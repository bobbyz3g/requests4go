package requests4go

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
)

// ErrResultConsumed indicates a closed, transferred or otherwise consumed result.
var ErrResultConsumed = errors.New("result has already been consumed")

// ErrResponseTooLarge indicates that a read terminal exceeded LimitBody.
var ErrResponseTooLarge = errors.New("response body exceeds limit")

// StatusError records a response rejected by Require2xx or RequireStatus.
// Header is a snapshot. The response body has been closed without being read.
type StatusError struct {
	StatusCode int
	Status     string
	Header     http.Header
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("unexpected HTTP status: %d", e.StatusCode)
}

// Result owns the response returned by RequestBuilder.Send, or its first error.
// It must not be copied or used concurrently. Call Response to transfer ownership
// or Close to release it. A zero Result has no response and is already consumed.
type Result struct {
	response *Response
	err      error
	consumed bool
	limit    int64
	limited  bool
}

func (r *Result) available() error {
	if r.err != nil {
		return r.err
	}
	if r.consumed || r.response == nil {
		return ErrResultConsumed
	}
	return nil
}

// Require2xx accepts status codes 200 through 299. A rejected response is closed
// immediately without reading its body and records a *StatusError.
func (r *Result) Require2xx() *Result {
	if err := r.available(); err != nil {
		if r.err == nil {
			r.err = err
		}
		return r
	}
	return r.require(r.response.StatusCode >= 200 && r.response.StatusCode < 300)
}

// RequireStatus accepts exactly code. A mismatch records a *StatusError and
// closes the response without reading it. Checks after an error do nothing.
func (r *Result) RequireStatus(code int) *Result {
	if err := r.available(); err != nil {
		if r.err == nil {
			r.err = err
		}
		return r
	}
	return r.require(r.response.StatusCode == code)
}

func (r *Result) require(ok bool) *Result {
	if !ok {
		r.err = &StatusError{
			StatusCode: r.response.StatusCode,
			Status:     r.response.Status,
			Header:     r.response.Header.Clone(),
		}
		_ = r.Close()
	}
	return r
}

// LimitBody limits the bytes read by Text, Bytes, Save and JSON terminals, including
// JSON whitespace. It applies to bytes exposed by Body, after any automatic
// decompression. Reading one extra byte detects overflow and returns
// ErrResponseTooLarge. Zero accepts only an empty body. A negative limit records
// an error and closes the response. The default is unlimited; the last limit wins.
// This does not limit reads after Response transfers ownership to the caller.
func (r *Result) LimitBody(n int64) *Result {
	if err := r.available(); err != nil {
		if r.err == nil {
			r.err = err
		}
		return r
	}
	if n < 0 {
		r.err = errors.New("negative response body limit")
		_ = r.Close()
		return r
	}
	r.limit, r.limited = n, true
	return r
}

// consume preserves the operation error over a subsequent cleanup error.
func (r *Result) consume(read func(*Response) error) error {
	if err := r.available(); err != nil {
		return err
	}
	r.err = read(r.response)
	return r.Close()
}

func (r *Result) readBody(resp *Response) ([]byte, error) {
	if !r.limited {
		return io.ReadAll(resp.Body)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, r.limit))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) < r.limit {
		return data, nil
	}
	// Probe separately so even MaxInt64 does not overflow when adding one.
	if err := checkBodyOverflow(resp.Body); err != nil {
		return nil, err
	}
	return data, nil
}

// Bytes reads and closes the response, returning nil on any read or close error.
// It consumes the Result once and observes LimitBody. It does not check status.
func (r *Result) Bytes() ([]byte, error) {
	var data []byte
	err := r.consume(func(resp *Response) (err error) {
		data, err = r.readBody(resp)
		return err
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Text reads and closes the response, returning an empty string on error.
// It observes LimitBody and converts bytes directly to a string without charset conversion.
func (r *Result) Text() (string, error) {
	data, err := r.Bytes()
	return string(data), err
}

// DecodeJSONInto decodes one complete JSON value into dst and closes the response
// on success or failure. It uses encoding/json semantics, observes LimitBody and
// requires application/json (with optional media type parameters). A mismatch
// returns ErrNotJSONContent without reading the body. dst must be a non-nil
// pointer; decoding errors may leave it partially modified, as with json.Unmarshal.
func (r *Result) DecodeJSONInto(dst any) error {
	return r.consume(func(resp *Response) error {
		mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil || mediaType != AppJSON {
			return ErrNotJSONContent
		}
		data, err := r.readBody(resp)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, dst)
	})
}

// DecodeJSON decodes one complete JSON value and closes the response, following
// DecodeJSONInto's rules. On any error it returns the zero value of T.
func (r *Result) DecodeJSON[T any]() (T, error) {
	var value T
	if err := r.DecodeJSONInto(&value); err != nil {
		var zero T
		return zero, err
	}
	return value, nil
}

// Discard closes the response without reading it and consumes the Result once.
// Unlike Close, a repeated successful Discard returns ErrResultConsumed.
func (r *Result) Discard() error {
	return r.consume(func(*Response) error { return nil })
}

// Response returns the request error, or transfers the response to the caller.
// On success the caller must close it. Subsequent Response calls return
// ErrResultConsumed; Close after transfer does not close the caller's response.
func (r *Result) Response() (*Response, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.consumed || r.response == nil {
		return nil, ErrResultConsumed
	}
	resp := r.response
	r.response = nil
	r.consumed = true
	return resp, nil
}

// Close releases an owned response without reading its remaining content.
// It returns the first request or cleanup error. Repeated calls do not close
// again and return the same error. After transfer, Close is a no-op.
func (r *Result) Close() error {
	if r.response != nil {
		err := r.response.Close()
		r.response = nil
		if r.err == nil {
			r.err = err
		}
	}
	r.consumed = true
	return r.err
}

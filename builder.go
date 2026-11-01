package requests4go

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

// ErrRequestSent indicates that Send has already been called on a builder.
var ErrRequestSent = errors.New("request has already been sent")

// ErrBodyAlreadySet indicates conflicting body sources in a builder.
var ErrBodyAlreadySet = errors.New("request body has already been set")

// RequestBuilder configures a single request. Create one with Client.NewRequest.
// It must not be copied or used concurrently. Setters retain the first error;
// subsequent setters do nothing. Setters also do nothing after Send.
type RequestBuilder struct {
	client   *Client
	ctx      context.Context
	method   string
	url      string
	headers  http.Header
	query    url.Values
	body     io.Reader
	openBody func() (io.ReadCloser, error)
	bodySet  bool
	err      error
	sent     bool
}

// NewRequest starts a builder without sending a request. ctx must not be nil.
// Use Request or an HTTP method setter to select the destination, then Send.
func (c *Client) NewRequest(ctx context.Context) *RequestBuilder {
	b := &RequestBuilder{client: c, ctx: ctx, headers: make(http.Header), query: make(url.Values)}
	if ctx == nil {
		b.err = errors.New("nil request context")
	}
	return b
}

func (b *RequestBuilder) active() bool { return b.err == nil && !b.sent }

// Request selects the method and URL. A later successful call replaces both.
// URL and method validation follows NewRequestWithContext; transport-level
// restrictions, such as supported URL schemes, are checked at Send time.
func (b *RequestBuilder) Request(method, url string) *RequestBuilder {
	if !b.active() {
		return b
	}
	if _, err := NewRequestWithContext(b.ctx, method, url); err != nil {
		b.err = err
		return b
	}
	b.method, b.url = method, url
	return b
}

// Get selects a GET request.
func (b *RequestBuilder) Get(url string) *RequestBuilder { return b.Request(http.MethodGet, url) }

// Post selects a POST request.
func (b *RequestBuilder) Post(url string) *RequestBuilder { return b.Request(http.MethodPost, url) }

// Put selects a PUT request.
func (b *RequestBuilder) Put(url string) *RequestBuilder { return b.Request(http.MethodPut, url) }

// Patch selects a PATCH request.
func (b *RequestBuilder) Patch(url string) *RequestBuilder { return b.Request(http.MethodPatch, url) }

// Delete selects a DELETE request.
func (b *RequestBuilder) Delete(url string) *RequestBuilder { return b.Request(http.MethodDelete, url) }

// Head selects a HEAD request.
func (b *RequestBuilder) Head(url string) *RequestBuilder { return b.Request(http.MethodHead, url) }

// Options selects an OPTIONS request.
func (b *RequestBuilder) Options(url string) *RequestBuilder {
	return b.Request(http.MethodOptions, url)
}

// Header replaces all values for name. Names are case insensitive.
// Special HTTP fields such as Host follow net/http.Header semantics.
func (b *RequestBuilder) Header(name, value string) *RequestBuilder {
	if b.active() {
		b.headers.Set(name, value)
	}
	return b
}

// Headers copies headers, replacing the values of each supplied name.
// As with http.Header, use canonical names when supplying multiple values.
func (b *RequestBuilder) Headers(headers http.Header) *RequestBuilder {
	if b.active() {
		for name, values := range headers {
			b.headers[http.CanonicalHeaderKey(name)] = slices.Clone(values)
		}
	}
	return b
}

// Query copies values, replacing all existing values for each supplied key,
// including values in the destination URL. Other URL query keys are preserved.
func (b *RequestBuilder) Query(values url.Values) *RequestBuilder {
	if b.active() {
		for name, values := range values {
			b.query[name] = slices.Clone(values)
		}
	}
	return b
}

func (b *RequestBuilder) acceptBody() bool {
	if !b.active() {
		return false
	}
	if b.bodySet {
		b.err = ErrBodyAlreadySet
		return false
	}
	b.bodySet = true
	return true
}

// JSON encodes a snapshot of value and sets Content-Type to application/json.
// A second body source records ErrBodyAlreadySet without encoding its value.
func (b *RequestBuilder) JSON(value any) *RequestBuilder {
	if b.acceptBody() {
		data, err := json.Marshal(value)
		if err != nil {
			b.err = err
			return b
		}
		b.body = bytes.NewReader(data)
		b.headers.Set("Content-Type", AppJSON)
	}
	return b
}

// Form encodes a snapshot of values, including repeated fields, and sets
// Content-Type to application/x-www-form-urlencoded. It is a body source.
func (b *RequestBuilder) Form(values url.Values) *RequestBuilder {
	if b.acceptBody() {
		b.body = strings.NewReader(values.Encode())
		b.headers.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return b
}

// Body selects reader as the single body source, without reading or closing it.
// Even nil selects a body source. The caller owns an accepted reader until Send;
// Send transfers ownership of its closer, including on construction failure.
// Rejected or skipped readers remain the caller's responsibility.
func (b *RequestBuilder) Body(reader io.Reader) *RequestBuilder {
	if b.acceptBody() {
		b.body = reader
	}
	return b
}

// Send attempts the request once and always returns a non-nil Result.
// Repeated calls return ErrRequestSent without changing the previous Result.
// Construction errors prevent sending and close an accepted body. Once handed
// to Client.Do, body cleanup follows net/http, which may close it asynchronously.
// The caller must consume or close the Result, or transfer its Response.
func (b *RequestBuilder) Send() *Result {
	if b.sent {
		return &Result{err: ErrRequestSent}
	}
	b.sent = true
	if b.err == nil {
		if b.url == "" {
			b.err = errors.New("missing request URL")
		} else if b.client == nil {
			b.err = errors.New("missing request client")
		}
	}
	var req *http.Request
	if b.err == nil && b.openBody != nil {
		if b.err = b.ctx.Err(); b.err == nil {
			b.body, b.err = b.openBody()
		}
		b.openBody = nil
	}
	if b.err == nil {
		req, b.err = NewRequestWithContext(b.ctx, b.method, b.url, func(req *http.Request) error {
			req.Header = b.headers.Clone()
			query := req.URL.Query()
			for key, values := range b.query {
				query[key] = slices.Clone(values)
			}
			if len(b.query) > 0 {
				req.URL.RawQuery = query.Encode()
			}
			return setRequestBody(req, b.body)
		})
	}
	if b.err != nil {
		if closer, ok := b.body.(io.Closer); ok {
			// Preserve the construction error if cleanup also fails.
			_ = closer.Close()
		}
		b.body = nil
		return &Result{err: b.err}
	}
	b.body = nil
	resp, err := b.client.Do(req)
	return &Result{response: resp, err: err}
}

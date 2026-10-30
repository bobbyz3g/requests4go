package requests4go

import "errors"

// ErrResultConsumed indicates a closed, transferred or otherwise consumed result.
var ErrResultConsumed = errors.New("result has already been consumed")

// Result owns the response returned by RequestBuilder.Send, or its first error.
// It must not be copied or used concurrently. Call Response to transfer ownership
// or Close to release it. A zero Result has no response and is already consumed.
type Result struct {
	response *Response
	err      error
	consumed bool
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

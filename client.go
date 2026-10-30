package requests4go

import (
	"context"
	"net/http"
)

var defaultHTTPClient = &http.Client{}
var defaultClient Client

// Client sends HTTP requests using an underlying net/http client.
// Its zero value uses a shared client with no timeout or cookie jar.
// A Client can be shared by goroutines, provided its underlying client and
// transport are safe for concurrent use and are not reconfigured during use.
type Client struct {
	httpClient *http.Client
}

// NewClient wraps client without copying or modifying it. Configure its timeout,
// transport, redirect policy and cookie jar before use. A nil client uses the
// same defaults as the zero value of Client, independently of http.DefaultClient.
func NewClient(client *http.Client) *Client {
	return &Client{httpClient: client}
}

// Request builds and sends a request. ctx must not be nil and controls the
// request through response body reads. Options have the same behavior as in
// NewRequestWithContext. A construction error prevents the request from being sent.
// Resource ownership during construction follows the supplied options' behavior.
// The caller must close the response on success, including for non-2xx statuses.
func (c *Client) Request(ctx context.Context, method, url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequestWithContext(ctx, method, url, opts...)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Do sends req using its context and the underlying client's policies.
// HTTP error status codes do not themselves cause an error.
// On success the caller owns the response and must close it. Do does not read,
// decode or cache its body. Request body ownership follows http.Client.Do.
func (c *Client) Do(req *http.Request) (*Response, error) {
	client := c.httpClient
	if client == nil {
		client = defaultHTTPClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return NewResponse(resp), nil
}

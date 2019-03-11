package requests4go

import "net/http"

type RequestArguments struct {
	Client *http.Client
	// Headers is a map of HTTP header.
	Headers map[string]string

	// Params is a map of query strings in Get request.
	Params map[string]string

	// Auth is basic HTTP authentication formatting the username and password in base64,
	// the format is:
	// []string{username, password}
	Auth []string

	Cookies []*http.Cookie

	Json map[string]string
}

var DefaultRequestArguments = &RequestArguments{
	Client:  http.DefaultClient,
	Headers: defaultHeaders,
	Params:  nil,
	Auth:    nil,
	Cookies: nil,
	Json:    nil,
}

// NewRequestArguments returns a new *RequestArguments with args.
// TODO: Create RequestArguments according map.
func NewRequestArguments(kwargs map[string]interface{}) *RequestArguments {
	return nil
}

// sendRequest sends http request and returns the response.
func sendRequest(method, url string, args *RequestArguments) (*Response, error) {
	req, err := prepareRequest(method, url, nil)

	if err != nil {
		return nil, err
	}

	for k, v := range args.Headers {
		req.Header.Set(k, v)
	}

	resp, err := args.Client.Do(req)

	if err != nil {
		return nil, err
	}

	return &Response{resp}, nil
}

// prepareRequest prepares http.Request according to method, url and RequestArguments.
func prepareRequest(method, url string, args *RequestArguments) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if args == nil {
		args = DefaultRequestArguments
	}
	if args.Auth != nil {
		req.SetBasicAuth(args.Auth[0], args.Auth[1])
	}
	return req, err
}

func Get(url string, args *RequestArguments) (*Response, error) {
	if args == nil {
		args = DefaultRequestArguments
	}

	return sendRequest("GET", url, args)
}

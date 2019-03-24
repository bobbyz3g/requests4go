package requests4go

import (
	"net/http"
	"strings"
)

// Session allows user use cookies between HTTP requests.
type Session struct {
	Args   *RequestArguments
	Client *http.Client
}

// NewSession returns a session struct.
func NewSession(args *RequestArguments) *Session {
	if args == nil {
		args = &RequestArguments{
			Client: &http.Client{
				Jar: setDefaultJar(),
			},
			Headers:       defaultHeaders,
			RedirectLimit: defaultRedirectLimit,
		}
	}

	if args.Client == nil {
		args.Client = &http.Client{
			Jar: setDefaultJar(),
		}
	}

	return &Session{args, args.Client}
}

// mergeRequestsArguments merges the param args and Session.Args.
func (s *Session) mergeRequestsArguments(args *RequestArguments) *RequestArguments {
	if args == nil {
		args = &RequestArguments{}
	}

	if args.Client == nil {
		args.Client = s.Client
	}

	if len(s.Args.Headers) > 0 || len(args.Headers) > 0 {
		headers := make(map[string]string)
		for k, v := range s.Args.Headers {
			headers[k] = v
		}
		for k, v := range args.Headers {
			headers[k] = v
		}
		args.Headers = headers
	}
	return args
}

func (s *Session) request(method string, url string, args *RequestArguments) (*Response, error) {
	method = strings.ToUpper(method)
	args = s.mergeRequestsArguments(args)
	return sendRequest(method, url, args)
}

// Get sends a GET request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func (s *Session) Get(url string, args *RequestArguments) (*Response, error) {
	return s.request("GET", url, args)
}

// Put sends a PUT request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func (s *Session) Put(url string, args *RequestArguments) (*Response, error) {
	return s.request("PUT", url, args)
}

// Post sends a POST request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func (s *Session) Post(url string, args *RequestArguments) (*Response, error) {
	return s.request("POST", url, args)
}

// Delete sends a DELETE request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func (s *Session) Delete(url string, args *RequestArguments) (*Response, error) {
	return s.request("DELETE", url, args)
}

// Patch sends a PATCH request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func (s *Session) Patch(url string, args *RequestArguments) (*Response, error) {
	return s.request("PATCH", url, args)
}

// Head sends a HEAD request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func (s *Session) Head(url string, args *RequestArguments) (*Response, error) {
	return s.request("HEAD", url, args)
}

// Options sends a OPTIONS request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func (s *Session) Option(url string, args *RequestArguments) (*Response, error) {
	return s.request("OPTION", url, args)
}

package requests4go

import (
	"errors"
	"net/http"
)

const (
	defaultUserAgent     = "Request4go"
	defaultContentType   = "application/x-www-form-urlencoded; charset=utf-8"
	defaultJsonType      = "application/json; charset=utf-8"
	defaultRedirectLimit = 10
)

var (
	defaultHeaders = map[string]string{
		"Connection":      "keep-alive",
		"Accept-Encoding": "gzip, deflate",
		"Accept":          "*/*",
		"User-Agent":      defaultUserAgent,
	}

	ErrRedirectLimitExceeded = errors.New("requests4go: Request exceeded redirect count limit")
)

func addCheckRedirectLimit(args *RequestArguments) {
	if args.Client.CheckRedirect != nil {
		return
	}
	args.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= args.RedirectLimit {
			return ErrRedirectLimitExceeded
		}
		for key, vv := range via[0].Header {
			for _, val := range vv {
				req.Header.Add(key, val)
			}
		}
		return nil
	}
}

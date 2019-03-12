package requests4go

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-querystring/query"
)

type RequestArguments struct {
	Client *http.Client
	// Headers is a map of HTTP header.
	Headers map[string]string

	// Params is a map of URL query params in GET request.
	Params map[string]string

	// ObjectParam is a struct that encapsulates URL query params in GET request.
	ObjectParam interface{}

	// Auth is basic HTTP authentication formatting the username and password in base64,
	// the format is:
	// []string{username, password}
	Auth []string

	Cookies []*http.Cookie

	Json map[string]string
}

var DefaultRequestArguments = &RequestArguments{
	Client:      http.DefaultClient,
	Headers:     defaultHeaders,
	Params:      nil,
	ObjectParam: nil,
	Auth:        nil,
	Cookies:     nil,
	Json:        nil,
}

// NewRequestArguments returns a new *RequestArguments with args.
// Optional keys:
//   "Client", "Headers", "Params", "Auth", "Cookies", "Json".
func NewRequestArguments(args map[string]interface{}) *RequestArguments {
	a := DefaultRequestArguments

	for k, v := range args {
		switch k {
		case "Client":
			a.Client = v.(*http.Client)
		case "Headers":
			a.Headers = v.(map[string]string)
		case "Params":
			a.Params = v.(map[string]string)
		case "ObjectParam":
			a.ObjectParam = v.(interface{})
		case "Auth":
			a.Auth = v.([]string)
		case "Cookies":
			a.Cookies = v.([]*http.Cookie)
		case "Json":
			a.Json = v.(map[string]string)
		}
	}
	return a
}

// sendRequest sends http request and returns the response.
func sendRequest(method, reqUrl string, args *RequestArguments) (*Response, error) {
	if args == nil {
		args = DefaultRequestArguments
	}

	req, err := prepareRequest(method, reqUrl, args)

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
func prepareRequest(method, originUrl string, args *RequestArguments) (*http.Request, error) {
	reqUrl, err := prepareURL(originUrl, args.Params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, reqUrl, nil)

	if args.Auth != nil {
		req.SetBasicAuth(args.Auth[0], args.Auth[1])
	}

	return req, err
}

// prepareURL prepares new URL with url query params.
func prepareURL(originUrl string, params map[string]string) (string, error) {
	if len(params) == 0 {
		return originUrl, nil
	}

	parsedUrl, err := url.Parse(originUrl)

	if err != nil {
		return "", err
	}

	rawQuery, err := url.ParseQuery(parsedUrl.RawQuery)

	if err != nil {
		return "", err
	}

	for k, v := range params {
		rawQuery.Set(k, v)
	}

	return mergeParams(parsedUrl, rawQuery), nil
}

// prepareUrlWithStruct prepares new Url with object param.
func prepareURLWithStruct(originUrl string, paramStruct interface{}) (string, error) {
	parsedUrl, err := url.Parse(originUrl)

	if err != nil {
		return "", err
	}

	rawQuery, err := url.ParseQuery(parsedUrl.RawQuery)
	if err != nil {
		return "", err
	}

	params, err := query.Values(paramStruct)
	if err != nil {
		return "", err
	}

	for k, value := range params {
		for _, v := range value {
			rawQuery.Add(k, v)
		}
	}

	return mergeParams(parsedUrl, rawQuery), nil
}

// mergeParams merges the url and params, returns new url.
func mergeParams(parsedUrl *url.URL, rawQuery url.Values) string {
	newUrl := strings.Replace(parsedUrl.String(), "?"+parsedUrl.RawQuery, "", -1)
	return newUrl + "?" + rawQuery.Encode()
}

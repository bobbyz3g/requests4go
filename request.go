package requests4go

import (
	"bytes"
	"encoding/json"
	"io"
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

	// ObjectParam is a struct that encapsulates URL query params within GET request.
	ObjectParam interface{}

	// Auth is basic HTTP authentication formatting the username and password in base64,
	// the format is:
	// []string{username, password}
	Auth []string

	Cookies []*http.Cookie

	// Json can be []byte, string or struct.
	// When you want to send a JSON within request, you can use it.
	Json interface{}

	// Data is a map stores the key values, will be converted into the body of
	// Post request.
	Data map[string]string
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
			a.Json = v
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

	return NewResponse(args.Client.Do(req))
}

// prepareRequest prepares http.Request according to method, url and RequestArguments.
func prepareRequest(method, reqUrl string, args *RequestArguments) (*http.Request, error) {
	var err error

	switch {
	case len(args.Params) != 0:
		if reqUrl, err = prepareURL(reqUrl, args.Params); err != nil {
			return nil, err
		}
	case args.ObjectParam != nil:
		if reqUrl, err = prepareURLWithStruct(reqUrl, args.ObjectParam); err != nil {
			return nil, err
		}
	}

	body, err := prepareBody(args)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, reqUrl, body)

	if args.Auth != nil {
		req.SetBasicAuth(args.Auth[0], args.Auth[1])
	}

	for k, v := range args.Headers {
		req.Header.Set(k, v)
	}
	return req, err
}

// prepareBody prepares the give HTTP body.
func prepareBody(args *RequestArguments) (io.Reader, error) {
	if args.Json != nil {
		args.Headers["Content-Type"] = defaultJsonType
		return prepareJsonBody(args.Json)
	}

	if args.Data != nil {
		args.Headers["Content-Type"] = defaultContentType
		return prepareDataBody(args.Data)
	}
	return nil, nil
}

// prepareDataBody prepares the given HTTP Data body.
func prepareDataBody(data map[string]string) (io.Reader, error) {
	reader := strings.NewReader(encodeParams(data))
	return reader, nil
}

// encodeParams encodes parameters in a piece of data.
func encodeParams(data map[string]string) string {
	vs := &url.Values{}
	for k, v := range data {
		vs.Set(k, v)
	}
	return vs.Encode()
}

// prepareJsonBody prepares the give HTTP Json body.
func prepareJsonBody(JSON interface{}) (io.Reader, error) {
	var reader io.Reader
	switch JSON.(type) {
	case string:
		reader = strings.NewReader(JSON.(string))
	case []byte:
		reader = bytes.NewReader(JSON.([]byte))
	default:
		byteS, err := json.Marshal(JSON)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(byteS)
	}
	return reader, nil
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

package requests4go

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-querystring/query"
)

type FileField struct {
	// FileName specifies name of file that you wish to upload.
	FileName string

	// FieldName specifies form field name.
	FieldName string

	FileContent io.ReadCloser
}

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

	// Files specifies the files you wish to post.
	Files []FileField
}

var DefaultRequestArguments = &RequestArguments{
	Client:      http.DefaultClient,
	Headers:     defaultHeaders,
	Params:      nil,
	ObjectParam: nil,
	Auth:        nil,
	Cookies:     nil,
	Json:        nil,
	Data:        nil,
	Files:       nil,
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

	if args.Files != nil {
		body, contentType, err := prepareFilesBody(args.Files, args.Data)
		args.Headers["Content-type"] = contentType
		return body, err
	}

	if args.Data != nil {
		args.Headers["Content-Type"] = defaultContentType
		return prepareDataBody(args.Data)
	}

	return nil, nil
}

// prepareFilesBody prepares the body for a multipart/form-data request.
// It returns body, contentType and error.
func prepareFilesBody(files []FileField, data map[string]string) (io.Reader, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, file := range files {
		fileWriter, err := writer.CreateFormFile(file.FieldName, file.FileName)
		if err != nil {
			return nil, "", err
		}

		if _, err := io.Copy(fileWriter, file.FileContent); err != nil {
			return nil, "", err
		}

		if err := file.FileContent.Close(); err != nil {
			return nil, "", err
		}
	}

	for key, value := range data {
		err := writer.WriteField(key, value)
		if err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	contentType := writer.FormDataContentType()
	return body, contentType, nil
}

// prepareDataBody prepares the body for a application/x-www-form-urlencoded request.
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

// prepareJsonBody prepares the body for application/json request.
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

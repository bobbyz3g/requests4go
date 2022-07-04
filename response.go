// Developed by Kaiser925 on 2021/2/2.
// Lasted modified 2021/2/2.
// Copyright (c) 2021.  All rights reserved
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package requests4go

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

var ErrNotJSONContent = errors.New("content type not application/json")

// Response is a wrapper of the http.Response.
// It opens up new methods for http.Response.
type Response struct {
	// Embed an HTTP response directly. This makes a *http.Response act exactly
	// like an *http.Response so that all meta methods are supported.
	*http.Response
}

// NewResponse returns new Response
func NewResponse(resp *http.Response) *Response {
	return &Response{
		resp,
	}
}

// Ok returns true if the status code is less than 400.
func (r *Response) Ok() bool {
	return r.StatusCode < 400 && r.StatusCode >= 200
}

// Close is to support io.ReadCloser.
func (r *Response) Close() error {
	_, err := io.Copy(ioutil.Discard, r)
	if err != nil {
		return err
	}
	if r.Body == nil {
		return nil
	}
	return r.Body.Close()
}

// Read is to support io.ReadCloser.
func (r *Response) Read(p []byte) (n int, err error) {
	return r.Body.Read(p)
}

// Text reads body of response and returns content of response in string.
func (r *Response) Text() (string, error) {
	content, err := r.Content()
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// Content reads body of response and returns content of response in bytes.
func (r *Response) Content() ([]byte, error) {
	content, err := ioutil.ReadAll(r.Body)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return content, nil
}

// SaveContent reads body of response and saves response body to file.
func (r *Response) SaveContent(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, r.Body); err != nil {
		return err
	}
	return nil
}

// JSON reads body of response and unmarshal the response content to v.
func (r *Response) JSON(v interface{}) error {
	ct := r.Header.Get("Content-Type")
	if ct != AppJSON {
		return errors.New("content type not application/json")
	}
	content, err := r.Content()
	if err != nil {
		return err
	}
	return json.Unmarshal(content, v)
}

// XML unmarshal the response content as XML.
func (r *Response) XML(v interface{}) error {
	content, err := r.Content()
	if err != nil {
		return err
	}
	return xml.Unmarshal(content, v)
}

// Unmarshaler is the interface implemented by types
// that can unmarshal from response content.
type Unmarshaler interface {
	Unmarshal([]byte) error
}

func (r *Response) Unmarshal(u Unmarshaler) error {
	content, err := r.Content()
	if err != nil {
		return err
	}
	return u.Unmarshal(content)
}

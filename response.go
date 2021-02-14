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
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/bitly/go-simplejson"
)

// Response is a wrapper of the http.Response.
// It opens up new methods for http.Response.
type Response struct {
	*http.Response

	content []byte
}

// NewResponse returns new Response
func NewResponse(resp *http.Response) *Response {
	return &Response{
		resp,
		nil,
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
	return r.Body.Close()
}

// Read is to support io.ReadCloser.
func (r *Response) Read(p []byte) (n int, err error) {
	return r.Body.Read(p)
}

// Text returns content of response in string.
// It will close response.
func (r *Response) Text() (string, error) {
	content, err := r.loadContent()
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// Content returns content of response in bytes.
// It will close response.
func (r *Response) Content() ([]byte, error) {
	content, err := r.loadContent()
	if err != nil {
		return nil, err
	}
	return content, nil
}

// JSON returns simplejson.Json.
// See the usage of simplejson on https://godoc.org/github.com/bitly/go-simplejson.
func (r *Response) SimpleJSON() (*simplejson.Json, error) {
	content, err := r.loadContent()
	if err != nil {
		return nil, fmt.Errorf("Json error: %w", err)
	}
	return simplejson.NewJson(content)
}

// SaveContent saves response body to file and closes the response.
func (r *Response) SaveContent(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	content, err := r.loadContent()
	if err != nil {
		return err
	}

	_, err = f.Write(content)
	if err != nil {
		return err
	}
	return nil
}

// JSON unmarshal the response content to v.
func (r *Response) JSON(v interface{}) error {
	content, err := r.loadContent()
	if err != nil {
		return err
	}
	return json.Unmarshal(content, v)
}

func (r *Response) loadContent() ([]byte, error) {
	if r.content != nil {
		return r.content, nil
	}
	var reader io.ReadCloser

	defer func() {
		reader.Close()
	}()

	var err error
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		if reader, err = gzip.NewReader(r); err != nil {
			return nil, err
		}
	case "deflate":
		if reader, err = zlib.NewReader(r); err != nil {
			return nil, err
		}
	default:
		reader = r
	}
	content, err := ioutil.ReadAll(reader)
	if err != nil && err != io.EOF {
		return nil, err
	}
	r.content = content
	return content, nil
}

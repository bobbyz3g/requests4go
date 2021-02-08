// Developed by Kaiser925 on 2021/1/25.
// Lasted modified 2021/1/25.
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
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

// A RequestOption is represent a option of request.
// You can use it to custom request.
// You can also define your own RequestOption.
type RequestOption func(req *http.Request) error

// Params sets url query parameters for the request.
// It replaces any existing values.
func Params(params map[string]string) RequestOption {
	return func(req *http.Request) error {
		q := req.URL.Query()
		for key, value := range params {
			q.Set(key, value)
		}
		req.URL.RawQuery = q.Encode()
		return nil
	}
}

// Auth sets basic auth for the request.
func Auth(name string, password string) RequestOption {
	return func(req *http.Request) error {
		req.SetBasicAuth(name, password)
		return nil
	}
}

// Headers sets the header for the request.
// It replaces any existing values.
// The key is case insensitive.
func Headers(headers map[string]string) RequestOption {
	return func(req *http.Request) error {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return nil
	}
}

// JSON encodes the value to json data and set content-type be application/json.
func JSON(v interface{}) RequestOption {
	return func(req *http.Request) error {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		v := bytes.NewReader(b)
		req.ContentLength = int64(v.Len())
		req.Body = ioutil.NopCloser(v)
		snapshot := *v
		req.GetBody = func() (io.ReadCloser, error) {
			r := snapshot
			return ioutil.NopCloser(&r), nil
		}
		req.Header.Set("Content-Type", "application/json")
		return nil
	}
}

// File sets request body be file content.
func File(filename string) RequestOption {
	return func(req *http.Request) error {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		// Copy file to new ReaderCloser, not use file directly
		b := &bytes.Buffer{}
		if _, err := io.Copy(b, f); err != nil {
			return err
		}
		req.ContentLength = int64(b.Len())
		req.Body = ioutil.NopCloser(b)
		snapshot := *b
		req.GetBody = func() (io.ReadCloser, error) {
			r := snapshot
			return ioutil.NopCloser(&r), nil
		}
		return nil
	}
}

// Cookies add the cookie to http.Request.
func Cookies(c map[string]string) RequestOption {
	return func(req *http.Request) error {
		for k, v := range c {
			req.AddCookie(&http.Cookie{Name: k, Value: v})
		}
		return nil
	}
}

// Data sends form-encoded data.
func Data(form map[string]string) RequestOption {
	return func(req *http.Request) error {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if len(req.PostForm) == 0 {
			req.PostForm = make(url.Values)
		}
		for k, v := range form {
			req.PostForm.Set(k, v)
		}
		return nil
	}
}

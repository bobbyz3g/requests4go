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
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

// A RequestOption is represent a option of request.
type RequestOption func(req *http.Request) error

// NewRequestWithOpt wrappers the NewRequestWithOptCtx
func NewRequestWithOpt(
	method string,
	url string,
	opts ...RequestOption,
) (*http.Request, error) {
	return NewRequestWithOptCtx(context.Background(), method, url, opts...)
}

// NewRequestWithOptCtx builds a new *http.Request with Context and RequestOption
func NewRequestWithOptCtx(
	ctx context.Context,
	method string,
	url string,
	opts ...RequestOption,
) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		if err := opt(req); err != nil {
			return nil, err
		}
	}
	return req, nil
}

// Params sets url query parameters for the request.
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

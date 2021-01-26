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
	"context"
	"net/http"
)

// A RequestOption is represent a option of request.
type RequestOption func(req *http.Request) error

// NewRequestWithOpt wrapper the NewRequestWithOptCtx
func NewRequestWithOpt(
	method string,
	url string,
	opts ...RequestOption,
) (*http.Request, error) {
	return NewRequestWithOptCtx(context.Background(), method, url, opts...)
}

// NewRequestWithOptCtx build a new *http.Request with Context and RequestOption
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

// Params set url query parameters for the request.
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

// Auth set basic auth for the request.
func Auth(name string, password string) RequestOption {
	return func(req *http.Request) error {
		req.SetBasicAuth(name, password)
		return nil
	}
}

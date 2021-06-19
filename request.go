// Developed by Kaiser925 on 2021/2/2.
// Lasted modified 2021/2/1.
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

// NewRequest wrappers the NewRequestWithContext
func NewRequest(method, url string, opts ...RequestOption) (*http.Request, error) {
	return NewRequestWithContext(context.Background(), method, url, opts...)
}

// NewRequestWithContext builds a new *http.Request with Context and RequestOption
//
// Note:
// If there are multiple options will modify the request body,
// only the last one will take effect.
// The order of options all will effect final request status.
func NewRequestWithContext(ctx context.Context, method, url string, opts ...RequestOption) (*http.Request, error) {
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

// Get sends "GET" request.
func Get(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("GET", url, opts...)
	if err != nil {
		return nil, err
	}
	return Do(req)
}

// Post sends "POST" request.
func Post(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("POST", url, opts...)
	if err != nil {
		return nil, err
	}
	return Do(req)
}

// Put sends "Put" request.
func Put(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("PUT", url, opts...)
	if err != nil {
		return nil, err
	}
	return Do(req)
}

// Patch sends "PATCH" request.
func Patch(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("PATCH", url, opts...)
	if err != nil {
		return nil, err
	}
	return Do(req)
}

// Head sends "HEAD" request.
func Head(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("HEAD", url, opts...)
	if err != nil {
		return nil, err
	}
	return Do(req)
}

// Options sends "OPTIONS" request.
func Options(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("OPTIONS", url, opts...)
	if err != nil {
		return nil, err
	}
	return Do(req)
}

// Delete sends "DELETE" request.
func Delete(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("DELETE", url, opts...)
	if err != nil {
		return nil, err
	}
	return Do(req)
}

// Do sends request and return the response.
func Do(req *http.Request) (*Response, error) {
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	return NewResponse(resp), nil
}

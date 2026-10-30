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
	"context"
	"net/http"
	"net/http/cookiejar"

	"golang.org/x/net/publicsuffix"
)

// Session allows user use cookies between HTTP requests.
// Configure Client before concurrent use. Replacing Client between requests is
// supported; a nil Client uses the package's shared client without a cookie jar.
type Session struct {
	Client *http.Client
}

// NewSession returns a session struct.
func NewSession() *Session {
	return &Session{
		Client: &http.Client{
			Jar: getDefaultJar(),
		},
	}
}

// Get sends a GET request, returns Response struct.
func (s *Session) Get(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(context.Background(), http.MethodGet, url, opts...)
}

// Put sends a PUT request, returns Response struct.
func (s *Session) Put(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(context.Background(), http.MethodPut, url, opts...)
}

// Post sends a POST request, returns Response struct.
func (s *Session) Post(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(context.Background(), http.MethodPost, url, opts...)
}

// Delete sends a DELETE request, returns Response struct.
func (s *Session) Delete(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(context.Background(), http.MethodDelete, url, opts...)
}

// Patch sends a PATCH request, returns Response struct.
func (s *Session) Patch(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(context.Background(), http.MethodPatch, url, opts...)
}

// Head sends a HEAD request, returns Response struct.
func (s *Session) Head(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(context.Background(), http.MethodHead, url, opts...)
}

// Options sends a OPTIONS request, returns Response struct.
func (s *Session) Options(url string, opts ...RequestOption) (*Response, error) {
	return s.Request(context.Background(), http.MethodOptions, url, opts...)
}

// Request builds and sends a request with ctx using the session's current Client.
// Its construction, status and response ownership rules are those of Client.Request.
func (s *Session) Request(ctx context.Context, method, url string, opts ...RequestOption) (*Response, error) {
	return NewClient(s.Client).Request(ctx, method, url, opts...)
}

// Do sends a prepared request using the session's current Client and cookie jar.
// Its status and response ownership rules are those of Client.Do.
func (s *Session) Do(req *http.Request) (*Response, error) {
	return NewClient(s.Client).Do(req)
}

func getDefaultJar() *cookiejar.Jar {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, _ := cookiejar.New(&options)
	return jar
}

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
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/http/cookiejar"
)

// Session allows user use cookies between HTTP requests.
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
	req, err := NewRequest("GET", url, opts...)
	if err != nil {
		return nil, err
	}
	return s.do(req)
}

// Put sends a PUT request, returns Response struct.
func (s *Session) Put(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("PUT", url, opts...)
	if err != nil {
		return nil, err
	}
	return s.do(req)
}

// Post sends a POST request, returns Response struct.
func (s *Session) Post(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("POST", url, opts...)
	if err != nil {
		return nil, err
	}
	return s.do(req)
}

// Delete sends a DELETE request, returns Response struct.
func (s *Session) Delete(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("DELETE", url, opts...)
	if err != nil {
		return nil, err
	}
	return s.do(req)
}

// Patch sends a PATCH request, returns Response struct.
func (s *Session) Patch(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("PATCH", url, opts...)
	if err != nil {
		return nil, err
	}
	return s.do(req)
}

// Head sends a HEAD request, returns Response struct.
func (s *Session) Head(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("POST", url, opts...)
	if err != nil {
		return nil, err
	}
	return s.do(req)
}

// Options sends a OPTIONS request, returns Response struct.
func (s *Session) Options(url string, opts ...RequestOption) (*Response, error) {
	req, err := NewRequest("OPTIONS", url, opts...)
	if err != nil {
		return nil, err
	}
	return s.do(req)
}

func (s *Session) do(req *http.Request) (*Response, error) {
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return NewResponse(resp), nil
}

func getDefaultJar() *cookiejar.Jar {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, _ := cookiejar.New(&options)
	return jar
}

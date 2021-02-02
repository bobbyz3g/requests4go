// Copyright 2019 The Kaiser925. All rights reserved.
// Use of this source code is governed by a Apache
// license that can be found in the LICENSE file.

package requests4go

import (
	"net/http"
)

// Session allows user use cookies between HTTP requests.
type Session struct {
	Client *http.Client
}

// NewSession returns a session struct.
func NewSession() *Session {
	return &Session{
		Client: &http.Client{
			Jar: setDefaultJar(),
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

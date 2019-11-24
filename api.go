// Copyright 2019 The Kaiser925. All rights reserved.
// Use of this source code is governed by a Apache
// license that can be found in the LICENSE file.

package requests4go

import "strings"

// DoRequest constructs and sends request, returns Response struct.
// Three options:
//   1. method.
//   2. url.
//   3. RequestArguments struct, can be nil.
func DoRequest(method string, url string, args *RequestArguments) (*Response, error) {
	method = strings.ToUpper(method)
	return sendRequest(method, url, args)
}

// Get sends a GET request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func Get(url string, args *RequestArguments) (*Response, error) {
	return DoRequest("GET", url, args)
}

// Post sends a POST request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func Post(url string, args *RequestArguments) (*Response, error) {
	return DoRequest("POST", url, args)
}

// Put sends a PUT request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func Put(url string, args *RequestArguments) (*Response, error) {
	return DoRequest("PUT", url, args)
}

// Delete sends a DELETE request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func Delete(url string, args *RequestArguments) (*Response, error) {
	return DoRequest("DELETE", url, args)
}

// Head sends a HEAD request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func Head(url string, args *RequestArguments) (*Response, error) {
	return DoRequest("HEAD", url, args)
}

// Options sends a OPTIONS request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func Options(url string, args *RequestArguments) (*Response, error) {
	return DoRequest("OPTIONS", url, args)
}

// Patch sends a PATCH request, returns Response struct.
// Two options:
//   1. Url.
//   2. RequestArguments struct, can be nil.
func Patch(url string, args *RequestArguments) (*Response, error) {
	return DoRequest("PATCH", url, args)
}

// Developed by Kaiser925 on 2021/2/2.
// Lasted modified 2021/1/27.
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

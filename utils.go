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

import (
	"errors"
	"net/http"
	"net/http/cookiejar"

	"golang.org/x/net/publicsuffix"
)

const (
	defaultUserAgent     = "Request4go"
	defaultContentType   = "application/x-www-form-urlencoded; charset=utf-8"
	defaultJSONType      = "application/json; charset=utf-8"
	defaultRedirectLimit = 10
)

var (
	defaultHeaders = map[string]string{
		"Connection":      "keep-alive",
		"Accept-Encoding": "gzip, deflate",
		"Accept":          "*/*",
		"User-Agent":      defaultUserAgent,
	}

	// ErrRedirectLimitExceeded will be returned when redirect times over limit.
	ErrRedirectLimitExceeded = errors.New("requests4go: Request exceeded redirect count limit")
)

func setDefaultJar() *cookiejar.Jar {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, _ := cookiejar.New(&options)
	return jar
}

func addCheckRedirectLimit(args *RequestArguments) {
	if args.Client.CheckRedirect != nil {
		return
	}
	args.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= args.RedirectLimit {
			return ErrRedirectLimitExceeded
		}
		for key, vv := range via[0].Header {
			for _, val := range vv {
				req.Header.Add(key, val)
			}
		}
		return nil
	}
}

func cookiesFromMap(m map[string]string) []*http.Cookie {
	l := len(m)
	cookies := make([]*http.Cookie, l)
	index := 0
	for key, val := range m {
		cookies[index] = &http.Cookie{Name: key, Value: val}
		index++
	}
	return cookies
}

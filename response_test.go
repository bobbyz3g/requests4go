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
	"bufio"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

type respTest struct {
	Raw  string
	Resp http.Response
	Body string
}

type respData struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func dummyReq(method string) *http.Request {
	return &http.Request{Method: method}
}

func TestResponse(t *testing.T) {
	resp, err := Get("http://httpbin.org/get", Params(map[string]string{
		"a": "1",
		"b": "2",
	}))

	if err != nil {
		t.Errorf("GET request error: got %s", err)
	}

	assert.Equal(t, err, nil)
	assert.Equal(t, resp.Ok(), true)
	text, err := resp.Text()
	assert.NotEqual(t, text, "")
	assert.Equal(t, err, nil)
}

func TestResponse_Content(t *testing.T) {
	var testcases = map[string]respTest{
		"text body": {
			Raw: "HTTP/1.0 200 OK\r\n" +
				"Connection: close\r\n" +
				"\r\n" +
				"Body here\n",
			Resp: http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Proto:      "HTTP/1.0",
				ProtoMajor: 1,
				ProtoMinor: 0,
				Request:    dummyReq("GET"),
				Header: http.Header{
					"Connection": {"close"},
				},
				Close:         true,
				ContentLength: -1,
			},
			Body: "Body here\n",
		},
		"multiple line body": {
			Raw: "HTTP/1.0 200 OK\r\n" +
				"Connection: close\r\n" +
				"\r\n" +
				"Body line one\n" +
				"Body line two\n",
			Resp: http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Proto:      "HTTP/1.0",
				ProtoMajor: 1,
				ProtoMinor: 0,
				Request:    dummyReq("GET"),
				Header: http.Header{
					"Connection": {"close"},
				},
				Close:         true,
				ContentLength: -1,
			},
			Body: "Body line one\nBody line two\n",
		},
		"body should not be read": {
			"HTTP/1.1 204 No Content\r\n" +
				"\r\n" +
				"Body should not be read!\n",

			http.Response{
				Status:        "204 No Content",
				StatusCode:    204,
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				Header:        http.Header{},
				Request:       dummyReq("GET"),
				Close:         false,
				ContentLength: 0,
			},

			"",
		},
		"json body": {
			Raw: "HTTP/1.0 200 OK\r\n" +
				"Connection: close\r\n" +
				"Content-Type: application/json\r\n" +
				"\r\n" +
				"{\"name\": \"foo\", \"age\": 10}\n",
			Resp: http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Proto:      "HTTP/1.0",
				ProtoMajor: 1,
				ProtoMinor: 0,
				Request:    dummyReq("GET"),
				Header: http.Header{
					"Connection": {"close"},
				},
				Close:         true,
				ContentLength: -1,
			},
			Body: "{\"name\": \"foo\", \"age\": 10}\n",
		},
		"chunked response without Content-Length.": {
			"HTTP/1.1 200 OK\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"\r\n" +
				"0a\r\n" +
				"Body here\n\r\n" +
				"09\r\n" +
				"continued\r\n" +
				"0\r\n" +
				"\r\n",

			http.Response{
				Status:           "200 OK",
				StatusCode:       200,
				Proto:            "HTTP/1.1",
				ProtoMajor:       1,
				ProtoMinor:       1,
				Request:          dummyReq("GET"),
				Header:           http.Header{},
				Close:            false,
				ContentLength:    -1,
				TransferEncoding: []string{"chunked"},
			},

			"Body here\ncontinued",
		},
		"chunked response with Content-Length.": {
			"HTTP/1.1 200 OK\r\n" +
				"Transfer-Encoding: chunked\r\n" +
				"Content-Length: 10\r\n" +
				"\r\n" +
				"0a\r\n" +
				"Body here\n\r\n" +
				"0\r\n" +
				"\r\n",

			http.Response{
				Status:           "200 OK",
				StatusCode:       200,
				Proto:            "HTTP/1.1",
				ProtoMajor:       1,
				ProtoMinor:       1,
				Request:          dummyReq("GET"),
				Header:           http.Header{},
				Close:            false,
				ContentLength:    -1,
				TransferEncoding: []string{"chunked"},
			},

			"Body here\n",
		},
	}
	for name, tc := range testcases {
		hresp, err := http.ReadResponse(bufio.NewReader(strings.NewReader(tc.Raw)), tc.Resp.Request)
		if err != nil {
			t.Errorf("#%s: %v", name, err)
			continue
		}
		resp := NewResponse(hresp)
		p, err := resp.Content()
		if err != nil {
			t.Errorf("#%s: Error = %v ", name, err)
			continue
		}
		if string(p) != tc.Body {
			t.Errorf("#%s: Body = %q want %q", name, p, tc.Body)
		}

		// read repeatedly
		p, err = resp.Content()
		if err != nil {
			t.Errorf("#%s: Error = %v ", name, err)
			continue
		}
		if len(p) != 0 {
			t.Errorf("#%s: Body = %q want []", name, p)
		}
	}
}

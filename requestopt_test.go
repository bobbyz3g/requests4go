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
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestParams(t *testing.T) {
	params := map[string]string{
		"b": "b",
		"a": "c",
	}
	req, _ := NewRequestWithOpt("get", "http://simple.org/path/?b=a", Params(params))
	assert.Equal(t, req.URL.String(), "http://simple.org/path/?a=c&b=b")
}

var authTests = []struct {
	username, password string
	ok                 bool
}{
	{"Aladdin", "open sesame", true},
	{"Aladdin", "open:sesame", true},
	{"", "", true},
}

func TestAuth(t *testing.T) {
	for _, tt := range authTests {
		r, _ := NewRequestWithOpt("GET", "http://example.com/", Auth(tt.username, tt.password))
		username, password, ok := r.BasicAuth()
		assert.Equal(t, ok, tt.ok)
		assert.Equal(t, username, tt.username)
		assert.Equal(t, password, tt.password)
	}
}

var headerTests = []struct {
	k string
	v string
}{
	{"apple", "ok"},
	{"banana", "okokok"},
}

func TestHeaders(t *testing.T) {
	for _, tt := range headerTests {
		r, _ := NewRequestWithOpt("GET", "http://example.com", Headers(map[string]string{tt.k: tt.v}))
		v := r.Header.Get(tt.k)
		assert.Equal(t, v, tt.v)
	}
}

func TestAll(t *testing.T) {
	params := map[string]string{
		"b": "b",
		"a": "c",
	}
	req, _ := NewRequestWithOpt("get", "http://simple.org/path/?b=a",
		Params(params),
		Auth(authTests[0].username, authTests[0].password))
	assert.Equal(t, req.URL.String(), "http://simple.org/path/?a=c&b=b")
	username, password, ok := req.BasicAuth()
	assert.Equal(t, ok, authTests[0].ok)
	assert.Equal(t, username, authTests[0].username)
	assert.Equal(t, password, authTests[0].password)
}

var jsonTests = struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}{
	"name",
	21,
}

func TestJSON(t *testing.T) {
	req, _ := NewRequestWithOpt("POST", "http://httpbin.org/post", JSON(jsonTests))
	b, _ := json.Marshal(jsonTests)
	reqE, _ := http.NewRequest("POST", "http://httpbin.org/post", bytes.NewReader(b))
	assert.Equal(t, req.Body, reqE.Body)
	assert.Equal(t, req.ContentLength, reqE.ContentLength)
}

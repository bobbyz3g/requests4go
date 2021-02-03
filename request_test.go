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
	"github.com/stretchr/testify/assert"
	"testing"
)

var testMap = map[string]string{
	"name": "name",
	"age":  "age",
}

func TestBaseGet(t *testing.T) {
	resp, err := Get("http://httpbin.org/get")

	if err != nil {
		t.Errorf("Get error: excepted no error, got %v", err)
	}

	assert.Equal(t, true, resp.Ok())
}

func TestCookieGet(t *testing.T) {
	resp, err := Get("http://httpbin.org/cookies", Cookies(testMap))

	if err != nil {
		t.Fatalf("Request error: %v", err)
	}

	JSON, err := resp.SimpleJSON()
	if err != nil {
		t.Fatalf("Get json error: %v", err)
	}

	name, _ := JSON.Get("cookies").Get("name").String()
	age, _ := JSON.Get("cookies").Get("age").String()

	assert.Equal(t, testMap["name"], name)
	assert.Equal(t, testMap["age"], age)
}

func TestBaseJsonPost(t *testing.T) {
	jsonStruct := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		"name",
		21,
	}

	resp, err := Post("http://httpbin.org/post", JSON(jsonStruct))
	if err != nil {
		t.Errorf("Reqeust erro: got %v", err)
	}

	JSON, err := resp.SimpleJSON()
	if err != nil {
		t.Errorf("%v \n", err)
	}
	name, _ := JSON.Get("json").Get("name").String()
	age, _ := JSON.Get("json").Get("age").Int()

	assert.Equal(t, jsonStruct.Name, name)
	assert.Equal(t, jsonStruct.Age, age)
}

func TestPut(t *testing.T) {
	resp, err := Put("http://httpbin.org/put")
	assert.Equal(t, err, nil)
	assert.Equal(t, resp.Ok(), true)
}

func TestDelete(t *testing.T) {
	resp, err := Delete("http://httpbin.org/delete")
	assert.Equal(t, err, nil)
	assert.Equal(t, resp.Ok(), true)
}

func TestPatch(t *testing.T) {
	resp, err := Options("http://httpbin.org/patch")
	assert.Equal(t, err, nil)
	assert.Equal(t, resp.Ok(), true)
}

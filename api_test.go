// Copyright 2019 The Kaiser925. All rights reserved.
// Use of this source code is governed by a Apache
// license that can be found in the LICENSE file.

package requests4go

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var testMap = map[string]string{
	"name": "name",
	"age":  "age",
}

func TestBaseFilePost(t *testing.T) {
	args := NewRequestArguments()
	f, err := os.Open("testdata/file4upload")

	if err != nil {
		t.Fatalf("Open file error: %v", err)
	}

	f2, err := os.Open("testdata/file2")

	if err != nil {
		t.Fatalf("Open file error: %v", err)
	}

	files := []FileField{
		{
			FileName:    "file4upload",
			FieldName:   "file",
			FileContent: f,
		},
		{
			FileName:    "file2",
			FieldName:   "file2",
			FileContent: f2,
		},
	}

	args.Files = files
	resp, err := Post("http://www.httpbin.org/post", args)
	if err != nil {
		t.Errorf("Reqeust error: %v", err)
	}
	fString1 := "Hey, I am test file."
	fString2 := "Hey, I am test file too."
	JSON, err := resp.JSON()
	if err != nil {
		t.Fatalf("Get response JSON error: %v", err)
	}

	file1, _ := JSON.Get("files").Get("file").String()
	file2, _ := JSON.Get("files").Get("file2").String()

	assert.Equal(t, fString1, file1)
	assert.Equal(t, fString2, file2)
}
func TestBaseGet(t *testing.T) {
	resp, err := Get("http://httpbin.org/get", nil)

	if err != nil {
		t.Errorf("Get error: excepted no error, got %v", err)
	}

	assert.Equal(t, true, resp.Ok())
}

func TestCookieGet(t *testing.T) {
	args := NewRequestArguments()
	args.Cookies = testMap

	resp, err := Get("http://httpbin.org/cookies", args)

	if err != nil {
		t.Fatalf("Request error: %v", err)
	}

	JSON, err := resp.JSON()
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
	args := NewRequestArguments()
	args.JSON = jsonStruct

	resp, err := Post("http://httpbin.org/post", args)
	if err != nil {
		t.Errorf("Reqeust erro: got %v", err)
	}

	JSON, err := resp.JSON()
	if err != nil {
		t.Errorf("%v \n", err)
	}
	name, _ := JSON.Get("json").Get("name").String()
	age, _ := JSON.Get("json").Get("age").Int()

	assert.Equal(t, jsonStruct.Name, name)
	assert.Equal(t, jsonStruct.Age, age)
}

func TestBaseDataPost(t *testing.T) {
	args := NewRequestArguments()
	args.Data = testMap

	resp, err := Post("http://httpbin.org/post", args)
	if err != nil {
		t.Errorf("Request error: got %v", err)
	}

	JSON, err := resp.JSON()
	if err != nil {
		t.Errorf("Get json error: %v", err)
	}

	name, _ := JSON.Get("form").Get("name").String()
	age, _ := JSON.Get("form").Get("age").String()

	assert.Equal(t, testMap["name"], name)
	assert.Equal(t, testMap["age"], age)
}

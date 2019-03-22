package requests4go

import (
	"testing"
)

var testMap = map[string]string{
	"name": "name",
	"age":  "age",
}

func TestBaseGet(t *testing.T) {
	resp, err := Get("http://httpbin.org/get", nil)

	if err != nil {
		t.Errorf("Get error: excepted no error, got %v", err)
	}

	if !resp.Ok() {
		t.Errorf("Request got wrong response, excepted \"200 OK\", got \"%v\"", resp.Status)
	}
}

func TestCookieGet(t *testing.T) {
	args := DefaultRequestArguments
	args.Cookies = testMap

	resp, err := Get("http://httpbin.org/cookies", args)

	if err != nil {
		t.Fatalf("Request error: %v", err)
	}

	JSON, err := resp.Json()
	if err != nil {
		t.Fatalf("Get json error: %v", err)
	}

	if name, _ := JSON.Get("cookies").Get("name").String(); name != testMap["name"] {
		t.Errorf("Cookies set error: excepted %v, got %v", testMap["name"], name)
	}

	if age, _ := JSON.Get("cookies").Get("age").String(); age != testMap["age"] {
		t.Errorf("Cookies set error: excepted %v, got %v", testMap["age"], age)
	}
}

package requests4go

import (
	"net/http"
	"testing"
)

func TestNewRequestArguments(t *testing.T) {
	a := struct {
		a int
		b int
		c int
	}{
		1, 2, 3,
	}
	testArgs := map[string]interface{}{
		"Client":      http.DefaultClient,
		"Header":      defaultHeaders,
		"ObjectParam": a,
	}

	args := NewRequestArguments(testArgs)

	if args.Client != http.DefaultClient {
		t.Error("NewRequestArguments Error")
	}

	if args.ObjectParam != a {
		t.Error("NewRequestArguments Error")
	}

	for k, v := range args.Headers {
		if defaultHeaders[k] != v {
			t.Error("NewRequestArgument Headers Error")
		}
	}
}

func TestPrepareURl(t *testing.T) {
	reqUrl, err := prepareURL("https://www.example.com/", map[string]string{"a": "1", "b": "2"})

	if err != nil {
		t.Errorf("Url error: got %s", err)
	}

	if reqUrl != "https://www.example.com/?a=1&b=2" {
		t.Errorf("Prepare url error: excepted %s, got %s", "https://www.example.com/?a=1&b=2", reqUrl)
	}

	reqUrl3, err := prepareURL("https://www.example.com/?c=3", map[string]string{"a": "1", "b": "2"})

	if err != nil {
		t.Errorf("Url error: got %s", err)
	}

	if reqUrl3 != "https://www.example.com/?a=1&b=2&c=3" {
		t.Errorf("Prepare url error: excepted %s, got %s", "https://www.example.com/?a=1&b=2&c=3", reqUrl3)
	}
}

func TestPrepareUrlWithStruct(t *testing.T) {
	type Options struct {
		A string `url:"a"`
		B string `url:"b"`
	}
	opt := Options{"1", "2"}
	reqUrl, err := prepareURLWithStruct("https://www.example.com/", opt)
	if err != nil {
		t.Errorf("Url error: got %s", err)
	}

	target := "https://www.example.com/?a=1&b=2"
	if reqUrl != target {
		t.Errorf("Prepare url error: excepted %s, got %s", target, reqUrl)
	}

}

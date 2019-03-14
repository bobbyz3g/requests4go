package requests4go

import (
	"testing"
)

func TestBaseGet(t *testing.T) {
	resp, err := Get("http://httpbin.org/get", nil)

	if err != nil {
		t.Errorf("Get error: excepted no error, got %v", err)
	}

	if !resp.Ok() {
		t.Errorf("Request got wrong response, excepted \"200 OK\", got \"%v\"", resp.Status)
	}
}

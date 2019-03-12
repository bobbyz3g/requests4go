package requests4go

import "testing"

func TestBaseGet(t *testing.T) {
	resp, err := Get("http://httpbin.org/get", nil)
	if err != nil {
		t.Error(err)
	}
	if resp.Status() != "200 OK" {
		t.Error("Get error.")
	}
}

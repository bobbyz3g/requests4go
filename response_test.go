package requests4go

import "testing"

func TestResponse(t *testing.T) {
	args := DefaultRequestArguments
	args.Params = map[string]string{
		"a": "1",
		"b": "2",
	}
	resp, err := Get("http://httpbin.org/get", args)

	if err != nil {
		t.Errorf("GET request error: got %s", err)
	}

	if flag := resp.Ok(); !flag {
		t.Errorf("GET request status error: excepted %v, got %v", true, flag)
	}

	text, err := resp.Text()

	if err != nil {
		t.Errorf("Response.Text() error: got %s", err)
	}

	content, err := resp.Content()

	if err != nil {
		t.Errorf("Response.Content() error: got %s", err)
	}

	if string(content[:]) != text {
		t.Errorf("Internal content error: \n text is %v, \n string of content is %v", text, content)
	}

	json, err := resp.Json()

	if err != nil {
		t.Errorf("Response.Json error: %v", err)
	}

	agent, _ := json.Get("headers").Get("User-Agent").String()

	if agent != "Request4go" {
		t.Errorf("Response heeder error: excpeted Request4go, go %v", agent)
	}
}

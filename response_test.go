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

import "testing"

func TestResponse(t *testing.T) {
	args := NewRequestArguments()
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

	json, err := resp.SimpleJSON()

	if err != nil {
		t.Errorf("Response.Json error: %v", err)
	}

	agent, _ := json.Get("headers").Get("User-Agent").String()

	if agent != "Request4go" {
		t.Errorf("Response heeder error: excpeted Request4go, go %v", agent)
	}
}

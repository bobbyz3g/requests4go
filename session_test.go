// Copyright 2019 The Kaiser925. All rights reserved.
// Use of this source code is governed by a Apache
// license that can be found in the LICENSE file.

package requests4go

import "testing"

func TestBaseSession(t *testing.T) {
	s := NewSession(nil)
	s.Get("http://httpbin.org/cookies/set/sessioncookie/123456789", nil)
	resp, _ := s.Get("http://httpbin.org/cookies", nil)
	JSON, _ := resp.Json()
	if cookie, _ := JSON.Get("cookies").Get("sessioncookie").String(); cookie != "123456789" {
		t.Errorf("Session cookie error: excepted \"123456789\", got %v", cookie)
	}
}

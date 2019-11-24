// Copyright 2019 The Kaiser925. All rights reserved.
// Use of this source code is governed by a Apache
// license that can be found in the LICENSE file.

package requests4go

import (
	"testing"
)

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

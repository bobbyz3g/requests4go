// Copyright 2019 The Kaiser925. All rights reserved.
// Use of this source code is governed by a Apache
// license that can be found in the LICENSE file.

package requests4go

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareURl(t *testing.T) {
	reqURL, err := prepareURL("https://www.example.com/", map[string]string{"a": "1", "b": "2"})

	if err != nil {
		t.Errorf("Url error: got %s", err)
	}

	assert.Equal(t, "https://www.example.com/?a=1&b=2", reqURL)

	reqURL3, err := prepareURL("https://www.example.com/?c=3", map[string]string{"a": "1", "b": "2"})

	if err != nil {
		t.Errorf("Url error: got %s", err)
	}

	assert.Equal(t, "https://www.example.com/?a=1&b=2&c=3", reqURL3)
}

func TestPrepareUrlWithStruct(t *testing.T) {
	type Options struct {
		A string `url:"a"`
		B string `url:"b"`
	}
	opt := Options{"1", "2"}
	reqURL, err := prepareURLWithStruct("https://www.example.com/", opt)
	if err != nil {
		t.Errorf("Url error: got %s", err)
	}

	target := "https://www.example.com/?a=1&b=2"

	assert.Equal(t, target, reqURL)
}

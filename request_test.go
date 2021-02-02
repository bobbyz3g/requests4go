// Developed by Kaiser925 on 2021/2/2.
// Lasted modified 2021/1/27.
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

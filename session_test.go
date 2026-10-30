// Copyright 2019 The Kaiser925. All rights reserved.
// Use of this source code is governed by a Apache
// license that can be found in the LICENSE file.

package requests4go

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionSharesCookiesAcrossEntrypoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "active", Path: "/"})
			return
		}
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != "active" {
			t.Errorf("session cookie missing: %v", err)
		}
	}))
	defer server.Close()
	session := NewSession()
	resp, err := session.Get(server.URL + "/login")
	if err != nil {
		t.Fatal(err)
	}
	resp.Close()
	resp, err = session.Request(t.Context(), http.MethodGet, server.URL+"/check")
	if err != nil {
		t.Fatal(err)
	}
	resp.Close()
	req, err := NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/check")
	if err != nil {
		t.Fatal(err)
	}
	resp, err = session.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Close()
}

func TestSessionUsesReplacedClient(t *testing.T) {
	session := NewSession()
	session.Client = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusAccepted, Body: http.NoBody, Header: make(http.Header)}, nil
	})}
	resp, err := session.Get("http://example.test")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status %d, want 202", resp.StatusCode)
	}
}

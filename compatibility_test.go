package requests4go_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	requests "github.com/Kaiser925/requests4go"
)

func TestLegacySessionAndOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				t.Error(err)
			}
			if r.PostForm.Get("first") != "one" || r.PostForm.Get("second") != "two" || r.URL.Query().Get("page") != "1" {
				t.Errorf("legacy request form=%v query=%v", r.PostForm, r.URL.Query())
			}
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "legacy"})
		} else if cookie, err := r.Cookie("session"); err != nil || cookie.Value != "legacy" {
			t.Errorf("legacy session cookie: %v", err)
		}
		io.WriteString(w, "response")
	}))
	defer server.Close()
	session := requests.NewSession()
	resp, err := session.Post(server.URL, requests.Params(requests.M{"page": "1"}),
		requests.Data(requests.M{"first": "one"}), requests.Data(requests.M{"second": "two"}))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if text, err := resp.Text(); err != nil || text != "response" {
		t.Fatalf("text=%q err=%v", text, err)
	}
	if content, err := resp.Content(); err != nil || len(content) != 0 {
		t.Fatal("legacy response became cached")
	}
	resp2, err := session.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Close()
}

func ExampleClient_NewRequest() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"name":"Alice"}`)
	}))
	defer server.Close()
	client := requests.NewClient(server.Client())
	value, err := client.NewRequest(context.Background()).Get(server.URL).Send().
		Require2xx().LimitBody(1024).DecodeJSON[struct{ Name string }]()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(value.Name)
	// Output: Alice
}

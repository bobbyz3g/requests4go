package requests4go

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestBuilderSnapshotsAndRedirect(t *testing.T) {
	query := url.Values{"tag": {"go", "http"}}
	headers := http.Header{"X-Value": {"one", "two"}}
	form := url.Values{"item": {"a+b", "中文"}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		wantQuery := url.Values{"keep": {"yes"}, "tag": {"go", "http"}}
		if !reflect.DeepEqual(r.URL.Query(), wantQuery) {
			t.Errorf("query = %v, want %v", r.URL.Query(), wantQuery)
		}
		if !reflect.DeepEqual(r.Header.Values("X-Value"), []string{"one", "two"}) {
			t.Errorf("headers = %v", r.Header)
		}
		if err := r.ParseForm(); err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(r.PostForm, url.Values{"item": {"a+b", "中文"}}) {
			t.Errorf("form = %v", r.PostForm)
		}
		if r.URL.Path == "/start" {
			w.Header().Set("Location", "/finish?keep=yes&tag=go&tag=http")
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}))
	defer server.Close()
	b := NewClient(server.Client()).NewRequest(t.Context()).
		Post(server.URL + "/start?keep=yes&tag=old").Query(query).Headers(headers).Form(form)
	query["tag"][0] = "changed"
	headers["X-Value"][0] = "changed"
	form["item"][0] = "changed"
	resp, err := b.Send().Response()
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if resp.StatusCode != http.StatusOK || resp.Request.URL.Path != "/finish" {
		t.Fatal("request body did not replay through redirect")
	}
}

func TestBuilderJSONSnapshotAndContext(t *testing.T) {
	ctx := context.WithValue(t.Context(), struct{}{}, "value")
	client := NewClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		defer req.Body.Close()
		data, err := io.ReadAll(req.Body)
		if err != nil || string(data) != `{"name":"Alice"}` {
			t.Errorf("body = %q, error %v", data, err)
		}
		if req.Context().Value(struct{}{}) != "value" || req.Header.Get("Content-Type") != AppJSON {
			t.Error("context or content type lost")
		}
		return &http.Response{StatusCode: 422, Body: http.NoBody}, nil
	})})
	payload := M{"name": "Alice"}
	b := client.NewRequest(ctx).Post("http://example.test").JSON(payload)
	payload["name"] = "changed"
	resp, err := b.Send().Response()
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if resp.StatusCode != 422 {
		t.Fatal("HTTP error response should remain available")
	}
}

type failingJSON struct {
	calls *int
	err   error
}

func (v failingJSON) MarshalJSON() ([]byte, error) {
	*v.calls++
	return nil, v.err
}

func TestBuilderFirstErrorShortCircuits(t *testing.T) {
	sends, encodes := 0, 0
	want := errors.New("encode failed")
	client := NewClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		sends++
		return nil, errors.New("unexpected transport")
	})})
	skipped := &countedBody{Reader: strings.NewReader("unused")}
	b := client.NewRequest(t.Context()).Post("http://example.test").
		JSON(failingJSON{&encodes, want}).
		JSON(failingJSON{&encodes, errors.New("later error")}).
		Post("http://%").Body(skipped)
	result := b.Send()
	resp, err := result.Response()
	var marshalErr *json.MarshalerError
	if resp != nil || !errors.Is(err, want) || !errors.As(err, &marshalErr) {
		t.Fatalf("response %v, error %v", resp, err)
	}
	if err := result.Close(); !errors.Is(err, want) {
		t.Fatalf("Close lost first error: %v", err)
	}
	if sends != 0 || encodes != 1 || skipped.closes != 0 || skipped.reads != 0 {
		t.Fatalf("sends=%d encodes=%d skipped=%+v", sends, encodes, skipped)
	}
}

type countedBody struct {
	io.Reader
	reads, closes int
	closeErr      error
}

func (b *countedBody) Read(p []byte) (int, error) { b.reads++; return b.Reader.Read(p) }
func (b *countedBody) Close() error               { b.closes++; return b.closeErr }

func TestBuilderBodyConflictOwnership(t *testing.T) {
	first := &countedBody{Reader: strings.NewReader("first"), closeErr: errors.New("close failed")}
	second := &countedBody{Reader: strings.NewReader("second")}
	b := NewClient(nil).NewRequest(t.Context()).Post("http://example.test").Body(first).Body(second)
	if first.closes != 0 || second.closes != 0 {
		t.Fatal("configuration closed a reader")
	}
	_, err := b.Send().Response()
	if !errors.Is(err, ErrBodyAlreadySet) || first.closes != 1 || second.closes != 0 || first.reads != 0 {
		t.Fatalf("error=%v first=%+v second=%+v", err, first, second)
	}
	if _, err := b.Send().Response(); !errors.Is(err, ErrRequestSent) || first.closes != 1 {
		t.Fatalf("repeated Send: %v, closes=%d", err, first.closes)
	}
}

func TestBuilderValidationFailureClosesAcceptedBody(t *testing.T) {
	body := &countedBody{Reader: strings.NewReader("body")}
	_, err := NewClient(nil).NewRequest(t.Context()).Body(body).Post("http://%").Send().Response()
	if err == nil || body.closes != 1 || body.reads != 0 {
		t.Fatalf("error=%v body=%+v", err, body)
	}
	for name, builder := range map[string]*RequestBuilder{
		"nil context": NewClient(nil).NewRequest(nil).Get("http://example.test"),
		"no URL":      NewClient(nil).NewRequest(t.Context()),
		"zero value":  {},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := builder.Send().Response(); err == nil {
				t.Fatal("invalid builder succeeded")
			}
		})
	}
}

func TestBuilderSendOnceAndTransfer(t *testing.T) {
	sends := 0
	body := &countedBody{Reader: strings.NewReader("response")}
	client := NewClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		sends++
		return &http.Response{StatusCode: 200, Body: body}, nil
	})})
	b := client.NewRequest(t.Context()).Get("http://example.test")
	result := b.Send()
	if _, err := b.Post("http://other.test").Send().Response(); !errors.Is(err, ErrRequestSent) {
		t.Fatalf("repeated Send: %v", err)
	}
	resp, err := result.Response()
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if _, err := result.Response(); !errors.Is(err, ErrResultConsumed) {
		t.Fatalf("repeated transfer: %v", err)
	}
	if err := result.Close(); err != nil || body.closes != 0 || sends != 1 {
		t.Fatalf("transfer ownership: error=%v closes=%d sends=%d", err, body.closes, sends)
	}
	if text, err := resp.Text(); err != nil || text != "response" {
		t.Fatalf("caller cannot read response: %q, %v", text, err)
	}
}

func TestResultCloseOnce(t *testing.T) {
	want := errors.New("close failed")
	body := &countedBody{Reader: strings.NewReader("unread"), closeErr: want}
	client := NewClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: body}, nil
	})})
	result := client.NewRequest(t.Context()).Get("http://example.test").Send()
	for range 2 {
		if err := result.Close(); !errors.Is(err, want) {
			t.Fatalf("Close error = %v", err)
		}
	}
	if body.closes != 1 || body.reads != 0 {
		t.Fatalf("body = %+v", body)
	}
	if _, err := result.Response(); !errors.Is(err, want) {
		t.Fatalf("cleanup error lost: %v", err)
	}
}

func TestBuilderSendFailure(t *testing.T) {
	want := errors.New("transport failed")
	body := &countedBody{Reader: strings.NewReader("body")}
	client := NewClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.Body.Close()
		return nil, want
	})})
	result := client.NewRequest(t.Context()).Post("http://example.test").Body(body).Send()
	if _, err := result.Response(); !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
	if err := result.Close(); !errors.Is(err, want) || body.closes != 1 {
		t.Fatalf("error=%v closes=%d", err, body.closes)
	}
}

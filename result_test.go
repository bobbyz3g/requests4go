package requests4go

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newReadResult(payload string, status int) (*Result, *countedBody) {
	body := &countedBody{Reader: strings.NewReader(payload)}
	return &Result{response: NewResponse(&http.Response{
		StatusCode: status, Status: http.StatusText(status),
		Header: http.Header{"Content-Type": {"application/json; charset=utf-8"}}, Body: body,
	})}, body
}

func TestResultFluentJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, `{"name":"Alice"}`)
	}))
	defer server.Close()
	type user struct{ Name string }
	value, err := NewClient(server.Client()).NewRequest(t.Context()).
		Post(server.URL).JSON(M{"name": "Alice"}).Send().
		Require2xx().RequireStatus(http.StatusCreated).LimitBody(1024).DecodeJSON[user]()
	if err != nil || value.Name != "Alice" {
		t.Fatalf("value=%+v error=%v", value, err)
	}
}

func TestResultStatusFailureShortCircuits(t *testing.T) {
	r, body := newReadResult(`{"name":"ignored"}`, http.StatusNotFound)
	body.closeErr = errors.New("close failed")
	header := r.response.Header
	r.Require2xx()
	if body.closes != 1 || body.reads != 0 {
		t.Fatalf("status rejection: %+v", body)
	}
	header.Set("Content-Type", "changed")
	var value map[string]any
	err := r.RequireStatus(200).LimitBody(-1).DecodeJSONInto(&value)
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != 404 || statusErr.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("status error = %v", err)
	}
	if value != nil || body.reads != 0 || body.closes != 1 {
		t.Fatal("operation after status failure touched the body")
	}
	if r.Close() != err {
		t.Fatal("cleanup replaced first error")
	}
}

func TestResultTerminalsConsumeOnce(t *testing.T) {
	for _, terminal := range []struct {
		name string
		run  func(*Result) error
	}{
		{"bytes", func(r *Result) error { _, err := r.Bytes(); return err }},
		{"text", func(r *Result) error { _, err := r.Text(); return err }},
		{"json", func(r *Result) error { _, err := r.DecodeJSON[map[string]int](); return err }},
		{"into", func(r *Result) error { var v map[string]int; return r.DecodeJSONInto(&v) }},
		{"discard", func(r *Result) error { return r.Discard() }},
	} {
		t.Run(terminal.name, func(t *testing.T) {
			r, body := newReadResult(`{"n":1}`, 200)
			if err := terminal.run(r); err != nil {
				t.Fatal(err)
			}
			reads := body.reads
			if body.closes != 1 || (terminal.name == "discard" && reads != 0) {
				t.Fatalf("body=%+v", body)
			}
			if err := terminal.run(r); !errors.Is(err, ErrResultConsumed) {
				t.Fatalf("repeat terminal: %v", err)
			}
			if _, err := r.Response(); !errors.Is(err, ErrResultConsumed) {
				t.Fatalf("transfer after consumption: %v", err)
			}
			r.Close()
			if body.closes != 1 || body.reads != reads {
				t.Fatal("body touched after consumption")
			}
		})
	}
}

func TestResultLimitBody(t *testing.T) {
	for _, tc := range []struct {
		name, payload string
		limit         int64
		wantErr       bool
	}{
		{"empty", "", 0, false},
		{"zero", "x", 0, true},
		{"exact", "abc", 3, false},
		{"below", "ab", 3, false},
		{"above", "abcd", 3, true},
		{"max", "abc", math.MaxInt64, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, body := newReadResult(tc.payload, 200)
			data, err := r.LimitBody(tc.limit).Bytes()
			if tc.wantErr {
				if !errors.Is(err, ErrResponseTooLarge) || data != nil {
					t.Fatalf("data=%q err=%v", data, err)
				}
			} else if err != nil || string(data) != tc.payload {
				t.Fatalf("data=%q err=%v", data, err)
			}
			if body.closes != 1 {
				t.Fatalf("closes=%d", body.closes)
			}
		})
	}
	r, body := newReadResult("unused", 200)
	if _, err := r.LimitBody(-1).Text(); err == nil || body.reads != 0 || body.closes != 1 {
		t.Fatalf("negative limit: error=%v body=%+v", err, body)
	}
	r, body = newReadResult("", 200)
	source := strings.NewReader(strings.Repeat("x", 4096))
	body.Reader = source
	if _, err := r.LimitBody(3).Bytes(); !errors.Is(err, ErrResponseTooLarge) || 4096-source.Len() != 4 {
		t.Fatalf("limit did not bound reads: error=%v bytes read=%d", err, 4096-source.Len())
	}
}

func TestResultStatusChecksAreExplicit(t *testing.T) {
	r, body := newReadResult("server error", 500)
	if text, err := r.Text(); err != nil || text != "server error" || body.closes != 1 {
		t.Fatalf("implicit status rejection: text=%q err=%v", text, err)
	}
	r, body = newReadResult("created", 201)
	err := r.Require2xx().RequireStatus(200).Discard()
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != 201 || body.reads != 0 || body.closes != 1 {
		t.Fatalf("exact status check: err=%v body=%+v", err, body)
	}
}

func TestResultPreservesConstructionError(t *testing.T) {
	want := errors.New("encoding failed")
	calls := 0
	value, err := NewClient(nil).NewRequest(t.Context()).Post("http://example.test").
		JSON(failingJSON{&calls, want}).Send().Require2xx().LimitBody(1).DecodeJSON[map[string]int]()
	if value != nil || !errors.Is(err, want) || calls != 1 {
		t.Fatalf("value=%v err=%v calls=%d", value, err, calls)
	}
}

type readError struct{ err error }

func (r readError) Read([]byte) (int, error) { return 0, r.err }

func TestResultReadAndCloseErrors(t *testing.T) {
	readErr, closeErr := errors.New("read failed"), errors.New("close failed")
	for _, limit := range []bool{false, true} {
		r, body := newReadResult("", 200)
		body.Reader, body.closeErr = readError{readErr}, closeErr
		if limit {
			r.LimitBody(1)
		}
		if data, err := r.Bytes(); data != nil || !errors.Is(err, readErr) || body.closes != 1 {
			t.Fatalf("data=%q err=%v closes=%d", data, err, body.closes)
		}
	}
	r, body := newReadResult(`{"n":1}`, 200)
	body.closeErr = closeErr
	value, err := r.DecodeJSON[map[string]int]()
	if value != nil || !errors.Is(err, closeErr) || body.closes != 1 {
		t.Fatalf("value=%v error=%v closes=%d", value, err, body.closes)
	}
}

func TestResultJSONFailuresClose(t *testing.T) {
	for _, payload := range []string{`{"n":`, `{"n":1} {"n":2}`, `{"n":"wrong type"}`} {
		r, body := newReadResult(payload, 200)
		body.closeErr = errors.New("close failed")
		value, err := r.DecodeJSON[map[string]int]()
		if err == nil || errors.Is(err, body.closeErr) || value != nil || body.closes != 1 {
			t.Fatalf("payload=%q value=%v error=%v closes=%d", payload, value, err, body.closes)
		}
	}
	r, body := newReadResult(`{"n":1}`, 200)
	r.response.Header.Set("Content-Type", "text/plain")
	if err := r.DecodeJSONInto(new(any)); !errors.Is(err, ErrNotJSONContent) || body.reads != 0 || body.closes != 1 {
		t.Fatalf("media type: err=%v body=%+v", err, body)
	}
	r, body = newReadResult(`{"n":1}`, 200)
	if _, err := r.LimitBody(3).DecodeJSON[map[string]int](); !errors.Is(err, ErrResponseTooLarge) || body.closes != 1 {
		t.Fatalf("JSON limit: err=%v body=%+v", err, body)
	}
	r, body = newReadResult(`{"n":1}`, 200)
	var invalid *json.InvalidUnmarshalError
	if err := r.DecodeJSONInto(nil); !errors.As(err, &invalid) || body.closes != 1 {
		t.Fatalf("invalid destination: err=%v body=%+v", err, body)
	}
}

func TestResultTransferredResponseIsUntouched(t *testing.T) {
	r, body := newReadResult("caller owned", 200)
	resp, err := r.LimitBody(0).Response()
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if _, err := r.Require2xx().Text(); !errors.Is(err, ErrResultConsumed) || body.closes != 0 || body.reads != 0 {
		t.Fatalf("err=%v body=%+v", err, body)
	}
	text, err := resp.Text()
	if err != nil || text != "caller owned" {
		t.Fatalf("limit affected transferred body: %q, %v", text, err)
	}
}

func TestResultCanceledReadCloses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		<-req.Context().Done()
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	r := NewClient(server.Client()).NewRequest(ctx).Get(server.URL).Send()
	if r.err != nil {
		t.Fatal(r.err)
	}
	body := &countedBody{Reader: r.response.Body}
	original := r.response.Body
	r.response.Body = &readCloseTracker{countedBody: body, original: original}
	cancel()
	if _, err := r.Text(); !errors.Is(err, context.Canceled) || body.closes != 1 {
		t.Fatalf("error=%v closes=%d", err, body.closes)
	}
}

type readCloseTracker struct {
	*countedBody
	original io.Closer
}

func (b *readCloseTracker) Close() error { b.closes++; return b.original.Close() }

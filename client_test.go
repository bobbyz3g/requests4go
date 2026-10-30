package requests4go

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestClientUsesTransportAndPreservesResponse(t *testing.T) {
	ctx := context.WithValue(t.Context(), struct{}{}, "request value")
	body := &observedBody{Reader: strings.NewReader("error details")}
	raw := &http.Response{StatusCode: http.StatusTeapot, Body: body, Header: make(http.Header)}
	client := NewClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Context().Value(struct{}{}) != "request value" {
			t.Error("request context was lost")
		}
		payload, err := io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil || string(payload) != `{"name":"Alice"}` || req.Header.Get("X-Test") != "yes" {
			t.Errorf("unexpected request: payload=%q header=%q err=%v", payload, req.Header.Get("X-Test"), err)
		}
		return raw, nil
	})})
	resp, err := client.Request(ctx, http.MethodPost, "http://example.test", JSON(M{"name": "Alice"}), Headers(M{"X-Test": "yes"}))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if resp.Response != raw || body.closed || body.reads != 0 {
		t.Fatal("response was replaced, read or closed before returning to the caller")
	}
	text, err := resp.Text()
	if err != nil || text != "error details" || body.closed {
		t.Fatalf("Text = %q, error %v, closed %v", text, err, body.closed)
	}
}

type observedBody struct {
	io.Reader
	reads  int
	closed bool
}

func (b *observedBody) Read(p []byte) (int, error) { b.reads++; return b.Reader.Read(p) }
func (b *observedBody) Close() error               { b.closed = true; return nil }

func TestClientConstructionFailureDoesNotSend(t *testing.T) {
	client := NewClient(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Error("transport called after construction failure")
		return nil, errors.New("unexpected send")
	})})
	want := errors.New("invalid option")
	resp, err := client.Request(t.Context(), http.MethodGet, "http://example.test", func(*http.Request) error { return want })
	if resp != nil || !errors.Is(err, want) {
		t.Fatalf("response %v, error %v", resp, err)
	}
	if _, err := client.Request(nil, http.MethodGet, "http://example.test"); err == nil {
		t.Fatal("nil context accepted")
	}
}

func TestClientCancellationAndTimeout(t *testing.T) {
	for _, timeout := range []bool{false, true} {
		t.Run(map[bool]string{false: "cancellation", true: "timeout"}[timeout], func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			hc := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if !timeout {
					cancel()
				}
				<-req.Context().Done()
				return nil, req.Context().Err()
			})}
			want := context.Canceled
			if timeout {
				hc.Timeout = 10 * time.Millisecond
				want = context.DeadlineExceeded
			}
			_, err := NewClient(hc).Request(ctx, http.MethodGet, "http://example.test")
			var urlErr *url.Error
			if !errors.Is(err, want) || !errors.As(err, &urlErr) {
				t.Fatalf("error %v does not preserve %v and *url.Error", err, want)
			}
		})
	}
}

func TestClientContextCancelsResponseRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	resp, err := NewClient(nil).Request(ctx, http.MethodGet, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	cancel()
	if _, err := resp.Content(); !errors.Is(err, context.Canceled) {
		t.Fatalf("body read error = %v, want context.Canceled", err)
	}
}

func TestClientRedirectPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/start" {
			t.Error("redirect policy was ignored")
		}
		w.Header().Set("Location", "/finish")
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()
	client := NewClient(&http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}})
	req, err := NewRequestWithContext(t.Context(), http.MethodGet, server.URL+"/start")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("status %d, want 302", resp.StatusCode)
	}
}

func TestClientZeroValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	var client Client
	resp, err := client.Request(t.Context(), http.MethodGet, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d, want 204", resp.StatusCode)
	}
}

func TestClientConcurrentRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, r.Header.Get("X-Request"))
	}))
	defer server.Close()
	client := NewClient(server.Client())
	var workers sync.WaitGroup
	for _, value := range []string{"one", "two", "three", "four"} {
		workers.Go(func() {
			resp, err := client.Request(t.Context(), http.MethodGet, server.URL, Headers(M{"X-Request": value}))
			if err != nil {
				t.Error(err)
				return
			}
			defer resp.Close()
			got, err := resp.Text()
			if err != nil || got != value {
				t.Errorf("response %q, error %v; want %q", got, err, value)
			}
		})
	}
	workers.Wait()
}

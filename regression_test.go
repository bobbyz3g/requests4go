package requests4go

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestDataSendsMergedForm(t *testing.T) {
	want := url.Values{"keep": {"first"}, "replace": {"last"}, "escaped": {"a+b &中文"}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		if !reflect.DeepEqual(r.PostForm, want) {
			t.Errorf("received form %v, want %v", r.PostForm, want)
		}
		if r.URL.Path == "/start" {
			w.Header().Set("Location", "/finish")
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	}))
	defer server.Close()

	resp, err := Post(server.URL+"/start",
		Data(M{"keep": "first", "replace": "old"}),
		Data(M{"replace": "last", "escaped": "a+b &中文"}))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if resp.StatusCode != http.StatusOK || resp.Request.URL.Path != "/finish" {
		t.Fatalf("form was not replayed after redirect: %s", resp.Status)
	}
}

func TestSessionHeadSendsHead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("received method %s, want HEAD", r.Method)
		}
		w.Header().Set("X-Result", "head")
	}))
	defer server.Close()
	resp, err := NewSession().Head(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if resp.Header.Get("X-Result") != "head" {
		t.Fatal("missing response header")
	}
}

func TestReplacingBodyRedirects(t *testing.T) {
	for _, status := range []int{http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
		for _, mode := range []string{"replayable", "stream", "nil", "empty"} {
			t.Run(strconv.Itoa(status)+"/"+mode, func(t *testing.T) {
				var body io.Reader
				want := "fresh"
				switch mode {
				case "replayable":
					body = strings.NewReader(want)
				case "stream":
					// Hide the concrete reader type to make this body non-replayable.
					body = struct{ io.Reader }{strings.NewReader(want)}
				case "nil":
					want = ""
				case "empty":
					want = ""
					body = bytes.NewReader(nil)
				}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					got, err := io.ReadAll(r.Body)
					if err != nil || string(got) != want {
						t.Errorf("%s received %q, error %v; want %q", r.URL.Path, got, err, want)
					}
					if mode == "stream" && r.URL.Path != "/start" {
						t.Error("non-replayable body followed redirect")
					}
					if r.URL.Path == "/start" {
						w.Header().Set("Location", "/finish")
						w.WriteHeader(status)
					}
				}))
				defer server.Close()
				// The old JSON body has the same length as "fresh", so stale GetBody
				// would otherwise replay it without a length error masking the bug.
				resp, err := Post(server.URL+"/start", JSON("old"), Body(body))
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Close()
				wantStatus := http.StatusOK
				if mode == "stream" {
					wantStatus = status
				}
				if resp.StatusCode != wantStatus {
					t.Errorf("status %d, want %d", resp.StatusCode, wantStatus)
				}
			})
		}
	}
}

func TestReplacingBodyClearsLength(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, err := io.ReadAll(r.Body)
		if err != nil || string(got) != "a longer replacement" {
			t.Errorf("received %q, error %v", got, err)
		}
	}))
	defer server.Close()
	body := struct{ io.Reader }{strings.NewReader("a longer replacement")}
	resp, err := Post(server.URL, JSON("old"), Body(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
}

type closeProbe struct {
	reads  int
	closed bool
	err    error
}

func (b *closeProbe) Read([]byte) (int, error) {
	b.reads++
	return 0, io.ErrUnexpectedEOF
}

func (b *closeProbe) Close() error {
	b.closed = true
	return b.err
}

func TestResponseCloseDoesNotRead(t *testing.T) {
	closeErr := errors.New("close failed")
	body := &closeProbe{err: closeErr}
	resp := NewResponse(&http.Response{Body: body})
	if err := resp.Close(); !errors.Is(err, closeErr) {
		t.Errorf("Close error %v, want %v", err, closeErr)
	}
	if body.reads != 0 || !body.closed {
		t.Errorf("Close read %d times; closed = %v", body.reads, body.closed)
	}
	if err := NewResponse(&http.Response{}).Close(); err != nil {
		t.Fatalf("closing absent body: %v", err)
	}
}

func TestResponseJSONMediaType(t *testing.T) {
	for _, tc := range []struct {
		contentType string
		accepted    bool
	}{
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"Application/JSON; Charset=\"UTF-8\"", true},
		{"", false},
		{"text/plain", false},
		{"application/json; charset=", false},
	} {
		t.Run(tc.contentType, func(t *testing.T) {
			const payload = `{"name":"Alice"}`
			resp := NewResponse(&http.Response{
				Header: http.Header{"Content-Type": {tc.contentType}},
				Body:   io.NopCloser(strings.NewReader(payload)),
			})
			defer resp.Close()
			var got struct{ Name string }
			err := resp.JSON(&got)
			if tc.accepted {
				if err != nil || got.Name != "Alice" {
					t.Fatalf("decoded %+v, error %v", got, err)
				}
			} else {
				if !errors.Is(err, ErrNotJSONContent) {
					t.Fatalf("error %v, want ErrNotJSONContent", err)
				}
				remaining, err := resp.Text()
				if err != nil || remaining != payload {
					t.Fatalf("rejected content was consumed: %q, error %v", remaining, err)
				}
			}
		})
	}
}

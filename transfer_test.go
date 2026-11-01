package requests4go

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMultipartStreamsRepeatedFiles(t *testing.T) {
	first := &countedBody{Reader: strings.NewReader("first")}
	second := &countedBody{Reader: strings.NewReader("second")}
	parts := []Part{
		{Name: "title", Body: strings.NewReader("hello")},
		{Name: "files", Filename: "one.txt", ContentType: "text/plain", Body: first},
		{Name: "files", Filename: "two.txt", Body: second},
	}
	client := NewClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		defer req.Body.Close()
		if req.GetBody != nil || req.ContentLength > 0 {
			t.Error("multipart stream should have unknown length and no replay")
		}
		_, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		if err != nil {
			t.Fatal(err)
		}
		reader := multipart.NewReader(req.Body, params["boundary"])
		for i, expected := range []struct{ name, filename, contentType, data string }{
			{"title", "", "", "hello"},
			{"files", "one.txt", "text/plain", "first"},
			{"files", "two.txt", "application/octet-stream", "second"},
		} {
			part, err := reader.NextPart()
			if err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(part)
			if err != nil || part.FormName() != expected.name || part.FileName() != expected.filename || part.Header.Get("Content-Type") != expected.contentType || string(data) != expected.data {
				t.Fatalf("part %d: header=%v data=%q err=%v", i, part.Header, data, err)
			}
		}
		if _, err := reader.NextPart(); err != io.EOF {
			t.Fatalf("missing final boundary: %v", err)
		}
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	})})
	b := client.NewRequest(t.Context()).Post("http://example.test").Multipart(parts...)
	parts[1].Filename = "changed"
	if first.reads != 0 || second.reads != 0 || first.closes != 0 {
		t.Fatal("multipart construction touched payloads")
	}
	if err := b.Send().Discard(); err != nil {
		t.Fatal(err)
	}
	if first.closes != 1 || second.closes != 1 {
		t.Fatal("multipart sources were not closed exactly once")
	}
}

func TestMultipartValidationClosesAcceptedParts(t *testing.T) {
	body := &countedBody{Reader: strings.NewReader("unread")}
	r := NewClient(nil).NewRequest(t.Context()).Post("http://example.test").Multipart(
		Part{Name: "file", Filename: "file", ContentType: "text/plain; charset=", Body: body},
	).Send()
	if r.Close() == nil || body.reads != 0 || body.closes != 1 {
		t.Fatalf("validation: err=%v body=%+v", r.err, body)
	}
}

func TestMultipartReadError(t *testing.T) {
	want := errors.New("upload failed")
	source := &countedBody{Reader: readError{want}}
	client := NewClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		defer req.Body.Close()
		_, err := io.Copy(io.Discard, req.Body)
		return nil, err
	})})
	err := client.NewRequest(t.Context()).Post("http://example.test").Multipart(Part{Name: "file", Body: source}).Send().Discard()
	if !errors.Is(err, want) || source.closes != 1 {
		t.Fatalf("error=%v closes=%d", err, source.closes)
	}
}

func TestMultipartCancellationUnblocksReader(t *testing.T) {
	source, writer := io.Pipe()
	defer writer.Close()
	started := make(chan struct{})
	finished := make(chan struct{})
	client := NewClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		go func() {
			close(started)
			io.Copy(io.Discard, req.Body)
			close(finished)
		}()
		<-req.Context().Done()
		req.Body.Close()
		<-finished
		return nil, req.Context().Err()
	})})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- client.NewRequest(ctx).Post("http://example.test").Multipart(Part{Name: "stream", Body: source}).Send().Discard()
	}()
	<-started
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation error=%v", err)
	}
}

func TestFileOpensAtSendAndDoesNotReplay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "upload")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/start" {
			t.Error("file body unexpectedly replayed")
		}
		data, err := io.ReadAll(req.Body)
		if err != nil || string(data) != "created after configuration" {
			t.Errorf("data=%q err=%v", data, err)
		}
		w.Header().Set("Location", "/finish")
		w.WriteHeader(307)
	}))
	defer server.Close()
	b := NewClient(server.Client()).NewRequest(t.Context()).Post(server.URL + "/start").File(path)
	if err := os.WriteFile(path, []byte("created after configuration"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := b.Send().RequireStatus(307).Discard(); err != nil {
		t.Fatal(err)
	}
}

func TestFileErrorPreservesIdentity(t *testing.T) {
	err := NewClient(nil).NewRequest(t.Context()).Post("http://example.test").File(filepath.Join(t.TempDir(), "absent")).Send().Discard()
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("open error=%v", err)
	}
}

func TestSaveStagesFileAndCloses(t *testing.T) {
	for _, mode := range []string{"success", "limit", "read", "close", "status", "rename"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "download")
			if mode == "rename" {
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			r, body := newReadResult("downloaded", 200)
			want := errors.New("failure")
			switch mode {
			case "limit":
				r.LimitBody(3)
				want = ErrResponseTooLarge
			case "read":
				body.Reader = readError{want}
			case "close":
				body.closeErr = want
			case "status":
				r.RequireStatus(201)
			}
			err := r.Save(path)
			if mode == "success" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("expected save failure")
			} else if mode != "status" && mode != "rename" && !errors.Is(err, want) {
				t.Fatalf("error=%v, want %v", err, want)
			}
			if mode != "rename" {
				data, err := os.ReadFile(path)
				wantData := "original"
				if mode == "success" {
					wantData = "downloaded"
				}
				if err != nil || string(data) != wantData {
					t.Fatalf("file=%q error=%v", data, err)
				}
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 1 || body.closes != 1 {
				t.Fatalf("entries=%v error=%v closes=%d", entries, err, body.closes)
			}
			if err := r.Save(filepath.Join(dir, "second")); err == nil {
				t.Fatal("save consumed twice")
			}
		})
	}
}

func TestSaveCreateErrorClosesResponse(t *testing.T) {
	r, body := newReadResult("unread", 200)
	err := r.Save(filepath.Join(t.TempDir(), "absent", "file"))
	if !errors.Is(err, os.ErrNotExist) || body.reads != 0 || body.closes != 1 {
		t.Fatalf("error=%v body=%+v", err, body)
	}
}

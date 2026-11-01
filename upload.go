package requests4go

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"os"
	"sync"
)

// Part is one multipart form field or file. Name identifies the field; Filename
// distinguishes a file from a text field. ContentType is optional (files default
// to application/octet-stream). Body must not be nil. Use separate Parts for
// repeated names. Metadata is copied; Body is consumed once, not buffered.
type Part struct {
	Name        string
	Filename    string
	ContentType string
	Body        io.Reader
}

// File selects a raw file body, opened only when Send starts. The library closes
// it on completion or failure. It streams with unknown length and cannot replay
// for 307/308 redirects. Use a fresh builder for another attempt. No Content-Type
// is inferred. File is a body source and conflicts with other body setters.
func (b *RequestBuilder) File(path string) *RequestBuilder {
	if b.acceptBody() {
		b.openBody = func() (io.ReadCloser, error) {
			file, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			return file, nil
		}
	}
	return b
}

// Multipart selects a streaming multipart/form-data body and snapshots metadata.
// Parts are sent in order, including repeated names. Closers of accepted parts
// transfer to the library at Send, including if construction fails; rejected
// parts remain caller-owned. Each part should have its own reader. A reader that
// may block must support concurrent Close to interrupt Read on cancellation.
// Multipart bodies have unknown length and cannot replay for 307/308 redirects.
func (b *RequestBuilder) Multipart(parts ...Part) *RequestBuilder {
	if !b.acceptBody() {
		return b
	}
	body := &multipartBody{}
	for _, part := range parts {
		if closer, ok := part.Body.(io.Closer); ok {
			body.closers = append(body.closers, closer)
		}
	}
	b.body = body
	var headers bytes.Buffer
	writer := multipart.NewWriter(&headers)
	var readers []io.Reader
	for _, part := range parts {
		if part.Name == "" || part.Body == nil {
			b.err = fmt.Errorf("multipart part requires a name and body")
			return b
		}
		params := map[string]string{"name": part.Name}
		if part.Filename != "" {
			params["filename"] = part.Filename
		}
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", mime.FormatMediaType("form-data", params))
		contentType := part.ContentType
		if contentType == "" && part.Filename != "" {
			contentType = "application/octet-stream"
		}
		if contentType != "" {
			mediaType, params, err := mime.ParseMediaType(contentType)
			if err != nil {
				b.err = fmt.Errorf("multipart content type: %w", err)
				return b
			}
			header.Set("Content-Type", mime.FormatMediaType(mediaType, params))
		}
		if _, err := writer.CreatePart(header); err != nil {
			b.err = err
			return b
		}
		// Buffer only framing; MultiReader pulls payloads from their sources.
		readers = append(readers, bytes.NewReader(bytes.Clone(headers.Bytes())), part.Body)
		headers.Reset()
	}
	if err := writer.Close(); err != nil {
		b.err = err
		return b
	}
	readers = append(readers, bytes.NewReader(bytes.Clone(headers.Bytes())))
	body.Reader = io.MultiReader(readers...)
	b.headers.Set("Content-Type", writer.FormDataContentType())
	return b
}

type multipartBody struct {
	io.Reader
	closers []io.Closer
	once    sync.Once
	err     error
}

func (b *multipartBody) Close() error {
	b.once.Do(func() {
		for _, closer := range b.closers {
			if err := closer.Close(); b.err == nil {
				b.err = err
			}
		}
	})
	return b.err
}

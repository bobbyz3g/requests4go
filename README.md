# request4go

![test](https://github.com/Kaiser925/requests4go/workflows/test/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/Kaiser925/requests4go.svg)](https://pkg.go.dev/github.com/Kaiser925/requests4go)

Go HTTP Requests. ✨🎉✨ Send requests quickly and humanely.

Requires Go 1.27.3 or newer. The module path remains
`github.com/Kaiser925/requests4go`, including when using this fork.

Quick Start
=======

Get module.

~~~
go get -u github.com/Kaiser925/requests4go
~~~

### Make a request

~~~go
package main

import (
	"log"

	"github.com/Kaiser925/requests4go"
)

func main() {
	r, err := requests4go.Get("http://httpbin.org/get")
	if err != nil {
		log.Fatal(err.Error())
	}

	defer r.Close()
	txt, err := r.Text()
	if err != nil {
		log.Print(err)
		return
	}
	log.Println(txt)
}
~~~

You can also send a POST request.

```go
package main

import (
	"log"

	"github.com/Kaiser925/requests4go"
)

func main() {
	// JSON will set body be json data.
	data := requests4go.JSON(requests4go.M{"key": "value"})
	r, err := requests4go.Post("http://httpbin.org/post", data)
	
	if err != nil {
		log.Print(err)
		return
	}
	defer r.Close()
	log.Println(r.Status)
}
```

### Passing Parameters In URLS

The snippets below are intended for functions that return an error, unless a full
package is shown.

```go
params := requests4go.Params(requests4go.M{"key1": "value1", "key2": "value2"})
r, err := requests4go.Get("http://httpbin.org/get", params)
if err != nil {
	return err
}
defer r.Close()
```

### Custom Headers

```go
headers := requests4go.Headers(requests4go.M{"key1": "value1", "key2": "value2"})

r, err := requests4go.Get("http://httpbin.org/get", headers)
if err != nil {
	return err
}
defer r.Close()
```

### Response Content

Read the response body with `Text`, `Content`, or a decoding method. These methods
consume the body without caching or closing it. Always close the response when
finished. `Close` closes the body without reading the remaining content.

~~~go
resp, err := requests4go.Get("https://httpbin.org/get")
if err != nil {
	return err
}
defer resp.Close()
txt, err := resp.Text()
if err != nil {
	return err
}
log.Println(txt)
~~~

### Decode Response Content

There are two methods to handle JSON response content.

1. Decode into a Go value using JSON.

```go
var foo map[string]any
resp, err := requests4go.Get("https://httpbin.org/json")
if err != nil {
	return err
}
defer resp.Close()
if err := resp.JSON(&foo); err != nil {
	return err
}
fmt.Printf("%v\n", foo)
```

2. Implement Unmarshaler on a custom type.

```go
package foo

import (
	"encoding/json"
	"fmt"

	"github.com/Kaiser925/requests4go"
)

type Foo map[string]any

func (f *Foo) Unmarshal(p []byte) error {
	return json.Unmarshal(p, f)
}

func F() error {
	var f Foo
	resp, err := requests4go.Get("https://httpbin.org/json")
	if err != nil {
		return err
	}
	defer resp.Close()
	if err := resp.Unmarshal(&f); err != nil {
		return err
	}
	fmt.Printf("%v\n", f)
	return nil
}
```

### Configured clients and context

`Client` wraps a standard `*http.Client` without copying or modifying it. Configure
timeouts, redirects, proxies, transports and cookie jars through the standard
library client before use:

```go
client := requests4go.NewClient(&http.Client{Timeout: 10 * time.Second})
resp, err := client.Request(ctx, http.MethodGet, "https://httpbin.org/get",
	requests4go.Params(requests4go.M{"page": "1"}))
if err != nil {
	return err
}
defer resp.Close()
```

Here `ctx` is the caller's non-nil `context.Context`; the snippet also uses
`net/http` and `time`. To send an existing `*http.Request`, use `client.Do(req)`.
Its context controls the request and response body reads. Cancellation and timeout
errors retain standard library error identities for `errors.Is` and `errors.As`.

The zero value of `Client` and `NewClient(nil)` use a shared default HTTP client,
with no timeout or cookie jar. This client is independent of `http.DefaultClient`.
Clients can be reused concurrently if their underlying transport supports it;
do not mutate client configuration during use. Each response belongs to its caller.

`NewSession()` still creates a cookie jar. Its existing methods remain available,
and `session.Request(ctx, method, url, opts...)` and `session.Do(req)` use the same
current `session.Client`, including its jar and policies. Replacing that field
between requests remains supported. Setting it to nil selects the shared default
client without cookies.

Existing package-level methods keep their signatures and use the same execution
core. No entrypoint automatically converts non-2xx statuses into errors, reads or
caches a response, or closes a successful response for the caller. Existing request
option order, form merging and body replacement behavior remain unchanged.

### Fluent requests

`Client.NewRequest(ctx)` builds a request through chained setters. `Send` attempts
it once and returns a `Result`; `Response` returns the first error or transfers the
response to the caller:

```go
client := requests4go.NewClient(nil)
resp, err := client.NewRequest(ctx).
	Post("https://httpbin.org/post?source=example").
	Header("Accept", "application/json").
	Query(url.Values{"tag": {"go", "http"}}).
	JSON(map[string]any{"name": "Alice"}).
	Send().
	Response()
if err != nil {
	return err
}
defer resp.Close()
text, err := resp.Text()
if err != nil {
	return err
}
fmt.Println(text)
```

The snippet uses `net/url` and the caller's `ctx`. `Request(method, url)` supports
custom methods; Get, Post, Put, Patch, Delete, Head and Options are shortcuts.
`Header` replaces a header's values; `Headers(http.Header)` copies multiple values.
`Query(url.Values)` replaces values for supplied keys and preserves other URL keys.
`JSON`, `Form(url.Values)` and `Body(io.Reader)` select one body source. Selecting
another body records `ErrBodyAlreadySet`, including repeated Form calls; the old
`Data` request option still merges fields as before. JSON/Form set Content-Type;
a later Header call can override it.

Setters do no work after the first error or after Send. Send reports construction
errors without sending; calling it again returns `ErrRequestSent` without changing
the first Result. The builder is single-use even when the first attempt fails.
Go still evaluates arguments to skipped setters.

If retaining the Result instead of immediately transferring its response, call
`Close()` to release it. Close does not read the remaining response and returns any
request or cleanup error. Repeated Close does not close again. After Response
transfers ownership, Result.Close is harmless and the caller must close the response.
Repeated Response calls return `ErrResultConsumed`; a stored failure is returned
unchanged. A zero Result is empty and a RequestBuilder must be created by NewRequest.

### Fluent response reading

Status checks are explicit. Read terminals consume and close the response on both
success and failure, so this chain does not need a separate Close:

```go
client := requests4go.NewClient(nil)
value, err := client.NewRequest(ctx).
    Post("https://httpbin.org/post").
    JSON(map[string]any{"name": "Alice"}).
    Send().
    Require2xx().
    LimitBody(1 << 20).
    DecodeJSON[map[string]any]()
if err != nil {
    return err
}
fmt.Println(value)
```

`Require2xx()` accepts 200–299; `RequireStatus(code)` accepts exactly that code.
On rejection, the Result stores a `*StatusError` with StatusCode, Status and a
Header snapshot, then closes without reading the body. Use `errors.As` to inspect
it. No response body is retained in the error, and later chain operations do no work.

`Bytes()` and `Text()` return the complete body; Text does not perform charset
conversion. `DecodeJSON[T]()` returns a typed value, and `DecodeJSONInto(dst)` fills
an existing non-nil pointer. Both use `encoding/json`, require `application/json`
(parameters such as charset are allowed), and reject trailing non-whitespace data.
Type and syntax errors retain their standard library identities. Into may partly
modify dst on a decoding error; the generic method returns the zero value on error.
`Discard()` closes without draining the body and returns the first error.

Read and save terminals default to unlimited size. `LimitBody(n)` explicitly bounds bytes
read from Body (after automatic decompression), including JSON whitespace. It reads
at most n+1 bytes to detect overflow, returns `ErrResponseTooLarge`, and closes the
body. Zero permits only an empty body; a negative value records an error and closes
without reading. A later valid limit replaces an earlier limit. Limits do not apply
to responses transferred by `Response()`; callers control those reads themselves.
Use a context deadline or client timeout to bound waiting as well as size.

Successful terminals consume the Result once. Further reads return
`ErrResultConsumed`; after a failure, later operations retain that first error.
Read/decoding/status errors take precedence over cleanup errors. If only closing
fails, that error is returned; Bytes/Text/generic JSON return a zero result on error.

The minimal surface and ownership rules are:

| Stage | Operations | Contract |
| --- | --- | --- |
| RequestBuilder (available) | Request, Get/Post/etc., Header, Headers, Query, JSON, Form, Body, File, Multipart | Mutable, per-request, not safe for concurrent use. Retains its context and first error. No network activity before Send. |
| Send (available) | Returns a non-nil Result, including on error | Single use; repeated sends fail without sending again. Builds through the shared request core. |
| Result checks (available) | Require2xx, RequireStatus, LimitBody | First failure stops subsequent work and closes any owned response. Status checking and size limits are explicit. |
| Result terminals (available) | Text, Bytes, DecodeJSON[T], DecodeJSONInto, Discard, Save | Consume once and close on both success and failure; return the first error. Discard closes without unbounded draining. |
| Response escape (available) | Response() | Transfers ownership to the caller, who must close the response. The Result can no longer consume or close it. |
| Explicit cleanup (available) | Close() | Releases an unconsumed Result without reading it. A chain abandoned after Send still requires cleanup. |

Further contract decisions:

- Builder and Result instances must not be copied or used concurrently. Reusing a
  terminal or transferred Result fails without touching the response again;
  repeated Close is harmless. A successful Send consumes the builder, as does a
  failed Send attempt.
- Methods after a recorded error do no work. Go still evaluates their arguments
  before calling them; argument-expression side effects cannot be short-circuited.
- JSON and form setters encode a snapshot; header and query setters copy values.
  Reader bodies retain standard library replay support for bytes/string readers;
  arbitrary streams are not made replayable. File paths open at Send time.
  A second body source is an error in the new builder only. Old RequestOptions
  retain their existing replacement and form-merging rules.
- A caller retains ownership of an accepted Reader before Send. Calling Send
  transfers ownership of accepted closers to the library, even if building or
  sending fails. Configuration alone does not open or close external resources.
  A setter skipped because of an earlier error does not take ownership of its input.
- Builder failures are not sent. Result failures retain the first error and its
  identity; cleanup errors only become the returned error if no earlier error
  exists. Status errors retain response metadata without buffering an unlimited
  error body.
- Existing Response methods continue to consume a stream without caching or
  automatic closing. Automatic closing applies only to the new Result terminals.
- There is no implicit retry, default status rejection, base URL resolution, or
  garbage-collection-based cleanup. Request URLs follow net/http semantics.

### Streaming uploads

Use `File(path)` for a raw file body. It opens the path only during Send and closes
it through the HTTP client's lifecycle. It does not infer a Content-Type.
`Multipart(parts...)` supports ordered fields, memory-backed files and repeated
file names without buffering entire payloads:

```go
client := requests4go.NewClient(nil)
err := client.NewRequest(ctx).Post("https://httpbin.org/post").Multipart(
    requests4go.Part{Name: "title", Body: strings.NewReader("Reports")},
    requests4go.Part{Name: "files", Filename: "one.txt", ContentType: "text/plain", Body: strings.NewReader("first")},
    requests4go.Part{Name: "files", Filename: "two.txt", Body: strings.NewReader("second")},
).Send().Require2xx().Discard()
if err != nil {
    return err
}
```

This snippet uses `strings`. To stream an existing file in a multipart part, pass
an open `*os.File` as Body. Each accepted part's closer transfers to the library
at Send, including when construction fails. Before Send, the caller still owns
the readers. Rejected or skipped parts remain caller-owned. Use a separate reader
for each part. Readers that may block must allow Close to interrupt Read; arbitrary
blocking non-closable readers cannot be forcibly canceled by this library.
Request body closing follows net/http and may finish asynchronously; its close
errors are not promoted to response errors. Upload read errors propagate through Do.

Multipart buffers only framing metadata and reads payloads sequentially. It does
not launch producer goroutines. File and multipart bodies have unknown length
(HTTP/1.1 normally uses chunked transfer) and no GetBody replay function: 307/308
redirects requiring replay are returned to the caller. Standard 301/302/303 policies
still apply. Create a fresh builder and fresh readers for another attempt. The old
`FileContent` and `MultipartForm` options retain their existing buffered behavior.

### Saving responses

```go
client := requests4go.NewClient(nil)
err := client.NewRequest(ctx).Get("https://httpbin.org/bytes/1024").Send().
    Require2xx().LimitBody(10 << 20).Save("download.bin")
if err != nil {
    return err
}
```

Save streams to a temporary file in the target directory. Only after transfer,
file close and response close succeed does it rename the temporary file to the
target. Earlier failures leave an existing target unchanged and attempt to remove
the temporary file. Replacement follows the host's `os.Rename` semantics; rename
failure is returned. The directory must exist. The new file uses mode 0600 where
supported, does not preserve old permissions, and is not fsynced for crash durability.
Save consumes the Result and observes LimitBody without buffering the whole file.
Use an explicit status check if non-success response bodies must not be saved.

Legacy `Response.SaveContent` keeps its direct-create/truncate behavior, may leave
a partial file on error, and leaves response closing to the caller. It now reports
file close failures. For custom download writers or progress reporting, transfer
the Response and use its Body with `io.Copy`; the caller then owns all cleanup.

### Compatibility and validation

Existing package functions, RequestOption signatures and Session methods remain
available. Response reads still consume a stream; they do not cache or automatically
close it. JSON encoding continues to use encoding/json semantics. New behavior is
provided through Client, RequestBuilder and Result, rather than replacing old APIs.

CI runs tests and vet on Linux, Windows and macOS at the go.mod toolchain version,
plus latest stable Go on Linux. Linux runs the race detector. Module checks verify
checksums and that go.mod/go.sum are tidy. Runnable examples and external-package
legacy API tests are included in `go test ./...`. Cross-compilation alone does not
substitute for running platform tests.

Before a release, require successful CI for the exact commit and run
`go run golang.org/x/vuln/cmd/govulncheck@latest ./...`. Review any Dependabot alerts
alongside scanner results; inability to access the database is not a clean scan.

### Unreleased changes

- Raised the Go baseline to 1.27.3 and updated dependencies and CI actions.
- Fixed form encoding, Session.Head, stale body length/replay metadata, JSON media
  type parameters, and response closing without unbounded draining.
- Added configurable Client and context-aware Request/Do entrypoints for Session.
- Added single-use fluent request building, typed JSON response decoding, explicit
  status checks, response size limits, and automatic-closing Result terminals.
- Added streaming file/multipart uploads and staged response file saving.
- Existing API signatures and the module path are preserved. Callers relying on
  the previous erroneous form/HEAD/Close behavior should review the fixes above.

License
=======

Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text

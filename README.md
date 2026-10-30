# request4go

![test](https://github.com/Kaiser925/requests4go/workflows/test/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/Kaiser925/requests4go.svg)](https://pkg.go.dev/github.com/Kaiser925/requests4go)

Go HTTP Requests. ✨🎉✨ Send requests quickly and humanely.

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

### Fluent response design (next stage)

RequestBuilder, Send, Result.Response and Result.Close are implemented. Stage 4
will add status checks and automatic-closing read terminals. The intended shape is:

```text
value, err := client.NewRequest(ctx).
    Post(url).
    JSON(payload).
    Send().
    Require2xx().
    DecodeJSON[Value]()
```

The minimal surface and ownership rules are:

| Stage | Operations | Contract |
| --- | --- | --- |
| RequestBuilder (available) | Request, Get/Post/etc., Header, Headers, Query, JSON, Form, Body | Mutable, per-request, not safe for concurrent use. Retains its context and first error. No network activity before Send. |
| Send (available) | Returns a non-nil Result, including on error | Single use; repeated sends fail without sending again. Builds through the shared request core. |
| Result checks (planned) | Require2xx, RequireStatus | First failure stops subsequent work and closes any owned response. Status checking is explicit. |
| Result terminals (planned) | Text, Bytes, DecodeJSON[T], DecodeJSONInto, Discard | Consume once and close on both success and failure; return the first error. Discard closes without unbounded draining. |
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
  arbitrary streams are not made replayable. Future path-based files open at Send time.
  A second body source is an error in the new builder only. Old RequestOptions
  retain their existing replacement and form-merging rules.
- A caller retains ownership of an accepted Reader before Send. Calling Send
  transfers ownership of accepted closers to the library, even if building or
  sending fails. Configuration alone does not open or close external resources.
  A setter skipped because of an earlier error does not take ownership of its input.
- Builder failures are not sent. Result failures retain the first error and its
  identity; cleanup errors only become the returned error if no earlier error
  exists. Status errors retain response metadata without buffering an unlimited
  error body. Exact error types will be introduced with their implementing stage.
- Existing Response methods continue to consume a stream without caching or
  automatic closing. Automatic closing applies only to the new Result terminals.
- There is no implicit retry, default status rejection, base URL resolution, or
  garbage-collection-based cleanup. Request URLs follow net/http semantics.

Streaming upload helpers and file save terminals remain later work; they are not
required for the initial fluent API.

License
=======

Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text

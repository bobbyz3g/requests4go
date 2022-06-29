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

	txt, _ := r.Text()
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
	
	// handle r and err
}
```

### Passing Parameters In URLS

```go
params := requests4go.Params(requests4go.M{"key1": "value1", "key2": "value2"})
r, err := requests4go.Get("http://httpbin.org/get", params)
```

### Custom Headers

```go
	headers := requests4go.Headers(requests4go.M{"key1": "value1", "key2": "value2"})

	r, err := requests4go.Get("http://httpbin.org/get", headers)
```

### Response Content

We can read the content of the server's response.

~~~go
resp, _ := requests4go.Get("https://httpbin.org/get", nil)
txt, _ := resp.Text()
log.Println(txt)
~~~

### Decode Response Content

There are two methods to handle JSON response content.

1. We can unmarshal the struct by using JSON.

```go
foo := &Foo{}
resp, _ := requests4go.Get("https://example.com")
j, _ := resp.JSON(&foo)
fmt.Printf("%v\n", foo)
```

2. Struct implements Unmarshaler.

```go
package foo

type Foo struct {}

type (f *Foo) Unmarshal(p []byte) error {
	// ... unmarshal logic
	return nil
}

func F() {
	f := &Foo{}
	resp, _ := requests4go.Get("https://example.com")
	err := resp.Unmarshal(f)
	if err != nil {
		// error handle
    }
	fmt.Printf("%v\n", foo)
}
```

License
=======

Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text

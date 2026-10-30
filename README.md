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

License
=======

Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text

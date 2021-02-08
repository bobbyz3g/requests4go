# request4go

![test](https://github.com/Kaiser925/requests4go/workflows/test/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/Kaiser925/requests4go.svg)](https://pkg.go.dev/github.com/Kaiser925/requests4go)

Go HTTP Requests. ✨🎉✨

Install
=======

~~~
go get -u github.com/Kaiser925/requests4go
~~~

Usage
=====

First, we import module.

~~~go
import "github.com/Kaiser925/requests4go"
~~~

### Make a Request

~~~go
// get
resp, err := requests4go.Get("http://httpbin.org/get")
if err != nil {
	fmt.Println(err)
}
fmt.Println(resp.Status)
~~~

You can also set Request with RequestOption

```go
headers := map[string]string{
	"x-test-header": "value"
}

resp, _ := requests4go.Get("http://httpbin.org/get", Headers(headers))
```

Now, we hava a **Response** object called resp. We can get all the information we need from this object.

### Response Content

We can read the content of the server's response.

~~~go
resp, _ := requests4go.Get("https://httpbin.org/get", nil)
txt, _ := resp.Text()
fmt.Println(txt)

// Output:
// {
// "args": {},
// "headers": {
// 	"Accept": "/*",
// 	"Accept-Encoding": "gzip, deflate",
// ...
~~~

### JSON Response Content

There are two methods to handle JSON response content.

1. We can deal with SimpleJSON witch use [go-simplejson](https://github.com/bitly/go-simplejson) parse json data.

~~~go
resp, _ := requests4go.Get("https://httpbin.org/get")
j, _ := resp.SimpleJSON()
url, _ := j.Get("url").String()
fmt.Println(url)

// Output:
// https://httpbin.org/get
~~~

2. We can unmarshal the struct by using JSON.

```go
foo := &Foo{}
resp, _ := requests4go.Get("https://example.com")
j, _ := resp.JSON(foo)
fmt.Println(j.bar)
```

License
=======

Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text

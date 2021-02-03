# request4go

![test](https://github.com/Kaiser925/requests4go/workflows/test/badge.svg)

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

We can deal with JSON data by using [go-simplejson](https://github.com/bitly/go-simplejson).

~~~go
resp, _ := requests4go.Get("https://httpbin.org/get")
j, _ := resp.SimpleJSON()
url, _ := j.Get("url").String()
fmt.Println(url)

// Output:
// https://httpbin.org/get
~~~

License
=======

Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text

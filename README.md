# request4go

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
resp, err := requests4go.Get("http://httpbin.org/get", nil)
if err != nil {
	fmt.Println(err)
}
fmt.Println(resp.Status)
~~~

Now, we hava a **Response** object called resp. We can get all the information we need from this object.

Make an HTTP POST request.

~~~go
args := requests4go.NewRequestArguments()

args.Data = map[string]string{
	"name": "Apple",
}

resp, err := requests4go.Post("https://httpbin.org/post", args)

if err != nil {
	fmt.Println(err)
}
fmt.Println(resp.Status)
~~~

Other HTTP request types: PUT, DELETE, HEAD and OPTIONS.

~~~go
args := requests4go.NewRequestArguments()

args.Data = map[string]string{
	"name": "Apple",
}

resp, err := requests4go.Put("https://httpbin.org/put", args)

if err != nil {
	fmt.Println(err)
}
fmt.Println(resp.Status)

resp, err = requests4go.Delete("https://httpbin.org/delete", nil)

resp, err = requests4go.Head("https://httpbin.org/get",nil)

resp, err = requests4go.Options("https://httpbin.org/get",nil)
~~~

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
resp, _ := requests4go.Get("https://httpbin.org/get", nil)
j, _ := resp.Json()
url, _ := j.Get("url").String()
fmt.Println(url)

// Output:
// https://httpbin.org/get
~~~

License
=======

Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text
# request4go

[English](README.md) | [简体中文](README.zh-CN.md)

![test](https://github.com/Kaiser925/requests4go/workflows/test/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/Kaiser925/requests4go.svg)](https://pkg.go.dev/github.com/Kaiser925/requests4go)

基于 Go 标准库构建的 HTTP 客户端，让请求的构建、发送和响应处理更易用。

需要 Go 1.27.3 或更高版本。模块路径仍为 `github.com/Kaiser925/requests4go`，使用此 fork 时也保持不变。

## 快速开始

安装模块：

~~~
go get -u github.com/Kaiser925/requests4go
~~~

### 发送请求

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

也可以发送 POST 请求：

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

### URL 查询参数

除完整展示 package 的示例外，下面的代码片段应放在返回 error 的函数中。

```go
params := requests4go.Params(requests4go.M{"key1": "value1", "key2": "value2"})
r, err := requests4go.Get("http://httpbin.org/get", params)
if err != nil {
	return err
}
defer r.Close()
```

### 自定义请求头

```go
headers := requests4go.Headers(requests4go.M{"key1": "value1", "key2": "value2"})

r, err := requests4go.Get("http://httpbin.org/get", headers)
if err != nil {
	return err
}
defer r.Close()
```

### 读取响应内容

使用 `Text`、`Content` 或解码方法读取响应体。这些方法会消费响应流，不会缓存内容，也不会自动关闭响应。使用完毕后必须关闭响应；`Close` 直接关闭响应体，不读取剩余内容。

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

### 解码响应内容

处理 JSON 响应有两种方式。

1. 使用 `JSON` 解码到 Go 值。

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

2. 为自定义类型实现 `Unmarshaler`。

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

### 配置客户端与 context

`Client` 包装标准库的 `*http.Client`，不会复制或修改它。使用前，通过标准库客户端配置超时、重定向、代理、Transport 和 Cookie Jar：

```go
client := requests4go.NewClient(&http.Client{Timeout: 10 * time.Second})
resp, err := client.Request(ctx, http.MethodGet, "https://httpbin.org/get",
	requests4go.Params(requests4go.M{"page": "1"}))
if err != nil {
	return err
}
defer resp.Close()
```

这里的 `ctx` 是调用者提供的非 nil `context.Context`；示例还使用了 `net/http` 和 `time`。发送已有的 `*http.Request` 时，使用 `client.Do(req)`。请求携带的 context 控制请求及响应体的读取。取消和超时错误保留标准库错误的身份，可使用 `errors.Is` 和 `errors.As` 判断。

`Client` 的零值和 `NewClient(nil)` 使用共享的默认 HTTP 客户端，不设置超时或 Cookie Jar。它独立于 `http.DefaultClient`。底层 Transport 支持并发时，Client 可以并发复用；使用过程中不要修改客户端配置。每个响应由对应的调用者负责管理。

`NewSession()` 仍会创建 Cookie Jar。原有方法继续可用，新增的 `session.Request(ctx, method, url, opts...)` 和 `session.Do(req)` 使用当前的 `session.Client`，包括它的 Cookie Jar 和相关策略。仍可在请求之间替换该字段；将它设为 nil 时，会使用不保存 Cookie 的共享默认客户端。

已有包级函数保留原来的签名，并共用同一个执行核心。这些发送入口不会自动将非 2xx 状态转换为错误，不会读取或缓存响应，也不会替调用者关闭成功返回的响应。原有请求选项的顺序、表单合并及请求体覆盖行为保持不变。

### 链式构建请求

通过 `Client.NewRequest(ctx)` 和链式配置方法构建请求。`Send` 只尝试发送一次，返回 `Result`；`Response` 返回首个错误，或将响应的所有权移交给调用者：

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

示例使用了 `net/url` 和调用者提供的 `ctx`。`Request(method, url)` 支持自定义 HTTP 方法；`Get`、`Post`、`Put`、`Patch`、`Delete`、`Head`、`Options` 是快捷方法。

`Header` 替换指定请求头的值；`Headers(http.Header)` 复制多值请求头。`Query(url.Values)` 替换传入键对应的查询参数，保留 URL 中其他键的值。

`JSON`、`Form(url.Values)` 和 `Body(io.Reader)` 用于选择请求体来源。再次设置请求体会记录 `ErrBodyAlreadySet`，重复调用 `Form` 也一样；旧的 `Data` 请求选项仍按原规则合并字段。`JSON` 和 `Form` 会设置 Content-Type，后续的 `Header` 调用可以覆盖它。

记录首个错误或调用 `Send` 后，后续配置方法不再执行操作。构建出错时，`Send` 返回错误而不发送请求；重复调用返回 `ErrRequestSent`，不会改变第一次返回的 Result。即使首次尝试失败，构建器也不能再次发送。Go 仍会求值被跳过方法的参数表达式。

如果保留 Result，而不立即移交响应，需要调用 `Close()` 释放它。Close 不读取剩余响应体，并返回已有的请求错误或清理错误。重复 Close 不会再次关闭资源。`Response` 移交所有权后，Result.Close 不再关闭该响应，调用者必须自行关闭它。重复调用 Response 返回 `ErrResultConsumed`；如果已经记录错误，则继续返回该错误。Result 的零值不包含响应；RequestBuilder 必须通过 NewRequest 创建。

### 链式读取响应

状态检查需要显式调用。读取终结方法无论成功或失败，都会消费并关闭响应，因此下面的调用链不需要额外调用 Close：

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

`Require2xx()` 接受 200–299；`RequireStatus(code)` 只接受指定状态码。不符合要求时，Result 保存 `*StatusError`，其中包含 StatusCode、Status 和 Header 快照，然后关闭响应而不读取响应体。可以用 `errors.As` 获取该错误。错误中不保留响应体，后续链式操作不再执行。

`Bytes()` 和 `Text()` 返回完整响应体；Text 不进行字符集转换。`DecodeJSON[T]()` 返回指定类型的值；`DecodeJSONInto(dst)` 将结果填入已有的非 nil 指针。二者都使用 `encoding/json`，要求 Content-Type 为 `application/json`，允许 charset 等参数，并拒绝完整 JSON 值之后的非空白数据。类型和语法错误保留标准库错误的身份。Into 解码失败时可能部分修改 dst；泛型方法在错误时返回类型零值。`Discard()` 不排空响应体，直接关闭并返回首个错误。

读取和保存方法默认不限制大小。`LimitBody(n)` 显式限制从 Body 读取的字节数，按自动解压后的内容计算，包括 JSON 中的空白字符。它最多读取 n+1 字节来检测超限；超限时返回 `ErrResponseTooLarge` 并关闭响应体。零表示仅允许空响应体；负数会记录错误并直接关闭，不读取内容。后一次有效设置覆盖前一次限制。通过 `Response()` 移交的响应不受该限制约束，后续读取由调用者控制。需要限制等待时间时，还应设置 context 截止时间或客户端超时。

成功的终结操作只能消费 Result 一次，再次读取返回 `ErrResultConsumed`；失败后，后续操作保留首个错误。读取、解码和状态错误优先于清理错误。如果仅关闭失败，则返回关闭错误；Bytes、Text 和泛型 JSON 方法在错误时返回零值。

接口及资源管理规则如下：

| 阶段 | 操作 | 契约 |
| --- | --- | --- |
| RequestBuilder | Request、Get/Post 等、Header、Headers、Query、JSON、Form、Body、File、Multipart | 每个请求单独创建，可修改，不支持并发使用。保留 context 和首个错误；Send 之前没有网络操作。 |
| Send | 返回非 nil Result，包括失败时 | 只能调用一次；重复调用不会再次发送。使用统一的请求构建核心。 |
| Result 检查 | Require2xx、RequireStatus、LimitBody | 首个错误阻止后续操作，并关闭仍归 Result 管理的响应。状态检查和大小限制均需显式设置。 |
| Result 终结方法 | Text、Bytes、DecodeJSON[T]、DecodeJSONInto、Discard、Save | 消费一次，无论成功或失败都关闭响应，返回首个错误。Discard 不会无限制排空响应体。 |
| 移交响应 | Response() | 将所有权交给调用者，由调用者关闭。Result 此后不能再次消费或关闭该响应。 |
| 显式清理 | Close() | 不读取内容，释放尚未消费的 Result。如果调用链停在 Send，仍需进行清理。 |

其他约定：

- Builder 和 Result 不能复制或并发使用。对已消费或已移交的 Result 再次调用终结方法会失败，不会再次操作响应；重复 Close 不会重复关闭资源。无论成功还是失败，Send 都会消耗构建器的一次发送机会。
- 记录错误后，后续方法不再执行操作。但 Go 会在调用方法前求值参数，因此无法跳过参数表达式本身的副作用。
- JSON 和表单配置方法会编码快照；请求头和查询参数会复制输入值。Reader 请求体保留标准库对字节、字符串 Reader 的重放支持，不会使任意流自动变得可重放。File 在 Send 时打开路径。仅新构建器会将第二个请求体来源视为错误；旧 RequestOption 保留原有覆盖和表单合并规则。
- Send 前，已接受的 Reader 仍归调用者所有。调用 Send 后，它们的 closer 移交给库，包括构建或发送失败的情况。配置操作本身不打开或关闭外部资源。因已有错误而跳过的配置方法不会接管其输入。
- 构建失败不会发送请求。Result 保留首个错误及其身份；只有不存在更早错误时，清理错误才成为返回错误。状态错误保留响应元数据，不会无限制缓存错误响应体。
- 原有 Response 方法仍然消费响应流，不缓存、不自动关闭。自动关闭仅适用于新的 Result 终结方法。
- 库不会额外添加隐式重试、默认状态拒绝、基础 URL 解析或依赖垃圾回收的清理机制。请求 URL 遵循 net/http 语义。

### 流式上传

使用 `File(path)` 将原始文件内容作为请求体。它只在 Send 时打开路径，并通过 HTTP 客户端的生命周期关闭文件，不自动推断 Content-Type。

`Multipart(parts...)` 支持有序字段、内存文件和同名多文件，不会将整个请求体读入内存：

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

示例使用了 `strings`。要在 multipart 中流式发送已有文件，可以将打开的 `*os.File` 作为 Body。每个已接受 part 的 closer 都会在 Send 时移交给库，包括构建失败的情况。Send 前，Reader 仍归调用者所有；被拒绝或被跳过的 part 也仍由调用者管理。每个 part 应使用独立的 Reader。可能阻塞的 Reader 必须允许 Close 中断 Read；库无法强制取消任意不可关闭的阻塞 Reader。

请求体关闭遵循 net/http 的行为，可能异步完成；其关闭错误不会转换为响应错误。上传读取错误通过 Do 传播。

Multipart 仅缓冲边界和头部等元数据，按顺序读取各部分内容，不启动后台生产 goroutine。File 和 multipart 请求体长度未知，HTTP/1.1 通常使用分块传输，并且没有 GetBody 重放函数：需要重放请求体的 307/308 重定向会直接返回给调用者。标准库的 301/302/303 策略仍然适用。再次尝试时应创建新的构建器和 Reader。旧 `FileContent`、`MultipartForm` 请求选项保留原有的缓冲行为。

### 保存响应

```go
client := requests4go.NewClient(nil)
err := client.NewRequest(ctx).Get("https://httpbin.org/bytes/1024").Send().
    Require2xx().LimitBody(10 << 20).Save("download.bin")
if err != nil {
    return err
}
```

Save 将响应流写入目标目录中的临时文件。只有传输、文件关闭和响应关闭都成功后，才将临时文件重命名为目标文件。此前发生错误时，已有目标文件保持不变，并尝试移除临时文件。文件替换遵循当前平台的 `os.Rename` 语义，重命名失败会返回错误。目标目录必须已经存在。平台支持时，新文件使用 0600 权限，不保留原文件权限，也不执行用于保证崩溃后持久性的 fsync。

Save 消费 Result 并遵守 LimitBody，不会缓存整个文件。如果不希望保存非成功状态的响应体，应显式检查状态码。

旧 `Response.SaveContent` 保留直接创建或截断文件的行为，错误时可能留下部分内容，响应仍由调用者关闭。它现在也会返回文件关闭错误。需要自定义下载 Writer 或显示进度时，可以移交 Response，再通过 `io.Copy` 读取其 Body；此时所有清理由调用者负责。

### 兼容性

需要 Go 1.27.3 或更高版本。我们保留已有公开 API 的签名和文档约定的行为，同时修复违反这些约定的错误。新增功能通过新增 API 提供；已有包级函数、RequestOption 和 Session 继续受到支持。

## 许可证

采用 Apache License, Version 2.0。完整许可证见 [LICENSE](LICENSE)。

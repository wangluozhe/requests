# requests
[![Gitee link address](https://img.shields.io/badge/gitee-reference-red?logo=gitee&logoColor=red&labelColor=white)](https://gitee.com/leegene/requests)[![Github link address](https://img.shields.io/badge/github-reference-blue?logo=github&logoColor=black&labelColor=white&color=black)](https://github.com/wangluozhe/requests)[![Go Version](https://img.shields.io/badge/Go%20Version-1.27.0-blue?logo=go&logoColor=white&labelColor=gray)]()[![Release Version](https://img.shields.io/badge/release-v1.6.0-blue)]()[![go documentation](https://img.shields.io/badge/go-documentation-blue)](https://pkg.go.dev/github.com/wangluozhe/requests)[![license GPL-3.0](https://img.shields.io/badge/license-GPL3.0-orange)](https://github.com/wangluozhe/requests/blob/main/LICENSE)

requests 是基于 [chttp](https://github.com/wangluozhe/chttp) 的 Go HTTP 请求库，支持：

1. HTTP/1.1 和 HTTP/2，以及请求头和 HTTP/2 伪头排序。
2. 自定义 JA3、TLS 扩展和 HTTP/2 指纹。
3. 通过 `ClientHelloID` 选择 uTLS 预设，或通过 `ClientHelloHexStream` 使用抓包中的 ClientHello。
4. Session Cookie 持久化、连接复用、有上限的 Transport 缓存和显式关闭。
5. 灵活的请求参数、Hook、中间件、流式读取和响应体保留策略。
6. C API 普通请求、流式请求及资源回收。

未配置指纹和 HTTP/2 参数时，默认使用 HTTP/1.1。设置 JA3、ClientHello 或 `HTTP2Settings` 后会准备 HTTP/2 支持，HTTPS 最终使用的协议由 ALPN 协商决定；`ForceHTTP1` 可以强制 HTTP/1.1。详见[协议选择](#协议选择)。

**此模块参考于Python的[requests模块](https://github.com/psf/requests/tree/main/requests)**


## 文档导航

- [快速上手](#快速上手)、[Session 会话](#session-会话)、[超时](#超时)
- [流式请求](#stream模式-流式请求)、[响应体保留策略](#响应体保留策略)
- [协议选择](#协议选择)、[JA3 指纹](#ja3指纹)、[ClientHello 预设与抓包](#clienthello-预设与抓包)
- [TLS 指纹](#tls指纹)、[HTTP/2 指纹](#http2指纹)、[有序请求头](#有序请求头)
- [C API](#c-api)、[升级与兼容性说明](#升级与兼容性说明)


## 加入微信群聊
如果你想跟更多志同道合的人一起学习，那么加入群聊来吧！(备注：加群)
![微信群聊](./imgs/wechat.jpg)


## 下载requests库

```bash
go get github.com/wangluozhe/requests
```



## 下载指定版

```bash
go get github.com/wangluozhe/requests@v1.6.0
```



# 快速上手

迫不及待了吗？本页内容为如何入门 Requests 提供了很好的指引。其假设你已经安装了 Requests。如果还没有，去安装看看吧。

首先，确认一下：

- Requests已安装
- Requests是最新的

让我们从一些简单的示例开始吧。



## DEBUG模式

`requests.SetDebug(true)` 全局开启调试输出，打印到标准输出（stdout），包括请求配置、请求和响应报文及 Body。调试输出会读取响应体，因此也会影响流式请求的内存占用和返回时机。

```go
import (
    "github.com/wangluozhe/requests"
)

// 打开Debug模式
requests.SetDebug(true)

// 请求代码...

// 取消Debug模式
requests.SetDebug(false)
```



## 更灵活的参数传递 (Flexible Parameter Passing)

`Request` 结构体中的参数（`Params`, `Headers`, `Cookies`, `Data`, `Files`, `Body`）现在支持 `any` 类型。您不再局限于特定的结构体，可以直接传入 `map`、`string` 或其他常见类型，库会自动进行转换。

**支持的类型示例 (Supported Types Examples):**

* **Params (Query String):**
* `*url.Params`
* `string` (e.g. "key=value&a=1")
* `map[string]string`
* `map[string][]string`
* `map[string]int`
* `map[string][]int`
* `map[string]float64`
* `map[string][]float64`
* `map[string]interface{}` (支持递归解析)


* **Headers:**
* `*http.Header`
* `string` (e.g. "User-Agent: abc\nAccept: */*")
* `map[string]string`
* `map[string][]string`
* `map[string]interface{}` (值支持 string, int, float64, bool)
* `map[string][]interface{}`


* **Cookies:**
* `*cookiejar.Jar`
* `string` (e.g. "name=value; a=1")
* `map[string]string`
* `map[string]int`
* `map[string]float64`
* `map[string]interface{}` (值支持 string, int, float64, bool)


* **Data (Form Data):**
* `*url.Values`
* `string` (e.g. "key=value&a=1")
* `map[string]string`
* `map[string][]string`
* `map[string]int`
* `map[string][]int`
* `map[string]float64`
* `map[string][]float64`
* `map[string]interface{}` (支持递归解析)


* **Files**
* `*url.Files`
* `map[string]string` (key为字段名, value为文件路径. 会自动提取文件名，ContentType默认为空)


* **Body:**
* `string`
* `[]byte`
* `io.Reader`

> **注意：** `AllowRedirects`、`Verify`、`RandomJA3`、`ForceHTTP1` 字段类型为 `*bool`，需要使用 `url.Bool()` 函数来设置。`Verify`、`RandomJA3`、`ForceHTTP1` 为 `nil` 时继承 Session；`AllowRedirects` 为 `nil` 时默认允许重定向（顶层 `requests.Head` 除外，见[重定向与请求历史](#重定向与请求历史)）。`url.Bool(true)` / `url.Bool(false)` 表示显式设置。
>
> ```go
> req.Verify = url.Bool(false)        // 显式关闭 TLS 证书验证
> req.AllowRedirects = url.Bool(false) // 显式禁用重定向
> req.ForceHTTP1 = url.Bool(true)      // 显式强制使用 HTTP/1
> ```

**使用示例 (Usage):**

```go
import "github.com/wangluozhe/requests/url"

req := &url.Request{
    // 直接使用 map 传递查询参数
    Params: map[string]string{
        "page": "1",
        "size": "20",
    },
    // 直接使用 map 传递 Header
    Headers: map[string]string{
        "Authorization": "Bearer token123",
    },
    // 直接传递字符串 Body
    Body: "raw body content",
}

```



## 发送请求

使用 requests 发送网络请求非常简单。

一开始要导入requests 模块：

```go
import (
    "github.com/wangluozhe/requests"
    "github.com/wangluozhe/requests/url"
)
```

然后，尝试获取某个网页。本例子中，我们来获取 Github 的公共时间线：

```go
r, err := requests.Get("https://api.github.com/events", nil)
```

现在，我们有一个名为 `r` 的 [`models.Response`](https://pkg.go.dev/github.com/wangluozhe/requests/models#Response) 对象。我们可以从这个对象中获取所有我们想要的信息。

Requests 简便的 API 意味着所有 HTTP 请求类型都是显而易见的。例如，你可以这样发送一个 HTTP POST 请求：

```go
data := url.NewData()
data.Set("key","value")
r, err := requests.Post("http://httpbin.org/post", &url.Request{Data: data})
```

漂亮，对吧？那么其他 HTTP 请求类型：PUT，DELETE，HEAD 以及 OPTIONS 又是如何的呢？都是一样的简单：

```go
data := url.NewData()
data.Set("key","value")
r, err := requests.Post("http://httpbin.org/post", &url.Request{Data: data})
r, err = requests.Delete("http://httpbin.org/delete", nil)
r, err = requests.Head("http://httpbin.org/get", nil)
r, err = requests.Options("http://httpbin.org/get", nil)
```

都很不错吧，但这也仅是 requests 的冰山一角呢。



## Session 会话

`requests.Get`、`requests.Post` 等顶层函数共用一个默认 Session，因此会跨调用复用连接和 Cookie。需要隔离账号或自定义会话默认值时，使用 `requests.NewSession()`：

```go
package main

import (
	"fmt"
	"time"

	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	session := requests.NewSession()
	defer session.Close()
	session.Timeout = 10 * time.Second
	session.Headers.Set("User-Agent", "my-app/1.0")
	session.Params = url.NewParams()
	session.Params.Set("source", "example")

	// 后续请求会自动携带符合域名、路径等作用域规则的 Cookie。
	first, err := session.Get("https://httpbin.org/cookies/set/session/example", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	first.Close()

	r, err := session.Get("https://httpbin.org/cookies", &url.Request{
		Timeout: 5 * time.Second, // 仅覆盖本次请求
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Close()
	fmt.Println(r.Text())
}
```

Session 可以设置公共 `Params`、`Headers`、`Cookies`、`Auth`、`Proxies`、`Cert`、TLS/HTTP2 指纹和 `ResponseRetention`。请求级 Params 和 Headers 会合并公共值，并覆盖同名项；请求 Cookie 按其作用域参与本次请求，服务器返回的 Cookie 会更新会话 Cookie Jar。

| 配置 | 新建 Session 的默认值 / 继承规则 |
| --- | --- |
| `Timeout` | 30 秒；Request 为 0 时继承 |
| `Verify` | `true`；Request 为 `nil` 时继承 |
| `RandomJA3`、`ForceHTTP1` | `false`；Request 为 `nil` 时继承 |
| `MaxRedirects` | 30；是否跟随由 Request 的 `AllowRedirects` 控制 |
| `TransportCacheLimit` | 32；小于等于 0 时仍使用默认 32 |

空字符串和 `nil` 配置通常表示继承，不表示清空 Session 的值。例如，空 `req.Proxies` 不会清除会话代理；TLS/HTTP2 配置非 `nil` 时按整个配置覆盖，不逐字段合并。

Transport 缓存按配置复用连接，超过上限时按最近使用情况淘汰并关闭空闲连接。`Session.Close()` 可重复调用，会释放空闲连接并拒绝新请求，不中断已经在用的连接；它不代替 `Response.Close()`。

并发发送前应完成 Session 配置和中间件注册。请求会对 TLS、HTTP/2 和 ClientHello 预设创建独立快照，但复制期间仍不能并发修改原配置、内部切片或 map。`Params`、`Values`、`Files` 的方法有锁，不代表整个 Session 的公开字段都支持并发读写。


## 设置代理

requests设置代理也非常简单，只需给req.Proxies传入代理ip即可。

```go
req := url.NewRequest()
req.Proxies = "http://127.0.0.1:32768"
// 设置带账号密码的代理：req.Proxies = "http://username:password@ip:port"

r, err := requests.Get("https://api.github.com/events", req)
```



## 传递 URL 参数

你也许经常想为 URL 的查询字符串(query string)传递某种数据。如果你是手工构建 URL，那么数据会以键/值对的形式置于 URL 中，跟在一个问号的后面。例如， `httpbin.org/get?key=val`。 Requests 允许你使用 `params` 关键字参数，以一个字符串字典来提供这些参数。举例来说，如果你想传递 `key1=value1` 和 `key2=value2` 到 `httpbin.org/get` ，那么你可以使用如下代码：

```go
params := url.NewParams()
params.Set("key1","value1")
params.Set("key2","value2")
r, err := requests.Get("http://httpbin.org/get",&url.Request{Params: params})
if err != nil {
	fmt.Println(err)
}
```

或者：

```golang
// map[string]type，type = string|int|float64|interface{}|[]string|[]int|[]float64|[]interface{}
req := url.NewRequest()
//params := map[string]string{
//	"page":  "1",
//	"limit": "20",
//	"skip":  "5",
//}
//params := map[string]int{
//	"page":  1,
//	"limit": 20,
//	"skip":  5,
//}
//params := map[string][]string{
//	"page":  []string{"1", "2"},
//	"limit": []string{"20"},
//	"skip":  []string{"5"},
//}
//params := map[string][]int{
//	"page":  []int{1, 2},
//	"limit": []int{20},
//	"skip":  []int{5},
//}
params := map[string]interface{}{
    "page":  []interface{}{"1", 2},
    "limit": []string{"20"},
    "skip":  []int{5},
}
req.Params = url.ParseParams(params)
r, err := requests.Get("https://httpbin.org/get", req)
if err != nil {
    fmt.Println(err)
}
fmt.Println(r.Text())
```

或者：

```go
params := url.ParseParams("key1=value1&key2=value2")
r, err := requests.Get("http://httpbin.org/get",&url.Request{Params: params})
if err != nil {
	fmt.Println(err)
}
```

通过打印输出该 URL，你能看到查询参数已拼接到 URL：

```go
fmt.Println(r.Url)
http://httpbin.org/get?key1=value1&key2=value2
```

查询参数沿用旧版行为，不自动进行百分号编码。键或值包含空格、`&`、`=`、中文等字符时，请先对该键或值调用 `utils.EncodeURIComponent` 或标准库 `net/url.QueryEscape`，再传入 Params。已有的 `%20`、`%26` 等编码会保留，不会二次编码。

你还可以使Params有多个值传入：

```go
params := url.NewParams()
params.Set("key1","value1")
params.Add("key1","value2")
params.Set("key2","value2")
r, err := requests.Get("http://httpbin.org/get",&url.Request{Params: params})
if err != nil {
	fmt.Println(err)
}
fmt.Println(r.Url)

http://httpbin.org/post?key1=value1&key1=value2&key2=value2
```



## 响应内容

我们能读取服务器响应的内容。再次以 GitHub 时间线为例：

```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
)

func main() {
	r, err := requests.Get("https://api.github.com/events", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r.Text())
}
```

示例输出：

```text
[{"repository":{"open_issues":0, "url":"https://github.com/...
```

`Text()` 返回 `string(r.Content)`，不会自动转换 GBK 等字符集。响应解压与字符集转换是两回事；非 UTF-8 文本需要调用方自行转码。



## 二进制响应内容

你也能以字节数组的方式访问请求响应体，对于非文本请求：

```go
fmt.Println(r.Content)
[91 123 34 105 100 34 58 34 50 48 55 49 52 50 51 57 56 48 53 34 44 34 116 121 112 101 34 58 34...
```

Requests 会自动为你解码 `gzip` 和 `deflate` 以及`br`传输编码的响应数据。

例如，将二进制内容保存到文件：

```go
package main

import (
	"fmt"
	"os"

	"github.com/wangluozhe/requests"
)

func main() {
	r, err := requests.Get("https://httpbin.org/image/png", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Close()
	if err := os.WriteFile("image.png", r.Content, 0644); err != nil {
		fmt.Println(err)
	}
}
```

## Stream模式 (流式请求)

默认请求会将响应体完整读入 `Content`。下载大文件时，可以选择以下两种流式方式；两者都通过 `r.Body` 读取，`Content` 初始为 `nil`，使用后应调用 `r.Close()`：

| 设置 | Body 的内容 | RawResponse |
| --- | --- | --- |
| `req.Stream = true` | 直接返回底层响应流，本库不再解压；是否已解压取决于底层 Transport | 仅响应头 |
| `req.ResponseRetention = url.ResponseRetentionBody` | 按 `Content-Encoding` 包装的解压流，不整体读入内存 | 不保留 |

需要直接保存解压后文件时，可以使用 `ResponseRetentionBody`：

```go
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := &url.Request{ResponseRetention: url.ResponseRetentionBody}
	r, err := requests.Get("https://httpbin.org/stream-bytes/10240", req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Close()
	if err := r.RaiseForStatus(); err != nil {
		fmt.Println(err)
		return
	}

	out, err := os.Create("large_file.dat")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer out.Close()
	n, err := io.Copy(out, r.Body)
	if err != nil {
		fmt.Println("下载失败:", err)
		return
	}
	fmt.Printf("已保存 %d 字节\n", n)
}
```

设置 `Stream = true` 时，读取方式优先于响应保留策略，即使选了 `ResponseRetentionAll` 也不会缓存完整响应体；非法的保留策略仍会返回错误。流式返回后，请求的 `Timeout` 仍会限制后续 Body 读取。


## JSON 响应内容

requests 内置了解析 JSON 的方法。现在支持直接解析到 `struct`、`map`，或者使用快捷方法直接获取字典。

### 1. 解析到结构体 (推荐)

利用 Go 的强类型特性，你可以定义一个结构体，直接将响应解析进去：

```go
type GithubEvent struct {
	Type string `json:"type"`
	Repo struct {
		Name string `json:"name"`
	} `json:"repo"`
}

var event []GithubEvent
// 传入结构体指针
err := r.Json(&event)
if err != nil {
    fmt.Println("解析失败:", err)
}
fmt.Println(event[0].Repo.Name)

```

### 2. 解析到 Map

如果你不想定义结构体，可以将结果解析到 map 中：

```go
// 方法 A: 使用 Json() 传入 map 指针
var result map[string]interface{}
r.Json(&result)

// 方法 B: 使用 JsonMap() 快捷方法
result, err := r.JsonMap()
fmt.Println(result["type"])

```

### 3. 使用 simplejson (可选)

如果你习惯使用第三方库 [go-simplejson](https://github.com/bitly/go-simplejson) 进行链式操作，可以使用 `r.SimpleJson()`：

```go
js, err := r.SimpleJson()
// 链式获取值
repoName := js.GetIndex(0).Get("repo").Get("name").MustString()
fmt.Println(repoName)

```

如果 JSON 解码失败，解析方法会返回错误。

`Text()`、`Json()`、`JsonMap()` 只使用 `Content`，不会自动读取流式 Body。`SimpleJson()` 优先解析 `Content`，不改变已有 Body 的读取位置；没有 `Content` 时，它会读完 Body 的**剩余内容**、关闭原流、缓存 `Content` 并恢复为内存 Body。此时会失去流式读取的低内存优势。已读取的部分不会被恢复，未解压的流也不会在此方法中自动解压。

对 `ResponseRetentionBody` 返回的 JSON 流，可直接使用标准库 `encoding/json`：

```go
var result map[string]any
err := json.NewDecoder(r.Body).Decode(&result)
```

需要注意的是，成功调用 `r.Json()` 并**不**意味着响应的成功。有的服务器会在失败的响应中包含一个 JSON 对象（比如 HTTP 500 的错误细节）。这种 JSON 会被解码返回。要检查请求是否成功，请检查 `r.StatusCode` 是否和你的期望相同。



## 响应体保留策略

通过 `Request.ResponseRetention` 或 `Session.ResponseRetention` 选择保留内容。Request 非空值优先，其次使用 Session，均为空时使用 `ResponseRetentionDefault`。

| 策略 | Content | Body | RawResponse |
| --- | --- | --- | --- |
| `url.ResponseRetentionDefault` | 完整解压内容 | 复用 Content 底层数组的内存 Reader | 不保留 |
| `url.ResponseRetentionContent` | 完整解压内容 | `nil` | 不保留 |
| `url.ResponseRetentionBody` | `nil` | 解压流 | 不保留 |
| `url.ResponseRetentionAll` | 完整解压内容 | 复用 Content 底层数组的内存 Reader | 响应头及本库解压前的实体内容 |

`r.Retention` 记录实际采用的策略。默认和 `All` 策略下，`Body` 是有读取位置的 `io.ReadCloser`，读到 EOF 后不会自动回卷。需要重复解析时使用 `Content`，或通过 `bytes.NewReader(r.Content)` 新建 Reader。`Content` 与内存 Body 共享数据，不应边读取边修改它。

需要保存完整响应时，显式选择 `All`：

```go
req := &url.Request{ResponseRetention: url.ResponseRetentionAll}
r, err := requests.Get("https://example.com", req)
if err != nil {
    return
}
defer r.Close()
raw := r.RawResponse
_ = raw // 保存或处理报文
```

`RawResponse` 是格式化后的响应头和实体内容，不是 TLS 数据包或 HTTP/2 帧的网络抓包；它也不是 `r.Body`。`All` 会增加内存占用，普通请求无需开启。

`r.Close()` 会关闭 Body 和重定向历史中的响应，并清空 `Content`、`Body`、`RawResponse`。应在所有读取、解析和保存操作完成后关闭。


## 定制请求头

如果你想为请求添加 HTTP 头部，只要简单地`url.NewHeaders()` 给 `Headers` 参数就可以了。

例如，在前一个示例中我们没有指定 content-type:

```go
rawurl := "https://api.github.com/some/endpoint"
headers := url.NewHeaders()
headers.Set("user-agent", "my-app/0.0.1")
req := url.NewRequest()
req.Headers = headers

r, err := requests.Get(rawurl, req)
```
或者
```go
rawurl := "https://api.github.com/some/endpoint"
req := url.NewRequest()
req.Headers = url.ParseHeaders(map[string]interface{}{
    ":authority":                "spidertools.cn",
    ":method":                   "GET",
    ":path":                     "/",
    ":scheme":                   "https",
    "accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
    "accept-encoding":           "gzip, deflate, br",
    "accept-language":           "zh-CN,zh;q=0.9",
    "cache-control":             "no-cache",
    "cookie":                    "_ga=GA1.1.630251354.1645893020; Hm_lvt_def79de877408c7bd826e49b694147bc=1647245863,1647936048,1648296630; Hm_lpvt_def79de877408c7bd826e49b694147bc=1648296630",
    "pragma":                    "no-cache",
    "sec-ch-ua":                 "\" Not A;Brand\";v=\"99\", \"Chromium\";v=\"98\", \"Google Chrome\";v=\"98\"",
    "sec-ch-ua-mobile":          "?0",
    "sec-ch-ua-platform":        "\"Windows\"",
    "sec-fetch-dest":            "document",
    "sec-fetch-mode":            "navigate",
    "sec-fetch-site":            "same-origin",
    "sec-fetch-user":            "?1",
    "upgrade-insecure-requests": "1",
    "user-agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36",
}) // 注意：此方式生成的headers是无序的，因为map是无序的。
r, err := requests.Get(rawurl, req)
```

注意: 定制 header 的优先级低于某些特定的信息源，例如：

- `Content-Length`请求头不能随便设置，可能会有错误发生。
- 如果被重定向到别的主机，授权 header 就会被删除。
- 代理授权 header 会被 URL 中提供的代理身份覆盖掉。
- 在我们能判断内容长度的情况下，header 的 Content-Length 会被改写。

更进一步讲，Requests 不会基于定制 header 的具体情况改变自己的行为。只不过在最后的请求中，所有的 header 信息都会被传递进去。

注意: 所有的 header 值必须是 `string`。



## 有序请求头

使用 `chttp` 提供的特殊键设置顺序，这些键本身不会作为 HTTP Header 发送：

| 键 | 用途 |
| --- | --- |
| `http.HeaderOrderKey` | 普通 Header 的发送顺序，名称使用小写 |
| `http.PHeaderOrderKey` | HTTP/2 伪头顺序，例如 `:method`、`:authority`、`:scheme`、`:path` |
| `http.UnChangedHeaderKey` | HTTP/1.1 中需要保留名称大小写的 Header；HTTP/2 的普通 Header 仍使用小写 |

```go
package main

import (
	"fmt"

	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	headers := url.NewHeaders()
	headers.Set("User-Agent", "my-app/1.0")
	headers.Set("Accept", "application/json")
	headers.Set("Accept-Language", "zh-CN,zh;q=0.9")
	(*headers)[http.HeaderOrderKey] = []string{
		"user-agent", "accept", "accept-language",
	}
	(*headers)[http.PHeaderOrderKey] = []string{
		":method", ":authority", ":scheme", ":path",
	}
	req := &url.Request{
		Headers:       headers,
		HTTP2Settings: &http.HTTP2Settings{},
	}
	r, err := requests.Get("https://httpbin.org/get", req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Close()
	fmt.Println(r.Text())
}
```

也可以向 `url.ParseHeaders` 传入多行 Header 字符串，它会记录普通 Header 和伪头的顺序；传入 map 时，应另行设置顺序键。请求的方法、路径、协议和主机由请求参数及 URL 决定，不应使用普通 `Path`、`Scheme` Header 来代替 HTTP/2 伪头。


## 更加复杂的 POST 请求

通常，你想要发送一些编码为表单形式的数据——非常像一个 HTML 表单。要实现这个，只需简单地`url.NewData()`给 `Data` 参数。你的数据字典在发出请求时会自动编码为表单形式：

```go
data := url.NewData()
data.Set("key1","value1")
data.Set("key2","value2")

req := url.NewRequest()
req.Data = data
r, err := requests.Post("http://www.baidu.com",req)
if err != nil {
    fmt.Println(err)
}

fmt.Println(r.Text())

...
"form": {
    "key1": "value1", 
    "key2": "value2"
}
...
```

或者：

```go
req := url.NewRequest()
// 跟url.ParseParams同一用法，只是名称不同而已。
// data := map[string]interface{}{
//    "page":  []interface{}{"1", 2},
//    "limit": []string{"20"},
//    "skip":  []int{5},
// }
// req.Data = url.ParseData(data)
req.Data = url.ParseData("key1=value1&key2=value2")
r, err := requests.Post("http://www.baidu.com",req)
if err != nil {
    fmt.Println(err)
}

fmt.Println(r.Text())

...
"form": {
    "key1": "value1", 
    "key2": "value2"
}
...
```

你还可以为 `data`在表单中多个元素使用同一 key 的时候，这种方式尤其有效：

```go
data := url.NewData()
data.Set("key1","value1")
data.Add("key1","value3")
data.Set("key2","value2")
data.Add("key2","value4")

req := url.NewRequest()
req.Data = data
r, err := requests.Post("http://httpbin.org/post",req)
if err != nil {
    fmt.Println(err)
}
fmt.Println(r.Text())

...
"form": {
    "key1": [
      "value1", 
      "value3"
    ], 
    "key2": [
      "value2", 
      "value4"
    ]
}
...
```

很多时候你想要发送的数据并非编码为表单形式的。如果你想传递一个 `json` 数据那么用下面的方法。

你可以使用 `Json` 参数直接传递，然后它就会被自动编码(任何可被 json.Marshal 处理的类型)。

```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	req.Json = map[string]interface{}{ // 修正老版map[string]string
		"some":   "data",
		"name":   "测试",
		"colors": []string{"蓝色", "绿色", "紫色"},
		"data": map[string]interface{}{
			"json": true,
			"age":  15,
		},
	}
	r, err := requests.Post("http://httpbin.org/post", req)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r.Text())
}
```

示例输出：

```text
{
  "args": {}, 
  "data": "{\"colors\":[\"\u84dd\u8272\",\"\u7eff\u8272\",\"\u7d2b\u8272\"],\"data\":{\"age\":15,\"json\":true},\"name\":\"\u6d4b\u8bd5\",\"some\":\"data\"}", 
  "files": {}, 
  "form": {}, 
  "headers": {
    "Accept": "*/*", 
    "Accept-Encoding": "gzip, deflate, br", 
    "Content-Length": "99", 
    "Content-Type": "application/json", 
    "Host": "httpbin.org", 
    "User-Agent": "golang-requests 1.0", 
    "X-Amzn-Trace-Id": "Root=1-6241ac54-4345fed071127ed54d2ae83b"
  }, 
  "json": {
    "colors": [
      "\u84dd\u8272", 
      "\u7eff\u8272", 
      "\u7d2b\u8272"
    ], 
    "data": {
      "age": 15, 
      "json": true
    }, 
    "name": "\u6d4b\u8bd5", 
    "some": "data"
  }, 
  "origin": "220.249.16.210", 
  "url": "http://httpbin.org/post"
}
```



## POST一个多部分编码(Multipart-Encoded)的文件或FormData

requests 使得上传多部分编码文件变得很简单：

```go
files := url.NewFiles()
// SetFile(name,fileName,filePath,contentType)
// name为字段名，fileName为上传的文件名，filePath为上传文件的绝对路径，contentType为上传的文件类型
// 如果contentType设置为""，则默认为"application/octet-stream"
files.SetFile("api","api","D:\\Go\\github.com\\wangluozhe\\requests\\api.go","")
req := url.NewRequest()
req.Files = files
r, err := requests.Post("http://httpbin.org/post",req)
if err != nil {
    fmt.Println(err)
}
fmt.Println(r.Text())

...
"files": {
    "api": "文件内容"
}
...
```

requests使得FormData的使用也方便多了：

```go
files := url.NewFiles()
// SetField(name,value)
// name为字段名，value为值
files.SetField("name","value")
req := url.NewRequest()
req.Files = files
r, err := requests.Post("http://httpbin.org/post",req)
if err != nil {
    fmt.Println(err)
}
fmt.Println(r.Text())

...
"form": {
	"name": "value"
}
...
```

`AddFile`、`AddField` 可以追加多个同名项；`SetFile`、`SetField` 会覆盖同名项。同时设置 `Files` 和 `Data` 时，表单字段会合并到文件集合的副本，不会累积到原始 `Files`。

## POST发送text/plain文本

`Body` 支持 `string`、`[]byte` 和 `io.Reader`。未指定 Content-Type 时使用 `application/octet-stream`；发送文本时显式设置 `text/plain`：
```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	headers := url.NewHeaders()
	headers.Set("User-Agent", "123")
	headers.Set("Content-Type", "text/plain; charset=utf-8")
	req.Headers = headers
	req.Body = "testdata"
	r, err := requests.Post("https://httpbin.org/post", req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Close()
	fmt.Println(r.Text())
}
```



## 响应状态码

我们可以检测响应状态码：

```go
r, err := requests.Get("http://httpbin.org/get", nil)
fmt.Println(r.StatusCode)

200
```

为方便引用，可以直接使用此方法：

```go
fmt.Println(r.StatusCode == http.StatusOK)

True
```

如果发送了一个错误请求(一个 4XX 客户端错误，或者 5XX 服务器错误响应)，我们可以通过 `Response.RaiseForStatus()` 获取 Go `error`（不会抛异常）：

```go
r, err := requests.Get("http://httpbin.org/status/404", nil)
fmt.Println(r.StatusCode)

404

fmt.Println(r.RaiseForStatus())

404 Client Error
```

但是，由于我们的例子中 `r` 的 `StatusCode` 是 `200` ，当我们调用 `RaiseForStatus()` 时，得到的是：

```go
fmt.Println(r.RaiseForStatus())

nil
```

还可以使用 `r.Ok()` 判断状态码是否在 `[200, 400)` 范围内。`r.IsRedirect()` 和 `r.IsPermanentRedirect()` 同时检查重定向状态码和非空 `Location`，并不表示已经跟随了重定向。



## 响应头

我们可以查看以一个 http.Header形式（实际是一个map[string]\[]string类型）展示的服务器响应头：

```go
fmt.Println(r.Headers)

map[Access-Control-Allow-Credentials:[true] Access-Control-Allow-Origin:[*] Connection:[keep-alive] Content-Length:[1976] Content-Type:[application/json] Date:[Sat, 12 Mar 2022 15:59:05 GMT] Server:[gunicorn/19.9.0]]
```

但是这个类型比较特殊：它是仅为 HTTP 头部而生的。根据 [RFC 2616](http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html)， HTTP 头部是大小写不敏感的。

因此，我们可以使用任意大写形式来访问这些响应头字段：

```go
fmt.Println(r.Headers["Content-Type"][0])
application/json

fmt.Println(r.Headers.Get("content-type"))
application/json
```

它还有一个特殊点，那就是服务器可以多次接受同一 header，每次都使用不同的值。但 Requests 会将它们合并，这样它们就可以用一个映射来表示出来，参见 [RFC 7230](http://tools.ietf.org/html/rfc7230#section-3.2):

> A recipient MAY combine multiple header fields with the same field name into one "field-name: field-value" pair, without changing the semantics of the message, by appending each subsequent field value to the combined field value in order, separated by a comma.
>
> 接收者可以合并多个相同名称的 header 栏位，把它们合为一个 "field-name: field-value" 配对，将每个后续的栏位值依次追加到合并的栏位值中，用逗号隔开即可，这样做不会改变信息的语义。



## Cookie

如果某个响应中包含一些 cookie，你可以快速访问它们：

```go
url := "https://www.baidu.com"
r, err := requests.Get(url, nil)

fmt.Println(r.Cookies)

[BD_NOT_HTTPS=1; Path=/; Max-Age=300 BIDUPSID=7DB59A8D47E763943295969C33979837; Path=/; Domain=baidu.com; Max-Age=2147483647 PSTM=1647233990; Path=/; Domain=baidu.com; Max-Age=2147483647 BAIDUID=7DB59A8D47E7639495833AF6370F9985:FG=1; Path=/; Domain=baidu.com; Max-Age=31536000]
```

要想发送你的cookies到服务器，可以使用 `cookies` 参数：

```go
req := url.NewRequest()
cookies,_ := cookiejar.New(nil)
// cookies := url.NewCookies() // 推荐使用这种
urls, _ := url.Parse("http://httpbin.org/cookies")
cookies.SetCookies(urls,[]*http.Cookie{&http.Cookie{
    Name:       "cookies_are",
    Value:      "working",
}})
req.Cookies = cookies
r, err := requests.Get("http://httpbin.org/cookies", req)
if err != nil {
    fmt.Println(err)
}

fmt.Println(r.Text())

{"cookies": {"cookies_are": "working"}}
```

或者：

```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	rawUrl := "http://httpbin.org/cookies"
	req := url.NewRequest()
	//cookies := map[string]string{
	//	"name":     "zhangsan",
	//	"age":      "10",
	//	"language": "en",
	//}
	//cookies := map[string]int{
	//	"age":   20,
	//	"page":  2,
	//	"limit": 10,
	//}
	cookies := map[string]interface{}{
		"Hm_lpvt_def79de877408c7bd826e49b694147bc": 1648301329,
		"Hm_lvt_def79de877408c7bd826e49b694147bc":  "1647245863,1647936048,1648296630",
		"_ga": "GA1.1.630251354.1645893020",
	}
	req.Cookies = url.ParseCookies(rawUrl, cookies)
	//req.Cookies = url.ParseCookies(rawUrl,"_ga=GA1.1.630251354.1645893020; Hm_lvt_def79de877408c7bd826e49b694147bc=1647245863,1647936048,1648296630; Hm_lpvt_def79de877408c7bd826e49b694147bc=1648301329")
	r, err := requests.Get(rawUrl, req)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r.Text())
}
```

示例输出：

```text
{
  "cookies": {
    "Hm_lpvt_def79de877408c7bd826e49b694147bc": "1648301329", 
    "Hm_lvt_def79de877408c7bd826e49b694147bc": "1647245863,1647936048,1648296630", 
    "_ga": "GA1.1.630251354.1645893020"
  }
}
```



## 重定向与请求历史

默认允许重定向，`req.AllowRedirects = url.Bool(false)` 可禁用。Session 的 `MaxRedirects` 默认是 30。

`r.History` 是响应字段，保存从最早到最近的中间响应，可用于检查跳转过程：

```go
r, err := requests.Get("http://github.com", nil)
if err != nil {
    return
}
defer r.Close()
fmt.Println(r.Url)
for _, previous := range r.History {
    fmt.Println(previous.StatusCode, previous.Url)
}
```

当前顶层 `requests.Head()` 会强制关闭重定向；`Session.Head()` 则遵循一般重定向设置，默认允许。需要通过顶层 API 发起跟随重定向的 HEAD 请求时，使用：

```go
req := &url.Request{AllowRedirects: url.Bool(true)}
r, err := requests.Request("HEAD", "http://github.com", req)
if err != nil {
    return
}
defer r.Close()
fmt.Println(r.Url)
```

## 超时

Go API 的 `Timeout` 类型为 `time.Duration`，例如 `5 * time.Second`。新建 Session 默认 30 秒，`req.Timeout == 0` 表示继承 Session，不表示无限等待。

超时覆盖连接、重定向和响应体读取。流式请求返回后，计时仍然有效，后续读取 Body 也可能返回超时错误；它不是每次读取后重新计时的空闲超时。

```go
package main

import (
	"fmt"
	"time"

	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := &url.Request{Timeout: 5 * time.Second}
	r, err := requests.Get("https://example.com", req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Close()
	fmt.Println(r.StatusCode)
}
```

将 `session.Timeout = 0` 可关闭 Client 的总时限，但拨号、TLS 握手和响应头仍有各自默认的 30 秒超时。请求返回错误后，应先处理错误，再访问响应。


## 基本身份认证

许多要求身份认证的web服务都接受 HTTP Basic Auth。这是最简单的一种身份认证，并且 requests 对这种认证方式的支持是直接开箱即可用。

以 HTTP Basic Auth 发送请求非常简单：

```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	req.Auth = []string{"user", "password"}
	r, err := requests.Get("http://httpbin.org/basic-auth/user/password", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("text:", r.Text())
}
```

示例输出：

```text
{
  "authenticated": true, 
  "user": "user"
}
```



## 客户端证书

你也可以指定一个本地证书用作客户端证书，可以是一个包含两个文件路径的数组（cert，key）或一个包含三个文件路径的数组（cert，key，根证书）：

```go
req := url.NewRequest()
req.Cert = []string{"cert","key"}
// req.Cert = []string{"cert","key","rootca"}
r, err := requests.Get("xxx",req)
if err != nil{
    fmt.Println(err)
}
```

或者保持在会话中：

```go
session := requests.NewSession()
session.Cert = []string{"cert","key"}
```




## Hook 功能 (Hooks)

支持在请求参数解析阶段注入自定义逻辑。如果您需要对传入的 `any` 类型参数进行特殊的预处理，可以通过设置 Hook 函数来实现。

**支持的 Hook 类型:**

* `ParamsHook`
* `HeadersHook`
* `CookiesHook`
* `DataHook`
* `FilesHook`
* `BodyHook`

**示例 (Example):**

```go
req := &url.Request{
    Params: "custom_format_string",
    // 自定义 Params 解析逻辑
    ParamsHook: func(p any) *url.Params {
        if s, ok := p.(string); ok {
            // 假设只需处理特殊字符串
            params := url.NewParams()
            params.Set("key", s)
            return params
        }
        return url.ParseParams(p) // 可以使用ParseParams进行兜底，Files和Body没有Parse，但可以自己手动创建
    },
}

```



## Session 中间件 (Session Middleware)

Session 级别支持中间件（Middleware），允许您在请求发送前后拦截并处理请求。这采用了经典的“洋葱模型”，适用于日志记录、鉴权、统一错误处理等场景。

**定义中间件:**

中间件函数的类型定义为：`type Middleware func(next Handler) Handler`

**示例：日志与鉴权中间件 (Example: Logging & Auth)**

```go
package main

import (
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/models"
	"github.com/wangluozhe/requests/url"
	"log"
	"time"
)

// 1. 定义日志中间件
func LoggingMiddleware(next requests.Handler) requests.Handler {
	return func(preq *models.PrepareRequest, req *url.Request) (*models.Response, error) {
		start := time.Now()
		log.Printf("[Middleware] Start Request: %s %s", preq.Method, preq.Url)

		// 执行下一个 Handler
		resp, err := next(preq, req)

		duration := time.Since(start)
		log.Printf("[Middleware] End Request. Duration: %v", duration)
		return resp, err
	}
}

// 2. 定义鉴权中间件
func AuthMiddleware(token string) requests.Middleware {
	return func(next requests.Handler) requests.Handler {
		return func(preq *models.PrepareRequest, req *url.Request) (*models.Response, error) {
			// 自动添加 Token 到 Header
			if preq.Headers != nil {
				preq.Headers.Set("Authorization", "Bearer "+token)
			}
			return next(preq, req)
		}
	}
}

func main() {
	// 创建 Session
	s := requests.NewSession()

	// 注册中间件（注意顺序：先注册的先执行）
	s.Use(LoggingMiddleware)
	s.Use(AuthMiddleware("my-secret-token"))

	// 发送请求
	// 该请求会自动经过 LoggingMiddleware 和 AuthMiddleware
	resp, err := s.Get("https://httpbin.org/get", nil)
	if err != nil {
		panic(err)
	}
	println(resp.Text())
}
```



## 协议选择

当前版本无指纹配置时默认使用 HTTP/1.1。以下设置会准备 HTTP/2 Transport：`Ja3`、`ClientHelloHexStream`、`ClientHelloID`、旧的 `TLSExtensions.ClientHelloHexStream`，或非 `nil` 的 `HTTP2Settings`。单独设置其他 TLS 扩展不等于开启 HTTP/2。

如果只想启用 HTTP/2 支持、使用底层默认参数，可以传入空配置：

```go
// http 为 github.com/wangluozhe/chttp，不是标准库 net/http。
req := &url.Request{HTTP2Settings: &http.HTTP2Settings{}}
r, err := requests.Get("https://example.com", req)
```

HTTPS 是否实际使用 HTTP/2 取决于 ClientHello 的 ALPN 和服务端协商结果。没有协商到 h2 时可使用 HTTP/1.1，这不是任意请求失败后自动重试。设置 `req.ForceHTTP1 = url.Bool(true)` 会强制使用 HTTP/1.1。

Header、Cookie、TLS 和 HTTP/2 配置类型来自 `github.com/wangluozhe/chttp` 及其子包；需要底层类型时，不要直接混用 `net/http` 中的同名类型。


## JA3指纹

requests也支持JA3指纹的修改，可以让你在访问的时候使用你自己定义的JA3指纹进行TLS握手访问，但是请注意，JA3指纹必须符合要求，不能随便更改，最好使用wireshark或者ja3er.com获取标准指纹，而不是随便输入一串数字。

```go
req := url.NewRequest()
headers := url.NewHeaders()
headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
req.Headers = headers
req.Ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-21,29-23-24,0"
r, err := requests.Get("https://ja3er.com/json", req)
if err != nil {
	fmt.Println(err)
}
fmt.Println(r.Text())

{"ja3_hash":"b32309a26951912be7dba376398abc3b", "transport": "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-21,29-23-24,0", "User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"}
```

## ClientHello 预设与抓包

除了 JA3，还可以选择 uTLS 内置预设。预设控制 TLS ClientHello，不会自动配置相应浏览器的 User-Agent、普通 Header 或 HTTP/2 指纹；这些参数仍需按需单独设置。

```go
package main

import (
	"fmt"

	utls "github.com/refraction-networking/utls"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	id := utls.HelloFirefox_148
	req := &url.Request{ClientHelloID: &id}
	r, err := requests.Get("https://tls.peet.ws/api/all", req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Close()
	fmt.Println(r.Text())
}
```

`ClientHelloHexStream` 接受 TLS ClientHello 记录的十六进制字符串，不是 JA3 字符串或整个 pcap 文件。例如，将抓包提取的十六进制内容保存为 `clienthello.hex` 后读取：

```go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	data, err := os.ReadFile("clienthello.hex")
	if err != nil {
		fmt.Println(err)
		return
	}
	// 文件中仅包含十六进制字节，可用空白分隔，不包含偏移量或 0x 前缀。
	hello := strings.Join(strings.Fields(string(data)), "")
	req := &url.Request{ClientHelloHexStream: hello}
	r, err := requests.Get("https://tls.peet.ws/api/all", req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Close()
	fmt.Println(r.Text())
}
```

这些字段也可以设置在 Session 上。请求与 Session 合并后，选择优先级如下：

1. `TLSExtensions.ClientHelloHexStream`（兼容旧字段）。
2. `ClientHelloHexStream`。
3. `ClientHelloID`。
4. `Ja3`。

空字符串或 `nil` 会继承 Session。仅设置请求的 `Ja3` 不会清除从 Session 继承的更高优先级配置；切换模式时，应在没有并发请求的情况下调整 Session 配置，或使用独立 Session。


## TLS指纹

使用 `transport.Extensions` 配置 TLS 扩展，再通过 `ToTLSExtensionsChecked` 转换为 `chttp` 类型。扩展内容应与目标握手匹配，JA3 的扩展列表中也需要包含对应扩展。
```go
package main

import (
	"fmt"
	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/transport"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	headers := &http.Header{
		"User-Agent":                []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/112.0"},
		"accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
		"accept-language":           []string{"zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2"},
		"accept-encoding":           []string{"gzip, deflate, br"},
		"upgrade-insecure-requests": []string{"1"},
		"sec-fetch-dest":            []string{"document"},
		"sec-fetch-mode":            []string{"navigate"},
		"sec-fetch-site":            []string{"none"},
		"sec-fetch-user":            []string{"?1"},
		"te":                        []string{"trailers"},
		http.PHeaderOrderKey: []string{
			":method",
			":path",
			":authority",
			":scheme",
		},
		http.HeaderOrderKey: []string{
			"user-agent",
			"accept",
			"accept-language",
			"accept-encoding",
			"upgrade-insecure-requests",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"te",
		},
	}
	req.Headers = headers
	req.Ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-21,29-23-24,0"
	es := &transport.Extensions{
		SupportedSignatureAlgorithms: []string{
			"ECDSAWithP256AndSHA256",
			"ECDSAWithP384AndSHA384",
			"ECDSAWithP521AndSHA512",
			"PSSWithSHA256",
			"PSSWithSHA384",
			"PSSWithSHA512",
			"PKCS1WithSHA256",
			"PKCS1WithSHA384",
			"PKCS1WithSHA512",
			"ECDSAWithSHA1",
			"PKCS1WithSHA1",
		},
		//CertCompressionAlgo: []string{
		//	"brotli",
		//},
		RecordSizeLimit: 4001,
		DelegatedCredentials: []string{
			"ECDSAWithP256AndSHA256",
			"ECDSAWithP384AndSHA384",
			"ECDSAWithP521AndSHA512",
			"ECDSAWithSHA1",
		},
		SupportedVersions: []string{
			"1.3",
			"1.2",
		},
		PSKKeyExchangeModes: []string{
			"PskModeDHE",
		},
		KeyShareCurves: []string{
			"X25519",
			"P256",
		},
	}
	tes, err := transport.ToTLSExtensionsChecked(es)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.TLSExtensions = tes
	r, err := requests.Get("https://tls.peet.ws/api/all", req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Close()
	fmt.Println(r.Request.Headers)
	fmt.Println("url:", r.Url)
	fmt.Println("headers:", r.Headers)
	fmt.Println("text:", r.Text())
}
```

### TLS 扩展字段与校验

除上例中的字段外，还支持：

| 字段 | 用法 |
| --- | --- |
| `SignatureAlgorithmsCert` | 证书签名算法列表，用于 `signature_algorithms_cert` 扩展 |
| `TrustAnchors` | `[][]byte`，用于编号 51764 的 TLS 扩展 ID 列表；不是根 CA 证书配置 |
| `NotUsedGREASE` | 为 `true` 时禁用 GREASE |
| `KeyShareCurves` | 支持 `GREASE`、`P256`、`P384`、`P521`、`X25519` 及受支持的数字字符串 |

例如，以下配置可用于补充 JA3 握手中的相应扩展：

```go
extensions, err := transport.ToTLSExtensionsChecked(&transport.Extensions{
    SignatureAlgorithmsCert: []string{"rsa_pss_rsae_sha256", "0x0403"},
    KeyShareCurves: []string{"4588", "X25519"}, // X25519MLKEM768、X25519
    TrustAnchors: [][]byte{{0x01, 0x02}},
    NotUsedGREASE: true,
})
if err != nil {
    return
}
req.TLSExtensions = extensions
```

常用后量子 KeyShare 编号包括 `"4588"`（X25519MLKEM768）、`"4587"`（SecP256r1MLKEM768）、`"4589"`（SecP384r1MLKEM1024）、`"514"`（MLKEM1024）。这里应传数字字符串，不要直接传这些算法名称。

签名算法字段接受支持的名称或十六进制编号（可带 `0x`）。`RecordSizeLimit` 为 0 时使用默认值，显式值必须在 64～16384 之间。`TrustAnchors` 为 `nil` 时使用默认配置，非 `nil` 的空切片表示明确发送空列表；单项长度为 1～255 字节，含长度前缀的列表总长不超过 65533 字节。JSON 中 `[]byte` 使用 Base64，例如 `"TrustAnchors": ["AQI="]`。

新代码建议使用 `ToTLSExtensionsChecked` 和 `ToHTTP2SettingsChecked`，处理转换错误。兼容旧接口的 `ToTLSExtensions` 在非法输入时返回空扩展配置；`ToHTTP2Settings` 会跳过非法条目，并对无效的可选字段回退默认值。旧接口不会提供同样的错误反馈。


## HTTP2指纹

通过 `transport.H2Settings` 配置 SETTINGS、连接窗口增量、首个 HEADERS 流及优先级帧，再用 `ToHTTP2SettingsChecked` 转换。请求和 Session 都使用已有的 `HTTP2Settings` 字段。

```go
package main

import (
	"fmt"
	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/transport"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	headers := &http.Header{
		"User-Agent":                []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/112.0"},
		"accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
		"accept-language":           []string{"zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2"},
		"accept-encoding":           []string{"gzip, deflate, br"},
		"upgrade-insecure-requests": []string{"1"},
		"sec-fetch-dest":            []string{"document"},
		"sec-fetch-mode":            []string{"navigate"},
		"sec-fetch-site":            []string{"none"},
		"sec-fetch-user":            []string{"?1"},
		"te":                        []string{"trailers"},
		http.PHeaderOrderKey: []string{
			":method",
			":path",
			":authority",
			":scheme",
		},
		http.HeaderOrderKey: []string{
			"user-agent",
			"accept",
			"accept-language",
			"accept-encoding",
			"upgrade-insecure-requests",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"te",
		},
	}
	req.Headers = headers
	req.Ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-21,29-23-24,0"
	h2s := &transport.H2Settings{
		Settings: map[string]int{
			"HEADER_TABLE_SIZE": 65536,
			//"ENABLE_PUSH":            0,
			//"MAX_HEADER_LIST_SIZE":   262144,
			//"MAX_CONCURRENT_STREAMS": 1000,
			"INITIAL_WINDOW_SIZE": 131072,
			"MAX_FRAME_SIZE":      16384,
		},
		SettingsOrder: []string{
			"HEADER_TABLE_SIZE",
			"INITIAL_WINDOW_SIZE",
			"MAX_FRAME_SIZE",
		},
		ConnectionFlow: 12517377,
		HeaderPriority: map[string]interface{}{
			"weight":    42,
			"streamDep": 13,
			"exclusive": false,
		},
		PriorityFrames: []map[string]interface{}{
			{
				"streamID": 3,
				"priorityParam": map[string]interface{}{
					"weight":    201,
					"streamDep": 0,
					"exclusive": false,
				},
			},
			{
				"streamID": 5,
				"priorityParam": map[string]interface{}{
					"weight":    101,
					"streamDep": 0,
					"exclusive": false,
				},
			},
			{
				"streamID": 7,
				"priorityParam": map[string]interface{}{
					"weight":    1,
					"streamDep": 0,
					"exclusive": false,
				},
			},
			{
				"streamID": 9,
				"priorityParam": map[string]interface{}{
					"weight":    1,
					"streamDep": 7,
					"exclusive": false,
				},
			},
			{
				"streamID": 11,
				"priorityParam": map[string]interface{}{
					"weight":    1,
					"streamDep": 3,
					"exclusive": false,
				},
			},
			{
				"streamID": 13,
				"priorityParam": map[string]interface{}{
					"weight":    241,
					"streamDep": 0,
					"exclusive": false,
				},
			},
		},
	}
	h2ss, err := transport.ToHTTP2SettingsChecked(h2s)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.HTTP2Settings = h2ss
	r, err := requests.Get("https://tls.peet.ws/api/all", req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Close()
	fmt.Println(r.Request.Headers)
	fmt.Println("url:", r.Url)
	fmt.Println("headers:", r.Headers)
	fmt.Println("text:", r.Text())
}
```


### HTTP/2 配置规则

- `SettingsOrder` 非空时，只发送该列表选中的 SETTINGS，保留顺序和重复项；对应名称必须存在于 `Settings`。省略时，按数字 SETTINGS ID 排序。
- 除常见 SETTINGS 名称外，支持 `ENABLE_CONNECT_PROTOCOL`、`NO_RFC7540_PRIORITIES`，以及数字 ID 字符串或 `UNKNOWN_SETTING_数字`。
- `ConnectionFlow` 是在初始 65535 字节连接窗口上增加的额度，不是最终窗口大小；0 使用底层默认值，合法范围为 0～2147418112。
- `HeadersID` 为 0 时使用默认值，非零时必须是 31 位奇数。`PriorityFrames` 也可能影响首个实际请求的流 ID。
- `HeaderPriority` 和 `PriorityFrames` 中的 `weight` 为 0～255 的编码值，对应协议权重 1～256。优先级依赖、流 ID、缺失字段和错误类型都会由 Checked 转换器校验。

非空的自定义 `Settings` 列表会替换底层默认列表，不自动补齐未列出的 SETTINGS。特别是没有发送 `MAX_FRAME_SIZE` 时，对端按协议默认的 16 KiB 帧大小发送数据，这可能增加大响应的处理开销。调整此值也会改变 HTTP/2 指纹，应按目标指纹选择，不能只为性能随意添加。


## JA4指纹

`JA4`是什么，怎么组成的，请看华总的文章[JA4概要](https://blog.csdn.net/Y_morph/article/details/133747866?spm=1001.2014.3001.5501)

```golang
package main

import (
	"fmt"
	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/transport"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	headers := &http.Header{
		"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"Accept-Encoding":           []string{"gzip, deflate, br"},
		"Accept-Language":           []string{"zh-CN,zh;q=0.9"},
		"Cache-Control":             []string{"no-cache"},
		"Connection":                []string{"keep-alive"},
		"Host":                      []string{"gospider2.gospiderb.asia:8998"},
		"Pragma":                    []string{"no-cache"},
		"sec-ch-ua":                 []string{"\".Not/A)Brand\";v=\"99\", \"Google Chrome\";v=\"103\", \"Chromium\";v=\"103\""},
		"sec-ch-ua-mobile":          []string{"\"?0\""},
		"sec-ch-ua-platform":        []string{"\"Windows\""},
		"Sec-Fetch-Dest":            []string{"document"},
		"Sec-Fetch-Mode":            []string{"navigate"},
		"Sec-Fetch-Site":            []string{"none"},
		"Sec-Fetch-User":            []string{"\"?1\""},
		"Upgrade-Insecure-Requests": []string{"1"},
		"User-Agent":                []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: []string{
			"HOST",
			"connection",
			"pragma",
			"Cache-Control",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"upgrade-insecure-requests",
			"user-agent",
			"accept",
			"sec-fetch-site",
			"sec-fetch-mode",
			"sec-fetch-user",
			"sec-fetch-dest",
			"accept-encoding",
			"accept-language",
		},
		http.UnChangedHeaderKey: []string{
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
		},
	}
	req.Headers = headers
	req.Ja3 = "772,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21-41,29-23-24,0"
	req.ForceHTTP1 = url.Bool(true)
	es := &transport.Extensions{
		SupportedSignatureAlgorithms: []string{
			"ecdsa_secp256r1_sha256",
			"rsa_pss_rsae_sha256",
			"rsa_pkcs1_sha256",
			"ecdsa_secp384r1_sha384",
			"rsa_pss_rsae_sha384",
			"rsa_pkcs1_sha384",
			"rsa_pss_rsae_sha512",
			"rsa_pkcs1_sha512",
		},
		CertCompressionAlgo: []string{
			"brotli",
		},
		RecordSizeLimit: 4001,
		DelegatedCredentials: []string{
			"ECDSAWithP256AndSHA256",
			"ECDSAWithP384AndSHA384",
			"ECDSAWithP521AndSHA512",
			"ECDSAWithSHA1",
		},
		SupportedVersions: []string{
			"GREASE",
			"1.3",
			"1.2",
		},
		PSKKeyExchangeModes: []string{
			"PskModeDHE",
		},
		KeyShareCurves: []string{
			"GREASE",
			"X25519",
		},
	}
	tes, err := transport.ToTLSExtensionsChecked(es)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.TLSExtensions = tes
	r, err := requests.Get("https://gospider2.gospiderb.asia:8998/", req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer r.Close()
	fmt.Println(r.Request.Headers)
	fmt.Println("url:", r.Url)
	fmt.Println("headers:", r.Headers)
	fmt.Println("text:", r.Text())
}

// 模拟后的结果
// JA4:  t13d1516h1_dea800f94266_27c5da80acb3
// JA4H: ge11nn13zhcn_d8f538a17def_e3b0c44298fc_e3b0c44298fc
```


## C API

C 导出层位于 `libs_dist`。安装 Go 和可用的 C 编译器、启用 CGO 后，可构建共享库及配套头文件：

```bash
# Windows
go build -buildmode=c-shared -o requests.dll ./libs_dist
# Linux
go build -buildmode=c-shared -o librequests.so ./libs_dist
```

### JSON 请求参数

`request` 和 `stream_request` 接收 UTF-8 JSON 字符串。`Method`、`Url` 必填，建议显式指定 `Id` 作为会话 ID；同一 ID 复用 Cookie 和连接，不同 ID 隔离会话。省略 `Id` 的调用会共用空字符串 ID 对应的 Session。

```json
{
  "Id": "account-1",
  "Method": "GET",
  "Url": "https://tls.peet.ws/api/all",
  "Timeout": 10,
  "ClientHelloID": {"Client": "Firefox", "Version": "148"},
  "HTTP2Settings": "{\"Settings\":{\"HEADER_TABLE_SIZE\":65536,\"ENABLE_PUSH\":0},\"SettingsOrder\":[\"HEADER_TABLE_SIZE\",\"ENABLE_PUSH\"]}"
}
```

C API 的类型与 Go API 并不完全相同：

| 字段 | C API JSON 格式 |
| --- | --- |
| `Params`、`Headers`、`Cookies`、`Data` | 字符串键值对象，例如 `{"page":"1"}`；Params 同样需要预编码 |
| `Json` | JSON 对象 |
| `Body` | Base64 字符串，不是明文；例如 `hello` 为 `aGVsbG8=` |
| `Timeout` | 整数秒；0 或省略时继承默认超时 |
| `AllowRedirects`、`Verify` | 省略时默认开启，显式 `false` 表示关闭 |
| `RandomJA3`、`ForceHTTP1` | 可选布尔值，默认关闭 |
| `Auth`、`Cert` | 字符串数组，分别用于账号密码和证书文件路径 |
| `Proxies`、`Ja3`、`ClientHelloHexStream` | 字符串 |
| `ClientHelloID` | uTLS 预设对象，例如 `{"Client":"Firefox","Version":"148"}` |
| `HeadersOrder`、`PseudoHeaderOrder`、`UnChangedHeaderKey` | 字符串数组；前者为普通头顺序，后两者为伪头顺序和保留大小写的名称 |
| `TLSExtensions`、`HTTP2Settings` | **内层 JSON 序列化后的字符串**，不是直接嵌套对象；字段分别对应 `transport.Extensions`、`transport.H2Settings` |

设置 `HeadersOrder` 或 `UnChangedHeaderKey` 时也应提供 `Headers` 对象。TLS/HTTP2 转换使用 Checked 校验，格式或范围错误会返回 `err`。仅启用默认 HTTP/2 参数时，可传 `"HTTP2Settings": "{}"`。

C API 没有暴露 `Files`、`ResponseRetention` 字段。普通 `request` 默认保留完整 `raw`；需要流式读取时请调用 `stream_request`，不要只给普通 `request` 设置 `Stream: true`。

### 返回值与释放

返回值都是指向 JSON 字符串的指针。先复制所需 JSON 和 ID 到调用方内存，再调用释放函数；不要在释放后继续读取原指针或其中的 ID。

| 调用结果 | JSON 字段 | 字符串释放方式 |
| --- | --- | --- |
| `request` 成功 | `id`、`url`、`headers`、`cookies`、`status_code`、`content`、`raw` | `freeMemory(id)` |
| `stream_request` 成功 | `stream_id`、`url`、`headers`、`cookies`、`status_code` | `freeMemory(stream_id)`，或随 `stream_close(stream_id)` 一并释放 |
| `stream_read` 成功 | `data`、`eof`，没有 `id` | `freeMemory(stream_id + "_read")`，或由下次同流读取 / 关闭自动释放 |
| 上述任意调用失败 | `id`、`err` | `freeMemory(id)` |

`content`、`raw`、`data` 均为 Base64 字符串。`freeMemory` 只释放返回字符串；`stream_close` 关闭响应流，并释放该流的元信息和最近一次读取字符串；`freeSession` 关闭会话和关联流。普通请求返回的字符串仍需按各自 `id` 释放。

下面是 Python `ctypes` 调用普通请求的完整示例：

```python
import base64
import ctypes
import json
from pathlib import Path

# Linux 改为 librequests.so。
lib = ctypes.CDLL(str(Path("requests.dll").resolve()))
lib.request.argtypes = [ctypes.c_char_p]
lib.request.restype = ctypes.c_void_p
lib.freeMemory.argtypes = [ctypes.c_char_p]
lib.freeMemory.restype = None
lib.freeSession.argtypes = [ctypes.c_char_p]
lib.freeSession.restype = None

session_id = "example-session"
params = {"Id": session_id, "Method": "GET", "Url": "https://example.com"}
try:
    pointer = lib.request(json.dumps(params).encode("utf-8"))
    if not pointer:
        raise RuntimeError("request 返回空指针")
    result = json.loads(ctypes.string_at(pointer).decode("utf-8"))
    # result 已在 Python 内存中，可以释放 C 字符串。
    lib.freeMemory(result["id"].encode("utf-8"))
    if "err" in result:
        raise RuntimeError(result["err"])
    print(result["status_code"])
    print(base64.b64decode(result["content"]))
finally:
    lib.freeSession(session_id.encode("utf-8"))
```

流式调用顺序：

1. 调用 `stream_request`，复制返回 JSON 并保存 `stream_id`。
2. 循环调用 `stream_read(stream_id, size)`。`size <= 0` 使用 4096 字节；每次先复制、解析返回值，再进行下一次读取。同一流不要并发调用 `stream_read`。
3. 解码并消费 `data`，再检查 `eof`。即使 `eof == true`，当前结果仍可能包含数据。失败时根据返回的 `id` 释放错误字符串。
4. 无论读完还是提前结束，都调用 `stream_close(stream_id)`；会话不再使用时调用 `freeSession(session_id)`。

流式读取沿用 `Stream=true` 的行为，本库不会额外解压响应流；需要解压时由调用方根据响应头处理。

后台每分钟扫描一次：流空闲 5 分钟后可被关闭，Session 空闲 30 分钟后可被回收，返回字符串存活 30 分钟后可被释放。托管字符串总量上限为 4096，超出时提前淘汰最旧项。这些是兜底回收规则，不能替代主动释放，也不能作为返回指针的有效期保证。`setDebug(bool)` 可控制与 Go API 相同的全局调试输出。


# 编码

requests支持一些常用的编码方式，并且命名更使人易懂。传递的参数可为`字符串（string）`或`字符数组（[]byte）`类型。



## 16进制编码

```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	url := "https://www.baidu.com"
	hexen := utils.HexEncode(url)
	fmt.Println(hexen)
	fmt.Println(string(hexen))
	hexde := utils.HexDecode(hexen)
	fmt.Println(hexde)
	fmt.Println(string(hexde))
}
```

示例输出：

```text
[54 56 55 52 55 52 55 48 55 51 51 97 50 102 50 102 55 55 55 55 55 55 50 101 54 50 54 49 54 57 54 52 55 53 50 101 54 51 54 102 54 100]
68747470733a2f2f7777772e62616964752e636f6d
[104 116 116 112 115 58 47 47 119 119 119 46 98 97 105 100 117 46 99 111 109]
https://www.baidu.com
```



## URI编码

```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	url := "https://www.baidu.com?page=10&abc=123&name=你好啊"
	encode := utils.EncodeURIComponent(url)
	fmt.Println(encode)
	decode := utils.DecodeURIComponent(encode)
	fmt.Println(decode)
	encode1 := utils.EncodeURI(url)
	fmt.Println(encode1)
	decode1 := utils.DecodeURI(encode1)
	fmt.Println(decode1)
	escape := utils.Escape(url)
	fmt.Println(escape)
	unescape := utils.UnEscape(escape)
	fmt.Println(unescape)
}
```

示例输出：

```text
https%3A%2F%2Fwww.baidu.com%3Fpage%3D10%26abc%3D123%26name%3D%E4%BD%A0%E5%A5%BD%E5%95%8A
https://www.baidu.com?page=10&abc=123&name=你好啊
https://www.baidu.com?page=10&abc=123&name=%E4%BD%A0%E5%A5%BD%E5%95%8A
https://www.baidu.com?page=10&abc=123&name=你好啊
https%3A//www.baidu.com%3Fpage%3D10%26abc%3D123%26name%3D%u4f60%u597d%u554a
https://www.baidu.com?page=10&abc=123&name=你好啊
```



## Base32 和 Base64 编码

```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	url := "https://www.baidu.com?page=10&abc=123&name=你好啊"
	base32en := utils.Base32Encode(url)
	fmt.Println(base32en)
	base32de := utils.Base32Decode(base32en)
	fmt.Println(base32de)
	base64en := utils.Base64Encode(url)
	fmt.Println(base64en)
	base64de := utils.Base64Decode(base64en)
	fmt.Println(base64de)
	// 或者像JavaScript语言一样
	btoa := utils.Btoa(url)
	fmt.Println(btoa)
	atob := utils.Atob(btoa)
	fmt.Println(atob)
}
```

示例输出：

```text
NB2HI4DTHIXS653XO4XGEYLJMR2S4Y3PNU7XAYLHMU6TCMBGMFRGGPJRGIZSM3TBNVST3ZF5UDS2LPPFSWFA====
https://www.baidu.com?page=10&abc=123&name=你好啊
aHR0cHM6Ly93d3cuYmFpZHUuY29tP3BhZ2U9MTAmYWJjPTEyMyZuYW1lPeS9oOWlveWVig==
https://www.baidu.com?page=10&abc=123&name=你好啊
aHR0cHM6Ly93d3cuYmFpZHUuY29tP3BhZ2U9MTAmYWJjPTEyMyZuYW1lPeS9oOWlveWVig==
https://www.baidu.com?page=10&abc=123&name=你好啊
```



二进制数据可使用 `utils.Base64DecodeToBytes` 或 `utils.Base32DecodeToBytes` 解码为 `[]byte`，避免先转成字符串。

# 加密算法

requests支持一些常用的加密算法，并且命名更使人易懂。传递的参数可为`字符串（string）`或`字符数组（[]byte）`类型。



## 线性散列（Hash）算法

`MD系列`：

```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	m := utils.MD4("123")
	fmt.Println("MD4:", m)

	m = utils.RIPEMD160("123")
	fmt.Println("RIPEMD160:", m)

	m = utils.MD5("123")
	fmt.Println("MD5:", m)
}
```

示例输出：

```text
MD4: c58cda49f00748a3bc0fcfa511d516cb
RIPEMD160: e3431a8e0adbf96fd140103dc6f63a3f8fa343ab
MD5: 202cb962ac59075b964b07152d234b70
```

`SHA系列`：

```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	s1 := utils.SHA1("123")
	b64 := utils.Btoa(s1)
	h16 := utils.HexEncode(s1)
	fmt.Println("SHA1-base64:", b64)
	fmt.Println("SHA1-hex:", string(h16))

	s2 := utils.SHA224("123")
	b64 = utils.Btoa(s2)
	h16 = utils.HexEncode(s2)
	fmt.Println("SHA224-base64:", b64)
	fmt.Println("SHA224-hex:", string(h16))

	s2 = utils.SHA256("123")
	b64 = utils.Btoa(s2)
	h16 = utils.HexEncode(s2)
	fmt.Println("SHA256-base64:", b64)
	fmt.Println("SHA256-hex:", string(h16))

	s5 := utils.SHA384("123")
	b64 = utils.Btoa(s5)
	h16 = utils.HexEncode(s5)
	fmt.Println("SHA384-base64:", b64)
	fmt.Println("SHA384-hex:", string(h16))

	s5 = utils.SHA512("123")
	b64 = utils.Btoa(s5)
	h16 = utils.HexEncode(s5)
	fmt.Println("SHA512-base64:", b64)
	fmt.Println("SHA512-hex:", string(h16))
}
```

示例输出：

```text
SHA1-base64: QL0AFWMIX8NRZTKeof9cXsvbvu8=
SHA1-hex: 40bd001563085fc35165329ea1ff5c5ecbdbbeef
SHA224-base64: eNgEXWhKvS7s6SN1jzzXgUid86SOEniYJGYBfw==
SHA224-hex: 78d8045d684abd2eece923758f3cd781489df3a48e1278982466017f
SHA256-base64: pmWkWSBCL51Bfkhn79xPuKBKHz//H6B+mY6G9/eieuM=
SHA256-hex: a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3
SHA384-base64: mgqC8MDPMUcNev/t40BsyaqEEGcVILcnBE7aFbTCVTKptc2Kr5zsSRnXYlW2v7AP
SHA384-hex: 9a0a82f0c0cf31470d7affede3406cc9aa8410671520b727044eda15b4c25532a9b5cd8aaf9cec4919d76255b6bfb00f
SHA512-base64: PJkJr+wlNU1VHa4hWQuybjjVPyFzuNPcPu5MBH56scHri4UQPjvnumE7MbtcnDYhTcnxSkL9ei/bhIVrylxEwg==
SHA512-hex: 3c9909afec25354d551dae21590bb26e38d53f2173b8d3dc3eee4c047e7ab1c1eb8b85103e3be7ba613b31bb5c9c36214dc9f14a42fd7a2fdb84856bca5c44c2
```

`Hmac系列`：

```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	md4 := utils.HmacMD4("123", "123")
	bs64 := utils.Btoa(md4)
	hex := utils.HexEncode(md4)
	fmt.Println("HmacMD4-base64:", bs64)
	fmt.Println("HmacMD4-hex:", string(hex))

	r160 := utils.HmacRIPEMD160("123", "123")
	bs64 = utils.Btoa(r160)
	hex = utils.HexEncode(r160)
	fmt.Println("HmacRIPEMD160-base64:", bs64)
	fmt.Println("HmacRIPEMD160-hex:", string(hex))

	md5 := utils.HmacMD5("123", "123")
	bs64 = utils.Btoa(md5)
	hex = utils.HexEncode(md5)
	fmt.Println("HmacMD5-base64:", bs64)
	fmt.Println("HmacMD5-hex:", string(hex))

	sha1 := utils.HmacSHA1("123", "123")
	bs64 = utils.Btoa(sha1)
	hex1 := utils.HexEncode(sha1)
	fmt.Println("HmacSHA1-base64:", bs64)
	fmt.Println("HmacSHA1-hex:", string(hex1))

	sha224 := utils.HmacSHA224("123", "123")
	bs64 = utils.Btoa(sha224)
	hex2 := utils.HexEncode(sha224)
	fmt.Println("HmacSHA224-base64:", bs64)
	fmt.Println("HmacSHA224-hex:", string(hex2))

	sha256 := utils.HmacSHA256("123", "123")
	bs64 = utils.Btoa(sha256)
	hex3 := utils.HexEncode(sha256)
	fmt.Println("HmacSHA256-base64:", bs64)
	fmt.Println("HmacSHA256-hex:", string(hex3))

	sha384 := utils.HmacSHA384("123", "123")
	bs64 = utils.Btoa(sha384)
	hex4 := utils.HexEncode(sha384)
	fmt.Println("HmacSHA384-base64:", bs64)
	fmt.Println("HmacSHA384-hex:", string(hex4))

	sha512 := utils.HmacSHA512("123", "123")
	bs64 = utils.Btoa(sha512)
	hex5 := utils.HexEncode(sha512)
	fmt.Println("HmacSHA512-base64:", bs64)
	fmt.Println("HmacSHA512-hex:", string(hex5))
}
```

示例输出：

```text
HmacMD4-base64: u3owGP+65If/IdHlG8klJg==
HmacMD4-hex: bb7a3018ffbae487ff21d1e51bc92526
HmacRIPEMD160-base64: Wt1vTR04ulq64/WbCZsoGMY1wM8=
HmacRIPEMD160-hex: 5add6f4d1d38ba5abae3f59b099b2818c635c0cf
HmacMD5-base64: sqHsDz4GBwmdfzl5HATppA==
HmacMD5-hex: b2a1ec0f3e0607099d7f39791c04e9a4
HmacSHA1-base64: o8Ak8BzMs7Y0V9hIsNL4nB90Sj0=
HmacSHA1-hex: a3c024f01cccb3b63457d848b0d2f89c1f744a3d
HmacSHA224-base64: +S2OBxrerlFMGsD5RQLnW7XCi4F+DgoB52I0jw==
HmacSHA224-hex: f92d8e071adeae514c1ac0f94502e75bb5c28b817e0e0a01e762348f
HmacSHA256-base64: PK/kD5K+asd9J5K0smfC2hHj8wh7k7sZxsUTN4aYS0Q=
HmacSHA256-hex: 3cafe40f92be6ac77d2792b4b267c2da11e3f3087b93bb19c6c5133786984b44
HmacSHA384-base64: b2i1J57kVppw+dC071Gk25fXfLkBmcILY38gkOOXgPjuyE8wK+LuP+w2pzTZ5tpb
HmacSHA384-hex: 6f68b5279ee4569a70f9d0b4ef51a4db97d77cb90199c20b637f2090e39780f8eec84f302be2ee3fec36a734d9e6da5b
HmacSHA512-base64: BjT9BDgLuvUGnIxGp0x9Id90FIiNmAwnoW1eJiy4yQWROcIS0JJgAPrwJuSDkEzvri9enZvV9R+8KsTE3lGBFQ==
HmacSHA512-hex: 0634fd04380bbaf5069c8c46a74c7d21df7414888d980c27a16d5e262cb8c9059139c212d0926000faf026e483904cefae2f5e9d9bd5f51fbc2ac4c4de518115
```


## RC4

`utils.RC4(data, key)` 接受字符串或 `[]byte`，返回 `[]byte`；对密文使用相同 key 再调用一次即可解密。key 长度必须为 1～256 字节，否则底层初始化错误会触发 panic。

```go
ciphertext := utils.RC4("hello", "example-key")
plaintext := utils.RC4(ciphertext, "example-key")
```

## 升级与兼容性说明

从旧版本升级到当前版本时，普通的 Go 请求调用方式保持兼容，通常不需要重写请求代码。以下行为在当前版本中有明确变化，升级前请检查相关代码。

### 查询参数编码

查询参数与旧版本保持一致：`Params.Encode()` 原样拼接键和值，`ParseParams(string)` 不自动解码。键或值中的空格、`&`、`=`、中文等特殊字符由调用方预先编码，例如使用项目的 `utils.EncodeURIComponent`：

```go
params := url.NewParams()
params.Set("q", utils.EncodeURIComponent("a b&c=中文"))

// q=a%20b%26c%3D%E4%B8%AD%E6%96%87
req := &url.Request{Params: params}
resp, err := requests.Get("https://example.com/search", req)
```

旧代码传入的 `a%20b` 等已编码值可以继续使用，不会把 `%` 再次编码为 `%25`。也可使用标准库 `net/url.QueryEscape`，其空格编码为 `+`。空值沿用旧版的裸键名形式，例如 `empty`；同名参数仍会归组。这一约定适用于完整 URL 的查询字符串以及字符串、map、`*url.Params` 形式的 Params，表单 `url.Values` 则仍会自动编码。

### Params、Values 和 Files

这些集合现在使用读写锁保护，`Values()`、`Files.Get()` 等读取方法返回独立副本。修改返回值不会再反向修改原集合，请使用 `Set`、`Add` 和 `Del` 修改数据。

集合包含锁和内部 map，不应按值复制。需要副本时使用对应的 `Clone()` 方法：

```go
paramsCopy := params.Clone()
dataCopy := data.Clone()
filesCopy := files.Clone()
```

文件上传和表单字段合并时，当前版本会在副本中合并字段，不会把表单字段永久追加到原始 `Files`。依赖这种累积行为的代码需要改为显式调用 `AddField`。

### 响应体保留变化

默认保留解压后的 `Content` 和共享其数据的内存 `Body`，不再默认保留 `RawResponse`。依赖原始报文的代码需要显式选择 `url.ResponseRetentionAll`。Body 读到 EOF 后不会自动回卷，重复解析应从 Content 新建 Reader。

`SimpleJson()` 优先使用 Content；流式响应没有 Content 时，它会缓存剩余 Body。完整说明见[响应体保留策略](#响应体保留策略)和[JSON 响应内容](#json-响应内容)。

### Session 和 Transport

Transport 缓存现在有数量上限和淘汰机制，并会在关闭时释放空闲连接。长期使用的 Session 建议显式关闭：

```go
session := requests.NewSession()
defer session.Close()

// 可按需调整 Transport 缓存上限
session.TransportCacheLimit = 32
```

请求开始时会复制 TLS、HTTP/2 和 ClientHello 配置快照，缓存不持有调用方的可变配置。复制期间仍需避免并发修改源配置，具体见[Session 会话](#session-会话)。

证书与 JA3 配置会在请求入口校验；TLS/HTTP2 的新代码建议使用 `ToTLSExtensionsChecked`、`ToHTTP2SettingsChecked` 显式处理转换错误。旧版单返回值转换函数仍保留兼容行为，不能据此认为所有非法输入都会返回错误。

另外，`req.Timeout == 0` 现在会正确继承 `Session.Timeout`，URL 中的用户认证也不会再重复拼接密码。

### C API 生命周期

导出函数签名保持不变，相同 Session ID 会复用 Cookie Jar。省略 `Verify`、`AllowRedirects` 时默认开启，显式 `false` 仍表示关闭。

普通请求和错误响应按 `id` 释放；流式元信息使用 `stream_id`，成功的 `stream_read` 使用 `stream_id + "_read"`，并非所有返回值都有 `id`。后台会回收长期未释放的资源，不应长期保存 C 指针。完整参数、示例和释放规则见 [C API](#c-api)。

### 升级检查清单

升级时重点检查以下代码：

1. 查询参数中的特殊字符是否已由调用方编码；旧版预编码用法可以继续使用。
2. 是否修改了 `Params.Values()` 或 `Files.Get()` 返回的对象并期望影响原集合。
3. 是否依赖默认的 `Response.RawResponse`。
4. 是否按值复制 `Params`、`Values` 或 `Files`。
5. C API 是否在使用后调用 `freeMemory`、`stream_close` 和 `freeSession`。

除此之外，常见的 `requests.Get`、`requests.Post`、`Session.Get` 和 `url.Request` 用法可以继续使用。

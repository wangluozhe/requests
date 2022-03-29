# requests
[![Gitee link address](https://img.shields.io/badge/gitee-reference-red?logo=gitee&logoColor=red&labelColor=white)](https://gitee.com/leegene/requests)[![Github link address](https://img.shields.io/badge/github-reference-blue?logo=github&logoColor=black&labelColor=white&color=black)](https://github.com/wangluozhe/requests)[![Go Version](https://img.shields.io/badge/Go%20Version-1.15.6-blue?logo=go&logoColor=white&labelColor=gray)]()[![Release Version](https://img.shields.io/badge/release-v1.0.04-blue)]()

requests支持以下新特性：

1. 支持http2，默认以http2进行连接，连接失败后会进行退化而进行http1.1连接
2. 支持JA3指纹修改
3. 支持http2+JA3指纹
4. 支持在使用代理的基础上修改JA3指纹

**此模块参考于Python的[requests模块](https://github.com/psf/requests/tree/main/requests)**



## 下载requests库

```bash
go get github.com/wangluozhe/requests
```



## 下载指定版本

```bash
go get github.com/wangluozhe/requests@v1.0.04
```



# 快速上手

迫不及待了吗？本页内容为如何入门 Requests 提供了很好的指引。其假设你已经安装了 Requests。如果还没有，去安装看看吧。

首先，确认一下：

- Requests已安装
- Requests是最新的

让我们从一些简单的示例开始吧。



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

现在，我们有一个名为 `r` 的 [`Response`](https://docs.python-requests.org/zh_CN/latest/api.html#requests.Response) 对象。我们可以从这个对象中获取所有我们想要的信息。

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
r := requests.Delete("http://httpbin.org/delete")
r := requests.Head("http://httpbin.org/get")
r := requests.Options("http://httpbin.org/get")
```

都很不错吧，但这也仅是 requests 的冰山一角呢。



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

```go
params := url.ParseParams("key1=value1&key2=value2")
r, err := requests.Get("http://httpbin.org/get",&url.Request{Params: params})
if err != nil {
	fmt.Println(err)
}
```

通过打印输出该 URL，你能看到 URL 已被正确编码：

```go
fmt.Println(r.Url)
http://httpbin.org/get?key1=value1&key2=value2
```

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
	}
	fmt.Println(r.Text)
}

[{"repository":{"open_issues":0, "url":"https://github.com/...
```

requests 会自动解码来自服务器的内容。大多数 unicode 字符集都能被无缝地解码。



## 二进制响应内容

你也能以字节数组的方式访问请求响应体，对于非文本请求：

```go
fmt.Println(r.Content)
[91 123 34 105 100 34 58 34 50 48 55 49 52 50 51 57 56 48 53 34 44 34 116 121 112 101 34 58 34...
```

Requests 会自动为你解码 `gzip` 和 `deflate` 以及`br`传输编码的响应数据。

例如，以请求返回的二进制数据创建一张图片，你可以使用如下代码：

```go
package main

import "github.com/wangluozhe/requests"

func main(){
    r, err := requests.Get("图片URL", nil)
	if err != nil {
		fmt.Println(err)
	}
	jpg,_ := os.Create("1.jpg")
	io.Copy(jpg,resp.Body)		// 第一种
	//jpg.Write(resp.Content)	// 第二种
}
```



## JSON 响应内容

Requests 中也有一个内置的 JSON 解码器，助你处理 JSON 数据：

```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
)

func main(){
    r, err := requests.Get("https://api.github.com/events", nil)
    if err != nil{
    	fmt.Println(err)
    }
    json, err := r.Json()
    fmt.Println(json, err)
}

{"Accept-Ranges":["bytes"],"Access-Control-Allow-Origin":["*"],"Access-Control...
```

如果 JSON 解码失败， `r.Json()` 就会返回一个异常。例如，响应内容是 401 (Unauthorized)，尝试访问 `r.Json()` 将会抛出 `map[] invalid character '(' after top-level value` 异常。

需要注意的是，成功调用 `r.Json()` 并**不**意味着响应的成功。有的服务器会在失败的响应中包含一个 JSON 对象（比如 HTTP 500 的错误细节）。这种 JSON 会被解码返回。要检查请求是否成功，请检查 `r.StatusCode` 是否和你的期望相同。



## 原始响应内容

在罕见的情况下，你可能想获取来自服务器的原始套接字响应，那么你可以访问 `r.Body`。 具体你可以这么做：

```go
r, err := requests.Get("http://www.baidu.com", nil)
if err != nil {
    fmt.Println(err)
}
fmt.Println(r.Body)

// 返回的是io.ReadCloser类型
```

但一般情况下，你应该以下面的模式将文本流保存到文件：

```go
f, _ := os.Create("baidu.txt")
io.Copy(f,resp.Body)
```



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

注意: 定制 header 的优先级低于某些特定的信息源，例如：

- `Content-Length`请求头不能随便设置，可能会有错误发生。
- 如果被重定向到别的主机，授权 header 就会被删除。
- 代理授权 header 会被 URL 中提供的代理身份覆盖掉。
- 在我们能判断内容长度的情况下，header 的 Content-Length 会被改写。

更进一步讲，Requests 不会基于定制 header 的具体情况改变自己的行为。只不过在最后的请求中，所有的 header 信息都会被传递进去。

注意: 所有的 header 值必须是 `string`。



## 有序请求头

如果你想让你的请求头变为有序的话，请添加一个`"Header-Order:"`参数即可，值为有序请求头数组，注：值必须全部为小写。

例如：

```go
package main

import (
	"github.com/wangluozhe/requests/url"
)

func main(){
    headers := url.NewHeaders()
    headers.Set("Path", "/get")
    headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
    headers.Set("accept-language", "zh-CN,zh;q=0.9")
    headers.Set("Scheme", "https")
    headers.Set("accept-encoding", "gzip, deflate, br")
    //headers.Set("Content-Length", "100")	// 注意，不能随便改变Content-Length大小
    headers.Set("Host", "httpbin.org")
    headers.Set("Accept", "application/json, text/javascript, */*; q=0.01")
    (*headers)["Header-Order:"] = []string{	// 请求头排序，值必须为小写
        "user-agent",
        "path",
        "accept-language",
        "scheme",
        "connection",
        "accept-encoding",
        "content-length",
        "host",
        "accept",
    }
    req.Headers = headers
    r, err := requests.Get("https://httpbin.org/get", req)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println("text:", r.Text)

    // 最好用fiddler抓包工具查看一下
}
```

或者：

```go
package main

import (
	"github.com/wangluozhe/requests/url"
)

func main(){
    req := url.NewRequest()
    req.Headers = url.ParseHeaders(`
    :authority: spidertools.cn
    :method: GET
    :path: /
    :scheme: https
    accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9
    accept-encoding: gzip, deflate, br
    accept-language: zh-CN,zh;q=0.9
    cache-control: no-cache
    cookie: _ga=GA1.1.630251354.1645893020; Hm_lvt_def79de877408c7bd826e49b694147bc=1647245863,1647936048,1648296630; Hm_lpvt_def79de877408c7bd826e49b694147bc=1648296630
    pragma: no-cache
    sec-ch-ua: " Not A;Brand";v="99", "Chromium";v="98", "Google Chrome";v="98"
    sec-ch-ua-mobile: ?0
    sec-ch-ua-platform: "Windows"
    sec-fetch-dest: document
    sec-fetch-mode: navigate
    sec-fetch-site: same-origin
    sec-fetch-user: ?1
    upgrade-insecure-requests: 1
    user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36
    `) // 注意：这是反引号，不是单引号
    r, err := requests.Get("https://httpbin.org/get", req)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println("text:", r.Text)

    // 最好用fiddler抓包工具查看一下
}
```



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

fmt.Println(resp.Text)

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
req.Data = url.ParseData("key1=value1&key2=value2")
r, err := requests.Post("http://www.baidu.com",req)
if err != nil {
    fmt.Println(err)
}

fmt.Println(resp.Text)

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
fmt.Println(resp.Text)

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

你可以使用 `Json` 参数直接传递，然后它就会被自动编码。

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
	}
	fmt.Println(r.Text)
}

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
fmt.Println(r.Text)

...
"files": {
    "api": "文件内容"
}
...
```

requests使得FormData的使用也方便多了：

```go
files := url.NewFiles()
// SetFile(name,value)
// name为字段名，value为值
files.SetField("name","value")
req := url.NewRequest()
req.Files = files
r, err := requests.Post("http://httpbin.org/post",req)
if err != nil {
    fmt.Println(err)
}
fmt.Println(r.Text)

...
"form": {
	"name": "value"
}
...
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

如果发送了一个错误请求(一个 4XX 客户端错误，或者 5XX 服务器错误响应)，我们可以通过 `Response.RaiseForStatus()` 来抛出异常：

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

一切都挺和谐哈。



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

fmt.Println(r.Text)

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
	req.Cookies = url.ParseCookies(rawUrl,"_ga=GA1.1.630251354.1645893020; Hm_lvt_def79de877408c7bd826e49b694147bc=1647245863,1647936048,1648296630; Hm_lpvt_def79de877408c7bd826e49b694147bc=1648301329")
	r, err := requests.Get(rawUrl, req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.Text)
}

{
  "cookies": {
    "Hm_lpvt_def79de877408c7bd826e49b694147bc": "1648301329", 
    "Hm_lvt_def79de877408c7bd826e49b694147bc": "1647245863,1647936048,1648296630", 
    "_ga": "GA1.1.630251354.1645893020"
  }
}
```



## 重定向与请求历史

默认情况下，除了 HEAD, Requests 会自动处理所有重定向。

可以使用响应对象的 `history` 方法来追踪重定向。

`Response.History` 是一个 `Response` 对象的列表，为了完成请求而创建了这些对象。这个对象列表按照从最老到最近的请求进行排序。

例如，Github 将所有的 HTTP 请求重定向到 HTTPS：

```go
r, err := requests.Get("http://github.com", nil)

fmt.Println(r.Url)

https://github.com/

fmt.Println(r.StatusCode)

200

fmt.Println(r.History)

[0xc0001803f0]
```

如果你使用的是GET、OPTIONS、POST、PUT、PATCH 或者 DELETE，那么你可以通过 `allow_redirects` 参数禁用重定向处理：

```go
req := url.NewRequest()
req.AllowRedirects = false
r, err := requests.Get("http://github.com", req)
if err != nil {
    fmt.Println(err)
}
fmt.Println(r.StatusCode)

301

fmt.Println(r.History)

[]
```

如果你使用了 HEAD，你也可以启用重定向：

```go
req := url.NewRequest()
req.AllowRedirects = true
r, err := requests.Get("http://github.com", req)
if err != nil {
	fmt.Println(err)
}
fmt.Println(r.Url)

https://github.com/

fmt.Println(r.History)

[0xc0001803f0]
```



## 超时

你可以告诉 requests 在经过以 `Timeout` 参数设定的秒数时间之后停止等待响应。基本上所有的生产代码都应该使用这一参数。如果不使用，你的程序可能会永远失去响应：

```go
req := url.NewRequest()
req.Timeout = 1 * time.Millisecond
r, err := requests.Get("http://github.com",req)
if err != nil {
    fmt.Println(err)
}

panic: runtime error: invalid memory address or nil pointer dereference
[signal 0xc0000005 code=0x0 addr=0x0 pc=0x253460]

goroutine 1 [running]:
main.main()
	D:/Go/github.com/wangluozhe/requests/examples/test.go:27 +0xc0
```

注意

`Timeout` 仅对连接过程有效，与响应体的下载无关。 `Timeout` 并不是整个下载响应的时间限制，而是如果服务器在 `Timeout` 秒内没有应答，将会引发一个异常（更精确地说，是在 `Timeout` 秒内没有从基础套接字上接收到任何字节的数据时）If no timeout is specified explicitly, requests do not time out.



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
	req.Auth = []string{"user","password"}
	r, err := requests.Get("http://httpbin.org/basic-auth/user/password",req)
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println("text:",r.Text)
}

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
fmt.Println(r.Text)

{"ja3_hash":"b32309a26951912be7dba376398abc3b", "ja3": "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-21,29-23-24,0", "User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36"}
```



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

https%3A%2F%2Fwww.baidu.com%3Fpage%3D10%26abc%3D123%26name%3D%E4%BD%A0%E5%A5%BD%E5%95%8A
https://www.baidu.com?page=10&abc=123&name=你好啊
https://www.baidu.com?page=10&abc=123&name=%E4%BD%A0%E5%A5%BD%E5%95%8A
https://www.baidu.com?page=10&abc=123&name=你好啊
https%3A//www.baidu.com%3Fpage%3D10%26abc%3D123%26name%3D%u4f60%u597d%u554a
https://www.baidu.com?page=10&abc=123&name=你好啊
```



## Base64编码

```go
package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	url := "https://www.baidu.com"
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

aHR0cHM6Ly93d3cuYmFpZHUuY29t
https://www.baidu.com
aHR0cHM6Ly93d3cuYmFpZHUuY29t
https://www.baidu.com
```


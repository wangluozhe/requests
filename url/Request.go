package url

import (
	"bytes"
	"encoding/json"
	"hash/fnv"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wangluozhe/chttp"
	"github.com/wangluozhe/chttp/cookiejar"
)

func NewRequest() *Request {
	return &Request{
		AllowRedirects: true,
		Verify:         true,
	}
}

type Request struct {
	// Params 设置 URL 查询参数.
	// 支持的类型 (Accepted types):
	//   - *url.Params
	//   - string (e.g. "key=value&a=1")
	//   - map[string]string
	//   - map[string][]string
	//   - map[string]int, map[string][]int
	//   - map[string]float64, map[string][]float64
	//   - map[string]interface{} (支持递归解析)
	Params any

	// Headers 设置请求头.
	// 支持的类型 (Accepted types):
	//   - *http.Header
	//   - string (e.g. "User-Agent: abc\nAccept: */*")
	//   - map[string]string
	//   - map[string][]string
	//   - map[string]interface{} (值支持 string, int, float64, bool)
	//   - map[string][]interface{}
	Headers any

	// Cookies 设置请求 Cookies.
	// 支持的类型 (Accepted types):
	//   - *cookiejar.Jar
	//   - string (e.g. "name=value; a=1")
	//   - map[string]string
	//   - map[string]int
	//   - map[string]float64
	//   - map[string]interface{} (值支持 string, int, float64, bool)
	Cookies any

	// Data 设置表单数据 (application/x-www-form-urlencoded).
	// 支持的类型 (Accepted types):
	//   - *url.Values
	//   - string (e.g. "key=value&a=1")
	//   - map[string]string
	//   - map[string][]string
	//   - map[string]int, map[string][]int
	//   - map[string]float64, map[string][]float64
	//   - map[string]interface{} (支持递归解析)
	Data any

	// Files 设置上传的文件 (multipart/form-data).
	// 支持的类型 (Accepted types):
	//   - *url.Files
	//   - map[string]string (key为字段名, value为文件路径. 会自动提取文件名，ContentType默认为空)
	Files any

	// Json 设置 JSON 请求体 (application/json).
	// 支持的类型 (Accepted types):
	//   - map[string]interface{}
	//   - struct, array, slice, int, bool, string... (任何可被 json.Marshal 处理的类型)
	Json any

	// Body 设置原始请求体.
	// 支持的类型 (Accepted types):
	//   - io.Reader
	//   - []byte
	//   - string
	Body           any
	Auth           []string
	Timeout        time.Duration
	AllowRedirects bool
	Proxies        string
	Verify         bool
	Cert           []string
	Stream         bool
	Ja3            string
	RandomJA3      bool
	ForceHTTP1     bool
	TLSExtensions  *http.TLSExtensions
	HTTP2Settings  *http.HTTP2Settings
}

func (req *Request) Hash() string {
	bytes, err := json.Marshal(req)
	if err != nil {
		return ""
	}

	h := fnv.New64a()
	h.Write(bytes)
	return strconv.Itoa(int(h.Sum64()))
}

// GetParams 获取 Params 结构体
func (req *Request) GetParams() *Params {
	if req.Params == nil {
		return nil
	}
	if p, ok := req.Params.(*Params); ok {
		return p
	}
	return ParseParams(req.Params)
}

// GetHeaders 获取 Headers 结构体
func (req *Request) GetHeaders() *http.Header {
	if req.Headers == nil {
		return nil
	}
	if h, ok := req.Headers.(*http.Header); ok {
		return h
	}
	return ParseHeaders(req.Headers)
}

// GetCookies 获取 Cookies 结构体
// 注意：解析字符串 Cookies 需要 rawurl 来确定域名，请在 Session 逻辑中传入
func (req *Request) GetCookies(rawurl string) *cookiejar.Jar {
	if req.Cookies == nil {
		return nil
	}
	if c, ok := req.Cookies.(*cookiejar.Jar); ok {
		return c
	}
	return ParseCookies(rawurl, req.Cookies)
}

// GetData 获取 Data (Values) 结构体
func (req *Request) GetData() *Values {
	if req.Data == nil {
		return nil
	}
	if v, ok := req.Data.(*Values); ok {
		return v
	}
	return ParseData(req.Data)
}

// GetFiles 获取 Files 结构体
func (req *Request) GetFiles() *Files {
	if req.Files == nil {
		return nil
	}
	if f, ok := req.Files.(*Files); ok {
		return f
	}
	// 为了方便，支持简单的 map[string]string { "field": "path/to/file" }
	if m, ok := req.Files.(map[string]string); ok {
		f := NewFiles()
		for name, path := range m {
			// 自动提取文件名，ContentType 留空由 SetFile 内部处理
			f.SetFile(name, filepath.Base(path), path, "")
		}
		return f
	}
	return nil
}

// GetBody 获取 Body io.Reader
func (req *Request) GetBody() io.Reader {
	if req.Body == nil {
		return nil
	}
	switch v := req.Body.(type) {
	case io.Reader:
		return v
	case string:
		return strings.NewReader(v)
	case []byte:
		return bytes.NewReader(v)
	default:
		return nil
	}
}

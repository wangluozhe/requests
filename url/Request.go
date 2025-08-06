package url

import (
	"io"
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
	Params         *Params
	Headers        *http.Header
	Cookies        *cookiejar.Jar
	Data           *Values
	Files          *Files
	Json           map[string]interface{}
	Body           io.Reader
	Auth           []string
	Timeout        time.Duration
	AllowRedirects bool
	Proxies        string
	Verify         bool
	Cert           []string
	Stream         bool
	Ja3            string
	ForceHTTP1     bool
	TLSExtensions  *http.TLSExtensions
	HTTP2Settings  *http.HTTP2Settings
}

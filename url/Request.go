package url

import (
	"github.com/Danny-Dasilva/fhttp"
	"github.com/Danny-Dasilva/fhttp/cookiejar"
	"time"
)

func NewRequest() *Request {
	return &Request{
		AllowRedirects: true,
		Verify:         true,
	}
}

type Proxies struct {
	Scheme   string
	Host     string
	Port     string
	User     string
	Password string
}

const (
	PROXIES_SCHEME_HTTP   = "http"
	PROXIES_SCHEME_HTTPS  = "https"
	PROXIES_SCHEME_SOCKS5 = "socks5"
)

type Request struct {
	Params         *Params
	Headers        *http.Header
	Cookies        *cookiejar.Jar
	Data           *Values
	Files          *Files
	Json           map[string]interface{}
	Auth           []string
	Timeout        time.Duration
	AllowRedirects bool
	Proxies        Proxies
	Verify         bool
	Cert           []string
	Ja3            string
}

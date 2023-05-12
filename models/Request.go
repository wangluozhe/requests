package models

import (
	"github.com/wangluozhe/fhttp"
	"github.com/wangluozhe/fhttp/cookiejar"
	"github.com/wangluozhe/requests/url"
)

type Request struct {
	Method  string
	Url     string
	Params  *url.Params
	Headers *http.Header
	Cookies *cookiejar.Jar
	Data    *url.Values
	Files   *url.Files
	Json    map[string]interface{}
	Auth    []string
}

func (req *Request) Prepare() *PrepareRequest {
	p := &PrepareRequest{}
	p.Prepare(
		req.Method,
		req.Url,
		req.Params,
		req.Headers,
		req.Cookies,
		req.Data,
		req.Files,
		req.Json,
		req.Auth,
	)
	return p
}

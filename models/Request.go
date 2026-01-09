package models

import (
	"io"

	"github.com/wangluozhe/chttp"
	"github.com/wangluozhe/chttp/cookiejar"
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
	Body    io.Reader
	Json    any
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
		req.Body,
		req.Auth,
	)
	return p
}

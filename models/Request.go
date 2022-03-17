package models

import (
	"github.com/Danny-Dasilva/fhttp"
	"github.com/Danny-Dasilva/fhttp/cookiejar"
	"github.com/wangluozhe/requests/url"
)

type Request struct {
	Method string
	Url string
	Params *url.Params
	Headers *http.Header
	Cookies *cookiejar.Jar
	Data *url.Values
	Files *url.Files
	Json map[string]string
	Auth []string
}

func (this *Request) Prepare() *PrepareRequest {
	p := &PrepareRequest{}
	p.Prepare(
		this.Method,
		this.Url,
		this.Params,
		this.Headers,
		this.Cookies,
		this.Data,
		this.Files,
		this.Json,
		this.Auth,
	)
	return p
}
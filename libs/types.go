package libs

import (
	"encoding/json"
	"strings"
)

type RequestParams struct {
	Id                 string                 `json:"Id"`
	Method             string                 `json:"Method"`
	Url                string                 `json:"Url"`
	Params             map[string]string      `json:"Params"`
	Headers            map[string]string      `json:"Headers"`
	HeadersOrder       []string               `json:"HeadersOrder"`
	UnChangedHeaderKey []string               `json:"UnChangedHeaderKey"`
	Cookies            map[string]string      `json:"Cookies"`
	Data               map[string]string      `json:"Data"`
	Json               map[string]interface{} `json:"Json"`
	Body               string                 `json:"Body"`
	Auth               []string               `json:"Auth"`
	Timeout            int                    `json:"Timeout"`
	AllowRedirects     bool                   `json:"AllowRedirects"`
	Proxies            string                 `json:"Proxies"`
	Verify             bool                   `json:"Verify"`
	Cert               []string               `json:"Cert"`
	Ja3                string                 `json:"Ja3"`
	RandomJA3          bool                   `json:"RandomJA3"`
	ForceHTTP1         bool                   `json:"ForceHTTP1"`
	PseudoHeaderOrder  []string               `json:"PseudoHeaderOrder"`
	TLSExtensions      string                 `json:"TLSExtensions"`
	HTTP2Settings      string                 `json:"HTTP2Settings"`
	Stream             bool                   `json:"Stream"`

	allowRedirectsSet bool
	verifySet         bool
	randomJA3Set      bool
	forceHTTP1Set     bool
}

// UnmarshalJSON 记录可选布尔字段是否被省略，使 C 接口能够继承 Session 默认值并接受显式 false。
func (p *RequestParams) UnmarshalJSON(data []byte) error {
	type requestParams RequestParams
	var decoded requestParams
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = RequestParams(decoded)

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for key := range fields {
		switch {
		case strings.EqualFold(key, "AllowRedirects"):
			p.allowRedirectsSet = true
		case strings.EqualFold(key, "Verify"):
			p.verifySet = true
		case strings.EqualFold(key, "RandomJA3"):
			p.randomJA3Set = true
		case strings.EqualFold(key, "ForceHTTP1"):
			p.forceHTTP1Set = true
		}
	}
	return nil
}

func (p RequestParams) AllowRedirectsSet() bool { return p.allowRedirectsSet }
func (p RequestParams) VerifySet() bool         { return p.verifySet }
func (p RequestParams) RandomJA3Set() bool      { return p.randomJA3Set }
func (p RequestParams) ForceHTTP1Set() bool     { return p.forceHTTP1Set }

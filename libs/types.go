package libs

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
	ForceHTTP1         bool                   `json:"ForceHTTP1"`
	PseudoHeaderOrder  []string               `json:"PseudoHeaderOrder"`
	TLSExtensions      string                 `json:"TLSExtensions"`
	HTTP2Settings      string                 `json:"HTTP2Settings"`
}

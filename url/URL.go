package url

import (
	"net/url"
)

// 解析URL
func Parse(rawurl string) (*URL, error) {
	p, err := url.Parse(rawurl)
	return &URL{
		Scheme:      p.Scheme,
		User:        p.User,
		Host:        p.Host,
		Path:        p.Path,
		RawParams:   p.RawQuery,
		Params:      ParseParams(p.RawQuery),
		RawFragment: p.EscapedFragment(),
		Fragment:    p.Fragment,
	}, err
}

// URL结构体
type URL struct {
	Scheme      string        // 协议
	User        *url.Userinfo // 用户信息
	Host        string        // 地址
	Path        string        // 路径
	RawParams   string        // GET参数
	Params      *Params       // GET参数
	RawFragment string        // 原始锚点
	Fragment    string        // 锚点
}

// URL结构体转字符串
func (this *URL) String() string {
	s := this.Scheme + "://"
	if this.User != nil {
		s += this.User.String() + "@"
	}
	s += this.Host
	if this.Path == "" {
		s += "/"
	}
	s += this.Path
	this.RawParams = this.Params.Encode()
	if this.RawParams != "" {
		s += "?" + this.RawParams
	}
	return s
}

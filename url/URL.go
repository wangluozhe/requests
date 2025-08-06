package url

import (
	"net/url"
	"sync"
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
		mutex:       &sync.RWMutex{},
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
	mutex       *sync.RWMutex
}

// URL结构体转字符串
func (u *URL) String() string {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	s := u.Scheme + "://"
	if u.User != nil {
		s += u.User.String() + "@"
	}
	s += u.Host
	if u.Path == "" {
		s += "/"
	}
	s += u.Path
	u.RawParams = u.Params.Encode()
	if u.RawParams != "" {
		s += "?" + u.RawParams
	}
	if u.RawFragment != "" {
		s += "#" + u.RawFragment
	}
	return s
}

package url

import (
	"errors"
	http "github.com/Danny-Dasilva/fhttp"
	"strings"
)

// 初始化Headers结构体
func NewHeaders() *http.Header {
	return &http.Header{
		http.PHeaderOrderKey: {":method", ":authority", ":scheme", ":path"},
	}
}

// 解析Headers字符串为结构体
func ParseHeaders(headers string) *http.Header {
	h := http.Header{}
	headerOrder := []string{}
	lines := strings.Split(headers, "\n")
	for _, header := range lines {
		header = strings.TrimSpace(header)
		if header == "" || strings.Index(header, ":") == 0 || strings.Index(header, "/") == 0 || strings.Index(header, "#") == 0 {
			continue
		}
		keyValue := strings.SplitN(header, ":", 2)
		if len(keyValue) != 2 {
			panic(errors.New("该字符串不符合http头部标准！"))
		}
		key := keyValue[0]
		value := keyValue[1]
		h.Set(key, value)
		headerOrder = append(headerOrder, strings.ToLower(key))
	}
	h[http.HeaderOrderKey] = headerOrder
	return &h
}

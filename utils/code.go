package utils

import (
	"net/url"
)

// URI编码
func EncodeURIComponent(s string) string {
	return url.QueryEscape(s)
}

// URI解码
func DecodeURIComponent(s string) string {
	t, _ := url.QueryUnescape(s)
	return t
}


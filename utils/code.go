package utils

import (
	"encoding/base64"
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

// Base64编码
func Btoa(s interface{}) string {
	byte_s := stringAndByte(s)
	return base64.StdEncoding.EncodeToString(byte_s)
}

// Base64编码，同上
func Base64Encode(s interface{}) string {
	byte_s := stringAndByte(s)
	return Btoa(byte_s)
}

// Base64解码
func Atob(s interface{}) string {
	byte_s := stringAndByte(s)
	str, err := base64.StdEncoding.DecodeString(string(byte_s))
	if err != nil{
		panic(err)
	}
	return string(str)
}

// Base64解码，同上
func Base64Decode(s interface{}) string {
	byte_s := stringAndByte(s)
	return Atob(byte_s)
}

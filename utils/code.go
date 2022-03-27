package utils

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/url"
)

// 只接受string和[]byte类型
func stringAndByte(s interface{}) []byte {
	var byte_s []byte
	var err error
	switch s.(type) {
	case string:
		byte_s = []byte(s.(string))
	case []byte:
		byte_s = s.([]byte)
	default:
		err = errors.New("Please check whether the type is string and []byte.")
		panic(err)
	}
	return byte_s
}

// Hex编码
func HexEncode(s interface{}) []byte {
	byte_s := stringAndByte(s)
	dst := make([]byte, hex.EncodedLen(len(byte_s)))
	n := hex.Encode(dst, byte_s)
	return dst[:n]
}

// Hex解码
func HexDecode(s interface{}) []byte {
	byte_s := stringAndByte(s)
	dst := make([]byte, hex.DecodedLen(len(byte_s)))
	n, err := hex.Decode(dst, byte_s)
	if err != nil {
		panic(err)
	}
	return dst[:n]
}

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

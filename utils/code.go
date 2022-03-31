package utils

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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
func EncodeURIComponent(s interface{}) string {
	byte_s := stringAndByte(s)
	es := url.QueryEscape(string(byte_s))
	es = strings.ReplaceAll(es, "+", "%20")
	es = strings.ReplaceAll(es, "%21", "!")
	es = strings.ReplaceAll(es, "%27", "'")
	es = strings.ReplaceAll(es, "%28", "(")
	es = strings.ReplaceAll(es, "%29", ")")
	es = strings.ReplaceAll(es, "%2A", "*")
	return es
}

// URI解码
func DecodeURIComponent(s interface{}) string {
	byte_s := stringAndByte(s)
	t, err := url.QueryUnescape(strings.ReplaceAll(strings.ReplaceAll(string(byte_s), "+", "%2B"), "%20", "+"))
	if err != nil {
		panic(err)
	}
	return t
}

// URI编码
func EncodeURI(s interface{}) string {
	byte_s := stringAndByte(s)
	es := EncodeURIComponent(string(byte_s))
	ss := "!#$&'()*+,-./:=?@_~"
	for i := 0; i < len(ss); i++ {
		es = strings.ReplaceAll(es, "%"+strings.ToUpper(string(HexEncode(string(ss[i])))), string(ss[i]))
	}
	return strings.ReplaceAll(es, "%3B", ";")
}

// URI解码
func DecodeURI(s interface{}) string {
	byte_s := stringAndByte(s)
	es := string(byte_s)
	ss := "!#$&'()*+,-./:=?@_~"
	for i := 0; i < len(ss); i++ {
		es = strings.ReplaceAll(es, "%"+strings.ToUpper(string(HexEncode(string(ss[i])))), "$"+"%"+strings.ToUpper(string(HexEncode(string(ss[i]))))+"$")
	}
	es = DecodeURIComponent(es)
	for i := 0; i < len(ss); i++ {
		es = strings.ReplaceAll(es, "$"+string(ss[i])+"$", "%"+strings.ToUpper(string(HexEncode(string(ss[i])))))
	}
	return es
}

// Base32编码
func Base32Encode(s interface{}) string {
	byte_s := stringAndByte(s)
	return base32.StdEncoding.EncodeToString(byte_s)
}

// Base32解码
func Base32Decode(s interface{}) string {
	byte_s := stringAndByte(s)
	str, err := base32.StdEncoding.DecodeString(string(byte_s))
	if err != nil {
		panic(err)
	}
	return string(str)
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
	if err != nil {
		panic(err)
	}
	return string(str)
}

// Base64解码，同上
func Base64Decode(s interface{}) string {
	byte_s := stringAndByte(s)
	return Atob(byte_s)
}

// 中文转Unicode
func Escape(s interface{}) string {
	byte_s := stringAndByte(s)
	str := string(byte_s)
	es := ""
	for _, s := range str {
		switch {
		case s >= '0' && s <= '9':
			es += string(s)
		case s >= 'a' && s <= 'z':
			es += string(s)
		case s >= 'A' && s <= 'Z':
			es += string(s)
		case strings.Contains("*+-./@_", string(s)):
			es += string(s)
		case int(s) <= 127:
			es += "%" + strings.ToUpper(string(HexEncode(string(s))))
		case int(s) >= 128:
			es += strings.ReplaceAll(strings.ReplaceAll(strconv.QuoteToASCII(string(s)), "\"", ""), "\\u", "%u")
		}
	}
	return es
}

// Unicode转中文
func UnEscape(s interface{}) string {
	byte_s := stringAndByte(s)
	str := string(byte_s)
	re, _ := regexp.Compile("(%u)[0-9a-zA-Z]{4}")
	str = re.ReplaceAllStringFunc(str, func(st string) string {
		bs, _ := hex.DecodeString(strings.ReplaceAll(st, "%u", ""))
		r := uint16(0)
		binary.Read(bytes.NewReader(bs), binary.BigEndian, &r)
		return string(r)
	})
	str = DecodeURIComponent(str)
	return str
}

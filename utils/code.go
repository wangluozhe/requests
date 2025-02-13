package utils

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// 只接受string和[]byte类型
func stringAndByte(s interface{}) []byte {
	switch v := s.(type) {
	case string:
		return []byte(v)
	case []byte:
		return v
	default:
		panic(errors.New("Please check whether the type is string or []byte."))
	}
}

// Hex编码
func HexEncode(s interface{}) []byte {
	byte_s := stringAndByte(s)
	dst := make([]byte, hex.EncodedLen(len(byte_s)))
	hex.Encode(dst, byte_s)
	return dst
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
	for _, char := range ss {
		es = strings.ReplaceAll(es, "%"+strings.ToUpper(string(HexEncode(string(char)))), string(char))
	}
	return strings.ReplaceAll(es, "%3B", ";")
}

// URI解码
func DecodeURI(s interface{}) string {
	byte_s := stringAndByte(s)
	es := string(byte_s)
	ss := "!#$&'()*+,-./:=?@_~"
	for _, char := range ss {
		es = strings.ReplaceAll(es, "%"+strings.ToUpper(string(HexEncode(string(char)))), "$"+"%"+strings.ToUpper(string(HexEncode(string(char))))+"$")
	}
	es = DecodeURIComponent(es)
	for _, char := range ss {
		es = strings.ReplaceAll(es, "$"+string(char)+"$", "%"+strings.ToUpper(string(HexEncode(string(char)))))
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
	return Btoa(s)
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
	return Atob(s)
}

// 中文转Unicode
func Escape(s interface{}) string {
	byte_s := stringAndByte(s)
	str := string(byte_s)
	var es strings.Builder
	for _, r := range str {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', strings.ContainsRune("*+-./@_", r):
			es.WriteRune(r)
		case r <= 127:
			es.WriteString("%" + strings.ToUpper(string(HexEncode(string(r)))))
		default:
			es.WriteString(strings.ReplaceAll(strings.ReplaceAll(strconv.QuoteToASCII(string(r)), "\"", ""), "\\u", "%u"))
		}
	}
	return es.String()
}

// Unicode转中文
func UnEscape(s interface{}) string {
	byte_s := stringAndByte(s)
	str := string(byte_s)
	re := regexp.MustCompile(`%u[0-9a-fA-F]{4}`)
	str = re.ReplaceAllStringFunc(str, func(st string) string {
		bs, _ := hex.DecodeString(strings.TrimPrefix(st, "%u"))
		r := rune(0)
		binary.Read(bytes.NewReader(bs), binary.BigEndian, &r)
		return string(r)
	})
	return DecodeURIComponent(str)
}

// Marshal 避免json.Marshal对 "<", ">", "&" 等字符进行HTML编码
func Marshal(data interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	// 禁用HTML转义
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(data)
	return buffer.Bytes(), err
}

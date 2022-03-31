package utils

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
)

// SHA1加密
func SHA1(s interface{}) []byte {
	byte_s := stringAndByte(s)
	h := sha1.New()
	h.Write(byte_s)
	return h.Sum(nil)
}

// SHA224加密
func SHA224(s interface{}) []byte {
	byte_s := stringAndByte(s)
	h := sha256.New224()
	h.Write(byte_s)
	return h.Sum(nil)
}

// SHA256加密
func SHA256(s interface{}) []byte {
	byte_s := stringAndByte(s)
	h := sha256.New()
	h.Write(byte_s)
	return h.Sum(nil)
}

// SHA384加密
func SHA384(s interface{}) []byte {
	byte_s := stringAndByte(s)
	h := sha512.New384()
	h.Write(byte_s)
	return h.Sum(nil)
}

// SHA512加密
func SHA512(s interface{}) []byte {
	byte_s := stringAndByte(s)
	h := sha512.New()
	h.Write(byte_s)
	return h.Sum(nil)
}

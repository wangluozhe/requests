package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"
)

// HmacMD4加密
func HmacMD4(s, key interface{}) []byte {
	byte_s := stringAndByte(s)
	byte_key := stringAndByte(key)
	h := hmac.New(md4.New, byte_key)
	h.Write(byte_s)
	return h.Sum(nil)
}

// HmacRIPEMD160加密
func HmacRIPEMD160(s, key interface{}) []byte {
	byte_s := stringAndByte(s)
	byte_key := stringAndByte(key)
	h := hmac.New(ripemd160.New, byte_key)
	h.Write(byte_s)
	return h.Sum(nil)
}

// HmacMD5加密
func HmacMD5(s, key interface{}) []byte {
	byte_s := stringAndByte(s)
	byte_key := stringAndByte(key)
	h := hmac.New(md5.New, byte_key)
	h.Write(byte_s)
	return h.Sum(nil)
}

// HmacSHA1加密
func HmacSHA1(s, key interface{}) []byte {
	byte_s := stringAndByte(s)
	byte_key := stringAndByte(key)
	h := hmac.New(sha1.New, byte_key)
	h.Write(byte_s)
	return h.Sum(nil)
}

// HmacSHA224加密
func HmacSHA224(s, key interface{}) []byte {
	byte_s := stringAndByte(s)
	byte_key := stringAndByte(key)
	h := hmac.New(sha256.New224, byte_key)
	h.Write(byte_s)
	return h.Sum(nil)
}

// HmacSHA256加密
func HmacSHA256(s, key interface{}) []byte {
	byte_s := stringAndByte(s)
	byte_key := stringAndByte(key)
	h := hmac.New(sha256.New, byte_key)
	h.Write(byte_s)
	return h.Sum(nil)
}

// HmacSHA384加密
func HmacSHA384(s, key interface{}) []byte {
	byte_s := stringAndByte(s)
	byte_key := stringAndByte(key)
	h := hmac.New(sha512.New384, byte_key)
	h.Write(byte_s)
	return h.Sum(nil)
}

// HmacSHA512加密
func HmacSHA512(s, key interface{}) []byte {
	byte_s := stringAndByte(s)
	byte_key := stringAndByte(key)
	h := hmac.New(sha512.New, byte_key)
	h.Write(byte_s)
	return h.Sum(nil)
}

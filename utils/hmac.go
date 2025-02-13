package utils

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"
	"hash"
)

// hmacHash 是一个辅助函数，用于生成 HMAC 哈希值
func hmacHash(hashFunc func() hash.Hash, s, key interface{}) []byte {
	byte_s := stringAndByte(s)
	byte_key := stringAndByte(key)
	h := hmac.New(hashFunc, byte_key)
	h.Write(byte_s)
	return h.Sum(nil)
}

// HmacMD4加密
func HmacMD4(s, key interface{}) []byte {
	return hmacHash(md4.New, s, key)
}

// HmacRIPEMD160加密
func HmacRIPEMD160(s, key interface{}) []byte {
	return hmacHash(ripemd160.New, s, key)
}

// HmacMD5加密
func HmacMD5(s, key interface{}) []byte {
	return hmacHash(md5.New, s, key)
}

// HmacSHA1加密
func HmacSHA1(s, key interface{}) []byte {
	return hmacHash(sha1.New, s, key)
}

// HmacSHA224加密
func HmacSHA224(s, key interface{}) []byte {
	return hmacHash(sha256.New224, s, key)
}

// HmacSHA256加密
func HmacSHA256(s, key interface{}) []byte {
	return hmacHash(sha256.New, s, key)
}

// HmacSHA384加密
func HmacSHA384(s, key interface{}) []byte {
	return hmacHash(sha512.New384, s, key)
}

// HmacSHA512加密
func HmacSHA512(s, key interface{}) []byte {
	return hmacHash(sha512.New, s, key)
}

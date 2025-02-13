package utils

import (
	"crypto/md5"
	"fmt"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"
	"hash"
)

// hashString 是一个辅助函数，用于生成哈希值
func hashString(hashFunc func() hash.Hash, s interface{}) string {
	byte_s := stringAndByte(s)
	h := hashFunc()
	h.Write(byte_s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// MD4加密
func MD4(s interface{}) string {
	return hashString(md4.New, s)
}

// RIPEMD160加密
func RIPEMD160(s interface{}) string {
	return hashString(ripemd160.New, s)
}

// MD5加密
func MD5(s interface{}) string {
	return hashString(md5.New, s)
}

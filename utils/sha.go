package utils

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

// hashBytes 是一个辅助函数，用于生成哈希值
func hashBytes(hashFunc func() hash.Hash, s interface{}) []byte {
	byte_s := stringAndByte(s)
	h := hashFunc()
	h.Write(byte_s)
	return h.Sum(nil)
}

// SHA1加密
func SHA1(s interface{}) []byte {
	return hashBytes(sha1.New, s)
}

// SHA224加密
func SHA224(s interface{}) []byte {
	return hashBytes(sha256.New224, s)
}

// SHA256加密
func SHA256(s interface{}) []byte {
	return hashBytes(sha256.New, s)
}

// SHA384加密
func SHA384(s interface{}) []byte {
	return hashBytes(sha512.New384, s)
}

// SHA512加密
func SHA512(s interface{}) []byte {
	return hashBytes(sha512.New, s)
}

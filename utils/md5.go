package utils

import (
	"crypto/md5"
	"fmt"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"
)

// MD4加密
func MD4(s interface{}) string {
	byte_s := stringAndByte(s)
	m := md4.New()
	m.Write(byte_s)
	return fmt.Sprintf("%x", m.Sum(nil))
}

// RIPEMD160加密
func RIPEMD160(s interface{}) string {
	byte_s := stringAndByte(s)
	r := ripemd160.New()
	r.Write(byte_s)
	return fmt.Sprintf("%x", r.Sum(nil))
}

// MD5加密
func MD5(s interface{}) string {
	byte_s := stringAndByte(s)
	m := md5.New()
	m.Write(byte_s)
	return fmt.Sprintf("%x", m.Sum(nil))
}

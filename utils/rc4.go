package utils

import (
	"crypto/rc4"
)

// RC4加密
func RC4(data, key interface{}) []byte {
	byte_data := stringAndByte(data)
	byte_key := stringAndByte(key)
	c, err := rc4.NewCipher(byte_key)
	if err != nil {
		panic(err)
	}
	plaintext := make([]byte, len(byte_data))
	c.XORKeyStream(plaintext, byte_data)
	return plaintext
}

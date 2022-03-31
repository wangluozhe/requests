package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	md4 := utils.HmacMD4("123", "123")
	bs64 := utils.Btoa(md4)
	hex := utils.HexEncode(md4)
	fmt.Println("HmacMD4-base64:", bs64)
	fmt.Println("HmacMD4-hex:", string(hex))

	r160 := utils.HmacRIPEMD160("123", "123")
	bs64 = utils.Btoa(r160)
	hex = utils.HexEncode(r160)
	fmt.Println("HmacRIPEMD160-base64:", bs64)
	fmt.Println("HmacRIPEMD160-hex:", string(hex))

	md5 := utils.HmacMD5("123", "123")
	bs64 = utils.Btoa(md5)
	hex = utils.HexEncode(md5)
	fmt.Println("HmacMD5-base64:", bs64)
	fmt.Println("HmacMD5-hex:", string(hex))

	sha1 := utils.HmacSHA1("123", "123")
	bs64 = utils.Btoa(sha1)
	hex1 := utils.HexEncode(sha1)
	fmt.Println("HmacSHA1-base64:", bs64)
	fmt.Println("HmacSHA1-hex:", string(hex1))

	sha224 := utils.HmacSHA224("123", "123")
	bs64 = utils.Btoa(sha224)
	hex2 := utils.HexEncode(sha224)
	fmt.Println("HmacSHA224-base64:", bs64)
	fmt.Println("HmacSHA224-hex:", string(hex2))

	sha256 := utils.HmacSHA256("123", "123")
	bs64 = utils.Btoa(sha256)
	hex3 := utils.HexEncode(sha256)
	fmt.Println("HmacSHA256-base64:", bs64)
	fmt.Println("HmacSHA256-hex:", string(hex3))

	sha384 := utils.HmacSHA384("123", "123")
	bs64 = utils.Btoa(sha384)
	hex4 := utils.HexEncode(sha384)
	fmt.Println("HmacSHA384-base64:", bs64)
	fmt.Println("HmacSHA384-hex:", string(hex4))

	sha512 := utils.HmacSHA512("123", "123")
	bs64 = utils.Btoa(sha512)
	hex5 := utils.HexEncode(sha512)
	fmt.Println("HmacSHA512-base64:", bs64)
	fmt.Println("HmacSHA512-hex:", string(hex5))
}

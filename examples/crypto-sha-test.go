package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	s1 := utils.SHA1("123")
	b64 := utils.Btoa(s1)
	h16 := utils.HexEncode(s1)
	fmt.Println("SHA1-base64:", b64)
	fmt.Println("SHA1-hex:", string(h16))

	s2 := utils.SHA224("123")
	b64 = utils.Btoa(s2)
	h16 = utils.HexEncode(s2)
	fmt.Println("SHA224-base64:", b64)
	fmt.Println("SHA224-hex:", string(h16))

	s2 = utils.SHA256("123")
	b64 = utils.Btoa(s2)
	h16 = utils.HexEncode(s2)
	fmt.Println("SHA256-base64:", b64)
	fmt.Println("SHA256-hex:", string(h16))

	s5 := utils.SHA384("123")
	b64 = utils.Btoa(s5)
	h16 = utils.HexEncode(s5)
	fmt.Println("SHA384-base64:", b64)
	fmt.Println("SHA384-hex:", string(h16))

	s5 = utils.SHA512("123")
	b64 = utils.Btoa(s5)
	h16 = utils.HexEncode(s5)
	fmt.Println("SHA512-base64:", b64)
	fmt.Println("SHA512-hex:", string(h16))
}

package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	rc4 := utils.RC4("123", "123")
	fmt.Println("RC4-base64:", utils.Btoa(rc4))
	fmt.Println("RC4-hex:",string(utils.HexEncode(rc4)))
}

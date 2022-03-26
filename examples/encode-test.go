package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	url := "https://www.baidu.com"
	encode := utils.EncodeURIComponent(url)
	fmt.Println(encode)
	decode := utils.DecodeURIComponent(encode)
	fmt.Println(decode)
	base64en := utils.Base64Encode(url)
	fmt.Println(base64en)
	base64de := utils.Base64Decode(base64en)
	fmt.Println(base64de)
	btoa := utils.Btoa(url)
	fmt.Println(btoa)
	atob := utils.Atob(btoa)
	fmt.Println(atob)
}
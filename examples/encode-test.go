package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	url := "https://www.baidu.com?page=10&abc=123&name=你好啊"
	hexen := utils.HexEncode(url)
	fmt.Println(hexen)
	fmt.Println(string(hexen))
	hexde := utils.HexDecode(hexen)
	fmt.Println(hexde)
	fmt.Println(string(hexde))
	encode := utils.EncodeURIComponent(url)
	fmt.Println(encode)
	decode := utils.DecodeURIComponent(encode)
	fmt.Println(decode)
	encode1 := utils.EncodeURI(url)
	fmt.Println(encode1)
	decode1 := utils.DecodeURI(encode1)
	fmt.Println(decode1)
	base32en := utils.Base32Encode(url)
	fmt.Println(base32en)
	base32de := utils.Base32Decode(base32en)
	fmt.Println(base32de)
	base64en := utils.Base64Encode(url)
	fmt.Println(base64en)
	base64de := utils.Base64Decode(base64en)
	fmt.Println(base64de)
	btoa := utils.Btoa(url)
	fmt.Println(btoa)
	atob := utils.Atob(btoa)
	fmt.Println(atob)
	escape := utils.Escape(url)
	fmt.Println(escape)
	unescape := utils.UnEscape(escape)
	fmt.Println(unescape)
}
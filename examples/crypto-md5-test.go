package main

import (
	"fmt"
	"github.com/wangluozhe/requests/utils"
)

func main() {
	m := utils.MD4("123")
	fmt.Println("MD4:", m)

	m = utils.RIPEMD160("123")
	fmt.Println("RIPEMD160:", m)

	m = utils.MD5("123")
	fmt.Println("MD5:", m)
}

package main

import (
	"fmt"
	"github.com/wangluozhe/requests/url"
)

func main() {
	params := map[string]interface{}{
		"ip":   "127.0.0.1",
		"port": 8888,
	}
	p := url.ParseParams(params)
	fmt.Println(p)
}

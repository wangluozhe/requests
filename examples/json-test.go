package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	req.Json = map[string]interface{}{
		"some":   "data",
		"name":   "测试",
		"colors": []string{"蓝色", "绿色", "紫色"},
		"data": map[string]interface{}{
			"json": true,
			"age":  15,
		},
	}
	r, err := requests.Post("http://httpbin.org/post", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.Text)
}

package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	headers := url.NewHeaders()
	headers.Set("User-Agent","123")
	req.Headers = headers
	req.Proxies = "http://127.0.0.1:8888"
	//req := &url.Request{}
	r, _ := requests.Get("https://httpbin.org/get",req)
	//r, _ := requests.Get("https://httpbin.org/get",nil)
	fmt.Println(r.Text)
}

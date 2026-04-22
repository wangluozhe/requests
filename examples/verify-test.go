package main

import (
	"fmt"

	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	headers := url.NewHeaders()
	req.Headers = headers
	req.Verify = url.Bool(false)
	req.Proxies = "http://127.0.0.1:7890"
	r, err := requests.Get("https://tls.peet.ws/api/all", req)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(r.Text())
}

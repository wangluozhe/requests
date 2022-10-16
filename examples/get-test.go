package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	headers := url.NewHeaders()
	headers.Set("User-Agent", "123")
	req.Headers = headers
	req.Proxies = url.Proxies{
		//2022-11-10日测试代理到期
		Scheme:   url.PROXIES_SCHEME_HTTP,
		Host:     "125.124.226.10",
		Port:     "4848",
		User:     "yzxsk1665463105",
		Password: "ha9yi8",
	}

	r, _ := requests.Get("https://httpbin.org/get", req)

	fmt.Println(r.Text)
}

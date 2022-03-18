package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	headers := url.NewHeaders()
	headers.Set("Path", "/get")
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
	headers.Set("accept-language", "zh-CN,zh;q=0.9")
	headers.Set("Scheme", "https")
	headers.Set("accept-encoding", "gzip, deflate, br")
	headers.Set("Host", "httpbin.org")
	headers.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	(*headers)["Header-Order:"] = []string{
		"user-agent",
		"path",
		"accept-language",
		"scheme",
		"connection",
		"accept-encoding",
		"content-length",
		"host",
		"accept",
	}
	//headers := &http.Header{
	//	"authority": {"www.wayfair.com"},
	//	"Accept": {"*/*"},
	//	"Accept-Encoding": {"gzip, deflate, br"},
	//	"Authority": {"www.wayfair.com"},
	//	//"Content-Length": {"0"},
	//	"Host": {"httpbin.org"},
	//	"User-Agent": {"golang-requests 1.0"},
	//	http.HeaderOrderKey: []string{
	//		"user-agent",
	//		"host",
	//		"accept-encoding",
	//		"connection",
	//		"accept",
	//		"authority",
	//		"content-length",
	//	},
	//}
	req.Headers = headers
	r, err := requests.Get("https://httpbin.org/get", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.Request.Headers)
	fmt.Println("url:", r.Url)
	fmt.Println("headers:", r.Headers)
	fmt.Println("text:", r.Text)
}

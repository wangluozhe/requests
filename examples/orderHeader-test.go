package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	//headers := url.NewHeaders()
	//headers.Set("Path", "/get")
	//headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
	//headers.Set("accept-language", "zh-CN,zh;q=0.9")
	//headers.Set("Scheme", "https")
	//headers.Set("accept-encoding", "gzip, deflate, br")
	//headers.Set("Host", "httpbin.org")
	//headers.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	//(*headers)["Header-Order:"] = []string{
	//	"user-agent",
	//	"path",
	//	"accept-language",
	//	"scheme",
	//	"connection",
	//	"accept-encoding",
	//	"content-length",
	//	"host",
	//	"accept",
	//}

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

	req.Headers = url.ParseHeaders(`
	:authority: spidertools.cn
	:method: GET
	:path: /
	:scheme: https
	accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9
	accept-encoding: gzip, deflate, br
	accept-language: zh-CN,zh;q=0.9
	cache-control: no-cache
	cookie: _ga=GA1.1.630251354.1645893020; Hm_lvt_def79de877408c7bd826e49b694147bc=1647245863,1647936048,1648296630; Hm_lpvt_def79de877408c7bd826e49b694147bc=1648296630
	pragma: no-cache
	sec-ch-ua: " Not A;Brand";v="99", "Chromium";v="98", "Google Chrome";v="98"
	sec-ch-ua-mobile: ?0
	sec-ch-ua-platform: "Windows"
	sec-fetch-dest: document
	sec-fetch-mode: navigate
	sec-fetch-site: same-origin
	sec-fetch-user: ?1
	upgrade-insecure-requests: 1
	user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36
	`)
	r, err := requests.Get("https://httpbin.org/get", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.Request.Headers)
	fmt.Println("url:", r.Url)
	fmt.Println("headers:", r.Headers)
	fmt.Println("text:", r.Text)
}

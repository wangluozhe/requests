package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	headers := url.NewHeaders()
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
	req.Headers = headers
	req.Proxies = "http://127.0.0.1:7890"
	req.Ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-21,29-23-24,0"
	r, err := requests.Get("https://ja3er.com/json", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.Request.Headers)
	fmt.Println("url:", r.Url)
	fmt.Println("headers:", r.Headers)
	fmt.Println("text:", r.Text)
}
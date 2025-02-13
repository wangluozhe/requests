package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
	"time"
)

func main() {
	session := requests.NewSession()
	req := url.NewRequest()
	req.Timeout = 100000 * time.Second
	headers := url.NewHeaders()
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
	req.Headers = headers
	req.Ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,5-0-16-10-35-65037-27-13-17513-11-18-23-45-65281-51-43-41,29-23-24,0"
	r1, err := session.Get("https://tls.peet.ws/api/all", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r1.Request.Headers)
	fmt.Println("url:", r1.Url)
	fmt.Println("headers:", r1.Headers)
	fmt.Println("text:", r1.Text)
	r2, err := session.Get("https://tls.peet.ws/api/all", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r2.Request.Headers)
	fmt.Println("url:", r2.Url)
	fmt.Println("headers:", r2.Headers)
	fmt.Println("text:", r2.Text)
}

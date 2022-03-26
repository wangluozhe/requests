package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	rawUrl := "http://httpbin.org/cookies"
	req := url.NewRequest()
	req.Cookies = url.ParseCookies(rawUrl,"_ga=GA1.1.630251354.1645893020; Hm_lvt_def79de877408c7bd826e49b694147bc=1647245863,1647936048,1648296630; Hm_lpvt_def79de877408c7bd826e49b694147bc=1648301329")
	r, err := requests.Get(rawUrl, req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.Text)
}
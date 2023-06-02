package main

import (
	"log"

	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func getIp(session *requests.Session, proxy string) {
	req := url.NewRequest()
	req.Proxies = proxy
	headers := url.NewHeaders()
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36")
	req.Headers = headers
	r, err := session.Get("https://tools.scrapfly.io/api/fp/anything", req)
	if err != nil {
		log.Panic(err)
	} else {
		log.Print(r.Text)
	}
}
func main() {
	session := requests.NewSession()
	getIp(session, "")
}

package url

import (
	"errors"
	http "github.com/Danny-Dasilva/fhttp"
	"github.com/Danny-Dasilva/fhttp/cookiejar"
	"net/url"
	"strings"
)

func NewCookies() *cookiejar.Jar {
	cookies, _ := cookiejar.New(nil)
	return cookies
}

func ParseCookies(rawurl, cookies string) *cookiejar.Jar {
	c := NewCookies()
	cookies = "_ga=GA1.1.630251354.1645893020; Hm_lvt_def79de877408c7bd826e49b694147bc=1647245863,1647936048,1648296630; Hm_lpvt_def79de877408c7bd826e49b694147bc=1648301329"
	cookieList := strings.Split(cookies, ";")
	urls, _ := url.Parse(rawurl)
	for _, cookie := range cookieList {
		if cookie == "" {
			continue
		}
		cookie = strings.TrimSpace(cookie)
		keyValue := strings.Split(cookie, "=")
		if len(keyValue) != 2 {
			panic(errors.New("该字符串不符合Cookies标准"))
		}
		key := keyValue[0]
		value := keyValue[1]
		c.SetCookies(urls, []*http.Cookie{&http.Cookie{
			Name:  key,
			Value: value,
		}})
	}
	return c
}

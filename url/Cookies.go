package url

import (
	"errors"
	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/chttp/cookiejar"
	"net/url"
	"strconv"
	"strings"
)

func NewCookies() *cookiejar.Jar {
	cookies, _ := cookiejar.New(nil)
	return cookies
}

func ParseCookies(rawurl string, cookies interface{}) *cookiejar.Jar {
	urls, _ := url.Parse(rawurl)
	jar := NewCookies()
	switch cookies.(type) {
	case string:
		var cookie_list []*http.Cookie
		cookieList := strings.Split(cookies.(string), ";")
		for _, cookie := range cookieList {
			cookie = strings.TrimSpace(cookie)
			if cookie == "" {
				continue
			}
			keyValue := strings.SplitN(cookie, "=", 2)
			if len(keyValue) != 2 {
				panic(errors.New("该字符串不符合Cookies标准"))
			}
			key := keyValue[0]
			value := keyValue[1]
			cookie_list = append(cookie_list, &http.Cookie{
				Name:  key,
				Value: value,
			})
		}
		jar.SetCookies(urls, cookie_list)
	case map[string]string:
		var cookie_list []*http.Cookie
		v := cookies.(map[string]string)
		for key, value := range v {
			value = strings.TrimSpace(value)
			if key == "" || value == "" {
				continue
			}
			cookie_list = append(cookie_list, &http.Cookie{
				Name:  key,
				Value: value,
			})
		}
		jar.SetCookies(urls, cookie_list)
	case map[string]int:
		var cookie_list []*http.Cookie
		v := cookies.(map[string]int)
		for key, value := range v {
			val := strings.TrimSpace(strconv.Itoa(value))
			if key == "" || val == "" {
				continue
			}
			cookie_list = append(cookie_list, &http.Cookie{
				Name:  key,
				Value: val,
			})
		}
		jar.SetCookies(urls, cookie_list)
	case map[string]float64:
		var cookie_list []*http.Cookie
		v := cookies.(map[string]float64)
		for key, value := range v {
			val := strings.TrimSpace(strconv.Itoa(int(value)))
			if key == "" || val == "" {
				continue
			}
			cookie_list = append(cookie_list, &http.Cookie{
				Name:  key,
				Value: val,
			})
		}
		jar.SetCookies(urls, cookie_list)
	case map[string]interface{}:
		var cookie_list []*http.Cookie
		v := cookies.(map[string]interface{})
		for key, value := range v {
			switch value.(type) {
			case string:
				cookie_list = append(cookie_list, &http.Cookie{
					Name:  key,
					Value: value.(string),
				})
			case int:
				cookie_list = append(cookie_list, &http.Cookie{
					Name:  key,
					Value: strconv.Itoa(value.(int)),
				})
			case float64:
				cookie_list = append(cookie_list, &http.Cookie{
					Name:  key,
					Value: strconv.Itoa(int(value.(float64))),
				})
			case bool:
				cookie_list = append(cookie_list, &http.Cookie{
					Name:  key,
					Value: strconv.FormatBool(value.(bool)),
				})
			}
		}
		jar.SetCookies(urls, cookie_list)
	}
	return jar
}

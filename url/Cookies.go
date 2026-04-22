package url

import (
	"net/url"
	"strconv"
	"strings"

	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/chttp/cookiejar"
)

func NewCookies() *cookiejar.Jar {
	cookies, _ := cookiejar.New(nil)
	return cookies
}

func parseStringCookies(cookies string) []*http.Cookie {
	var cookieList []*http.Cookie
	for _, cookie := range strings.Split(cookies, ";") {
		cookie = strings.TrimSpace(cookie)
		if cookie == "" {
			continue
		}

		key, value, ok := strings.Cut(cookie, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		cookieList = append(cookieList, &http.Cookie{
			Name:  key,
			Value: strings.TrimSpace(value),
		})
	}
	return cookieList
}

func parseMapCookies(cookies map[string]interface{}) []*http.Cookie {
	var cookieList []*http.Cookie
	for key, value := range cookies {
		var val string
		switch v := value.(type) {
		case string:
			val = v
		case int:
			val = strconv.Itoa(v)
		case float64:
			val = strconv.Itoa(int(v))
		case bool:
			val = strconv.FormatBool(v)
		default:
			continue
		}
		if key == "" || val == "" {
			continue
		}
		cookieList = append(cookieList, &http.Cookie{
			Name:  key,
			Value: val,
		})
	}
	return cookieList
}

func ParseCookies(rawurl string, cookies interface{}) *cookiejar.Jar {
	urls, _ := url.Parse(rawurl)
	jar := NewCookies()
	var cookieList []*http.Cookie

	switch v := cookies.(type) {
	case string:
		cookieList = parseStringCookies(v)
	case map[string]string:
		cookieList = parseMapCookies(convertToInterfaceMap(v))
	case map[string]int:
		cookieList = parseMapCookies(convertToInterfaceMap(v))
	case map[string]float64:
		cookieList = parseMapCookies(convertToInterfaceMap(v))
	case map[string]interface{}:
		cookieList = parseMapCookies(v)
	}

	jar.SetCookies(urls, cookieList)
	return jar
}

func convertToInterfaceMap(m interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	switch v := m.(type) {
	case map[string]string:
		for key, value := range v {
			result[key] = value
		}
	case map[string]int:
		for key, value := range v {
			result[key] = value
		}
	case map[string]float64:
		for key, value := range v {
			result[key] = value
		}
	}
	return result
}

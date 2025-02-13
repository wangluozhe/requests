package url

import (
	"errors"
	http "github.com/wangluozhe/chttp"
	"strconv"
	"strings"
)

// 初始化Headers结构体
func NewHeaders() *http.Header {
	headers := &http.Header{}
	(*headers)[http.PHeaderOrderKey] = []string{":method", ":authority", ":scheme", ":path"}
	return headers
}

// 解析Headers字符串为结构体
func ParseHeaders(headers interface{}) *http.Header {
	h := NewHeaders()
	headerOrder := []string{}
	pHeaderOrder := []string{}

	addHeader := func(key, value string, isPseudo bool) {
		key = strings.ToLower(key)
		if isPseudo {
			if SearchStrings((*h)[http.PHeaderOrderKey], key) == -1 || SearchStrings(pHeaderOrder, key) != -1 {
				return
			}
			pHeaderOrder = append(pHeaderOrder, key)
		} else {
			h.Add(key, value)
			headerOrder = append(headerOrder, key)
		}
	}

	switch v := headers.(type) {
	case string:
		lines := strings.Split(v, "\n")
		for _, header := range lines {
			header = strings.TrimSpace(header)
			if header == "" || strings.HasPrefix(header, "/") || strings.HasPrefix(header, "#") {
				continue
			}
			keyValue := strings.SplitN(header, ":", 2)
			if len(keyValue) != 2 {
				panic(errors.New("该字符串不符合http头部标准！"))
			}
			addHeader(keyValue[0], keyValue[1], strings.HasPrefix(header, ":"))
		}
	case map[string]string:
		for key, value := range v {
			isPseudo := strings.HasPrefix(key, ":")
			addHeader(key, value, isPseudo)
		}
	case map[string]interface{}:
		for key, value := range v {
			isPseudo := strings.HasPrefix(key, ":")
			switch val := value.(type) {
			case string:
				addHeader(key, val, isPseudo)
			case int:
				addHeader(key, strconv.Itoa(val), isPseudo)
			case float64:
				addHeader(key, strconv.Itoa(int(val)), isPseudo)
			case bool:
				addHeader(key, strconv.FormatBool(val), isPseudo)
			}
		}
	case map[string][]string:
		for key, values := range v {
			isPseudo := strings.HasPrefix(key, ":")
			for _, value := range values {
				addHeader(key, value, isPseudo)
			}
		}
	case map[string][]interface{}:
		for key, values := range v {
			isPseudo := strings.HasPrefix(key, ":")
			for _, value := range values {
				switch val := value.(type) {
				case string:
					addHeader(key, val, isPseudo)
				case int:
					addHeader(key, strconv.Itoa(val), isPseudo)
				case float64:
					addHeader(key, strconv.Itoa(int(val)), isPseudo)
				case bool:
					addHeader(key, strconv.FormatBool(val), isPseudo)
				}
			}
		}
	}

	(*h)[http.HeaderOrderKey] = headerOrder
	if len(pHeaderOrder) == 4 {
		(*h)[http.PHeaderOrderKey] = pHeaderOrder
	}
	return h
}

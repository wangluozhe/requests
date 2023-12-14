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
	switch headers.(type) {
	case string:
		lines := strings.Split(headers.(string), "\n")
		for _, header := range lines {
			header = strings.TrimSpace(header)
			if header == "" || strings.Index(header, "/") == 0 || strings.Index(header, "#") == 0 {
				continue
			} else if strings.Index(header, ":") == 0 {
				header = strings.TrimLeft(header, ":")
				keyValue := strings.SplitN(header, ":", 2)
				if len(keyValue) != 2 {
					panic(errors.New("该字符串不符合http头部标准！"))
				}
				key := ":" + strings.ToLower(keyValue[0])
				if SearchStrings((*h)[http.PHeaderOrderKey], key) == -1 || SearchStrings(pHeaderOrder, key) != -1 {
					continue
				}
				pHeaderOrder = append(pHeaderOrder, key)
			} else {
				keyValue := strings.SplitN(header, ":", 2)
				if len(keyValue) != 2 {
					panic(errors.New("该字符串不符合http头部标准！"))
				}
				key := keyValue[0]
				value := keyValue[1]
				h.Set(key, value)
				headerOrder = append(headerOrder, strings.ToLower(key))
			}
		}
	case map[string]string:
		for key, value := range headers.(map[string]string) {
			key = strings.ToLower(key)
			if strings.Index(key, ":") == 0 {
				if SearchStrings((*h)[http.PHeaderOrderKey], key) == -1 || SearchStrings(pHeaderOrder, key) != -1 {
					continue
				}
				pHeaderOrder = append(pHeaderOrder, key)
			} else {
				h.Add(key, value)
				headerOrder = append(headerOrder, key)
			}
		}
	case map[string]interface{}:
		for key, value := range headers.(map[string]interface{}) {
			key = strings.ToLower(key)
			if strings.Index(key, ":") == 0 {
				if SearchStrings((*h)[http.PHeaderOrderKey], key) == -1 || SearchStrings(pHeaderOrder, key) != -1 {
					continue
				}
				pHeaderOrder = append(pHeaderOrder, key)
			} else {
				switch value.(type) {
				case string:
					h.Add(key, value.(string))
				case int:
					h.Add(key, strconv.Itoa(value.(int)))
				case float64:
					h.Add(key, strconv.Itoa(int(value.(float64))))
				case bool:
					h.Add(key, strconv.FormatBool(value.(bool)))
				}
				headerOrder = append(headerOrder, key)
			}
		}
	case map[string][]string:
		for key, values := range headers.(map[string][]string) {
			key = strings.ToLower(key)
			if strings.Index(key, ":") == 0 {
				if SearchStrings((*h)[http.PHeaderOrderKey], key) == -1 || SearchStrings(pHeaderOrder, key) != -1 {
					continue
				}
				pHeaderOrder = append(pHeaderOrder, key)
			} else {
				for _, value := range values {
					h.Add(key, value)
					headerOrder = append(headerOrder, key)
				}
			}
		}
	case map[string][]interface{}:
		for key, values := range headers.(map[string]interface{}) {
			key = strings.ToLower(key)
			if strings.Index(key, ":") == 0 {
				if SearchStrings((*h)[http.PHeaderOrderKey], key) == -1 || SearchStrings(pHeaderOrder, key) != -1 {
					continue
				}
				pHeaderOrder = append(pHeaderOrder, key)
			} else {
				for _, value := range values.([]interface{}) {
					switch value.(type) {
					case string:
						h.Add(key, value.(string))
					case int:
						h.Add(key, strconv.Itoa(value.(int)))
					case float64:
						h.Add(key, strconv.Itoa(int(value.(float64))))
					case bool:
						h.Add(key, strconv.FormatBool(value.(bool)))
					}
				}
				headerOrder = append(headerOrder, key)
			}
		}
	}
	(*h)[http.HeaderOrderKey] = headerOrder
	if len(pHeaderOrder) == 4 {
		(*h)[http.PHeaderOrderKey] = pHeaderOrder
	}
	return h
}

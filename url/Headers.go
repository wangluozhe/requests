package url

import (
	"strconv"
	"strings"

	http "github.com/wangluozhe/chttp"
)

var defaultPseudoHeaderOrder = []string{":method", ":authority", ":scheme", ":path"}

// 初始化Headers结构体
func NewHeaders() *http.Header {
	headers := &http.Header{}
	(*headers)[http.PHeaderOrderKey] = append([]string(nil), defaultPseudoHeaderOrder...)
	return headers
}

func parseHeaderLine(header string) (key, value string, isPseudo, ok bool) {
	header = strings.TrimSpace(header)
	if header == "" || strings.HasPrefix(header, "/") || strings.HasPrefix(header, "#") {
		return "", "", false, false
	}

	if strings.HasPrefix(header, ":") {
		name, value, found := strings.Cut(header[1:], ":")
		if !found {
			return "", "", true, false
		}
		name = strings.TrimSpace(name)
		if name == "" {
			return "", "", true, false
		}
		return ":" + name, strings.TrimSpace(value), true, true
	}

	var found bool
	key, value, found = strings.Cut(header, ":")
	if !found {
		return "", "", false, false
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", "", false, false
	}
	return key, strings.TrimSpace(value), false, true
}

// 解析Headers字符串为结构体
func ParseHeaders(headers interface{}) *http.Header {
	h := NewHeaders()
	headerOrder := []string{}
	pHeaderOrder := []string{}

	addHeader := func(key, value string, isPseudo bool) {
		lowerKey := strings.ToLower(key)
		if isPseudo {
			if SearchStrings(defaultPseudoHeaderOrder, lowerKey) == -1 || SearchStrings(pHeaderOrder, lowerKey) != -1 {
				return
			}
			pHeaderOrder = append(pHeaderOrder, lowerKey)
			(*h)[http.PHeaderOrderKey] = pHeaderOrder
			return
		}

		headerOrder = append(headerOrder, lowerKey)
		(*h)[http.HeaderOrderKey] = headerOrder
		h.Add(key, value)
	}

	switch v := headers.(type) {
	case string:
		lines := strings.Split(v, "\n")
		for _, header := range lines {
			key, value, isPseudo, ok := parseHeaderLine(header)
			if !ok {
				continue
			}
			addHeader(key, value, isPseudo)
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

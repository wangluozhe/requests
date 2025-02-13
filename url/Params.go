package url

import (
	"strconv"
	"strings"
	"sync"
)

// 查询下标
func SearchStrings(str []string, substr string) int {
	for index, value := range str {
		if value == substr {
			return index
		}
	}
	return -1
}

// 解析params字符串为Params结构体
func ParseParams(params interface{}) *Params {
	p := NewParams()
	switch v := params.(type) {
	case string:
		if v == "" {
			return p
		}
		for _, l := range strings.Split(v, "&") {
			value := strings.SplitN(l, "=", 2)
			if len(value) == 2 {
				p.Add(value[0], value[1])
			}
		}
	case map[string]string:
		for key, value := range v {
			p.Set(key, value)
		}
	case map[string][]string:
		for key, values := range v {
			for _, value := range values {
				p.Add(key, value)
			}
		}
	case map[string]int:
		for key, value := range v {
			p.Set(key, strconv.Itoa(value))
		}
	case map[string][]int:
		for key, values := range v {
			for _, value := range values {
				p.Add(key, strconv.Itoa(value))
			}
		}
	case map[string]float64:
		for key, value := range v {
			p.Set(key, strconv.FormatFloat(value, 'f', -1, 64))
		}
	case map[string][]float64:
		for key, values := range v {
			for _, value := range values {
				p.Add(key, strconv.FormatFloat(value, 'f', -1, 64))
			}
		}
	case map[string]interface{}:
		for key, value := range v {
			switch v := value.(type) {
			case string:
				p.Add(key, v)
			case []string:
				for _, s2 := range v {
					p.Add(key, s2)
				}
			case int:
				p.Add(key, strconv.Itoa(v))
			case []int:
				for _, s2 := range v {
					p.Add(key, strconv.Itoa(s2))
				}
			case float64:
				p.Add(key, strconv.FormatFloat(v, 'f', -1, 64))
			case []float64:
				for _, s2 := range v {
					p.Add(key, strconv.FormatFloat(s2, 'f', -1, 64))
				}
			case bool:
				p.Add(key, strconv.FormatBool(v))
			case []interface{}:
				for _, s2 := range v {
					switch s2 := s2.(type) {
					case string:
						p.Add(key, s2)
					case int:
						p.Add(key, strconv.Itoa(s2))
					case float64:
						p.Add(key, strconv.FormatFloat(s2, 'f', -1, 64))
					case bool:
						p.Add(key, strconv.FormatBool(s2))
					}
				}
			}
		}
	}
	return p
}

// 初始化Params结构体
func NewParams() *Params {
	return &Params{}
}

// Params结构体
type Params struct {
	values   sync.Map
	indexKey []string
}

// 设置Params参数
func (p *Params) Set(key, value string) {
	p.values.Store(key, []string{value})
	if SearchStrings(p.indexKey, key) == -1 {
		p.indexKey = append(p.indexKey, key)
	}
}

// 获取Params参数值
func (p *Params) Get(key string) string {
	if value, ok := p.values.Load(key); ok {
		return value.([]string)[0]
	}
	return ""
}

// 添加Params参数
func (p *Params) Add(key, value string) {
	if val, ok := p.values.Load(key); ok {
		p.values.Store(key, append(val.([]string), value))
	} else {
		p.Set(key, value)
	}
}

// 删除Params参数
func (p *Params) Del(key string) {
	if _, ok := p.values.Load(key); ok {
		p.values.Delete(key)
		if index := SearchStrings(p.indexKey, key); index != -1 {
			p.indexKey = append(p.indexKey[:index], p.indexKey[index+1:]...)
		}
	}
}

// 获取Params的所有Key
func (p *Params) Keys() []string {
	return p.indexKey
}

// Params结构体转字符串
func (p *Params) Encode() string {
	var text []string
	for _, key := range p.indexKey {
		if item, ok := p.values.Load(key); ok {
			for _, value := range item.([]string) {
				text = append(text, key+"="+value)
			}
		}
	}
	return strings.Join(text, "&")
}

// Params结构体返回map[string][]string
func (p *Params) Values() map[string][]string {
	values := make(map[string][]string)
	p.values.Range(func(k, v interface{}) bool {
		values[k.(string)] = v.([]string)
		return true
	})
	return values
}

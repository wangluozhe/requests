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
	switch params.(type) {
	case string:
		v := params.(string)
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
		v := params.(map[string]string)
		if v == nil {
			return p
		}
		for key, value := range v {
			p.Set(key, value)
		}
	case map[string][]string:
		v := params.(map[string][]string)
		if v == nil {
			return p
		}
		for key, values := range v {
			for _, value := range values {
				p.Add(key, value)
			}
		}
	case map[string]int:
		v := params.(map[string]int)
		if v == nil {
			return p
		}
		for key, value := range v {
			p.Set(key, strconv.Itoa(value))
		}
	case map[string][]int:
		v := params.(map[string][]int)
		if v == nil {
			return p
		}
		for key, values := range v {
			for _, value := range values {
				p.Add(key, strconv.Itoa(value))
			}
		}
	case map[string]float64:
		v := params.(map[string]float64)
		if v == nil {
			return p
		}
		for key, value := range v {
			p.Set(key, strconv.Itoa(int(value)))
		}
	case map[string][]float64:
		v := params.(map[string][]float64)
		if v == nil {
			return p
		}
		for key, values := range v {
			for _, value := range values {
				p.Add(key, strconv.Itoa(int(value)))
			}
		}
	case map[string]interface{}:
		v := params.(map[string]interface{})
		for key, value := range v {
			switch value.(type) {
			case string:
				p.Add(key, value.(string))
			case []string:
				for _, s2 := range value.([]string) {
					p.Add(key, s2)
				}
			case int:
				p.Add(key, strconv.Itoa(value.(int)))
			case []int:
				for _, s2 := range value.([]int) {
					p.Add(key, strconv.Itoa(s2))
				}
			case float64:
				p.Add(key, strconv.Itoa(int(value.(float64))))
			case []float64:
				for _, s2 := range value.([]float64) {
					p.Add(key, strconv.Itoa(int(s2)))
				}
			case bool:
				p.Add(key, strconv.FormatBool(value.(bool)))
			case []interface{}:
				for _, s2 := range value.([]interface{}) {
					switch s2.(type) {
					case string:
						p.Add(key, s2.(string))
					case int:
						p.Add(key, strconv.Itoa(s2.(int)))
					case float64:
						p.Add(key, strconv.Itoa(int(s2.(float64))))
					case bool:
						p.Add(key, strconv.FormatBool(s2.(bool)))
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
	index := SearchStrings(p.indexKey, key)
	if index == -1 {
		p.indexKey = append(p.indexKey, key)
	}
}

// 获取Params参数值
func (p *Params) Get(key string) string {
	value, ok := p.values.Load(key)
	if ok {
		return value.([]string)[0]
	}
	return ""
}

// 添加Params参数
func (p *Params) Add(key, value string) {
	val, ok := p.values.Load(key)
	if !ok {
		p.Set(key, value)
	} else {
		p.values.Store(key, append(val.([]string), value))
	}
}

// 删除Params参数
func (p *Params) Del(key string) {
	_, ok := p.values.Load(key)
	if !ok {
		return
	}
	p.values.Delete(key)
	index := SearchStrings(p.indexKey, key)
	if index != -1 {
		p.indexKey = append(p.indexKey[:index], p.indexKey[index+1:]...)
	}
}

// 获取Params的所有Key
func (p *Params) Keys() []string {
	return p.indexKey
}

// Params结构体转字符串
func (p *Params) Encode() string {
	text := []string{}
	for _, key := range p.indexKey {
		item, _ := p.values.Load(key)
		for _, value := range item.([]string) {
			text = append(text, key+"="+value)
		}
	}
	return strings.Join(text, "&")
}

// Params结构体返回map[string][]string
func (p *Params) Values() map[string][]string {
	values := make(map[string][]string)
	p.values.Range(func(k, p interface{}) bool {
		values[k.(string)] = p.([]string)
		return true
	})
	return values
}

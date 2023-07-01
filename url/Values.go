package url

import (
	"github.com/wangluozhe/requests/utils"
	"strconv"
	"strings"
	"sync"
)

// 解析Values字符串为Values结构体
func ParseValues(data interface{}) *Values {
	p := NewValues()
	switch data.(type) {
	case string:
		v := data.(string)
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
		v := data.(map[string]string)
		if v == nil {
			return p
		}
		for key, value := range v {
			p.Set(key, value)
		}
	case map[string][]string:
		v := data.(map[string][]string)
		if v == nil {
			return p
		}
		for key, values := range v {
			for _, value := range values {
				p.Add(key, value)
			}
		}
	case map[string]int:
		v := data.(map[string]int)
		if v == nil {
			return p
		}
		for key, value := range v {
			p.Set(key, strconv.Itoa(value))
		}
	case map[string][]int:
		v := data.(map[string][]int)
		if v == nil {
			return p
		}
		for key, values := range v {
			for _, value := range values {
				p.Add(key, strconv.Itoa(value))
			}
		}
	case map[string]float64:
		v := data.(map[string]float64)
		if v == nil {
			return p
		}
		for key, value := range v {
			p.Set(key, strconv.Itoa(int(value)))
		}
	case map[string][]float64:
		v := data.(map[string][]float64)
		if v == nil {
			return p
		}
		for key, values := range v {
			for _, value := range values {
				p.Add(key, strconv.Itoa(int(value)))
			}
		}
	case map[string]interface{}:
		v := data.(map[string]interface{})
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

// 解析Data字符串为Values结构体
func ParseData(data interface{}) *Values {
	return ParseValues(data)
}

// 初始化Values结构体
func NewValues() *Values {
	return &Values{}
}

// 初始化Data结构体
func NewData() *Values {
	return &Values{}
}

// Values结构体
type Values struct {
	values   sync.Map
	indexKey []string
}

// 设置Values参数
func (v *Values) Set(key, value string) {
	v.values.Store(key, []string{value})
	index := SearchStrings(v.indexKey, key)
	if index == -1 {
		v.indexKey = append(v.indexKey, key)
	}
}

// 获取Values参数值
func (v *Values) Get(key string) string {
	value, ok := v.values.Load(key)
	if ok {
		return value.([]string)[0]
	}
	return ""
}

// 添加Values参数
func (v *Values) Add(key, value string) {
	val, ok := v.values.Load(key)
	if !ok {
		v.Set(key, value)
	} else {
		v.values.Store(key, append(val.([]string), value))
	}
}

// 删除Values参数
func (v *Values) Del(key string) {
	_, ok := v.values.Load(key)
	if !ok {
		return
	}
	v.values.Delete(key)
	index := SearchStrings(v.indexKey, key)
	if index != -1 {
		v.indexKey = append(v.indexKey[:index], v.indexKey[index+1:]...)
	}
}

// 获取Values的所有Key
func (v *Values) Keys() []string {
	return v.indexKey
}

// Values结构体转字符串
func (v *Values) Encode() string {
	text := []string{}
	for _, key := range v.indexKey {
		item, _ := v.values.Load(key)
		for _, value := range item.([]string) {
			text = append(text, utils.EncodeURIComponent(key)+"="+utils.EncodeURIComponent(value))
		}
	}
	return strings.Join(text, "&")
}

// Values结构体返回map[string][]string
func (v *Values) Values() map[string][]string {
	values := make(map[string][]string)
	v.values.Range(func(k, v interface{}) bool {
		values[k.(string)] = v.([]string)
		return true
	})
	return values
}

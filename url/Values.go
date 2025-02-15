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
	switch v := data.(type) {
	case string:
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
			p.Set(key, strconv.Itoa(int(value)))
		}
	case map[string][]float64:
		for key, values := range v {
			for _, value := range values {
				p.Add(key, strconv.Itoa(int(value)))
			}
		}
	case map[string]interface{}:
		parseInterfaceMapValues(v, p)
	}
	return p
}

// 解析Data字符串为Values结构体
func ParseData(data interface{}) *Values {
	return ParseValues(data)
}

// 解析map[string]interface{}为Values结构体
func parseInterfaceMapValues(data map[string]interface{}, p *Values) {
	for key, value := range data {
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
			p.Add(key, strconv.Itoa(int(v)))
		case []float64:
			for _, s2 := range v {
				p.Add(key, strconv.Itoa(int(s2)))
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
					p.Add(key, strconv.Itoa(int(s2)))
				case bool:
					p.Add(key, strconv.FormatBool(s2))
				}
			}
		}
	}
}

// 初始化Values结构体
func NewValues() *Values {
	return &Values{mutex: &sync.RWMutex{}}
}

// 初始化Data结构体
func NewData() *Values {
	return &Values{mutex: &sync.RWMutex{}}
}

// Values结构体
type Values struct {
	values   sync.Map
	indexKey []string
	mutex    *sync.RWMutex
}

// 设置Values参数
func (v *Values) Set(key, value string) {
	v.values.Store(key, []string{value})
	v.mutex.Lock()
	defer v.mutex.Unlock()
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
	v.mutex.Lock()
	defer v.mutex.Unlock()
	index := SearchStrings(v.indexKey, key)
	if index != -1 {
		v.indexKey = append(v.indexKey[:index], v.indexKey[index+1:]...)
	}
}

// 获取Values的所有Key
func (v *Values) Keys() []string {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return append([]string(nil), v.indexKey...)
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

package url

import (
	"strconv"
	"strings"
	"sync"

	"github.com/wangluozhe/requests/utils"
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
			p.Set(key, strconv.FormatFloat(value, 'f', -1, 64))
		}
	case map[string][]float64:
		for key, values := range v {
			for _, value := range values {
				p.Add(key, strconv.FormatFloat(value, 'f', -1, 64))
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

// 初始化Values结构体
func NewValues() *Values {
	return &Values{values: make(map[string][]string)}
}

// 初始化Data结构体
func NewData() *Values {
	return NewValues()
}

// Values 按插入顺序保存表单值，零值可用，支持并发方法调用。使用后不能按值复制，请使用 Clone。
type Values struct {
	values   map[string][]string
	indexKey []string
	mutex    sync.RWMutex
}

// 设置Values参数
func (v *Values) Set(key, value string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.initLocked()
	if _, exists := v.values[key]; !exists {
		v.indexKey = append(v.indexKey, key)
	}
	v.values[key] = []string{value}
}

// 获取Values参数值
func (v *Values) Get(key string) string {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	if values := v.values[key]; len(values) > 0 {
		return values[0]
	}
	return ""
}

// 添加Values参数
func (v *Values) Add(key, value string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.initLocked()
	if _, exists := v.values[key]; !exists {
		v.indexKey = append(v.indexKey, key)
	}
	v.values[key] = append(v.values[key], value)
}

// 删除Values参数
func (v *Values) Del(key string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if _, exists := v.values[key]; !exists {
		return
	}
	delete(v.values, key)
	index := SearchStrings(v.indexKey, key)
	if index != -1 {
		copy(v.indexKey[index:], v.indexKey[index+1:])
		v.indexKey[len(v.indexKey)-1] = ""
		v.indexKey = v.indexKey[:len(v.indexKey)-1]
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
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	text := []string{}
	for _, key := range v.indexKey {
		for _, value := range v.values[key] {
			text = append(text, utils.EncodeURIComponent(key)+"="+utils.EncodeURIComponent(value))
		}
	}
	return strings.Join(text, "&")
}

// Values 返回独立的 map 和值切片，修改返回值不会影响原表单。
func (v *Values) Values() map[string][]string {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	values := make(map[string][]string, len(v.values))
	for key, item := range v.values {
		values[key] = append([]string(nil), item...)
	}
	return values
}

// Clone 返回 Values 的深拷贝。
func (v *Values) Clone() *Values {
	if v == nil {
		return nil
	}
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	clone := &Values{
		values:   make(map[string][]string, len(v.values)),
		indexKey: append([]string(nil), v.indexKey...),
	}
	for key, values := range v.values {
		clone.values[key] = append([]string(nil), values...)
	}
	return clone
}

func (v *Values) initLocked() {
	if v.values == nil {
		v.values = make(map[string][]string)
	}
}

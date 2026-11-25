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

// ParseParams 解析查询参数；字符串中的键和值保持原样，不进行百分号解码。
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
			} else if len(value) == 1 && value[0] != "" {
				p.Add(value[0], "")
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
	return &Params{values: make(map[string][]string)}
}

// Params 按插入顺序保存参数，零值可用，支持并发方法调用。使用后不能按值复制，请使用 Clone。
type Params struct {
	values   map[string][]string
	indexKey []string
	mutex    sync.RWMutex
}

// 设置Params参数
func (p *Params) Set(key, value string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.initLocked()
	if _, exists := p.values[key]; !exists {
		p.indexKey = append(p.indexKey, key)
	}
	p.values[key] = []string{value}
}

// 获取Params参数值
func (p *Params) Get(key string) string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	if values := p.values[key]; len(values) > 0 {
		return values[0]
	}
	return ""
}

// 添加Params参数
func (p *Params) Add(key, value string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.initLocked()
	if _, exists := p.values[key]; !exists {
		p.indexKey = append(p.indexKey, key)
	}
	p.values[key] = append(p.values[key], value)
}

// 删除Params参数
func (p *Params) Del(key string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if _, exists := p.values[key]; !exists {
		return
	}
	delete(p.values, key)
	if index := SearchStrings(p.indexKey, key); index != -1 {
		copy(p.indexKey[index:], p.indexKey[index+1:])
		p.indexKey[len(p.indexKey)-1] = ""
		p.indexKey = p.indexKey[:len(p.indexKey)-1]
	}
}

// 获取Params的所有Key
func (p *Params) Keys() []string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return append([]string(nil), p.indexKey...)
}

// Encode 按插入顺序原样拼接查询参数，空值只输出键名，与旧版行为一致。
// 键和值中的特殊字符应由调用方预先编码，本方法不进行百分号编码。
func (p *Params) Encode() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	var text []string
	for _, key := range p.indexKey {
		for _, value := range p.values[key] {
			if value == "" {
				text = append(text, key)
			} else {
				text = append(text, key+"="+value)
			}
		}
	}
	return strings.Join(text, "&")
}

// Values 返回独立的 map 和值切片，修改返回值不会影响 Params。
func (p *Params) Values() map[string][]string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	values := make(map[string][]string, len(p.values))
	for key, item := range p.values {
		values[key] = append([]string(nil), item...)
	}
	return values
}

// Clone 返回 Params 的深拷贝。
func (p *Params) Clone() *Params {
	if p == nil {
		return nil
	}
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	clone := &Params{
		values:   make(map[string][]string, len(p.values)),
		indexKey: append([]string(nil), p.indexKey...),
	}
	for key, values := range p.values {
		clone.values[key] = append([]string(nil), values...)
	}
	return clone
}

func (p *Params) initLocked() {
	if p.values == nil {
		p.values = make(map[string][]string)
	}
}

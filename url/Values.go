package url

import (
	"github.com/wangluozhe/requests/utils"
	"strings"
)

// 解析Values字符串为Values结构体
func ParseValues(params string) *Values {
	p := NewValues()
	if params == "" {
		return p
	}
	for _, l := range strings.Split(params, "&") {
		value := strings.SplitN(l, "=", 2)
		if len(value) == 2 {
			p.Add(value[0], value[1])
		}
	}
	return p
}

// 解析Data字符串为Values结构体
func ParseData(data string) *Values {
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
	values   map[string][]string
	indexKey []string
}

// 设置Values参数
func (v *Values) Set(key, value string) {
	if v.values == nil {
		v.values = make(map[string][]string)
	}
	v.values[key] = []string{
		value,
	}
	index := SearchStrings(v.indexKey, key)
	if index == -1 {
		v.indexKey = append(v.indexKey, key)
	}
}

// 获取Values参数值
func (v *Values) Get(key string) string {
	if _, ok := v.values[key]; ok {
		return v.values[key][0]
	}
	return ""
}

// 添加Values参数
func (v *Values) Add(key, value string) {
	index := SearchStrings(v.indexKey, key)
	if index == -1 {
		v.Set(key, value)
	} else {
		v.values[key] = append(v.values[key], value)
	}
	return
}

// 删除Values参数
func (v *Values) Del(key string) {
	index := SearchStrings(v.indexKey, key)
	if index == -1 {
		return
	}
	delete(v.values, key)
	v.indexKey = append(v.indexKey[:index], v.indexKey[index+1:]...)
	return
}

// 获取Values的所有Key
func (v *Values) Keys() []string {
	return v.indexKey
}

// Values结构体转字符串
func (v *Values) Encode() string {
	text := []string{}
	for _, key := range v.indexKey {
		item := v.values[key]
		for _, value := range item {
			text = append(text, utils.EncodeURIComponent(key)+"="+utils.EncodeURIComponent(value))
		}
	}
	return strings.Join(text, "&")
}

// Values结构体返回values
func (v *Values) Values() map[string][]string {
	return v.values
}

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
		value := strings.Split(l, "=")
		if len(value) == 2{
			p.Add(value[0], value[1])
		}
	}
	return p
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
	values   []map[string][]string
	indexKey []string
}

// 设置Values参数
func (this *Values) Set(key, value string) {
	p := map[string][]string{
		key: []string{value,},
	}
	index := SearchStrings(this.indexKey, key)
	if len(this.indexKey) == 0 || index == -1 {
		this.values = append(this.values, p)
		this.indexKey = append(this.indexKey, key)
	} else {
		this.values[index] = p
	}
}

// 获取Values参数值
func (this *Values) Get(key string) string {
	if len(this.values) != 0 {
		index := SearchStrings(this.indexKey, key)
		if index != -1 {
			return this.values[index][key][0]
		}
		return ""
	}
	return ""
}

// 添加Values参数
func (this *Values) Add(key, value string) bool {
	index := SearchStrings(this.indexKey, key)
	if len(this.indexKey) == 0 || index == -1 {
		this.Set(key, value)
	} else {
		this.values[index][key] = append(this.values[index][key], value)
	}
	return true
}

// 删除Values参数
func (this *Values) Del(key string) bool {
	index := SearchStrings(this.indexKey, key)
	if len(this.indexKey) == 0 || index == -1 {
		return false
	}
	this.values = append(this.values[:index], this.values[index+1:]...)
	this.indexKey = append(this.indexKey[:index], this.indexKey[index+1:]...)
	return true
}

// 获取Values的所有Key
func (this *Values) Keys() []string {
	return this.indexKey
}

// Values结构体转字符串
func (this *Values) Encode() string {
	text := []string{}
	for index, key := range this.indexKey {
		item := this.values[index][key]
		for _, value := range item {
			text = append(text, utils.EncodeURIComponent(key)+"="+utils.EncodeURIComponent(value))
		}
	}
	return strings.Join(text, "&")
}

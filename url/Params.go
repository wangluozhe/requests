package url

import (
	"github.com/wangluozhe/requests/utils"
	"strings"
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
func ParseParams(params string) *Params {
	p := NewParams()
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

// 初始化Params结构体
func NewParams() *Params {
	return &Params{}
}

// Params结构体
type Params struct {
	params   []map[string][]string
	indexKey []string
}

// 设置Params参数
func (this *Params) Set(key, value string) {
	p := map[string][]string{
		key: []string{value,},
	}
	index := SearchStrings(this.indexKey, key)
	if len(this.indexKey) == 0 || index == -1 {
		this.params = append(this.params, p)
		this.indexKey = append(this.indexKey, key)
	} else {
		this.params[index] = p
	}
}

// 获取Params参数值
func (this *Params) Get(key string) string {
	if len(this.params) != 0 {
		index := SearchStrings(this.indexKey, key)
		if index != -1 {
			return this.params[index][key][0]
		}
	}
	return ""
}

// 添加Params参数
func (this *Params) Add(key, value string) bool {
	index := SearchStrings(this.indexKey, key)
	if len(this.indexKey) == 0 || index == -1 {
		this.Set(key, value)
	} else {
		this.params[index][key] = append(this.params[index][key], value)
	}
	return true
}

// 删除Params参数
func (this *Params) Del(key string) bool {
	index := SearchStrings(this.indexKey, key)
	if len(this.indexKey) == 0 || index == -1 {
		return false
	}
	this.params = append(this.params[:index], this.params[index+1:]...)
	this.indexKey = append(this.indexKey[:index], this.indexKey[index+1:]...)
	return true
}

// 获取Params的所有Key
func (this *Params) Keys() []string {
	return this.indexKey
}

// Params结构体转字符串
func (this *Params) Encode() string {
	text := []string{}
	for index, key := range this.indexKey {
		item := this.params[index][key]
		for _, value := range item {
			text = append(text, utils.EncodeURIComponent(key)+"="+utils.EncodeURIComponent(value))
		}
	}
	return strings.Join(text, "&")
}

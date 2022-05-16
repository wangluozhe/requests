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
func (p *Params) Set(key, value string) {
	pm := map[string][]string{
		key: []string{value,},
	}
	index := SearchStrings(p.indexKey, key)
	if len(p.indexKey) == 0 || index == -1 {
		p.params = append(p.params, pm)
		p.indexKey = append(p.indexKey, key)
	} else {
		p.params[index] = pm
	}
}

// 获取Params参数值
func (p *Params) Get(key string) string {
	if len(p.params) != 0 {
		index := SearchStrings(p.indexKey, key)
		if index != -1 {
			return p.params[index][key][0]
		}
	}
	return ""
}

// 添加Params参数
func (p *Params) Add(key, value string) bool {
	index := SearchStrings(p.indexKey, key)
	if len(p.indexKey) == 0 || index == -1 {
		p.Set(key, value)
	} else {
		p.params[index][key] = append(p.params[index][key], value)
	}
	return true
}

// 删除Params参数
func (p *Params) Del(key string) bool {
	index := SearchStrings(p.indexKey, key)
	if len(p.indexKey) == 0 || index == -1 {
		return false
	}
	p.params = append(p.params[:index], p.params[index+1:]...)
	p.indexKey = append(p.indexKey[:index], p.indexKey[index+1:]...)
	return true
}

// 获取Params的所有Key
func (p *Params) Keys() []string {
	return p.indexKey
}

// Params结构体转字符串
func (p *Params) Encode() string {
	text := []string{}
	for index, key := range p.indexKey {
		item := p.params[index][key]
		for _, value := range item {
			text = append(text, utils.EncodeURIComponent(key)+"="+utils.EncodeURIComponent(value))
		}
	}
	return strings.Join(text, "&")
}

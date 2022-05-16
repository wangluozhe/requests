package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Danny-Dasilva/fhttp"
	"github.com/bitly/go-simplejson"
	"github.com/wangluozhe/requests/url"
	"io"
)

// Response结构体
type Response struct {
	Url        string
	Headers    http.Header
	Cookies    []*http.Cookie
	Text       string
	Content    []byte
	Body       io.ReadCloser
	StatusCode int
	History    []*Response
	Request    *url.Request
}

// 使用自带库JSON解析
func (res *Response) Json() (map[string]interface{}, error) {
	js := make(map[string]interface{})
	err := json.Unmarshal(res.Content, &js)
	return js, err
}

// 使用go-simplejson解析
func (res *Response) SimpleJson() (*simplejson.Json, error) {
	return simplejson.NewFromReader(res.Body)
}

// 状态码是否错误
func (res *Response) RaiseForStatus() error {
	var err error
	if res.StatusCode >= 400 && res.StatusCode < 500 {
		err = errors.New(fmt.Sprintf("%d Client Error", res.StatusCode))
	} else if res.StatusCode >= 500 && res.StatusCode < 600 {
		err = errors.New(fmt.Sprintf("%d Server Error", res.StatusCode))
	}
	return err
}

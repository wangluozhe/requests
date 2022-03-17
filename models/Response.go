package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/wangluozhe/requests/url"
	"io"
	"github.com/Danny-Dasilva/fhttp"
)

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

func (this *Response) Json() (map[string]interface{}, error) {
	js := make(map[string]interface{})
	err := json.Unmarshal(this.Content, &js)
	return js, err
}

func (this *Response) RaiseForStatus() error {
	var err error
	if this.StatusCode >= 400 && this.StatusCode < 500{
		err = errors.New(fmt.Sprintf("%d Client Error", this.StatusCode))
	} else if this.StatusCode >= 500 && this.StatusCode < 600{
		err = errors.New(fmt.Sprintf("%d Server Error", this.StatusCode))
	}
	return err
}
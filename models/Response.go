package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests/url"
	"io"
)

var RedirectStatusCodes = []int{
	http.StatusMovedPermanently,
	http.StatusFound,
	http.StatusSeeOther,
	http.StatusTemporaryRedirect,
	http.StatusPermanentRedirect,
}

func inRedirectStatusCodes(statusCode int) bool {
	for _, code := range RedirectStatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

var PermanentRedirectStatusCodes = []int{
	http.StatusMovedPermanently,
	http.StatusPermanentRedirect,
}

func inPermanentRedirectStatusCodes(statusCode int) bool {
	for _, code := range PermanentRedirectStatusCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}

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

// 状态码是否合格
func (res Response) Ok() bool {
	// Returns True if :attr:`status_code` is less than 400, False if not.
	//
	// This attribute checks if the status code of the response is between
	// 400 and 600 to see if there was a client error or a server error. If
	// the status code is between 200 and 400, this will return True. This
	// is **not** a check to see if the response code is ``200 OK``.
	err := res.RaiseForStatus()
	if err != nil {
		return false
	}
	return true
}

func (res Response) IsRedirect() bool {
	// True if this Response is a well-formed HTTP redirect that could have
	// been processed automatically (by :meth:`Session.resolve_redirects`).
	if res.Headers.Get("locaion") == "" || inRedirectStatusCodes(res.StatusCode) {
		return false
	}
	return true
}

func (res Response) IsPermanentRedirect() bool {
	// True if this Response one of the permanent versions of redirect.
	if res.Headers.Get("locaion") == "" || inPermanentRedirectStatusCodes(res.StatusCode) {
		return false
	}
	return true
}

// 状态码是否错误
func (res *Response) RaiseForStatus() error {
	// Raises :class:`HTTPError`, if one occurred.
	var err error
	if res.StatusCode >= 400 && res.StatusCode < 500 {
		err = errors.New(fmt.Sprintf("%d Client Error", res.StatusCode))
	} else if res.StatusCode >= 500 && res.StatusCode < 600 {
		err = errors.New(fmt.Sprintf("%d Server Error", res.StatusCode))
	}
	return err
}

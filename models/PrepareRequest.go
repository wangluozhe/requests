package models

import (
	"bytes"
	"encoding/base64"
	jsonp "encoding/json"
	"errors"
	"fmt"
	"github.com/Danny-Dasilva/fhttp"
	"github.com/Danny-Dasilva/fhttp/cookiejar"
	"github.com/wangluozhe/requests/url"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

// HTTP的所有请求方法
var MethodNames = []string{http.MethodGet, http.MethodPost, http.MethodOptions, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodTrace}

// 是否为HTTP请求方法
func inMethod(method string) bool {
	for _, value := range MethodNames {
		if value == method {
			return true
		}
	}
	return false
}

func NewPrepareRequest() *PrepareRequest {
	return &PrepareRequest{}
}

// PrepareRequest结构体
type PrepareRequest struct {
	Method  string
	Url     string
	Headers *http.Header
	Cookies *cookiejar.Jar
	Body    io.ReadCloser
}

// 预处理所有数据
func (this *PrepareRequest) Prepare(method, url string, params *url.Params, headers *http.Header, cookies *cookiejar.Jar, data *url.Values, files *url.Files, json map[string]interface{}, auth []string) error {
	err := this.Prepare_method(method)
	if err != nil {
		return err
	}
	err = this.Prepare_url(url, params)
	if err != nil {
		return err
	}
	err = this.Prepare_headers(headers)
	if err != nil {
		return err
	}
	this.Prepare_cookies(cookies)
	err = this.Prepare_body(data, files, json)
	if err != nil {
		return err
	}
	err = this.Prepare_auth(auth, url)
	if err != nil {
		return err
	}
	return nil
}

// 预处理method
func (this *PrepareRequest) Prepare_method(method string) error {
	method = strings.ToUpper(method)
	if !inMethod(method) {
		return errors.New("Method does not conform to HTTP protocol!")
	}
	this.Method = method
	return nil
}

// 预处理url
func (this *PrepareRequest) Prepare_url(rawurl string, params *url.Params) error {
	rawurl = strings.TrimSpace(rawurl)
	urls, err := url.Parse(rawurl)
	if err != nil {
		return err
	}
	if urls.Scheme == "" {
		return errors.New(fmt.Sprintf("Invalid URL %s: No scheme supplied. Perhaps you meant http://%s?", rawurl, rawurl))
	} else if urls.Host == "" {
		return errors.New(fmt.Sprintf("Invalid URL %s: No host supplied", rawurl))
	}
	if urls.Path == "" {
		urls.Path = "/"
	}
	if params != nil {
		if urls.RawParams != "" {
			urls.Params = url.ParseParams(urls.RawParams + "&" + params.Encode())
		} else {
			urls.Params = url.ParseParams(params.Encode())
		}
	}
	this.Url = urls.String()
	return nil
}

// 预处理headers
func (this *PrepareRequest) Prepare_headers(headers *http.Header) error {
	this.Headers = url.NewHeaders()
	if headers != nil {
		for key, values := range *headers {
			if len(values) == 1 {
				this.Headers.Set(key, values[0])
			} else {
				for index, value := range values {
					if index == 0 {
						this.Headers.Set(key, value)
					} else {
						this.Headers.Add(key, value)
					}
				}
			}
		}
	}
	return nil
}

// 预处理body
func (this *PrepareRequest) Prepare_body(data *url.Values, files *url.Files, json map[string]interface{}) error {
	var body string
	var content_type string
	var err error
	if data == nil && json != nil {
		content_type = "application/json"
		json_byte, err := jsonp.Marshal(json)
		if err != nil {
			return err
		}
		body = string(json_byte)
	}
	if files != nil {
		var byteBuffer *bytes.Buffer
		if data != nil{
			for _, key := range data.Keys() {
				files.AddField(key, data.Get(key))
			}
		}
		byteBuffer, content_type, err = files.Encode()
		var body_byte []byte
		if byteBuffer != nil {
			body_byte, _ = ioutil.ReadAll(byteBuffer)
		}
		body = string(body_byte)
		if err != nil {
			return err
		}
	} else if data != nil {
		content_type = "application/x-www-form-urlencoded"
		body = data.Encode()
	}
	this.prepare_content_length(body)
	if content_type != "" && this.Headers.Get("Content-Type") == "" {
		this.Headers.Set("Content-Type", content_type)
	}
	this.Body = ioutil.NopCloser(strings.NewReader(body))
	return nil
}

// 预处理body大小
func (this *PrepareRequest) prepare_content_length(body string) {
	if body != "" {
		length := len(body)
		if length > 0 {
			this.Headers.Set("Content-Length", strconv.Itoa(length))
		}
	} else if (this.Method != "GET" || this.Method != "HEAD") && this.Headers.Get("Content-Length") == "" {
		this.Headers.Set("Content-Length", "0")
	}
}

// 预处理cookie
func (this *PrepareRequest) Prepare_cookies(cookies *cookiejar.Jar) {
	if cookies != nil {
		this.Cookies = cookies
	} else {
		this.Cookies, _ = cookiejar.New(nil)
	}
}

// 预处理auth
func (this *PrepareRequest) Prepare_auth(auth []string, rawurl string) error {
	if auth == nil {
		urls, err := url.Parse(rawurl)
		if err != nil {
			return err
		}
		user := urls.User.String()
		if user != "" {
			auth = append(auth, user)
		}
		pass, _ := urls.User.Password()
		if pass != "" {
			auth = append(auth, pass)
		}
	}
	if auth != nil && len(auth) == 2 {
		this.Headers.Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(strings.Join(auth, ":"))))
	}
	return nil
}

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
func (pr *PrepareRequest) Prepare(method, url string, params *url.Params, headers *http.Header, cookies *cookiejar.Jar, data *url.Values, files *url.Files, json map[string]interface{}, auth []string) error {
	err := pr.Prepare_method(method)
	if err != nil {
		return err
	}
	err = pr.Prepare_url(url, params)
	if err != nil {
		return err
	}
	err = pr.Prepare_headers(headers)
	if err != nil {
		return err
	}
	pr.Prepare_cookies(cookies)
	err = pr.Prepare_body(data, files, json)
	if err != nil {
		return err
	}
	err = pr.Prepare_auth(auth, url)
	if err != nil {
		return err
	}
	return nil
}

// 预处理method
func (pr *PrepareRequest) Prepare_method(method string) error {
	method = strings.ToUpper(method)
	if !inMethod(method) {
		return errors.New("Method does not conform to HTTP protocol!")
	}
	pr.Method = method
	return nil
}

// 预处理url
func (pr *PrepareRequest) Prepare_url(rawurl string, params *url.Params) error {
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
	pr.Url = urls.String()
	return nil
}

// 预处理headers
func (pr *PrepareRequest) Prepare_headers(headers *http.Header) error {
	pr.Headers = url.NewHeaders()
	if headers != nil {
		for key, values := range *headers {
			if len(values) == 1 {
				pr.Headers.Set(key, values[0])
			} else {
				for index, value := range values {
					if index == 0 {
						pr.Headers.Set(key, value)
					} else {
						pr.Headers.Add(key, value)
					}
				}
			}
		}
	}
	return nil
}

// 预处理body
func (pr *PrepareRequest) Prepare_body(data *url.Values, files *url.Files, json map[string]interface{}) error {
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
	pr.prepare_content_length(body)
	if content_type != "" && pr.Headers.Get("Content-Type") == "" {
		pr.Headers.Set("Content-Type", content_type)
	}
	pr.Body = ioutil.NopCloser(strings.NewReader(body))
	return nil
}

// 预处理body大小
func (pr *PrepareRequest) prepare_content_length(body string) {
	if body != "" {
		length := len(body)
		if length > 0 {
			pr.Headers.Set("Content-Length", strconv.Itoa(length))
		}
	} else if (pr.Method != "GET" || pr.Method != "HEAD") && pr.Headers.Get("Content-Length") == "" {
		pr.Headers.Set("Content-Length", "0")
	}
}

// 预处理cookie
func (pr *PrepareRequest) Prepare_cookies(cookies *cookiejar.Jar) {
	if cookies != nil {
		pr.Cookies = cookies
	} else {
		pr.Cookies, _ = cookiejar.New(nil)
	}
}

// 预处理auth
func (pr *PrepareRequest) Prepare_auth(auth []string, rawurl string) error {
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
		pr.Headers.Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(strings.Join(auth, ":"))))
	}
	return nil
}

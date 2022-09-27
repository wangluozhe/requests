package requests

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/Danny-Dasilva/fhttp"
	"github.com/Danny-Dasilva/fhttp/cookiejar"
	tls "github.com/Danny-Dasilva/utls"
	"github.com/andybalholm/brotli"
	"github.com/wangluozhe/requests/ja3"
	"github.com/wangluozhe/requests/models"
	"github.com/wangluozhe/requests/url"
	"io/ioutil"
	url2 "net/url"
	"strings"
	"time"
)

// 默认User—Agent
func default_user_agent() string {
	return USER_AGENT
}

// 默认请求头
func default_headers() *http.Header {
	headers := url.NewHeaders()
	headers.Set("User-Agent", default_user_agent())
	headers.Set("Accept-Encoding", "gzip, deflate, br")
	headers.Set("Accept", "*/*")
	headers.Set("Connection", "keep-alive")
	return headers
}

// 合并cookies
func merge_cookies(rawurl string, cookieJar *cookiejar.Jar, cookie *cookiejar.Jar) {
	urls, _ := url2.Parse(rawurl)
	cookieJar.SetCookies(urls, cookie.Cookies(urls))
}

// 合并参数
func merge_setting(request_setting, session_setting interface{}) interface{} {
	switch (request_setting).(type) {
	case *url.Params:
		merged_setting := session_setting.(*url.Params)
		if merged_setting == nil {
			return request_setting
		}
		requestd_setting := request_setting.(*url.Params)
		if requestd_setting == nil {
			return merged_setting
		}
		for _, key := range requestd_setting.Keys() {
			merged_setting.Set(key, requestd_setting.Get(key))
		}
		return merged_setting
	case *http.Header:
		merged_setting := session_setting.(*http.Header)
		if merged_setting == nil {
			return request_setting
		}
		requestd_setting := request_setting.(*http.Header)
		if requestd_setting == nil {
			return merged_setting
		}
		for key, _ := range *requestd_setting {
			merged_setting.Set(key,(*requestd_setting)[key][0])
		}
		return merged_setting
	case []string:
		merged_setting := session_setting.([]string)
		if merged_setting == nil {
			return request_setting
		}
		requestd_setting := request_setting.([]string)
		if requestd_setting == nil {
			return merged_setting
		}
		for index, value := range requestd_setting {
			merged_setting[index] = value
		}
		return merged_setting
	case bool:
		merged_setting := session_setting.(bool)
		requestd_setting := request_setting.(bool)
		if requestd_setting == true {
			merged_setting = requestd_setting
		}
		return merged_setting
	case string:
		merged_setting := session_setting.(string)
		if merged_setting == "" {
			return request_setting
		}
		requestd_setting := request_setting.(string)
		if requestd_setting == "" {
			return merged_setting
		}
	}
	return request_setting
}

// 禁用redirect
var disableRedirect = func(request *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

const (
	DEFAULT_REDIRECT_LIMIT = 30 // 默认redirect最大次数
	DEFAULT_TIMEOUT        = 10 // 默认client响应时间
)

// 新建默认Session
func NewSession() *Session {
	session := &Session{
		Headers:      default_headers(),
		Cookies:      nil,
		Verify:       true,
		MaxRedirects: DEFAULT_REDIRECT_LIMIT,
		transport:    nil,
		request:      nil,
		client:       nil,
	}
	cookies, _ := cookiejar.New(nil)
	session.Cookies = cookies
	session.transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: session.Verify,
		},
		DisableKeepAlives: true,
	}
	session.request = &http.Request{}
	session.client = &http.Client{
		Transport:     session.transport,
		CheckRedirect: nil,
		Jar:           cookies,
		Timeout:       DEFAULT_TIMEOUT * time.Second,
	}
	return session
}

// 新建默认Session，同上一模一样
func DefaultSession() *Session {
	return NewSession()
}

// Session结构体
type Session struct {
	Params       *url.Params
	Headers      *http.Header
	Cookies      *cookiejar.Jar
	Auth         []string
	Proxies      string
	Verify       bool
	Cert         []string
	Ja3          string
	MaxRedirects int
	transport    *http.Transport
	request      *http.Request
	client       *http.Client
}

// 预请求处理
func (s *Session) Prepare_request(request *models.Request) (*models.PrepareRequest, error) {
	var err error
	params := merge_setting(request.Params, s.Params).(*url.Params)
	headers := merge_setting(request.Headers, s.Headers).(*http.Header)
	c := request.Cookies
	if c == nil {
		c, _ = cookiejar.New(nil)
	}
	cookies, _ := cookiejar.New(nil)
	merge_cookies(request.Url, cookies, s.Cookies)
	merge_cookies(request.Url, cookies, c)
	auth := merge_setting(request.Auth, s.Auth).([]string)
	p := models.NewPrepareRequest()
	err = p.Prepare(
		request.Method,
		request.Url,
		params,
		headers,
		cookies,
		request.Data,
		request.Files,
		request.Json,
		auth,
	)
	if err != nil {
		return p, err
	}
	return p, nil
}

// http请求方式基础函数
func (s *Session) Request(method, rawurl string, request *url.Request) (*models.Response, error) {
	if request == nil {
		request = url.NewRequest()
	}
	req := &models.Request{
		Method:  strings.ToUpper(method),
		Url:     rawurl,
		Params:  request.Params,
		Headers: request.Headers,
		Cookies: request.Cookies,
		Data:    request.Data,
		Files:   request.Files,
		Json:    request.Json,
		Auth:    request.Auth,
	}
	preq, err := s.Prepare_request(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.Send(preq, request)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// get请求方式
func (s *Session) Get(rawurl string, req *url.Request) (*models.Response, error) {
	return s.Request(http.MethodGet, rawurl, req)
}

// post请求方式
func (s *Session) Post(rawurl string, req *url.Request) (*models.Response, error) {
	return s.Request(http.MethodPost, rawurl, req)
}

// options请求方式
func (s *Session) Options(rawurl string, req *url.Request) (*models.Response, error) {
	return s.Request(http.MethodOptions, rawurl, req)
}

// head请求方式
func (s *Session) Head(rawurl string, req *url.Request) (*models.Response, error) {
	return s.Request(http.MethodHead, rawurl, req)
}

// put请求方式
func (s *Session) Put(rawurl string, req *url.Request) (*models.Response, error) {
	return s.Request(http.MethodPut, rawurl, req)
}

// patch请求方式
func (s *Session) Patch(rawurl string, req *url.Request) (*models.Response, error) {
	return s.Request(http.MethodPatch, rawurl, req)
}

// delete请求方式
func (s *Session) Delete(rawurl string, req *url.Request) (*models.Response, error) {
	return s.Request(http.MethodDelete, rawurl, req)
}

// connect请求方式
func (s *Session) Connect(rawurl string, req *url.Request) (*models.Response, error) {
	return s.Request(http.MethodConnect, rawurl, req)
}

// trace请求方式
func (s *Session) Trace(rawurl string, req *url.Request) (*models.Response, error) {
	return s.Request(http.MethodTrace, rawurl, req)
}

// 发送数据
func (s *Session) Send(preq *models.PrepareRequest, req *url.Request) (*models.Response, error) {
	var err error
	var history []*models.Response

	// 修复报错tls: CurvePreferences includes unsupported curve
	s.transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: s.Verify,
	}

	// 设置代理
	proxies := merge_setting(s.Proxies, req.Proxies).(string)
	if proxies != "" {
		u1, err := url2.Parse(proxies)
		if err != nil {
			return nil, err
		}
		s.transport.Proxy = http.ProxyURL(u1)
	}

	// 是否验证证书
	verify := merge_setting(s.Verify, req.Verify).(bool)
	s.transport.TLSClientConfig.InsecureSkipVerify = verify

	// 设置证书
	cert := merge_setting(s.Cert, req.Cert).([]string)
	if cert != nil {
		var cert_byte []byte
		certs, err := tls.LoadX509KeyPair(cert[0], cert[1])
		if err != nil {
			return nil, err
		}
		if len(cert) == 3 {
			cert_byte, err = ioutil.ReadFile(cert[2])
		} else {
			cert_byte, err = ioutil.ReadFile(cert[0])
		}
		if err != nil {
			return nil, err
		}
		certPool := x509.NewCertPool()
		ok := certPool.AppendCertsFromPEM(cert_byte)
		if !ok {
			return nil, errors.New("failed to parse root certificate")
		}
		s.transport.TLSClientConfig.RootCAs = certPool
		fmt.Println(certs)
		s.transport.TLSClientConfig.Certificates = []tls.Certificate{certs}
	}

	// 设置JA3指纹信息
	ja3String := merge_setting(s.Ja3, req.Ja3).(string)
	if ja3String != "" && strings.HasPrefix(preq.Url, "https") {
		browser := ja3.Browser{
			JA3:       ja3String,
			UserAgent: s.Headers.Get("User-Agent"),
		}
		tr, err := ja3.NewJA3Transport(browser, proxies, s.transport.TLSClientConfig)
		if err != nil {
			return nil, err
		}
		s.client.Transport = tr
	}

	// 设置超时时间
	timeout := req.Timeout
	if timeout != 0 {
		s.client.Timeout = timeout
	}

	// 是否自动转发
	allowRedirect := req.AllowRedirects
	if allowRedirect {
		s.client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
			if request != nil {
				preq.Url = request.URL.String()
				p := models.NewPrepareRequest()
				c, _ := cookiejar.New(nil)
				c.SetCookies(request.URL, request.Cookies())
				p.Prepare(request.Method, request.URL.String(), nil, &request.Header, c, nil, nil, nil, nil)
				r := s.buildResponse(request.Response, p, &url.Request{})
				history = append(history, r)
			}
			if len(via) > s.MaxRedirects {
				return errors.New(fmt.Sprintf("redirects number gt %i", s.MaxRedirects))
			}
			return nil
		}
	} else {
		s.client.CheckRedirect = disableRedirect
	}

	u, _ := url2.Parse(preq.Url)
	// 设置有序请求头
	if req.Headers != nil{
		if (*req.Headers)[http.HeaderOrderKey] != nil {
			(*preq.Headers)[http.HeaderOrderKey] = (*req.Headers)[http.HeaderOrderKey]
		}
	}
	s.request = &http.Request{
		Method: preq.Method,
		URL:    u,
		Header: *preq.Headers,
		Body:   preq.Body,
	}
	s.client.Jar = preq.Cookies
	req.Headers = &s.request.Header
	resp, err := s.client.Do(s.request)
	if err != nil {
		return nil, err
	}
	response := s.buildResponse(resp, preq, req)
	response.History = history
	return response, nil
}

// 构建response参数
func (s *Session) buildResponse(resp *http.Response, preq *models.PrepareRequest, req *url.Request) *models.Response {
	content, _ := ioutil.ReadAll(resp.Body)
	encoding := resp.Header.Get("Content-Encoding")
	DecompressBody(&content, encoding)
	body := ioutil.NopCloser(bytes.NewReader(content))
	response := &models.Response{
		Url:        preq.Url,
		Headers:    resp.Header,
		Cookies:    resp.Cookies(),
		Text:       string(content),
		Content:    content,
		Body:       body,
		StatusCode: resp.StatusCode,
		History:    []*models.Response{},
		Request:    req,
	}
	if resp.Cookies() != nil {
		u, _ := url2.Parse(preq.Url)
		s.Cookies.SetCookies(u, resp.Cookies())
	}
	return response
}

// 解码Body数据
func DecompressBody(content *[]byte, encoding string) {
	if encoding != "" {
		if strings.ToLower(encoding) == "gzip" {
			decodeGZip(content)
		} else if strings.ToLower(encoding) == "deflate" {
			decodeDeflate(content)
		} else if strings.ToLower(encoding) == "br" {
			decodeBrotli(content)
		}
	}
}

// 解码GZip编码
func decodeGZip(content *[]byte) error {
	if content == nil {
		return nil
	}
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, content)
	r, err := gzip.NewReader(b)
	if err != nil {
		return err
	}
	defer r.Close()
	*content, err = ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return nil
}

// 解码deflate编码
func decodeDeflate(content *[]byte) error {
	var err error
	if content == nil {
		return err
	}
	r := flate.NewReader(bytes.NewReader(*content))
	defer r.Close()
	*content, err = ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return nil
}

// 解码br编码
func decodeBrotli(content *[]byte) error {
	var err error
	if content == nil {
		return err
	}
	r := brotli.NewReader(bytes.NewReader(*content))
	*content, err = ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return nil
}

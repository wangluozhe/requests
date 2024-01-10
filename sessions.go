package requests

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/andybalholm/brotli"
	utls "github.com/refraction-networking/utls"
	"github.com/wangluozhe/chttp"
	"github.com/wangluozhe/chttp/cookiejar"
	"github.com/wangluozhe/chttp/http2"
	"github.com/wangluozhe/requests/models"
	ja3 "github.com/wangluozhe/requests/transport"
	"github.com/wangluozhe/requests/url"
	"github.com/wangluozhe/requests/utils"
	"io/ioutil"
	"log"
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
	urls, _ := url2.Parse(utils.EncodeURI(rawurl))
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
			if key == http.PHeaderOrderKey || key == http.HeaderOrderKey || key == http.UnChangedHeaderKey {
				continue
			}
			merged_setting.Set(key, (*requestd_setting)[key][0])
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
	case *ja3.TLSExtensions:
		merged_setting := session_setting.(*ja3.TLSExtensions)
		if merged_setting == nil {
			return request_setting
		}
		requestd_setting := request_setting.(*ja3.TLSExtensions)
		if requestd_setting == nil {
			return merged_setting
		}
	case *http2.HTTP2Settings:
		merged_setting := session_setting.(*http2.HTTP2Settings)
		if merged_setting == nil {
			return request_setting
		}
		requestd_setting := request_setting.(*http2.HTTP2Settings)
		if requestd_setting == nil {
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
		TLSClientConfig: &utls.Config{
			InsecureSkipVerify: session.Verify,
			OmitEmptyPsk:       true,
		},
		DisableKeepAlives: false, // 这里问题很严重
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
	Params        *url.Params
	Headers       *http.Header
	Cookies       *cookiejar.Jar
	Auth          []string
	Proxies       string
	Verify        bool
	Cert          []string
	Ja3           string
	MaxRedirects  int
	TLSExtensions *ja3.TLSExtensions
	HTTP2Settings *http2.HTTP2Settings
	transport     *http.Transport
	request       *http.Request
	client        *http.Client
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
		request.Body,
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
		Body:    request.Body,
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

	// 设置代理
	proxies := merge_setting(s.Proxies, req.Proxies).(string)
	if proxies != "" {
		u1, err := url2.Parse(proxies)
		if err != nil {
			return nil, err
		}
		s.transport.Proxy = http.ProxyURL(u1)
	} else {
		s.transport.Proxy = nil
	}

	// 设置JA3指纹信息
	ja3String := merge_setting(s.Ja3, req.Ja3).(string)
	if ja3String != "" && strings.HasPrefix(preq.Url, "https") {
		browser := ja3.Browser{
			JA3:       ja3String,
			UserAgent: s.Headers.Get("User-Agent"),
		}

		// 自定义TLS指纹信息
		tlsExtensions := merge_setting(req.TLSExtensions, s.TLSExtensions).(*ja3.TLSExtensions)
		http2Settings := merge_setting(req.HTTP2Settings, s.HTTP2Settings).(*http2.HTTP2Settings)
		if strings.Index(strings.Split(browser.JA3, ",")[2], "-41") != -1 {
			config := s.transport.TLSClientConfig.Clone()
			if config.ClientSessionCache == nil {
				config.SessionTicketKey = [32]byte{}
				config.OmitEmptyPsk = true
				config.ClientSessionCache = utls.NewLRUClientSessionCache(0)
				s.transport.TLSClientConfig = config
			}
		}

		options := &ja3.Options{
			Browser:       browser,
			TLSExtensions: tlsExtensions,
			HTTP2Settings: http2Settings,
			ForceHTTP1:    req.ForceHTTP1,
			TLSConfig:     s.transport.TLSClientConfig,
		}

		if proxies != "" {
			options.Proxy = proxies
		}

		client, err := ja3.NewClient(options)
		if err != nil {
			return nil, err
		}
		s.client = &client
	}

	// 是否验证证书
	verify := merge_setting(s.Verify, req.Verify).(bool)
	s.transport.TLSClientConfig.InsecureSkipVerify = verify

	// 设置证书
	cert := merge_setting(s.Cert, req.Cert).([]string)
	if cert != nil {
		var cert_byte []byte
		certs, err := utls.LoadX509KeyPair(cert[0], cert[1])
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
		s.transport.TLSClientConfig.Certificates = []utls.Certificate{certs}
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
			if len(via) > s.MaxRedirects {
				return errors.New(fmt.Sprintf("redirects number gt %i", s.MaxRedirects))
			}
			if request != nil {
				preq.Url = request.URL.String()
				p := models.NewPrepareRequest()
				c, _ := cookiejar.New(nil)
				c.SetCookies(request.URL, request.Cookies())
				p.Prepare(request.Method, request.URL.String(), nil, &request.Header, c, nil, nil, nil, "", nil)
				r, err := s.buildResponse(request.Response, p, &url.Request{})
				if err != nil {
					return err
				}
				history = append(history, r)
			}
			// 获取上一次请求的Cookies
			var cookies []*http.Cookie
			lastReq := via[len(via)-1]
			if lastReq.Response != nil {
				cookies = lastReq.Response.Cookies()
			}

			// 将上一次请求的Cookies与当前请求的Cookies合并
			reqCookies := append(cookies, request.Cookies()...)

			// 设置合并后的Cookies到重定向请求的Header中
			for _, cookie := range reqCookies {
				request.AddCookie(cookie)
			}
			return nil
		}
	} else {
		s.client.CheckRedirect = disableRedirect
	}

	// 设置有序请求头
	if req.Headers != nil {
		if (*req.Headers)[http.HeaderOrderKey] != nil {
			(*preq.Headers)[http.HeaderOrderKey] = (*req.Headers)[http.HeaderOrderKey]
		}
		if (*req.Headers)[http.PHeaderOrderKey] != nil {
			(*preq.Headers)[http.PHeaderOrderKey] = (*req.Headers)[http.PHeaderOrderKey]
		}
		if (*req.Headers)[http.UnChangedHeaderKey] != nil {
			(*preq.Headers)[http.UnChangedHeaderKey] = (*req.Headers)[http.UnChangedHeaderKey]
		}
	}

	s.request, err = http.NewRequest(preq.Method, preq.Url, preq.Body)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	s.request.Header = *preq.Headers
	s.client.Jar = preq.Cookies
	req.Headers = &s.request.Header
	resp, err := s.client.Do(s.request)
	if err != nil {
		return nil, err
	}
	response, err := s.buildResponse(resp, preq, req)
	if err != nil {
		return nil, err
	}
	response.History = history
	return response, nil
}

// 构建response参数
func (s *Session) buildResponse(resp *http.Response, preq *models.PrepareRequest, req *url.Request) (*models.Response, error) {
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
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
	return response, nil
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

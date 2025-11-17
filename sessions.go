package requests

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net"
	url2 "net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	utls "github.com/refraction-networking/utls"
	"github.com/wangluozhe/chttp"
	"github.com/wangluozhe/chttp/cookiejar"
	"github.com/wangluozhe/chttp/httputil"
	"github.com/wangluozhe/requests/models"
	"github.com/wangluozhe/requests/url"
	"github.com/wangluozhe/requests/utils"
)

// 默认User—Agent
func default_user_agent() string {
	return USER_AGENT
}

// 默认请求头
func default_headers() *http.Header {
	headers := url.NewHeaders()
	headers.Set("User-Agent", default_user_agent())
	return headers
}

// 合并cookies
func merge_cookies(rawurl string, cookieJar *cookiejar.Jar, cookie *cookiejar.Jar) {
	urls, _ := url2.Parse(utils.EncodeURI(rawurl))

	// 防止因为传入 nil 对象导致程序崩溃
	if cookieJar == nil || cookie == nil {
		return
	}

	// cookieJar的name表
	var cookieJarNames = make(map[string]struct{})
	for _, c := range cookieJar.Cookies(urls) {
		cookieJarNames[c.Name] = struct{}{}
	}

	var newCookies []*http.Cookie

	// 检查目标 cookieJar 中是否已存在相同name，存在则不更新
	for _, c := range cookie.Cookies(urls) {
		if _, exists := cookieJarNames[c.Name]; !exists {
			newCookies = append(newCookies, c)
		}
	}

	// 如果列表不为空才执行设置操作
	if len(newCookies) > 0 {
		cookieJar.SetCookies(urls, newCookies)
	}
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
	case *http.TLSExtensions:
		merged_setting := session_setting.(*http.TLSExtensions)
		if merged_setting == nil {
			return request_setting
		}
		requestd_setting := request_setting.(*http.TLSExtensions)
		if requestd_setting == nil {
			return merged_setting
		}
	case *http.HTTP2Settings:
		merged_setting := session_setting.(*http.HTTP2Settings)
		if merged_setting == nil {
			return request_setting
		}
		requestd_setting := request_setting.(*http.HTTP2Settings)
		if requestd_setting == nil {
			return merged_setting
		}
	case nil:
		return session_setting
	}
	return request_setting
}

// 禁用redirect
var disableRedirect = func(request *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

const (
	DEFAULT_REDIRECT_LIMIT = 30 // 默认redirect最大次数
	DEFAULT_TIMEOUT        = 30 // 默认client响应时间
)

// 新建默认Session
func NewSession() *Session {
	session := &Session{
		Headers:        default_headers(),
		Cookies:        nil,
		Verify:         true,
		MaxRedirects:   DEFAULT_REDIRECT_LIMIT,
		transport:      nil,
		transportCache: make(map[string]*http.Transport),
		cacheLock:      sync.Mutex{},
	}
	cookies, _ := cookiejar.New(nil)
	session.Cookies = cookies
	session.transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   DEFAULT_TIMEOUT * time.Second,
			KeepAlive: DEFAULT_TIMEOUT * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   DEFAULT_TIMEOUT * time.Second,
		ResponseHeaderTimeout: DEFAULT_TIMEOUT * time.Second,
		TLSClientConfig: &utls.Config{
			InsecureSkipVerify:                 session.Verify,
			ClientSessionCache:                 utls.NewLRUClientSessionCache(0),
			OmitEmptyPsk:                       true,
			PreferSkipResumptionOnNilExtension: true,
			SessionTicketsDisabled:             false,
		},
		DisableKeepAlives:   false,
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 200,
		MaxConnsPerHost:     200,
		IdleConnTimeout:     time.Duration(DEFAULT_TIMEOUT) * time.Second,
	}
	return session
}

var defaultSession = NewSession()

// Session结构体
type Session struct {
	Params         *url.Params
	Headers        *http.Header
	Cookies        *cookiejar.Jar
	Auth           []string
	Proxies        string
	Verify         bool
	Cert           []string
	Ja3            string
	RandomJA3      bool
	ForceHTTP1     bool
	TLSExtensions  *http.TLSExtensions
	HTTP2Settings  *http.HTTP2Settings
	MaxRedirects   int
	transport      *http.Transport
	transportCache map[string]*http.Transport
	cacheLock      sync.Mutex
}

// 预请求处理
func (s *Session) Prepare_request(request *models.Request) (*models.PrepareRequest, error) {
	var err error
	params := merge_setting(request.Params, s.Params).(*url.Params)
	var h http.Header
	if request.Headers != nil {
		h = request.Headers.Clone()
	}
	s_h := s.Headers.Clone()
	headers := merge_setting(&h, &s_h).(*http.Header)
	c := request.Cookies
	if c == nil {
		c, _ = cookiejar.New(nil)
	}
	cookies, _ := cookiejar.New(nil)
	merge_cookies(request.Url, cookies, c)
	merge_cookies(request.Url, cookies, s.Cookies)
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
	proxies := merge_setting(req.Proxies, s.Proxies).(string)
	// 是否验证证书
	verify := merge_setting(req.Verify, s.Verify).(bool)
	// 设置证书
	cert := merge_setting(req.Cert, s.Cert).([]string)
	// 设置超时时间
	timeout := merge_setting(req.Timeout, DEFAULT_TIMEOUT).(time.Duration)
	// 设置ja3
	ja3String := merge_setting(req.Ja3, s.Ja3).(string)
	// 设置随机ja3
	randomJA3 := merge_setting(req.RandomJA3, s.RandomJA3).(bool)
	// 设置强制http1
	forceHTTP1 := merge_setting(req.ForceHTTP1, s.ForceHTTP1).(bool)
	// 设置tls
	tlsExtensions := merge_setting(req.TLSExtensions, s.TLSExtensions).(*http.TLSExtensions)
	// 设置http2
	http2Settings := merge_setting(req.HTTP2Settings, s.HTTP2Settings).(*http.HTTP2Settings)
	if http2Settings != nil && http2Settings.HeadersID == 0 {
		http2Settings.HeadersID = 1
	}

	// --- 缓存逻辑开始 ---

	// 1. 根据 Transport 级别的配置生成缓存 Key
	// 注意：只有会影响 Transport 的配置才需要加入 Key
	certKey := ""
	if len(cert) > 0 {
		certKey = strings.Join(cert, "|")
	}

	// 不使用连接复用
	disableKeepAlives := strings.ToLower(preq.Headers.Get("Connection")) == "close"

	cacheKey := fmt.Sprintf("proxies=%s&verify=%t&cert=%s&ja3=%s&randomJA3=%t&forceHTTP1=%t&tlsExtensions=%s&http2Settings=%s&DisableKeepAlives=%t",
		proxies, verify, certKey, ja3String, randomJA3, forceHTTP1, tlsExtensionsHash(tlsExtensions), http2SettingsHash(http2Settings), disableKeepAlives,
	)

	s.cacheLock.Lock() // 加锁保护缓存

	transport, found := s.transportCache[cacheKey]
	if !found {
		// 缓存未命中，创建新的 Transport
		transport = s.transport.Clone() // 从会话的基础 transport 克隆
		transport.TLSClientConfig = s.transport.TLSClientConfig.Clone()

		if proxies != "" {
			u1, err := url2.Parse(proxies)
			if err != nil {
				return nil, err
			}
			transport.Proxy = http.ProxyURL(u1)
		}

		if verify != transport.TLSClientConfig.InsecureSkipVerify {
			transport.TLSClientConfig.InsecureSkipVerify = verify
		}

		if cert != nil {
			var cert_byte []byte
			certs, err := utls.LoadX509KeyPair(cert[0], cert[1])
			if err != nil {
				return nil, err
			}
			if len(cert) == 3 {
				cert_byte, err = os.ReadFile(cert[2])
			} else {
				cert_byte, err = os.ReadFile(cert[0])
			}
			if err != nil {
				return nil, err
			}
			certPool := x509.NewCertPool()
			ok := certPool.AppendCertsFromPEM(cert_byte)
			if !ok {
				return nil, errors.New("failed to parse root certificate")
			}
			transport.TLSClientConfig.RootCAs = certPool
			transport.TLSClientConfig.Certificates = []utls.Certificate{certs}
		}

		if ja3String != "" && strings.HasPrefix(preq.Url, "https") && transport.H2Transport == nil {
			if transport.TLSClientConfig.ClientSessionCache == nil {
				transport.TLSClientConfig.ClientSessionCache = utls.NewLRUClientSessionCache(0)
			}
			if transport.TLSClientConfig.OmitEmptyPsk == false {
				transport.TLSClientConfig.OmitEmptyPsk = true
			}
			if strings.Index(strings.Split(ja3String, ",")[2], "-41") != -1 {
				transport.TLSClientConfig.SessionTicketsDisabled = false
			}
			if transport.H2Transport == nil {
				var h2 *http.HTTP2Transport
				h2, _ = http.HTTP2ConfigureTransports(transport)
				transport.JA3 = ja3String
				transport.UserAgent = s.Headers.Get("User-Agent")
				transport.RandomJA3 = randomJA3
				transport.ForceHTTP1 = forceHTTP1
				// 自定义TLS指纹信息
				h2.HTTP2Settings = http2Settings
				if http2Settings != nil {
					if http2Settings.Settings != nil {
						for _, setting := range http2Settings.Settings {
							switch setting.ID {
							case http.HTTP2SettingHeaderTableSize:
								h2.MaxEncoderHeaderTableSize = setting.Val
								h2.MaxDecoderHeaderTableSize = setting.Val
							case http.HTTP2SettingMaxConcurrentStreams:
								h2.StrictMaxConcurrentStreams = true
							case http.HTTP2SettingMaxFrameSize:
								h2.MaxReadFrameSize = setting.Val
							case http.HTTP2SettingMaxHeaderListSize:
								h2.MaxHeaderListSize = setting.Val
							}
						}
					}
				}
				transport.TLSExtensions = tlsExtensions
				transport.H2Transport = h2
			}
		} else if ja3String != "" && strings.HasPrefix(preq.Url, "https") && transport.H2Transport != nil {
			transport.JA3 = ja3String
			transport.UserAgent = s.Headers.Get("User-Agent")
			transport.RandomJA3 = randomJA3
			transport.ForceHTTP1 = forceHTTP1
			transport.TLSExtensions = tlsExtensions
			transport.H2Transport.(*http.HTTP2Transport).HTTP2Settings = http2Settings
		}

		if disableKeepAlives {
			transport.DisableKeepAlives = true
		}

		s.transportCache[cacheKey] = transport // 将新创建的 transport 存入缓存
	}

	s.cacheLock.Unlock() // 解锁

	// --- 缓存逻辑结束 ---

	// 2. 使用获取到的 Transport 创建一个 Client
	// Client 的创建开销很小，每次请求都可以创建一个新的，以便设置请求特定的 Timeout
	client := &http.Client{
		Transport:     transport,
		CheckRedirect: nil,
		Timeout:       timeout,
	}

	// 是否自动转发
	if req.AllowRedirects {
		if client.CheckRedirect == nil {
			client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
				if len(via) > s.MaxRedirects {
					return errors.New(fmt.Sprintf("redirects number gt %i", s.MaxRedirects))
				}
				if request != nil {
					preq.Url = request.URL.String()
					p := models.NewPrepareRequest()
					c, _ := cookiejar.New(nil)
					c.SetCookies(request.URL, request.Cookies())
					p.Prepare(request.Method, request.URL.String(), nil, &request.Header, c, nil, nil, nil, nil, nil)
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
		}
	} else {
		client.CheckRedirect = disableRedirect
	}

	// 设置cookies
	client.Jar = preq.Cookies

	// 设置有序请求头
	if req.Headers != nil {
		var preqHeaders http.Header
		for _, key := range []string{http.HeaderOrderKey, http.PHeaderOrderKey, http.UnChangedHeaderKey} {
			reqValue := (*req.Headers)[key]
			preqValue := (*preq.Headers)[key]
			if reqValue == nil {
				continue
			}
			if !reflect.DeepEqual(reqValue, preqValue) {
				if preqHeaders == nil {
					preqHeaders = preq.Headers.Clone()
				}
				preqHeaders[key] = reqValue
			}
		}
		if preqHeaders != nil {
			preq.Headers = &preqHeaders
		}
	}

	request, err := http.NewRequest(preq.Method, preq.Url, preq.Body)
	if err != nil {
		return nil, err
	}
	request.Header = preq.Headers.Clone()
	resp, err := client.Do(request)
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
	rawResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}

	if req.Stream {
		response := &models.Response{
			Url:         preq.Url,
			Headers:     resp.Header,
			Cookies:     resp.Cookies(),
			Text:        "",
			Content:     nil,
			Body:        resp.Body,
			StatusCode:  resp.StatusCode,
			History:     []*models.Response{},
			Request:     req,
			RawResponse: rawResponse,
		}
		if resp.Cookies() != nil {
			u, _ := url2.Parse(preq.Url)
			s.Cookies.SetCookies(u, resp.Cookies())
		}
		return response, nil
	}

	response := &models.Response{
		Url:         preq.Url,
		Headers:     resp.Header,
		Cookies:     resp.Cookies(),
		Text:        "",
		Content:     nil,
		Body:        nil,
		StatusCode:  resp.StatusCode,
		History:     []*models.Response{},
		Request:     req,
		RawResponse: rawResponse,
	}

	// 创建解压流
	decompressed_body, err := DecompressBody(resp.Body, resp.Header.Get("Content-Encoding"))
	if err != nil {
		return nil, err
	}
	defer decompressed_body.Close()

	// 流式读取到缓冲区
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, decompressed_body); err != nil {
		return nil, err
	}

	response.Content = buf.Bytes()
	response.Text = buf.String()
	response.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))

	if resp.Cookies() != nil {
		u, _ := url2.Parse(preq.Url)
		s.Cookies.SetCookies(u, resp.Cookies())
	}
	resp.Body.Close()
	return response, nil
}

// TLSExtensions的hash
func tlsExtensionsHash(tlsExtensions *http.TLSExtensions) string {
	bytes, err := json.Marshal(tlsExtensions)
	if err != nil {
		return ""
	}

	h := fnv.New64a()
	h.Write(bytes)
	return strconv.Itoa(int(h.Sum64()))
}

// HTTP2Settings的hash
func http2SettingsHash(http2Settings *http.HTTP2Settings) string {
	bytes, err := json.Marshal(http2Settings)
	if err != nil {
		return ""
	}

	h := fnv.New64a()
	h.Write(bytes)
	return strconv.Itoa(int(h.Sum64()))
}

// 解码Body数据
func DecompressBody(body io.ReadCloser, encoding string) (io.ReadCloser, error) {
	switch strings.ToLower(encoding) {
	case "gzip":
		r, err := gzip.NewReader(body)
		if err != nil {
			return nil, err
		}
		return r, nil // gzip.Reader 自动处理流式解压[7](@ref)

	case "deflate":
		return flate.NewReader(body), nil // flate.Reader 直接返回流式接口[7,8](@ref)

	case "br":
		return io.NopCloser(brotli.NewReader(body)), nil // brotli.Reader 实现 io.Reader 接口[1](@ref)

	default:
		return body, nil // 非压缩数据原样返回
	}
}

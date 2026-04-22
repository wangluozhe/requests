package requests

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/x509"
	"encoding/json"
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
	"github.com/google/uuid"
	utls "github.com/refraction-networking/utls"
	"github.com/wangluozhe/chttp"
	"github.com/wangluozhe/chttp/cookiejar"
	"github.com/wangluozhe/chttp/httputil"
	"github.com/wangluozhe/requests/models"
	"github.com/wangluozhe/requests/url"
	"github.com/wangluozhe/requests/utils"
)

// Handler 定义处理请求的函数类型
type Handler func(preq *models.PrepareRequest, req *url.Request) (*models.Response, error)

// Middleware 定义中间件函数类型
type Middleware func(next Handler) Handler

// Bool 返回指向给定 bool 值的指针，用于将 Session 的 bool 字段包装为 *bool 传给 merge_setting。
func Bool(v bool) *bool {
	return &v
}

// 默认User-Agent
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

func cloneParams(params *url.Params) *url.Params {
	if params == nil {
		return nil
	}
	cloned := url.NewParams()
	for key, values := range params.Values() {
		for _, value := range values {
			cloned.Add(key, value)
		}
	}
	return cloned
}

// 合并参数
func merge_setting(request_setting, session_setting interface{}) interface{} {
	switch (request_setting).(type) {
	case *url.Params:
		sessionParams := session_setting.(*url.Params)
		if sessionParams == nil {
			return request_setting
		}
		requestParams := request_setting.(*url.Params)
		if requestParams == nil {
			return cloneParams(sessionParams)
		}
		merged := cloneParams(sessionParams)
		for key, values := range requestParams.Values() {
			merged.Del(key)
			for _, value := range values {
				merged.Add(key, value)
			}
		}
		return merged
	case *http.Header:
		sessionHeaders := session_setting.(*http.Header)
		if sessionHeaders == nil {
			return request_setting
		}
		requestHeaders := request_setting.(*http.Header)
		if requestHeaders == nil {
			cloned := sessionHeaders.Clone()
			return &cloned
		}
		merged := sessionHeaders.Clone()
		for key := range *requestHeaders {
			if key == http.PHeaderOrderKey || key == http.HeaderOrderKey || key == http.UnChangedHeaderKey {
				continue
			}
			values := append([]string(nil), (*requestHeaders)[key]...)
			merged[key] = values
		}
		return &merged
	case []string:
		sessionSlice := session_setting.([]string)
		if sessionSlice == nil {
			return request_setting
		}
		requestSlice := request_setting.([]string)
		if requestSlice == nil {
			return append([]string(nil), sessionSlice...)
		}
		return append([]string(nil), requestSlice...)
	case *bool:
		// *bool 类型：nil 表示"未设置"，非 nil 则使用 Request 的显式值
		requestd_setting := request_setting.(*bool)
		if requestd_setting != nil {
			return requestd_setting
		}
		return session_setting
	case string:
		merged_setting := session_setting.(string)
		if merged_setting == "" {
			return request_setting
		}
		requestd_setting := request_setting.(string)
		if requestd_setting == "" {
			return merged_setting
		}
		return requestd_setting
	case *http.TLSExtensions:
		merged_setting := session_setting.(*http.TLSExtensions)
		if merged_setting == nil {
			return request_setting
		}
		requestd_setting := request_setting.(*http.TLSExtensions)
		if requestd_setting == nil {
			return merged_setting
		}
		return requestd_setting
	case *http.HTTP2Settings:
		merged_setting := session_setting.(*http.HTTP2Settings)
		if merged_setting == nil {
			return request_setting
		}
		requestd_setting := request_setting.(*http.HTTP2Settings)
		if requestd_setting == nil {
			return merged_setting
		}
		return requestd_setting
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
		Timeout:        time.Duration(DEFAULT_TIMEOUT) * time.Second,
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
			InsecureSkipVerify:                 !session.Verify,
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
	Timeout        time.Duration
	Proxies        string
	Verify         bool
	Cert           []string
	Ja3            string
	RandomJA3      bool
	ForceHTTP1     bool
	TLSExtensions  *http.TLSExtensions
	HTTP2Settings  *http.HTTP2Settings
	MaxRedirects   int
	Middlewares    []Middleware // 中间件列表
	transport      *http.Transport
	transportCache map[string]*http.Transport
	cacheLock      sync.Mutex
}

// Use 添加中间件
func (s *Session) Use(middleware ...Middleware) {
	s.Middlewares = append(s.Middlewares, middleware...)
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
		Params:  request.GetParams(),
		Headers: request.GetHeaders(),
		Cookies: request.GetCookies(rawurl),
		Data:    request.GetData(),
		Files:   request.GetFiles(),
		Json:    request.Json,
		Body:    request.GetBody(),
		Auth:    request.Auth,
	}
	preq, err := s.Prepare_request(req)
	if err != nil {
		return nil, err
	}

	// 构造中间件调用链，s.Send 作为最底层的 Handler
	var handler Handler = s.Send
	// 倒序应用中间件，保证 Use 的顺序即为执行顺序
	for i := len(s.Middlewares) - 1; i >= 0; i-- {
		handler = s.Middlewares[i](handler)
	}

	// 执行 Handler (中间件链 -> Send)
	resp, err := handler(preq, request)
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
	verify := *merge_setting(req.Verify, Bool(s.Verify)).(*bool)
	// 设置证书
	cert := merge_setting(req.Cert, s.Cert).([]string)
	// 设置超时时间
	timeout := merge_setting(req.Timeout, s.Timeout).(time.Duration)
	// 设置ja3
	ja3String := merge_setting(req.Ja3, s.Ja3).(string)
	// 设置随机ja3
	randomJA3 := *merge_setting(req.RandomJA3, Bool(s.RandomJA3)).(*bool)
	// 设置强制http1
	forceHTTP1 := *merge_setting(req.ForceHTTP1, Bool(s.ForceHTTP1)).(*bool)
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

	transport, err := func() (*http.Transport, error) {
		s.cacheLock.Lock() // 加锁保护缓存
		defer s.cacheLock.Unlock()

		tr, found := s.transportCache[cacheKey]
		if found {
			return tr, nil
		}

		// 缓存未命中，创建新的 Transport
		tr = s.transport.Clone() // 从会话的基础 transport 克隆
		tr.TLSClientConfig = s.transport.TLSClientConfig.Clone()

		if proxies != "" {
			u1, err := url2.Parse(proxies)
			if err != nil {
				return nil, err
			}
			tr.Proxy = http.ProxyURL(u1)
		}

		if !verify != tr.TLSClientConfig.InsecureSkipVerify {
			tr.TLSClientConfig.InsecureSkipVerify = !verify
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
				return nil, fmt.Errorf("failed to parse root certificate")
			}
			tr.TLSClientConfig.RootCAs = certPool
			tr.TLSClientConfig.Certificates = []utls.Certificate{certs}
		}

		if ja3String != "" && strings.HasPrefix(preq.Url, "https") && tr.H2Transport == nil {
			if tr.TLSClientConfig.ClientSessionCache == nil {
				tr.TLSClientConfig.ClientSessionCache = utls.NewLRUClientSessionCache(0)
			}
			if tr.TLSClientConfig.OmitEmptyPsk == false {
				tr.TLSClientConfig.OmitEmptyPsk = true
			}
			if strings.Contains(strings.Split(ja3String, ",")[2], "-41") {
				tr.TLSClientConfig.SessionTicketsDisabled = false
			}
			if tr.H2Transport == nil {
				var h2 *http.HTTP2Transport
				h2, _ = http.HTTP2ConfigureTransports(tr)
				tr.JA3 = ja3String
				tr.UserAgent = s.Headers.Get("User-Agent")
				tr.RandomJA3 = randomJA3
				tr.ForceHTTP1 = forceHTTP1
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
				tr.TLSExtensions = tlsExtensions
				tr.H2Transport = h2
			}
		} else if ja3String != "" && strings.HasPrefix(preq.Url, "https") && tr.H2Transport != nil {
			tr.JA3 = ja3String
			tr.UserAgent = s.Headers.Get("User-Agent")
			tr.RandomJA3 = randomJA3
			tr.ForceHTTP1 = forceHTTP1
			tr.TLSExtensions = tlsExtensions
			tr.H2Transport.(*http.HTTP2Transport).HTTP2Settings = http2Settings
		}

		if disableKeepAlives {
			tr.DisableKeepAlives = true
		}

		s.transportCache[cacheKey] = tr // 将新创建的 transport 存入缓存
		return tr, nil
	}()
	if err != nil {
		return nil, err
	}

	// --- 缓存逻辑结束 ---

	// 2. 使用获取到的 Transport 创建一个 Client
	// Client 的创建开销很小，每次请求都可以创建一个新的，以便设置请求特定的 Timeout
	client := &http.Client{
		Transport:     transport,
		CheckRedirect: nil,
		Timeout:       timeout,
	}

		// 是否自动转发（nil 表示跟随 Session 默认行为，即允许重定向）
		allowRedirects := true
		if req.AllowRedirects != nil {
			allowRedirects = *req.AllowRedirects
		}
		if allowRedirects {
		if client.CheckRedirect == nil {
			client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
				if len(via) > s.MaxRedirects {
					return fmt.Errorf("redirects number gt %d", s.MaxRedirects)
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
			reqValue := (*req.GetHeaders())[key]
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

	// ================= START DEBUG LOGIC =================
	var debugID string

	if IsDebug() {
		debugID = uuid.New().String()

		// 2. 打印原始 HTTP 请求报文 (Wire Format)
		if http2Settings != nil {
			request.ProtoMajor = 2
		}
		dumpReq, _ := httputil.DumpRequestOut(request, true)
		fmt.Printf("%s [ID: %s] [Wire] HTTP Raw Request:\n%s\n", time.Now().Format("2006/01/02 15:04:05"), debugID, string(dumpReq))
	}
	// ================= END DEBUG LOGIC =================

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	// ================= START DEBUG RESPONSE =================
	if IsDebug() {
		dumpResp, _ := httputil.DumpResponse(resp, true)
		fmt.Printf("%s [ID: %s] [Wire] HTTP Raw Response:\n%s\n", time.Now().Format("2006/01/02 15:04:05"), debugID, string(dumpResp))
	}
	// ================= END DEBUG RESPONSE =================

	response, err := s.buildResponse(resp, preq, req)
	if err != nil {
		return nil, err
	}
	response.History = history
	return response, nil
}

// 构建response参数
func (s *Session) buildResponse(resp *http.Response, preq *models.PrepareRequest, req *url.Request) (*models.Response, error) {
	response := &models.Response{
		Url:        preq.Url,
		Headers:    resp.Header,
		Cookies:    resp.Cookies(),
		Body:       resp.Body,
		StatusCode: resp.StatusCode,
		History:    []*models.Response{},
		Request:    req,
	}

	// 更新 Cookies
	if resp.Cookies() != nil {
		u, _ := url2.Parse(preq.Url)
		s.Cookies.SetCookies(u, resp.Cookies())
	}

	// 获取 Header 部分的原始数据 (不包含 Body)
	// DumpResponse(resp, false) 只会 Dump Header
	rawHeader, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return nil, err
	}

	// Stream 模式，直接返回
	if req.Stream {
		response.Body = resp.Body
		response.RawResponse = rawHeader
		return response, nil
	}

	// 确保最后关闭原始 Body
	defer resp.Body.Close()

	// 读取原始 Body 数据 (可能是压缩的，也可能是加密的)
	rawBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 拼装完整的 RawResponse (Header + RawBody)
	// 注意：httputil.DumpResponse 返回的数据末尾通常已经包含 \r\n\r\n，但如果是 false 模式可能需要检查一下分隔符
	// 这里直接拼接作为近似的 Wire Format
	response.RawResponse = append(rawHeader, rawBodyBytes...)

	// 处理解压逻辑
	// 无论服务端是否压缩，我们都尝试通过 DecompressBody 处理
	// 因为我们需要一个新的 Reader 来读取 rawBodyBytes
	var bodyReader io.ReadCloser = io.NopCloser(bytes.NewReader(rawBodyBytes))

	// 创建解压流
	decompressedStream, err := DecompressBody(bodyReader, resp.Header.Get("Content-Encoding"))
	if err != nil {
		return nil, err
	}
	defer decompressedStream.Close()

	// 读取解压后的数据到 Content
	contentBytes, err := io.ReadAll(decompressedStream)
	if err != nil {
		return nil, err
	}

	response.Content = contentBytes
	// 重新填充 Body，这样用户如果习惯去 Read Body 也能读到解压后的数据
	response.Body = io.NopCloser(bytes.NewReader(contentBytes))
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
		return r, nil // gzip.Reader 自动处理流式解压

	case "deflate":
		return flate.NewReader(body), nil // flate.Reader 直接返回流式接口

	case "br":
		return io.NopCloser(brotli.NewReader(body)), nil // brotli.Reader 实现 io.Reader 接口

	default:
		return body, nil // 非压缩数据原样返回
	}
}

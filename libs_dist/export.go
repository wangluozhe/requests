package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	url2 "net/url"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/google/uuid"
	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/chttp/cookiejar"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/libs"
	"github.com/wangluozhe/requests/models"
	ja3 "github.com/wangluozhe/requests/transport"
	"github.com/wangluozhe/requests/url"
	"github.com/wangluozhe/requests/utils"
)

var unsafePointers = make(map[string]*C.char)
var unsafePointersLock = sync.Mutex{}

var sessionsPool = sync.Map{}

// StreamEntry 保存一个活跃的流式响应连接
type StreamEntry struct {
	Response *models.Response
	Body     io.ReadCloser
	Session  *sessionEntry
	mu       sync.Mutex
	readMu   sync.Mutex
	closed   bool
}

var streamPool = sync.Map{}

// sessionEntry 管理 C API 对 Session 的所有权和活动调用。
// freeSession 会先标记关闭，再等待引用归零，避免请求使用已关闭的 Session。
type sessionEntry struct {
	session *requests.Session
	mu      sync.Mutex
	refs    int
	closing bool
	done    chan struct{}
	closed  chan struct{}
}

func newSessionEntry() *sessionEntry {
	return &sessionEntry{
		session: requests.NewSession(),
		done:    make(chan struct{}),
		closed:  make(chan struct{}),
	}
}

// sessionForID 返回可复用的 Session。Cookie Jar 只在首次创建时初始化。
func sessionForID(id string) *sessionEntry {
	for {
		if actual, ok := sessionsPool.Load(id); ok {
			entry := actual.(*sessionEntry)
			entry.mu.Lock()
			closing := entry.closing
			closed := entry.closed
			entry.mu.Unlock()
			if !closing {
				return entry
			}
			<-closed
			continue
		}

		candidate := newSessionEntry()
		actual, loaded := sessionsPool.LoadOrStore(id, candidate)
		if !loaded {
			return candidate
		}
		// 并发创建失败的候选 Session 不再归池管理，立即释放其 Transport。
		_ = candidate.session.Close()
		entry := actual.(*sessionEntry)
		entry.mu.Lock()
		closing := entry.closing
		closed := entry.closed
		entry.mu.Unlock()
		if !closing {
			return entry
		}
		<-closed
	}
}

func acquireSession(id string) *sessionEntry {
	for {
		entry := sessionForID(id)
		entry.mu.Lock()
		if !entry.closing {
			entry.refs++
			entry.mu.Unlock()
			return entry
		}
		closed := entry.closed
		entry.mu.Unlock()
		<-closed
	}
}

func releaseSession(entry *sessionEntry) {
	if entry == nil {
		return
	}
	entry.mu.Lock()
	entry.refs--
	if entry.closing && entry.refs == 0 {
		close(entry.done)
	}
	entry.mu.Unlock()
}

func (entry *StreamEntry) close() {
	if entry == nil {
		return
	}
	entry.mu.Lock()
	if entry.closed {
		entry.mu.Unlock()
		return
	}
	entry.closed = true
	body := entry.Body
	entry.mu.Unlock()
	if body != nil {
		_ = body.Close()
	}
}

func (entry *StreamEntry) read(buf []byte) (int, error) {
	entry.readMu.Lock()
	defer entry.readMu.Unlock()
	entry.mu.Lock()
	if entry.closed || entry.Body == nil {
		entry.mu.Unlock()
		return 0, errors.New("stream is closed")
	}
	body := entry.Body
	entry.mu.Unlock()
	// 不持有状态锁执行 Read，允许 close 同时关闭底层连接以唤醒阻塞读取。
	return body.Read(buf)
}

func closeStreamsForSession(session *sessionEntry) {
	streamPool.Range(func(key, value interface{}) bool {
		entry, ok := value.(*StreamEntry)
		if !ok || entry.Session != session {
			return true
		}
		streamID, ok := key.(string)
		if !ok {
			return true
		}
		if actual, deleted := streamPool.LoadAndDelete(streamID); deleted {
			if streamEntry, ok := actual.(*StreamEntry); ok {
				streamEntry.close()
			}
		}
		freeManagedCString(streamID)
		freeManagedCString(streamID + "_read")
		return true
	})
}

func closeSession(id string) error {
	actual, ok := sessionsPool.Load(id)
	if !ok {
		return nil
	}
	entry := actual.(*sessionEntry)
	entry.mu.Lock()
	if entry.closing {
		closed := entry.closed
		entry.mu.Unlock()
		<-closed
		return nil
	}
	entry.closing = true
	if entry.refs == 0 {
		close(entry.done)
	}
	done := entry.done
	entry.mu.Unlock()

	<-done
	closeStreamsForSession(entry)
	err := entry.session.Close()
	sessionsPool.CompareAndDelete(id, entry)
	close(entry.closed)
	return err
}

func GetSession(id string) *requests.Session {
	return sessionForID(id).session
}

// managedCString 统一登记返回给 C 调用方的字符串，调用 freeMemory 释放。
func managedCString(id, value string) *C.char {
	ptr := C.CString(value)
	unsafePointersLock.Lock()
	old := unsafePointers[id]
	unsafePointers[id] = ptr
	unsafePointersLock.Unlock()
	if old != nil {
		C.free(unsafe.Pointer(old))
	}
	return ptr
}

func freeManagedCString(id string) {
	unsafePointersLock.Lock()
	ptr, ok := unsafePointers[id]
	if ok {
		delete(unsafePointers, id)
	}
	unsafePointersLock.Unlock()
	if ok && ptr != nil {
		C.free(unsafe.Pointer(ptr))
	}
}

func errorCString(message string) *C.char {
	id := uuid.New().String()
	payload, err := json.Marshal(map[string]string{"id": id, "err": message})
	if err != nil {
		payload = []byte("{\"id\":\"" + id + "\",\"err\":\"C API error\"}")
	}
	return managedCString(id, string(payload))
}

func errorf(format string, args ...interface{}) *C.char {
	return errorCString(fmt.Sprintf(format, args...))
}

func goString(value *C.char) string {
	if value == nil {
		return ""
	}
	return C.GoString(value)
}

//export request
func request(requestParamsChar *C.char) *C.char {
	if requestParamsChar == nil {
		return errorCString("request 参数不能为空")
	}
	requestParamsString := goString(requestParamsChar)
	requestParams := libs.RequestParams{}
	err := json.Unmarshal([]byte(requestParamsString), &requestParams)
	if err != nil {
		return errorf("request 参数解析失败: %v", err)
	}

	req, err := buildRequest(requestParams)
	if err != nil {
		return errorf("构建 request 失败: %v", err)
	}

	sessionEntry := acquireSession(requestParams.Id)
	defer releaseSession(sessionEntry)
	response, err := sessionEntry.session.Request(requestParams.Method, requestParams.Url, req)
	if err != nil {
		return errorf("执行 request 失败: %v", err)
	}
	if response.Body != nil {
		defer response.Body.Close()
	}

	responseParams := make(map[string]interface{})
	responseID := uuid.New().String()
	responseParams["id"] = responseID
	responseParams["url"] = response.Url
	responseParams["headers"] = response.Headers
	responseParams["cookies"] = response.Cookies
	responseParams["status_code"] = response.StatusCode
	responseParams["content"] = utils.Base64Encode(response.Content)
	responseParams["raw"] = response.RawResponse

	responseParamsString, err := json.Marshal(responseParams)
	if err != nil {
		return errorf("序列化 response 失败: %v", err)
	}
	return managedCString(responseID, string(responseParamsString))
}

func buildRequest(requestParams libs.RequestParams) (*url.Request, error) {
	if requestParams.Method == "" {
		return nil, errors.New("method is null")
	}

	if requestParams.Url == "" {
		return nil, errors.New("url is null")
	}

	req := url.NewRequest()
	if requestParams.Params != nil {
		params := url.NewParams()
		for key, value := range requestParams.Params {
			params.Set(key, value)
		}
		req.Params = params
	}

	if requestParams.Headers != nil {
		headers := url.NewHeaders()
		if requestParams.UnChangedHeaderKey != nil {
			(*headers)[http.UnChangedHeaderKey] = requestParams.UnChangedHeaderKey
		}
		if requestParams.HeadersOrder != nil {
			(*headers)[http.HeaderOrderKey] = requestParams.HeadersOrder
		}
		for key, value := range requestParams.Headers {
			if strings.ToLower(key) != "content-length" {
				headers.Set(key, value)
			}
		}
		req.Headers = headers
	}

	if requestParams.Cookies != nil {
		cookies, _ := cookiejar.New(nil)
		u, err := url2.Parse(requestParams.Url)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, errors.New("invalid URL")
		}
		for key, value := range requestParams.Cookies {
			cookies.SetCookies(u, []*http.Cookie{&http.Cookie{
				Name:  key,
				Value: value,
			}})
		}
		req.Cookies = cookies
	}

	if requestParams.Data != nil {
		data := url.NewData()
		for key, value := range requestParams.Data {
			data.Set(key, value)
		}
		req.Data = data
	}

	if requestParams.Json != nil {
		req.Json = requestParams.Json
	}

	if requestParams.Body != "" {
		req.Body = bytes.NewReader(utils.Base64DecodeToBytes(requestParams.Body))
	}

	if requestParams.Auth != nil {
		req.Auth = requestParams.Auth
	}

	if requestParams.Timeout != 0 {
		timeout := requestParams.Timeout
		req.Timeout = time.Duration(timeout) * time.Second
	}

	if requestParams.AllowRedirectsSet() {
		req.AllowRedirects = url.Bool(requestParams.AllowRedirects)
	}

	if requestParams.Proxies != "" {
		req.Proxies = requestParams.Proxies
	}

	if requestParams.VerifySet() {
		req.Verify = url.Bool(requestParams.Verify)
	}

	if requestParams.Cert != nil {
		req.Cert = requestParams.Cert
	}

	if requestParams.Ja3 != "" {
		req.Ja3 = requestParams.Ja3
	}

	if requestParams.RandomJA3Set() {
		req.RandomJA3 = url.Bool(requestParams.RandomJA3)
	}

	if requestParams.ForceHTTP1Set() {
		req.ForceHTTP1 = url.Bool(requestParams.ForceHTTP1)
	}

	if requestParams.PseudoHeaderOrder != nil {
		if req.Headers == nil {
			req.Headers = url.NewHeaders()
		}
		(*req.GetHeaders())[http.PHeaderOrderKey] = requestParams.PseudoHeaderOrder
	}

	if requestParams.TLSExtensions != "" {
		tlsExtensions := &ja3.Extensions{}
		err := json.Unmarshal([]byte(requestParams.TLSExtensions), tlsExtensions)
		if err != nil {
			return nil, err
		}
		req.TLSExtensions, err = ja3.ToTLSExtensionsChecked(tlsExtensions)
		if err != nil {
			return nil, err
		}
	}

	if requestParams.HTTP2Settings != "" {
		http2Settings := &ja3.H2Settings{}
		err := json.Unmarshal([]byte(requestParams.HTTP2Settings), http2Settings)
		if err != nil {
			return nil, err
		}
		req.HTTP2Settings = ja3.ToHTTP2Settings(http2Settings)
	}

	if requestParams.Stream {
		req.Stream = true
	}
	return req, nil
}

//export freeMemory
func freeMemory(responseId *C.char) {
	if responseId == nil {
		return
	}
	freeManagedCString(goString(responseId))
}

//export freeSession
func freeSession(sessionId *C.char) {
	if sessionId == nil {
		return
	}
	_ = closeSession(goString(sessionId))
}

//export setDebug
func setDebug(enable bool) {
	requests.SetDebug(enable)
}

//export stream_request
func stream_request(requestParamsChar *C.char) *C.char {
	if requestParamsChar == nil {
		return errorCString("stream_request 参数不能为空")
	}
	requestParamsString := goString(requestParamsChar)
	requestParams := libs.RequestParams{}
	err := json.Unmarshal([]byte(requestParamsString), &requestParams)
	if err != nil {
		return errorf("stream_request 参数解析失败: %v", err)
	}

	// 强制开启 Stream 模式
	requestParams.Stream = true

	req, err := buildRequest(requestParams)
	if err != nil {
		return errorf("构建 stream request 失败: %v", err)
	}

	sessionEntry := acquireSession(requestParams.Id)
	defer releaseSession(sessionEntry)
	response, err := sessionEntry.session.Request(requestParams.Method, requestParams.Url, req)
	if err != nil {
		return errorf("执行 stream request 失败: %v", err)
	}

	// 生成 stream_id，将 response 存入流连接池（不关闭 body）
	streamId := uuid.New().String()
	streamPool.Store(streamId, &StreamEntry{
		Response: response,
		Body:     response.Body,
		Session:  sessionEntry,
	})

	// 返回元信息（不包含 body 内容）
	responseParams := make(map[string]interface{})
	responseParams["stream_id"] = streamId
	responseParams["url"] = response.Url
	responseParams["headers"] = response.Headers
	responseParams["cookies"] = response.Cookies
	responseParams["status_code"] = response.StatusCode

	responseParamsString, err := json.Marshal(responseParams)
	if err != nil {
		if actual, loaded := streamPool.LoadAndDelete(streamId); loaded {
			if entry, ok := actual.(*StreamEntry); ok {
				entry.close()
			}
		}
		return errorf("序列化 stream response 失败: %v", err)
	}
	return managedCString(streamId, string(responseParamsString))
}

//export stream_read
func stream_read(streamIdChar *C.char, size C.int) *C.char {
	if streamIdChar == nil {
		return errorCString("stream_read 的 stream_id 不能为空")
	}
	streamId := goString(streamIdChar)

	entry, ok := streamPool.Load(streamId)
	if !ok {
		return errorf("stream_read 找不到流: %s", streamId)
	}

	streamEntry := entry.(*StreamEntry)
	bufSize := int(size)
	if bufSize <= 0 {
		bufSize = 4096
	}

	buf := make([]byte, bufSize)
	n, err := streamEntry.read(buf)

	result := make(map[string]interface{})
	if n > 0 {
		result["data"] = utils.Base64Encode(buf[:n])
	} else {
		result["data"] = ""
	}

	if err == io.EOF {
		result["eof"] = true
	} else if err != nil {
		return errorf("stream_read 读取失败: %v", err)
	} else {
		result["eof"] = false
	}

	resultString, err := json.Marshal(result)
	if err != nil {
		return errorf("序列化 stream_read 结果失败: %v", err)
	}

	// stream_read 返回的指针用 stream_id + "_read" 作为 key 管理
	// 每次调用会覆盖上一次的指针，避免内存泄漏
	readKey := streamId + "_read"
	return managedCString(readKey, string(resultString))
}

//export stream_close
func stream_close(streamIdChar *C.char) {
	if streamIdChar == nil {
		return
	}
	streamId := goString(streamIdChar)
	if actual, ok := streamPool.LoadAndDelete(streamId); ok {
		if streamEntry, ok := actual.(*StreamEntry); ok {
			streamEntry.close()
		}
	}
	freeManagedCString(streamId)
	readKey := streamId + "_read"
	freeManagedCString(readKey)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			// 处理 panic，可以记录日志或采取其他措施
			fmt.Println("Recovered from panic:", r)
		}
	}()
}

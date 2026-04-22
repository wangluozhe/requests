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
var errorFormat = "{\"err\": \"%v\"}"

var sessionsPool = sync.Map{}

// StreamEntry 保存一个活跃的流式响应连接
type StreamEntry struct {
	Response *models.Response
	Body     io.ReadCloser
}

var streamPool = sync.Map{}

func GetSession(id string) *requests.Session {
	cookies, _ := cookiejar.New(nil)
	if actual, ok := sessionsPool.Load(id); ok {
		s := actual.(*requests.Session)
		s.Cookies = cookies
		return s
	}
	s := requests.NewSession()
	s.Cookies = cookies
	actual, _ := sessionsPool.LoadOrStore(id, s)
	return actual.(*requests.Session)
}

//export request
func request(requestParamsChar *C.char) *C.char {
	requestParamsString := C.GoString(requestParamsChar)
	requestParams := libs.RequestParams{}
	err := json.Unmarshal([]byte(requestParamsString), &requestParams)
	if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, "request->err := json.Unmarshal([]byte(requestParamsString), &requestParams) failed: "+err.Error()))
	}

	req, err := buildRequest(requestParams)
	if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, "request->req, err := buildRequest(requestParams) failed: "+err.Error()))
	}

	session := GetSession(requestParams.Id)
	response, err := session.Request(requestParams.Method, requestParams.Url, req)
	if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, "request->response, err := GetSession(requestParams.Id).Request(requestParams.Method, requestParams.Url, req) failed: "+err.Error()))
	}
	defer response.Body.Close()

	responseParams := make(map[string]interface{})
	responseParams["id"] = uuid.New().String()
	responseParams["url"] = response.Url
	responseParams["headers"] = response.Headers
	responseParams["cookies"] = response.Cookies
	responseParams["status_code"] = response.StatusCode
	responseParams["content"] = utils.Base64Encode(response.Content)
	responseParams["raw"] = response.RawResponse

	responseParamsString, err := json.Marshal(responseParams)
	if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, "request->responseParamsString, err := json.Marshal(responseParams) failed: "+err.Error()))
	}
	responseString := C.CString(string(responseParamsString))

	unsafePointersLock.Lock()
	unsafePointers[responseParams["id"].(string)] = responseString
	defer unsafePointersLock.Unlock()

	return responseString
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
		u, _ := url2.Parse(requestParams.Url)
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

	req.AllowRedirects = url.Bool(requestParams.AllowRedirects)

	if requestParams.Proxies != "" {
		req.Proxies = requestParams.Proxies
	}

	req.Verify = url.Bool(requestParams.Verify)

	if requestParams.Cert != nil {
		req.Cert = requestParams.Cert
	}

	if requestParams.Ja3 != "" {
		req.Ja3 = requestParams.Ja3
	}

	if requestParams.RandomJA3 {
		req.RandomJA3 = url.Bool(requestParams.RandomJA3)
	}

	if requestParams.ForceHTTP1 {
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
		req.TLSExtensions = ja3.ToTLSExtensions(tlsExtensions)
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
	responseIdString := C.GoString(responseId)

	unsafePointersLock.Lock()
	defer unsafePointersLock.Unlock()

	ptr, ok := unsafePointers[responseIdString]

	if !ok {
		fmt.Println("freeMemory:", ok)
		return
	}

	if ptr != nil {
		defer C.free(unsafe.Pointer(ptr))
	}

	delete(unsafePointers, responseIdString)
}

//export freeSession
func freeSession(sessionId *C.char) {
	sessionIdString := C.GoString(sessionId)
	sessionsPool.Delete(sessionIdString)
}

//export setDebug
func setDebug(enable bool) {
	requests.SetDebug(enable)
}

//export stream_request
func stream_request(requestParamsChar *C.char) *C.char {
	requestParamsString := C.GoString(requestParamsChar)
	requestParams := libs.RequestParams{}
	err := json.Unmarshal([]byte(requestParamsString), &requestParams)
	if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, "stream_request->json.Unmarshal failed: "+err.Error()))
	}

	// 强制开启 Stream 模式
	requestParams.Stream = true

	req, err := buildRequest(requestParams)
	if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, "stream_request->buildRequest failed: "+err.Error()))
	}

	session := GetSession(requestParams.Id)
	response, err := session.Request(requestParams.Method, requestParams.Url, req)
	if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, "stream_request->session.Request failed: "+err.Error()))
	}

	// 生成 stream_id，将 response 存入流连接池（不关闭 body）
	streamId := uuid.New().String()
	streamPool.Store(streamId, &StreamEntry{
		Response: response,
		Body:     response.Body,
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
		return C.CString(fmt.Sprintf(errorFormat, "stream_request->json.Marshal failed: "+err.Error()))
	}
	responseString := C.CString(string(responseParamsString))

	unsafePointersLock.Lock()
	unsafePointers[streamId] = responseString
	defer unsafePointersLock.Unlock()

	return responseString
}

//export stream_read
func stream_read(streamIdChar *C.char, size C.int) *C.char {
	streamId := C.GoString(streamIdChar)

	entry, ok := streamPool.Load(streamId)
	if !ok {
		return C.CString(fmt.Sprintf(errorFormat, "stream_read->stream not found: "+streamId))
	}

	streamEntry := entry.(*StreamEntry)
	bufSize := int(size)
	if bufSize <= 0 {
		bufSize = 4096
	}

	buf := make([]byte, bufSize)
	n, err := streamEntry.Body.Read(buf)

	result := make(map[string]interface{})
	if n > 0 {
		result["data"] = utils.Base64Encode(buf[:n])
	} else {
		result["data"] = ""
	}

	if err == io.EOF {
		result["eof"] = true
	} else if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, "stream_read->Read failed: "+err.Error()))
	} else {
		result["eof"] = false
	}

	resultString, err := json.Marshal(result)
	if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, "stream_read->json.Marshal failed: "+err.Error()))
	}

	// stream_read 返回的指针用 stream_id + "_read" 作为 key 管理
	// 每次调用会覆盖上一次的指针，避免内存泄漏
	readKey := streamId + "_read"
	unsafePointersLock.Lock()
	if oldPtr, exists := unsafePointers[readKey]; exists && oldPtr != nil {
		C.free(unsafe.Pointer(oldPtr))
	}
	responseString := C.CString(string(resultString))
	unsafePointers[readKey] = responseString
	unsafePointersLock.Unlock()

	return responseString
}

//export stream_close
func stream_close(streamIdChar *C.char) {
	streamId := C.GoString(streamIdChar)

	entry, ok := streamPool.Load(streamId)
	if !ok {
		return
	}

	streamEntry := entry.(*StreamEntry)
	if streamEntry.Body != nil {
		streamEntry.Body.Close()
	}

	streamPool.Delete(streamId)

	// 清理相关的 unsafePointers
	unsafePointersLock.Lock()
	defer unsafePointersLock.Unlock()

	// 清理 stream_id 对应的指针
	if ptr, exists := unsafePointers[streamId]; exists && ptr != nil {
		C.free(unsafe.Pointer(ptr))
		delete(unsafePointers, streamId)
	}

	// 清理 stream_read 对应的指针
	readKey := streamId + "_read"
	if ptr, exists := unsafePointers[readKey]; exists && ptr != nil {
		C.free(unsafe.Pointer(ptr))
		delete(unsafePointers, readKey)
	}
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			// 处理 panic，可以记录日志或采取其他措施
			fmt.Println("Recovered from panic:", r)
		}
	}()
}

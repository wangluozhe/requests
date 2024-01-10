package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/chttp/cookiejar"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/libs"
	ja3 "github.com/wangluozhe/requests/transport"
	"github.com/wangluozhe/requests/url"
	"github.com/wangluozhe/requests/utils"
	url2 "net/url"
	"strings"
	"sync"
	"time"
	"unsafe"
)

var unsafePointers = make(map[string]*C.char)
var unsafePointersLock = sync.Mutex{}
var errorFormat = "{\"err\": \"%v\"}"

var sessionsPool = make(map[string]*sync.Pool)
var sessionsPoolLock = sync.Mutex{}

func GetSession(id string) *requests.Session {
	sessionsPoolLock.Lock()
	defer sessionsPoolLock.Unlock()
	if sp, ok := sessionsPool[id]; ok {
		s := sp.Get().(*requests.Session)
		sp.Put(s)
		return s
	}
	sp := &sync.Pool{
		New: func() interface{} {
			return requests.NewSession()
		},
	}
	sessionsPool[id] = sp
	s := sp.Get().(*requests.Session)
	sp.Put(s)
	return s
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

	response, err := GetSession(requestParams.Id).Request(requestParams.Method, requestParams.Url, req)
	if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, "request->response, err := GetSession(requestParams.Id).Request(requestParams.Method, requestParams.Url, req) failed: "+err.Error()))
	}

	responseParams := make(map[string]interface{})
	responseParams["id"] = uuid.New().String()
	responseParams["url"] = response.Url
	responseParams["headers"] = response.Headers
	responseParams["cookies"] = response.Cookies
	responseParams["status_code"] = response.StatusCode
	responseParams["content"] = utils.Base64Encode(response.Text)

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
		for key, value := range requestParams.Headers {
			if strings.ToLower(key) != "content-length" {
				headers.Set(key, value)
			}
		}
		req.Headers = headers
		if requestParams.HeadersOrder != nil {
			(*req.Headers)[http.HeaderOrderKey] = requestParams.HeadersOrder
		}
		if requestParams.UnChangedHeaderKey != nil {
			(*req.Headers)[http.UnChangedHeaderKey] = requestParams.UnChangedHeaderKey
		}
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
		req.Body = requestParams.Body
	}

	if requestParams.Auth != nil {
		req.Auth = requestParams.Auth
	}

	if requestParams.Timeout != 0 {
		timeout := requestParams.Timeout
		req.Timeout = time.Duration(timeout) * time.Second
	}

	req.AllowRedirects = requestParams.AllowRedirects

	if requestParams.Proxies != "" {
		req.Proxies = requestParams.Proxies
	}

	req.Verify = requestParams.Verify

	if requestParams.Cert != nil {
		req.Cert = requestParams.Cert
	}

	if requestParams.Ja3 != "" {
		req.Ja3 = requestParams.Ja3
	}

	if requestParams.ForceHTTP1 {
		req.ForceHTTP1 = requestParams.ForceHTTP1
	}

	if requestParams.PseudoHeaderOrder != nil {
		(*req.Headers)[http.PHeaderOrderKey] = requestParams.PseudoHeaderOrder
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

func main() {
	defer func() {
		if r := recover(); r != nil {
			// 处理 panic，可以记录日志或采取其他措施
			fmt.Println("Recovered from panic:", r)
		}
	}()
}

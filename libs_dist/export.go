package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	http "github.com/wangluozhe/fhttp"
	"github.com/wangluozhe/fhttp/cookiejar"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/libs"
	ja3 "github.com/wangluozhe/requests/transport"
	"github.com/wangluozhe/requests/url"
	url2 "net/url"
	"time"
)

//export request
func request(requestParamsChar *C.char) *C.char {
	errorFormat := "{\"err\": \"%v\"}"
	requestParamsString := C.GoString(requestParamsChar)
	requestParams := libs.RequestParams{}
	err := json.Unmarshal([]byte(requestParamsString), &requestParams)
	if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, err.Error()))
	}

	if requestParams.Method == "" {
		return C.CString(fmt.Sprintf(errorFormat, "method is null"))
	}

	if requestParams.Url == "" {
		return C.CString(fmt.Sprintf(errorFormat, "url is null"))
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
			headers.Set(key, value)
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

	if requestParams.TLSExtensions != "" {
		tlsExtensions := &ja3.Extensions{}
		err = json.Unmarshal([]byte(requestParams.TLSExtensions), tlsExtensions)
		if err != nil {
			return C.CString(fmt.Sprintf(errorFormat, err.Error()))
		}
		req.TLSExtensions = ja3.ToTLSExtensions(tlsExtensions)
	}

	if requestParams.HTTP2Settings != "" {
		http2Settings := &ja3.H2Settings{}
		err = json.Unmarshal([]byte(requestParams.HTTP2Settings), http2Settings)
		if err != nil {
			return C.CString(fmt.Sprintf(errorFormat, err.Error()))
		}
		req.HTTP2Settings = ja3.ToHTTP2Settings(http2Settings)
	}

	response, err := requests.Request(requestParams.Method, requestParams.Url, req)
	if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, err.Error()))
	}

	responseParams := make(map[string]interface{})
	responseParams["url"] = response.Url
	responseParams["headers"] = response.Headers
	responseParams["cookies"] = response.Cookies
	responseParams["status_code"] = response.StatusCode
	responseParams["content"] = response.Text

	responseParamsString, err := json.Marshal(responseParams)
	if err != nil {
		return C.CString(fmt.Sprintf(errorFormat, err.Error()))
	}
	return C.CString(string(responseParamsString))
}

func main() {

}

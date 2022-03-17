package requests

import (
	"github.com/wangluozhe/requests/models"
	"github.com/wangluozhe/requests/url"
	"net/http"
)

func Request(method, rawurl string, req *url.Request) (*models.Response, error) {
	session := NewSession()
	return session.Request(method, rawurl, req)
}

func Get(rawurl string, req *url.Request) (*models.Response, error) {
	return Request(http.MethodGet, rawurl, req)
}

func Post(rawurl string, req *url.Request) (*models.Response, error) {
	return Request(http.MethodPost, rawurl, req)
}

func Options(rawurl string, req *url.Request) (*models.Response, error) {
	return Request(http.MethodOptions, rawurl, req)
}

func Head(rawurl string, req *url.Request) (*models.Response, error) {
	req.AllowRedirects = false
	return Request(http.MethodHead, rawurl, req)
}

func Put(rawurl string, req *url.Request) (*models.Response, error) {
	return Request(http.MethodPut, rawurl, req)
}

func Patch(rawurl string, req *url.Request) (*models.Response, error) {
	return Request(http.MethodPatch, rawurl, req)
}

func Delete(rawurl string, req *url.Request) (*models.Response, error) {
	return Request(http.MethodDelete, rawurl, req)
}

func Connect(rawurl string, req *url.Request) (*models.Response, error) {
	return Request(http.MethodConnect, rawurl, req)
}

func Trace(rawurl string, req *url.Request) (*models.Response, error) {
	return Request(http.MethodTrace, rawurl, req)
}
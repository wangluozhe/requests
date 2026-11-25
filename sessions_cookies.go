package requests

import (
	"net/url"
	"sync"

	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/chttp/cookiejar"
)

// requestCookieJar 按目标 URL 查询原 Jar，不把缺少作用域信息的 Cookies() 结果重新存储。
// 保持请求 Cookie 按名称覆盖 Session Cookie 的既有规则；响应更新只写入 Session，
// 不修改调用方的请求 Jar。每个请求链独立记录更新，防止删除后重新选中请求旧值。
type requestCookieJar struct {
	mu          sync.Mutex
	session     *cookiejar.Jar
	request     *cookiejar.Jar
	updated     *cookiejar.Jar
	skipInitial bool
}

func newRequestCookieJar(session, request *cookiejar.Jar, skipInitial bool) *requestCookieJar {
	if session == nil {
		session, _ = cookiejar.New(nil)
	}
	updated, _ := cookiejar.New(nil)
	return &requestCookieJar{session: session, request: request, updated: updated, skipInitial: skipInitial}
}

func (j *requestCookieJar) Cookies(u *url.URL) []*http.Cookie {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.skipInitial {
		j.skipInitial = false
		return nil
	}
	if j.request == nil || j.request == j.session {
		return j.session.Cookies(u)
	}
	updated := make(map[string]bool)
	for _, cookie := range j.updated.Cookies(u) {
		updated[cookie.Name] = true
	}
	var cookies []*http.Cookie
	requestNames := make(map[string]bool)
	for _, cookie := range j.request.Cookies(u) {
		if !updated[cookie.Name] {
			cookies = append(cookies, cookie)
			requestNames[cookie.Name] = true
		}
	}
	for _, cookie := range j.session.Cookies(u) {
		if !requestNames[cookie.Name] {
			cookies = append(cookies, cookie)
		}
	}
	return cookies
}

func (j *requestCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.session.SetCookies(u, cookies)
	for _, cookie := range cookies {
		// 标记与响应相同的域名和路径，包括删除、过期及 Secure 更新。
		// 标记在本请求链内保留，防止已过期或只能通过 HTTPS 发送的新值回退为旧值。
		j.updated.SetCookies(u, []*http.Cookie{{
			Name: cookie.Name, Value: "updated", Domain: cookie.Domain, Path: cookie.Path,
		}})
	}
}

package requests

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	stdurl "net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wangluozhe/requests/url"
)

func TestSessionRedirectCookieScopeAndHistory(t *testing.T) {
	for _, test := range []struct {
		name              string
		cookie            http.Cookie
		targetHost        string
		crossHostFirstHop bool
		wantCookie        string
	}{
		{name: "host-only cross-host first hop", cookie: http.Cookie{Path: "/"}, targetHost: "target.example.test", crossHostFirstHop: true},
		{name: "host-only cross-host", cookie: http.Cookie{Path: "/"}, targetHost: "target.example.test"},
		{name: "explicit domain cross-host", cookie: http.Cookie{Domain: "source.example.test", Path: "/"}, targetHost: "target.example.test"},
		{name: "default path", cookie: http.Cookie{}, targetHost: "source.example.test"},
		{name: "explicit path", cookie: http.Cookie{Path: "/source"}, targetHost: "source.example.test"},
		{name: "secure over HTTP", cookie: http.Cookie{Path: "/", Secure: true}, targetHost: "source.example.test"},
		{name: "same host", cookie: http.Cookie{Path: "/"}, targetHost: "source.example.test", wantCookie: "session=value"},
		{name: "shared parent domain", cookie: http.Cookie{Domain: "example.test", Path: "/"}, targetHost: "target.example.test", wantCookie: "session=value"},
	} {
		t.Run(test.name, func(t *testing.T) {
			const sourceURL = "http://source.example.test/source/start"
			middleURL := "http://source.example.test/middle"
			if test.crossHostFirstHop {
				middleURL = "http://" + test.targetHost + "/middle"
			}
			targetURL := "http://" + test.targetHost + "/final"
			cookie := test.cookie
			cookie.Name, cookie.Value = "session", "value"
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/source/start":
					http.SetCookie(w, &cookie)
					w.Header().Set("Location", middleURL)
					w.WriteHeader(http.StatusFound)
					_, _ = io.WriteString(w, "first")
				case "/middle":
					w.Header().Set("Location", targetURL)
					w.WriteHeader(http.StatusTemporaryRedirect)
					_, _ = io.WriteString(w, "second")
				default:
					_, _ = io.WriteString(w, r.Header.Get("Cookie"))
				}
			}))
			defer server.Close()
			session := NewSession()
			defer session.Close()
			session.Timeout = 5 * time.Second
			// Keep separate cookie origins while sending every request to a local server.
			session.transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
			}
			response, err := session.Get(sourceURL, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Close()
			if got := response.Text(); got != test.wantCookie {
				t.Errorf("redirected Cookie = %q, want %q", got, test.wantCookie)
			}
			if response.Url != targetURL || len(response.History) != 2 {
				t.Fatalf("final URL = %q, history length = %d", response.Url, len(response.History))
			}
			for i, want := range []struct {
				url, body string
				status    int
			}{
				{sourceURL, "first", http.StatusFound},
				{middleURL, "second", http.StatusTemporaryRedirect},
			} {
				history := response.History[i]
				if history.Url != want.url || history.Text() != want.body || history.StatusCode != want.status {
					t.Errorf("history[%d] = URL %q, body %q, status %d; want %+v", i, history.Url, history.Text(), history.StatusCode, want)
				}
			}
			// Query the original HTTPS origin to also verify that Secure cookies were retained.
			origin, _ := stdurl.Parse(strings.Replace(sourceURL, "http:", "https:", 1))
			if cookies := session.Cookies.Cookies(origin); len(cookies) != 1 || cookies[0].Value != "value" {
				t.Errorf("origin cookies = %v, want session=value", cookies)
			}
			// A later request must not receive cookies persisted under the wrong redirect origin/path.
			later, err := session.Get(targetURL, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer later.Close()
			if got := later.Text(); got != test.wantCookie {
				t.Errorf("later Cookie = %q, want %q", got, test.wantCookie)
			}
		})
	}
}

func TestSessionRedirectExplicitCookie(t *testing.T) {
	for _, test := range []struct {
		name       string
		setCookie  string
		wantCookie string
	}{
		{name: "unchanged", wantCookie: ""},
		{name: "updated", setCookie: "session=new; Path=/", wantCookie: "session=new"},
		{name: "deleted", setCookie: "session=; Path=/; Max-Age=0"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/start":
					if test.setCookie != "" {
						w.Header().Set("Set-Cookie", test.setCookie)
					}
					http.Redirect(w, r, "/middle", http.StatusFound)
				case "/middle":
					http.Redirect(w, r, "/final", http.StatusFound)
				default:
					_, _ = io.WriteString(w, r.Header.Get("Cookie"))
				}
			}))
			defer server.Close()
			session := NewSession()
			defer session.Close()
			session.Timeout = 5 * time.Second
			headers := url.NewHeaders()
			headers.Set("Cookie", "session=old")
			response, err := session.Get(server.URL+"/start", &url.Request{Headers: headers})
			if err != nil {
				t.Fatal(err)
			}
			defer response.Close()
			if got := response.Text(); got != test.wantCookie {
				t.Errorf("redirected Cookie = %q, want %q", got, test.wantCookie)
			}
		})
	}
}

func TestSessionRedirectControls(t *testing.T) {
	for _, test := range []struct {
		name           string
		allow          bool
		maxRedirects   int
		wantRequests   int32
		wantLimitError bool
	}{
		{name: "disabled", maxRedirects: 2, wantRequests: 1},
		{name: "limit exceeded", allow: true, maxRedirects: 1, wantRequests: 2, wantLimitError: true},
		{name: "at limit", allow: true, maxRedirects: 2, wantRequests: 3},
	} {
		t.Run(test.name, func(t *testing.T) {
			var count atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count.Add(1)
				switch r.URL.Path {
				case "/start":
					http.Redirect(w, r, "/middle", http.StatusFound)
				case "/middle":
					http.Redirect(w, r, "/final", http.StatusFound)
				default:
					_, _ = io.WriteString(w, "ok")
				}
			}))
			defer server.Close()
			session := NewSession()
			defer session.Close()
			session.Timeout = 5 * time.Second
			session.MaxRedirects = test.maxRedirects
			response, err := session.Get(server.URL+"/start", &url.Request{AllowRedirects: url.Bool(test.allow)})
			if response != nil {
				defer response.Close()
			}
			if got := count.Load(); got != test.wantRequests {
				t.Errorf("request count = %d, want %d", got, test.wantRequests)
			}
			if test.wantLimitError {
				if err == nil || !strings.Contains(err.Error(), "redirects number gt") {
					t.Fatalf("expected redirect limit error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !test.allow && (response.Url != server.URL+"/start" || response.StatusCode != http.StatusFound || len(response.History) != 0) {
				t.Fatalf("disabled redirect returned URL %q, status %d, history length %d", response.Url, response.StatusCode, len(response.History))
			}
			if test.allow && (response.Url != server.URL+"/final" || response.Text() != "ok" || len(response.History) != 2) {
				t.Fatalf("followed redirect returned URL %q, body %q, history length %d", response.Url, response.Text(), len(response.History))
			}
		})
	}
}

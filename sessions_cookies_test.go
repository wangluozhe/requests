package requests

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	stdurl "net/url"
	"strings"
	"testing"
	"time"

	chttp "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests/url"
)

func TestSessionRedirectSavedCookieScope(t *testing.T) {
	for _, source := range []string{"session", "request"} {
		for _, test := range []struct {
			name, initial, target, seed string
			cookie                      chttp.Cookie
			want                        string
		}{
			{name: "target host", initial: "http://source.example.test/start", target: "http://target.example.test/final", seed: "http://target.example.test/", cookie: chttp.Cookie{Path: "/"}, want: "saved=value"},
			{name: "narrow path", initial: "http://source.example.test/private/start", target: "http://source.example.test/private/other", cookie: chttp.Cookie{Path: "/private/start"}},
			{name: "broad path", initial: "http://source.example.test/private/start", target: "http://source.example.test/public/end", cookie: chttp.Cookie{Path: "/"}, want: "saved=value"},
			{name: "parent domain", initial: "http://source.example.test/start", target: "http://target.example.test/final", cookie: chttp.Cookie{Domain: "example.test", Path: "/"}, want: "saved=value"},
			{name: "host-only", initial: "http://source.example.test/start", target: "http://child.source.example.test/final", cookie: chttp.Cookie{Path: "/"}},
			{name: "secure downgrade", initial: "https://source.example.test/start", target: "http://source.example.test/final", cookie: chttp.Cookie{Path: "/", Secure: true}},
			{name: "expired", initial: "http://source.example.test/start", target: "http://source.example.test/final", cookie: chttp.Cookie{Path: "/", Expires: time.Unix(1, 0)}},
		} {
			t.Run(source+"/"+test.name, func(t *testing.T) {
				initialURL, _ := stdurl.Parse(test.initial)
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == initialURL.Path {
						http.Redirect(w, r, test.target, http.StatusFound)
						return
					}
					_, _ = io.WriteString(w, r.Header.Get("Cookie"))
				})
				plain := httptest.NewServer(handler)
				defer plain.Close()
				secure := httptest.NewTLSServer(handler)
				defer secure.Close()
				session := NewSession()
				defer session.Close()
				session.Verify = false
				session.Timeout = 5 * time.Second
				session.transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
					local := plain.Listener.Addr().String()
					if strings.HasSuffix(address, ":443") {
						local = secure.Listener.Addr().String()
					}
					return (&net.Dialer{}).DialContext(ctx, network, local)
				}
				jar := session.Cookies
				req := url.NewRequest()
				if source == "request" {
					jar = url.NewCookies()
					req.Cookies = jar
				}
				seedURL := initialURL
				if test.seed != "" {
					seedURL, _ = stdurl.Parse(test.seed)
				}
				cookie := test.cookie
				cookie.Name, cookie.Value = "saved", "value"
				jar.SetCookies(seedURL, []*chttp.Cookie{&cookie})
				response, err := session.Get(test.initial, req)
				if err != nil {
					t.Fatal(err)
				}
				defer response.Close()
				if got := response.Text(); got != test.want {
					t.Fatalf("redirected Cookie = %q, want %q", got, test.want)
				}
			})
		}
	}
}

func TestSessionRedirectRequestCookieUpdates(t *testing.T) {
	for _, test := range []struct {
		name, setCookie, want, wantStored string
	}{
		{name: "request overrides session", want: "session=request", wantStored: "stored"},
		{name: "updated", setCookie: "session=new; Path=/", want: "session=new", wantStored: "new"},
		{name: "deleted", setCookie: "session=; Path=/; Max-Age=0"},
		{name: "expired", setCookie: "session=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/start" {
					w.Header().Set("X-Received-Cookie", r.Header.Get("Cookie"))
					if test.setCookie != "" {
						w.Header().Set("Set-Cookie", test.setCookie)
					}
					http.Redirect(w, r, "/final", http.StatusFound)
					return
				}
				_, _ = io.WriteString(w, r.Header.Get("Cookie"))
			}))
			defer server.Close()
			session := NewSession()
			defer session.Close()
			session.Timeout = 5 * time.Second
			origin, _ := stdurl.Parse(server.URL)
			session.Cookies.SetCookies(origin, []*chttp.Cookie{{Name: "session", Value: "stored", Path: "/"}})
			requestJar := url.NewCookies()
			requestJar.SetCookies(origin, []*chttp.Cookie{{Name: "session", Value: "request", Path: "/"}})
			response, err := session.Get(server.URL+"/start", &url.Request{Cookies: requestJar})
			if err != nil {
				t.Fatal(err)
			}
			defer response.Close()
			if got := response.History[0].Headers.Get("X-Received-Cookie"); got != "session=request" {
				t.Errorf("initial Cookie = %q, want session=request", got)
			}
			if got := response.Text(); got != test.want {
				t.Errorf("redirected Cookie = %q, want %q", got, test.want)
			}
			if cookies := requestJar.Cookies(origin); len(cookies) != 1 || cookies[0].Value != "request" {
				t.Errorf("caller Cookie Jar was modified: %v", cookies)
			}
			stored := session.Cookies.Cookies(origin)
			if test.wantStored == "" {
				if len(stored) != 0 {
					t.Errorf("deleted Cookie remains in Session: %v", stored)
				}
			} else if len(stored) != 1 || stored[0].Value != test.wantStored {
				t.Errorf("Session cookies = %v, want session=%s", stored, test.wantStored)
			}
		})
	}
}

func TestSessionRedirectExplicitCookieRebuildsFromJar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			w.Header().Set("X-Received-Cookie", r.Header.Get("Cookie"))
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		_, _ = io.WriteString(w, r.Header.Get("Cookie"))
	}))
	defer server.Close()
	session := NewSession()
	defer session.Close()
	session.Timeout = 5 * time.Second
	origin, _ := stdurl.Parse(server.URL)
	session.Cookies.SetCookies(origin, []*chttp.Cookie{{Name: "stored", Value: "value", Path: "/"}})
	headers := url.NewHeaders()
	headers.Set("Cookie", "manual=one")
	response, err := session.Get(server.URL+"/start", &url.Request{Headers: headers})
	if err != nil {
		t.Fatal(err)
	}
	defer response.Close()
	if got := response.History[0].Headers.Get("X-Received-Cookie"); got != "manual=one" {
		t.Errorf("initial Cookie = %q, want manual=one", got)
	}
	if got := response.Text(); got != "stored=value" {
		t.Errorf("redirected Cookie = %q, want stored=value", got)
	}
}

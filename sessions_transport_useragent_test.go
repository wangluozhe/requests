package requests

import (
	"crypto/tls"
	stdhttp "net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/wangluozhe/requests/models"
	"github.com/wangluozhe/requests/url"
)

func TestSessionTransportUsesFinalUserAgent(t *testing.T) {
	const chrome = "Mozilla/5.0 Chrome/155.0.0.0"
	const firefox = "Mozilla/5.0 Firefox/148.0"
	for _, source := range []string{"session", "request", "middleware"} {
		for _, forceHTTP1 := range []bool{true, false} {
			t.Run(source+"/forceHTTP1="+strconv.FormatBool(forceHTTP1), func(t *testing.T) {
				var connectionGREASE sync.Map
				server := httptest.NewUnstartedServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
					grease, ok := connectionGREASE.Load(r.RemoteAddr)
					if !ok {
						t.Error("missing ClientHello for connection")
					}
					w.Header().Set("X-Test-Grease", strconv.FormatBool(grease == true))
					w.Header().Set("X-Test-Connection", r.RemoteAddr)
					w.Header().Set("X-Test-Protocol", r.Proto)
					_, _ = w.Write([]byte(r.UserAgent()))
				}))
				server.TLS = &tls.Config{GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
					grease := false
					for _, cipher := range hello.CipherSuites {
						if cipher>>8 == cipher&0xff && cipher&0x0f == 0x0a {
							grease = true
						}
					}
					connectionGREASE.Store(hello.Conn.RemoteAddr().String(), grease)
					return nil, nil
				}}
				server.EnableHTTP2 = true
				server.StartTLS()
				defer server.Close()

				session := NewSession()
				defer session.Close()
				session.Verify = false
				session.Timeout = 5 * time.Second
				session.Ja3 = transportTestJA3
				session.ForceHTTP1 = forceHTTP1
				session.Headers.Set("User-Agent", chrome)
				session.TLSExtensions = transportTestTLSSettings()
				session.TLSExtensions.NotUsedGREASE = false
				if source == "middleware" {
					session.Use(func(next Handler) Handler {
						return func(preq *models.PrepareRequest, req *url.Request) (*models.Response, error) {
							preq.Headers.Set("User-Agent", preq.Headers.Get("X-Test-User-Agent"))
							preq.Headers.Del("X-Test-User-Agent")
							return next(preq, req)
						}
					})
				}

				connections := make(map[string]string)
				owners := make(map[string]string)
				for _, test := range []struct {
					ua     string
					grease bool
				}{
					{firefox, false},
					{chrome, true},
					{firefox, false},
					{"Mozilla/5.0 Chrome/155.0.0.1", true},
					{chrome, true},
					{"", true}, // An explicit empty UA uses chttp's default fingerprint family.
					{"", true},
					{`Custom/1.0 ("quoted") &verify=false&ja3=x`, true},
				} {
					request := url.NewRequest()
					switch source {
					case "session":
						session.Headers.Set("User-Agent", test.ua)
					case "request":
						request.Headers = map[string]string{"User-Agent": test.ua}
					case "middleware":
						request.Headers = map[string]string{"User-Agent": chrome, "X-Test-User-Agent": test.ua}
					}
					response, err := session.Get(server.URL, request)
					if err != nil {
						t.Fatal(err)
					}
					wantWireUA, wantProtocol := test.ua, "HTTP/1.1"
					if !forceHTTP1 {
						wantProtocol = "HTTP/2.0"
						if wantWireUA == "" {
							// chttp HTTP/2 replaces an empty UA with its protocol default.
							wantWireUA = "Go-http-client/2.0"
						}
					}
					if got := response.Text(); got != wantWireUA {
						t.Errorf("wire User-Agent = %q, want %q", got, wantWireUA)
					}
					if got := response.Headers.Get("X-Test-Protocol"); got != wantProtocol {
						t.Errorf("protocol = %q, want %q", got, wantProtocol)
					}
					if got := response.Headers.Get("X-Test-Grease"); got != strconv.FormatBool(test.grease) {
						t.Errorf("UA %q sent a ClientHello with GREASE=%s, want %t", test.ua, got, test.grease)
					}
					connection := response.Headers.Get("X-Test-Connection")
					response.Close()
					if previous, exists := connections[test.ua]; exists && previous != connection {
						t.Errorf("same UA %q did not reuse its connection", test.ua)
					}
					if previous, exists := owners[connection]; exists && previous != test.ua {
						t.Errorf("different UAs %q and %q shared a connection", previous, test.ua)
					}
					connections[test.ua] = connection
					owners[connection] = test.ua
				}

				session.cacheLock.Lock()
				defer session.cacheLock.Unlock()
				if len(session.transportCache) != len(connections) {
					t.Errorf("cached transports = %d, want %d", len(session.transportCache), len(connections))
				}
				for _, tr := range session.transportCache {
					if _, exists := connections[tr.UserAgent]; !exists {
						t.Errorf("cached unexpected Transport.UserAgent = %q", tr.UserAgent)
					}
				}
			})
		}
	}
}

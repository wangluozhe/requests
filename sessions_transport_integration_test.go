package requests

import (
	"crypto/tls"
	"encoding/hex"
	"net"
	stdhttp "net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	utls "github.com/refraction-networking/utls"
	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests/url"
)

const transportTestJA3 = "771,4865-4866-4867,0-10-11-13-16-43-45-51,29-23,0"

func transportTestTLSSettings() *http.TLSExtensions {
	return &http.TLSExtensions{
		NotUsedGREASE: true,
		SupportedVersions: &utls.SupportedVersionsExtension{
			Versions: []uint16{utls.VersionTLS13},
		},
		KeyShareCurves: &utls.KeyShareExtension{KeyShares: []utls.KeyShare{{Group: utls.X25519}}},
	}
}

// Build a valid captured ClientHello locally, without contacting an external server.
func transportTestClientHello(t *testing.T) string {
	t.Helper()
	client, peer := net.Pipe()
	defer client.Close()
	defer peer.Close()
	conn := utls.UClient(client, &utls.Config{ServerName: "example.com"}, utls.HelloCustom)
	spec, err := transportTestTLSSettings().StringToSpec(transportTestJA3, "", false, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.ApplyPreset(spec); err != nil {
		t.Fatal(err)
	}
	if err := conn.BuildHandshakeState(); err != nil {
		t.Fatal(err)
	}
	hello := conn.HandshakeState.Hello.Raw
	record := append([]byte{22, 3, 1, byte(len(hello) >> 8), byte(len(hello))}, hello...)
	return hex.EncodeToString(record)
}

func TestSessionTransportInitialization(t *testing.T) {
	replay := transportTestClientHello(t)
	for _, test := range []struct {
		name            string
		ja3             bool
		randomJA3       bool
		replay          bool
		h2Settings      bool
		preconfiguredH2 bool
		forceHTTP1      bool
		wantProto       string
	}{
		{name: "default", wantProto: "HTTP/1.1"},
		{name: "JA3", ja3: true, h2Settings: true, wantProto: "HTTP/2.0"},
		{name: "random JA3", ja3: true, randomJA3: true, wantProto: "HTTP/2.0"},
		{name: "HTTP2 settings without JA3", h2Settings: true, wantProto: "HTTP/2.0"},
		{name: "existing HTTP2 transport", preconfiguredH2: true, h2Settings: true, wantProto: "HTTP/2.0"},
		{name: "existing HTTP2 forced HTTP1", preconfiguredH2: true, forceHTTP1: true, wantProto: "HTTP/1.1"},
		{name: "replay without JA3", replay: true, wantProto: "HTTP/2.0"},
		{name: "JA3 forced HTTP1", ja3: true, forceHTTP1: true, wantProto: "HTTP/1.1"},
		{name: "HTTP2 settings forced HTTP1", h2Settings: true, forceHTTP1: true, wantProto: "HTTP/1.1"},
		{name: "replay forced HTTP1", replay: true, forceHTTP1: true, wantProto: "HTTP/1.1"},
	} {
		for _, route := range []string{"direct HTTPS", "HTTP then HTTPS", "HTTP redirect"} {
			t.Run(test.name+"/"+route, func(t *testing.T) {
				hellos := make(chan []uint16, 1)
				secure := httptest.NewUnstartedServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
					_, _ = w.Write([]byte(r.Proto))
				}))
				secure.EnableHTTP2 = true
				secure.TLS = &tls.Config{GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
					select {
					case hellos <- append([]uint16(nil), hello.CipherSuites...):
					default:
					}
					return nil, nil
				}}
				secure.StartTLS()
				defer secure.Close()
				plain := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
					if r.URL.Path == "/redirect" {
						stdhttp.Redirect(w, r, secure.URL, stdhttp.StatusFound)
						return
					}
					_, _ = w.Write([]byte("plain"))
				}))
				defer plain.Close()

				session := NewSession()
				defer session.Close()
				session.Verify = false
				session.Timeout = 5 * time.Second
				if test.preconfiguredH2 {
					if _, err := http.HTTP2ConfigureTransports(session.transport); err != nil {
						t.Fatal(err)
					}
				}
				request := &url.Request{ForceHTTP1: Bool(test.forceHTTP1), RandomJA3: Bool(test.randomJA3)}
				if test.ja3 {
					session.Ja3 = transportTestJA3
					request.TLSExtensions = transportTestTLSSettings()
				}
				if test.replay {
					request.TLSExtensions = &http.TLSExtensions{ClientHelloHexStream: replay}
				}
				if test.h2Settings {
					request.HTTP2Settings = &http.HTTP2Settings{
						HeadersID: 3,
						Settings: []http.HTTP2Setting{
							{ID: http.HTTP2SettingEnablePush, Val: 0},
							{ID: http.HTTP2SettingMaxFrameSize, Val: 32768},
						},
					}
				}
				if route == "HTTP then HTTPS" {
					response, err := session.Get(plain.URL, request)
					if err != nil {
						t.Fatal(err)
					}
					response.Close()
				}
				target := secure.URL
				if route == "HTTP redirect" {
					target = plain.URL + "/redirect"
				}
				response, err := session.Get(target, request)
				if err != nil {
					t.Fatal(err)
				}
				defer response.Close()
				if got := response.Text(); got != test.wantProto {
					t.Errorf("server received %s, want %s", got, test.wantProto)
				}
				if test.ja3 || test.replay {
					select {
					case ciphers := <-hellos:
						if !reflect.DeepEqual(ciphers, []uint16{4865, 4866, 4867}) {
							t.Errorf("configured ClientHello was not sent: cipher suites = %v", ciphers)
						}
					default:
						t.Error("no TLS ClientHello observed")
					}
				}
				if route == "HTTP redirect" && len(response.History) != 1 {
					t.Errorf("redirect history length = %d, want 1", len(response.History))
				}
				session.cacheLock.Lock()
				defer session.cacheLock.Unlock()
				if len(session.transportCache) != 1 {
					t.Errorf("transport cache entries = %d, want 1", len(session.transportCache))
				}
				if test.h2Settings && !test.forceHTTP1 {
					for _, tr := range session.transportCache {
						h2, ok := tr.H2Transport.(*http.HTTP2Transport)
						if !ok || !reflect.DeepEqual(h2.HTTP2Settings, request.HTTP2Settings) {
							t.Error("HTTP2 settings were not applied to the cached transport")
						}
					}
				}
			})
		}
	}
}

func TestSessionConcurrentTransportInitialization(t *testing.T) {
	server := httptest.NewUnstartedServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		_, _ = w.Write([]byte(r.Proto))
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()
	session := NewSession()
	defer session.Close()
	session.Verify = false
	session.Timeout = 5 * time.Second
	session.HTTP2Settings = &http.HTTP2Settings{HeadersID: 3}
	var workers sync.WaitGroup
	for range 12 {
		workers.Go(func() {
			response, err := session.Get(server.URL, nil)
			if err != nil {
				t.Error(err)
				return
			}
			defer response.Close()
			if got := response.Text(); got != "HTTP/2.0" {
				t.Errorf("server received %s, want HTTP/2.0", got)
			}
		})
	}
	workers.Wait()
	session.cacheLock.Lock()
	defer session.cacheLock.Unlock()
	if len(session.transportCache) != 1 {
		t.Errorf("concurrent requests created %d transports, want 1", len(session.transportCache))
	}
}

func TestSessionRejectsInvalidReplayWithoutJA3(t *testing.T) {
	server := httptest.NewTLSServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {}))
	defer server.Close()
	session := NewSession()
	defer session.Close()
	session.Verify = false
	session.Timeout = 5 * time.Second
	response, err := session.Get(server.URL, &url.Request{
		TLSExtensions: &http.TLSExtensions{ClientHelloHexStream: "invalid hex"},
	})
	if response != nil {
		response.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "decoding ClientHello") {
		t.Fatalf("invalid capture was silently ignored: response=%v, err=%v", response, err)
	}
}

package requests

import (
	"crypto/tls"
	"fmt"
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

func TestSessionClientHelloSelection(t *testing.T) {
	replay := transportTestClientHello(t)
	preset := utls.HelloFirefox_148
	spec, err := utls.UTLSIdToSpec(preset)
	if err != nil {
		t.Fatal(err)
	}
	var hellos sync.Map
	secure := httptest.NewUnstartedServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		ciphers, ok := hellos.Load(r.RemoteAddr)
		if !ok {
			t.Error("missing ClientHello")
		}
		w.Header().Set("X-Test-Ciphers", fmt.Sprint(ciphers))
		_, _ = w.Write([]byte(r.Proto))
	}))
	secure.EnableHTTP2 = true
	secure.TLS = &tls.Config{GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
		hellos.Store(hello.Conn.RemoteAddr().String(), append([]uint16(nil), hello.CipherSuites...))
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
	invalidID := &utls.ClientHelloID{Client: "unknown", Version: "0"}
	for _, test := range []struct {
		name       string
		sessionHex string
		sessionID  *utls.ClientHelloID
		sessionTLS *http.TLSExtensions
		request    url.Request
		wantPreset bool
	}{
		{name: "session capture", sessionHex: replay},
		{name: "request capture overrides session capture", sessionHex: "invalid", request: url.Request{ClientHelloHexStream: replay}},
		{name: "session preset", sessionID: &preset, wantPreset: true},
		{name: "request preset overrides session preset", sessionID: invalidID, request: url.Request{ClientHelloID: &preset}, wantPreset: true},
		{name: "legacy capture wins", request: url.Request{TLSExtensions: &http.TLSExtensions{ClientHelloHexStream: replay}, ClientHelloHexStream: "invalid", ClientHelloID: invalidID, Ja3: "invalid"}},
		{name: "direct capture wins", request: url.Request{ClientHelloHexStream: replay, ClientHelloID: invalidID, Ja3: "invalid"}},
		{name: "preset wins over JA3", request: url.Request{ClientHelloID: &preset, Ja3: "invalid"}, wantPreset: true},
		{name: "inherited legacy capture wins", sessionTLS: &http.TLSExtensions{ClientHelloHexStream: replay}, request: url.Request{ClientHelloHexStream: "invalid"}},
		{name: "inherited direct capture wins", sessionHex: replay, request: url.Request{ClientHelloID: invalidID}},
		{name: "request TLS replaces inherited legacy capture", sessionTLS: &http.TLSExtensions{ClientHelloHexStream: "invalid"}, request: url.Request{TLSExtensions: &http.TLSExtensions{}, ClientHelloHexStream: replay}},
	} {
		for _, forceHTTP1 := range []bool{false, true} {
			for _, route := range []string{"HTTPS", "HTTP then HTTPS", "HTTP redirect"} {
				t.Run(fmt.Sprintf("%s/forceHTTP1=%t/%s", test.name, forceHTTP1, route), func(t *testing.T) {
					session := NewSession()
					defer session.Close()
					session.Verify = false
					session.Timeout = 5 * time.Second
					session.ForceHTTP1 = forceHTTP1
					session.ClientHelloHexStream = test.sessionHex
					session.ClientHelloID = test.sessionID
					session.TLSExtensions = test.sessionTLS
					request := test.request
					if route == "HTTP then HTTPS" {
						response, err := session.Get(plain.URL, &request)
						if err != nil {
							t.Fatal(err)
						}
						response.Close()
					}
					target := secure.URL
					if route == "HTTP redirect" {
						target = plain.URL + "/redirect"
					}
					response, err := session.Get(target, &request)
					if err != nil {
						t.Fatal(err)
					}
					defer response.Close()
					wantProto := "HTTP/2.0"
					if forceHTTP1 {
						wantProto = "HTTP/1.1"
					}
					if got := response.Text(); got != wantProto {
						t.Errorf("protocol = %q, want %q", got, wantProto)
					}
					wantCiphers := []uint16{4865, 4866, 4867}
					if test.wantPreset {
						wantCiphers = spec.CipherSuites
					}
					if got := response.Headers.Get("X-Test-Ciphers"); got != fmt.Sprint(wantCiphers) {
						t.Errorf("ClientHello ciphers = %s, want %v", got, wantCiphers)
					}
					session.cacheLock.Lock()
					defer session.cacheLock.Unlock()
					if len(session.transportCache) != 1 {
						t.Errorf("cached transports = %d, want 1", len(session.transportCache))
					}
				})
			}
		}
	}
}

func TestSessionClientHelloCacheIsolation(t *testing.T) {
	server := httptest.NewUnstartedServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		_, _ = w.Write([]byte(r.RemoteAddr))
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()
	session := NewSession()
	defer session.Close()
	session.Verify = false
	session.Timeout = 5 * time.Second
	requests := []url.Request{
		{ClientHelloHexStream: transportTestClientHello(t)},
		{ClientHelloHexStream: transportTestClientHello(t)},
		{ClientHelloID: &utls.HelloFirefox_148},
		{ClientHelloID: &utls.HelloChrome_155},
	}
	connections := make(map[int]string)
	owners := make(map[string]int)
	for _, index := range []int{0, 1, 2, 3, 0, 2, 1, 3} {
		request := requests[index]
		// Equal values with different pointer addresses must share a Transport.
		request.ClientHelloID = cloneClientHelloID(request.ClientHelloID)
		response, err := session.Get(server.URL, &request)
		if err != nil {
			t.Fatal(err)
		}
		connection := response.Text()
		response.Close()
		if previous, ok := connections[index]; ok && previous != connection {
			t.Errorf("configuration %d did not reuse its connection", index)
		}
		if owner, ok := owners[connection]; ok && owner != index {
			t.Errorf("configurations %d and %d shared a connection", owner, index)
		}
		connections[index], owners[connection] = connection, index
	}
	if len(session.transportCache) != len(requests) {
		t.Errorf("cached transports = %d, want %d", len(session.transportCache), len(requests))
	}
}

func TestSessionClientHelloSnapshot(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {}))
	defer server.Close()
	for _, source := range []string{"session", "request"} {
		t.Run(source, func(t *testing.T) {
			session := NewSession()
			defer session.Close()
			weights := utls.DefaultWeights
			id := utls.HelloRandomizedALPN
			id.Seed, id.Weights = &utls.PRNGSeed{1, 2}, &weights
			original := cloneClientHelloID(&id)
			request := &url.Request{}
			if source == "session" {
				session.ClientHelloID = &id
			} else {
				request.ClientHelloID = &id
			}
			response, err := session.Get(server.URL, request)
			if err != nil {
				t.Fatal(err)
			}
			response.Close()
			id.Client = "changed"
			id.Seed[0]++
			id.Weights.Extensions_Append_ALPN = -1
			for _, tr := range session.transportCache {
				if !reflect.DeepEqual(tr.ClientHelloID, original) {
					t.Errorf("cached preset changed through the caller's configuration: %+v", tr.ClientHelloID)
				}
			}
		})
	}
}

func TestClientHelloHashCoversFields(t *testing.T) {
	base := &utls.ClientHelloID{Client: "Randomized", Version: "0", Seed: &utls.PRNGSeed{}, Weights: &utls.Weights{}}
	seen := make(map[string]string)
	check := func(name, stream string, id *utls.ClientHelloID) {
		t.Helper()
		hash := clientHelloSettingsHash(stream, id)
		if previous, ok := seen[hash]; ok {
			t.Errorf("%s shares a hash with %s", name, previous)
		}
		seen[hash] = name
		if got := clientHelloSettingsHash(stream, cloneClientHelloID(id)); got != hash {
			t.Errorf("%s hash depends on pointer identity", name)
		}
	}
	check("absent", "", nil)
	check("empty preset", "", &utls.ClientHelloID{})
	check("capture A", "aa", nil)
	check("capture B", "bb", nil)
	check("base", "aa", base)
	for _, field := range []string{"Client", "Version", "Seed", "Weights"} {
		id := cloneClientHelloID(base)
		switch field {
		case "Client":
			id.Client += "!"
		case "Version":
			id.Version += "!"
		case "Seed":
			id.Seed = nil
		case "Weights":
			id.Weights = nil
		}
		check(field, "aa", id)
	}
	for i := range len(*base.Seed) {
		id := cloneClientHelloID(base)
		id.Seed[i] = 1
		check(fmt.Sprintf("seed byte %d", i), "aa", id)
	}
	// Enumerate upstream fields so dependency updates expose missing hash coverage.
	for i := range reflect.TypeOf(*base.Weights).NumField() {
		id := cloneClientHelloID(base)
		weights := reflect.ValueOf(id.Weights).Elem()
		weights.Field(i).SetFloat(0.5)
		check(weights.Type().Field(i).Name, "aa", id)
	}
}

func TestSessionConcurrentClientHelloPreset(t *testing.T) {
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
	id := utls.HelloRandomizedALPN
	weights := utls.DefaultWeights
	weights.TLSVersMax_Set_VersionTLS13 = 1
	id.Seed, id.Weights = &utls.PRNGSeed{1, 2, 3}, &weights
	original := cloneClientHelloID(&id)
	session.ClientHelloID = &id
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
				t.Errorf("protocol = %s, want HTTP/2.0", got)
			}
		})
	}
	workers.Wait()
	if !reflect.DeepEqual(&id, original) {
		t.Fatal("concurrent requests changed the caller's preset")
	}
	if len(session.transportCache) != 1 {
		t.Errorf("cached transports = %d, want 1", len(session.transportCache))
	}
}

func TestSessionRejectsInvalidClientHello(t *testing.T) {
	server := httptest.NewTLSServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {}))
	defer server.Close()
	for _, test := range []struct {
		name    string
		request url.Request
		wantErr string
	}{
		{"invalid hex", url.Request{ClientHelloHexStream: "invalid"}, "decoding ClientHello"},
		{"invalid record", url.Request{ClientHelloHexStream: "0000"}, "parsing ClientHello"},
		{"unknown preset", url.Request{ClientHelloID: &utls.ClientHelloID{Client: "unknown"}}, "selecting TLS ClientHello preset"},
		{"empty preset", url.Request{ClientHelloID: &utls.ClientHelloID{}}, "selecting TLS ClientHello preset"},
	} {
		t.Run(test.name, func(t *testing.T) {
			session := NewSession()
			defer session.Close()
			session.Verify = false
			session.Timeout = 5 * time.Second
			response, err := session.Get(server.URL, &test.request)
			if response != nil {
				response.Close()
			}
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("invalid configuration returned response=%v, err=%v", response, err)
			}
		})
	}
}

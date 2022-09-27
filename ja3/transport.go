package ja3

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	http "github.com/Danny-Dasilva/fhttp"
	http2 "github.com/Danny-Dasilva/fhttp/http2"
	"golang.org/x/net/proxy"

	utls "github.com/Danny-Dasilva/utls"
)

var errProtocolNegotiated = errors.New("protocol negotiated")

type Browser struct {
	JA3       string
	UserAgent string
}

func NewJA3Transport(browser Browser, proxyURL string, config *utls.Config) (http.RoundTripper, error) {
	if proxyURL != "" {
		dialer, err := NewConnectDialer(proxyURL, browser.UserAgent)
		if err != nil {
			return nil, err
		}
		if dialer != nil {

			return &JA3Transport{
				dialer: dialer,

				TLSClientConfig:   config,
				JA3:               browser.JA3,
				UserAgent:         browser.UserAgent,
				cachedTransports:  make(map[string]http.RoundTripper),
				cachedConnections: make(map[string]net.Conn),
			}, err
		}
	}
	return &JA3Transport{
		dialer: proxy.Direct,

		TLSClientConfig:   config,
		JA3:               browser.JA3,
		UserAgent:         browser.UserAgent,
		cachedTransports:  make(map[string]http.RoundTripper),
		cachedConnections: make(map[string]net.Conn),
	}, nil
}

type JA3Transport struct {
	sync.Mutex

	JA3             string
	UserAgent       string
	TLSClientConfig *utls.Config

	cachedConnections map[string]net.Conn
	cachedTransports  map[string]http.RoundTripper

	dialer proxy.ContextDialer
}

func (rt *JA3Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	addr := rt.getDialTLSAddr(req)
	if _, ok := rt.cachedTransports[addr]; !ok {
		if err := rt.getTransport(req, addr); err != nil {
			return nil, err
		}
	}
	return rt.cachedTransports[addr].RoundTrip(req)
}

func (rt *JA3Transport) getTransport(req *http.Request, addr string) error {
	switch strings.ToLower(req.URL.Scheme) {
	case "http":
		rt.cachedTransports[addr] = &http.Transport{DialContext: rt.dialer.DialContext, DisableKeepAlives: true}
		return nil
	case "https":
	default:
		return fmt.Errorf("invalid URL scheme: [%v]", req.URL.Scheme)
	}

	_, err := rt.dialTLS(context.Background(), "tcp", addr)
	switch err {
	case errProtocolNegotiated:
	case nil:
		// Should never happen.
		panic("dialTLS returned no error when determining cachedTransports")
	default:
		return err
	}

	return nil
}

func (rt *JA3Transport) dialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	rt.Lock()
	defer rt.Unlock()

	// If we have the connection from when we determined the HTTPS
	// cachedTransports to use, return that.
	if conn := rt.cachedConnections[addr]; conn != nil {
		delete(rt.cachedConnections, addr)
		return conn, nil
	}
	rawConn, err := rt.dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	var host string
	if host, _, err = net.SplitHostPort(addr); err != nil {
		host = addr
	}
	//////////////////

	spec, err := StringToSpec(rt.JA3, rt.UserAgent)
	if err != nil {
		return nil, err
	}

	conn := utls.UClient(rawConn, &utls.Config{ServerName: host, InsecureSkipVerify: true}, // MinVersion:         tls.VersionTLS10,
		// MaxVersion:         tls.VersionTLS13,

		utls.HelloCustom)

	if err := conn.ApplyPreset(spec); err != nil {
		return nil, err
	}

	if err = conn.Handshake(); err != nil {
		_ = conn.Close()

		if err.Error() == "tls: CurvePreferences includes unsupported curve" {
			//fix this
			return nil, fmt.Errorf("conn.Handshake() error for tls 1.3 (please retry request): %+v", err)
		}
		return nil, fmt.Errorf("uTlsConn.Handshake() error: %+v", err)
	}

	//////////
	if rt.cachedTransports[addr] != nil {
		return conn, nil
	}

	// No http.Transport constructed yet, create one based on the results
	// of ALPN.
	switch conn.ConnectionState().NegotiatedProtocol {
	case http2.NextProtoTLS:
		// t2 := http2.Transport{DialTLS: rt.dialTLSHTTP2}
		parsedUserAgent := parseUserAgent(rt.UserAgent)

		t2 := http2.Transport{DialTLS: rt.dialTLSHTTP2,
			PushHandler: &http2.DefaultPushHandler{},
			Navigator:   parsedUserAgent,
		}
		//	t2.Settings = []http2.Setting{
		//		{ID: http2.SettingMaxHeaderListSize, Val: 262144},
		//		{ID: http2.SettingMaxConcurrentStreams, Val: 1000},
		//
		//	}
		//// 	rTableSize:      "HEADER_TABLE_SIZE",
		//// SettingEnablePush:           "ENABLE_PUSH",
		//// SettingMaxConcurrentStreams: "MAX_CONCURRENT_STREAMS",
		//// SettingInitialWindowSize:    "INITIAL_WINDOW_SIZE",
		//// SettingMaxFrameSize:         "MAX_FRAME_SIZE",
		//// SettingMaxHeaderListSize:    "MAX_HEADER_LIST_SIZE",
		//
		//	t2.InitialWindowSize = 6291456
		//	t2.HeaderTableSize = 65536
		//	// t2.PushHandler = &http2.DefaultPushHandler{}
		//	// rt.cachedTransports[addr] = &t2
		rt.cachedTransports[addr] = &t2
	default:
		// Assume the remote peer is speaking HTTP 1.x + TLS.
		rt.cachedTransports[addr] = &http.Transport{DialTLSContext: rt.dialTLS}

	}

	// Stash the connection just established for use servicing the
	// actual request (should be near-immediate).
	rt.cachedConnections[addr] = conn

	return nil, errProtocolNegotiated
}

func (rt *JA3Transport) dialTLSHTTP2(network, addr string, _ *utls.Config) (net.Conn, error) {
	return rt.dialTLS(context.Background(), network, addr)
}

func (rt *JA3Transport) getDialTLSAddr(req *http.Request) string {
	host, port, err := net.SplitHostPort(req.URL.Host)
	if err == nil {
		return net.JoinHostPort(host, port)
	}
	return net.JoinHostPort(req.URL.Host, "443") // we can assume port is 443 at this point
}

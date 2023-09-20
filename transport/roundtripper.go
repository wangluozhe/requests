package transport

import (
	"context"
	"errors"
	"fmt"
	"net"

	"strings"
	"sync"

	utls "github.com/refraction-networking/utls"
	http "github.com/wangluozhe/fhttp"
	http2 "github.com/wangluozhe/fhttp/http2"
	"golang.org/x/net/proxy"
)

var errProtocolNegotiated = errors.New("protocol negotiated")

type roundTripper struct {
	sync.Mutex
	// fix typing
	JA3       string
	UserAgent string

	cachedConnections map[string]net.Conn
	cachedTransports  map[string]http.RoundTripper

	dialer        proxy.ContextDialer
	config        *utls.Config
	tlsExtensions *TLSExtensions
	http2Settings *http2.HTTP2Settings
	forceHTTP1    bool
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("user-agent") == "" {
		req.Header.Set("user-agent", rt.UserAgent)
	}
	addr := rt.getDialTLSAddr(req)
	if _, ok := rt.cachedTransports[addr]; !ok {
		if err := rt.getTransport(req, addr); err != nil {
			return nil, err
		}
	}
	return rt.cachedTransports[addr].RoundTrip(req)
}

func (rt *roundTripper) getTransport(req *http.Request, addr string) error {
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

func (rt *roundTripper) dialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
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

	spec, err := StringToSpec(rt.JA3, rt.UserAgent, rt.tlsExtensions, rt.forceHTTP1)
	if err != nil {
		return nil, err
	}

	rt.config.ServerName = host
	conn := utls.UClient(rawConn, rt.config.Clone(), utls.HelloCustom)

	if err := conn.ApplyPreset(spec); err != nil {
		return nil, err
	}

	if err = conn.HandshakeContext(ctx); err != nil {
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
		parsedUserAgent := parseUserAgent(rt.UserAgent)
		t2 := http2.Transport{
			DialTLS:         rt.dialTLSHTTP2,
			TLSClientConfig: rt.config,
			PushHandler:     &http2.DefaultPushHandler{},
			Navigator:       parsedUserAgent,
		}
		if rt.http2Settings != nil {
			t2.HTTP2Settings = rt.http2Settings
			if rt.http2Settings.Settings != nil {
				t2.Settings = rt.http2Settings.Settings
				for _, v := range rt.http2Settings.Settings {
					switch v.ID {
					case http2.SettingHeaderTableSize:
						t2.HeaderTableSize = v.Val
					case http2.SettingMaxConcurrentStreams:
						if v.Val == 0 {
							t2.StrictMaxConcurrentStreams = true
						} else {
							t2.StrictMaxConcurrentStreams = false
						}
					case http2.SettingInitialWindowSize:
						t2.InitialWindowSize = v.Val
					case http2.SettingMaxHeaderListSize:
						t2.MaxHeaderListSize = v.Val
					}
				}
			}
		}
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

func (rt *roundTripper) dialTLSHTTP2(network, addr string, _ *utls.Config) (net.Conn, error) {
	return rt.dialTLS(context.Background(), network, addr)
}

func (rt *roundTripper) getDialTLSAddr(req *http.Request) string {
	host, port, err := net.SplitHostPort(req.URL.Host)
	if err == nil {
		return net.JoinHostPort(host, port)
	}
	return net.JoinHostPort(req.URL.Host, "443") // we can assume port is 443 at this point
}

func newRoundTripper(browser Browser, config *utls.Config, tlsExtensions *TLSExtensions, http2Settings *http2.HTTP2Settings, forceHTTP1 bool, dialer ...proxy.ContextDialer) http.RoundTripper {
	if config == nil {
		if strings.Index(strings.Split(browser.JA3, ",")[2], "-41") == -1 {
			config = &utls.Config{
				InsecureSkipVerify: true,
			}
		} else {
			config = &utls.Config{
				InsecureSkipVerify: true,
				SessionTicketKey:   [32]byte{},
				ClientSessionCache: utls.NewLRUClientSessionCache(0),
				OmitEmptyPsk:       true,
			}
		}
	}
	if len(dialer) > 0 {

		return &roundTripper{
			dialer: dialer[0],

			JA3:               browser.JA3,
			UserAgent:         browser.UserAgent,
			cachedTransports:  make(map[string]http.RoundTripper),
			cachedConnections: make(map[string]net.Conn),
			config:            config,
			tlsExtensions:     tlsExtensions,
			http2Settings:     http2Settings,
			forceHTTP1:        forceHTTP1,
		}
	}

	return &roundTripper{
		dialer: proxy.Direct,

		JA3:               browser.JA3,
		UserAgent:         browser.UserAgent,
		cachedTransports:  make(map[string]http.RoundTripper),
		cachedConnections: make(map[string]net.Conn),
		config:            config,
		tlsExtensions:     tlsExtensions,
		http2Settings:     http2Settings,
		forceHTTP1:        forceHTTP1,
	}
}

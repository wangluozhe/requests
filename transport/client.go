package transport

import (
	utls "github.com/refraction-networking/utls"
	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/chttp/http2"

	"time"

	"golang.org/x/net/proxy"
)

type Browser struct {
	// Return a greeting that embeds the name in a message.
	JA3       string
	UserAgent string
}

var disabledRedirect = func(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func clientBuilder(browser Browser, config *utls.Config, tlsExtensions *TLSExtensions, http2Settings *http2.HTTP2Settings, forceHTTP1 bool, dialer proxy.ContextDialer, timeout int) http.Client {
	//if timeout is not set in call default to 15
	if timeout == 0 {
		timeout = 15
	}
	client := http.Client{
		Transport: newRoundTripper(browser, config, tlsExtensions, http2Settings, forceHTTP1, dialer),
		Timeout:   time.Duration(timeout) * time.Second,
	}
	return client
}

// newClient creates a new http transport
func newClient(browser Browser, timeout int, config *utls.Config, tlsExtensions *TLSExtensions, http2Settings *http2.HTTP2Settings, forceHTTP1 bool, proxyURL ...string) (http.Client, error) {
	//fix check PR
	if len(proxyURL) > 0 && len(proxyURL[0]) > 0 {
		dialer, err := newConnectDialer(proxyURL[0], browser.UserAgent)
		if err != nil {
			return http.Client{
				Timeout: time.Duration(timeout) * time.Second,
			}, err
		}
		return clientBuilder(
			browser,
			config,
			tlsExtensions,
			http2Settings,
			forceHTTP1,
			dialer,
			timeout,
		), nil
	}

	return clientBuilder(
		browser,
		config,
		tlsExtensions,
		http2Settings,
		forceHTTP1,
		proxy.Direct,
		timeout,
	), nil

}

type Options struct {
	Browser       Browser
	Timeout       int
	TLSConfig     *utls.Config
	TLSExtensions *TLSExtensions
	HTTP2Settings *http2.HTTP2Settings
	ForceHTTP1    bool
	Proxy         string
}

func NewClient(options *Options) (http.Client, error) {
	return newClient(options.Browser, options.Timeout, options.TLSConfig, options.TLSExtensions, options.HTTP2Settings, options.ForceHTTP1, options.Proxy)
}

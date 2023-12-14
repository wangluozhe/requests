package main

import (
	"fmt"
	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/transport"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	headers := &http.Header{
		"User-Agent":                []string{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/112.0"},
		"accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"},
		"accept-language":           []string{"zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2"},
		"accept-encoding":           []string{"gzip, deflate, br"},
		"upgrade-insecure-requests": []string{"1"},
		"sec-fetch-dest":            []string{"document"},
		"sec-fetch-mode":            []string{"navigate"},
		"sec-fetch-site":            []string{"none"},
		"sec-fetch-user":            []string{"?1"},
		"te":                        []string{"trailers"},
		http.PHeaderOrderKey: []string{
			":method",
			":path",
			":authority",
			":scheme",
		},
		http.HeaderOrderKey: []string{
			"user-agent",
			"accept",
			"accept-language",
			"accept-encoding",
			"upgrade-insecure-requests",
			"sec-fetch-dest",
			"sec-fetch-mode",
			"sec-fetch-site",
			"sec-fetch-user",
			"te",
		},
	}
	req.Headers = headers
	req.Ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-21,29-23-24,0"
	es := &transport.Extensions{
		SupportedSignatureAlgorithms: []string{
			"ECDSAWithP256AndSHA256",
			"ECDSAWithP384AndSHA384",
			"ECDSAWithP521AndSHA512",
			"PSSWithSHA256",
			"PSSWithSHA384",
			"PSSWithSHA512",
			"PKCS1WithSHA256",
			"PKCS1WithSHA384",
			"PKCS1WithSHA512",
			"ECDSAWithSHA1",
			"PKCS1WithSHA1",
		},
		//CertCompressionAlgo: []string{
		//	"brotli",
		//},
		RecordSizeLimit: 4001,
		DelegatedCredentials: []string{
			"ECDSAWithP256AndSHA256",
			"ECDSAWithP384AndSHA384",
			"ECDSAWithP521AndSHA512",
			"ECDSAWithSHA1",
		},
		SupportedVersions: []string{
			"1.3",
			"1.2",
		},
		PSKKeyExchangeModes: []string{
			"PskModeDHE",
		},
		KeyShareCurves: []string{
			"X25519",
			"P256",
		},
	}
	tes := transport.ToTLSExtensions(es)
	req.TLSExtensions = tes
	r, err := requests.Get("https://tls.peet.ws/api/all", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.Request.Headers)
	fmt.Println("url:", r.Url)
	fmt.Println("headers:", r.Headers)
	fmt.Println("text:", r.Text)
}

package main

import (
	"fmt"
	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/transport"
	"github.com/wangluozhe/requests/url"
)

func main() {
	session := requests.NewSession()
	req := url.NewRequest()
	headers := &http.Header{
		"User-Agent":                []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"},
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
	req.Ja3 = "771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,45-10-23-35-27-16-18-13-17513-65037-11-51-65281-5-0-43-41,4588-29-23-24,0"
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
		CertCompressionAlgo: []string{
			"brotli",
		},
		RecordSizeLimit:      0,
		DelegatedCredentials: nil,
		SupportedVersions: []string{
			"GREASE",
			"1.3",
			"1.2",
		},
		PSKKeyExchangeModes: []string{
			"PskModeDHE",
		},
		KeyShareCurves: []string{
			"GREASE",
			"4588",
			"X25519",
		},
	}
	tes := transport.ToTLSExtensions(es)
	req.TLSExtensions = tes
	r, err := session.Get("https://tls.peet.ws/api/all", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("text:", r.Text)
	r, err = session.Get("https://tls.peet.ws/api/all", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("text:", r.Text)
}

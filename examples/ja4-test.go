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
		"Accept":                    []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"},
		"Accept-Encoding":           []string{"gzip, deflate, br"},
		"Accept-Language":           []string{"zh-CN,zh;q=0.9"},
		"Cache-Control":             []string{"no-cache"},
		"Connection":                []string{"keep-alive"},
		"Host":                      []string{"gospider2.gospiderb.asia:8998"},
		"Pragma":                    []string{"no-cache"},
		"sec-ch-ua":                 []string{"\".Not/A)Brand\";v=\"99\", \"Google Chrome\";v=\"103\", \"Chromium\";v=\"103\""},
		"sec-ch-ua-mobile":          []string{"\"?0\""},
		"sec-ch-ua-platform":        []string{"\"Windows\""},
		"Sec-Fetch-Dest":            []string{"document"},
		"Sec-Fetch-Mode":            []string{"navigate"},
		"Sec-Fetch-Site":            []string{"none"},
		"Sec-Fetch-User":            []string{"\"?1\""},
		"Upgrade-Insecure-Requests": []string{"1"},
		"User-Agent":                []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"},
		http.HeaderOrderKey: []string{
			"HOST",
			"connection",
			"pragma",
			"Cache-Control",
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
			"upgrade-insecure-requests",
			"user-agent",
			"accept",
			"sec-fetch-site",
			"sec-fetch-mode",
			"sec-fetch-user",
			"sec-fetch-dest",
			"accept-encoding",
			"accept-language",
		},
		http.UnChangedHeaderKey: []string{
			"sec-ch-ua",
			"sec-ch-ua-mobile",
			"sec-ch-ua-platform",
		},
	}
	req.Headers = headers
	req.Ja3 = "772,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-17513-21-41,29-23-24,0"
	req.ForceHTTP1 = true
	es := &transport.Extensions{
		SupportedSignatureAlgorithms: []string{
			"ecdsa_secp256r1_sha256",
			"rsa_pss_rsae_sha256",
			"rsa_pkcs1_sha256",
			"ecdsa_secp384r1_sha384",
			"rsa_pss_rsae_sha384",
			"rsa_pkcs1_sha384",
			"rsa_pss_rsae_sha512",
			"rsa_pkcs1_sha512",
		},
		CertCompressionAlgo: []string{
			"brotli",
		},
		RecordSizeLimit: 4001,
		DelegatedCredentials: []string{
			"ECDSAWithP256AndSHA256",
			"ECDSAWithP384AndSHA384",
			"ECDSAWithP521AndSHA512",
			"ECDSAWithSHA1",
		},
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
			"X25519",
		},
	}
	tes := transport.ToTLSExtensions(es)
	req.TLSExtensions = tes
	r, err := requests.Get("https://gospider2.gospiderb.asia:8998/", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.Request.Headers)
	fmt.Println("url:", r.Url)
	fmt.Println("headers:", r.Headers)
	fmt.Println("text:", r.Text)
}

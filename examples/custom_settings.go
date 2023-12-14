package main

import (
	"fmt"
	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests/transport"
	"io"
	"log"
	"strings"
)

func main() {
	url := "https://tls.peet.ws/api/all"
	headers := http.Header{
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
	var browser = transport.Browser{
		JA3:       "771,4865-4867-4866-49195-49199-52393-52392-49196-49200-49162-49161-49171-49172-156-157-47-53,0-23-65281-10-11-35-16-5-34-51-43-13-45-28-21,29-23-24-25-256-257,0",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:109.0) Gecko/20100101 Firefox/112.0",
	}

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
	h2s := &transport.H2Settings{
		Settings: map[string]int{
			"HEADER_TABLE_SIZE": 65536,
			//"ENABLE_PUSH":            0,
			//"MAX_HEADER_LIST_SIZE":   262144,
			//"MAX_CONCURRENT_STREAMS": 1000,
			"INITIAL_WINDOW_SIZE": 131072,
			"MAX_FRAME_SIZE":      16384,
		},
		SettingsOrder: []string{
			"HEADER_TABLE_SIZE",
			"INITIAL_WINDOW_SIZE",
			"MAX_FRAME_SIZE",
		},
		ConnectionFlow: 12517377,
		HeaderPriority: map[string]interface{}{
			"weight":    42,
			"streamDep": 13,
			"exclusive": false,
		},
		PriorityFrames: []map[string]interface{}{
			{
				"streamID": 3,
				"priorityParam": map[string]interface{}{
					"weight":    201,
					"streamDep": 0,
					"exclusive": false,
				},
			},
			{
				"streamID": 5,
				"priorityParam": map[string]interface{}{
					"weight":    101,
					"streamDep": 0,
					"exclusive": false,
				},
			},
			{
				"streamID": 7,
				"priorityParam": map[string]interface{}{
					"weight":    1,
					"streamDep": 0,
					"exclusive": false,
				},
			},
			{
				"streamID": 9,
				"priorityParam": map[string]interface{}{
					"weight":    1,
					"streamDep": 7,
					"exclusive": false,
				},
			},
			{
				"streamID": 11,
				"priorityParam": map[string]interface{}{
					"weight":    1,
					"streamDep": 3,
					"exclusive": false,
				},
			},
			{
				"streamID": 13,
				"priorityParam": map[string]interface{}{
					"weight":    241,
					"streamDep": 0,
					"exclusive": false,
				},
			},
		},
	}
	tes := transport.ToTLSExtensions(es)
	h2ss := transport.ToHTTP2Settings(h2s)
	options := &transport.Options{
		Browser:       browser,
		Timeout:       30,
		TLSExtensions: tes,
		HTTP2Settings: h2ss,
	}
	client, err := transport.NewClient(options)
	request, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		log.Fatalln(err)
	}
	request.Header = headers
	response, err := client.Do(request)
	if err != nil {
		log.Print("Request Failed: " + err.Error())
	}
	content, _ := io.ReadAll(response.Body)
	fmt.Println(string(content))
	fmt.Println(response.Header)
}

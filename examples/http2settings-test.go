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
	h2ss := transport.ToHTTP2Settings(h2s)
	req.HTTP2Settings = h2ss
	r, err := requests.Get("https://tls.peet.ws/api/all", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.Request.Headers)
	fmt.Println("url:", r.Url)
	fmt.Println("headers:", r.Headers)
	fmt.Println("text:", r.Text)
}

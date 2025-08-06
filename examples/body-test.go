package main

import (
	"bytes"
	"fmt"

	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	req := url.NewRequest()
	headers := url.NewHeaders()
	headers.Set("User-Agent", "123")
	req.Headers = headers
	req.Body = bytes.NewReader([]byte{0x00, 0x01, 0x02, 0x03, 0x04})
	r, _ := requests.Post("https://httpbin.org/post", req)
	fmt.Println(r.Text)
}

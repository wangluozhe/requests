package main

import (
	"fmt"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/url"
)

func main() {
	data := url.NewData()
	// SetFile(name,fileName,filePath,contentType)
	// name为字段名，fileName为上传的文件名，filePath为上传文件的绝对路径，contentType为上传的文件类型
	// 如果contentType设置为""，则默认为"application/octet-stream"
	data.Set("page", "1")
	data.Set("limit", "10")
	req := url.NewRequest()
	req.Data = data
	r, err := requests.Post("http://httpbin.org/post", req)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r.Text)
}

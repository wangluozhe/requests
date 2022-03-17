package url

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"strings"
)

// 初始化Files结构体
func NewFiles() *Files {
	return &Files{}
}

// Files结构体
type Files struct {
	files    []map[string][]map[string]string
	indexKey []string
}

// Files设置Field参数
func (this *Files) SetField(name, value string) {
	f := map[string][]map[string]string{
		name: {{
			"type":  "field",
			"value": value,
		}},
	}
	index := SearchStrings(this.indexKey, name)
	if len(this.indexKey) == 0 || index == -1 {
		this.files = append(this.files, f)
		this.indexKey = append(this.indexKey, name)
	} else {
		this.files[index] = f
	}
}

// Files设置File参数
func (this *Files) SetFile(name, fileName, filePath, contentType string) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	f := map[string][]map[string]string{
		name: {{
			"type":        "file",
			"value":       fileName,
			"path":        filePath,
			"contentType": contentType,
		}},
	}
	index := SearchStrings(this.indexKey, name)
	if len(this.indexKey) == 0 || index == -1 {
		this.files = append(this.files, f)
		this.indexKey = append(this.indexKey, name)
	} else {
		this.files[index] = f
	}
}

// 获取Files参数值
func (this *Files) Get(name string) map[string]string {
	if len(this.files) != 0 {
		index := SearchStrings(this.indexKey, name)
		if index != -1 {
			return this.files[index][name][0]
		}
	}
	return nil
}

// Files添加Field参数
func (this *Files) AddField(name, value string) {
	f := map[string]string{
		"type":  "field",
		"value": value,
	}
	index := SearchStrings(this.indexKey, name)
	if len(this.indexKey) == 0 || index == -1 {
		this.SetField(name, value)
	} else {
		this.files[index][name] = append(this.files[index][name], f)
	}
}

// Files添加File参数
func (this *Files) AddFile(name, fileName, filePath, contentType string) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	f := map[string]string{
		"type":        "file",
		"value":       fileName,
		"path":        filePath,
		"contentType": contentType,
	}
	index := SearchStrings(this.indexKey, name)
	if len(this.indexKey) == 0 || index == -1 {
		this.SetFile(name, fileName, filePath, contentType)
	} else {
		this.files[index][name] = append(this.files[index][name], f)
	}
}

// 删除Files参数
func (this *Files) Del(name string) bool {
	index := SearchStrings(this.indexKey, name)
	if len(this.indexKey) == 0 || index == -1 {
		return false
	}
	this.files = append(this.files[:index], this.files[index+1:]...)
	this.indexKey = append(this.indexKey[:index], this.indexKey[index+1:]...)
	return true
}

// Files结构体转FormFile
func (this *Files) Encode() (*bytes.Buffer, string, error) {
	var uploadWriter io.Writer
	var uploadFile *os.File
	var err error

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for index, name := range this.indexKey {
		itemList := this.files[index][name]
		for _, item := range itemList {
			if item["type"] == "field" {
				writer.WriteField(name, item["value"])
			} else {
				contentType := item["contentType"]
				if contentType == "" {
					contentType = "application/octet-stream"
				}
				h := this.createFormFileHeader(name, item["value"], contentType)
				uploadWriter, err = writer.CreatePart(h)
				if err != nil {
					return nil, "", err
				}
				uploadFile, err = os.Open(item["path"])
				if err != nil {
					return nil, "", err
				}
				_, err = io.Copy(uploadWriter, uploadFile)
				if err != nil {
					return nil, "", err
				}
				err = uploadFile.Close()
				if err != nil {
					return nil, "", err
				}
				err = writer.Close()
				if err != nil {
					return nil, "", err
				}
			}
		}
	}
	return body, writer.FormDataContentType(), nil
}

// 创建文件Header
func (this *Files) createFormFileHeader(name, fileName, contentType string) textproto.MIMEHeader {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
		strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(name),
		strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(fileName)))
	h.Set("Content-Type", contentType)
	return h
}

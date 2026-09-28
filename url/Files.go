package url

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"strings"
	"sync"
)

// 初始化Files结构体
func NewFiles() *Files {
	return &Files{files: make(map[string][]map[string]string)}
}

// Files 按插入顺序保存上传条目，零值可用，支持并发方法调用。使用后不能按值复制，请使用 Clone。
type Files struct {
	files    map[string][]map[string]string
	indexKey []string
	mutex    sync.RWMutex
}

// Files设置Field参数
func (fs *Files) SetField(name, value string) {
	fs.setParam(name, map[string]string{
		"type":  "field",
		"value": value,
	})
}

// Files设置File参数
func (fs *Files) SetFile(name, fileName, filePath, contentType string) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	fs.setParam(name, map[string]string{
		"type":        "file",
		"value":       fileName,
		"path":        filePath,
		"contentType": contentType,
	})
}

// 设置参数的通用方法
func (fs *Files) setParam(name string, param map[string]string) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.initLocked()
	if _, exists := fs.files[name]; !exists {
		fs.indexKey = append(fs.indexKey, name)
	}
	fs.files[name] = []map[string]string{param}
}

// Get 返回第一个上传条目的独立副本；不存在时返回 nil。
func (fs *Files) Get(name string) map[string]string {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	if items := fs.files[name]; len(items) > 0 {
		return cloneFileParam(items[0])
	}
	return nil
}

// Files添加Field参数
func (fs *Files) AddField(name, value string) {
	fs.addParam(name, map[string]string{
		"type":  "field",
		"value": value,
	})
}

// Files添加File参数
func (fs *Files) AddFile(name, fileName, filePath, contentType string) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	fs.addParam(name, map[string]string{
		"type":        "file",
		"value":       fileName,
		"path":        filePath,
		"contentType": contentType,
	})
}

// 添加参数的通用方法
func (fs *Files) addParam(name string, param map[string]string) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.initLocked()
	if _, exists := fs.files[name]; !exists {
		fs.indexKey = append(fs.indexKey, name)
	}
	fs.files[name] = append(fs.files[name], param)
}

// 删除Files参数
func (fs *Files) Del(name string) bool {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	if _, exists := fs.files[name]; !exists {
		return false
	}
	delete(fs.files, name)
	index := SearchStrings(fs.indexKey, name)
	if index != -1 {
		copy(fs.indexKey[index:], fs.indexKey[index+1:])
		fs.indexKey[len(fs.indexKey)-1] = ""
		fs.indexKey = fs.indexKey[:len(fs.indexKey)-1]
	}
	return true
}

// Files结构体转FormFile
func (fs *Files) Encode() (*bytes.Buffer, string, error) {
	// 只在创建元数据快照时持锁，文件读取不阻塞其他方法。
	snapshot := fs.Clone()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, name := range snapshot.indexKey {
		itemList := snapshot.files[name]
		for _, item := range itemList {
			if item["type"] == "field" {
				if err := writer.WriteField(name, item["value"]); err != nil {
					return nil, "", err
				}
			} else {
				if err := fs.writeFile(writer, name, item); err != nil {
					return nil, "", err
				}
			}
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return body, writer.FormDataContentType(), nil
}

// Clone 深拷贝条目和顺序信息；磁盘文件仍在 Encode 时按路径读取。
func (fs *Files) Clone() *Files {
	if fs == nil {
		return nil
	}
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	return fs.cloneLocked()
}

func (fs *Files) cloneLocked() *Files {
	clone := &Files{
		indexKey: append([]string(nil), fs.indexKey...),
		files:    make(map[string][]map[string]string, len(fs.files)),
	}
	for name, items := range fs.files {
		cloneItems := make([]map[string]string, len(items))
		for index, item := range items {
			cloneItems[index] = cloneFileParam(item)
		}
		clone.files[name] = cloneItems
	}
	return clone
}

func (fs *Files) initLocked() {
	if fs.files == nil {
		fs.files = make(map[string][]map[string]string)
	}
}

func cloneFileParam(param map[string]string) map[string]string {
	if param == nil {
		return nil
	}
	clone := make(map[string]string, len(param))
	for key, value := range param {
		clone[key] = value
	}
	return clone
}

// 写入文件
func (fs *Files) writeFile(writer *multipart.Writer, name string, item map[string]string) error {
	contentType := item["contentType"]
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	h := fs.createFormFileHeader(name, item["value"], contentType)
	uploadWriter, err := writer.CreatePart(h)
	if err != nil {
		return err
	}
	uploadFile, err := os.Open(item["path"])
	if err != nil {
		return err
	}
	defer uploadFile.Close()
	if _, err = io.Copy(uploadWriter, uploadFile); err != nil {
		return err
	}
	return nil
}

// 创建文件Header
func (fs *Files) createFormFileHeader(name, fileName, contentType string) textproto.MIMEHeader {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
		strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(name),
		strings.NewReplacer("\\", "\\\\", `"`, "\\\"").Replace(fileName)))
	h.Set("Content-Type", contentType)
	return h
}

package models

import (
	"io"
	"mime"
	"mime/multipart"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/wangluozhe/requests/url"
)

func TestPrepareAuthUsesURLPasswordOnce(t *testing.T) {
	request := NewPrepareRequest()
	if err := request.Prepare_headers(nil); err != nil {
		t.Fatalf("Prepare_headers returned error: %v", err)
	}
	if err := request.Prepare_auth(nil, "https://alice:secret@example.com/"); err != nil {
		t.Fatalf("Prepare_auth returned error: %v", err)
	}

	const want = "Basic YWxpY2U6c2VjcmV0"
	if got := request.Headers.Get("Authorization"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

func TestPrepareFilesBodyCanReuseSharedCollections(t *testing.T) {
	files := url.NewFiles()
	files.AddField("tag", "base")
	files.AddField("fixed", "kept")
	data := url.NewValues()
	data.Add("tag", "one")
	data.Add("tag", "two")
	data.Add("later", "end")

	check := func(data *url.Values, want []string) {
		t.Helper()
		body, contentType, err := prepareFilesBody(files, data)
		if err != nil {
			t.Error(err)
			return
		}
		_, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			t.Error(err)
			return
		}
		reader := multipart.NewReader(strings.NewReader(body), params["boundary"])
		var got []string
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Error(err)
				return
			}
			value, err := io.ReadAll(part)
			if err != nil {
				t.Error(err)
				return
			}
			got = append(got, part.FormName()+"="+string(value))
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("multipart fields = %v, want %v", got, want)
		}
	}
	want := []string{"tag=base", "tag=one", "tag=two", "fixed=kept", "later=end"}
	// 顺序复用和并发复用都不能将 Data 累积到共享的 Files 中。
	check(data, want)
	check(data, want)
	var group sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			check(data, want)
		}()
	}
	group.Wait()
	check(nil, []string{"tag=base", "fixed=kept"})
	if got := data.Encode(); got != "tag=one&tag=two&later=end" {
		t.Fatalf("Data changed after preparing requests: %s", got)
	}
}

package url

import (
	"io"
	"mime"
	"mime/multipart"
	neturl "net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"sync"
	"testing"
)

func TestParamsValuesAreDeepCopies(t *testing.T) {
	params := NewParams()
	params.Add("key", "one")
	params.Add("key", "two")

	values := params.Values()
	values["key"][0] = "changed"
	values["new"] = []string{"value"}
	if got := params.Encode(); got != "key=one&key=two" {
		t.Fatalf("Params was modified through Values: %q", got)
	}

	clone := params.Clone()
	clone.values["key"][0] = "clone-only"
	clone.Add("key", "three")
	if got := params.Encode(); got != "key=one&key=two" {
		t.Fatalf("Params was modified through Clone: %q", got)
	}
}

func TestValuesAreDeepCopies(t *testing.T) {
	values := NewValues()
	values.Add("key", "one")
	values.Add("key", "two")

	snapshot := values.Values()
	snapshot["key"][0] = "changed"
	if got := values.Encode(); got != "key=one&key=two" {
		t.Fatalf("Values was modified through Values: %q", got)
	}

	clone := values.Clone()
	clone.values["key"][0] = "clone-only"
	clone.Set("key", "three")
	if got := values.Encode(); got != "key=one&key=two" {
		t.Fatalf("Values was modified through Clone: %q", got)
	}
}

func TestFilesGetAndCloneAreDeepCopies(t *testing.T) {
	files := NewFiles()
	files.SetFile("upload", "one.txt", "/tmp/one.txt", "text/plain")

	item := files.Get("upload")
	item["value"] = "changed.txt"
	if got := files.Get("upload")["value"]; got != "one.txt" {
		t.Fatalf("Files was modified through Get: %q", got)
	}

	clone := files.Clone()
	clone.files["upload"][0]["path"] = "clone-only"
	if files.Get("upload")["path"] != "/tmp/one.txt" {
		t.Fatal("Clone shares nested file metadata")
	}
	clone.SetFile("upload", "two.txt", "/tmp/two.txt", "text/plain")
	if got := files.Get("upload")["value"]; got != "one.txt" {
		t.Fatalf("Files was modified through Clone: %q", got)
	}
}

func TestCollectionsSupportConcurrentAccess(t *testing.T) {
	params := NewParams()
	values := NewValues()
	files := NewFiles()
	var group sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		worker := worker
		group.Add(1)
		go func() {
			defer group.Done()
			for i := 0; i < 100; i++ {
				key := "key-" + strconv.Itoa(worker) + "-" + strconv.Itoa(i)
				params.Set(key, "value")
				values.Set(key, "value")
				files.AddField(key, "value")
				_ = params.Get(key)
				_ = values.Get(key)
				_ = files.Get(key)
			}
		}()
	}
	group.Wait()
	if len(params.Keys()) != 800 || len(values.Keys()) != 800 {
		t.Fatalf("unexpected key count: params=%d values=%d", len(params.Keys()), len(values.Keys()))
	}
}

func TestCollectionsDoNotLoseConcurrentAdds(t *testing.T) {
	const workers = 8
	const addsPerWorker = 100
	params := NewParams()
	values := NewValues()
	files := NewFiles()
	var group sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for i := 0; i < addsPerWorker; i++ {
				params.Add("shared", "value")
				values.Add("shared", "value")
				files.AddField("shared", "value")
			}
		}()
	}
	group.Wait()

	want := workers * addsPerWorker
	if got := len(params.Values()["shared"]); got != want {
		t.Fatalf("Params retained %d concurrent values, want %d", got, want)
	}
	if got := len(values.Values()["shared"]); got != want {
		t.Fatalf("Values retained %d concurrent values, want %d", got, want)
	}
	if got := len(files.Clone().files["shared"]); got != want {
		t.Fatalf("Files retained %d concurrent values, want %d", got, want)
	}
}

func TestCollectionsOrderAndZeroValues(t *testing.T) {
	type collection interface {
		Set(string, string)
		Add(string, string)
		Del(string)
		Keys() []string
		Values() map[string][]string
		Encode() string
	}
	for name, c := range map[string]collection{"Params": new(Params), "Values": new(Values)} {
		t.Run(name, func(t *testing.T) {
			if c.Encode() != "" || len(c.Values()) != 0 {
				t.Fatal("nonempty zero value")
			}
			c.Del("missing")
			c.Add("b", "first")
			c.Add("a", "old")
			c.Add("b", "second")
			c.Set("a", "new")
			if got := c.Encode(); got != "b=first&b=second&a=new" {
				t.Fatalf("unexpected order: %s", got)
			}
			keys := c.Keys()
			keys[0] = "changed"
			if !reflect.DeepEqual(c.Keys(), []string{"b", "a"}) {
				t.Fatal("Keys shares internal storage")
			}
			c.Del("b")
			c.Add("b", "last")
			if got := c.Encode(); got != "a=new&b=last" {
				t.Fatalf("deleted key was not reinserted at the end: %s", got)
			}
		})
	}
	if got := ParseParams("b=x%20y&a=&b=z").Clone().Encode(); got != "b=x+y&b=z&a=" {
		t.Fatalf("Params encoding changed: %s", got)
	}
	if got := ParseValues("b=x y&a=&b=z&c=1=2").Clone().Encode(); got != "b=x%20y&b=z&a=&c=1%3D2" {
		t.Fatalf("Values encoding changed: %s", got)
	}
	var p *Params
	var v *Values
	var f *Files
	if p.Clone() != nil || v.Clone() != nil || f.Clone() != nil {
		t.Fatal("nil Clone should return nil")
	}
}

func TestParamsEncodeEscapesQueryComponents(t *testing.T) {
	params := NewParams()
	params.Add("a b&c", "x&y=z 中文")
	params.Add("empty", "")

	const want = "a+b%26c=x%26y%3Dz+%E4%B8%AD%E6%96%87&empty="
	if got := params.Encode(); got != want {
		t.Fatalf("Params.Encode() = %q, want %q", got, want)
	}

	parsed, err := neturl.ParseQuery(params.Encode())
	if err != nil {
		t.Fatalf("ParseQuery returned error: %v", err)
	}
	if got := parsed.Get("a b&c"); got != "x&y=z 中文" {
		t.Fatalf("decoded value = %q, want %q", got, "x&y=z 中文")
	}
	if got := parsed.Get("empty"); got != "" {
		t.Fatalf("decoded empty value = %q, want empty", got)
	}
}

func TestFilesMultipartOrderAndClone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "upload.txt")
	if err := os.WriteFile(path, []byte("file content"), 0600); err != nil {
		t.Fatal(err)
	}
	var files Files
	if files.Get("missing") != nil || files.Del("missing") {
		t.Fatal("nonempty zero value")
	}
	files.AddField("tag", "old")
	files.SetField("tag", "first")
	files.AddFile("upload", "one.txt", path, "")
	files.AddField("tag", "second")
	files.SetField("removed", "value")
	if !files.Del("removed") || files.Del("removed") {
		t.Fatal("incorrect Del result")
	}
	files.AddFile("upload", "two.txt", path, "text/plain")
	clone := files.Clone()
	files.Del("tag")
	files.AddField("tag", "source-only")
	body, contentType, err := clone.Encode()
	if err != nil {
		t.Fatal(err)
	}
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatal(err)
	}
	reader := multipart.NewReader(body, params["boundary"])
	var got []string
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, part.FormName()+"/"+part.FileName()+"/"+part.Header.Get("Content-Type")+"/"+string(content))
	}
	want := []string{"tag///first", "tag///second", "upload/one.txt/application/octet-stream/file content", "upload/two.txt/text/plain/file content"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("multipart parts = %#v, want %#v", got, want)
	}
	clone.SetFile("upload", "missing.txt", path+".missing", "")
	if _, _, err := clone.Encode(); err == nil {
		t.Fatal("expected missing file error")
	}
}

func TestCollectionsConcurrentSnapshotsAndDeletes(t *testing.T) {
	var params Params
	var values Values
	var files Files
	var group sync.WaitGroup
	start := make(chan struct{})
	for worker := 0; worker < 8; worker++ {
		group.Add(1)
		go func(worker int) {
			defer group.Done()
			<-start
			for i := 0; i < 100; i++ {
				key := strconv.Itoa(i % 8)
				if worker < 4 {
					params.Del(key)
					values.Del(key)
					files.Del(key)
					params.Set(key, "first")
					values.Set(key, "first")
					files.SetField(key, "first")
					params.Add(key, "second")
					values.Add(key, "second")
					files.AddField(key, "second")
					continue
				}
				_ = params.Encode()
				_ = values.Encode()
				_ = params.Keys()
				_ = values.Keys()
				for _, snapshot := range []map[string][]string{params.Values(), values.Values()} {
					for _, entries := range snapshot {
						entries[0] = "external-only"
					}
				}
				params.Clone().Set(key, "clone-only")
				values.Clone().Set(key, "clone-only")
				if item := files.Get(key); item != nil {
					item["value"] = "external-only"
				}
				files.Clone().SetField(key, "clone-only")
				if _, _, err := files.Encode(); err != nil {
					t.Error(err)
					return
				}
			}
		}(worker)
	}
	close(start)
	group.Wait()
}

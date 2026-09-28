package url

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

type benchmarkValues interface {
	Set(string, string)
	Add(string, string)
	Get(string) string
	Values() map[string][]string
	Encode() string
}

// 所有工作集固定为 16 个键，每键 2 个值，避免随 b.N 增长而扭曲 GC 成本。
func benchmarkCollection(b *testing.B, newCollection func() benchmarkValues) {
	keys := make([]string, 16)
	for i := range keys {
		keys[i] = "key" + strconv.Itoa(i)
	}
	build := func() benchmarkValues {
		v := newCollection()
		for _, key := range keys {
			v.Add(key, "first value")
			v.Add(key, "second&value")
		}
		return v
	}
	b.Run("Build16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = build()
		}
	})
	b.Run("SetExisting", func(b *testing.B) {
		v := build()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v.Set(keys[i%len(keys)], "updated")
		}
	})
	b.Run("Get", func(b *testing.B) {
		v := build()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if v.Get(keys[i%len(keys)]) == "" {
				b.Fatal("missing value")
			}
		}
	})
	b.Run("Snapshot", func(b *testing.B) {
		v := build()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if len(v.Values()) != len(keys) {
				b.Fatal("incomplete snapshot")
			}
		}
	})
	b.Run("Encode", func(b *testing.B) {
		v := build()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if v.Encode() == "" {
				b.Fatal("empty encoding")
			}
		}
	})
	// 旧版 Set/Get 可安全用于此固定键场景；不使用旧版会丢数据的并发 Add。
	b.Run("ParallelSetGet", func(b *testing.B) {
		v := build()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := keys[i%len(keys)]
				v.Set(key, "updated")
				if v.Get(key) == "" {
					b.Error("missing value")
				}
				i++
			}
		})
	})
}

func BenchmarkCollectionsParams(b *testing.B) {
	benchmarkCollection(b, func() benchmarkValues { return NewParams() })
}

func BenchmarkCollectionsValues(b *testing.B) {
	benchmarkCollection(b, func() benchmarkValues { return NewValues() })
}

func BenchmarkCollectionsFiles(b *testing.B) {
	keys := make([]string, 16)
	for i := range keys {
		keys[i] = "field" + strconv.Itoa(i)
	}
	build := func() *Files {
		fs := NewFiles()
		for _, key := range keys {
			fs.AddField(key, "first value")
			fs.AddField(key, "second&value")
		}
		return fs
	}
	b.Run("Build16", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = build()
		}
	})
	b.Run("Get", func(b *testing.B) {
		fs := build()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if fs.Get(keys[i%len(keys)])["value"] == "" {
				b.Fatal("missing field")
			}
		}
	})
	b.Run("EncodeFields", func(b *testing.B) {
		fs := build()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, _, err := fs.Encode(); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("EncodeFile64KiB", func(b *testing.B) {
		path := filepath.Join(b.TempDir(), "upload.bin")
		if err := os.WriteFile(path, bytes.Repeat([]byte("x"), 64*1024), 0600); err != nil {
			b.Fatal(err)
		}
		fs := NewFiles()
		fs.SetFile("upload", "upload.bin", path, "")
		b.SetBytes(64 * 1024)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, _, err := fs.Encode(); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("ParallelSetGet", func(b *testing.B) {
		fs := build()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := keys[i%len(keys)]
				fs.SetField(key, "updated")
				if fs.Get(key)["value"] == "" {
					b.Error("missing field")
				}
				i++
			}
		})
	})
}

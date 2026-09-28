package main

import (
	"sync"
	"testing"

	"github.com/wangluozhe/chttp/cookiejar"
	"github.com/wangluozhe/requests"
)

// BenchmarkCSessionLookupBefore 模拟重构前每次查找都重置 Cookie Jar 的路径。
func BenchmarkCSessionLookupBefore(b *testing.B) {
	var pool sync.Map
	id := "benchmark"
	session := requests.NewSession()
	pool.Store(id, session)
	defer session.Close()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cookies, _ := cookiejar.New(nil)
		actual, _ := pool.Load(id)
		actual.(*requests.Session).Cookies = cookies
	}
}

// BenchmarkCSessionLookupAfter 测量重构后复用 Session 与 Cookie Jar 的查找路径。
func BenchmarkCSessionLookupAfter(b *testing.B) {
	id := "benchmark-after"
	entry := sessionForID(id)
	defer closeSession(id)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if sessionForID(id) != entry {
			b.Fatal("session entry changed")
		}
	}
}

// BenchmarkCSessionLifecycleBefore 模拟重构前没有引用管理的直接查表路径。
func BenchmarkCSessionLifecycleBefore(b *testing.B) {
	var pool sync.Map
	id := "benchmark-lifecycle"
	session := requests.NewSession()
	pool.Store(id, session)
	defer session.Close()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := pool.Load(id); !ok {
			b.Fatal("session missing")
		}
	}
}

// BenchmarkCSessionAcquireReleaseAfter 测量生命周期引用计数的请求路径。
func BenchmarkCSessionAcquireReleaseAfter(b *testing.B) {
	id := "benchmark-acquire-release"
	_ = sessionForID(id)
	defer closeSession(id)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry := acquireSession(id)
		releaseSession(entry)
	}
}

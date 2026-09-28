package requests

import (
	"testing"

	utls "github.com/refraction-networking/utls"
	http "github.com/wangluozhe/chttp"
)

func benchmarkTLSExtensions() *http.TLSExtensions {
	return &http.TLSExtensions{
		SupportedVersions:    &utls.SupportedVersionsExtension{Versions: []uint16{utls.GREASE_PLACEHOLDER, utls.VersionTLS13, utls.VersionTLS12}},
		KeyShareCurves:       &utls.KeyShareExtension{KeyShares: []utls.KeyShare{{Group: utls.X25519}, {Group: utls.CurveP256}}},
		NotUsedGREASE:        true,
		ClientHelloHexStream: "1603010000",
	}
}

func benchmarkHTTP2Settings() *http.HTTP2Settings {
	return &http.HTTP2Settings{
		Settings:       []http.HTTP2Setting{{ID: http.HTTP2SettingHeaderTableSize, Val: 4096}, {ID: http.HTTP2SettingMaxConcurrentStreams, Val: 100}},
		ConnectionFlow: 65535,
		HeadersID:      1,
		HeaderPriority: &http.HTTP2PriorityParam{StreamDep: 3, Weight: 7, Exclusive: true},
		PriorityFrames: []http.HTTP2PriorityFrame{{HTTP2FrameHeader: http.HTTP2FrameHeader{StreamID: 9}, HTTP2PriorityParam: http.HTTP2PriorityParam{StreamDep: 2, Weight: 5}}},
	}
}

// BenchmarkSettingsSnapshot 对比 TLS/HTTP2 配置从请求合并到 Transport 前的快照成本。
func BenchmarkSettingsSnapshotTLS(b *testing.B) {
	request := benchmarkTLSExtensions()
	session := benchmarkTLSExtensions()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if merge_setting(request, session).(*http.TLSExtensions) == nil {
			b.Fatal("TLS snapshot is nil")
		}
	}
}

func BenchmarkSettingsSnapshotHTTP2(b *testing.B) {
	request := benchmarkHTTP2Settings()
	session := benchmarkHTTP2Settings()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if merge_setting(request, session).(*http.HTTP2Settings) == nil {
			b.Fatal("HTTP/2 snapshot is nil")
		}
	}
}

// BenchmarkTransportCacheLookup 只测量旧版和新版共有的 map 查找与互斥锁成本。
// 新版额外的 LRU 队列移动由专项测试覆盖，不把两种不同实现伪装成同一场景。
func BenchmarkTransportCacheLookup(b *testing.B) {
	session := NewSession()
	session.transportCache["stable"] = session.transport
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session.cacheLock.Lock()
		if session.transportCache["stable"] == nil {
			b.Fatal("transport cache miss")
		}
		session.cacheLock.Unlock()
	}
}

// BenchmarkTransportCacheHitWithSnapshot 模拟旧的缓存命中路径：先复制配置，再计算哈希和查找缓存。
func BenchmarkTransportCacheHitWithSnapshot(b *testing.B) {
	session := NewSession()
	session.transportCache["stable"] = session.transport
	tlsSettings := benchmarkTLSExtensions()
	h2Settings := benchmarkHTTP2Settings()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tlsExtensionsHash(cloneTLSExtensions(tlsSettings))
		_ = http2SettingsHash(cloneHTTP2Settings(h2Settings))
		session.cacheLock.Lock()
		_, found := session.cachedTransportLocked("stable")
		session.cacheLock.Unlock()
		if !found {
			b.Fatal("transport cache miss")
		}
	}
}

// BenchmarkTransportCacheHitWithoutSnapshot 模拟当前路径：缓存命中时只计算哈希和查找缓存。
func BenchmarkTransportCacheHitWithoutSnapshot(b *testing.B) {
	session := NewSession()
	session.transportCache["stable"] = session.transport
	tlsSettings := benchmarkTLSExtensions()
	h2Settings := benchmarkHTTP2Settings()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tlsExtensionsHash(tlsSettings)
		_ = http2SettingsHash(h2Settings)
		session.cacheLock.Lock()
		_, found := session.cachedTransportLocked("stable")
		session.cacheLock.Unlock()
		if !found {
			b.Fatal("transport cache miss")
		}
	}
}

func BenchmarkTLSSettingsHash(b *testing.B) {
	settings := benchmarkTLSExtensions()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if tlsExtensionsHash(settings) == "" {
			b.Fatal("TLS settings hash is empty")
		}
	}
}

func BenchmarkHTTP2SettingsHash(b *testing.B) {
	settings := benchmarkHTTP2Settings()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if http2SettingsHash(settings) == "" {
			b.Fatal("HTTP/2 settings hash is empty")
		}
	}
}

package requests

import (
	"reflect"
	"testing"

	utls "github.com/refraction-networking/utls"
	http "github.com/wangluozhe/chttp"
)

func TestMergeSettingCreatesImmutableTLSSnapshot(t *testing.T) {
	original := &http.TLSExtensions{
		SupportedSignatureAlgorithms: &utls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{utls.ECDSAWithP256AndSHA256}},
		CertCompressionAlgo:          &utls.UtlsCompressCertExtension{Algorithms: []utls.CertCompressionAlgo{utls.CertCompressionBrotli}},
		RecordSizeLimit:              &utls.FakeRecordSizeLimitExtension{Limit: 0x4001},
		DelegatedCredentials:         &utls.DelegatedCredentialsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{utls.ECDSAWithP384AndSHA384}},
		SupportedVersions:            &utls.SupportedVersionsExtension{Versions: []uint16{utls.VersionTLS13, utls.VersionTLS12}},
		PSKKeyExchangeModes:          &utls.PSKKeyExchangeModesExtension{Modes: []uint8{1, 2}},
		SignatureAlgorithmsCert:      &utls.SignatureAlgorithmsCertExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{utls.PSSWithSHA256}},
		KeyShareCurves:               &utls.KeyShareExtension{KeyShares: []utls.KeyShare{{Group: utls.X25519, Data: []byte{1, 2}}}},
		NotUsedGREASE:                true,
		ClientHelloHexStream:         "hello",
	}
	snapshot := merge_setting(original, (*http.TLSExtensions)(nil)).(*http.TLSExtensions)
	if snapshot == nil || snapshot == original {
		t.Fatal("TLS merge should return an independent snapshot")
	}
	snapshot.SupportedSignatureAlgorithms.SupportedSignatureAlgorithms[0] = utls.PKCS1WithSHA256
	snapshot.CertCompressionAlgo.Algorithms[0] = utls.CertCompressionZlib
	snapshot.RecordSizeLimit.Limit = 0x4002
	snapshot.DelegatedCredentials.SupportedSignatureAlgorithms[0] = utls.PKCS1WithSHA256
	snapshot.SupportedVersions.Versions[0] = utls.VersionTLS10
	snapshot.PSKKeyExchangeModes.Modes[0] = 3
	snapshot.SignatureAlgorithmsCert.SupportedSignatureAlgorithms[0] = utls.PKCS1WithSHA256
	snapshot.KeyShareCurves.KeyShares[0].Group = utls.CurveP256
	snapshot.KeyShareCurves.KeyShares[0].Data[0] = 9
	snapshot.NotUsedGREASE = false
	if !reflect.DeepEqual(original.SupportedSignatureAlgorithms.SupportedSignatureAlgorithms, []utls.SignatureScheme{utls.ECDSAWithP256AndSHA256}) ||
		!reflect.DeepEqual(original.CertCompressionAlgo.Algorithms, []utls.CertCompressionAlgo{utls.CertCompressionBrotli}) ||
		original.RecordSizeLimit.Limit != 0x4001 ||
		!reflect.DeepEqual(original.DelegatedCredentials.SupportedSignatureAlgorithms, []utls.SignatureScheme{utls.ECDSAWithP384AndSHA384}) ||
		!reflect.DeepEqual(original.SupportedVersions.Versions, []uint16{utls.VersionTLS13, utls.VersionTLS12}) ||
		!reflect.DeepEqual(original.PSKKeyExchangeModes.Modes, []uint8{1, 2}) ||
		!reflect.DeepEqual(original.SignatureAlgorithmsCert.SupportedSignatureAlgorithms, []utls.SignatureScheme{utls.PSSWithSHA256}) ||
		original.KeyShareCurves.KeyShares[0].Group != utls.X25519 ||
		!reflect.DeepEqual(original.KeyShareCurves.KeyShares[0].Data, []byte{1, 2}) ||
		!original.NotUsedGREASE {
		t.Fatal("TLS nested fields changed through the snapshot")
	}
}

func TestMergeSettingCreatesImmutableHTTP2Snapshot(t *testing.T) {
	original := &http.HTTP2Settings{
		Settings:       []http.HTTP2Setting{{ID: http.HTTP2SettingHeaderTableSize, Val: 4096}},
		HeadersID:      1,
		ConnectionFlow: 65535,
		HeaderPriority: &http.HTTP2PriorityParam{StreamDep: 3, Weight: 7, Exclusive: true},
		PriorityFrames: []http.HTTP2PriorityFrame{{HTTP2FrameHeader: http.HTTP2FrameHeader{StreamID: 9}, HTTP2PriorityParam: http.HTTP2PriorityParam{StreamDep: 2, Weight: 5}}},
	}
	snapshot := merge_setting((*http.HTTP2Settings)(nil), original).(*http.HTTP2Settings)
	if snapshot == nil || snapshot == original {
		t.Fatal("HTTP/2 merge should return an independent snapshot")
	}
	snapshot.Settings[0].Val = 8192
	snapshot.HeaderPriority.StreamDep = 99
	snapshot.PriorityFrames[0].HTTP2PriorityParam.StreamDep = 88
	if original.Settings[0].Val != 4096 || original.HeaderPriority.StreamDep != 3 || original.PriorityFrames[0].HTTP2PriorityParam.StreamDep != 2 {
		t.Fatal("HTTP/2 nested fields changed through the snapshot")
	}
}

func TestTransportCacheEvictsLeastRecentlyUsed(t *testing.T) {
	session := NewSession()
	session.TransportCacheLimit = 2
	session.cacheLock.Lock()
	session.storeTransportLocked("first", &http.Transport{})
	session.storeTransportLocked("second", &http.Transport{})
	if _, ok := session.cachedTransportLocked("first"); !ok {
		t.Fatal("expected first cache entry")
	}
	session.storeTransportLocked("third", &http.Transport{})
	session.cacheLock.Unlock()

	if len(session.transportCache) != 2 {
		t.Fatalf("cache size = %d, want 2", len(session.transportCache))
	}
	if _, ok := session.transportCache["second"]; ok {
		t.Fatal("least recently used transport was not evicted")
	}
	if _, ok := session.transportCache["first"]; !ok {
		t.Fatal("recently used transport was evicted")
	}
}

func TestSessionCloseClearsTransportCache(t *testing.T) {
	session := NewSession()
	session.cacheLock.Lock()
	session.storeTransportLocked("one", &http.Transport{})
	session.storeTransportLocked("two", &http.Transport{})
	session.cacheLock.Unlock()

	if err := session.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if len(session.transportCache) != 0 || session.transportLRU.Len() != 0 {
		t.Fatal("Close did not clear the transport cache")
	}
	if err := session.Close(); err != nil {
		t.Fatalf("Close should be idempotent: %v", err)
	}
}

func TestSessionCloseRejectsRequests(t *testing.T) {
	session := NewSession()
	if err := session.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	response, err := session.Get("http://127.0.0.1:1", nil)
	if err == nil || response != nil {
		t.Fatalf("closed session returned response=%v, err=%v", response, err)
	}
}

func TestSettingsHashesCoverConfigurationFields(t *testing.T) {
	tlsSettings := &http.TLSExtensions{
		SupportedSignatureAlgorithms: &utls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{utls.ECDSAWithP256AndSHA256}},
		CertCompressionAlgo:          &utls.UtlsCompressCertExtension{Algorithms: []utls.CertCompressionAlgo{utls.CertCompressionBrotli}},
		RecordSizeLimit:              &utls.FakeRecordSizeLimitExtension{Limit: 0x4001},
		DelegatedCredentials:         &utls.DelegatedCredentialsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{utls.ECDSAWithP384AndSHA384}},
		SupportedVersions:            &utls.SupportedVersionsExtension{Versions: []uint16{utls.VersionTLS13}},
		PSKKeyExchangeModes:          &utls.PSKKeyExchangeModesExtension{Modes: []uint8{1}},
		SignatureAlgorithmsCert:      &utls.SignatureAlgorithmsCertExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{utls.PSSWithSHA256}},
		KeyShareCurves:               &utls.KeyShareExtension{KeyShares: []utls.KeyShare{{Group: utls.X25519, Data: []byte{1}}}},
		NotUsedGREASE:                true,
		ClientHelloHexStream:         "hello",
	}
	tlsBaseHash := tlsExtensionsHash(tlsSettings)
	tlsVariants := []struct {
		name   string
		change func(*http.TLSExtensions)
	}{
		{"signature algorithms", func(v *http.TLSExtensions) { v.SupportedSignatureAlgorithms.SupportedSignatureAlgorithms[0]++ }},
		{"certificate compression", func(v *http.TLSExtensions) { v.CertCompressionAlgo.Algorithms[0]++ }},
		{"record size limit", func(v *http.TLSExtensions) { v.RecordSizeLimit.Limit++ }},
		{"delegated credentials", func(v *http.TLSExtensions) { v.DelegatedCredentials.SupportedSignatureAlgorithms[0]++ }},
		{"supported versions", func(v *http.TLSExtensions) { v.SupportedVersions.Versions[0]++ }},
		{"PSK modes", func(v *http.TLSExtensions) { v.PSKKeyExchangeModes.Modes[0]++ }},
		{"certificate signatures", func(v *http.TLSExtensions) { v.SignatureAlgorithmsCert.SupportedSignatureAlgorithms[0]++ }},
		{"key share group", func(v *http.TLSExtensions) { v.KeyShareCurves.KeyShares[0].Group++ }},
		{"key share data", func(v *http.TLSExtensions) { v.KeyShareCurves.KeyShares[0].Data[0]++ }},
		{"GREASE flag", func(v *http.TLSExtensions) { v.NotUsedGREASE = !v.NotUsedGREASE }},
		{"hello stream", func(v *http.TLSExtensions) { v.ClientHelloHexStream += "!" }},
	}
	for _, variant := range tlsVariants {
		t.Run(variant.name, func(t *testing.T) {
			copy := cloneTLSExtensions(tlsSettings)
			variant.change(copy)
			if got := tlsExtensionsHash(copy); got == tlsBaseHash {
				t.Fatalf("TLS hash did not change after changing %s", variant.name)
			}
		})
	}

	h2Settings := &http.HTTP2Settings{
		Settings:       []http.HTTP2Setting{{ID: http.HTTP2SettingHeaderTableSize, Val: 4096}},
		ConnectionFlow: 65535,
		HeadersID:      0,
		SettingsAck:    true,
		HeaderPriority: &http.HTTP2PriorityParam{StreamDep: 3, Weight: 7, Exclusive: true},
		PriorityFrames: []http.HTTP2PriorityFrame{{HTTP2FrameHeader: http.HTTP2FrameHeader{StreamID: 9}, HTTP2PriorityParam: http.HTTP2PriorityParam{StreamDep: 2, Weight: 5}}},
	}
	h2DefaultHash := http2SettingsHash(h2Settings)
	h2ExplicitDefault := cloneHTTP2Settings(h2Settings)
	h2ExplicitDefault.HeadersID = 1
	if http2SettingsHash(h2ExplicitDefault) != h2DefaultHash {
		t.Fatal("HTTP/2 default HeadersID should have one cache key")
	}
	h2Variants := []struct {
		name   string
		change func(*http.HTTP2Settings)
	}{
		{"settings", func(v *http.HTTP2Settings) { v.Settings[0].Val++ }},
		{"connection flow", func(v *http.HTTP2Settings) { v.ConnectionFlow++ }},
		{"settings ack", func(v *http.HTTP2Settings) { v.SettingsAck = !v.SettingsAck }},
		{"header priority", func(v *http.HTTP2Settings) { v.HeaderPriority.Weight++ }},
		{"priority frame", func(v *http.HTTP2Settings) { v.PriorityFrames[0].HTTP2PriorityParam.Weight++ }},
	}
	for _, variant := range h2Variants {
		t.Run("HTTP/2 "+variant.name, func(t *testing.T) {
			copy := cloneHTTP2Settings(h2Settings)
			variant.change(copy)
			if got := http2SettingsHash(copy); got == h2DefaultHash {
				t.Fatalf("HTTP/2 hash did not change after changing %s", variant.name)
			}
		})
	}
}

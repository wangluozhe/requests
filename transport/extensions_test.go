package transport

import (
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"

	utls "github.com/refraction-networking/utls"
)

func TestToTLSExtensionsCheckedTrustAnchors(t *testing.T) {
	// 255 * (1 + 255) + (1 + 252) = 65533, the maximum encoded list size.
	maxList := make([][]byte, 256)
	for i := 0; i < 255; i++ {
		maxList[i] = make([]byte, 255)
	}
	maxList[255] = make([]byte, 252)
	oversizedList := append([][]byte(nil), maxList...)
	oversizedList[255] = make([]byte, 253)

	for _, test := range []struct {
		name    string
		ids     [][]byte
		wantErr bool
	}{
		{"omitted", nil, false},
		{"empty list", [][]byte{}, false},
		{"minimum ID", [][]byte{{1}}, false},
		{"maximum ID", [][]byte{make([]byte, 255)}, false},
		{"maximum list", maxList, false},
		{"empty ID", [][]byte{{}}, true},
		{"nil ID", [][]byte{nil}, true},
		{"oversized ID", [][]byte{make([]byte, 256)}, true},
		{"oversized list", oversizedList, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			extensions, err := ToTLSExtensionsChecked(&Extensions{TrustAnchors: test.ids})
			if test.wantErr {
				if err == nil || extensions != nil {
					t.Fatalf("invalid TrustAnchors returned extensions=%v, err=%v", extensions, err)
				}
				if !strings.Contains(err.Error(), "TrustAnchors") {
					t.Fatalf("error does not identify TrustAnchors: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			ext := extensions.TrustAnchors
			if test.ids == nil {
				if ext != nil {
					t.Fatal("omitted TrustAnchors should keep chttp defaults")
				}
				return
			}
			if ext == nil || !reflect.DeepEqual(ext.TrustAnchorIDs, test.ids) {
				t.Fatalf("TrustAnchors was not fully converted: %#v", ext)
			}
			wire := make([]byte, ext.Len())
			if n, err := ext.Read(wire); n != len(wire) || err != io.EOF {
				t.Fatalf("uTLS rejected valid TrustAnchors: n=%d, err=%v", n, err)
			}
		})
	}
}

func TestToTLSExtensionsCheckedTrustAnchorsOwnsData(t *testing.T) {
	config := &Extensions{TrustAnchors: [][]byte{{1, 2}, {3, 4}}}
	extensions, err := ToTLSExtensionsChecked(config)
	if err != nil {
		t.Fatal(err)
	}
	config.TrustAnchors[0][0] = 9
	config.TrustAnchors[1] = []byte{8}
	if !reflect.DeepEqual(extensions.TrustAnchors.TrustAnchorIDs, [][]byte{{1, 2}, {3, 4}}) {
		t.Fatal("converted TrustAnchors shares nested data with input")
	}
	extensions.TrustAnchors.TrustAnchorIDs[0][1] = 7
	extensions.TrustAnchors.TrustAnchorIDs[1] = []byte{6}
	if !reflect.DeepEqual(config.TrustAnchors, [][]byte{{9, 2}, {8}}) {
		t.Fatal("input changed through converted TrustAnchors")
	}
}

func TestToTLSExtensionsCheckedRecordSizeLimit(t *testing.T) {
	for _, limit := range []int{-1, 63, 16385} {
		t.Run("invalid_"+itoa(limit), func(t *testing.T) {
			_, err := ToTLSExtensionsChecked(&Extensions{RecordSizeLimit: limit})
			if err == nil {
				t.Fatalf("RecordSizeLimit %d should be rejected", limit)
			}
		})
	}

	for _, limit := range []int{64, 16384} {
		t.Run("valid_"+itoa(limit), func(t *testing.T) {
			extensions, err := ToTLSExtensionsChecked(&Extensions{RecordSizeLimit: limit})
			if err != nil {
				t.Fatalf("RecordSizeLimit %d returned error: %v", limit, err)
			}
			if extensions.RecordSizeLimit == nil || extensions.RecordSizeLimit.Limit != uint16(limit) {
				t.Fatalf("converted RecordSizeLimit = %#v, want %d", extensions.RecordSizeLimit, limit)
			}
		})
	}
}

func TestToTLSExtensionsCheckedRejectsInvalidSignatureNumbers(t *testing.T) {
	for _, value := range []string{"", "-1", "0", "0x10000", "not-a-number"} {
		t.Run(valueOrEmpty(value), func(t *testing.T) {
			_, err := ToTLSExtensionsChecked(&Extensions{SupportedSignatureAlgorithms: []string{value}})
			if err == nil {
				t.Fatalf("signature algorithm %q should be rejected", value)
			}
		})
	}

	extensions, err := ToTLSExtensionsChecked(&Extensions{SupportedSignatureAlgorithms: []string{"0x0401"}})
	if err != nil {
		t.Fatalf("valid signature algorithm returned error: %v", err)
	}
	if len(extensions.SupportedSignatureAlgorithms.SupportedSignatureAlgorithms) != 1 ||
		extensions.SupportedSignatureAlgorithms.SupportedSignatureAlgorithms[0] != utls.SignatureScheme(0x0401) {
		t.Fatalf("converted signature algorithms = %#v", extensions.SupportedSignatureAlgorithms)
	}
}

func TestToTLSExtensionsCheckedRejectsUnknownEnums(t *testing.T) {
	tests := []struct {
		name string
		set  func(*Extensions)
	}{
		{"certificate compression", func(e *Extensions) { e.CertCompressionAlgo = []string{"unknown"} }},
		{"supported version", func(e *Extensions) { e.SupportedVersions = []string{"1.4"} }},
		{"PSK mode", func(e *Extensions) { e.PSKKeyExchangeModes = []string{"unknown"} }},
		{"key share curve", func(e *Extensions) { e.KeyShareCurves = []string{"unknown"} }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &Extensions{}
			test.set(config)
			if _, err := ToTLSExtensionsChecked(config); err == nil {
				t.Fatalf("unknown %s should be rejected", test.name)
			}
		})
	}
}

func TestToTLSExtensionsCheckedConvertsAllSupportedFields(t *testing.T) {
	config := &Extensions{
		SupportedSignatureAlgorithms: []string{"PKCS1WithSHA256"},
		CertCompressionAlgo:          []string{"zlib"},
		RecordSizeLimit:              4096,
		DelegatedCredentials:         []string{"0x0403"},
		SupportedVersions:            []string{"1.3"},
		PSKKeyExchangeModes:          []string{"PskModeDHE"},
		SignatureAlgorithmsCert:      []string{"ECDSAWithP256AndSHA256"},
		KeyShareCurves:               []string{"P256"},
		NotUsedGREASE:                true,
	}

	extensions, err := ToTLSExtensionsChecked(config)
	if err != nil {
		t.Fatalf("supported configuration returned error: %v", err)
	}
	if len(extensions.SupportedSignatureAlgorithms.SupportedSignatureAlgorithms) != 1 ||
		extensions.CertCompressionAlgo.Algorithms[0] != utls.CertCompressionZlib ||
		extensions.RecordSizeLimit.Limit != 4096 ||
		len(extensions.DelegatedCredentials.SupportedSignatureAlgorithms) != 1 ||
		len(extensions.SupportedVersions.Versions) != 1 ||
		len(extensions.PSKKeyExchangeModes.Modes) != 1 ||
		len(extensions.SignatureAlgorithmsCert.SupportedSignatureAlgorithms) != 1 ||
		len(extensions.KeyShareCurves.KeyShares) != 1 ||
		!extensions.NotUsedGREASE {
		t.Fatalf("supported configuration was not fully converted: %#v", extensions)
	}
}

func itoa(value int) string {
	return strconv.Itoa(value)
}

func valueOrEmpty(value string) string {
	if value == "" {
		return "empty"
	}
	return value
}

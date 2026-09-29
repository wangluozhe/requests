package transport

import (
	"fmt"
	"strconv"
	"strings"

	utls "github.com/refraction-networking/utls"
	http "github.com/wangluozhe/chttp"
)

var supportedSignatureAlgorithmsExtensions = map[string]utls.SignatureScheme{
	"PKCS1WithSHA256":                     utls.PKCS1WithSHA256,
	"PKCS1WithSHA384":                     utls.PKCS1WithSHA384,
	"PKCS1WithSHA512":                     utls.PKCS1WithSHA512,
	"PSSWithSHA256":                       utls.PSSWithSHA256,
	"PSSWithSHA384":                       utls.PSSWithSHA384,
	"PSSWithSHA512":                       utls.PSSWithSHA512,
	"ECDSAWithP256AndSHA256":              utls.ECDSAWithP256AndSHA256,
	"ECDSAWithP384AndSHA384":              utls.ECDSAWithP384AndSHA384,
	"ECDSAWithP521AndSHA512":              utls.ECDSAWithP521AndSHA512,
	"Ed25519":                             utls.Ed25519,
	"PKCS1WithSHA1":                       utls.PKCS1WithSHA1,
	"ECDSAWithSHA1":                       utls.ECDSAWithSHA1,
	"rsa_pkcs1_sha1":                      utls.SignatureScheme(0x0201),
	"Reserved for backward compatibility": utls.SignatureScheme(0x0202),
	"ecdsa_sha1":                          utls.SignatureScheme(0x0203),
	"rsa_pkcs1_sha256":                    utls.SignatureScheme(0x0401),
	"ecdsa_secp256r1_sha256":              utls.SignatureScheme(0x0403),
	"rsa_pkcs1_sha256_legacy":             utls.SignatureScheme(0x0420),
	"rsa_pkcs1_sha384":                    utls.SignatureScheme(0x0501),
	"ecdsa_secp384r1_sha384":              utls.SignatureScheme(0x0503),
	"rsa_pkcs1_sha384_legacy":             utls.SignatureScheme(0x0520),
	"rsa_pkcs1_sha512":                    utls.SignatureScheme(0x0601),
	"ecdsa_secp521r1_sha512":              utls.SignatureScheme(0x0603),
	"rsa_pkcs1_sha512_legacy":             utls.SignatureScheme(0x0620),
	"eccsi_sha256":                        utls.SignatureScheme(0x0704),
	"iso_ibs1":                            utls.SignatureScheme(0x0705),
	"iso_ibs2":                            utls.SignatureScheme(0x0706),
	"iso_chinese_ibs":                     utls.SignatureScheme(0x0707),
	"sm2sig_sm3":                          utls.SignatureScheme(0x0708),
	"gostr34102012_256a":                  utls.SignatureScheme(0x0709),
	"gostr34102012_256b":                  utls.SignatureScheme(0x070A),
	"gostr34102012_256c":                  utls.SignatureScheme(0x070B),
	"gostr34102012_256d":                  utls.SignatureScheme(0x070C),
	"gostr34102012_512a":                  utls.SignatureScheme(0x070D),
	"gostr34102012_512b":                  utls.SignatureScheme(0x070E),
	"gostr34102012_512c":                  utls.SignatureScheme(0x070F),
	"rsa_pss_rsae_sha256":                 utls.SignatureScheme(0x0804),
	"rsa_pss_rsae_sha384":                 utls.SignatureScheme(0x0805),
	"rsa_pss_rsae_sha512":                 utls.SignatureScheme(0x0806),
	"ed25519":                             utls.SignatureScheme(0x0807),
	"ed448":                               utls.SignatureScheme(0x0808),
	"rsa_pss_pss_sha256":                  utls.SignatureScheme(0x0809),
	"rsa_pss_pss_sha384":                  utls.SignatureScheme(0x080A),
	"rsa_pss_pss_sha512":                  utls.SignatureScheme(0x080B),
	"ecdsa_brainpoolP256r1tls13_sha256":   utls.SignatureScheme(0x081A),
	"ecdsa_brainpoolP384r1tls13_sha384":   utls.SignatureScheme(0x081B),
	"ecdsa_brainpoolP512r1tls13_sha512":   utls.SignatureScheme(0x081C),
}

var certCompressionAlgoExtensions = map[string]utls.CertCompressionAlgo{
	"zlib":   utls.CertCompressionZlib,
	"brotli": utls.CertCompressionBrotli,
	"zstd":   utls.CertCompressionZstd,
}

var supportedVersionsExtensions = map[string]uint16{
	"GREASE": utls.GREASE_PLACEHOLDER,
	"1.3":    utls.VersionTLS13,
	"1.2":    utls.VersionTLS12,
	"1.1":    utls.VersionTLS11,
	"1.0":    utls.VersionTLS10,
}

var pskKeyExchangeModesExtensions = map[string]uint8{
	"PskModeDHE":   utls.PskModeDHE,
	"PskModePlain": utls.PskModePlain,
}

var keyShareCurvesExtensions = map[string]utls.KeyShare{
	"GREASE": utls.KeyShare{Group: utls.CurveID(utls.GREASE_PLACEHOLDER), Data: []byte{0}},
	"P256":   utls.KeyShare{Group: utls.CurveP256},
	"P384":   utls.KeyShare{Group: utls.CurveP384},
	"P521":   utls.KeyShare{Group: utls.CurveP521},
	"X25519": utls.KeyShare{Group: utls.X25519},
	"4588":   utls.KeyShare{Group: utls.X25519MLKEM768},
	"25497":  utls.KeyShare{Group: utls.CurveID(0x6399)},
	"65072":  utls.KeyShare{Group: utls.CurveID(0xfe30)},
	"65073":  utls.KeyShare{Group: utls.CurveID(0xfe31)},
	"65074":  utls.KeyShare{Group: utls.CurveID(0xfe32)},
}

type Extensions struct {
	//PKCS1WithSHA256 SignatureScheme = 0x0401
	//PKCS1WithSHA384 SignatureScheme = 0x0501
	//PKCS1WithSHA512 SignatureScheme = 0x0601
	//PSSWithSHA256 SignatureScheme = 0x0804
	//PSSWithSHA384 SignatureScheme = 0x0805
	//PSSWithSHA512 SignatureScheme = 0x0806
	//ECDSAWithP256AndSHA256 SignatureScheme = 0x0403
	//ECDSAWithP384AndSHA384 SignatureScheme = 0x0503
	//ECDSAWithP521AndSHA512 SignatureScheme = 0x0603
	//Ed25519 SignatureScheme = 0x0807
	//PKCS1WithSHA1 SignatureScheme = 0x0201
	//ECDSAWithSHA1 SignatureScheme = 0x0203
	SupportedSignatureAlgorithms []string `json:"SupportedSignatureAlgorithms"`
	//CertCompressionZlib   CertCompressionAlgo = 0x0001
	//CertCompressionBrotli CertCompressionAlgo = 0x0002
	//CertCompressionZstd   CertCompressionAlgo = 0x0003
	CertCompressionAlgo []string `json:"CertCompressionAlgo"`
	// Limit: 0x4001
	RecordSizeLimit int `json:"RecordSizeLimit"`
	//PKCS1WithSHA256 SignatureScheme = 0x0401
	//PKCS1WithSHA384 SignatureScheme = 0x0501
	//PKCS1WithSHA512 SignatureScheme = 0x0601
	//PSSWithSHA256 SignatureScheme = 0x0804
	//PSSWithSHA384 SignatureScheme = 0x0805
	//PSSWithSHA512 SignatureScheme = 0x0806
	//ECDSAWithP256AndSHA256 SignatureScheme = 0x0403
	//ECDSAWithP384AndSHA384 SignatureScheme = 0x0503
	//ECDSAWithP521AndSHA512 SignatureScheme = 0x0603
	//Ed25519 SignatureScheme = 0x0807
	//PKCS1WithSHA1 SignatureScheme = 0x0201
	//ECDSAWithSHA1 SignatureScheme = 0x0203
	DelegatedCredentials []string `json:"DelegatedCredentials"`
	//GREASE_PLACEHOLDER = 0x0a0a
	//VersionTLS10 = 0x0301
	//VersionTLS11 = 0x0302
	//VersionTLS12 = 0x0303
	//VersionTLS13 = 0x0304
	//VersionSSL30 = 0x0300
	SupportedVersions []string `json:"SupportedVersions"`
	//PskModePlain uint8 = pskModePlain
	//PskModeDHE   uint8 = pskModeDHE
	PSKKeyExchangeModes []string `json:"PSKKeyExchangeModes"`
	//PKCS1WithSHA256 SignatureScheme = 0x0401
	//PKCS1WithSHA384 SignatureScheme = 0x0501
	//PKCS1WithSHA512 SignatureScheme = 0x0601
	//PSSWithSHA256 SignatureScheme = 0x0804
	//PSSWithSHA384 SignatureScheme = 0x0805
	//PSSWithSHA512 SignatureScheme = 0x0806
	//ECDSAWithP256AndSHA256 SignatureScheme = 0x0403
	//ECDSAWithP384AndSHA384 SignatureScheme = 0x0503
	//ECDSAWithP521AndSHA512 SignatureScheme = 0x0603
	//Ed25519 SignatureScheme = 0x0807
	//PKCS1WithSHA1 SignatureScheme = 0x0201
	//ECDSAWithSHA1 SignatureScheme = 0x0203
	SignatureAlgorithmsCert []string `json:"SignatureAlgorithmsCert"`
	//GREASE_PLACEHOLDER = 0x0a0a
	//CurveP256 CurveID = 23
	//CurveP384 CurveID = 24
	//CurveP521 CurveID = 25
	//X25519    CurveID = 29
	KeyShareCurves []string `json:"KeyShareCurves"`
	//default is false, default is used grease, if not used grease the NotUsedGREASE param is true
	NotUsedGREASE        bool   `json:"NotUsedGREASE"`
	ClientHelloHexStream string `json:"ClientHelloHexStream"`
}

func parseSignatureScheme(value string) (utls.SignatureScheme, error) {
	if parsed, ok := supportedSignatureAlgorithmsExtensions[value]; ok {
		return parsed, nil
	}
	raw := strings.TrimPrefix(strings.TrimPrefix(value, "0x"), "0X")
	parsed, err := strconv.ParseUint(raw, 16, 16)
	if err != nil || parsed == 0 {
		return 0, fmt.Errorf("unknown signature algorithm %q", value)
	}
	return utls.SignatureScheme(parsed), nil
}

func parseKeyShareCurve(value string) (utls.KeyShare, error) {
	if parsed, ok := keyShareCurvesExtensions[value]; ok {
		return parsed, nil
	}
	curveID, err := strconv.ParseUint(value, 10, 16)
	if err != nil || curveID == 0 {
		return utls.KeyShare{}, fmt.Errorf("unknown key share curve %q", value)
	}
	return utls.KeyShare{Group: utls.CurveID(curveID), Data: []byte{0}}, nil
}

func appendSignatureAlgorithms(dst *[]utls.SignatureScheme, values []string, field string) error {
	for index, value := range values {
		parsed, err := parseSignatureScheme(value)
		if err != nil {
			return fmt.Errorf("invalid %s[%d]: %w", field, index, err)
		}
		*dst = append(*dst, parsed)
	}
	return nil
}

// ToTLSExtensionsChecked 将 TLS 扩展配置转换为经过校验的运行时配置。
func ToTLSExtensionsChecked(e *Extensions) (*http.TLSExtensions, error) {
	extensions := &http.TLSExtensions{}
	if e == nil {
		return extensions, nil
	}
	if e.ClientHelloHexStream != "" {
		extensions.ClientHelloHexStream = e.ClientHelloHexStream
		return extensions, nil
	}
	if e.RecordSizeLimit != 0 && (e.RecordSizeLimit < 64 || e.RecordSizeLimit > 16384) {
		return nil, fmt.Errorf("invalid RecordSizeLimit %d: must be between 64 and 16384", e.RecordSizeLimit)
	}
	if e.SupportedSignatureAlgorithms != nil {
		extensions.SupportedSignatureAlgorithms = &utls.SignatureAlgorithmsExtension{}
		if err := appendSignatureAlgorithms(&extensions.SupportedSignatureAlgorithms.SupportedSignatureAlgorithms, e.SupportedSignatureAlgorithms, "SupportedSignatureAlgorithms"); err != nil {
			return nil, err
		}
	}
	if e.CertCompressionAlgo != nil {
		extensions.CertCompressionAlgo = &utls.UtlsCompressCertExtension{}
		for index, value := range e.CertCompressionAlgo {
			parsed, ok := certCompressionAlgoExtensions[value]
			if !ok {
				return nil, fmt.Errorf("invalid CertCompressionAlgo[%d]: unknown value %q", index, value)
			}
			extensions.CertCompressionAlgo.Algorithms = append(extensions.CertCompressionAlgo.Algorithms, parsed)
		}
	}
	if e.RecordSizeLimit != 0 {
		extensions.RecordSizeLimit = &utls.FakeRecordSizeLimitExtension{Limit: uint16(e.RecordSizeLimit)}
	}
	if e.DelegatedCredentials != nil {
		extensions.DelegatedCredentials = &utls.DelegatedCredentialsExtension{}
		if err := appendSignatureAlgorithms(&extensions.DelegatedCredentials.SupportedSignatureAlgorithms, e.DelegatedCredentials, "DelegatedCredentials"); err != nil {
			return nil, err
		}
	}
	if e.SupportedVersions != nil {
		extensions.SupportedVersions = &utls.SupportedVersionsExtension{}
		for index, value := range e.SupportedVersions {
			parsed, ok := supportedVersionsExtensions[value]
			if !ok {
				return nil, fmt.Errorf("invalid SupportedVersions[%d]: unknown value %q", index, value)
			}
			extensions.SupportedVersions.Versions = append(extensions.SupportedVersions.Versions, parsed)
		}
	}
	if e.PSKKeyExchangeModes != nil {
		extensions.PSKKeyExchangeModes = &utls.PSKKeyExchangeModesExtension{}
		for index, value := range e.PSKKeyExchangeModes {
			parsed, ok := pskKeyExchangeModesExtensions[value]
			if !ok {
				return nil, fmt.Errorf("invalid PSKKeyExchangeModes[%d]: unknown value %q", index, value)
			}
			extensions.PSKKeyExchangeModes.Modes = append(extensions.PSKKeyExchangeModes.Modes, parsed)
		}
	}
	if e.SignatureAlgorithmsCert != nil {
		extensions.SignatureAlgorithmsCert = &utls.SignatureAlgorithmsCertExtension{}
		if err := appendSignatureAlgorithms(&extensions.SignatureAlgorithmsCert.SupportedSignatureAlgorithms, e.SignatureAlgorithmsCert, "SignatureAlgorithmsCert"); err != nil {
			return nil, err
		}
	}
	if e.KeyShareCurves != nil {
		extensions.KeyShareCurves = &utls.KeyShareExtension{}
		for index, value := range e.KeyShareCurves {
			parsed, err := parseKeyShareCurve(value)
			if err != nil {
				return nil, fmt.Errorf("invalid KeyShareCurves[%d]: %w", index, err)
			}
			extensions.KeyShareCurves.KeyShares = append(extensions.KeyShareCurves.KeyShares, parsed)
		}
	}
	extensions.NotUsedGREASE = e.NotUsedGREASE
	return extensions, nil
}

// ToTLSExtensions 保留原有 API，非法输入时返回空扩展配置。
func ToTLSExtensions(e *Extensions) *http.TLSExtensions {
	extensions, err := ToTLSExtensionsChecked(e)
	if err != nil {
		return &http.TLSExtensions{}
	}
	return extensions
}

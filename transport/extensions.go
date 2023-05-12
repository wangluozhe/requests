package transport

import (
	"fmt"
	utls "github.com/refraction-networking/utls"
	"strconv"
)

var supportedSignatureAlgorithmsExtensions = map[string]utls.SignatureScheme{
	"PKCS1WithSHA256":        utls.PKCS1WithSHA256,
	"PKCS1WithSHA384":        utls.PKCS1WithSHA384,
	"PKCS1WithSHA512":        utls.PKCS1WithSHA512,
	"PSSWithSHA256":          utls.PSSWithSHA256,
	"PSSWithSHA384":          utls.PSSWithSHA384,
	"PSSWithSHA512":          utls.PSSWithSHA512,
	"ECDSAWithP256AndSHA256": utls.ECDSAWithP256AndSHA256,
	"ECDSAWithP384AndSHA384": utls.ECDSAWithP384AndSHA384,
	"ECDSAWithP521AndSHA512": utls.ECDSAWithP521AndSHA512,
	"Ed25519":                utls.Ed25519,
	"PKCS1WithSHA1":          utls.PKCS1WithSHA1,
	"ECDSAWithSHA1":          utls.ECDSAWithSHA1,
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
	SupportedSignatureAlgorithms []string
	//CertCompressionZlib   CertCompressionAlgo = 0x0001
	//CertCompressionBrotli CertCompressionAlgo = 0x0002
	//CertCompressionZstd   CertCompressionAlgo = 0x0003
	CertCompressionAlgo []string
	// Limit: 0x4001
	RecordSizeLimit int
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
	DelegatedCredentials []string
	//GREASE_PLACEHOLDER = 0x0a0a
	//VersionTLS10 = 0x0301
	//VersionTLS11 = 0x0302
	//VersionTLS12 = 0x0303
	//VersionTLS13 = 0x0304
	//VersionSSL30 = 0x0300
	SupportedVersions []string
	//PskModePlain uint8 = pskModePlain
	//PskModeDHE   uint8 = pskModeDHE
	PSKKeyExchangeModes []string
	//GREASE_PLACEHOLDER = 0x0a0a
	//CurveP256 CurveID = 23
	//CurveP384 CurveID = 24
	//CurveP521 CurveID = 25
	//X25519    CurveID = 29
	KeyShareCurves []string
}

type TLSExtensions struct {
	SupportedSignatureAlgorithms *utls.SignatureAlgorithmsExtension
	CertCompressionAlgo          *utls.UtlsCompressCertExtension
	RecordSizeLimit              *utls.FakeRecordSizeLimitExtension
	DelegatedCredentials         *utls.DelegatedCredentialsExtension
	SupportedVersions            *utls.SupportedVersionsExtension
	PSKKeyExchangeModes          *utls.PSKKeyExchangeModesExtension
	KeyShareCurves               *utls.KeyShareExtension
}

func ToTLSExtensions(e *Extensions) (extensions *TLSExtensions) {
	extensions = &TLSExtensions{}
	if e == nil {
		return extensions
	}
	if e.SupportedSignatureAlgorithms != nil {
		extensions.SupportedSignatureAlgorithms = &utls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{}}
		for _, s := range e.SupportedSignatureAlgorithms {
			extensions.SupportedSignatureAlgorithms.SupportedSignatureAlgorithms = append(extensions.SupportedSignatureAlgorithms.SupportedSignatureAlgorithms, supportedSignatureAlgorithmsExtensions[s])
		}
	}
	if e.CertCompressionAlgo != nil {
		extensions.CertCompressionAlgo = &utls.UtlsCompressCertExtension{Algorithms: []utls.CertCompressionAlgo{}}
		for _, s := range e.CertCompressionAlgo {
			extensions.CertCompressionAlgo.Algorithms = append(extensions.CertCompressionAlgo.Algorithms, certCompressionAlgoExtensions[s])
		}
	}
	if e.RecordSizeLimit != 0 {
		hexStr := fmt.Sprintf("0x%v", e.RecordSizeLimit)
		hexInt, _ := strconv.ParseInt(hexStr, 0, 0)
		extensions.RecordSizeLimit = &utls.FakeRecordSizeLimitExtension{uint16(hexInt)}
	}
	if e.DelegatedCredentials != nil {
		extensions.DelegatedCredentials = &utls.DelegatedCredentialsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{}}
		for _, s := range e.DelegatedCredentials {
			extensions.DelegatedCredentials.SupportedSignatureAlgorithms = append(extensions.DelegatedCredentials.SupportedSignatureAlgorithms, supportedSignatureAlgorithmsExtensions[s])
		}
	}
	if e.SupportedVersions != nil {
		extensions.SupportedVersions = &utls.SupportedVersionsExtension{Versions: []uint16{}}
		for _, s := range e.SupportedVersions {
			extensions.SupportedVersions.Versions = append(extensions.SupportedVersions.Versions, supportedVersionsExtensions[s])
		}
	}
	if e.PSKKeyExchangeModes != nil {
		extensions.PSKKeyExchangeModes = &utls.PSKKeyExchangeModesExtension{Modes: []uint8{}}
		for _, s := range e.PSKKeyExchangeModes {
			extensions.PSKKeyExchangeModes.Modes = append(extensions.PSKKeyExchangeModes.Modes, pskKeyExchangeModesExtensions[s])
		}
	}
	if e.KeyShareCurves != nil {
		extensions.KeyShareCurves = &utls.KeyShareExtension{KeyShares: []utls.KeyShare{}}
		for _, s := range e.KeyShareCurves {
			extensions.KeyShareCurves.KeyShares = append(extensions.KeyShareCurves.KeyShares, keyShareCurvesExtensions[s])
		}
	}
	return extensions
}

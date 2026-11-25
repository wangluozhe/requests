package requests

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/iotest"
	"time"

	chttp "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests/models"
	"github.com/wangluozhe/requests/url"
)

func deflateTestData(t *testing.T, format string, content []byte, level int) []byte {
	t.Helper()
	var buffer bytes.Buffer
	var writer io.WriteCloser
	var err error
	if format == "zlib" {
		writer, err = zlib.NewWriterLevel(&buffer, level)
	} else {
		writer, err = flate.NewWriter(&buffer, level)
	}
	if err != nil {
		t.Fatal(err)
	}
	if _, err = writer.Write(content); err != nil {
		t.Fatal(err)
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestDecompressDeflateFormats(t *testing.T) {
	for _, format := range []string{"zlib", "raw"} {
		for _, test := range []struct {
			name    string
			content []byte
		}{
			{name: "empty"},
			{name: "short", content: []byte("a")},
			{name: "text", content: []byte("标准 deflate 响应\x00\xff")},
		} {
			t.Run(format+"/"+test.name, func(t *testing.T) {
				compressed := deflateTestData(t, format, test.content, flate.DefaultCompression)
				// 单字节读取同时覆盖头部分片与解压过程中的分片。
				body := io.NopCloser(iotest.OneByteReader(bytes.NewReader(compressed)))
				reader, err := DecompressBody(body, "DeFlAtE")
				if err != nil {
					t.Fatal(err)
				}
				defer reader.Close()
				got, err := io.ReadAll(reader)
				if err != nil || !bytes.Equal(got, test.content) {
					t.Fatalf("decoded=%q, err=%v; want %q", got, err, test.content)
				}
			})
		}
	}
}

func TestDecompressDeflateErrors(t *testing.T) {
	zlibData := deflateTestData(t, "zlib", []byte("payload"), flate.DefaultCompression)
	rawData := deflateTestData(t, "raw", []byte("payload"), flate.DefaultCompression)
	badChecksum := bytes.Clone(zlibData)
	badChecksum[len(badChecksum)-1] ^= 1
	readErr := errors.New("source read failed")
	for _, test := range []struct {
		name   string
		source io.Reader
		want   error
	}{
		{name: "empty wire body", source: bytes.NewReader(nil), want: io.ErrUnexpectedEOF},
		{name: "short header", source: bytes.NewReader(zlibData[:1]), want: io.ErrUnexpectedEOF},
		{name: "zlib header only", source: bytes.NewReader(zlibData[:2]), want: io.ErrUnexpectedEOF},
		{name: "truncated zlib", source: bytes.NewReader(zlibData[:len(zlibData)-1]), want: io.ErrUnexpectedEOF},
		{name: "zlib checksum", source: bytes.NewReader(badChecksum), want: zlib.ErrChecksum},
		{name: "truncated raw", source: bytes.NewReader(rawData[:len(rawData)-1]), want: io.ErrUnexpectedEOF},
		{name: "source error", source: iotest.ErrReader(readErr), want: readErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			reader, err := DecompressBody(io.NopCloser(test.source), "deflate")
			if err == nil {
				defer reader.Close()
				_, err = io.ReadAll(reader)
			}
			if !errors.Is(err, test.want) {
				t.Fatalf("err=%v, want %v", err, test.want)
			}
		})
	}
}

func TestSessionDeflateResponses(t *testing.T) {
	content := []byte("标准 deflate 与原始 DEFLATE 均可读取")
	for _, format := range []string{"zlib", "raw"} {
		t.Run(format, func(t *testing.T) {
			compressed := deflateTestData(t, format, content, flate.DefaultCompression)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Encoding", "deflate")
				_, _ = w.Write(compressed)
			}))
			defer server.Close()
			session := NewSession()
			defer session.Close()
			session.Timeout = 5 * time.Second
			for _, test := range []struct {
				name      string
				retention url.ResponseRetention
				stream    bool
			}{
				{name: "default", retention: url.ResponseRetentionDefault},
				{name: "content", retention: url.ResponseRetentionContent},
				{name: "body", retention: url.ResponseRetentionBody},
				{name: "all", retention: url.ResponseRetentionAll},
				{name: "stream", stream: true},
			} {
				t.Run(test.name, func(t *testing.T) {
					response, err := session.Get(server.URL, &url.Request{ResponseRetention: test.retention, Stream: test.stream})
					if err != nil {
						t.Fatal(err)
					}
					defer response.Close()
					if test.stream || test.retention == url.ResponseRetentionBody {
						if response.Content != nil {
							t.Fatal("streaming response buffered Content")
						}
					} else if !bytes.Equal(response.Content, content) {
						t.Fatalf("Content=%q, want %q", response.Content, content)
					}
					if test.retention == url.ResponseRetentionContent {
						if response.Body != nil {
							t.Fatal("content-only response has a Body")
						}
					} else {
						if response.Body == nil {
							t.Fatal("missing Body")
						}
						want := content
						if test.stream {
							want = compressed // Stream 保持返回原始压缩字节的既有语义。
						}
						got, err := io.ReadAll(response.Body)
						if err != nil || !bytes.Equal(got, want) {
							t.Fatalf("Body=%q, err=%v; want %q", got, err, want)
						}
					}
					if test.retention == url.ResponseRetentionAll {
						_, rawBody, found := bytes.Cut(response.RawResponse, []byte("\r\n\r\n"))
						if !found || !bytes.Equal(rawBody, compressed) {
							t.Fatalf("RawResponse body=%q, want original compressed bytes", rawBody)
						}
					}
				})
			}
		})
	}
}

func TestBuildResponseDeflateStreamingOwnership(t *testing.T) {
	content := bytes.Repeat([]byte("streaming content"), 16384)
	for _, format := range []string{"zlib", "raw"} {
		t.Run(format, func(t *testing.T) {
			// 大于解码器缓冲区的压缩流，用于验证构造响应时不会读完整个 Body。
			source := bytes.NewReader(deflateTestData(t, format, content, flate.NoCompression))
			body := &retentionTrackingBody{Reader: source}
			session := NewSession()
			defer session.Close()
			response, err := session.buildResponse(&chttp.Response{
				StatusCode: 200, Header: chttp.Header{"Content-Encoding": {"deflate"}}, Body: body,
			}, &models.PrepareRequest{Url: "http://example.test/"}, &url.Request{ResponseRetention: url.ResponseRetentionBody})
			if err != nil {
				t.Fatal(err)
			}
			defer response.Close()
			if source.Len() == 0 || body.closes != 0 || response.Content != nil || response.RawResponse != nil {
				t.Fatal("streaming response was buffered or closed before return")
			}
			prefix := make([]byte, 8)
			if _, err := io.ReadFull(response.Body, prefix); err != nil || !bytes.Equal(prefix, content[:8]) {
				t.Fatalf("prefix=%q, err=%v", prefix, err)
			}
			_ = response.Close()
			if body.closes != 1 {
				t.Fatalf("Body closes=%d, want one close by caller", body.closes)
			}
		})
	}
}

func TestBuildResponseDeflateErrorsCloseBody(t *testing.T) {
	compressed := deflateTestData(t, "zlib", []byte("payload"), flate.DefaultCompression)
	badChecksum := bytes.Clone(compressed)
	badChecksum[len(badChecksum)-1] ^= 1
	for _, test := range []struct {
		name string
		data []byte
		want error
	}{
		{name: "incomplete header", data: []byte{0x78, 0x20}, want: io.ErrUnexpectedEOF}, // FDICT 缺少字典标识。
		{name: "checksum", data: badChecksum, want: zlib.ErrChecksum},
	} {
		for _, retention := range []url.ResponseRetention{
			url.ResponseRetentionDefault, url.ResponseRetentionContent, url.ResponseRetentionAll, url.ResponseRetentionBody,
		} {
			t.Run(test.name+"/"+string(retention), func(t *testing.T) {
				body := &retentionTrackingBody{Reader: bytes.NewReader(test.data)}
				session := NewSession()
				defer session.Close()
				response, err := session.buildResponse(&chttp.Response{
					StatusCode: 200, ProtoMajor: 1, ProtoMinor: 1,
					Header: chttp.Header{"Content-Encoding": {"deflate"}}, Body: body,
				}, &models.PrepareRequest{Url: "http://example.test/"}, &url.Request{ResponseRetention: retention})
				if response != nil {
					defer response.Close()
					if retention != url.ResponseRetentionBody || err != nil {
						t.Fatalf("unexpected response for corrupt data: response=%v, err=%v", response, err)
					}
					_, err = io.ReadAll(response.Body)
					_ = response.Close()
				}
				if !errors.Is(err, test.want) || body.closes != 1 {
					t.Fatalf("err=%v, closes=%d; want %v and one close", err, body.closes, test.want)
				}
			})
		}
	}
}

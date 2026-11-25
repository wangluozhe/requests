package requests

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	chttp "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests/models"
	"github.com/wangluozhe/requests/url"
)

func TestSessionValidatesRetentionBeforeSending(t *testing.T) {
	for _, test := range []struct {
		name    string
		session url.ResponseRetention
		request *url.Request
		wantErr bool
	}{
		{name: "invalid session", session: "invalid", wantErr: true},
		{name: "invalid request", session: url.ResponseRetentionDefault, request: &url.Request{ResponseRetention: "invalid"}, wantErr: true},
		{name: "inherited invalid session", session: "invalid", request: &url.Request{}, wantErr: true},
		{name: "invalid streaming request", request: &url.Request{ResponseRetention: "invalid", Stream: true}, wantErr: true},
		{name: "default"},
		{name: "valid request overrides invalid session", session: "invalid", request: &url.Request{ResponseRetention: url.ResponseRetentionDefault}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				_, _ = io.WriteString(w, "ok")
			}))
			defer server.Close()
			session := NewSession()
			defer session.Close()
			session.Timeout = 5 * time.Second
			session.ResponseRetention = test.session
			response, err := session.Post(server.URL, test.request)
			if response != nil {
				defer response.Close()
			}
			if test.wantErr {
				if response != nil || err == nil || !strings.Contains(err.Error(), `invalid response retention "invalid"`) {
					t.Fatalf("response=%v, err=%v; want retention error", response, err)
				}
				if got := calls.Load(); got != 0 {
					t.Errorf("invalid retention sent %d requests", got)
				}
				if len(session.transportCache) != 0 {
					t.Error("invalid retention created a cached Transport")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if calls.Load() != 1 || response.Text() != "ok" || response.Retention != url.ResponseRetentionDefault {
				t.Fatalf("unexpected valid response: calls=%d, response=%+v", calls.Load(), response)
			}
		})
	}
}

type retentionTrackingBody struct {
	io.Reader
	reads, closes int
	closeErr      error
}

func (b *retentionTrackingBody) Read(p []byte) (int, error) {
	b.reads++
	return b.Reader.Read(p)
}

func (b *retentionTrackingBody) Close() error {
	b.closes++
	return b.closeErr
}

func TestBuildResponseInvalidRetentionClosesBody(t *testing.T) {
	for _, test := range []struct {
		name     string
		session  url.ResponseRetention
		request  *url.Request
		nilBody  bool
		closeErr error
	}{
		{name: "session", session: "invalid"},
		{name: "request", request: &url.Request{ResponseRetention: "invalid"}},
		{name: "stream", request: &url.Request{ResponseRetention: "invalid", Stream: true}},
		{name: "nil body", session: "invalid", nilBody: true},
		{name: "close error", session: "invalid", closeErr: errors.New("close failed")},
	} {
		t.Run(test.name, func(t *testing.T) {
			session := NewSession()
			defer session.Close()
			session.ResponseRetention = test.session
			body := &retentionTrackingBody{Reader: strings.NewReader("unread"), closeErr: test.closeErr}
			resp := &chttp.Response{StatusCode: 200, Header: make(chttp.Header)}
			if !test.nilBody {
				resp.Body = body
			}
			response, err := session.buildResponse(resp, &models.PrepareRequest{Url: "http://example.test/"}, test.request)
			if response != nil || err == nil || !strings.Contains(err.Error(), `invalid response retention "invalid"`) {
				t.Fatalf("response=%v, err=%v; want retention error", response, err)
			}
			if !test.nilBody && (body.closes != 1 || body.reads != 0) {
				t.Fatalf("Body closes=%d, reads=%d; want one close without reading", body.closes, body.reads)
			}
		})
	}
}

func TestBuildResponseStreamingBodyOwnership(t *testing.T) {
	for _, req := range []*url.Request{
		{Stream: true},
		{ResponseRetention: url.ResponseRetentionBody},
	} {
		session := NewSession()
		body := &retentionTrackingBody{Reader: strings.NewReader("stream")}
		response, err := session.buildResponse(&chttp.Response{
			StatusCode: 200, Header: make(chttp.Header), Body: body,
		}, &models.PrepareRequest{Url: "http://example.test/"}, req)
		if err != nil {
			t.Fatal(err)
		}
		if body.closes != 0 || body.reads != 0 {
			t.Fatalf("stream was consumed or closed before return: reads=%d, closes=%d", body.reads, body.closes)
		}
		got, err := io.ReadAll(response.Body)
		if err != nil || string(got) != "stream" {
			t.Fatalf("stream=%q, err=%v", got, err)
		}
		_ = response.Close()
		_ = session.Close()
		if body.closes != 1 {
			t.Fatalf("Body closes=%d, want one close by caller", body.closes)
		}
	}
}

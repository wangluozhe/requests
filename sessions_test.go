package requests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wangluozhe/requests/url"
)

func TestMergeSettingParamsOrderAndIsolation(t *testing.T) {
	for _, test := range []struct {
		name    string
		request *url.Params
		session *url.Params
		want    string
	}{
		{name: "both_nil"},
		{name: "request_only", request: url.ParseParams("a=new&a=two&z=4"), want: "a=new&a=two&z=4"},
		{name: "session_only", session: url.ParseParams("b=1&a=old&c=3"), want: "b=1&a=old&c=3"},
		{
			name:    "override_and_append",
			request: url.ParseParams("a=new&a=two&z=4"),
			session: url.ParseParams("b=1&a=old&c=3"),
			want:    "b=1&c=3&a=new&a=two&z=4",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			merged := merge_setting(test.request, test.session).(*url.Params)
			if test.request == nil && test.session == nil {
				if merged != nil {
					t.Fatal("merging nil Params should return nil")
				}
				return
			}
			if merged == nil || merged == test.request || merged == test.session {
				t.Fatal("merged Params must be an independent collection")
			}
			if got := merged.Encode(); got != test.want {
				t.Fatalf("merged Params = %q, want %q", got, test.want)
			}
			originals := make(map[*url.Params]string)
			for _, source := range []*url.Params{test.request, test.session} {
				if source != nil {
					originals[source] = source.Encode()
				}
			}
			merged.Add("a", "merged-only")
			merged.Del("b")
			merged.Set("z", "merged-only")
			for source, want := range originals {
				if got := source.Encode(); got != want {
					t.Fatalf("source changed through merged Params: %q, want %q", got, want)
				}
			}
			want := merged.Encode()
			for source := range originals {
				source.Set("a", "source-only")
				source.Add("z", "source-only")
			}
			if got := merged.Encode(); got != want {
				t.Fatalf("merged Params changed through source: %q, want %q", got, want)
			}
		})
	}
}

func TestMergeSettingInheritsSessionTimeout(t *testing.T) {
	const sessionTimeout = 250 * time.Millisecond
	if got := merge_setting(time.Duration(0), sessionTimeout); got != sessionTimeout {
		t.Fatalf("merged timeout = %v, want %v", got, sessionTimeout)
	}

	const requestTimeout = 50 * time.Millisecond
	if got := merge_setting(requestTimeout, sessionTimeout); got != requestTimeout {
		t.Fatalf("merged timeout = %v, want request timeout %v", got, requestTimeout)
	}
}

func TestSessionRejectsInvalidURLWithoutPanic(t *testing.T) {
	session := NewSession()
	response, err := session.Get("://bad", nil)
	if err == nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		t.Fatal("invalid URL request unexpectedly succeeded")
	}
	if response != nil {
		t.Fatalf("invalid URL returned response: %#v", response)
	}
}

func TestSessionTimeoutIsAppliedToRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte("late"))
	}))
	defer server.Close()

	session := NewSession()
	session.Timeout = 20 * time.Millisecond
	response, err := session.Get(server.URL, nil)
	if err == nil {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
		t.Fatal("request succeeded after the session timeout")
	}
	if response != nil {
		t.Fatal("timed out request returned a response")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		message := strings.ToLower(err.Error())
		if !strings.Contains(message, "deadline") && !strings.Contains(message, "timeout") {
			t.Fatalf("timeout returned unexpected error: %v", err)
		}
	}
}

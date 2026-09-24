package requests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

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

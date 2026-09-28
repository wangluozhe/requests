package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestCSessionCookiePersistsAcrossRequests(t *testing.T) {
	var mu sync.Mutex
	var received []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		received = append(received, r.Header.Get("Cookie"))
		mu.Unlock()
		if r.Header.Get("Cookie") == "" {
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "persisted", Path: "/"})
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	id := "cookie-persistence-" + time.Now().Format("150405.000000000")
	entry := acquireSession(id)
	defer func() { _ = closeSession(id) }()
	defer releaseSession(entry)

	for i := 0; i < 2; i++ {
		response, err := entry.session.Get(server.URL, nil)
		if err != nil {
			t.Fatalf("request %d failed: %v", i+1, err)
		}
		if response.Body != nil {
			_ = response.Body.Close()
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 2 || received[0] != "" || received[1] != "session=persisted" {
		t.Fatalf("cookies sent = %#v, want first request empty and second request persisted", received)
	}
}

func TestCSessionCloseWaitsForRequestAndIsIdempotent(t *testing.T) {
	id := "session-close-" + time.Now().Format("150405.000000000")
	entry := acquireSession(id)

	closed := make(chan error, 1)
	go func() {
		closed <- closeSession(id)
	}()

	select {
	case err := <-closed:
		t.Fatalf("session closed while request reference was active: %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	releaseSession(entry)
	if err := <-closed; err != nil {
		t.Fatalf("closeSession failed: %v", err)
	}
	if err := closeSession(id); err != nil {
		t.Fatalf("repeated closeSession failed: %v", err)
	}
}

type closeTrackingBody struct {
	io.Reader
	closed chan struct{}
	once   sync.Once
}

func (body *closeTrackingBody) Close() error {
	body.once.Do(func() { close(body.closed) })
	return nil
}

func TestCSessionCloseClosesOwnedStreams(t *testing.T) {
	id := "stream-close-" + time.Now().Format("150405.000000000")
	entry := sessionForID(id)
	body := &closeTrackingBody{Reader: strings.NewReader("body"), closed: make(chan struct{})}
	streamID := "stream-" + id
	streamPool.Store(streamID, &StreamEntry{Body: body, Session: entry})
	defer closeSession(id)

	if err := closeSession(id); err != nil {
		t.Fatalf("closeSession failed: %v", err)
	}
	select {
	case <-body.closed:
	case <-time.After(time.Second):
		t.Fatal("owned stream body was not closed")
	}
	if _, ok := streamPool.Load(streamID); ok {
		t.Fatal("owned stream remained in stream pool")
	}
}

func TestManagedCStringReleaseIsIdempotent(t *testing.T) {
	const id = "managed-cstring-test"
	ptr := managedCString(id, "result")
	if ptr == nil {
		t.Fatal("managedCString returned nil")
	}
	freeManagedCString(id)
	freeManagedCString(id)
	if _, ok := unsafePointers[id]; ok {
		t.Fatal("managed CString remained registered after release")
	}
}

func TestStreamEntryReadAndClose(t *testing.T) {
	body := &closeTrackingBody{Reader: strings.NewReader("data"), closed: make(chan struct{})}
	entry := &StreamEntry{Body: body}
	buffer := make([]byte, 4)
	n, err := entry.read(buffer)
	if err != nil || string(buffer[:n]) != "data" {
		t.Fatalf("read = %d, %v, %q", n, err, buffer[:n])
	}
	entry.close()
	if _, err := entry.read(buffer); err == nil {
		t.Fatal("read succeeded after stream close")
	}
}

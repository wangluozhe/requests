package models

import "testing"

func TestPrepareAuthUsesURLPasswordOnce(t *testing.T) {
	request := NewPrepareRequest()
	if err := request.Prepare_headers(nil); err != nil {
		t.Fatalf("Prepare_headers returned error: %v", err)
	}
	if err := request.Prepare_auth(nil, "https://alice:secret@example.com/"); err != nil {
		t.Fatalf("Prepare_auth returned error: %v", err)
	}

	const want = "Basic YWxpY2U6c2VjcmV0"
	if got := request.Headers.Get("Authorization"); got != want {
		t.Fatalf("Authorization = %q, want %q", got, want)
	}
}

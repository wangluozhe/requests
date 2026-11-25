package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/wangluozhe/requests/libs"
)

func TestBuildRequestBodyBase64(t *testing.T) {
	for _, test := range []struct {
		name    string
		body    string
		want    []byte
		wantErr bool
	}{
		{name: "empty"},
		{name: "text", body: "aGVsbG8=", want: []byte("hello")},
		{name: "binary", body: "AP+A", want: []byte{0, 255, 128}},
		{name: "line breaks", body: "aGVs\r\nbG8=", want: []byte("hello")},
		{name: "invalid characters", body: "%%invalid%%", wantErr: true},
		{name: "missing padding", body: "aGVsbG8", wantErr: true},
		{name: "extra padding", body: "aGVsbG8==", wantErr: true},
		{name: "valid prefix with invalid suffix", body: "aGVsbG8h!", wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			req, err := buildRequest(libs.RequestParams{
				Method: "POST", Url: "http://example.com", Body: test.body,
			})
			if test.wantErr {
				var decodeErr base64.CorruptInputError
				if req != nil || !errors.As(err, &decodeErr) || !strings.Contains(err.Error(), "Body") {
					t.Fatalf("request=%v, err=%v; want Body decoding error and no partial request", req, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if test.body == "" {
				if req.Body != nil {
					t.Fatal("empty Body should remain unset")
				}
				return
			}
			if req.Body == nil {
				t.Fatal("decoded Body is missing")
			}
			got, err := io.ReadAll(req.GetBody())
			if err != nil || !bytes.Equal(got, test.want) {
				t.Fatalf("Body=%v, err=%v; want %v", got, err, test.want)
			}
		})
	}
}

func TestCAPIRejectsInvalidBodyBase64(t *testing.T) {
	for _, name := range []string{"request", "stream_request"} {
		t.Run(name, func(t *testing.T) {
			id := t.Name()
			t.Cleanup(func() { _ = closeSession(id) })
			payload, err := json.Marshal(libs.RequestParams{
				Id: id, Method: "POST", Url: "http://127.0.0.1:1", Body: "%%invalid%%",
			})
			if err != nil {
				t.Fatal(err)
			}
			inputID := id + "/input"
			input := managedCString(inputID, string(payload))
			defer freeManagedCString(inputID)
			call := request
			if name == "stream_request" {
				call = stream_request
			}
			result := call(input)
			var response struct {
				ID  string `json:"id"`
				Err string `json:"err"`
			}
			if err := json.Unmarshal([]byte(goString(result)), &response); err != nil {
				t.Fatalf("invalid error response: %v", err)
			}
			defer freeManagedCString(response.ID)
			if response.ID == "" || !strings.Contains(response.Err, "Body") || !strings.Contains(response.Err, "Base64") {
				t.Fatalf("expected managed Body Base64 error, got %+v", response)
			}
			if _, exists := sessionsPool.Load(id); exists {
				t.Fatal("invalid Body created a session")
			}
		})
	}
}

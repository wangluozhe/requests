package models

import (
	"io"
	"strings"
	"testing"
)

func TestSimpleJsonReadsContentWithoutConsumingBody(t *testing.T) {
	res := &Response{
		Content: []byte(`{"message":"from content"}`),
		Body:    io.NopCloser(strings.NewReader(`body remains readable`)),
	}

	value, err := res.SimpleJson()
	if err != nil {
		t.Fatalf("SimpleJson returned error: %v", err)
	}
	if got, err := value.Get("message").String(); err != nil || got != "from content" {
		t.Fatalf("parsed message = %q, error = %v", got, err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("reading Body returned error: %v", err)
	}
	if string(body) != "body remains readable" {
		t.Fatalf("Body = %q, want original body", body)
	}
}

func TestSimpleJsonRestoresBodyWhenContentIsUnavailable(t *testing.T) {
	res := &Response{Body: io.NopCloser(strings.NewReader(`{"ok":true}`))}

	value, err := res.SimpleJson()
	if err != nil {
		t.Fatalf("SimpleJson returned error: %v", err)
	}
	if ok, err := value.Get("ok").Bool(); err != nil || !ok {
		t.Fatalf("parsed ok = %v, error = %v", ok, err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("reading restored Body returned error: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("restored Body = %q", body)
	}
	if string(res.Content) != `{"ok":true}` {
		t.Fatalf("Content = %q", res.Content)
	}
}

func TestSimpleJsonRejectsMissingBody(t *testing.T) {
	if _, err := (&Response{}).SimpleJson(); err == nil {
		t.Fatal("SimpleJson should reject a response without Content or Body")
	}
}

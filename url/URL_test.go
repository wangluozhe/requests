package url

import "testing"

func TestParseInvalidURLReturnsError(t *testing.T) {
	parsed, err := Parse("://bad")
	if err == nil {
		t.Fatal("Parse returned nil error for an invalid URL")
	}
	if parsed != nil {
		t.Fatalf("Parse returned %#v with an error", parsed)
	}
}

func TestURLStringHandlesZeroValues(t *testing.T) {
	var zero URL
	_ = zero.String()

	var nilURL *URL
	if got := nilURL.String(); got != "" {
		t.Fatalf("nil URL String() = %q, want empty string", got)
	}
}

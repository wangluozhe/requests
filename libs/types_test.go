package libs

import (
	"encoding/json"
	"testing"
)

func TestRequestParamsTracksOptionalBooleans(t *testing.T) {
	var omitted RequestParams
	if err := json.Unmarshal([]byte(`{"Method":"GET"}`), &omitted); err != nil {
		t.Fatalf("unmarshal omitted fields: %v", err)
	}
	if omitted.VerifySet() || omitted.AllowRedirectsSet() || omitted.RandomJA3Set() || omitted.ForceHTTP1Set() {
		t.Fatal("omitted boolean fields were marked as set")
	}

	var explicit RequestParams
	if err := json.Unmarshal([]byte(`{"Verify":false,"AllowRedirects":false,"RandomJA3":false,"ForceHTTP1":false}`), &explicit); err != nil {
		t.Fatalf("unmarshal explicit false fields: %v", err)
	}
	if !explicit.VerifySet() || !explicit.AllowRedirectsSet() || !explicit.RandomJA3Set() || !explicit.ForceHTTP1Set() {
		t.Fatal("explicit false fields were not marked as set")
	}
}

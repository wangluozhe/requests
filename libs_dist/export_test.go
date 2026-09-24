package main

import (
	"encoding/json"
	"testing"

	"github.com/wangluozhe/requests/libs"
)

func TestBuildRequestPreservesCBooleanDefaults(t *testing.T) {
	var omitted libs.RequestParams
	if err := json.Unmarshal([]byte(`{"Method":"GET","Url":"http://example.com"}`), &omitted); err != nil {
		t.Fatalf("unmarshal omitted fields: %v", err)
	}
	omittedRequest, err := buildRequest(omitted)
	if err != nil {
		t.Fatalf("build omitted request: %v", err)
	}
	if omittedRequest.Verify != nil || omittedRequest.AllowRedirects != nil {
		t.Fatal("omitted booleans should inherit session defaults")
	}

	var explicit libs.RequestParams
	if err := json.Unmarshal([]byte(`{"Method":"GET","Url":"http://example.com","Verify":false,"AllowRedirects":false}`), &explicit); err != nil {
		t.Fatalf("unmarshal explicit fields: %v", err)
	}
	explicitRequest, err := buildRequest(explicit)
	if err != nil {
		t.Fatalf("build explicit request: %v", err)
	}
	if explicitRequest.Verify == nil || *explicitRequest.Verify {
		t.Fatal("explicit Verify=false was not preserved")
	}
	if explicitRequest.AllowRedirects == nil || *explicitRequest.AllowRedirects {
		t.Fatal("explicit AllowRedirects=false was not preserved")
	}
}

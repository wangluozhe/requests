package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/wangluozhe/requests/libs"
)

func TestBuildRequestTrustAnchors(t *testing.T) {
	for _, test := range []struct {
		name    string
		config  string
		present bool
		want    [][]byte
		wantErr bool
	}{
		{name: "omitted", config: `{}`},
		{name: "null", config: `{"TrustAnchors":null}`},
		{name: "empty list", config: `{"TrustAnchors":[]}`, present: true, want: [][]byte{}},
		{name: "base64 IDs", config: `{"TrustAnchors":["AQI=","AwQ="]}`, present: true, want: [][]byte{{1, 2}, {3, 4}}},
		{name: "empty ID", config: `{"TrustAnchors":[""]}`, wantErr: true},
		{name: "null ID", config: `{"TrustAnchors":[null]}`, wantErr: true},
		{name: "invalid base64", config: `{"TrustAnchors":["!"]}`, wantErr: true},
		{name: "wrong type", config: `{"TrustAnchors":42}`, wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			req, err := buildRequest(libs.RequestParams{
				Method: "GET", Url: "https://example.com", TLSExtensions: test.config,
			})
			if test.wantErr {
				if err == nil || req != nil {
					t.Fatalf("invalid TrustAnchors returned request=%v, err=%v", req, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if req.TLSExtensions == nil {
				t.Fatal("TLS configuration was dropped")
			}
			ext := req.TLSExtensions.TrustAnchors
			if (ext != nil) != test.present {
				t.Fatalf("TrustAnchors = %#v, want present=%v", ext, test.present)
			}
			if ext != nil && !reflect.DeepEqual(ext.TrustAnchorIDs, test.want) {
				t.Fatalf("TrustAnchorIDs = %v, want %v", ext.TrustAnchorIDs, test.want)
			}
		})
	}
}

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

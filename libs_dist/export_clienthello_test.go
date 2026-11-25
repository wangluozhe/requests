package main

import (
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	utls "github.com/refraction-networking/utls"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/libs"
)

func TestBuildRequestClientHello(t *testing.T) {
	for _, test := range []struct {
		name    string
		fields  string
		stream  string
		preset  *utls.ClientHelloID
		legacy  string
		wantErr bool
	}{
		{name: "omitted", fields: `{}`},
		{name: "null", fields: `{"ClientHelloID":null}`},
		{name: "direct capture", fields: `{"ClientHelloHexStream":"0102"}`, stream: "0102"},
		{name: "preset", fields: `{"ClientHelloID":{"Client":"Firefox","Version":"148"}}`, preset: &utls.HelloFirefox_148},
		{name: "randomized parameters", fields: `{"ClientHelloID":{"Client":"Randomized","Version":"0","Weights":{"Extensions_Append_ALPN":0.5}}}`, preset: &utls.ClientHelloID{Client: "Randomized", Version: "0", Weights: &utls.Weights{Extensions_Append_ALPN: 0.5}}},
		{name: "legacy capture", fields: `{"TLSExtensions":"{\"ClientHelloHexStream\":\"0304\"}"}`, legacy: "0304"},
		{name: "legacy and new capture", fields: `{"ClientHelloHexStream":"0102","TLSExtensions":"{\"ClientHelloHexStream\":\"0304\"}"}`, stream: "0102", legacy: "0304"},
		{name: "invalid capture type", fields: `{"ClientHelloHexStream":12}`, wantErr: true},
		{name: "invalid preset type", fields: `{"ClientHelloID":"Firefox"}`, wantErr: true},
		{name: "invalid version type", fields: `{"ClientHelloID":{"Version":148}}`, wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var params libs.RequestParams
			err := json.Unmarshal([]byte(test.fields), &params)
			if test.wantErr {
				if err == nil {
					t.Fatal("invalid JSON field type accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			params.Method, params.Url = "GET", "https://example.com"
			request, err := buildRequest(params)
			if err != nil {
				t.Fatal(err)
			}
			if request.ClientHelloHexStream != test.stream || !reflect.DeepEqual(request.ClientHelloID, test.preset) {
				t.Fatalf("ClientHello JSON configuration was not preserved: %+v", request)
			}
			if test.legacy != "" && (request.TLSExtensions == nil || request.TLSExtensions.ClientHelloHexStream != test.legacy) {
				t.Fatal("legacy ClientHello configuration was not preserved")
			}
		})
	}
}

func TestBuildRequestClientHelloPresetNegotiatesHTTP2(t *testing.T) {
	server := httptest.NewUnstartedServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		_, _ = w.Write([]byte(r.Proto))
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()
	var params libs.RequestParams
	if err := json.Unmarshal([]byte(`{"Method":"GET","Verify":false,"Timeout":5,"Ja3":"ignored","ClientHelloID":{"Client":"Firefox","Version":"148"}}`), &params); err != nil {
		t.Fatal(err)
	}
	params.Url = server.URL
	request, err := buildRequest(params)
	if err != nil {
		t.Fatal(err)
	}
	session := requests.NewSession()
	defer session.Close()
	response, err := session.Get(params.Url, request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Close()
	if response.Text() != "HTTP/2.0" {
		t.Fatalf("preset from C JSON negotiated %s, want HTTP/2.0", response.Text())
	}
}

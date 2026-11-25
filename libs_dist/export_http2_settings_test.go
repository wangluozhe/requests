package main

import (
	"crypto/tls"
	"io"
	stdhttp "net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	http "github.com/wangluozhe/chttp"
	"github.com/wangluozhe/requests"
	"github.com/wangluozhe/requests/libs"
	"golang.org/x/net/http2"
)

func TestBuildRequestHTTP2SettingsValidation(t *testing.T) {
	for _, test := range []struct {
		name, config string
		want         []http.HTTP2Setting
		wantErr      string
	}{
		{name: "empty", config: `{}`},
		{name: "null", config: `null`},
		{
			name:   "zero values and numeric order",
			config: `{"Settings":{"UNKNOWN_SETTING_10":0,"MAX_CONCURRENT_STREAMS":0,"HEADER_TABLE_SIZE":0,"ENABLE_PUSH":0,"INITIAL_WINDOW_SIZE":0}}`,
			want:   []http.HTTP2Setting{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 10}},
		},
		{
			name:   "explicit fingerprint order",
			config: `{"Settings":{"HEADER_TABLE_SIZE":0,"ENABLE_PUSH":0},"SettingsOrder":["ENABLE_PUSH","HEADER_TABLE_SIZE"]}`,
			want:   []http.HTTP2Setting{{ID: 2}, {ID: 1}},
		},
		{name: "negative setting", config: `{"Settings":{"HEADER_TABLE_SIZE":-1}}`, wantErr: `Settings["HEADER_TABLE_SIZE"]`},
		{name: "invalid known setting", config: `{"Settings":{"NO_RFC7540_PRIORITIES":2}}`, wantErr: `Settings["NO_RFC7540_PRIORITIES"]`},
		{name: "missing ordered setting", config: `{"SettingsOrder":["ENABLE_PUSH"]}`, wantErr: "SettingsOrder[0]"},
		{name: "invalid flow", config: `{"ConnectionFlow":2147483647}`, wantErr: "ConnectionFlow"},
		{name: "invalid stream", config: `{"HeadersID":2}`, wantErr: "HeadersID"},
		{name: "invalid priority", config: `{"HeaderPriority":{"weight":"256"}}`, wantErr: "HeaderPriority.weight"},
		{name: "missing frame field", config: `{"PriorityFrames":[{}]}`, wantErr: "PriorityFrames[0].streamID"},
	} {
		t.Run(test.name, func(t *testing.T) {
			request, err := buildRequest(libs.RequestParams{Method: "GET", Url: "https://example.com", HTTP2Settings: test.config})
			if test.wantErr != "" {
				if request != nil || err == nil || !strings.Contains(err.Error(), "HTTP2Settings") || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("request=%+v, err=%v; want error containing %q", request, err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if request.HTTP2Settings == nil || !reflect.DeepEqual(request.HTTP2Settings.Settings, test.want) {
				t.Fatalf("settings=%+v; want %v", request.HTTP2Settings, test.want)
			}
		})
	}
	for _, config := range []string{
		`{`, `{"Settings":{"ENABLE_PUSH":true}}`, `{"Settings":{"ENABLE_PUSH":0.5}}`,
		`{"SettingsOrder":42}`, `{"ConnectionFlow":"1"}`, `{"HeadersID":[]}`,
	} {
		if request, err := buildRequest(libs.RequestParams{Method: "GET", Url: "https://example.com", HTTP2Settings: config}); request != nil || err == nil {
			t.Fatalf("accepted malformed JSON %s", config)
		}
	}
}

func TestBuildRequestHTTP2FingerprintOnWire(t *testing.T) {
	type observation struct {
		settings []http2.Setting
		window   uint32
		streamID uint32
	}
	observations := make(chan observation, 1)
	server := httptest.NewUnstartedServer(nil)
	server.TLS = &tls.Config{NextProtos: []string{"h2"}}
	server.Config.TLSNextProto = map[string]func(*stdhttp.Server, *tls.Conn, stdhttp.Handler){
		"h2": func(_ *stdhttp.Server, conn *tls.Conn, _ stdhttp.Handler) {
			_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
			preface := make([]byte, len(http2.ClientPreface))
			if _, err := io.ReadFull(conn, preface); err != nil || string(preface) != http2.ClientPreface {
				t.Errorf("invalid client preface %q: %v", preface, err)
				return
			}
			framer := http2.NewFramer(conn, conn)
			if err := framer.WriteSettings(); err != nil {
				t.Error(err)
				return
			}
			var observed observation
			for observed.streamID == 0 {
				frame, err := framer.ReadFrame()
				if err != nil {
					t.Error(err)
					return
				}
				switch frame := frame.(type) {
				case *http2.SettingsFrame:
					if !frame.IsAck() {
						if err := frame.ForeachSetting(func(s http2.Setting) error {
							observed.settings = append(observed.settings, s)
							return nil
						}); err != nil {
							t.Error(err)
							return
						}
						if err := framer.WriteSettingsAck(); err != nil {
							t.Error(err)
							return
						}
					}
				case *http2.WindowUpdateFrame:
					if frame.StreamID == 0 {
						observed.window = frame.Increment
					}
				case *http2.HeadersFrame:
					observed.streamID = frame.StreamID
				}
			}
			observations <- observed
			if err := framer.WriteHeaders(http2.HeadersFrameParam{StreamID: observed.streamID, EndHeaders: true, EndStream: true, BlockFragment: []byte{0x88}}); err != nil {
				t.Error(err)
			}
			// Drain ACKs until Session.Close to avoid a TCP reset on Windows.
			for {
				if _, err := framer.ReadFrame(); err != nil {
					return
				}
			}
		},
	}
	server.StartTLS()
	defer server.Close()
	request, err := buildRequest(libs.RequestParams{
		Method: "GET", Url: server.URL, Timeout: 5,
		HTTP2Settings: `{"Settings":{"HEADER_TABLE_SIZE":0,"ENABLE_PUSH":0,"MAX_CONCURRENT_STREAMS":0,"INITIAL_WINDOW_SIZE":0,"UNKNOWN_SETTING_10":0},"SettingsOrder":["ENABLE_PUSH","HEADER_TABLE_SIZE","INITIAL_WINDOW_SIZE","MAX_CONCURRENT_STREAMS","UNKNOWN_SETTING_10"],"ConnectionFlow":1048576,"HeadersID":3}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	session := requests.NewSession()
	defer session.Close()
	session.Verify = false
	response, err := session.Get(server.URL, request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Close()
	want := []http2.Setting{{ID: 2}, {ID: 1}, {ID: 4}, {ID: 3}, {ID: 10}}
	select {
	case observed := <-observations:
		if !reflect.DeepEqual(observed.settings, want) || observed.window != 1048576 || observed.streamID != 3 {
			t.Fatalf("fingerprint changed on wire: %+v; want settings=%v, window=1048576, streamID=3", observed, want)
		}
	default:
		t.Fatal("no HTTP/2 request observed")
	}
}

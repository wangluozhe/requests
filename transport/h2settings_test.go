package transport

import (
	"testing"

	http "github.com/wangluozhe/chttp"
)

func TestToHTTP2SettingsRejectsInvalidValues(t *testing.T) {
	settings := ToHTTP2Settings(&H2Settings{
		Settings: map[string]int{
			"HEADER_TABLE_SIZE":     4096,
			"ENABLE_PUSH":           2,
			"MAX_FRAME_SIZE":        -1,
			"INITIAL_WINDOW_SIZE":   1 << 31,
			"UNKNOWN_SETTING_70000": 1,
			"not-a-setting":         1,
		},
		ConnectionFlow: -1,
		HeadersID:      -2,
	})

	if settings == nil {
		t.Fatal("invalid HTTP/2 settings should produce safe defaults")
	}
	if settings.HeadersID != 1 {
		t.Fatalf("HeadersID = %d, want default 1", settings.HeadersID)
	}
	if settings.ConnectionFlow != 0 {
		t.Fatalf("ConnectionFlow = %d, want default 0", settings.ConnectionFlow)
	}
	if len(settings.Settings) != 1 || settings.Settings[0].ID != http.HTTP2SettingHeaderTableSize || settings.Settings[0].Val != 4096 {
		t.Fatalf("unexpected validated settings: %#v", settings.Settings)
	}
}

func TestToHTTP2SettingsHandlesMalformedPriorityFields(t *testing.T) {
	settings := ToHTTP2Settings(&H2Settings{
		HeaderPriority: map[string]interface{}{
			"weight":    "invalid",
			"streamDep": -1,
			"exclusive": "invalid",
		},
		PriorityFrames: []map[string]interface{}{
			nil,
			{"streamID": float64(0)},
			{"streamID": float64(3), "priorityParam": "invalid"},
			{
				"streamID": float64(5),
				"priorityParam": map[string]interface{}{
					"weight":    float64(256),
					"streamDep": float64(7),
					"exclusive": true,
				},
			},
			{
				"streamID": float64(7),
				"priorityParam": map[string]interface{}{
					"streamDep": float64(1 << 63),
				},
			},
		},
	})

	if settings == nil || settings.HeaderPriority == nil {
		t.Fatal("malformed priority fields should retain safe defaults")
	}
	if *settings.HeaderPriority != (http.HTTP2PriorityParam{}) {
		t.Fatalf("malformed HeaderPriority = %#v, want zero priority", settings.HeaderPriority)
	}
	if len(settings.PriorityFrames) != 1 {
		t.Fatalf("PriorityFrames length = %d, want 1", len(settings.PriorityFrames))
	}
	frame := settings.PriorityFrames[0]
	if frame.HTTP2FrameHeader.StreamID != 5 || frame.HTTP2PriorityParam.StreamDep != 7 || frame.HTTP2PriorityParam.Weight != 255 || !frame.HTTP2PriorityParam.Exclusive {
		t.Fatalf("validated priority frame = %#v, want stream 5/dependency 7/weight 255/exclusive", frame)
	}
}

func TestToHTTP2SettingsHandlesMissingInput(t *testing.T) {
	settings := ToHTTP2Settings(nil)
	if settings == nil {
		t.Fatal("nil input should produce safe defaults")
	}
	if settings.HeadersID != 1 || settings.HeaderPriority == nil {
		t.Fatalf("nil input defaults = %#v", settings)
	}
}

package transport

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"
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

func TestToHTTP2SettingsPreservesZerosAndOrder(t *testing.T) {
	input := &H2Settings{Settings: map[string]int{
		"HEADER_TABLE_SIZE": 0, "ENABLE_PUSH": 0, "MAX_CONCURRENT_STREAMS": 0,
		"INITIAL_WINDOW_SIZE": 0, "MAX_FRAME_SIZE": 16384, "MAX_HEADER_LIST_SIZE": 0,
		"ENABLE_CONNECT_PROTOCOL": 0, "NO_RFC7540_PRIORITIES": 0,
		"10": 0, "UNKNOWN_SETTING_65535": 0,
	}}
	want := []http.HTTP2Setting{
		{ID: 1, Val: 0}, {ID: 2, Val: 0}, {ID: 3, Val: 0}, {ID: 4, Val: 0},
		{ID: 5, Val: 16384}, {ID: 6, Val: 0}, {ID: 8, Val: 0}, {ID: 9, Val: 0},
		{ID: 10, Val: 0}, {ID: 65535, Val: 0},
	}
	for range 32 {
		checked, err := ToHTTP2SettingsChecked(input)
		if err != nil {
			t.Fatal(err)
		}
		legacy := ToHTTP2Settings(input)
		if !reflect.DeepEqual(checked.Settings, want) || !reflect.DeepEqual(legacy, checked) {
			t.Fatalf("zero settings/order changed: checked=%+v, legacy=%+v, want=%v", checked, legacy, want)
		}
	}
	// Explicit order still selects a subset and can repeat a setting on the wire.
	input.SettingsOrder = []string{"NO_RFC7540_PRIORITIES", "HEADER_TABLE_SIZE", "NO_RFC7540_PRIORITIES"}
	want = []http.HTTP2Setting{{ID: 9, Val: 0}, {ID: 1, Val: 0}, {ID: 9, Val: 0}}
	checked, err := ToHTTP2SettingsChecked(input)
	if err != nil || !reflect.DeepEqual(checked.Settings, want) || !reflect.DeepEqual(ToHTTP2Settings(input), checked) {
		t.Fatalf("explicit order changed: checked=%+v, err=%v", checked, err)
	}
}

func TestToHTTP2SettingsCheckedRejectsInvalidInput(t *testing.T) {
	for _, test := range []struct {
		name, input, field string
	}{
		{"unknown name", `{"Settings":{"typo":1}}`, `Settings["typo"]`},
		{"negative ID", `{"Settings":{"-1":1}}`, `Settings["-1"]`},
		{"ID overflow", `{"Settings":{"UNKNOWN_SETTING_65536":1}}`, `Settings["UNKNOWN_SETTING_65536"]`},
		{"negative value", `{"Settings":{"HEADER_TABLE_SIZE":-1}}`, `Settings["HEADER_TABLE_SIZE"]`},
		{"enable push", `{"Settings":{"ENABLE_PUSH":2}}`, `Settings["ENABLE_PUSH"]`},
		{"connect protocol", `{"Settings":{"ENABLE_CONNECT_PROTOCOL":2}}`, `Settings["ENABLE_CONNECT_PROTOCOL"]`},
		{"numeric connect protocol", `{"Settings":{"8":2}}`, `Settings["8"]`},
		{"no priorities", `{"Settings":{"NO_RFC7540_PRIORITIES":2}}`, `Settings["NO_RFC7540_PRIORITIES"]`},
		{"numeric no priorities", `{"Settings":{"UNKNOWN_SETTING_9":2}}`, `Settings["UNKNOWN_SETTING_9"]`},
		{"frame minimum", `{"Settings":{"MAX_FRAME_SIZE":16383}}`, `Settings["MAX_FRAME_SIZE"]`},
		{"frame maximum", `{"Settings":{"MAX_FRAME_SIZE":16777216}}`, `Settings["MAX_FRAME_SIZE"]`},
		{"missing ordered field", `{"SettingsOrder":["HEADER_TABLE_SIZE"]}`, "SettingsOrder[0]"},
		{"negative connection window", `{"ConnectionFlow":-1}`, "ConnectionFlow"},
		{"connection window overflow", `{"ConnectionFlow":2147418113}`, "ConnectionFlow"},
		{"negative stream", `{"HeadersID":-1}`, "HeadersID"},
		{"even stream", `{"HeadersID":2}`, "HeadersID"},
		{"weight type", `{"HeaderPriority":{"weight":"256"}}`, "HeaderPriority.weight"},
		{"weight zero", `{"HeaderPriority":{"weight":0}}`, "HeaderPriority.weight"},
		{"weight overflow", `{"HeaderPriority":{"weight":257}}`, "HeaderPriority.weight"},
		{"weight fraction", `{"HeaderPriority":{"weight":1.5}}`, "HeaderPriority.weight"},
		{"dependency negative", `{"HeaderPriority":{"streamDep":-1}}`, "HeaderPriority.streamDep"},
		{"dependency overflow", `{"HeaderPriority":{"streamDep":2147483648}}`, "HeaderPriority.streamDep"},
		{"exclusive type", `{"HeaderPriority":{"exclusive":1}}`, "HeaderPriority.exclusive"},
		{"header self dependency", `{"HeaderPriority":{"streamDep":1}}`, "HeaderPriority.streamDep"},
		{"null frame", `{"PriorityFrames":[null]}`, "PriorityFrames[0].streamID"},
		{"missing stream", `{"PriorityFrames":[{"priorityParam":{}}]}`, "PriorityFrames[0].streamID"},
		{"stream zero", `{"PriorityFrames":[{"streamID":0,"priorityParam":{}}]}`, "PriorityFrames[0].streamID"},
		{"stream overflow", `{"PriorityFrames":[{"streamID":2147483648,"priorityParam":{}}]}`, "PriorityFrames[0].streamID"},
		{"stream type", `{"PriorityFrames":[{"streamID":"3","priorityParam":{}}]}`, "PriorityFrames[0].streamID"},
		{"missing priority", `{"PriorityFrames":[{"streamID":3}]}`, "PriorityFrames[0].priorityParam"},
		{"null priority", `{"PriorityFrames":[{"streamID":3,"priorityParam":null}]}`, "PriorityFrames[0].priorityParam"},
		{"priority type", `{"PriorityFrames":[{"streamID":3,"priorityParam":true}]}`, "PriorityFrames[0].priorityParam"},
		{"frame weight", `{"PriorityFrames":[{"streamID":3,"priorityParam":{"weight":257}}]}`, "PriorityFrames[0].priorityParam.weight"},
		{"frame self dependency", `{"PriorityFrames":[{"streamID":3,"priorityParam":{"streamDep":3}}]}`, "PriorityFrames[0].priorityParam.streamDep"},
		{"derived stream overflow", `{"PriorityFrames":[{"streamID":2147483647,"priorityParam":{}}]}`, "PriorityFrames[0].streamID"},
		{"derived self dependency", `{"HeaderPriority":{"streamDep":5},"PriorityFrames":[{"streamID":3,"priorityParam":{}}]}`, "HeaderPriority.streamDep"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var input H2Settings
			if err := json.Unmarshal([]byte(test.input), &input); err != nil {
				t.Fatal(err)
			}
			result, err := ToHTTP2SettingsChecked(&input)
			if result != nil || err == nil || !strings.Contains(err.Error(), test.field) {
				t.Fatalf("result=%+v, err=%v; want nil and error containing %q", result, err, test.field)
			}
			if legacy := ToHTTP2Settings(&input); legacy == nil {
				t.Fatal("legacy conversion must retain its safe defaults")
			}
		})
	}
}

func TestToHTTP2SettingsCheckedBoundaries(t *testing.T) {
	for _, input := range []*H2Settings{nil, {}} {
		result, err := ToHTTP2SettingsChecked(input)
		if err != nil || result.HeadersID != 1 || result.HeaderPriority == nil || !reflect.DeepEqual(result, ToHTTP2Settings(nil)) {
			t.Fatalf("default configuration: result=%+v, err=%v", result, err)
		}
	}
	input := &H2Settings{
		Settings: map[string]int{
			"INITIAL_WINDOW_SIZE": math.MaxInt32, "MAX_FRAME_SIZE": 1<<24 - 1,
			"ENABLE_CONNECT_PROTOCOL": 1, "NO_RFC7540_PRIORITIES": 1,
		},
		ConnectionFlow: math.MaxInt32 - 65535, HeadersID: math.MaxInt32, SettingsAck: true,
		HeaderPriority: map[string]interface{}{"weight": 256, "streamDep": 0, "exclusive": true},
		PriorityFrames: []map[string]interface{}{
			{"streamID": math.MaxInt32 - 1, "priorityParam": map[string]interface{}{"weight": 1, "streamDep": 0}},
		},
	}
	result, err := ToHTTP2SettingsChecked(input)
	if err != nil {
		t.Fatal(err)
	}
	if result.ConnectionFlow != input.ConnectionFlow || result.HeadersID != input.HeadersID || !result.SettingsAck || result.HeaderPriority.Weight != 255 || !result.HeaderPriority.Exclusive || result.PriorityFrames[0].HTTP2PriorityParam.Weight != 0 {
		t.Fatalf("boundary configuration changed: %+v", result)
	}
	for _, value := range []interface{}{math.NaN(), math.Inf(1), float32(math.Inf(-1)), uint64(math.MaxUint64)} {
		input = &H2Settings{HeaderPriority: map[string]interface{}{"weight": value}}
		if result, err := ToHTTP2SettingsChecked(input); result != nil || err == nil {
			t.Fatalf("accepted invalid number %v", value)
		}
	}
	if strconv.IntSize == 64 {
		maxUint32 := uint64(math.MaxUint32)
		input = &H2Settings{Settings: map[string]int{"UNKNOWN_SETTING_65535": int(maxUint32)}}
		result, err := ToHTTP2SettingsChecked(input)
		if err != nil || len(result.Settings) != 1 || result.Settings[0].Val != math.MaxUint32 {
			t.Fatalf("uint32 maximum: result=%+v, err=%v", result, err)
		}
		for _, field := range []string{"HEADER_TABLE_SIZE", "INITIAL_WINDOW_SIZE"} {
			input.Settings = map[string]int{field: int(maxUint32 + 1)}
			if result, err := ToHTTP2SettingsChecked(input); result != nil || err == nil {
				t.Fatalf("accepted overflow in %s", field)
			}
		}
	}
}

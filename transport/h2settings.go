package transport

import (
	"math"
	"strconv"
	"strings"

	http "github.com/wangluozhe/chttp"
)

var settings = map[string]http.HTTP2SettingID{
	"HEADER_TABLE_SIZE":      http.HTTP2SettingHeaderTableSize,
	"ENABLE_PUSH":            http.HTTP2SettingEnablePush,
	"MAX_CONCURRENT_STREAMS": http.HTTP2SettingMaxConcurrentStreams,
	"INITIAL_WINDOW_SIZE":    http.HTTP2SettingInitialWindowSize,
	"MAX_FRAME_SIZE":         http.HTTP2SettingMaxFrameSize,
	"MAX_HEADER_LIST_SIZE":   http.HTTP2SettingMaxHeaderListSize,
	"NO_RFC7540_PRIORITIES":  http.HTTP2SettingID(0x9),
}

type H2Settings struct {
	//HEADER_TABLE_SIZE
	//ENABLE_PUSH
	//MAX_CONCURRENT_STREAMS
	//INITIAL_WINDOW_SIZE
	//MAX_FRAME_SIZE
	//MAX_HEADER_LIST_SIZE
	Settings    map[string]int `json:"Settings"`
	SettingsAck bool           `json:"SettingsAck"`
	//HEADER_TABLE_SIZE
	//ENABLE_PUSH
	//MAX_CONCURRENT_STREAMS
	//INITIAL_WINDOW_SIZE
	//MAX_FRAME_SIZE
	//MAX_HEADER_LIST_SIZE
	SettingsOrder  []string                 `json:"SettingsOrder"`
	ConnectionFlow int                      `json:"ConnectionFlow"`
	HeadersID      int                      `json:"HeadersID"`
	HeaderPriority map[string]interface{}   `json:"HeaderPriority"`
	PriorityFrames []map[string]interface{} `json:"PriorityFrames"`
}

func settingNameToID(name string) (http.HTTP2SettingID, bool) {
	if id, ok := settings[name]; ok {
		return id, true
	}
	if strings.HasPrefix(name, "UNKNOWN_SETTING_") {
		name = strings.TrimPrefix(name, "UNKNOWN_SETTING_")
	}
	id, err := strconv.ParseUint(name, 10, 16)
	if err != nil {
		return 0, false
	}
	return http.HTTP2SettingID(id), true
}

// covertSettingNameToID 保留旧的内部函数，非法名称统一返回 0。
func covertSettingNameToID(name string) http.HTTP2SettingID {
	id, _ := settingNameToID(name)
	return id
}

func validHTTP2Setting(id http.HTTP2SettingID, value uint32) bool {
	switch id {
	case http.HTTP2SettingEnablePush:
		return value <= 1
	case http.HTTP2SettingInitialWindowSize:
		return value <= math.MaxInt32
	case http.HTTP2SettingMaxFrameSize:
		return value >= 1<<14 && value <= 1<<24-1
	default:
		return true
	}
}

func toUint32(value int) (uint32, bool) {
	if value < 0 || uint64(value) > math.MaxUint32 {
		return 0, false
	}
	return uint32(value), true
}

func numberValue(value interface{}) (int64, bool) {
	switch n := value.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case uint:
		if uint64(n) > math.MaxInt64 {
			return 0, false
		}
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		if n > math.MaxInt64 {
			return 0, false
		}
		return int64(n), true
	case float32:
		f := float64(n)
		if math.IsNaN(f) || math.IsInf(f, 0) || f != math.Trunc(f) || f < math.MinInt64 || f >= float64(math.MaxInt64) {
			return 0, false
		}
		return int64(f), true
	case float64:
		if math.IsNaN(n) || math.IsInf(n, 0) || n != math.Trunc(n) || n < math.MinInt64 || n >= float64(math.MaxInt64) {
			return 0, false
		}
		return int64(n), true
	default:
		return 0, false
	}
}

func priorityParamFromMap(source map[string]interface{}) (http.HTTP2PriorityParam, bool) {
	var priority http.HTTP2PriorityParam
	if source == nil {
		return priority, true
	}
	if raw, exists := source["weight"]; exists {
		weight, ok := numberValue(raw)
		if !ok || weight < 1 || weight > 256 {
			return priority, false
		}
		priority.Weight = uint8(weight - 1)
	}
	if raw, exists := source["streamDep"]; exists {
		streamDep, ok := numberValue(raw)
		if !ok || streamDep < 0 || streamDep > math.MaxInt32 {
			return priority, false
		}
		priority.StreamDep = uint32(streamDep)
	}
	if raw, exists := source["exclusive"]; exists {
		exclusive, ok := raw.(bool)
		if !ok {
			return priority, false
		}
		priority.Exclusive = exclusive
	}
	return priority, true
}

func priorityFrameFromMap(frame map[string]interface{}) (http.HTTP2PriorityFrame, bool) {
	var result http.HTTP2PriorityFrame
	if frame == nil {
		return result, false
	}
	rawStreamID, exists := frame["streamID"]
	if !exists {
		return result, false
	}
	streamID, ok := numberValue(rawStreamID)
	if !ok || streamID <= 0 || streamID > math.MaxInt32 {
		return result, false
	}
	rawPriority, exists := frame["priorityParam"]
	if !exists {
		return result, false
	}
	prioritySource, ok := rawPriority.(map[string]interface{})
	if !ok {
		return result, false
	}
	priority, ok := priorityParamFromMap(prioritySource)
	if !ok {
		return result, false
	}
	result.HTTP2FrameHeader.StreamID = uint32(streamID)
	result.HTTP2PriorityParam = priority
	return result, true
}

func ToHTTP2Settings(h2Settings *H2Settings) (http2Settings *http.HTTP2Settings) {
	http2Settings = &http.HTTP2Settings{
		Settings:       nil,
		ConnectionFlow: 0,
		HeaderPriority: &http.HTTP2PriorityParam{},
		PriorityFrames: nil,
	}
	if h2Settings == nil {
		http2Settings.HeadersID = 1
		return http2Settings
	}
	if h2Settings.Settings != nil {
		appendSetting := func(name string, value int) {
			id, ok := settingNameToID(name)
			if !ok {
				return
			}
			converted, ok := toUint32(value)
			if !ok || !validHTTP2Setting(id, converted) {
				return
			}
			if converted == 0 && id != http.HTTP2SettingEnablePush {
				return
			}
			http2Settings.Settings = append(http2Settings.Settings, http.HTTP2Setting{ID: id, Val: converted})
		}
		if len(h2Settings.SettingsOrder) > 0 {
			for _, orderKey := range h2Settings.SettingsOrder {
				if val, exists := h2Settings.Settings[orderKey]; exists {
					appendSetting(orderKey, val)
				}
			}
		} else {
			for id, val := range h2Settings.Settings {
				appendSetting(id, val)
			}
		}
	}
	if h2Settings.SettingsAck {
		http2Settings.SettingsAck = true
	}
	if h2Settings.ConnectionFlow > 0 && h2Settings.ConnectionFlow <= math.MaxInt32 {
		http2Settings.ConnectionFlow = h2Settings.ConnectionFlow
	}
	if h2Settings.HeadersID <= 0 || h2Settings.HeadersID > math.MaxInt32 {
		http2Settings.HeadersID = 1
	} else {
		http2Settings.HeadersID = h2Settings.HeadersID
	}
	if h2Settings.HeaderPriority != nil {
		if priority, ok := priorityParamFromMap(h2Settings.HeaderPriority); ok {
			http2Settings.HeaderPriority = &priority
		}
	}
	if h2Settings.PriorityFrames != nil {
		for _, frame := range h2Settings.PriorityFrames {
			if parsed, ok := priorityFrameFromMap(frame); ok {
				http2Settings.PriorityFrames = append(http2Settings.PriorityFrames, parsed)
			}
		}
	}
	return http2Settings
}

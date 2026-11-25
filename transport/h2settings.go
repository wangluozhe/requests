package transport

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	http "github.com/wangluozhe/chttp"
)

var settings = map[string]http.HTTP2SettingID{
	"HEADER_TABLE_SIZE":       http.HTTP2SettingHeaderTableSize,
	"ENABLE_PUSH":             http.HTTP2SettingEnablePush,
	"MAX_CONCURRENT_STREAMS":  http.HTTP2SettingMaxConcurrentStreams,
	"INITIAL_WINDOW_SIZE":     http.HTTP2SettingInitialWindowSize,
	"MAX_FRAME_SIZE":          http.HTTP2SettingMaxFrameSize,
	"MAX_HEADER_LIST_SIZE":    http.HTTP2SettingMaxHeaderListSize,
	"ENABLE_CONNECT_PROTOCOL": http.HTTP2SettingID(0x8),
	"NO_RFC7540_PRIORITIES":   http.HTTP2SettingID(0x9),
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
	// SettingsOrder 非空时只发送其中列出的字段，保留顺序及重复项；省略时按 SETTINGS ID 排序。
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
	if id == http.HTTP2SettingID(0x9) {
		return value <= 1
	}
	return (http.HTTP2Setting{ID: id, Val: value}).Valid() == nil
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

func priorityParamFromMap(source map[string]interface{}) (http.HTTP2PriorityParam, error) {
	var priority http.HTTP2PriorityParam
	if source == nil {
		return priority, nil
	}
	if raw, exists := source["weight"]; exists {
		weight, ok := numberValue(raw)
		if !ok || weight < 1 || weight > 256 {
			return priority, fmt.Errorf("weight must be an integer between 1 and 256")
		}
		priority.Weight = uint8(weight - 1)
	}
	if raw, exists := source["streamDep"]; exists {
		streamDep, ok := numberValue(raw)
		if !ok || streamDep < 0 || streamDep > math.MaxInt32 {
			return priority, fmt.Errorf("streamDep must be an integer between 0 and %d", math.MaxInt32)
		}
		priority.StreamDep = uint32(streamDep)
	}
	if raw, exists := source["exclusive"]; exists {
		exclusive, ok := raw.(bool)
		if !ok {
			return priority, fmt.Errorf("exclusive must be a boolean")
		}
		priority.Exclusive = exclusive
	}
	return priority, nil
}

func priorityFrameFromMap(frame map[string]interface{}) (http.HTTP2PriorityFrame, error) {
	var result http.HTTP2PriorityFrame
	rawStreamID, exists := frame["streamID"]
	if !exists {
		return result, fmt.Errorf("streamID is required")
	}
	streamID, ok := numberValue(rawStreamID)
	if !ok || streamID <= 0 || streamID > math.MaxInt32 {
		return result, fmt.Errorf("streamID must be an integer between 1 and %d", math.MaxInt32)
	}
	rawPriority, exists := frame["priorityParam"]
	if !exists {
		return result, fmt.Errorf("priorityParam is required")
	}
	prioritySource, ok := rawPriority.(map[string]interface{})
	if !ok {
		return result, fmt.Errorf("priorityParam must be an object")
	}
	priority, err := priorityParamFromMap(prioritySource)
	if err != nil {
		return result, fmt.Errorf("priorityParam.%w", err)
	}
	if priority.StreamDep == uint32(streamID) {
		return result, fmt.Errorf("priorityParam.streamDep must not equal streamID")
	}
	result.HTTP2FrameHeader.StreamID = uint32(streamID)
	result.HTTP2PriorityParam = priority
	return result, nil
}

// ToHTTP2Settings 保留原有单返回值接口，跳过无效项并对无效的可选字段使用默认值。
// 合法零值不会被丢弃。需要获知配置错误时使用 ToHTTP2SettingsChecked。
func ToHTTP2Settings(h2Settings *H2Settings) *http.HTTP2Settings {
	converted, _ := convertHTTP2Settings(h2Settings, false)
	return converted
}

// ToHTTP2SettingsChecked 转换 HTTP/2 指纹，遇到无效配置返回 nil 和包含字段路径的错误。
// SettingsOrder 非空时沿用原有的字段选择语义；未选中的 SETTINGS 不参与转换。
// nil 或空配置使用默认值，未提供 SettingsOrder 时按 ID 排序，保证指纹稳定。
func ToHTTP2SettingsChecked(h2Settings *H2Settings) (*http.HTTP2Settings, error) {
	return convertHTTP2Settings(h2Settings, true)
}

func convertHTTP2Settings(h2Settings *H2Settings, strict bool) (*http.HTTP2Settings, error) {
	http2Settings := &http.HTTP2Settings{
		HeadersID:      1,
		HeaderPriority: &http.HTTP2PriorityParam{},
	}
	if h2Settings == nil {
		return http2Settings, nil
	}
	names := h2Settings.SettingsOrder
	if len(names) == 0 {
		names = make([]string, 0, len(h2Settings.Settings))
		for name := range h2Settings.Settings {
			names = append(names, name)
		}
		// 数字 ID 决定默认顺序；同一 ID 的不同名称以字典序打破平局。
		sort.Slice(names, func(i, j int) bool {
			left, _ := settingNameToID(names[i])
			right, _ := settingNameToID(names[j])
			if left == right {
				return names[i] < names[j]
			}
			return left < right
		})
	}
	for index, name := range names {
		value, exists := h2Settings.Settings[name]
		if !exists {
			if strict {
				return nil, fmt.Errorf("invalid SettingsOrder[%d]: %q is missing from Settings", index, name)
			}
			continue
		}
		id, ok := settingNameToID(name)
		if !ok {
			if strict {
				return nil, fmt.Errorf("invalid Settings[%q]: expected a known name or a 16-bit setting ID", name)
			}
			continue
		}
		converted, ok := toUint32(value)
		if !ok || !validHTTP2Setting(id, converted) {
			if strict {
				return nil, fmt.Errorf("invalid Settings[%q] value %d: outside the HTTP/2 setting range", name, value)
			}
			continue
		}
		http2Settings.Settings = append(http2Settings.Settings, http.HTTP2Setting{ID: id, Val: converted})
	}
	http2Settings.SettingsAck = h2Settings.SettingsAck
	// ConnectionFlow 是在初始 65535 字节窗口之上增加的额度。
	if h2Settings.ConnectionFlow < 0 || h2Settings.ConnectionFlow > math.MaxInt32-65535 {
		if strict {
			return nil, fmt.Errorf("invalid ConnectionFlow %d: must be between 0 and %d", h2Settings.ConnectionFlow, math.MaxInt32-65535)
		}
	} else {
		http2Settings.ConnectionFlow = h2Settings.ConnectionFlow
	}
	if h2Settings.HeadersID < 0 || h2Settings.HeadersID > math.MaxInt32 || (h2Settings.HeadersID != 0 && h2Settings.HeadersID%2 == 0) {
		if strict {
			return nil, fmt.Errorf("invalid HeadersID %d: must be 0 (default) or an odd 31-bit stream ID", h2Settings.HeadersID)
		}
	} else if h2Settings.HeadersID != 0 {
		http2Settings.HeadersID = h2Settings.HeadersID
	}
	if h2Settings.HeaderPriority != nil {
		if priority, err := priorityParamFromMap(h2Settings.HeaderPriority); err == nil {
			http2Settings.HeaderPriority = &priority
		} else if strict {
			return nil, fmt.Errorf("invalid HeaderPriority.%w", err)
		}
	}
	for index, frame := range h2Settings.PriorityFrames {
		if parsed, err := priorityFrameFromMap(frame); err == nil {
			http2Settings.PriorityFrames = append(http2Settings.PriorityFrames, parsed)
		} else if strict {
			return nil, fmt.Errorf("invalid PriorityFrames[%d].%w", index, err)
		}
	}
	firstStreamID := uint32(http2Settings.HeadersID)
	for len(http2Settings.PriorityFrames) > 0 {
		last := len(http2Settings.PriorityFrames) - 1
		lastID := http2Settings.PriorityFrames[last].HTTP2FrameHeader.StreamID
		if lastID == math.MaxInt32 {
			if strict {
				return nil, fmt.Errorf("invalid PriorityFrames[%d].streamID: leaves no room for the first request stream", last)
			}
			http2Settings.PriorityFrames = http2Settings.PriorityFrames[:last]
			continue
		}
		firstStreamID = max(firstStreamID, (lastID+1)|1)
		break
	}
	if http2Settings.HeaderPriority.StreamDep == firstStreamID {
		if strict {
			return nil, fmt.Errorf("invalid HeaderPriority.streamDep: must not equal the first request stream ID %d", firstStreamID)
		}
		http2Settings.HeaderPriority = &http.HTTP2PriorityParam{}
	}
	return http2Settings, nil
}

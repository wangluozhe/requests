package transport

import (
	"github.com/wangluozhe/chttp/http2"
)

var settings = map[string]http2.SettingID{
	"HEADER_TABLE_SIZE":      http2.SettingHeaderTableSize,
	"ENABLE_PUSH":            http2.SettingEnablePush,
	"MAX_CONCURRENT_STREAMS": http2.SettingMaxConcurrentStreams,
	"INITIAL_WINDOW_SIZE":    http2.SettingInitialWindowSize,
	"MAX_FRAME_SIZE":         http2.SettingMaxFrameSize,
	"MAX_HEADER_LIST_SIZE":   http2.SettingMaxHeaderListSize,
}

type H2Settings struct {
	//HEADER_TABLE_SIZE
	//ENABLE_PUSH
	//MAX_CONCURRENT_STREAMS
	//INITIAL_WINDOW_SIZE
	//MAX_FRAME_SIZE
	//MAX_HEADER_LIST_SIZE
	Settings map[string]int `json:"Settings"`
	//HEADER_TABLE_SIZE
	//ENABLE_PUSH
	//MAX_CONCURRENT_STREAMS
	//INITIAL_WINDOW_SIZE
	//MAX_FRAME_SIZE
	//MAX_HEADER_LIST_SIZE
	SettingsOrder  []string                 `json:"SettingsOrder"`
	ConnectionFlow int                      `json:"ConnectionFlow"`
	HeaderPriority map[string]interface{}   `json:"HeaderPriority"`
	PriorityFrames []map[string]interface{} `json:"PriorityFrames"`
}

func ToHTTP2Settings(h2Settings *H2Settings) (http2Settings *http2.HTTP2Settings) {
	http2Settings = &http2.HTTP2Settings{
		Settings:       nil,
		ConnectionFlow: 0,
		HeaderPriority: &http2.PriorityParam{},
		PriorityFrames: nil,
	}
	if h2Settings.Settings != nil {
		if h2Settings.SettingsOrder != nil {
			for _, orderKey := range h2Settings.SettingsOrder {
				val := h2Settings.Settings[orderKey]
				if val != 0 || orderKey == "ENABLE_PUSH" {
					http2Settings.Settings = append(http2Settings.Settings, http2.Setting{
						ID:  settings[orderKey],
						Val: uint32(val),
					})
				}
			}
		} else {
			for id, val := range h2Settings.Settings {
				http2Settings.Settings = append(http2Settings.Settings, http2.Setting{
					ID:  settings[id],
					Val: uint32(val),
				})
			}
		}
	}
	if h2Settings.ConnectionFlow != 0 {
		http2Settings.ConnectionFlow = h2Settings.ConnectionFlow
	}
	if h2Settings.HeaderPriority != nil {
		var weight int
		var streamDep int
		w := h2Settings.HeaderPriority["weight"]
		switch w.(type) {
		case int:
			weight = w.(int)
		case float64:
			weight = int(w.(float64))
		}
		s := h2Settings.HeaderPriority["streamDep"]
		switch s.(type) {
		case int:
			streamDep = s.(int)
		case float64:
			streamDep = int(s.(float64))
		}
		var priorityParam *http2.PriorityParam
		if w == nil {
			priorityParam = &http2.PriorityParam{
				StreamDep: uint32(streamDep),
				Exclusive: h2Settings.HeaderPriority["exclusive"].(bool),
			}
		} else {
			priorityParam = &http2.PriorityParam{
				StreamDep: uint32(streamDep),
				Exclusive: h2Settings.HeaderPriority["exclusive"].(bool),
				Weight:    uint8(weight - 1),
			}
		}
		http2Settings.HeaderPriority = priorityParam
	}
	if h2Settings.PriorityFrames != nil {
		for _, frame := range h2Settings.PriorityFrames {
			var weight int
			var streamDep int
			var streamID int
			priorityParamSource := frame["priorityParam"].(map[string]interface{})
			w := priorityParamSource["weight"]
			switch w.(type) {
			case int:
				weight = w.(int)
			case float64:
				weight = int(w.(float64))
			}
			s := priorityParamSource["streamDep"]
			switch s.(type) {
			case int:
				streamDep = s.(int)
			case float64:
				streamDep = int(s.(float64))
			}
			sid := frame["streamID"]
			switch sid.(type) {
			case int:
				streamID = sid.(int)
			case float64:
				streamID = int(sid.(float64))
			}
			var priorityParam http2.PriorityParam
			if w == nil {
				priorityParam = http2.PriorityParam{
					StreamDep: uint32(streamDep),
					Exclusive: priorityParamSource["exclusive"].(bool),
				}
			} else {
				priorityParam = http2.PriorityParam{
					StreamDep: uint32(streamDep),
					Exclusive: priorityParamSource["exclusive"].(bool),
					Weight:    uint8(weight - 1),
				}
			}
			http2Settings.PriorityFrames = append(http2Settings.PriorityFrames, http2.PriorityFrame{
				FrameHeader: http2.FrameHeader{
					StreamID: uint32(streamID),
				},
				PriorityParam: priorityParam,
			})
		}
	}
	return http2Settings
}

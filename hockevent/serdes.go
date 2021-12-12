package hockevent

import (
	"encoding/json"
	"fmt"
)

func Serialize(event HockEvent) string {
	var d map[string]interface{}
	data, _ := json.Marshal(event)
	_ = json.Unmarshal(data, &d)
	d["type"] = event.GetType()
	data, _ = json.Marshal(&d)

	return string(data)
}

func Deserialize(data []byte) (HockEvent, error) {
	h := hockEvent{}
	err := json.Unmarshal(data, &h)
	if err != nil {
		return nil, err
	}

	var event HockEvent

	log.Debugf("received event %d", h.Type)

	switch h.Type {
	case EventMessage:
		event = &MessageEvent{}
	case EventPlayerJoin:
		event = &PlayerJoinEvent{}
	case EventPlayerLeft:
		event = &PlayerLeftEvent{}
	case EventPlayerDeath:
		event = &PlayerDeathEvent{}
	case EventPlayerUpdate:
		event = &PlayerUpdateEvent{}
	case EventPlayerList:
		event = &PlayerListEvent{}
	case EventPlayerDimensionChange:
		event = &PlayerDimensionChangeEvent{}
	case EventLog:
		event = &LogEvent{}
	default:
		return nil, fmt.Errorf("invalid type %d", h.Type)
	}

	err = json.Unmarshal(data, event)
	return event, err
}

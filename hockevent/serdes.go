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

	switch h.Type {
	case EVENT_MESSAGE:
		event = &MessageEvent{}
	case EVENT_PLAYER_JOIN:
		event = &PlayerJoinEvent{}
	case EVENT_PLAYER_LEFT:
		event = &PlayerLeftEvent{}
	case EVENT_PLAYER_DEATH:
		event = &PlayerDeathEvent{}
	case EVENT_PLAYER_UPDATE:
		event = &PlayerUpdateEvent{}
	case EVENT_PLAYER_LIST:
		event = &PlayerListEvent{}
	default:
		return nil, fmt.Errorf("invalid type %d", h.Type)
	}

	err = json.Unmarshal(data, event)
	return event, err
}

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
	case EventMessage: // 0
		event = &MessageEvent{}
	case EventPlayerJoin: // 1
		event = &PlayerJoinEvent{}
	case EventPlayerLeft: // 2
		event = &PlayerLeftEvent{}
	case EventPlayerDeath: // 3
		event = &PlayerDeathEvent{}
	case EventPlayerUpdate: // 4
		event = &PlayerUpdateEvent{}
	case EventPlayerList: // 5
		event = &PlayerListEvent{}
	case EventPlayerDimensionChange: // 6
		event = &PlayerDimensionChangeEvent{}
	case EventLog: // 7
		event = &LogEvent{}
	case EventFormRequest: // 8
		event = &FormRequestEvent{}
	case EventFormResponse: // 9
		event = &FormResponseEvent{}
	case EventScoreboard: // 10
		event = &ScoreBoardEvent{}
	case EventPlayerDeathCountRequest: // 11
		event = &PlayerDeathCountRequestEvent{}
	case EventPlayerDeathCountResponse: // 12
		event = &PlayerDeathCountResponseEvent{}
	default:
		return nil, fmt.Errorf("invalid type %d", h.Type)
	}

	err = json.Unmarshal(data, event)
	return event, err
}

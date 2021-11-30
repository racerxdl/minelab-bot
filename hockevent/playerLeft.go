package hockevent

type PlayerLeftEvent struct {
	PlayerStatusEvent
}

func (PlayerLeftEvent) GetType() Type {
	return EVENT_PLAYER_LEFT
}

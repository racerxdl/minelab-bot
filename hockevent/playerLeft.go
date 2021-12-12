package hockevent

type PlayerLeftEvent struct {
	PlayerStatusEvent
}

func (PlayerLeftEvent) GetType() Type {
	return EventPlayerLeft
}

package hockevent

type PlayerJoinEvent struct {
	PlayerStatusEvent
}

func (PlayerJoinEvent) GetType() Type {
	return EventPlayerJoin
}

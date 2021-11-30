package hockevent

type PlayerJoinEvent struct {
	PlayerStatusEvent
}

func (PlayerJoinEvent) GetType() Type {
	return EVENT_PLAYER_JOIN
}

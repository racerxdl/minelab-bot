package hockevent

type PlayerDeathEvent struct {
	PlayerStatusEvent
}

func (PlayerDeathEvent) GetType() Type {
	return EventPlayerDeath
}

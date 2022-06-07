package hockevent

type PlayerDeathEvent struct {
	PlayerStatusEvent
	DeathReason string `json:"deathReason"`
}

func (PlayerDeathEvent) GetType() Type {
	return EventPlayerDeath
}

type PlayerDeathCountRequestEvent struct{}

func (PlayerDeathCountRequestEvent) GetType() Type {
	return EventPlayerDeathCountRequest
}

type PlayerDeathCountResponseEvent struct {
	PlayerDeaths map[string]int `json:"playerDeaths"`
}

func (PlayerDeathCountResponseEvent) GetType() Type {
	return EventPlayerDeathCountResponse
}

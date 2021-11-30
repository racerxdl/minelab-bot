package hockevent

type PlayerStatusEvent struct {
	Username string `json:"username"`
	Xuid     string `json:"xuid"`
}

func (PlayerStatusEvent) GetType() Type {
	return EVENT_INVALID
}

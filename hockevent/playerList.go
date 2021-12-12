package hockevent

type PlayerListEvent struct {
	Players []string `json:"players"`
}

func (PlayerListEvent) GetType() Type {
	return EventPlayerList
}

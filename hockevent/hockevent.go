package hockevent

type Type int

const (
	EVENT_INVALID       Type = -1
	EVENT_MESSAGE       Type = 0
	EVENT_PLAYER_JOIN   Type = 1
	EVENT_PLAYER_LEFT   Type = 2
	EVENT_PLAYER_DEATH  Type = 3
	EVENT_PLAYER_UPDATE Type = 4
	EVENT_PLAYER_LIST   Type = 5
)

type HockEvent interface {
	GetType() Type
}

type hockEvent struct {
	Type Type `json:"type"`
}

package hockevent

type Type int

const (
	EventInvalid                  Type = -1
	EventMessage                  Type = 0
	EventPlayerJoin               Type = 1
	EventPlayerLeft               Type = 2
	EventPlayerDeath              Type = 3
	EventPlayerUpdate             Type = 4
	EventPlayerList               Type = 5
	EventPlayerDimensionChange    Type = 6
	EventLog                      Type = 7
	EventFormRequest              Type = 8
	EventFormResponse             Type = 9
	EventScoreboard               Type = 10
	EventPlayerDeathCountRequest  Type = 11
	EventPlayerDeathCountResponse Type = 12
)

type HockEvent interface {
	GetType() Type
}

type hockEvent struct {
	Type Type `json:"type"`
}

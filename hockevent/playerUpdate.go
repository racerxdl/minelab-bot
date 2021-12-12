package hockevent

type PlayerUpdateEvent struct {
	PlayerStatusEvent

	Pitch   float32 `json:"pitch"`
	Yaw     float32 `json:"yaw"`
	X       float32
	Y       float32
	Z       float32
	HeadYaw float32 `json:"headYaw"`
}

func (PlayerUpdateEvent) GetType() Type {
	return EventPlayerUpdate
}

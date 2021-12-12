package hockevent

import "fmt"

type PlayerDimensionChangeEvent struct {
	PlayerStatusEvent

	Dimension int `json:"dimension"`
}

func (PlayerDimensionChangeEvent) GetType() Type {
	return EventPlayerDimensionChange
}

func (e PlayerDimensionChangeEvent) DimensionName() string {
	return DimensionName(e.Dimension)
}

func DimensionName(dimension int) string {
	switch dimension {
	case 0:
		return "overworld"
	case 1:
		return "nether"
	case 2:
		return "the end"
	default:
		return fmt.Sprintf("unknown(%d)", dimension)
	}
}

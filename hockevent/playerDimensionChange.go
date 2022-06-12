package hockevent

import (
	"fmt"
	"strings"
)

type PlayerDimensionChangeEvent struct {
	PlayerStatusEvent

	Dimension     int `json:"dimension"`
	FromDimension int `json:"fromDimension"`
}

func (PlayerDimensionChangeEvent) GetType() Type {
	return EventPlayerDimensionChange
}

func (e PlayerDimensionChangeEvent) DimensionName() string {
	return DimensionName(e.Dimension)
}

func (e PlayerDimensionChangeEvent) FromDimensionName() string {
	return DimensionName(e.FromDimension)
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

func ToDimensionId(dimension string) int {
	dimension = strings.Trim(strings.ToLower(dimension), " \r\n")
	switch dimension {
	case "overworld":
		return 0
	case "nether":
		return 1
	case "the end":
		return 2
	case "0":
		return 0
	case "1":
		return 1
	case "2":
		return 2
	default:
		return -1
	}
}

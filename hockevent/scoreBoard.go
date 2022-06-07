package hockevent

type ScoreboardEntry struct {
	EntryID   int64  `json:"entryId"`
	Type      uint8  `json:"type"`
	PID       uint64 `json:"pid"`
	Score     uint32 `json:"score"`
	ActorID   int64  `json:"actorId"`
	EntryName string `json:"entryName"`
}

type ScoreBoardEvent struct {
	Name           string            `json:"name"`
	Type           string            `json:"type"`
	Title          string            `json:"title"`
	To             string            `json:"to"`
	SortDescending bool              `json:"sortDescending"`
	Entries        []ScoreboardEntry `json:"entries"`
}

func (ScoreBoardEvent) GetType() Type {
	return EventScoreboard
}

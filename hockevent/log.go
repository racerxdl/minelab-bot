package hockevent

type LogEvent struct {
	Category     int    `json:"category"`
	Bitset       int    `json:"bitset"`
	Rules        int    `json:"rules"`
	Area         int    `json:"area"`
	Unk0         int    `json:"unk0"`
	FunctionLine int    `json:"functionLine"`
	FunctionName string `json:"functionName"`
	Message      string `json:"message"`
}

func (LogEvent) GetType() Type {
	return EventLog
}

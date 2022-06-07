package hockevent

type FormRequestEvent struct {
	FormID int            `json:"formId"`
	Data   map[string]any `json:"jsonData"`
	To     string         `json:"to"`
}

func (FormRequestEvent) GetType() Type {
	return EventFormRequest
}

type FormResponseEvent struct {
	FormID   int    `json:"formId"`
	Response string `json:"response"`
	From     string `json:"from"`
}

func (FormResponseEvent) GetType() Type {
	return EventFormResponse
}

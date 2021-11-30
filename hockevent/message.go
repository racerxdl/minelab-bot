package hockevent

type MessageType int

const (
	MESSAGE_NORMAL  MessageType = 0
	MESSAGE_SYSTEM  MessageType = 1
	MESSAGE_ANOUNCE MessageType = 2
	MESSAGE_WHISPER MessageType = 3
	MESSAGE_JUKEBOX MessageType = 4
)

type MessageEvent struct {
	From         string      `json:"from"`
	To           string      `json:"to"`
	Message      string      `json:"message"`
	MsgType      MessageType `json:"msgType"`
	Translatable bool        `json:"translatable"`
}

func (MessageEvent) GetType() Type {
	return EVENT_MESSAGE
}

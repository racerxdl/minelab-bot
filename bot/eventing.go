package bot

import "github.com/racerxdl/minelab-bot/hockevent"

func (lab *Minelab) SendMessageToPlayer(source, username, message string) {
	lab.sender <- hockevent.MessageEvent{
		From:         source,
		To:           username,
		Message:      message,
		Translatable: true,
		MsgType:      hockevent.MESSAGE_SYSTEM,
	}
}

func (lab *Minelab) BroadcastMessage(source, message string) {
	lab.sender <- hockevent.MessageEvent{
		From:         source,
		Message:      message,
		Translatable: true,
		MsgType:      hockevent.MESSAGE_SYSTEM,
	}
}

func (lab *Minelab) JukeboxMessage(source, username, message string) {
	lab.sender <- hockevent.MessageEvent{
		From:         source,
		To:           username,
		Message:      message,
		Translatable: true,
		MsgType:      hockevent.MESSAGE_JUKEBOX,
	}
}

func (lab *Minelab) RequestPlayerList() {
	lab.sender <- hockevent.PlayerListEvent{}
}

func (lab *Minelab) RequestPlayerDeathCount() {
	lab.sender <- hockevent.PlayerDeathCountRequestEvent{}
}

func (lab *Minelab) FormRequest(formID int, data map[string]any, to string) {
	lab.sender <- hockevent.FormRequestEvent{
		FormID: formID,
		Data:   data,
		To:     to,
	}
}

//TODO: Make it better
func (lab *Minelab) Scoreboard(event hockevent.ScoreBoardEvent) {
	lab.sender <- event
}

package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

type logrusHook struct {
	dg        *discordgo.Session
	channelId string
}

func NewHook(dg *discordgo.Session, guildId, logCategory, logChannel string) (logrus.Hook, error) {
	channelId, err := GetChatID(dg, guildId, logCategory, logChannel)
	if err != nil {
		return nil, err
	}

	return &logrusHook{
		dg:        dg,
		channelId: channelId,
	}, nil
}

func (l *logrusHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
	}
}

func (l *logrusHook) Fire(e *logrus.Entry) error {
	_, err := l.dg.ChannelMessageSend(l.channelId, fmt.Sprintf("[%s]: %s", e.Level, e.Message))
	return err
}

package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func GetChatID(s *discordgo.Session, guildID, chatCategory, chatChannel string) (string, error) {
	chans, err := s.GuildChannels(guildID)
	if err != nil {
		return "", err
	}
	channelCategoryId := ""
	// Search category
	for _, c := range chans {
		if c.Name == chatCategory && c.Type == discordgo.ChannelTypeGuildCategory {
			channelCategoryId = c.ID
			break
		}
	}
	if channelCategoryId == "" {
		return "", fmt.Errorf("category with name %q not found", chatCategory)
	}

	channelId := ""
	for _, c := range chans {
		if c.ParentID == channelCategoryId && c.Name == chatChannel {
			channelId = c.ID
			break
		}
	}

	if channelId == "" {
		return "", fmt.Errorf("channel with topic %s and name %s not found", chatCategory, chatChannel)
	}

	return channelId, nil
}

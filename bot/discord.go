package bot

import (
	"fmt"
	"html"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/racerxdl/minelab-bot/discord"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

func (lab *Minelab) handleOnDiscordMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID || m.Content == "" {
		return
	}

	name := m.Author.Username

	if m.Member != nil && m.Member.Nick != "" {
		name = m.Member.Nick
	}

	if lab.dedupMsg("DISCORD"+name+m.Content) == "" {
		return
	}

	if m.ChannelID == lab.discord.chatChannelId {
		name = html.EscapeString(name)
		lab.BroadcastMessage("Discord", "<"+text.Colourf("<yellow>%s</yellow>", name)+"> "+text.Colourf("<yellow>%s</yellow>", m.Content))
		return
	}

	c, _ := s.Channel(m.ChannelID)
	cc, _ := s.Channel(c.ParentID)

	cname := ""

	if cc != nil {
		cname += cc.Name + "/"
	}

	cname += c.Name

	lab.log.Debugf("[DISCORD-%s] (%s) %s > %s\n", m.GuildID, cname, name, m.Content)
	//fmt.Printf("(%s) %s > %s\n", c.Name, m.Author.Username, m.Content)
}

func (lab *Minelab) dedupMsg(msg string) string {
	lab.discord.deduplock.Lock()
	defer lab.discord.deduplock.Unlock()

	if time.Since(lab.discord.lastupdate) > time.Minute {
		lab.discord.lastmessages = nil
	}
	lab.discord.lastupdate = time.Now()

	for _, v := range lab.discord.lastmessages {
		if v == msg {
			return ""
		}
	}
	lab.discord.lastmessages = append(lab.discord.lastmessages, msg)
	if len(lab.discord.lastmessages) > 20 {
		lab.discord.lastmessages = lab.discord.lastmessages[1:]
	}
	return msg
}

func (lab *Minelab) sendDiscordChat(sourceName, message string) {
	if lab.dg == nil || lab.discord.chatChannelId == "" {
		return
	}
	msg := fmt.Sprintf("<%s> %s", sourceName, message)
	if sourceName == "" {
		msg = message
	}

	if lab.dedupMsg("GAME"+msg) == "" {
		return
	}

	_, err := lab.dg.ChannelMessageSend(lab.discord.chatChannelId, msg)
	if err != nil {
		lab.log.Errorf("error sending discord message: %s\n", err)
	}
}

func (lab *Minelab) playerSetPlaying(playerName string) {
	if lab.dg == nil {
		return
	}
	lab.discord.RLock()
	defer lab.discord.RUnlock()

	if dsuser, ok := lab.discord.cachedPlayerUsers[playerName]; ok && lab.cfg.Bot.PlayingRoleID != "" {
		err := lab.dg.GuildMemberRoleAdd(lab.cfg.Bot.GuildID, dsuser, lab.cfg.Bot.PlayingRoleID)
		if err != nil {
			lab.log.Errorf("error granting role %s to %s: %s\n", lab.cfg.Bot.PlayingRoleID, playerName, err)
		}
	}
}

func (lab *Minelab) playerUnsetPlaying(playerName string) {
	if lab.dg == nil {
		return
	}
	lab.discord.RLock()
	defer lab.discord.RUnlock()

	if dsuser, ok := lab.discord.cachedPlayerUsers[playerName]; ok && lab.cfg.Bot.PlayingRoleID != "" {
		err := lab.dg.GuildMemberRoleRemove(lab.cfg.Bot.GuildID, dsuser, lab.cfg.Bot.PlayingRoleID)
		if err != nil {
			lab.log.Errorf("error removing role %s to %s: %s\n", lab.cfg.Bot.PlayingRoleID, playerName, err)
		}
	}
}

func (lab *Minelab) cacheDiscordUsers() {
	if lab.dg == nil {
		return
	}
	lab.discord.Lock()
	lab.log.Infoln("Caching discord player users start")
	if lab.discord.cachedPlayerUsers == nil {
		lab.discord.cachedPlayerUsers = make(map[string]string)
	}
	totalUsers := len(lab.cfg.Bot.UserMap)
	if totalUsers > 0 {
		start := ""
		usersCached := 0
		for usersCached < totalUsers {
			st, err := lab.dg.GuildMembers(lab.cfg.Bot.GuildID, start, 1000)
			if err != nil {
				lab.log.Errorf("error caching discord players: %s\n", err)
				break
			}
			for _, member := range st {
				u := lab.cfg.ReverseDiscordUser(member.User.Username)
				if u != "" {
					lab.discord.cachedPlayerUsers[u] = member.User.ID
					usersCached++
				}
			}
			if len(st) > 0 {
				start = st[len(st)-1].User.ID
			} else {
				break
			}
		}
		if usersCached != totalUsers {
			lab.log.Warnf("Some users were not cached: \n")
			for k, v := range lab.cfg.Bot.UserMap {
				_, ok := lab.discord.cachedPlayerUsers[v]
				if !ok {
					lab.log.Warnf("User %q not found (XBX Name: %q)\n", v, k)
				}
			}
		}
	}
	lab.discord.Unlock()

	for k := range lab.discord.cachedPlayerUsers {
		lab.playerUnsetPlaying(k)
	}
	lab.log.Infoln("Caching discord player users finished")
}

func (lab *Minelab) discordRoutine(stop chan struct{}) {
	if lab.cfg.Bot.Token == "" { // No discord just wait
		lab.log.Warnf("No discord token found. Not starting discord routine\n")
		<-stop
		return
	}

	lab.discord.lastupdate = time.Now()

	dg, err := discordgo.New("Bot " + lab.cfg.Bot.Token)
	if err != nil {
		lab.log.Errorf("error starting discord bot: %s\n", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(lab.handleOnDiscordMessage)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		lab.log.Errorf("error opening connection, %s\n", err)
		return
	}
	lab.dg = dg

	defer func() {
		// Cleanly close down the Discord session.
		lab.dg = nil
		_ = dg.Close()
	}()

	chatId, err := discord.GetChatID(dg, lab.cfg.Bot.GuildID, lab.cfg.Bot.ChatCategory, lab.cfg.Bot.ChatChannel)
	if err != nil {
		lab.log.Errorf("error getting chat channel id: %s\n", err)
		return
	}

	lab.discord.chatChannelId = chatId

	logHook, err := discord.NewHook(dg, lab.cfg.Bot.GuildID, lab.cfg.Bot.LogCategory, lab.cfg.Bot.LogChannel)
	if err != nil {
		lab.log.Errorf("error creating hook: %s\n", err)
		return
	}

	lab.log.AddHook(logHook)

	// Wait here until CTRL-C or other term signal is received.
	lab.log.Infoln("Discord Bot is now running.")

	lab.cacheDiscordUsers()

	<-stop
}

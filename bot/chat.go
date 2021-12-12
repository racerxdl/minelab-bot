package bot

import (
	"fmt"
	"github.com/racerxdl/minelab-bot/hockevent"
	"github.com/racerxdl/minelab-bot/lang"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"regexp"
	"strings"
)

var textRegex = regexp.MustCompile(`%(.*?)[\s]`)

func parseTranslationText(txt *packet.Text) string {
	msg := lang.GetString("ptbr", txt.Message)
	msg = msg + " " // Pad so regex will work

	keys := textRegex.FindAllString(msg, -1)
	if len(keys) > 0 {
		for _, k := range keys {
			if len(k) > 5 {
				val := lang.GetString("ptbr", k[1:])
				if val != k[1:] {
					msg = strings.ReplaceAll(msg, k, val)
				}
			}
		}
	}

	// Replace params
	params := make([]string, len(txt.Parameters))
	for i, v := range txt.Parameters {
		params[i] = v
		if strings.HasPrefix(v, "%") {
			params[i] = lang.GetString("ptbr", v[1:])
		}
	}

	iparams := make([]interface{}, len(params))
	for i, v := range params {
		id := i + 1
		prefix := fmt.Sprintf("%%%d", id)
		msg = strings.Replace(msg, prefix+"$s", v, -1)
		msg = strings.Replace(msg, prefix+"$d", v, -1)
		iparams[i] = v
	}

	if strings.Contains(msg, "%") {
		msg = fmt.Sprintf(msg, iparams...)
	}

	msg = text.ANSI(msg)

	return msg
}

func (lab *Minelab) handleOtherText(client, server *minecraft.Conn, txt *packet.Text) bool {
	if txt.TextType == packet.TextTypeTranslation {
		msg := parseTranslationText(txt)
		lab.log.Infof("%s\n", msg)
		lab.sendDiscordChat("", msg)
	}
	return false
}

func (lab *Minelab) handleChat(event hockevent.MessageEvent) bool {
	msg := text.ANSI(event.Message)
	lab.log.Infof("%s> %s\n", event.From, msg)

	if strings.HasPrefix(msg, "!") || strings.HasPrefix(msg, "?") {
		// COMMAND
		lab.handleCommand(event.From, msg)
		return true
	}
	lab.sendDiscordChat(event.From, msg)

	return false
}

func (lab *Minelab) handleCommand(playerName, message string) {
	message = message[1:]
	cmd := strings.SplitN(message, " ", 2)
	if len(cmd) < 1 {
		lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Especifique o comando</red>"))
		return
	}

	dimension := lab.getPlayerDimension(playerName)

	switch cmd[0] {
	case "commands", "comandos":
		lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<bold>ondemorri</bold>, <bold>ondeta</bold>, <bold>mark</bold>"))
	case "mark":
		if len(cmd) < 2 {
			lab.SendMessageToPlayer(ServerName, playerName, "<bold>mark</bold> nome_da_marcacao")
			return
		}
		pos := lab.GetPlayerPosition(playerName)
		err := lab.db.AddPositionMark(playerName, cmd[1], dimension, *pos)
		if err != nil {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<bold>Houve um erro salvando o marcador!</bold>"))
			lab.sendDiscordChat(ServerName, fmt.Sprintf("%s tried to save marker %q but got error %q", playerName, cmd[1], err))
		} else {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<bold>Marcador %q foi salvo em X: %.0f Y: %.0f Z: %.0f em %s! Use !ondeta %q para ler novamente</bold>", cmd[1], pos.X(), pos.Y(), pos.Z(), hockevent.DimensionName(dimension), cmd[1]))
		}
	case "ondemorri", "wheredididie":
		coord, hasDied := lab.GetPlayerLastDeathPosition(playerName)
		if !hasDied {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Você ainda não morreu</red>"))
		} else {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("Você morreu em X: %.0f Y: %.0f Z: %.0f", coord.X(), coord.Y(), coord.Z()))
			lab.sendDiscordChat(ServerName, fmt.Sprintf("%s died at X: %.0f Y: %.0f Z: %.0f", playerName, coord.X(), coord.Y(), coord.Z()))
		}
	case "whereis", "ondeta":
		if len(cmd) < 2 {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Uso: !ondeta playerName/marcador</red>"))
			return
		}
		pos := lab.GetPlayerPosition(cmd[1])
		if pos != nil {
			dimen := lab.getPlayerDimension(cmd[1])
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("Jogador <bold>%s</bold> está em X: %.0f Y: %.0f Z: %.0f de %s", cmd[1], pos.X(), pos.Y(), pos.Z(), hockevent.DimensionName(dimen)))
			lab.sendDiscordChat(ServerName, fmt.Sprintf("Player %q is at X: %.0f Y: %.0f Z: %.0f", cmd[1], pos.X(), pos.Y(), pos.Z()))
			return
		}

		mark, err := lab.db.GetPositionMark(playerName, cmd[1], dimension)
		if err != nil {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Não achei um jogador/marcador %q</red>", cmd[1]))
			lab.sendDiscordChat(ServerName, fmt.Sprintf("%s tried to read marker %q but got error %q", playerName, cmd[1], err))
			return
		}
		lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("Marcador <bold>%s</bold> está em X: %.0f Y: %.0f Z: %.0f", cmd[1], mark.X(), mark.Y(), mark.Z()))
		lab.sendDiscordChat(ServerName, fmt.Sprintf("Mark %q is at X: %.0f Y: %.0f Z: %.0f", cmd[1], mark.X(), mark.Y(), mark.Z()))
		return
	case "reload":
		//lab.reloadConfig()
		//lab.reloadSound()
		lab.BroadcastMessage(ServerName, "Configuracões recarregadas")
	default:
		lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Comando inválido %q</red>", cmd[0]))
	}
}

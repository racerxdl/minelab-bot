package bot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/racerxdl/minelab-bot/hockevent"
	"github.com/racerxdl/minelab-bot/lang"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/text"
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

var commandList = [][]string{
	{"comandos", "lista de comandos"},
	{"mark", "criar marcador"},
	{"delmark", "apagar marcador"},
	{"ondemorri", "local da ultima morte"},
	{"marcadores", "marcadores do usuario"},
	{"ondeta", "localizar marcador"},
	{"mortes", "lista de mortes"},
	{"portal", "coordenada do portal na outra dimensao do local atual"},
	{"n2o", "conversor de coordenada nether 2 overworld"},
	{"o2n", "conversor de coordenada overworld 2 nether"},
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
		for _, c := range commandList {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<bold>!%s</bold>: %s", c[0], c[1]))
		}
		return
		// lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<bold>ondemorri</bold>, <bold>ondeta</bold>, <bold>mark</bold>"))
	case "delmark":
		if len(cmd) < 2 {
			lab.SendMessageToPlayer(ServerName, playerName, "<bold>delmark</bold> dimensao nome_da_marcacao")
			return
		}
		s := strings.SplitN(cmd[1], " ", 2)
		if len(s) != 2 {
			lab.SendMessageToPlayer(ServerName, playerName, "<bold>delmark</bold> dimensao nome_da_marcacao")
			return
		}
		dimen := hockevent.ToDimensionId(s[0])
		if dimen == -1 {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Dimensao invalida: %q</red>", s[0]))
			return
		}
		err := lab.db.DelPositionMark(playerName, s[1], dimen)
		if err != nil {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Erro apagando %q: %s</red>", s[1], err))
			return
		}
		lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("Marcador %q apagado!", s[1]))
		return
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
	case "bookmarks", "marcadores":
		names, dimensions, positions, err := lab.db.GetPlayerPositionMarks(playerName)
		if err != nil {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<bold>Houve um erro listando marcadores!</bold>"))
			lab.sendDiscordChat(ServerName, fmt.Sprintf("%s tried to list markers but got error %q", playerName, err))
			return
		}

		lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("Você tem <red><bold>%d</bold><red> marcadores:", len(names)))
		for i, name := range names {
			dimen := hockevent.DimensionName(dimensions[i])
			pos := positions[i]
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("- <bold>%q</bold> em <bold>%s</bold> -> X: <red>%.0f</red> Y: <red>%.0f</red> Z: <red>%.0f</red>", name, dimen, pos.X(), pos.Y(), pos.Z()))
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
	case "deaths", "mortes":
		for p, d := range lab.playerDeaths {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("- <bold>%s</bold>: <red>%d</red>", p, d))
		}
		return
	case "portal":
		pos := lab.GetPlayerPosition(playerName)
		dimen := lab.getPlayerDimension(playerName)
		tdimen := 0
		if dimen != 0 && dimen != 1 {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>!portal só funciona no nether e overworld.</red>"))
			return
		}
		x := 0
		z := 0

		if dimen == 0 {
			x = int(pos[0] / 8)
			z = int(pos[2] / 8)
			tdimen = 1
		} else {
			x = int(pos[0] * 8)
			z = int(pos[2] * 8)
		}
		sname := hockevent.DimensionName(dimen)
		dname := hockevent.DimensionName(tdimen)
		lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<bold>%s<bold> X: %d Z: %d => <bold>%s</bold> X: %d Z: %d", sname, int(pos.X()), int(pos.Z()), dname, x, z))
		return
	case "n2o":
		if len(cmd) < 2 {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Uso: !n2o X Z</red>"))
			return
		}
		s := strings.Split(cmd[1], " ")
		if len(s) != 2 {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Uso: !n2o X Z</red>"))
			return
		}
		X, errx := strconv.ParseInt(s[0], 10, 32)
		Y, erry := strconv.ParseInt(s[1], 10, 32)
		errMsg := ""
		if errx != nil {
			errMsg += fmt.Sprintf("X invalido: %s. ", errx)
		}
		if erry != nil {
			errMsg += fmt.Sprintf("Y invalido: %s. ", erry)
		}
		if errMsg != "" {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Erro: %s</red>", errMsg))
			return
		}
		nx := X * 8
		ny := Y * 8
		lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<bold>N</bold> X: %d Z: %d => <bold>O</bold> X: %d Y %d", X, Y, nx, ny))
		return
	case "o2n":
		if len(cmd) < 2 {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Uso: !o2n X Z</red>"))
			return
		}
		s := strings.Split(cmd[1], " ")
		if len(s) != 2 {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Uso: !o2n X Z</red>"))
			return
		}
		X, errx := strconv.ParseInt(s[0], 10, 32)
		Y, erry := strconv.ParseInt(s[1], 10, 32)
		errMsg := ""
		if errx != nil {
			errMsg += fmt.Sprintf("X invalido: %s. ", errx)
		}
		if erry != nil {
			errMsg += fmt.Sprintf("Y invalido: %s. ", erry)
		}
		if errMsg != "" {
			lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Erro: %s</red>", errMsg))
			return
		}
		nx := X / 8
		ny := Y / 8
		lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<bold>O</bold> X: %d Z: %d => <bold>N</bold> X: %d Y %d", X, Y, nx, ny))
		return
	case "reload":
		// lab.reloadConfig()
		//lab.reloadSound()
		lab.BroadcastMessage(ServerName, "Configuracões recarregadas")
	default:
		lab.SendMessageToPlayer(ServerName, playerName, text.Colourf("<red>Comando inválido %q</red>", cmd[0]))
	}
}

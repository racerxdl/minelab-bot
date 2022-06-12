package bot

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/racerxdl/minelab-bot/hockevent"
	"github.com/racerxdl/minelab-bot/models"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

func (lab *Minelab) handlePlayerJoin(event hockevent.PlayerJoinEvent) {
	lab.AddPlayer(event.Username, event.Xuid)
	lab.playerSetPlaying(event.Username)
	lab.log.Infof("Player %s is now online", event.Username)
}
func (lab *Minelab) handlePlayerLeft(event hockevent.PlayerLeftEvent) {
	lab.DelPlayer(event.Username)
	lab.playerUnsetPlaying(event.Username)
	lab.log.Infof("Player %s is now offline", event.Username)
}
func (lab *Minelab) handlePlayerDeath(event hockevent.PlayerDeathEvent) {
	lab.UpdatePlayerDeath(event.Username)
	lab.log.Infof("Player %s is died", event.Username)
}
func (lab *Minelab) handlePlayerUpdate(event hockevent.PlayerUpdateEvent) {
	lab.UpdatePlayerPosAbsolute(event.Username, mgl32.Vec3{
		event.X,
		event.Y,
		event.Z,
	})
}
func (lab *Minelab) handlePlayerList(event hockevent.PlayerListEvent) {
	lab.log.Infof("Received player list with %d players", len(event.Players))
	for _, player := range event.Players {
		lab.log.Infof("PlayerList(%s)", player)
		lab.AddPlayer(player, "")
		lab.playerSetPlaying(player)
	}
}

func (lab *Minelab) handlePlayerDeathCount(event hockevent.PlayerDeathCountResponseEvent) {
	// lab.log.Infof("Received player death count with %d players", len(event.PlayerDeaths))
	for player, deaths := range event.PlayerDeaths {
		// lab.log.Infof("Player %s - Deaths %d", player, deaths)
		lab.playerDeaths[player] = deaths
	}
}

func (lab *Minelab) handlePlayerDimensionChanged(event hockevent.PlayerDimensionChangeEvent) {
	changed := lab.UpdatePlayerDimension(event.Username, event.Dimension)
	lab.log.Debugf("Received DimensionChangeEvent(%s,%d) -> %t", event.Username, event.Dimension, changed)
	if changed {
		lab.log.Infof("Player %s went to %s", event.Username, event.DimensionName())
		lab.BroadcastMessage(ServerName, text.Colourf("<b><yellow>%s</yellow></b> foi para <b>%s</b>", event.Username, event.DimensionName()))
		lab.sendDiscordChat(ServerName, fmt.Sprintf("%s foi para %s", event.Username, event.DimensionName()))
	}
}

func (lab *Minelab) getPlayerDimension(playerName string) int {
	player, ok := lab.players[playerName]
	if ok {
		return player.Dimension
	}
	return 0
}

func (lab *Minelab) AddPlayer(username, xuid string) {
	lab.playerLock.Lock()
	defer lab.playerLock.Unlock()

	if _, ok := lab.players[username]; !ok {
		lab.players[username] = &models.Player{
			Username: username,
			Xuid:     xuid,
		}
	}
}

func (lab *Minelab) DelPlayer(username string) {
	lab.playerLock.Lock()
	defer lab.playerLock.Unlock()

	delete(lab.players, username)
}

func (lab *Minelab) GetPlayerLastDeathPosition(username string) (mgl32.Vec3, bool) {
	lab.playerLock.RLock()
	defer lab.playerLock.RUnlock()

	if username[0] == '@' {
		username = username[1:]
	}

	player, ok := lab.players[username]
	if !ok || player.LastDeathPosition == nil {
		return mgl32.Vec3{}, false
	}

	return *player.LastDeathPosition, true
}

func (lab *Minelab) GetPlayerPosition(username string) *mgl32.Vec3 {
	lab.playerLock.RLock()
	defer lab.playerLock.RUnlock()

	if username[0] == '@' {
		username = username[1:]
	}

	player, ok := lab.players[username]
	if !ok {
		return nil
	}

	p := player.Position

	return &p
}

func (lab *Minelab) UpdatePlayerDimension(username string, dimension int) bool {
	lab.playerLock.Lock()
	defer lab.playerLock.Unlock()

	if username[0] == '@' {
		username = username[1:]
	}

	player, ok := lab.players[username]
	if !ok {
		lab.players[username] = &models.Player{
			Username:  username,
			Dimension: dimension,
		}
		go lab.playerSetPlaying(username) // Spawn in routine because this also locks the mutex
		return true
	}
	changed := player.Dimension != dimension
	player.Dimension = dimension
	return changed
}

func (lab *Minelab) UpdatePlayerPosAbsolute(username string, pos mgl32.Vec3) {
	lab.playerLock.Lock()
	defer lab.playerLock.Unlock()

	if username[0] == '@' {
		username = username[1:]
	}

	player, ok := lab.players[username]
	if !ok {
		lab.players[username] = &models.Player{
			Username: username,
			Position: pos,
		}
		go lab.playerSetPlaying(username) // Spawn in routine because this also locks the mutex
		return
	}

	player.Position = pos
}

func (lab *Minelab) UpdatePlayerDeath(username string) {
	lab.playerLock.Lock()
	defer lab.playerLock.Unlock()

	if username[0] == '@' {
		username = username[1:]
	}

	player, ok := lab.players[username]
	if !ok {
		return
	}
	dimen := hockevent.DimensionName(player.Dimension)
	lab.db.AddPositionMark(username, fmt.Sprintf("lastdeath_%s", dimen), player.Dimension, player.Position)

	player.LastDeathPosition = &mgl32.Vec3{}
	copy(player.LastDeathPosition[:], player.Position[:])
}

func (lab *Minelab) IsPlayerDead(username string) bool {
	lab.playerLock.RLock()
	defer lab.playerLock.RUnlock()
	if player, ok := lab.players[username]; ok {
		return player.IsDead
	}
	return false
}

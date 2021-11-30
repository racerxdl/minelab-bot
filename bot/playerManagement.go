package bot

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/racerxdl/minelab-bot/hockevent"
	"github.com/racerxdl/minelab-bot/models"
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
	for _, player := range event.Players {
		lab.AddPlayer(player, "")
	}
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

	player, ok := lab.players[username]
	if !ok || player.LastDeathPosition == nil {
		return mgl32.Vec3{}, false
	}

	return *player.LastDeathPosition, true
}

func (lab *Minelab) GetPlayerPosition(username string) mgl32.Vec3 {
	lab.playerLock.RLock()
	defer lab.playerLock.RUnlock()

	player, ok := lab.players[username]
	if !ok {
		lab.log.Warnf("player %s not found\n", username)
		return mgl32.Vec3{}
	}

	return player.Position
}

func (lab *Minelab) UpdatePlayerPosAbsolute(username string, pos mgl32.Vec3) {
	lab.playerLock.Lock()
	defer lab.playerLock.Unlock()

	player, ok := lab.players[username]
	if !ok {
		lab.log.Warnf("player %s not found\n", username)
		return // No player to update
	}

	if !pos.ApproxEqual(player.Position) {
		lab.log.Debugf("%s at %v\n", username, player.Position)
	}
	player.Position = pos
}

func (lab *Minelab) UpdatePlayerDeath(username string) {
	lab.playerLock.Lock()
	defer lab.playerLock.Unlock()

	player, ok := lab.players[username]
	if !ok {
		lab.log.Warnf("player %s not found\n", username)
		return
	}

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

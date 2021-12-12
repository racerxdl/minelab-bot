package models

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Player struct {
	Username          string
	Xuid              string
	Dimension         int
	Position          mgl32.Vec3
	LastDeathPosition *mgl32.Vec3
	IsDead            bool
}

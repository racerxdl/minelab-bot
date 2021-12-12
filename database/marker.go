package database

import "github.com/go-gl/mathgl/mgl32"

type positionMarker struct {
	Username  string
	Mark      string
	Dimension int
	Position  mgl32.Vec3
}

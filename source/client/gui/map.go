package gui

import (
	"strings"
)

type GameMap struct {
	Width     int
	Height    int
	GUIPlayer *GUIPlayer
}

func NewGameMap(width, height int) *GameMap {
	return &GameMap{
		Width:     width,
		Height:    height,
		GUIPlayer: NewPlayer(0, 0),
	}
}

func (m *GameMap) Render() string {
	var sb strings.Builder
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			if x == m.GUIPlayer.X && y == m.GUIPlayer.Y {
				sb.WriteString("P ")
			} else {
				sb.WriteString(". ")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

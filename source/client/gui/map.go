package gui

import (
	"fmt"
	"strings"
)

type GUIGameMap struct {
	Width   int
	Height  int
	Players map[int]*GUIPlayer
	LocalID int
}

func NewGameMap(width, height int) *GUIGameMap {
	return &GUIGameMap{
		Width:   width,
		Height:  height,
		Players: make(map[int]*GUIPlayer),
	}
}

func (m *GUIGameMap) Render() string {
	var sb strings.Builder
	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {
			playerFound := false
			for id, player := range m.Players {
				if x == player.X && y == player.Y {
					// 写入ID
					if id == m.LocalID {
						sb.WriteString(fmt.Sprintf("%d*", id))
					} else {
						sb.WriteString(fmt.Sprintf("%d ", id))
					}

					playerFound = true
					break
				}
			}
			if !playerFound {
				sb.WriteString(". ")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

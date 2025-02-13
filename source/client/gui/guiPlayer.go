package gui

type GUIPlayer struct {
	X, Y int
}

func NewPlayer(x, y int) *GUIPlayer {
	return &GUIPlayer{X: x, Y: y}
}

func (p *GUIPlayer) Move(dx, dy int, mapWidth, mapHeight int) bool {
	newX := p.X + dx
	newY := p.Y + dy

	if newX >= 0 && newX < mapWidth && newY >= 0 && newY < mapHeight {
		p.X = newX
		p.Y = newY
		return true
	}
	return false
}

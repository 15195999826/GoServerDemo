package gui

type GUIPlayer struct {
	ID, X, Y int
}

func NewGUIPlayer(id, x, y int) *GUIPlayer {
	return &GUIPlayer{ID: id, X: x, Y: y}
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

func (p *GUIPlayer) MoveTo(x int, y int, width int, height int) {
	if x >= 0 && x < width && y >= 0 && y < height {
		p.X = x
		p.Y = y
	}
}

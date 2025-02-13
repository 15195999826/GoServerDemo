package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	mainWindow       fyne.Window
	logEntry         *widget.Entry
	startButton      *widget.Button
	stopButton       *widget.Button
	OnConfigure      func(port, tickRate, maxPlayers, heartbeat, timeSyncTimes, appointedServerTimeDelay string) error
	OnStart          func() error
	OnStop           func()
	myApp            fyne.App
	playerCountLabel *widget.Label
	uptimeLabel      *widget.Label
)

type GameWindow struct {
	window   fyne.Window
	gameMap  *GameMap
	mapLabel *widget.Label
}

func NewGameWindow() *GameWindow {
	myApp = app.New()

	mainWindow = myApp.NewWindow("Game Server Control Panel")
	gw := &GameWindow{
		window:  mainWindow,
		gameMap: NewGameMap(10, 10),
	}

	gw.mapLabel = widget.NewLabel(gw.gameMap.Render())
	gw.mapLabel.TextStyle = fyne.TextStyle{Monospace: true}

	upBtn := widget.NewButton("Up", func() {
		gw.gameMap.GUIPlayer.Move(0, -1, gw.gameMap.Width, gw.gameMap.Height)
		gw.updateMap()
	})

	downBtn := widget.NewButton("Down", func() {
		gw.gameMap.GUIPlayer.Move(0, 1, gw.gameMap.Width, gw.gameMap.Height)
		gw.updateMap()
	})

	leftBtn := widget.NewButton("Left", func() {
		gw.gameMap.GUIPlayer.Move(-1, 0, gw.gameMap.Width, gw.gameMap.Height)
		gw.updateMap()
	})

	rightBtn := widget.NewButton("Right", func() {
		gw.gameMap.GUIPlayer.Move(1, 0, gw.gameMap.Width, gw.gameMap.Height)
		gw.updateMap()
	})

	controls := container.NewGridWithColumns(2,
		container.NewHBox(),
		upBtn,
		container.NewHBox(),
		leftBtn,
		rightBtn,
		container.NewHBox(),
		downBtn,
		container.NewHBox(),
	)

	content := container.NewHSplit(
		gw.mapLabel,
		controls,
	)

	gw.window.SetContent(content)
	return gw
}

func (gw *GameWindow) updateMap() {
	gw.mapLabel.SetText(gw.gameMap.Render())
}

func (gw *GameWindow) Show() {
	gw.window.ShowAndRun()
}

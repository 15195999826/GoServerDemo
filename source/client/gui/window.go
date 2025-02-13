package gui

import (
	"log"

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
	window    fyne.Window
	gameMap   *GameMap
	mapLabel  *widget.Label
	logEntry  *widget.Entry
	nickname  *widget.Entry
	onConnect func() error
	onStart   func()
}

func NewGameWindow() *GameWindow {
	myApp := app.New()
	mainWindow := myApp.NewWindow("Game Client")

	gw := &GameWindow{
		window:  mainWindow,
		gameMap: NewGameMap(10, 10),
	}

	// Create map display
	gw.mapLabel = widget.NewLabel(gw.gameMap.Render())
	gw.mapLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Create nickname input
	gw.nickname = widget.NewEntry()
	gw.nickname.SetPlaceHolder("Enter your nickname")

	// Create control buttons
	startBtn := widget.NewButton("Start Game", func() {
		if gw.onConnect != nil {
			if err := gw.onConnect(); err != nil {
				log.Printf("Failed to connect: %v", err)
				return
			}
			if gw.onStart != nil {
				gw.onStart()
			}
		}
	})

	// Movement buttons
	upBtn := widget.NewButton("↑", func() {
		gw.gameMap.GUIPlayer.Move(0, -1, gw.gameMap.Width, gw.gameMap.Height)
		gw.updateMap()
	})

	downBtn := widget.NewButton("↓", func() {
		gw.gameMap.GUIPlayer.Move(0, 1, gw.gameMap.Width, gw.gameMap.Height)
		gw.updateMap()
	})

	leftBtn := widget.NewButton("←", func() {
		gw.gameMap.GUIPlayer.Move(-1, 0, gw.gameMap.Width, gw.gameMap.Height)
		gw.updateMap()
	})

	rightBtn := widget.NewButton("→", func() {
		gw.gameMap.GUIPlayer.Move(1, 0, gw.gameMap.Width, gw.gameMap.Height)
		gw.updateMap()
	})

	// Movement controls layout
	controls := container.NewGridWithColumns(3,
		widget.NewLabel(""), upBtn, widget.NewLabel(""),
		leftBtn, widget.NewLabel(""), rightBtn,
		widget.NewLabel(""), downBtn, widget.NewLabel(""),
	)

	// Right side panel with controls and player info
	rightPanel := container.NewVBox(
		widget.NewLabel("Player Settings"),
		gw.nickname,
		widget.NewLabel(""),
		startBtn,
		widget.NewLabel("Movement Controls"),
		controls,
	)

	// Top row with game map and controls
	topRow := container.NewHBox(
		container.NewPadded(gw.mapLabel),
		container.NewPadded(rightPanel),
	)

	// Main layout
	mainContent := container.NewVBox(
		topRow,
	)

	gw.window.SetContent(mainContent)
	gw.window.Resize(fyne.NewSize(800, 600))

	return gw
}

// // Helper methods for log handling
// func (gw *GameWindow) newScrollableLabel(text string) *widget.Entry {
// 	entry := widget.NewMultiLineEntry()
// 	entry.SetText(text)
// 	entry.Disable()
// 	entry.TextStyle = fyne.TextStyle{}
// 	entry.Wrapping = fyne.TextWrapWord

// 	entry.OnChanged = func(string) {
// 		go func() {
// 			time.Sleep(100 * time.Millisecond)
// 			entry.CursorRow = len(strings.Split(entry.Text, "\n")) - 1
// 			entry.Refresh()
// 		}()
// 	}

// 	entry.OnChanged(entry.Text)
// 	return entry
// }

// type logWriter struct {
// 	gw *GameWindow
// }

// func (w *logWriter) Write(bytes []byte) (int, error) {
// 	w.gw.writeLog(string(bytes))
// 	return len(bytes), nil
// }

// func (gw *GameWindow) writeLog(msg string) {
// 	if gw.logEntry == nil {
// 		return
// 	}

// 	timestamp := time.Now().Format("15:04:05")
// 	logMsg := fmt.Sprintf("[%s] %s", timestamp, msg)

// 	current := gw.logEntry.Text
// 	if current != "" {
// 		current += "\n"
// 	}
// 	gw.logEntry.SetText(current + strings.TrimSpace(logMsg))
// }

func (gw *GameWindow) updateMap() {
	gw.mapLabel.SetText(gw.gameMap.Render())
}

func (gw *GameWindow) Show() {
	gw.window.ShowAndRun()
}

func (gw *GameWindow) SetCallbacks(onConnect func() error, onStart func()) {
	gw.onConnect = onConnect
	gw.onStart = onStart
}

func (gw *GameWindow) GetNickname() string {
	return gw.nickname.Text
}

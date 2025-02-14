package gui

import (
	"gameproject/source/client/backend"
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
	window             fyne.Window
	gameMap            *GUIGameMap
	mapLabel           *widget.Label
	nickname           *widget.Entry
	ip                 *widget.Entry
	lastPlayerInput    *widget.Label
	nextSendInputTimer *widget.Label
	onConnect          func() error
	onStart            func()
	onMovement         func(dx, dy int) // Add movement callback
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

	gw.ip = widget.NewEntry()
	gw.ip.SetPlaceHolder("Enter server IP")
	// 默认是本地服务器
	gw.ip.SetText("127.0.0.1")

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
		if gw.onMovement != nil {
			gw.onMovement(0, 1)
		}
	})

	downBtn := widget.NewButton("↓", func() {
		if gw.onMovement != nil {
			gw.onMovement(0, -1)
		}
	})

	leftBtn := widget.NewButton("←", func() {
		if gw.onMovement != nil {
			gw.onMovement(-1, 0)
		}
	})

	rightBtn := widget.NewButton("→", func() {
		if gw.onMovement != nil {
			gw.onMovement(1, 0)
		}
	})

	// Movement controls layout
	controls := container.NewGridWithColumns(3,
		widget.NewLabel(""), upBtn, widget.NewLabel(""),
		leftBtn, widget.NewLabel(""), rightBtn,
		widget.NewLabel(""), downBtn, widget.NewLabel(""),
	)

	// Right side panel with controls and player info
	controlPanel := container.NewVBox(
		widget.NewLabel(""),
		startBtn,
		widget.NewLabel("Movement Controls"),
		controls,
	)

	settingsPanel := container.NewVBox(
		widget.NewLabel("Player Settings"),
		gw.nickname,
		gw.ip,
	)

	gw.lastPlayerInput = widget.NewLabel("无输入")
	gw.lastPlayerInput.TextStyle = fyne.TextStyle{Monospace: true}

	gw.nextSendInputTimer = widget.NewLabel("-1")
	gw.nextSendInputTimer.TextStyle = fyne.TextStyle{Monospace: true}

	gameStatePanel := container.NewVBox(
		widget.NewLabel("当前输入:"),
		gw.lastPlayerInput,
		widget.NewLabel("下次发送计时器:"),
		gw.nextSendInputTimer,
	)

	// Top row with game map and controls
	topRow := container.NewHBox(
		container.NewPadded(gw.mapLabel),
		container.NewPadded(controlPanel),
		container.NewPadded(settingsPanel),
		container.NewPadded(gameStatePanel),
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

func (gw *GameWindow) SetCallbacks(onConnect func() error, onStart func(), onMovement func(dx, dy int)) {
	gw.onConnect = onConnect
	gw.onStart = onStart
	gw.onMovement = onMovement
}

func (gw *GameWindow) GetNickname() string {
	return gw.nickname.Text
}

func (gw *GameWindow) BindLocalPlayer(localID int) {
	gw.gameMap.LocalID = localID
}

func (gw *GameWindow) UpdatePlayers(player *backend.Player) {
	// 检查gw.gameMap.Players中是否已经存在该玩家
	if _, ok := gw.gameMap.Players[player.ID]; !ok {
		gw.gameMap.Players[player.ID] = NewGUIPlayer(player.ID, player.Position.X, player.Position.Y)
	} else {
		gw.gameMap.Players[player.ID].MoveTo(player.Position.X, player.Position.Y, gw.gameMap.Width, gw.gameMap.Height)
	}

	gw.updateMap()
}

package gui

import (
	"fmt"
	"image/color"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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

// ServerConfig holds the server configuration values
type ServerConfig struct {
	Port                     string
	TickRate                 string
	MaxPlayers               string
	HeartbeatInterval        string
	TimeSyncTimes            string
	AppointedServerTimeDelay string
}

// 创建可滚动到底部的多行文本框
func newScrollableLabel(text string) *widget.Entry {
	entry := widget.NewMultiLineEntry()
	entry.SetText(text)
	entry.Disable()
	entry.TextStyle = fyne.TextStyle{}
	entry.Wrapping = fyne.TextWrapWord

	// 使用 OnChanged 回调确保文本更新时滚动到底部
	entry.OnChanged = func(string) {
		// 使用 goroutine 确保在渲染完成后滚动
		go func() {
			time.Sleep(100 * time.Millisecond)
			entry.CursorRow = len(strings.Split(entry.Text, "\n")) - 1
			entry.Refresh()
		}()
	}

	// 触发一次 OnChanged 以执行初始滚动
	entry.OnChanged(entry.Text)
	return entry
}

// SetServerCallbacks sets the callback functions for server control
func SetServerCallbacks(configure func(port, tickRate, maxPlayers, heartbeat, timeSysncTimes, appointedServerTimeDelay string) error,
	start func() error,
	stop func()) {
	OnConfigure = configure
	OnStart = start
	OnStop = stop
}

// UpdatePlayerCount updates the player count display
func UpdatePlayerCount(count int) {
	if playerCountLabel != nil {
		playerCountLabel.SetText(fmt.Sprintf("Current Players: %d", count))
	}
}

// UpdateUptime updates the server uptime display
func UpdateUptime(duration time.Duration) {
	if uptimeLabel != nil {
		uptimeLabel.SetText(fmt.Sprintf("Server Uptime: %s", duration.Round(time.Second)))
	}
}

// CreateWindow creates and configures the window but doesn't start the event loop
func CreateWindow() {
	myApp = app.New()
	myApp.Settings().SetTheme(&customTheme{theme.DefaultTheme()})
	mainWindow = myApp.NewWindow("Game Server Control Panel")

	config := &ServerConfig{
		Port:                     "12345",
		TickRate:                 "50",
		MaxPlayers:               "100",
		HeartbeatInterval:        "5",
		TimeSyncTimes:            "10",
		AppointedServerTimeDelay: "3",
	}

	// Configuration section
	portEntry := widget.NewEntry()
	portEntry.SetText(config.Port)
	tickRateEntry := widget.NewEntry()
	tickRateEntry.SetText(config.TickRate)
	maxPlayersEntry := widget.NewEntry()
	maxPlayersEntry.SetText(config.MaxPlayers)
	heartbeatEntry := widget.NewEntry()
	heartbeatEntry.SetText(config.HeartbeatInterval)
	timeSyncTimesEntry := widget.NewEntry()
	timeSyncTimesEntry.SetText(config.TimeSyncTimes)
	appointedServerTimeDelayEntry := widget.NewEntry()
	appointedServerTimeDelayEntry.SetText(config.AppointedServerTimeDelay)

	configBox := container.NewGridWithColumns(2,
		widget.NewLabel("Port:"),
		portEntry,
		widget.NewLabel("Tick Rate (ms):"),
		tickRateEntry,
		widget.NewLabel("Max Players:"),
		maxPlayersEntry,
		widget.NewLabel("Heartbeat Interval (s):"),
		heartbeatEntry,
		widget.NewLabel("时间同步次数:"),
		timeSyncTimesEntry,
		widget.NewLabel("等待游戏开始时间 (s):"),
		appointedServerTimeDelayEntry,
	)

	// Control buttons
	startButton = widget.NewButton("Start Server", func() {
		if OnConfigure != nil {
			err := OnConfigure(portEntry.Text, tickRateEntry.Text, maxPlayersEntry.Text, heartbeatEntry.Text, timeSyncTimesEntry.Text, appointedServerTimeDelayEntry.Text)
			if err != nil {
				writeLog("Configuration error: " + err.Error())
				return
			}
		}

		if OnStart != nil {
			if err := OnStart(); err != nil {
				writeLog("Start error: " + err.Error())
				return
			}
		}
		startButton.Disable()
		stopButton.Enable()
	})

	stopButton = widget.NewButton("Stop Server", func() {
		if OnStop != nil {
			OnStop()
		}
		startButton.Enable()
		stopButton.Disable()
	})
	stopButton.Disable()

	buttonBox := container.NewVBox(startButton, stopButton)

	// Server stats
	playerCountLabel = widget.NewLabel("Players: 0")
	statsBox := container.NewVBox(
		playerCountLabel,
	)

	gameStateLabel := widget.NewLabel("游戏状态: 未开始")
	uptimeLabel = widget.NewLabel("游戏时间: 0s")
	statsBox2 := container.NewVBox(
		gameStateLabel,
		uptimeLabel,
	)

	// Combine controls and stats with padding
	controlsContainer := container.NewHBox(
		buttonBox,
		widget.NewLabel("    "), // Add some spacing
		statsBox,
		widget.NewLabel("    "), // Add some spacing
		statsBox2,
	)

	// Log output
	logEntry = newScrollableLabel("")

	// Redirect standard logger to our GUI
	log.SetOutput(&logWriter{})

	logScroll := container.NewScroll(logEntry)
	logScroll.SetMinSize(fyne.NewSize(0, 300)) // 设置最小高度为200

	// Main layout
	mainContainer := container.NewVBox(
		widget.NewCard("Server Configuration", "", configBox),
		widget.NewCard("Controls", "", controlsContainer),
		widget.NewCard("Server Logs", "", logScroll),
	)

	mainWindow.SetContent(mainContainer)
	mainWindow.Resize(fyne.NewSize(900, 600))
	mainWindow.Show()
}

// RunWindow starts the main event loop
func RunWindow() {
	if mainWindow != nil {
		mainWindow.ShowAndRun()
	}
}

// logWriter implements io.Writer to redirect logs to the GUI
type logWriter struct{}

func (w *logWriter) Write(bytes []byte) (int, error) {
	writeLog(string(bytes))
	return len(bytes), nil
}

// writeLog adds a new log entry to the log window
func writeLog(msg string) {
	if logEntry == nil {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	logMsg := fmt.Sprintf("[%s] %s", timestamp, msg)

	current := logEntry.Text
	if current != "" {
		current += "\n"
	}
	logEntry.SetText(current + strings.TrimSpace(logMsg))

	// Scroll to bottom
	logEntry.CursorRow = len(strings.Split(logEntry.Text, "\n")) - 1
	logEntry.Refresh()
}

// 自定义主题来覆盖禁用状态的文本颜色
type customTheme struct {
	fyne.Theme
}

func (t *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameDisabled:
		// Entry禁用状态显示浅灰色文字
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
		// case theme.ColorNameInputBackground:
		// Entry的背景色为黑色
		// return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
		// case theme.ColorNamePlaceHolder, theme.ColorNameInputBorder:
		// 	// Entry的占位符文字和边框颜色
		// 	return color.NRGBA{R: 160, G: 160, B: 160, A: 255}
		// default:
		// 	// 其他所有颜色保持默认
		// 	return t.Theme.Color(name, variant)
	}

	return t.Theme.Color(name, variant)
}

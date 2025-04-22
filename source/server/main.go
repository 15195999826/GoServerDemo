package main

import (
	"gameproject/source/server/backend"
	"gameproject/source/server/gui"
)

func main() {
	server := backend.NewGameServer()

	// Setup GUI callbacks
	gui.SetServerCallbacks(
		func(port, tickRate, maxPlayers, heartbeat, timeSysncTimes, appointedServerTimeDelay, sendInputInterval, executionDuration string) error {
			return server.Configure(port, tickRate, maxPlayers, heartbeat, timeSysncTimes, appointedServerTimeDelay, sendInputInterval, executionDuration)
		},
		func() error {
			return server.Start()
		},
		func() {
			server.Stop()
		},
	)

	// Create and run GUI
	gui.CreateWindow()
	gui.RunWindow() // 使用新的 RunWindow 函数
}

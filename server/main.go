package main

import (
	"gameproject/server/backend"
	"gameproject/server/gui"
)

func main() {
	server := backend.NewGameServer()

	// Setup GUI callbacks
	gui.SetServerCallbacks(
		func(port, tickRate, maxPlayers, heartbeat, timeSysncTimes, appointedServerTimeDelay string) error {
			return server.Configure(port, tickRate, maxPlayers, heartbeat, timeSysncTimes, appointedServerTimeDelay)
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

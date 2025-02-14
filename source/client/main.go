package main

import (
	"fmt"
	"log"

	"gameproject/source/client/backend"
	"gameproject/source/client/gui"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("\n程序发生错误,按回车键退出...")
			fmt.Scanln()
		}
	}()

	var client *backend.GameClient
	mainWindow := gui.NewGameWindow()

	// Set up callbacks
	mainWindow.SetCallbacks(
		// Connect callback
		func() error {
			client = backend.NewGameClient()
			client.SetOnPlayersUpdate(func(player *backend.Player) {
				mainWindow.UpdatePlayers(player)
			})

			client.SetBindLocalPlayer(func(localID int) {
				mainWindow.BindLocalPlayer(localID)
			})
			if err := client.Connect(); err != nil {
				return err
			}
			log.Println("Connected to server")
			return nil
		},
		// Start callback
		func() {
			go client.Start()
		},
		// Movement callback
		func(dx, dy int) {
			if client != nil {
				client.SendMovement(dx, dy)
			}
		},
	)

	mainWindow.Show()
}

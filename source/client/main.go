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
			// 保持窗口打开
			fmt.Println("\n程序发生错误,按回车键退出...")
			fmt.Scanln()
		}
	}()

	mainWindow := gui.NewGameWindow()
	mainWindow.Show()

	client := backend.NewGameClient()
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	log.Println("Connected to server")
	client.Start()
}

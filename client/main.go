package main

import (
	"fmt"
	"log"

	"gameproject/client/backend"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			// 保持窗口打开
			fmt.Println("\n程序发生错误，请查看error.log文件获取详细信息")
			fmt.Println("按回车键退出...")
			fmt.Scanln()
		}
	}()

	client := backend.NewGameClient()
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	log.Println("Connected to server")
	client.Start()
}

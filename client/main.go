package main

import (
	"log"

	"gameproject/client/backend"
)

func main() {
	client := backend.NewGameClient()
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	log.Println("Connected to server")
	client.Start()
}

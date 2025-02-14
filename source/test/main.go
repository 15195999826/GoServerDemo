package main

import (
	"gameproject/source/gametypes"
	"gameproject/source/serialization"
	"gameproject/source/test/fbtest"
	"log"
	"time"
)

func benchmarkTimeNow(iterations int) time.Duration {
	start := time.Now()
	for i := 0; i < iterations; i++ {
		time.Now()
	}
	return time.Since(start)
}

func main() {
	// 演示序列化
	serializedData := fbtest.SerializeMessage()

	// 演示反序列化
	fbtest.DeserializeMessage(serializedData)

	// 测试ticker
	// gameTicker := time.NewTicker(time.Second / 60)
	// defer gameTicker.Stop()

	// for range gameTicker.C {
	// 	// do something
	// 	log.Println("do something")
	// }

	// 测试time.Now()的性能
	iterations := 1000000
	duration := benchmarkTimeNow(iterations)
	log.Printf("Time.Now() called %d times took %v", iterations, duration)
	log.Printf("Average time per call: %v", duration/time.Duration(iterations))

	// 测试时间
	t := time.Second / time.Duration(20)
	log.Println(t.Milliseconds())

	fbtest.TestWorldSync()
	fbtest.TestPlayerInput()

	// 测试序列化玩家数据
	testPlayers := []gametypes.SerializePlayer{
		{
			ID:       1,
			Position: gametypes.Vector2Int{X: 100, Y: 200},
		},
		{
			ID:       2,
			Position: gametypes.Vector2Int{X: 300, Y: 400},
		},
		{
			ID:       3,
			Position: gametypes.Vector2Int{X: 500, Y: 600},
		},
	}

	testStartEnterGame := gametypes.StartEnterGame{
		Players: testPlayers,
	}

	startEnterGame := serialization.SerializeS2CStartEnterGame(&testStartEnterGame)

	retStartEnterGame := serialization.DeserializeS2CStartEnterGame(startEnterGame)
	log.Println("反序列化结果:%v", retStartEnterGame)
}

package main

import (
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
}

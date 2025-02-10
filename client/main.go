package main

import (
	"log"
	"time"

	"gameproject/GameProtocol"

	flatbuffers "github.com/google/flatbuffers/go"
	kcp "github.com/xtaci/kcp-go/v5"
)

type GameClient struct {
	conn *kcp.UDPSession
}

func NewGameClient() *GameClient {
	return &GameClient{}
}

func (c *GameClient) Connect() error {
	conn, err := kcp.DialWithOptions("127.0.0.1:12345", nil, 0, 0)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *GameClient) Start() {
	// 启动接收消息的goroutine
	go c.receiveMessages()
	// 定期发送心跳
	heartbeatTicker := time.NewTicker(1 * time.Second)
	go func() {
		for range heartbeatTicker.C {
			builder := flatbuffers.NewBuilder(1024)
			GameProtocol.MessageStart(builder)
			GameProtocol.MessageAddType(builder, GameProtocol.MessageTypeHeartbeat)
			message := GameProtocol.MessageEnd(builder)
			builder.Finish(message)
			data := builder.FinishedBytes()
			if _, err := c.conn.Write(data); err != nil {
				log.Println("Heartbeat error:", err)
				return
			}
		}
	}()

	// 定期发送更新
	ticker := time.NewTicker(50 * time.Millisecond)
	for range ticker.C {
		c.sendUpdate()
	}
}

func (c *GameClient) receiveMessages() {
	buffer := make([]byte, 1024)
	for {
		n, err := c.conn.Read(buffer)
		if err != nil {
			log.Println("Read error:", err)
			return
		}

		// 解析接收到的消息
		message := GameProtocol.GetRootAsMessage(buffer[:n], 0)

		// 根据消息类型处理
		switch message.Type() {
		case GameProtocol.MessageTypeGameState:
			// 处理游戏状态更新
			log.Println("Received game state update")
		}
	}
}

func (c *GameClient) sendUpdate() {
	builder := flatbuffers.NewBuilder(1024)

	// 创建玩家位置
	GameProtocol.Vector2Start(builder)
	GameProtocol.Vector2AddX(builder, 100.0)
	GameProtocol.Vector2AddY(builder, 100.0)
	position := GameProtocol.Vector2End(builder)

	// 创建玩家状态
	name := builder.CreateString("Player1")
	GameProtocol.PlayerStateStart(builder)
	GameProtocol.PlayerStateAddId(builder, 1)
	GameProtocol.PlayerStateAddPosition(builder, position)
	GameProtocol.PlayerStateAddName(builder, name)
	playerState := GameProtocol.PlayerStateEnd(builder)

	// 创建消息payload
	GameProtocol.MessageStart(builder)
	GameProtocol.MessageAddType(builder, GameProtocol.MessageTypePlayerMove)
	GameProtocol.MessageAddPayload(builder, playerState)
	message := GameProtocol.MessageEnd(builder)

	builder.Finish(message)

	// 发送消息
	data := builder.FinishedBytes()
	_, err := c.conn.Write(data)
	if err != nil {
		log.Println("Write error:", err)
	}
}

func (c *GameClient) Close() {
	if c.conn != nil {
		log.Println("Closing client connection...")
		c.conn.Close()
	}
}

func main() {
	client := NewGameClient()
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	log.Println("Connected to server")
	client.Start()
}

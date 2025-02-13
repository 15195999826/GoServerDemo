package backend

import (
	"fmt"
	"gameproject/fb"
	"log"
	"runtime/debug"
	"time"

	"github.com/xtaci/kcp-go"
)

type GameState int

const (
	Invalid GameState = iota
	Room
	Game
	GameOver
)

func (s GameState) String() string {
	return [...]string{"Invalid", "Room", "Game", "GameOver"}[s]
}

type GameClient struct {
	conn          *kcp.UDPSession
	commandSender *CommandSender

	heartbeatInterval time.Duration

	timeSyncedTimes          int
	systemTimeDiffWithServer time.Duration
	alreadyTimeSyncTimes     int
	lastSendTime             time.Time

	rtt time.Duration

	gameState GameState

	playerID int

	gameStartTime time.Time
}

func NewGameClient() *GameClient {
	client := &GameClient{
		gameState:            Invalid,
		alreadyTimeSyncTimes: 0,
	}
	client.commandSender = NewCommandSender()
	return client
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
	// 创建错误通道
	errChan := make(chan error, 1)

	// 启动接收消息的goroutine
	go func() {
		if err := c.receiveMessages(); err != nil {
			errChan <- err
		}
	}()

	// 定期发送心跳
	heartbeatTicker := time.NewTicker(1 * time.Second)
	defer heartbeatTicker.Stop()

	// 游戏主循环ticker
	gameTicker := time.NewTicker(time.Second / 60)
	defer gameTicker.Stop()

	for {
		select {
		case err := <-errChan:
			log.Printf("Error in receive messages: %v", err)
			c.Close()
			fmt.Println("\n程序发生错误, 按回车键退出...")
			fmt.Scanln()
			return
		case <-heartbeatTicker.C:
			c.commandSender.SendPing(c.conn)
		case <-gameTicker.C:
			c.Tick()
		}
	}
}

func (c *GameClient) receiveMessages() error {
	buffer := make([]byte, 1024)
	for {
		n, err := c.conn.Read(buffer)
		if err != nil {
			log.Println("Read error:", err)
			return fmt.Errorf("connection read error: %w", err)
		}

		// 解析接收到的消息
		s2cCommand := fb.GetRootAsS2CCommand(buffer[:n], 0)
		if s2cCommand == nil {
			return fmt.Errorf("failed to parse S2CCommand")
		}

		// 根据消息类型处理
		if err := c.handleMessage(s2cCommand); err != nil {
			return fmt.Errorf("message handling error: %w", err)
		}
	}
}

// 新增处理消息的方法
func (c *GameClient) handleMessage(s2cCommand *fb.S2CCommand) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			err = fmt.Errorf("panic in handleMessage:%v\nStack trace:\n%s", r, stack)
		}

	}()
	// 根据消息类型处理
	switch s2cCommand.Command() {
	default:
		log.Println("Unknown command from server:", s2cCommand.Command())
	case fb.ServerCommandS2C_COMMAND_PONG:
	case fb.ServerCommandS2C_COMMAND_ENTERROOM:
		enterRoom := fb.GetRootAsS2CEnterRoom(s2cCommand.BodyBytes(), 0)
		c.playerID = int(enterRoom.PlayerId())
		heartbeatInterval := float32(enterRoom.HeartbeatInterval()) / 2 // 这里用一般的时间发送Ping
		c.heartbeatInterval = time.Duration(heartbeatInterval) * time.Second
		c.timeSyncedTimes = int(enterRoom.TimeSyncTimes())
		log.Printf("Enter room, player id: %d, heartbeat interval: %v, time sync times: %d", c.playerID, c.heartbeatInterval, c.timeSyncedTimes)

	case fb.ServerCommandS2C_COMMAND_STARTENTERGAME:
	case fb.ServerCommandS2C_COMMAND_STARTGAME:
	case fb.ServerCommandS2C_COMMAND_WORLDSYNC:
	case fb.ServerCommandS2C_COMMAND_RESPONSETIME:
		c.alreadyTimeSyncTimes++
		responseTime := fb.GetRootAsS2CResponseTime(s2cCommand.BodyBytes(), 0)
		thzTimeRTT := time.Since(c.lastSendTime)
		serverTime := time.Unix(0, responseTime.ServerTime())
		thzTimeSystemTimeDiffWithServer := time.Until(serverTime.Add(-thzTimeRTT / 2))
		log.Printf("Response time, rtt: %v, server time: %v, system time diff with server: %v", thzTimeRTT, serverTime, thzTimeSystemTimeDiffWithServer)
		// rtt记录为平均值
		c.rtt = (c.rtt*time.Duration(c.alreadyTimeSyncTimes-1) + thzTimeRTT) / time.Duration(c.alreadyTimeSyncTimes)
		// 系统时间与服务器时间的差值记录为平均值
		c.systemTimeDiffWithServer = (c.systemTimeDiffWithServer*time.Duration(c.alreadyTimeSyncTimes-1) + thzTimeSystemTimeDiffWithServer) / time.Duration(c.alreadyTimeSyncTimes)
		log.Printf("AVG, rtt: %v, system time diff with server: %v", c.rtt, c.systemTimeDiffWithServer)
	}

	return nil
}

func (c *GameClient) Tick() {
	switch c.gameState {
	case Invalid:
	case Room:
		if c.alreadyTimeSyncTimes < c.timeSyncedTimes {
			c.commandSender.SendRequestTime(c.conn)
			c.lastSendTime = time.Now()
		}
	case Game:
	case GameOver:
	default:
		log.Println("未处理的 game state")
	}
}

func (c *GameClient) Close() {
	if c.conn != nil {
		log.Println("Closing client connection...")
		c.conn.Close()
		c.conn = nil
	}
}

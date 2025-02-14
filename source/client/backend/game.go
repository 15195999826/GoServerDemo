package backend

import (
	"fmt"
	"gameproject/fb"
	"gameproject/source/gametypes"
	"gameproject/source/serialization"
	"log"
	"runtime/debug"
	"time"

	"math/rand/v2"

	"github.com/xtaci/kcp-go"
)

type GameState int

const (
	Invalid GameState = iota
	Room
	GameCountDown // 游戏开始前倒计时阶段
	Game
	GameOver
)

func (s GameState) String() string {
	return [...]string{"Invalid", "Room", "GameCountDown", "Game", "GameOver"}[s]
}

type Player struct {
	ID       int
	Position gametypes.Vector2Int
}

type GameClient struct {
	conn *kcp.UDPSession

	heartbeatInterval time.Duration

	timeSyncedTimes          int
	systemTimeDiffWithServer int64
	alreadyTimeSyncTimes     int
	lastSendSyncTime         time.Time

	rtt time.Duration

	gameState GameState

	playerID int
	players  map[int]*Player

	desiredGameStartTime int64
	gameStartTime        time.Time

	gameMap           *gametypes.GameMap
	logicFrame        int
	bUpdateLogicFrame bool
	desiredLogicFrame int
	lastPlayInput     *gametypes.PlayerInput
	lastSendInputTime time.Time

	syncInputQueue []gametypes.PlayerInput

	// 回调函数
	bindLocalPlayer func(localID int)
	onPlayerUpdate  func(players *Player)
}

func NewGameClient() *GameClient {
	client := &GameClient{
		gameState:            Invalid,
		alreadyTimeSyncTimes: 0,
		logicFrame:           0,
		players:              make(map[int]*Player),
		syncInputQueue:       make([]gametypes.PlayerInput, 0),
	}

	client.gameMap = gametypes.NewGameMap(10, 10)
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
			sendPing(c.conn)
		case tickTime := <-gameTicker.C:
			c.tick(tickTime)
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
		c.bindLocalPlayer(c.playerID)
		heartbeatInterval := float32(enterRoom.HeartbeatInterval()) / 2 // 这里用一般的时间发送Ping
		c.heartbeatInterval = time.Duration(heartbeatInterval) * time.Second
		c.timeSyncedTimes = int(enterRoom.TimeSyncTimes()) // 首次请求看起来会存在冷启动的问题， 首次不计入平均值
		log.Printf("Enter room, player id: %d, heartbeat interval: %v, time sync times: %d", c.playerID, c.heartbeatInterval, c.timeSyncedTimes)
		c.gameState = Room

	case fb.ServerCommandS2C_COMMAND_STARTENTERGAME:
		startEntetGame := serialization.DeserializeS2CStartEnterGame(s2cCommand.BodyBytes())
		log.Printf("[StartEnterGame] players: %v", startEntetGame.Players)

		// 创建客户端本地角色
		for _, player := range startEntetGame.Players {
			// 检查是否已经存在， 存在的话存在逻辑错误， 抛出错误
			if _, ok := c.players[player.ID]; ok {
				return fmt.Errorf("Player %d already exists, current players: %v", player.ID, c.players)
			}
			c.players[player.ID] = &Player{
				ID:       player.ID,
				Position: player.Position,
			}

			// 通知UI更新玩家
			if c.onPlayerUpdate != nil {
				c.onPlayerUpdate(c.players[player.ID])
			}
		}

		// 模拟加载， 随机延迟后发送消息
		go func() {
			time.Sleep(time.Duration(0.5+float64(rand.IntN(2))) * time.Second)
			sendGameLoaded(c.conn)
		}()
	case fb.ServerCommandS2C_COMMAND_STARTGAME:
		startGame := fb.GetRootAsS2CStartGame(s2cCommand.BodyBytes(), 0)
		c.desiredGameStartTime = startGame.AppointedServerTime() + c.systemTimeDiffWithServer
		c.gameState = GameCountDown
	case fb.ServerCommandS2C_COMMAND_WORLDSYNC:
		worldSync := serialization.DeserializeWorldSync(s2cCommand.BodyBytes())
		c.bUpdateLogicFrame = true
		c.desiredLogicFrame = int(worldSync.LogicFrame)
	case fb.ServerCommandS2C_COMMAND_RESPONSETIME:
		c.alreadyTimeSyncTimes++
		responseTime := fb.GetRootAsS2CResponseTime(s2cCommand.BodyBytes(), 0)
		thzTimeRTT := time.Since(c.lastSendSyncTime)
		serverTime := responseTime.ServerTime()
		// int64
		thzTimeSystemTimeDiffWithServer := time.Now().UnixMilli() - serverTime

		log.Printf("Response time, rtt: %d ms, server time: %v ms, system time diff with server: %v ms", thzTimeRTT.Milliseconds(), serverTime, thzTimeSystemTimeDiffWithServer)
		if c.alreadyTimeSyncTimes > 1 {
			// rtt记录为平均值
			c.rtt = (c.rtt*time.Duration(c.alreadyTimeSyncTimes-2) + thzTimeRTT) / time.Duration(c.alreadyTimeSyncTimes-1)
			// 系统时间与服务器时间的差值记录为平均值
			c.systemTimeDiffWithServer = (c.systemTimeDiffWithServer*int64(c.alreadyTimeSyncTimes-2) + thzTimeSystemTimeDiffWithServer) / int64(c.alreadyTimeSyncTimes-1)
			log.Printf("AVG, rtt: %d ms, system time diff with server: %v ms", c.rtt.Milliseconds(), c.systemTimeDiffWithServer)
		}
	case fb.ServerCommandS2C_COMMAND_PLAYERINPUTSYNC:
		playerInput := serialization.DeserializePlayerInput(s2cCommand.BodyBytes())
		c.syncInputQueue = append(c.syncInputQueue, playerInput)
	}

	return nil
}

func (c *GameClient) tick(tickTime time.Time) {
	switch c.gameState {
	case Invalid:
	case Room:
		if c.alreadyTimeSyncTimes < c.timeSyncedTimes {
			sendRequestTime(c.conn)
			c.lastSendSyncTime = time.Now()
		}
	case GameCountDown:
		if time.Now().UnixMilli() >= c.desiredGameStartTime {
			c.gameStartTime = time.Now()
			c.gameState = Game
			c.lastSendInputTime = tickTime
		}
	case Game:
		// Todo: 在UE中实现时， 使用游戏时间累加计算， 在服务端使用系统时间
		// log.Printf("[%v]Game running...", tickTime.UnixMilli()-c.gameStartTime.UnixMilli())
		// 每2秒发送自己的输入
		if tickTime.Sub(c.lastSendInputTime) >= 2*time.Second && c.lastPlayInput != nil {
			sendPlayerInput(c.conn, c.lastPlayInput)
			c.lastPlayInput = nil
			c.lastSendInputTime = tickTime
		}

		if c.bUpdateLogicFrame {
			c.logicFrame = c.desiredLogicFrame
			c.bUpdateLogicFrame = false
		}

		// 处理收到的玩家输入，按顺序执行那些帧号小于当前逻辑帧的输入
		if len(c.syncInputQueue) > 0 {
			var remainingInputs []gametypes.PlayerInput
			for _, input := range c.syncInputQueue {
				if int(input.LogicFrame) <= c.logicFrame {
					// 执行输入
					if player, ok := c.players[input.ID]; ok {
						// 计算新位置
						newPos := player.Position
						switch input.CommandType {
						case gametypes.MoveLeft:
							newPos.X--
						case gametypes.MoveRight:
							newPos.X++
						case gametypes.MoveUp:
							newPos.Y--
						case gametypes.MoveDown:
							newPos.Y++
						}

						// 检查新位置是否在地图范围内
						if newPos.X >= 0 && newPos.X < c.gameMap.MapData.Width &&
							newPos.Y >= 0 && newPos.Y < c.gameMap.MapData.Height {
							// 更新位置
							player.Position = newPos

							// 通知UI更新玩家位置
							if c.onPlayerUpdate != nil {
								c.onPlayerUpdate(player)
							}
						}
					} else {
						log.Printf("[Error]tick Player %d not found", input.ID)
					}
				} else {
					// 将未处理的输入保存回队列
					remainingInputs = append(remainingInputs, input)
				}
			}

			// 更新输入队列，只保留未处理的输入
			c.syncInputQueue = remainingInputs
		}
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

func (c *GameClient) SendMovement(dx, dy int) error {
	if dx == 0 && dy == 0 {
		// 错误的输入
		err := fmt.Errorf("错误的指令")
		return err
	}

	var inputType gametypes.PlayerCommandType

	if dx > 0 {
		inputType = gametypes.MoveRight
	} else if dx < 0 {
		inputType = gametypes.MoveLeft
	} else if dy > 0 {
		inputType = gametypes.MoveUp
	} else if dy < 0 {
		inputType = gametypes.MoveDown
	}
	c.lastPlayInput = &gametypes.PlayerInput{
		ID:          c.playerID,
		LogicFrame:  c.logicFrame,
		CommandType: inputType,
	}

	return nil
}

func (c *GameClient) SetOnPlayersUpdate(callback func(player *Player)) {
	c.onPlayerUpdate = callback
}

func (c *GameClient) SetBindLocalPlayer(f func(localID int)) {
	c.bindLocalPlayer = f
}

package backend

import (
	"context"
	"fmt"
	"gameproject/fb"
	"gameproject/source/gametypes"
	"gameproject/source/serialization"
	"log"
	"math/rand/v2"
	"strconv"
	"sync"
	"time"

	"github.com/xtaci/kcp-go/v5"
)

type GameState int

const (
	Room GameState = iota
	WaitPlayersReady
	GameCountDown // 游戏开始前倒计时阶段
	Game
	GameOver
)

func (s GameState) String() string {
	return [...]string{"Room", "WaitPlayersReady", "GameCountDown", "Game", "GameOver"}[s]
}

type GameServer struct {
	players  map[int]*Player
	nextID   int
	listener *kcp.Listener
	config   *ServerConfig
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup

	gameState     GameState
	appointedTime int64

	gameMap      *gametypes.GameMap
	frameCounter int
	logicFrame   int
	inputQueue   []gametypes.PlayerInput
}

type ServerConfig struct {
	Port                     int
	TickRate                 int
	MaxPlayers               int
	HeartbeatInterval        time.Duration
	TimeSyncTimes            int
	AppointedServerTimeDelay time.Duration
	SendInputInterval        float32
	ExecutionDuration        float32
}

type Player struct {
	id              int
	conn            *kcp.UDPSession
	lastActive      time.Time
	timeSyncedTimes int
	isReady         bool
	position        gametypes.Vector2Int
}

func NewGameServer() *GameServer {
	ctx, cancel := context.WithCancel(context.Background())

	server := &GameServer{
		players:   make(map[int]*Player),
		nextID:    1,
		ctx:       ctx,
		cancel:    cancel,
		gameState: Room,
		gameMap:   gametypes.NewGameMap(10, 10),
	}

	return server
}

func (s *GameServer) Configure(port, tickRate, maxPlayers, heartbeat, timeSysncTimes, appointedServerTimeDelay, sendInputInterval, executionDuration string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	t, err := strconv.Atoi(tickRate)
	if err != nil {
		return err
	}

	m, err := strconv.Atoi(maxPlayers)
	if err != nil {
		return err
	}

	h, err := strconv.Atoi(heartbeat)
	if err != nil {
		return err
	}

	ts, err := strconv.Atoi(timeSysncTimes)
	if err != nil {
		return err
	}

	a, err := strconv.Atoi(appointedServerTimeDelay)
	if err != nil {
		return err
	}

	// 转化为float
	sf, err := strconv.ParseFloat(sendInputInterval, 64)
	if err != nil {
		return err
	}

	// 转化为float
	execf, err := strconv.ParseFloat(executionDuration, 64)
	if err != nil {
		return err
	}
	// 根据TickRate计算发送输入间隔帧数

	s.config = &ServerConfig{
		Port:                     p,
		TickRate:                 t,
		MaxPlayers:               m,
		HeartbeatInterval:        time.Duration(h) * time.Second,
		TimeSyncTimes:            ts,
		AppointedServerTimeDelay: time.Duration(a) * time.Second,
		SendInputInterval:        float32(sf),
		ExecutionDuration:        float32(execf),
	}
	return nil
}

func (s *GameServer) Start() error {
	if s.config == nil {
		return fmt.Errorf("server not configured")
	}

	var err error
	s.listener, err = kcp.ListenWithOptions(fmt.Sprintf(":%d", s.config.Port), nil, 0, 0)
	if err != nil {
		return err
	}

	// Game tick routine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(time.Second / time.Duration(s.config.TickRate))
		defer ticker.Stop()

		for {
			select {
			case tickTime := <-ticker.C:
				s.tick(tickTime)
			case <-s.ctx.Done():
				return
			}
		}
	}()

	// Heartbeat routine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.config.HeartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkHeartbeats()
			case <-s.ctx.Done():
				return
			}
		}
	}()

	// Accept connections routine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				conn, err := s.listener.AcceptKCP()
				if err != nil {
					if s.ctx.Err() != nil {
						return // Server is shutting down
					}
					log.Println("Accept error:", err)
					continue
				}

				if len(s.players) >= s.config.MaxPlayers {
					log.Printf("Rejected connection: server full")
					// Todo: 通知客户端连接失败
					conn.Close()
					continue
				}

				player := &Player{
					id:              s.nextID,
					conn:            conn,
					timeSyncedTimes: 0,
					isReady:         false,
				}
				s.nextID++
				s.players[player.id] = player

				// 创建进入房间消息，并发送给该玩家
				sendEnterRoomMessage(player, s)
				// _, err := player.conn.Write(data)
				// if err != nil {
				// 	log.Printf("Failed to send update to player %d: %v", player.id, err)
				// }
				s.wg.Add(1)
				go func() {
					defer s.wg.Done()
					s.handlePlayer(player)
				}()
			}
		}
	}()

	log.Printf("Server started on port %d", s.config.Port)
	return nil
}

func (s *GameServer) Stop() {
	log.Println("Stopping server...")
	s.cancel()
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
	log.Println("Server stopped")
}

func (s *GameServer) checkHeartbeats() {
	now := time.Now()
	disconnected := make([]int, 0)

	for id, player := range s.players {
		if now.Sub(player.lastActive) > 2*s.config.HeartbeatInterval {
			log.Printf("Player %d timeout", id)
			disconnected = append(disconnected, id)
		}
	}

	for _, id := range disconnected {
		if player, ok := s.players[id]; ok {
			player.conn.Close()
			delete(s.players, id)
		}
	}
}

func (s *GameServer) tick(tickTime time.Time) {
	// 打印tickTime time.Time, 通道中拿取的时间跟timeNow可能存在1s的误差
	// log.Printf("Tick at %v, TimeNow: %v", tickTime.UnixMilli(), time.Now().UnixMilli())
	switch s.gameState {
	case Room:
		// 如果房间人数满了，则开始游戏
		if len(s.players) == s.config.MaxPlayers {
			// 检查是否全部完成了校时
			allSynced := true
			for _, player := range s.players {
				if player.timeSyncedTimes < s.config.TimeSyncTimes {
					allSynced = false
				}
			}
			if allSynced {
				// 给每个玩家随机一个不重复的出生位置
				s.assignPlayerPositions()
				sendStartEnterGame(s)
				s.gameState = WaitPlayersReady
			}
		}
	case WaitPlayersReady:
		// 检查所有玩家是否准备好
		allReady := true
		for _, player := range s.players {
			if !player.isReady {
				allReady = false
				break
			}
		}
		if allReady {
			// 计算约定的游戏开始时间（当前时间 + 延迟时间）
			s.appointedTime = time.Now().Add(s.config.AppointedServerTimeDelay).UnixMilli()
			sendStartGame(s)
			s.gameState = GameCountDown
		}
	case GameCountDown:
		// 检查是否到达约定的游戏开始时间
		if tickTime.UnixMilli() >= s.appointedTime {
			log.Printf("Game start At:%v, AppointedTime:%v", tickTime.UnixMilli(), s.appointedTime)
			s.gameState = Game
			s.frameCounter = 0
			s.logicFrame = 0
		}
	case Game:
		// 游戏逻辑， 服务端目前只做指令转发
		s.frameCounter++
		logicFrameUpdated := false
		s.logicFrame++

		// 目前配置下，相当于每0.5秒进行一次world sync
		if s.frameCounter == s.config.TickRate/2 {
			s.frameCounter = 0
			logicFrameUpdated = true
		}

		// 筛选当前需要执行的命令， 服务端执行简单逻辑， 目前只计算位置， Todo: 可以考虑同步玩家位置状态做客户端校验
		var validInputs []gametypes.PlayerInput
		var remainingInputs []gametypes.PlayerInput

		for _, input := range s.inputQueue {
			if input.LogicFrame <= s.logicFrame {
				validInputs = append(validInputs, input)
			} else {
				remainingInputs = append(remainingInputs, input)
			}
		}

		s.inputQueue = remainingInputs
		if len(remainingInputs) != 0 {
			// 打印剩余输入
			log.Printf("[%d] 异常剩余玩家输入 remaining inputs: %v", s.logicFrame, remainingInputs)
		}

		// 服务端更新玩家位置
		if len(validInputs) != 0 {
			// Todo: 服务端更新玩家位置, 以后实现
		}

		if logicFrameUpdated {
			sendWorldSync(s)
		}

	default:
		return
	}
}

func (s *GameServer) handlePlayer(player *Player) {
	// 初始化最后活动时间
	player.lastActive = time.Now()

	// 在函数开始时记录客户端连接
	log.Printf("Player %d (%s) connected", player.id, player.conn.RemoteAddr())

	// 确保在函数返回时清理玩家
	defer func() {
		log.Printf("Player %d (%s) disconnected", player.id, player.conn.RemoteAddr())
		delete(s.players, player.id)
		player.conn.Close()
	}()

	buffer := make([]byte, 1024)
	for {
		n, err := player.conn.Read(buffer)
		if err != nil {
			log.Printf("Player %d connection error: %v", player.id, err)
			return
		}
		// 更新最后活动时间
		player.lastActive = time.Now()

		// 解析接收到的消息
		c2sCommand := fb.GetRootAsC2SCommand(buffer[:n], 0)

		// 根据消息类型处理
		switch c2sCommand.Command() {
		case fb.ClientCommandC2S_COMMAND_PING:
			// 返回Pong
			sendPong(player)
		case fb.ClientCommandC2S_COMMAND_REQUESTTIME:
			sendResponseTime(player)
			player.timeSyncedTimes++
		case fb.ClientCommandC2S_COMMAND_PLAYERINFO:
			// Todo: 更新玩家信息
		case fb.ClientCommandC2S_COMMAND_GAMELOADED:
			// 更新玩家准备状态
			player.isReady = true
		case fb.ClientCommandC2S_COMMAND_PLAYERINPUT:
			// 玩家输入存入缓存队列
			playerInput := serialization.DeserializePlayerInput(c2sCommand.BodyBytes())
			log.Printf("Player %d input: %v", player.id, playerInput)
			s.inputQueue = append(s.inputQueue, playerInput)
			// Todo: 目前直接转发, 以后考虑是否增加跟当前逻辑帧的校验关系
			sendPlayerInput(s, &playerInput)
		default:
			log.Printf("Unknown command from player %d: %d", player.id, c2sCommand.Command())
		}
	}
}

func (s *GameServer) assignPlayerPositions() {
	positions := make(map[int]*gametypes.Vector2Int)
	availablePositions := make([]gametypes.Vector2Int, 0)

	// Calculate center offsets
	centerX := s.gameMap.MapData.Width / 2
	centerY := s.gameMap.MapData.Height / 2

	// Create list of all possible positions
	// Excluding extreme edges for better gameplay
	for x := 1; x < s.gameMap.MapData.Width-1; x++ {
		for y := 1; y < s.gameMap.MapData.Height-1; y++ {
			// Convert to centered coordinate system where (0,0) is the center
			centeredX := x - centerX
			centeredY := y - centerY
			availablePositions = append(availablePositions, gametypes.Vector2Int{X: centeredX, Y: centeredY})
		}
	}

	// Randomly assign positions to players
	for playerID := range s.players {
		if len(availablePositions) == 0 {
			log.Printf("Warning: No more positions available for player %d", playerID)
			continue
		}

		// Pick random position from available positions
		idx := rand.IntN(len(availablePositions))
		pos := availablePositions[idx]

		// Remove used position by swapping with last element and shrinking slice
		availablePositions[idx] = availablePositions[len(availablePositions)-1]
		availablePositions = availablePositions[:len(availablePositions)-1]

		positions[playerID] = &gametypes.Vector2Int{X: pos.X, Y: pos.Y}
		log.Printf("Assigned position (%v, %v) to player %d", pos.X, pos.Y, playerID)
	}

	// 更新玩家位置
	for playerID, position := range positions {
		if player, ok := s.players[playerID]; ok {
			player.position = *position
		}
	}
}

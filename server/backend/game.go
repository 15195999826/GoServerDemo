package backend

import (
	"context"
	"fmt"
	"gameproject/fb"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/xtaci/kcp-go/v5"
)

type GameState int

const (
	Room GameState = iota
	WaitPlayersReady
	Game
	GameOver
)

func (s GameState) String() string {
	return [...]string{"Room", "WaitPlayersReady", "Game", "GameOver"}[s]
}

type GameServer struct {
	players       map[int]*Player
	nextID        int
	listener      *kcp.Listener
	config        *ServerConfig
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	commandSender *ServerCommandSender

	gameState GameState
}

type ServerConfig struct {
	Port                     int
	TickRate                 time.Duration
	MaxPlayers               int
	HeartbeatInterval        time.Duration
	TimeSyncTimes            int
	AppointedServerTimeDelay time.Duration
}

type Player struct {
	id              int
	conn            *kcp.UDPSession
	lastActive      time.Time
	timeSyncedTimes int
	isReady         bool
	position        struct{ x, y float32 }
}

func NewGameServer() *GameServer {
	ctx, cancel := context.WithCancel(context.Background())

	server := &GameServer{
		players:   make(map[int]*Player),
		nextID:    1,
		ctx:       ctx,
		cancel:    cancel,
		gameState: Room,
	}

	server.commandSender = NewServerCommandSender()
	return server
}

func (s *GameServer) Configure(port, tickRate, maxPlayers, heartbeat, timeSysncTimes, appointedServerTimeDelay string) error {
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

	s.config = &ServerConfig{
		Port:                     p,
		TickRate:                 time.Duration(t) * time.Millisecond,
		MaxPlayers:               m,
		HeartbeatInterval:        time.Duration(h) * time.Second,
		TimeSyncTimes:            ts,
		AppointedServerTimeDelay: time.Duration(a) * time.Second,
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
		ticker := time.NewTicker(s.config.TickRate)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.update()
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
				s.commandSender.SendEnterRoomMessage(player, s)
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

func (s *GameServer) update() {
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
				s.commandSender.SendStartEnterGame(s)
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
			s.commandSender.SendStartGame(s)
			s.gameState = Game
		}
	case Game:
		// 游戏逻辑， 服务端目前只做指令转发
	default:
		return
	}

	// // 创建游戏状态并广播
	// builder := flatbuffers.NewBuilder(1024)

	// // 创建玩家状态数组
	// playerStates := make([]flatbuffers.UOffsetT, 0, len(s.players))
	// for id, player := range s.players {
	// 	// 创建位置
	// 	GameProtocol.Vector2Start(builder)
	// 	GameProtocol.Vector2AddX(builder, player.position.x)
	// 	GameProtocol.Vector2AddY(builder, player.position.y)
	// 	pos := GameProtocol.Vector2End(builder)

	// 	// 创建玩家名称
	// 	name := builder.CreateString("Player" + string(id))

	// 	// 创建玩家状态
	// 	GameProtocol.PlayerStateStart(builder)
	// 	GameProtocol.PlayerStateAddId(builder, int32(player.id))
	// 	GameProtocol.PlayerStateAddPosition(builder, pos)
	// 	GameProtocol.PlayerStateAddName(builder, name)
	// 	playerState := GameProtocol.PlayerStateEnd(builder)

	// 	playerStates = append(playerStates, playerState)
	// }

	// // 创建玩家状态数组
	// GameProtocol.GameStateStartPlayersVector(builder, len(playerStates))
	// for i := len(playerStates) - 1; i >= 0; i-- {
	// 	builder.PrependUOffsetT(playerStates[i])
	// }
	// players := builder.EndVector(len(playerStates))

	// // 创建游戏状态
	// GameProtocol.GameStateStart(builder)
	// GameProtocol.GameStateAddPlayers(builder, players)
	// GameProtocol.GameStateAddTick(builder, 0)
	// gameState := GameProtocol.GameStateEnd(builder)

	// // 创建消息
	// GameProtocol.MessageStart(builder)
	// GameProtocol.MessageAddType(builder, GameProtocol.MessageTypeWorldSync)
	// GameProtocol.MessageAddPayload(builder, gameState)
	// message := GameProtocol.MessageEnd(builder)

	// builder.Finish(message)

	// // 广播给所有玩家
	// data := builder.FinishedBytes()
	// for _, player := range s.players {
	// 	_, err := player.conn.Write(data)
	// 	if err != nil {
	// 		log.Printf("Failed to send update to player %d: %v", player.id, err)
	// 	}
	// }
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
			s.commandSender.SendPong(player)
		case fb.ClientCommandC2S_COMMAND_REQUESTTIME:
			s.commandSender.SendResponseTime(player)
			player.timeSyncedTimes++
		case fb.ClientCommandC2S_COMMAND_PLAYERINFO:
			// Todo: 更新玩家信息
		case fb.ClientCommandC2S_COMMAND_GAMELOADED:
			// 更新玩家准备状态
			player.isReady = true
		case fb.ClientCommandC2S_COMMAND_PLAYERINPUT:
			// Todo: 玩家输入存入缓存队列
		default:
			log.Printf("Unknown command from player %d: %d", player.id, c2sCommand.Command())
		}
	}
}

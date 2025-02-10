package main

import (
	"gameproject/GameProtocol"
	"log"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/xtaci/kcp-go"
)

type GameServer struct {
	players map[int]*Player
	nextID  int
}

type Player struct {
	id         int
	conn       *kcp.UDPSession
	lastActive time.Time
	position   struct{ x, y float32 }
}

func NewGameServer() *GameServer {
	return &GameServer{
		players: make(map[int]*Player),
		nextID:  1,
	}
}

func (s *GameServer) Start() {
	listener, err := kcp.ListenWithOptions(":12345", nil, 0, 0)
	if err != nil {
		log.Fatal(err)
	}

	// 启动游戏tick
	ticker := time.NewTicker(50 * time.Millisecond)
	go func() {
		for range ticker.C {
			s.update()
		}
	}()

	// 启动心跳检测
	heartbeatTicker := time.NewTicker(5 * time.Second)
	go func() {
		for range heartbeatTicker.C {
			now := time.Now()
			// 使用临时map来避免在遍历时修改
			disconnected := make([]int, 0)

			for id, player := range s.players {
				if now.Sub(player.lastActive) > 2*time.Second {
					log.Printf("Player %d (%s) timeout", id, player.conn.RemoteAddr())
					disconnected = append(disconnected, id)
				}
			}

			// 清理断开的连接
			for _, id := range disconnected {
				if player, ok := s.players[id]; ok {
					player.conn.Close()
					delete(s.players, id)
				}
			}
		}
	}()

	log.Println("Server started on :12345")
	for {
		conn, err := listener.AcceptKCP()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}

		player := &Player{
			id:   s.nextID,
			conn: conn,
		}

		s.nextID++
		s.players[player.id] = player

		go s.handlePlayer(player)
	}
}

func (s *GameServer) update() {
	// 创建游戏状态并广播
	builder := flatbuffers.NewBuilder(1024)

	// 创建玩家状态数组
	playerStates := make([]flatbuffers.UOffsetT, 0, len(s.players))
	for id, player := range s.players {
		// 创建位置
		GameProtocol.Vector2Start(builder)
		GameProtocol.Vector2AddX(builder, player.position.x)
		GameProtocol.Vector2AddY(builder, player.position.y)
		pos := GameProtocol.Vector2End(builder)

		// 创建玩家名称
		name := builder.CreateString("Player" + string(id))

		// 创建玩家状态
		GameProtocol.PlayerStateStart(builder)
		GameProtocol.PlayerStateAddId(builder, int32(player.id))
		GameProtocol.PlayerStateAddPosition(builder, pos)
		GameProtocol.PlayerStateAddName(builder, name)
		playerState := GameProtocol.PlayerStateEnd(builder)

		playerStates = append(playerStates, playerState)
	}

	// 创建玩家状态数组
	GameProtocol.GameStateStartPlayersVector(builder, len(playerStates))
	for i := len(playerStates) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(playerStates[i])
	}
	players := builder.EndVector(len(playerStates))

	// 创建游戏状态
	GameProtocol.GameStateStart(builder)
	GameProtocol.GameStateAddPlayers(builder, players)
	GameProtocol.GameStateAddTick(builder, 0)
	gameState := GameProtocol.GameStateEnd(builder)

	// 创建消息
	GameProtocol.MessageStart(builder)
	GameProtocol.MessageAddType(builder, GameProtocol.MessageTypeGameState)
	GameProtocol.MessageAddPayload(builder, gameState)
	message := GameProtocol.MessageEnd(builder)

	builder.Finish(message)

	// 广播给所有玩家
	data := builder.FinishedBytes()
	for _, player := range s.players {
		_, err := player.conn.Write(data)
		if err != nil {
			log.Printf("Failed to send update to player %d: %v", player.id, err)
		}
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
		message := GameProtocol.GetRootAsMessage(buffer[:n], 0)

		// 根据消息类型处理
		switch message.Type() {
		case GameProtocol.MessageTypePlayerMove:
			// 解析玩家状态

		}
	}
}

func main() {
	server := NewGameServer()
	server.Start()
}

package backend

import (
	"gameproject/fb"
	"log"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
)

type ServerCommandSender struct {
}

func NewServerCommandSender() *ServerCommandSender {
	return &ServerCommandSender{}
}

func (ms *ServerCommandSender) SendPong(player *Player) error {
	builder := flatbuffers.NewBuilder(1024)

	// 创建 S2CCommand
	fb.S2CCommandStart(builder)
	fb.S2CCommandAddCommand(builder, fb.ServerCommandS2C_COMMAND_PONG)
	fb.S2CCommandAddStatus(builder, fb.S2CStatusS2C_STATUS_SUCCESS)
	command := fb.S2CCommandEnd(builder)

	builder.Finish(command)
	data := builder.FinishedBytes()

	_, err := player.conn.Write(data)
	if err != nil {
		log.Printf("Failed to send pong message to player %d: %v", player.id, err)
		return err
	}
	return nil
}

func (ms *ServerCommandSender) SendResponseTime(player *Player) error {
	builder := flatbuffers.NewBuilder(1024)

	// 创建 S2CResponseTime
	fb.S2CResponseTimeStart(builder)
	fb.S2CResponseTimeAddServerTime(builder, time.Now().Unix())
	responseTime := fb.S2CResponseTimeEnd(builder)

	// 创建 S2CCommand
	fb.S2CCommandStart(builder)
	fb.S2CCommandAddCommand(builder, fb.ServerCommandS2C_COMMAND_RESPONSETIME)
	fb.S2CCommandAddStatus(builder, fb.S2CStatusS2C_STATUS_SUCCESS)
	fb.S2CCommandAddBody(builder, responseTime)
	command := fb.S2CCommandEnd(builder)

	builder.Finish(command)
	data := builder.FinishedBytes()

	_, err := player.conn.Write(data)
	if err != nil {
		log.Printf("Failed to send response time message to player %d: %v", player.id, err)
		return err
	}
	return nil
}

// SendEnterRoomMessage 发送进入房间消息
func (ms *ServerCommandSender) SendEnterRoomMessage(player *Player, server *GameServer) error {
	builder := flatbuffers.NewBuilder(1024)

	// 创建 S2CEnterRoom
	fb.S2CEnterRoomStart(builder)
	fb.S2CEnterRoomAddPlayerId(builder, int32(player.id))
	fb.S2CEnterRoomAddTimeSyncTimes(builder, int32(server.config.TimeSyncTimes))
	fb.S2CEnterRoomAddHeartbeatInterval(builder, int32(server.config.HeartbeatInterval.Seconds()))
	enterRoom := fb.S2CEnterRoomEnd(builder)

	// 创建 S2CCommand
	fb.S2CCommandStart(builder)
	fb.S2CCommandAddCommand(builder, fb.ServerCommandS2C_COMMAND_ENTERROOM)
	fb.S2CCommandAddStatus(builder, fb.S2CStatusS2C_STATUS_SUCCESS)
	fb.S2CCommandAddBody(builder, enterRoom)
	command := fb.S2CCommandEnd(builder)

	builder.Finish(command)
	data := builder.FinishedBytes()

	_, err := player.conn.Write(data)
	if err != nil {
		log.Printf("Failed to send enter room message to player %d: %v", player.id, err)
		return err
	}
	return nil
}

func (ms *ServerCommandSender) SendStartEnterGame(server *GameServer) error {
	builder := flatbuffers.NewBuilder(1024)

	// 创建 S2CStartEnterGame
	fb.S2CStartEnterGameStart(builder)
	// Todo: 告知客户端其他所有玩家的数据
	startEnterGame := fb.S2CStartEnterGameEnd(builder)

	// 创建 S2CCommand
	fb.S2CCommandStart(builder)
	fb.S2CCommandAddCommand(builder, fb.ServerCommandS2C_COMMAND_STARTENTERGAME)
	fb.S2CCommandAddStatus(builder, fb.S2CStatusS2C_STATUS_SUCCESS)
	fb.S2CCommandAddBody(builder, startEnterGame)
	command := fb.S2CCommandEnd(builder)

	builder.Finish(command)
	data := builder.FinishedBytes()

	// 广播给所有玩家
	for _, player := range server.players {
		_, err := player.conn.Write(data)
		if err != nil {
			log.Printf("Failed to send start enter game message to player %d: %v", player.id, err)
			return err
		}
	}

	log.Printf("Sent start enter game message to all players")
	return nil
}

func (ms *ServerCommandSender) SendStartGame(server *GameServer) error {
	builder := flatbuffers.NewBuilder(1024)

	// 计算约定的游戏开始时间（当前时间 + 延迟时间）
	appointedTime := time.Now().Add(server.config.AppointedServerTimeDelay).Unix()

	// 创建 S2CStartGame
	fb.S2CStartGameStart(builder)
	fb.S2CStartGameAddAppointedServerTime(builder, appointedTime)
	startGame := fb.S2CStartGameEnd(builder)

	// 创建 S2CCommand
	fb.S2CCommandStart(builder)
	fb.S2CCommandAddCommand(builder, fb.ServerCommandS2C_COMMAND_STARTGAME)
	fb.S2CCommandAddStatus(builder, fb.S2CStatusS2C_STATUS_SUCCESS)
	fb.S2CCommandAddBody(builder, startGame)
	command := fb.S2CCommandEnd(builder)

	builder.Finish(command)
	data := builder.FinishedBytes()

	// 广播给所有玩家
	for _, player := range server.players {
		_, err := player.conn.Write(data)
		if err != nil {
			log.Printf("Failed to send start game message to player %d: %v", player.id, err)
			return err
		}
	}

	log.Printf("Sent start game message to all players, game will start at Unix time: %d", appointedTime)
	return nil
}

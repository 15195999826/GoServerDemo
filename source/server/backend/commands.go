package backend

import (
	"gameproject/fb"
	"log"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
)

func SendPong(player *Player) error {
	data := createS2CCommand(fb.ServerCommandS2C_COMMAND_PONG, fb.S2CStatusS2C_STATUS_SUCCESS, 0, "", nil)
	_, err := player.conn.Write(data)
	if err != nil {
		log.Printf("Failed to send pong message to player %d: %v", player.id, err)
		return err
	}
	return nil
}

func SendResponseTime(player *Player) error {
	builder := flatbuffers.NewBuilder(1024)

	// 创建 S2CResponseTime
	fb.S2CResponseTimeStart(builder)
	fb.S2CResponseTimeAddServerTime(builder, time.Now().UnixMilli())
	responseTimeOffset := fb.S2CResponseTimeEnd(builder)

	builder.Finish(responseTimeOffset)
	bodyBytes := builder.FinishedBytes()

	// 创建 S2CCommand
	data := createS2CCommand(fb.ServerCommandS2C_COMMAND_RESPONSETIME, fb.S2CStatusS2C_STATUS_SUCCESS, 0, "", bodyBytes)

	_, err := player.conn.Write(data)
	if err != nil {
		log.Printf("Failed to send response time message to player %d: %v", player.id, err)
		return err
	}
	return nil
}

// SendEnterRoomMessage 发送进入房间消息
func SendEnterRoomMessage(player *Player, server *GameServer) error {
	builder := flatbuffers.NewBuilder(1024)

	// 创建 S2CEnterRoom
	fb.S2CEnterRoomStart(builder)
	fb.S2CEnterRoomAddPlayerId(builder, int32(player.id))
	fb.S2CEnterRoomAddTimeSyncTimes(builder, int32(server.config.TimeSyncTimes))
	fb.S2CEnterRoomAddHeartbeatInterval(builder, int32(server.config.HeartbeatInterval.Seconds()))
	enterRoomOffset := fb.S2CEnterRoomEnd(builder)

	builder.Finish(enterRoomOffset)
	bodyBytes := builder.FinishedBytes()

	data := createS2CCommand(fb.ServerCommandS2C_COMMAND_ENTERROOM, fb.S2CStatusS2C_STATUS_SUCCESS, 0, "", bodyBytes)

	_, err := player.conn.Write(data)
	if err != nil {
		log.Printf("Failed to send enter room message to player %d: %v", player.id, err)
		return err
	}
	return nil
}

func SendStartEnterGame(server *GameServer) error {
	builder := flatbuffers.NewBuilder(1024)

	// 创建 S2CStartEnterGame
	fb.S2CStartEnterGameStart(builder)
	// Todo: 告知客户端其他所有玩家的数据
	startEnterGameOffset := fb.S2CStartEnterGameEnd(builder)

	builder.Finish(startEnterGameOffset)
	bodyBytes := builder.FinishedBytes()
	// 创建 S2CCommand
	data := createS2CCommand(fb.ServerCommandS2C_COMMAND_STARTENTERGAME, fb.S2CStatusS2C_STATUS_SUCCESS, 0, "", bodyBytes)

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

func SendStartGame(server *GameServer) error {
	builder := flatbuffers.NewBuilder(1024)

	// 计算约定的游戏开始时间（当前时间 + 延迟时间）
	appointedTime := time.Now().Add(server.config.AppointedServerTimeDelay).UnixMilli()

	// 创建 S2CStartGame
	fb.S2CStartGameStart(builder)
	fb.S2CStartGameAddAppointedServerTime(builder, appointedTime)
	startGameOffset := fb.S2CStartGameEnd(builder)

	builder.Finish(startGameOffset)
	bodyBytes := builder.FinishedBytes()

	// 创建 S2CCommand
	data := createS2CCommand(fb.ServerCommandS2C_COMMAND_STARTGAME, fb.S2CStatusS2C_STATUS_SUCCESS, 0, "", bodyBytes)

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

func createS2CCommand(command fb.ServerCommand, status fb.S2CStatus, code int64, message string, body []byte) []byte {
	builder := flatbuffers.NewBuilder(1024)

	// 创建message字符串
	var messageOffset flatbuffers.UOffsetT
	if message != "" {
		messageOffset = builder.CreateString(message)
	}

	// 创建body字节数组
	var bodyOffset flatbuffers.UOffsetT
	if body != nil {
		bodyOffset = builder.CreateByteVector(body)
	}

	// 开始构建S2CCommand
	fb.S2CCommandStart(builder)
	fb.S2CCommandAddCommand(builder, command)
	fb.S2CCommandAddStatus(builder, status)
	fb.S2CCommandAddCode(builder, code)
	if message != "" {
		fb.S2CCommandAddMessage(builder, messageOffset)
	}
	if body != nil {
		fb.S2CCommandAddBody(builder, bodyOffset)
	}
	rootOffset := fb.S2CCommandEnd(builder)

	// 完成构建
	builder.Finish(rootOffset)
	return builder.FinishedBytes()
}

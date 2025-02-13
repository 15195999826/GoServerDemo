package fbtest

import (
	"fmt"
	"gameproject/fb"

	flatbuffers "github.com/google/flatbuffers/go"
)

func SerializeMessage() []byte {
	// 1. 首先序列化内部的S2CEnterRoom
	builder := flatbuffers.NewBuilder(1024)
	fb.S2CEnterRoomStart(builder)
	fb.S2CEnterRoomAddPlayerId(builder, 12345)
	fb.S2CEnterRoomAddTimeSyncTimes(builder, 3)
	fb.S2CEnterRoomAddHeartbeatInterval(builder, 30)
	enterRoomOffset := fb.S2CEnterRoomEnd(builder)

	// 将S2CEnterRoom序列化为字节数组
	builder.Finish(enterRoomOffset)
	bodyBytes := builder.FinishedBytes()

	// 2. 使用通用函数构建外层S2CCommand
	return createCommand(
		fb.ServerCommandS2C_COMMAND_ENTERROOM,
		fb.S2CStatusS2C_STATUS_SUCCESS,
		0,
		"成功进入房间",
		bodyBytes,
	)
}

func DeserializeMessage(buf []byte) {
	// 解析外层S2CCommand
	command := fb.GetRootAsS2CCommand(buf, 0)

	fmt.Printf("Command: %v\n", command.Command())
	fmt.Printf("Status: %v\n", command.Status())
	fmt.Printf("Code: %v\n", command.Code())
	fmt.Printf("Message: %v\n", command.Message())

	// 获取body字节数组
	bodyBytes := command.BodyBytes()
	if len(bodyBytes) > 0 {
		// 解析内层S2CEnterRoom
		enterRoom := fb.GetRootAsS2CEnterRoom(bodyBytes, 0)

		fmt.Printf("Player ID: %v\n", enterRoom.PlayerId())
		fmt.Printf("Time Sync Times: %v\n", enterRoom.TimeSyncTimes())
		fmt.Printf("Heartbeat Interval: %v\n", enterRoom.HeartbeatInterval())
	}
}

func createCommand(command fb.ServerCommand, status fb.S2CStatus, code int64, message string, body []byte) []byte {
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

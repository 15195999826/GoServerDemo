package fbtest

import (
	"fmt"
	"gameproject/fb"
	"gameproject/source/gametypes"
	"gameproject/source/serialization"

	flatbuffers "github.com/google/flatbuffers/go"
)

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

func TestPlayerInput() {
	// 构造测试数据
	commands := []gametypes.PlayerCommand{
		{
			CommandType: gametypes.UseAbility,
			AbilityID:   101,
			Position: gametypes.Vector2Int{
				X: 200,
				Y: 300,
			},
			CustomStr: "技能1",
		},
	}

	testData := gametypes.PlayerInput{
		ID:         12345,
		LogicFrame: 100,
		Commands:   commands,
	}

	// 序列化
	data := serialization.SerializePlayerInput(&testData)
	fmt.Println("序列化数据长度:", len(data))

	// 反序列化
	fmt.Println("\n反序列化数据:")
	deserializedData := serialization.DeserializePlayerInput(data)

	// 打印基本信息
	fmt.Printf("玩家ID: %v\n", deserializedData.ID)
	fmt.Printf("逻辑帧: %v\n", deserializedData.LogicFrame)
	fmt.Printf("命令数量: %v\n", len(deserializedData.Commands))

	// 打印每个命令的详细信息
	for i, cmd := range deserializedData.Commands {
		fmt.Printf("\n命令 #%d:\n", i+1)
		fmt.Printf("  命令类型: %v\n", cmd.CommandType)
		fmt.Printf("  技能ID: %v\n", cmd.AbilityID)
		fmt.Printf("  目标位置: (%v, %v)\n", cmd.Position.X, cmd.Position.Y)
		fmt.Printf("  自定义字符串: %v\n", cmd.CustomStr)
	}

	// 验证数据是否一致
	fmt.Println("\n验证结果:")
	if deserializedData.ID != testData.ID {
		fmt.Printf("错误: 玩家ID不匹配. 预期 %v, 实际 %v\n", testData.ID, deserializedData.ID)
	}
	if deserializedData.LogicFrame != testData.LogicFrame {
		fmt.Printf("错误: 逻辑帧不匹配. 预期 %v, 实际 %v\n", testData.LogicFrame, deserializedData.LogicFrame)
	}
	if len(deserializedData.Commands) != len(testData.Commands) {
		fmt.Printf("错误: 命令数量不匹配. 预期 %v, 实际 %v\n", len(testData.Commands), len(deserializedData.Commands))
	} else {
		for i, origCmd := range testData.Commands {
			desCmd := deserializedData.Commands[i]

			if desCmd.CommandType != origCmd.CommandType {
				fmt.Printf("错误: 命令 #%d 命令类型不匹配. 预期 %v, 实际 %v\n", i+1, origCmd.CommandType, desCmd.CommandType)
			}
			if desCmd.AbilityID != origCmd.AbilityID {
				fmt.Printf("错误: 命令 #%d 技能ID不匹配. 预期 %v, 实际 %v\n", i+1, origCmd.AbilityID, desCmd.AbilityID)
			}
			if desCmd.Position.X != origCmd.Position.X || desCmd.Position.Y != origCmd.Position.Y {
				fmt.Printf("错误: 命令 #%d 位置不匹配. 预期 (%v, %v), 实际 (%v, %v)\n",
					i+1, origCmd.Position.X, origCmd.Position.Y, desCmd.Position.X, desCmd.Position.Y)
			}
			if desCmd.CustomStr != origCmd.CustomStr {
				fmt.Printf("错误: 命令 #%d 自定义字符串不匹配. 预期 %v, 实际 %v\n", i+1, origCmd.CustomStr, desCmd.CustomStr)
			}
		}
	}

	fmt.Println("验证完成，如无错误提示则表示测试通过！")

	// 添加第二个测试，测试多个命令的情况
	fmt.Println("\n\n=== 测试多个命令 ===")

	// 构造测试数据，使用两个命令
	multiCommands := []gametypes.PlayerCommand{
		{
			CommandType: gametypes.UseAbility,
			AbilityID:   101,
			Position: gametypes.Vector2Int{
				X: 200,
				Y: 300,
			},
			CustomStr: "命令1-技能",
		},
		{
			CommandType: gametypes.Invalid,
			AbilityID:   0,
			Position: gametypes.Vector2Int{
				X: 500,
				Y: 600,
			},
			CustomStr: "命令2-无效",
		},
	}

	multiTestData := gametypes.PlayerInput{
		ID:         54321,
		LogicFrame: 200,
		Commands:   multiCommands,
	}

	// 序列化
	multiData := serialization.SerializePlayerInput(&multiTestData)
	fmt.Println("序列化数据长度:", len(multiData))

	// 反序列化
	fmt.Println("\n反序列化数据:")
	multiDeserializedData := serialization.DeserializePlayerInput(multiData)

	// 打印基本信息
	fmt.Printf("玩家ID: %v\n", multiDeserializedData.ID)
	fmt.Printf("逻辑帧: %v\n", multiDeserializedData.LogicFrame)
	fmt.Printf("命令数量: %v\n", len(multiDeserializedData.Commands))

	// 打印每个命令的详细信息
	fmt.Println("\n原始命令:")
	for i, cmd := range multiCommands {
		fmt.Printf("命令 #%d: 类型=%v, 技能ID=%v, 位置=(%v,%v), 字符串=%v\n",
			i+1, cmd.CommandType, cmd.AbilityID, cmd.Position.X, cmd.Position.Y, cmd.CustomStr)
	}

	fmt.Println("\n反序列化命令:")
	for i, cmd := range multiDeserializedData.Commands {
		fmt.Printf("命令 #%d: 类型=%v, 技能ID=%v, 位置=(%v,%v), 字符串=%v\n",
			i+1, cmd.CommandType, cmd.AbilityID, cmd.Position.X, cmd.Position.Y, cmd.CustomStr)
	}

	// 验证数据是否一致
	fmt.Println("\n多命令验证结果:")
	if len(multiDeserializedData.Commands) != len(multiCommands) {
		fmt.Printf("错误: 命令数量不匹配. 预期 %v, 实际 %v\n", len(multiCommands), len(multiDeserializedData.Commands))
	} else {
		for i, origCmd := range multiCommands {
			desCmd := multiDeserializedData.Commands[i]

			if desCmd.CommandType != origCmd.CommandType {
				fmt.Printf("错误: 命令 #%d 命令类型不匹配. 预期 %v, 实际 %v\n", i+1, origCmd.CommandType, desCmd.CommandType)
			}
			if desCmd.AbilityID != origCmd.AbilityID {
				fmt.Printf("错误: 命令 #%d 技能ID不匹配. 预期 %v, 实际 %v\n", i+1, origCmd.AbilityID, desCmd.AbilityID)
			}
			if desCmd.Position.X != origCmd.Position.X || desCmd.Position.Y != origCmd.Position.Y {
				fmt.Printf("错误: 命令 #%d 位置不匹配. 预期 (%v, %v), 实际 (%v, %v)\n",
					i+1, origCmd.Position.X, origCmd.Position.Y, desCmd.Position.X, desCmd.Position.Y)
			}
			if desCmd.CustomStr != origCmd.CustomStr {
				fmt.Printf("错误: 命令 #%d 自定义字符串不匹配. 预期 %v, 实际 %v\n", i+1, origCmd.CustomStr, desCmd.CustomStr)
			}
		}
	}
}

package serialization

import (
	"gameproject/fb"
	"gameproject/source/gametypes"

	flatbuffers "github.com/google/flatbuffers/go"
)

func SerializePlayerInput(data *gametypes.PlayerInput) []byte {
	builder := flatbuffers.NewBuilder(1024)

	// 创建命令数组偏移量列表
	commandOffsets := make([]flatbuffers.UOffsetT, len(data.Commands))

	// 第一步：逆序创建所有命令对象
	for i := len(data.Commands) - 1; i >= 0; i-- {
		cmd := data.Commands[i]

		// 创建自定义字符串
		customStrOffset := builder.CreateString(cmd.CustomStr)

		// 创建Vector2Int（位置）
		fb.Vector2IntStart(builder)
		fb.Vector2IntAddX(builder, int32(cmd.Position.X))
		fb.Vector2IntAddY(builder, int32(cmd.Position.Y))
		positionOffset := fb.Vector2IntEnd(builder)

		// 创建PlayerCommand
		fb.PlayerCommandStart(builder)
		fb.PlayerCommandAddCommandType(builder, gametypes.ConvertPlayerCommandType(cmd.CommandType))
		fb.PlayerCommandAddAbilityId(builder, int32(cmd.AbilityID))
		fb.PlayerCommandAddPosition(builder, positionOffset)
		fb.PlayerCommandAddCustomString(builder, customStrOffset)
		cmdOffset := fb.PlayerCommandEnd(builder)

		// 存储偏移量到数组中，注意这里保持原始顺序
		commandOffsets[i] = cmdOffset
	}

	// 第二步：创建命令数组向量
	// 启动向量构建
	fb.PlayerInputStartCommandsVector(builder, len(commandOffsets))
	// 逆序添加命令，这样在FlatBuffers中会保持原始顺序
	for i := len(commandOffsets) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(commandOffsets[i])
	}
	commandsVector := builder.EndVector(len(commandOffsets))

	// 第三步：创建PlayerInput
	fb.PlayerInputStart(builder)
	fb.PlayerInputAddPlayerId(builder, int32(data.ID))
	fb.PlayerInputAddFrame(builder, int32(data.LogicFrame))
	fb.PlayerInputAddCommands(builder, commandsVector)
	playerInputOffset := fb.PlayerInputEnd(builder)

	builder.Finish(playerInputOffset)
	return builder.FinishedBytes()
}

func DeserializePlayerInput(buf []byte) gametypes.PlayerInput {
	playerInput := fb.GetRootAsPlayerInput(buf, 0)

	// 解析命令数组
	commands := make([]gametypes.PlayerCommand, 0, playerInput.CommandsLength())

	// 注意：FlatBuffers 在反序列化时保持了序列化时的顺序，无需特殊处理
	for i := 0; i < playerInput.CommandsLength(); i++ {
		command := new(fb.PlayerCommand)
		if playerInput.Commands(command, i) {
			// 获取位置
			position := command.Position(nil)
			var pos gametypes.Vector2Int
			if position != nil {
				pos = gametypes.Vector2Int{
					X: int(position.X()),
					Y: int(position.Y()),
				}
			}

			// 创建PlayerCommand
			commands = append(commands, gametypes.PlayerCommand{
				CommandType: gametypes.ConvertFBPlayerCommandType(command.CommandType()),
				AbilityID:   int(command.AbilityId()),
				Position:    pos,
				CustomStr:   string(command.CustomString()),
			})
		}
	}

	return gametypes.PlayerInput{
		ID:         int(playerInput.PlayerId()),
		LogicFrame: int(playerInput.Frame()),
		Commands:   commands,
	}
}

func SerializeWorldSync(data gametypes.WorldSync) []byte {
	builder := flatbuffers.NewBuilder(1024)
	// Create S2CWorldSync
	fb.S2CWorldSyncStart(builder)
	fb.S2CWorldSyncAddLogicFrame(builder, int32(data.LogicFrame))
	fb.S2CWorldSyncAddServerTime(builder, int64(data.ServerTime))
	worldSyncOffset := fb.S2CWorldSyncEnd(builder)

	builder.Finish(worldSyncOffset)
	return builder.FinishedBytes()
}

func DeserializeWorldSync(buf []byte) gametypes.WorldSync {
	worldSync := fb.GetRootAsS2CWorldSync(buf, 0)
	result := gametypes.WorldSync{
		LogicFrame: worldSync.LogicFrame(),
	}

	return result
}

func AddPlayersVector(inBuilder *flatbuffers.Builder, players []gametypes.SerializePlayer) flatbuffers.UOffsetT {
	// Create all players first (in reverse order)
	playerOffsets := make([]flatbuffers.UOffsetT, 0, len(players))

	// Note: Create vectors and objects in reverse order
	for i := len(players) - 1; i >= 0; i-- {
		player := players[i]

		// Create Vector2Int
		fb.Vector2IntStart(inBuilder)
		fb.Vector2IntAddX(inBuilder, int32(player.Position.X))
		fb.Vector2IntAddY(inBuilder, int32(player.Position.Y))
		positionOffset := fb.Vector2IntEnd(inBuilder)

		// Create Player
		fb.PlayerStart(inBuilder)
		fb.PlayerAddPlayerId(inBuilder, int32(player.ID))
		fb.PlayerAddPosition(inBuilder, positionOffset)
		playerOffset := fb.PlayerEnd(inBuilder)

		playerOffsets = append([]flatbuffers.UOffsetT{playerOffset}, playerOffsets...)
	}

	// Create players vector
	fb.S2CStartEnterGameStartPlayersVector(inBuilder, len(playerOffsets))
	for _, offset := range playerOffsets {
		inBuilder.PrependUOffsetT(offset)
	}

	return inBuilder.EndVector(len(playerOffsets))
}

func SerializeS2CStartEnterGame(startEnterGame *gametypes.StartEnterGame) []byte {
	builder := flatbuffers.NewBuilder(1024)
	playersVector := AddPlayersVector(builder, startEnterGame.Players)

	// Create S2CStartEnterGame
	fb.S2CStartEnterGameStart(builder)
	fb.S2CStartEnterGameAddPlayers(builder, playersVector)
	startEnterGameOffset := fb.S2CStartEnterGameEnd(builder)

	builder.Finish(startEnterGameOffset)
	return builder.FinishedBytes()
}

func DeserializeS2CStartEnterGame(buf []byte) gametypes.StartEnterGame {
	startEnterGame := fb.GetRootAsS2CStartEnterGame(buf, 0)
	players := make([]gametypes.SerializePlayer, 0, startEnterGame.PlayersLength())

	for i := 0; i < startEnterGame.PlayersLength(); i++ {
		player := new(fb.Player)
		if startEnterGame.Players(player, i) {
			position := player.Position(nil)
			players = append(players, gametypes.SerializePlayer{
				ID: int(player.PlayerId()),
				Position: gametypes.Vector2Int{
					X: int(position.X()),
					Y: int(position.Y()),
				},
			})
		}
	}

	return gametypes.StartEnterGame{
		Players: players,
	}
}

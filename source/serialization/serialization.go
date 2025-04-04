package serialization

import (
	"gameproject/fb"
	"gameproject/source/gametypes"

	flatbuffers "github.com/google/flatbuffers/go"
)

func SerializePlayerInput(data *gametypes.PlayerInput) []byte {
	builder := flatbuffers.NewBuilder(1024)

	fb.PlayerInputStart(builder)
	fb.PlayerInputAddPlayerId(builder, int32(data.ID))
	fb.PlayerInputAddFrame(builder, int32(data.LogicFrame))
	fb.PlayerInputAddCommandType(builder, gametypes.ConvertPlayerCommandType(data.CommandType))
	playerInputOffset := fb.PlayerInputEnd(builder)

	builder.Finish(playerInputOffset)
	return builder.FinishedBytes()
}

func DeserializePlayerInput(buf []byte) gametypes.PlayerInput {
	playerInput := fb.GetRootAsPlayerInput(buf, 0)
	return gametypes.PlayerInput{
		ID:          int(playerInput.PlayerId()),
		LogicFrame:  int(playerInput.Frame()),
		CommandType: gametypes.ConvertFBPlayerCommandType(playerInput.CommandType()),
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

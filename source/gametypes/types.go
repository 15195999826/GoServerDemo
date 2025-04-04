package gametypes

import (
	"gameproject/fb"
	"log"
)

type PlayerCommandType int

const (
	MoveLeft PlayerCommandType = iota
	MoveRight
	MoveUp
	MoveDown
)

func (s PlayerCommandType) String() string {
	return [...]string{"MoveLeft", "MoveRight", "MoveUp", "MoveDown"}[s]
}

var (
	// FB命令类型到内部命令类型的映射
	fbToInternalCmd = map[fb.PlayerCommandType]PlayerCommandType{
		fb.PlayerCommandTypeMoveLeft:  MoveLeft,
		fb.PlayerCommandTypeMoveRight: MoveRight,
		fb.PlayerCommandTypeMoveUp:    MoveUp,
		fb.PlayerCommandTypeMoveDown:  MoveDown,
	}

	// 内部命令类型到FB命令类型的映射
	internalToFBCmd = map[PlayerCommandType]fb.PlayerCommandType{
		MoveLeft:  fb.PlayerCommandTypeMoveLeft,
		MoveRight: fb.PlayerCommandTypeMoveRight,
		MoveUp:    fb.PlayerCommandTypeMoveUp,
		MoveDown:  fb.PlayerCommandTypeMoveDown,
	}
)

func ConvertFBPlayerCommandType(t fb.PlayerCommandType) PlayerCommandType {
	if cmd, ok := fbToInternalCmd[t]; ok {
		return cmd
	}
	log.Printf("未知的FB PlayerCommandType: %d", t)
	return MoveLeft // 默认值
}

func ConvertPlayerCommandType(t PlayerCommandType) fb.PlayerCommandType {
	if cmd, ok := internalToFBCmd[t]; ok {
		return cmd
	}
	log.Printf("未知的PlayerCommandType: %d", t)
	return fb.PlayerCommandTypeMoveLeft // 默认值
}

type SerializePlayer struct {
	ID       int
	Position Vector2Int
}

type StartEnterGame struct {
	Players []SerializePlayer
}

type PlayerInput struct {
	ID          int
	LogicFrame  int
	CommandType PlayerCommandType
}

type WorldSync struct {
	LogicFrame int32
	ServerTime int64
}

type GameMapData struct {
	Width  int
	Height int
}

// 运行时地图
type GameMap struct {
	MapData *GameMapData
	// Todo: 运行时的地图状态
}

func NewGameMap(width, height int) *GameMap {
	return &GameMap{
		MapData: &GameMapData{
			Width:  width,
			Height: height,
		},
	}
}

package gametypes

import (
	"gameproject/fb"
	"log"
)

type PlayerCommandType int

const (
	Invalid PlayerCommandType = iota
	MoveTop
	MoveTopLeft
	MoveTopRight
	MoveDown
	MoveDownLeft
	MoveDownRight
)

func (s PlayerCommandType) String() string {
	return [...]string{"Invalid", "MoveTop", "MoveTopLeft", "MoveTopRight", "MoveDown", "MoveDownLeft", "MoveDownRight"}[s]
}

var (
	// FB命令类型到内部命令类型的映射
	fbToInternalCmd = map[fb.PlayerCommandType]PlayerCommandType{
		fb.PlayerCommandTypeInvalid:       Invalid,
		fb.PlayerCommandTypeMoveTop:       MoveTop,
		fb.PlayerCommandTypeMoveTopLeft:   MoveTopLeft,
		fb.PlayerCommandTypeMoveTopRight:  MoveTopRight,
		fb.PlayerCommandTypeMoveDown:      MoveDown,
		fb.PlayerCommandTypeMoveDownLeft:  MoveDownLeft,
		fb.PlayerCommandTypeMoveDownRight: MoveDownRight,
	}

	// 内部命令类型到FB命令类型的映射
	internalToFBCmd = map[PlayerCommandType]fb.PlayerCommandType{
		Invalid:       fb.PlayerCommandTypeInvalid,
		MoveTop:       fb.PlayerCommandTypeMoveTop,
		MoveTopLeft:   fb.PlayerCommandTypeMoveTopLeft,
		MoveTopRight:  fb.PlayerCommandTypeMoveTopRight,
		MoveDown:      fb.PlayerCommandTypeMoveDown,
		MoveDownLeft:  fb.PlayerCommandTypeMoveDownLeft,
		MoveDownRight: fb.PlayerCommandTypeMoveDownRight,
	}
)

func ConvertFBPlayerCommandType(t fb.PlayerCommandType) PlayerCommandType {
	if cmd, ok := fbToInternalCmd[t]; ok {
		return cmd
	}
	log.Printf("未知的FB PlayerCommandType: %d", t)
	return MoveTopLeft // 默认值
}

func ConvertPlayerCommandType(t PlayerCommandType) fb.PlayerCommandType {
	if cmd, ok := internalToFBCmd[t]; ok {
		return cmd
	}
	log.Printf("未知的PlayerCommandType: %d", t)
	return fb.PlayerCommandTypeMoveTop // 默认值
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

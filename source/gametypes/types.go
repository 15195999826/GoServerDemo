package gametypes

import (
	"gameproject/fb"
	"log"
)

type PlayerCommandType int

const (
	Invalid PlayerCommandType = iota
	UseAbility
)

func (s PlayerCommandType) String() string {
	return [...]string{"Invalid", "UseAbility"}[s]
}

var (
	// FB命令类型到内部命令类型的映射
	fbToInternalCmd = map[fb.PlayerCommandType]PlayerCommandType{
		fb.PlayerCommandTypeInvalid:    Invalid,
		fb.PlayerCommandTypeUseAbility: UseAbility,
	}

	// 内部命令类型到FB命令类型的映射
	internalToFBCmd = map[PlayerCommandType]fb.PlayerCommandType{
		Invalid:    fb.PlayerCommandTypeInvalid,
		UseAbility: fb.PlayerCommandTypeUseAbility,
	}
)

func ConvertFBPlayerCommandType(t fb.PlayerCommandType) PlayerCommandType {
	if cmd, ok := fbToInternalCmd[t]; ok {
		return cmd
	}
	log.Printf("未知的FB PlayerCommandType: %d", t)
	return Invalid // 默认值
}

func ConvertPlayerCommandType(t PlayerCommandType) fb.PlayerCommandType {
	if cmd, ok := internalToFBCmd[t]; ok {
		return cmd
	}
	log.Printf("未知的PlayerCommandType: %d", t)
	return fb.PlayerCommandTypeInvalid // 默认值
}

type SerializePlayer struct {
	ID       int
	Position Vector2Int
}

type StartEnterGame struct {
	Players []SerializePlayer
}

type PlayerCommand struct {
	AbilityID int
	Position  Vector2Int
	CustomStr string
}

type PlayerInput struct {
	ID         int
	LogicFrame int
	Commands   []PlayerCommand
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

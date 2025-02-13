package gametypes

import "gameproject/fb"

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

func ConvertFBPlayerCommandType(t fb.PlayerCommandType) PlayerCommandType {
	switch t {
	case fb.PlayerCommandTypeMoveLeft:
		return MoveLeft
	case fb.PlayerCommandTypeMoveRight:
		return MoveRight
	case fb.PlayerCommandTypeMoveUp:
		return MoveUp
	case fb.PlayerCommandTypeMoveDown:
		return MoveDown
	default:
		return MoveLeft
	}
}

type PlayerInput struct {
	LogicFrame  int32
	CommandType PlayerCommandType
}

package backend

import (
	"gameproject/fb"
	"gameproject/source/gametypes"
	"gameproject/source/serialization"
	"log"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/xtaci/kcp-go"
)

func createC2SCommand(command fb.ClientCommand, body []byte) []byte {
	builder := flatbuffers.NewBuilder(1024)

	// 创建body字节数组
	var bodyOffset flatbuffers.UOffsetT
	if body != nil {
		bodyOffset = builder.CreateByteVector(body)
	}

	// 开始构建S2CCommand
	fb.C2SCommandStart(builder)
	fb.C2SCommandAddCommand(builder, command)
	if body != nil {
		fb.C2SCommandAddBody(builder, bodyOffset)
	}
	rootOffset := fb.C2SCommandEnd(builder)

	// 完成构建
	builder.Finish(rootOffset)
	return builder.FinishedBytes()
}

func sendPing(conn *kcp.UDPSession) error {
	data := createC2SCommand(fb.ClientCommandC2S_COMMAND_PING, nil)

	_, err := conn.Write(data)
	if err != nil {
		log.Printf("Failed to send ping message: %v", err)
		return err
	}
	return nil
}

func sendRequestTime(conn *kcp.UDPSession) error {
	data := createC2SCommand(fb.ClientCommandC2S_COMMAND_REQUESTTIME, nil)

	_, err := conn.Write(data)
	if err != nil {
		log.Printf("Failed to send request time message: %v", err)
		return err
	}
	return nil
}

func sendGameLoaded(conn *kcp.UDPSession) error {
	data := createC2SCommand(fb.ClientCommandC2S_COMMAND_GAMELOADED, nil)

	_, err := conn.Write(data)
	if err != nil {
		log.Printf("Failed to send game loaded message: %v", err)
		return err
	}
	return nil
}

func sendPlayerInput(conn *kcp.UDPSession, playerInput *gametypes.PlayerInput) error {
	bodyBytes := serialization.SerializePlayerInput(playerInput)

	data := createC2SCommand(fb.ClientCommandC2S_COMMAND_PLAYERINPUT, bodyBytes)

	_, err := conn.Write(data)
	if err != nil {
		log.Printf("Failed to send movement message: %v", err)
		return err
	}
	return nil
}

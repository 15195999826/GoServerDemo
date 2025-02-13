package backend

import (
	"gameproject/fb"
	"log"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/xtaci/kcp-go"
)

type CommandSender struct {
}

func NewCommandSender() *CommandSender {
	return &CommandSender{}
}

func (ms *CommandSender) SendPing(conn *kcp.UDPSession) error {
	builder := flatbuffers.NewBuilder(1024)

	// 创建 C2SCommand
	fb.C2SCommandStart(builder)
	fb.C2SCommandAddCommand(builder, fb.ClientCommandC2S_COMMAND_PING)
	command := fb.C2SCommandEnd(builder)

	builder.Finish(command)
	data := builder.FinishedBytes()

	_, err := conn.Write(data)
	if err != nil {
		log.Printf("Failed to send ping message: %v", err)
		return err
	}
	return nil
}

func (ms *CommandSender) SendRequestTime(conn *kcp.UDPSession) error {
	builder := flatbuffers.NewBuilder(1024)

	// 创建 C2SCommand
	fb.C2SCommandStart(builder)
	fb.C2SCommandAddCommand(builder, fb.ClientCommandC2S_COMMAND_REQUESTTIME)
	command := fb.C2SCommandEnd(builder)

	builder.Finish(command)
	data := builder.FinishedBytes()

	_, err := conn.Write(data)
	if err != nil {
		log.Printf("Failed to send request time message: %v", err)
		return err
	}
	return nil
}

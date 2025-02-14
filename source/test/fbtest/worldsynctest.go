package fbtest

import (
	"fmt"
	"gameproject/fb"
	"gameproject/source/gametypes"
	"gameproject/source/serialization"
)

func SerializeWorldSync() []byte {
	// Create sample player inputs
	bodyBytes := serialization.SerializeWorldSync(gametypes.WorldSync{
		LogicFrame: 100,
	})

	// Create outer S2CCommand
	return createCommand(
		fb.ServerCommandS2C_COMMAND_WORLDSYNC,
		fb.S2CStatusS2C_STATUS_SUCCESS,
		0,
		"",
		bodyBytes,
	)
}

func DeserializeWorldSync(buf []byte) {
	// Parse outer S2CCommand
	command := fb.GetRootAsS2CCommand(buf, 0)
	fmt.Printf("Command: %v\n", command.Command())
	fmt.Printf("Status: %v\n", command.Status())

	// Parse S2CWorldSync from body
	worldSync := serialization.DeserializeWorldSync(command.BodyBytes())
	fmt.Printf("LogicFrame: %v\n", worldSync.LogicFrame)
}

func TestWorldSync() {
	// Serialize
	data := SerializeWorldSync()
	fmt.Println("Serialized data length:", len(data))

	// Deserialize
	fmt.Println("\nDeserialized data:")
	DeserializeWorldSync(data)
}

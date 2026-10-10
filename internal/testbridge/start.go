// Package testbridge supplies the startup transport for synthetic sample hosts.
package testbridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"isolate"

	"github.com/mfateev/sdk-go-poc/workflow"
)

func ReplyStart(command *isolate.Command, start workflow.PayloadStart) {
	// Assert the protocol independently of the SDK's internal envelope helper.
	// ISOL followed by protocol, metadata, API and determinism versions (uint32 LE).
	header := []byte{'I', 'S', 'O', 'L', 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0}
	if command.Op != workflow.OpStartPayloads || !bytes.Equal(command.Payload, header) {
		panic(fmt.Sprintf("unexpected startup contract: op=%d payload=%x", command.Op, command.Payload))
	}
	body, err := json.Marshal(start)
	if err != nil {
		panic(err)
	}
	command.Reply(append(header, body...), nil)
}

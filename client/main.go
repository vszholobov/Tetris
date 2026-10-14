package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	lib "github.com/vszholobov/tetrisLib"
	"os"
)

var Version = "dev"
var serverAddress = "tetris.vszholobov.ru:8080"

type CreateSessionResponse struct {
	SessionId int64 `json:"sessionId"`
}
type SessionDto struct {
	SessionId int64 `json:"sessionId"`
	Started   bool  `json:"started"`
}

var helpMessage = `
Network menu: no arguments, or --network [--server host:port]
Create: --network --server host:port --create
Join: --network --server host:port --session ID
Offline: --single [--seed N] [--trace]
Controls: a/d: move, s: down, q/e: rotate, x/Escape/Ctrl+C: exit
Menu: c: create, r: reload, w/s: select, Enter: join, x: exit
`

func main() {
	var err error
	switch {
	case len(os.Args) == 1:
		err = runNetwork(nil)
	case os.Args[1] == "--network":
		err = runNetwork(os.Args[2:])
	case os.Args[1] == "--single":
		err = runSingle(os.Args[2:])
	default:
		fmt.Print(helpMessage)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func decodeGameStateMessage(data []byte) (*lib.GameStateMessage, error) {
	if len(data) != 40 {
		return nil, fmt.Errorf("invalid GameStateMessage size: %d", len(data))
	}
	packet := &lib.GameStateMessage{}
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, packet); err != nil {
		return nil, err
	}
	return packet, nil
}

package handler

import "go-redis/protocol"

type PingHander struct{}

func NewPingHandler() *PingHander {
	return &PingHander{}
}

func (h *PingHander) Handle(args []protocol.Value) *protocol.Value {
	if len(args) == 0 {
		return protocol.SimpleString("PONG")
	}

	if len(args) == 1 {
		return protocol.BulkString(args[0].Str)
	}

	return protocol.Error("ERR wrong number of arguments for 'ping' command")
}

package handler

import "go-redis/protocol"

type PingHander struct{}

func NewPingHandler() *PingHander {
	return &PingHander{}
}

func (h *PingHander) Handle(args []protocol.Value) *protocol.Value {
	if len(args) == 0 {
		return &protocol.Value{
			Type: protocol.StringType,
			Str:  "PONG",
		}
	}

	if len(args) == 1 {
		return &protocol.Value{
			Type: protocol.BulkStringType,
			Str:  args[0].Str,
		}
	}

	return &protocol.Value{
		Type: protocol.ErrorType,
		Str:  "ERR wrong number of arguments for 'ping' command",
	}
}

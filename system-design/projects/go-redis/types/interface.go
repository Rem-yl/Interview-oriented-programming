package types

import "go-redis/protocol"

type Handler interface {
	Handle(args []protocol.Value) *protocol.Value
}

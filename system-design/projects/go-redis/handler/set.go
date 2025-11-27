package handler

import (
	"go-redis/protocol"
	"go-redis/store"
)

type SetHandler struct {
	db *store.Store
}

func NewSetHandler(db *store.Store) *SetHandler {
	return &SetHandler{
		db: db,
	}
}

func (h *SetHandler) Handle(args []protocol.Value) *protocol.Value {
	if len(args) < 2 {
		return protocol.Error("too less args")
	}

	if len(args) > 2 {
		return protocol.Error("too many args")
	}

	keyVal, valueVal := args[0], args[1]
	if keyVal.Type != protocol.BulkStringType {
		return protocol.Error("ERR key type error")
	}

	if valueVal.Type != protocol.BulkStringType && valueVal.Type != protocol.IntType {
		return protocol.Error("ERR value type error")
	}

	var value any
	key := keyVal.Str
	if valueVal.Type == protocol.BulkStringType {
		value = valueVal.Str
	} else {
		value = valueVal.Int
	}

	h.db.Set(key, value)

	return protocol.SimpleString("OK")
}

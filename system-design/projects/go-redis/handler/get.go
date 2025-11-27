package handler

import (
	"go-redis/protocol"
	"go-redis/store"
)

type GetHandler struct {
	db *store.Store
}

func NewGetHandler(db *store.Store) *GetHandler {
	return &GetHandler{
		db: db,
	}
}

func (h *GetHandler) Handle(args []protocol.Value) *protocol.Value {
	if len(args) < 1 {
		return protocol.Error("too less args")
	}

	if len(args) > 1 {
		return protocol.Error("too many args")
	}

	keyVal := args[0]
	if keyVal.Type != protocol.BulkStringType {
		return protocol.Error("ERR key type")
	}

	key := keyVal.Str
	value, exists := h.db.Get(key)
	if !exists {
		return protocol.NullBulkString()
	}

	var res *protocol.Value
	valueInt, ok := value.(int64)
	if !ok {
		valueStr, ok := value.(string)
		if !ok {
			return protocol.Error("ERR unknown value type")
		}
		res = &protocol.Value{
			Type:   protocol.BulkStringType,
			Str:    valueStr,
			IsNull: false,
		}

		return res
	}

	res = &protocol.Value{
		Type:   protocol.IntType,
		Int:    valueInt,
		IsNull: false,
	}

	return res
}

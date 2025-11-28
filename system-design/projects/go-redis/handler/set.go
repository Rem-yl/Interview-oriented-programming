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
	if len(args) != 2 {
		return protocol.Error("ERR wrong number of arguments for 'set' command")
	}

	key := args[0].Str
	value := args[1].Str

	// SET 命令总是存储字符串
	// INCR/DECR 等命令会在需要时将字符串解析为整数
	h.db.Set(key, value)

	return protocol.SimpleString("OK")
}

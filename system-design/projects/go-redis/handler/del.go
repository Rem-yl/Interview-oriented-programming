package handler

import (
	"go-redis/protocol"
	"go-redis/store"
)

type DelHandler struct {
	db *store.Store
}

func NewDelHandler(db *store.Store) *DelHandler {
	return &DelHandler{
		db: db,
	}
}

func (h *DelHandler) Handle(args []protocol.Value) *protocol.Value {
	if len(args) < 1 {
		return protocol.Error("ERR wrong number of arguments for 'del' command")
	}

	var deletedCount int64 = 0

	for _, keyVal := range args {
		key := keyVal.Str
		if h.db.Delete(key) {
			deletedCount++
		}
	}

	return protocol.Integer(deletedCount)
}

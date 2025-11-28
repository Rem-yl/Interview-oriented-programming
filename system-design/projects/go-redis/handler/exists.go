package handler

import (
	"go-redis/protocol"
	"go-redis/store"
)

type ExistsHandler struct {
	db *store.Store
}

func NewExistsHandler(db *store.Store) *ExistsHandler {
	return &ExistsHandler{
		db: db,
	}
}

func (h *ExistsHandler) Handle(args []protocol.Value) *protocol.Value {
	if len(args) != 1 {
		return protocol.Error("len of args mistake")
	}

	keyVal := args[0]
	if keyVal.Type != protocol.BulkStringType {
		return protocol.Error("mistake args type")
	}

	exists := h.db.Exists(keyVal.Str)
	var n int64 = 0
	if exists {
		n = 1
	}

	return protocol.Integer(n)
}

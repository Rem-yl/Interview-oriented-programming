package handler

import (
	"go-redis/protocol"
	"go-redis/store"
	"strconv"
)

type IncrHandler struct {
	db *store.Store
}

func NewIncrHandler(db *store.Store) *IncrHandler {
	return &IncrHandler{
		db: db,
	}
}

func (h *IncrHandler) Handle(args []protocol.Value) *protocol.Value {
	if len(args) != 1 {
		return protocol.Error("ERR mistake args len")
	}

	keyVal := args[0]
	if keyVal.Type != protocol.BulkStringType {
		return protocol.Error("ERR key type")
	}

	if ok := h.db.Incr(keyVal.Str); !ok {
		return protocol.Error("ERR value is not an integer or out of range")
	}

	return protocol.SimpleString("OK")
}

// *****
type IncrByHandler struct {
	db *store.Store
}

func NewIncrByHandler(db *store.Store) *IncrByHandler {
	return &IncrByHandler{
		db: db,
	}
}

func (h *IncrByHandler) Handle(args []protocol.Value) *protocol.Value {
	if len(args) != 2 {
		return protocol.Error("ERR mistake args len")
	}

	keyVal, valueVal := args[0], args[1]
	if keyVal.Type != protocol.BulkStringType {
		return protocol.Error("ERR key type")
	}

	count, err := strconv.ParseInt(valueVal.Str, 10, 64)
	if err != nil {
		return protocol.Error("ERR count type")
	}

	if ok := h.db.IncrBy(keyVal.Str, count); !ok {
		return protocol.Error("ERR value is not an integer or out of range")
	}

	return protocol.SimpleString("OK")
}

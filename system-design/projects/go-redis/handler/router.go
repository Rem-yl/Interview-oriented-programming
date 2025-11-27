package handler

import (
	"go-redis/protocol"
	"go-redis/store"
	"go-redis/types"
	"strings"
)

type Router struct {
	handlers map[string]types.Handler
	db       *store.Store
}

func NewRouter(s *store.Store) *Router {
	r := &Router{
		handlers: make(map[string]types.Handler),
		db:       s,
	}

	r.registerDefaultHandlers()
	return r
}

func (r *Router) Route(cmd *protocol.Value) *protocol.Value {
	if cmd.Type != protocol.ArrayType {
		return protocol.Error("ERR expected array")
	}

	if len(cmd.Array) == 0 {
		return protocol.Error("ERR empty command")
	}

	cmdName := strings.ToUpper(cmd.Array[0].Str)

	handler, exists := r.handlers[cmdName]
	if !exists {
		return protocol.Error("ERR unknown command: " + cmdName)
	}

	args := cmd.Array[1:]

	return handler.Handle(args)
}

func (r *Router) Register(cmd string, handler types.Handler) {
	r.handlers[strings.ToUpper(cmd)] = handler
}

func (r *Router) registerDefaultHandlers() {
	r.Register("PING", NewPingHandler())
	r.Register("SET", NewSetHandler(r.db))
	r.Register("GET", NewGetHandler(r.db))
}

package handler

import (
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

func (r *Router) Register(cmd string, handler types.Handler) {
	r.handlers[strings.ToUpper(cmd)] = handler
}

func (r *Router) registerDefaultHandlers() {
	r.Register("PING", NewPingHandler())
}

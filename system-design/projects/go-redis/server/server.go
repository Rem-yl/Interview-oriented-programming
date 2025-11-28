package server

import (
	"fmt"
	"go-redis/handler"
	"go-redis/logger"
	"go-redis/store"
	"net"
	"sync"
	"sync/atomic"
)

type Server struct {
	addr     string
	listener net.Listener
	router   *handler.Router
	db       *store.Store
	clients  sync.Map
	shutdown chan struct{}
	wg       sync.WaitGroup
	clientID int64
}

func NewServer(addr string, s *store.Store) *Server {
	router := handler.NewRouter(s)

	return &Server{
		addr:     addr,
		router:   router,
		db:       s,
		shutdown: make(chan struct{}),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}

	s.listener = listener
	logger.Infof("Redis server listening on %s", s.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.shutdown:
				logger.Info("Server is shutting down")
				return nil
			default:
				logger.Errorf("Failed to accept connection: %v", err)
				continue
			}
		}

		client := NewClient(conn, s.router, s.nextClientID())
		s.clients.Store(client.id, client)

		s.wg.Add(1)

		go func() {
			defer s.wg.Done()
			client.Serve()
			s.clients.Delete(client.id)
		}()
	}
}

func (s *Server) Stop() error {
	logger.Info("Stopping server...")

	close(s.shutdown)

	if s.listener != nil {
		s.listener.Close()
	}

	s.clients.Range(func(key, value interface{}) bool {
		client := value.(*Client)
		client.Close()
		return true
	})

	s.wg.Wait()

	logger.Info("Server stopped")
	return nil
}

func (s *Server) nextClientID() string {
	id := atomic.AddInt64(&s.clientID, 1)
	return fmt.Sprintf("client-%d", id)
}

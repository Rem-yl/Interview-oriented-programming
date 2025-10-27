package main

import (
	"errors"
	"sync/atomic"
)

type Balancer interface {
	GetServer() (*Server, error)
}

// Server
type Server struct {
	URL                string
	Name               string
	activateConnection int32
}

func (s *Server) GetConnections() int32 {
	val := atomic.LoadInt32(&s.activateConnection)
	return val
}

func (s *Server) AddConnection() {
	atomic.AddInt32(&s.activateConnection, 1)
}

func (s *Server) DecConnection() {
	atomic.AddInt32(&s.activateConnection, -1)
}

// Balancer
type LeastConnectionBalancer struct {
	serverList []*Server
}

func NewLeastConnectionBalancer(serverList []*Server) *LeastConnectionBalancer {
	balancer := &LeastConnectionBalancer{
		serverList: serverList,
	}

	return balancer
}

func (b *LeastConnectionBalancer) GetServer() (*Server, error) {
	var leastServer *Server

	if len(b.serverList) <= 0 {
		return nil, errors.New("No server list.")
	}

	leastServer = b.serverList[0]
	if len(b.serverList) > 1 {
		for _, server := range b.serverList {
			if server.activateConnection < leastServer.activateConnection {
				leastServer = server
			}
		}
	}

	leastServer.AddConnection()
	return leastServer, nil
}

func main() {
	// Create test servers
	servers := []*Server{
		{URL: "http://server1.com", Name: "Server1", activateConnection: 0},
		{URL: "http://server2.com", Name: "Server2", activateConnection: 0},
		{URL: "http://server3.com", Name: "Server3", activateConnection: 0},
	}

	// Initialize balancer
	balancer := NewLeastConnectionBalancer(servers)

	// Simulate 10 requests
	println("Simulating least connection load balancing:")
	println("============================================")

	for i := 1; i <= 10; i++ {
		server, err := balancer.GetServer()
		if err != nil {
			println("Error:", err.Error())
			return
		}

		println("Request", i, "->", server.Name, "(Connections:", server.GetConnections(), ")")

		// Simulate some requests completing (decrement connections)
		if i%3 == 0 && i > 0 {
			// Every 3rd request, simulate a connection closing on server1
			servers[0].DecConnection()
			println("  -> Server1 connection closed (Connections:", servers[0].GetConnections(), ")")
		}
	}

	println("\nFinal connection counts:")
	println("=======================")
	for _, server := range servers {
		println(server.Name, "- Active Connections:", server.GetConnections())
	}
}

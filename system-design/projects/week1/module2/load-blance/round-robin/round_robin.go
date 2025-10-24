package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
)

type Server struct {
	URL  string
	Name string
}

type RoundRobinBalancer struct {
	ServerList []*Server
	index      int // 用于给出下一个服务器
	mu         sync.Mutex
}

func NewRoundRobinBalancer(serverList []*Server) *RoundRobinBalancer {
	balancer := &RoundRobinBalancer{
		ServerList: serverList,
		index:      0,
	}

	return balancer
}

// 轮询列表中的服务器
func (b *RoundRobinBalancer) GetServer() (*Server, error) {
	if len(b.ServerList) <= 0 {
		return nil, errors.New("no server in list")
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	server := b.ServerList[b.index]
	b.index = (b.index + 1) % len(b.ServerList)

	return server, nil
}

func creatServerList(n int) []*Server {
	serverList := make([]*Server, 0, n)
	for i := range n {
		server := &Server{
			URL:  fmt.Sprintf("http://127.0.0.1:%d", 8000+rand.Intn(1000)),
			Name: fmt.Sprintf("server-%d", i),
		}
		serverList = append(serverList, server)
	}

	return serverList
}

func main() {
	serverList := creatServerList(5)
	balancer := NewRoundRobinBalancer(serverList)
	fmt.Println("============================")
	fmt.Println("Server list: ")
	for _, server := range serverList {
		fmt.Printf("Name: %s, URL: %s \n", server.Name, server.URL)
	}
	fmt.Println("============================")
	fmt.Println("Start Round Robin")
	for i := range 10 {
		server, err := balancer.GetServer()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Round: %d, Name: %s, URL: %s \n", i, server.Name, server.URL)
	}

	fmt.Println("============================")
	fmt.Println("Done.")
}

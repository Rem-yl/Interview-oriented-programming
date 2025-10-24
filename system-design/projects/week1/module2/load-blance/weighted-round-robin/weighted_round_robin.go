package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
)

type Server struct {
	URL    string
	Name   string
	weight int
}

type Balancer interface {
	GetServer() (*Server, error)
}

func creatServerList(n int) []*Server {
	serverList := make([]*Server, 0, n)
	for i := range n {
		server := &Server{
			URL:    fmt.Sprintf("http://127.0.0.1:%d", 8000+rand.Intn(1000)),
			Name:   fmt.Sprintf("server-%d", i),
			weight: max(1, rand.Intn(5)),
		}
		serverList = append(serverList, server)
	}

	return serverList
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}

	return a
}

// ********** NaiveWeightedRoundRobinBalancer ********** //
type NaiveWeightedRoundRobinBalancer struct {
	serverList []*Server
	mu         sync.Mutex
	curIdx     int
}

func (b *NaiveWeightedRoundRobinBalancer) buildBalancer() {
	if len(b.serverList) <= 0 {
		return
	}

	newServerList := make([]*Server, 0)

	for _, server := range b.serverList {
		for _ = range server.weight {
			newServerList = append(newServerList, server)
		}
	}

	b.serverList = newServerList
}

func NewNaiveWeightedRoundRobinBalancer(serverList []*Server) *NaiveWeightedRoundRobinBalancer {
	balancer := &NaiveWeightedRoundRobinBalancer{
		serverList: serverList,
		curIdx:     0,
	}

	balancer.buildBalancer()
	return balancer
}

func (b *NaiveWeightedRoundRobinBalancer) GetServer() (*Server, error) {
	if len(b.serverList) <= 0 {
		return nil, errors.New("no server in list")
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	server := b.serverList[b.curIdx]
	b.curIdx = (b.curIdx + 1) % len(b.serverList)

	return server, nil
}

// *************************************************** //
// ********** GcdWeightedRoundRobinBalancer ********** //
// *************************************************** //

type GcdWeightedRoundRobinBalancer struct {
	serverList []*Server
	mu         sync.Mutex
	curIdx     int
	curWeight  int
	gcdWeight  int
	maxWeight  int
}

func (b *GcdWeightedRoundRobinBalancer) buildBalancer() {
	if len(b.serverList) <= 0 {
		return
	}

	var gcdWeight, maxWeight int

	for i, server := range b.serverList {
		if i == 0 {
			gcdWeight = server.weight
			maxWeight = server.weight
		} else {
			maxWeight = max(maxWeight, server.weight)
			gcdWeight = gcd(gcdWeight, server.weight)
		}
	}

	b.maxWeight = maxWeight
	b.gcdWeight = gcdWeight
}

func NewGcdWeightedRoundRobinBalancer(serverList []*Server) *GcdWeightedRoundRobinBalancer {
	balancer := &GcdWeightedRoundRobinBalancer{
		serverList: serverList,
		curIdx:     -1,
		curWeight:  0,
		gcdWeight:  0,
		maxWeight:  0,
	}

	balancer.buildBalancer()
	return balancer
}

func (b *GcdWeightedRoundRobinBalancer) GetServer() (*Server, error) {
	if len(b.serverList) <= 0 {
		return nil, errors.New("no server in list")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for {
		b.curIdx = (b.curIdx + 1) % len(b.serverList)
		if b.curIdx == 0 {
			b.curWeight -= b.gcdWeight
			if b.curWeight <= 0 {
				b.curWeight = b.maxWeight
			}
		}

		if b.serverList[b.curIdx].weight >= b.curWeight {
			return b.serverList[b.curIdx], nil
		}
	}
}

// *************************************************** //
// ********** SwrrWeightedRoundRobinBalancer ********* //
// *************************************************** //
type SwrrWeightedRoundRobinBalancer struct {
	serverList  []*SwrrServer
	curIdx      int
	mu          sync.Mutex
	totalWeight int
}

type SwrrServer struct {
	Server
	curWeight int
}

func (b *SwrrWeightedRoundRobinBalancer) build_balancer() {
	if len(b.serverList) <= 0 {
		return
	}

	totalWeight := 0

	for _, server := range b.serverList {
		totalWeight += server.weight
		server.curWeight = 0
	}

	b.totalWeight = totalWeight
}

func NewSwrrWeightedRoundRobinBalancer(serverList []*Server) *SwrrWeightedRoundRobinBalancer {
	swrrServerList := make([]*SwrrServer, 0)
	for _, server := range serverList {
		swrrServer := &SwrrServer{
			Server:    *server,
			curWeight: 0,
		}
		swrrServerList = append(swrrServerList, swrrServer)
	}

	balancer := &SwrrWeightedRoundRobinBalancer{
		serverList:  swrrServerList,
		curIdx:      0,
		totalWeight: 0,
	}

	balancer.build_balancer()

	return balancer
}

func (b *SwrrWeightedRoundRobinBalancer) GetServer() (*Server, error) {
	if len(b.serverList) <= 0 {
		return nil, errors.New("no server in list")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	selectServer := b.serverList[0]

	for _, swrrServer := range b.serverList {
		swrrServer.curWeight += swrrServer.weight
	}

	for _, swrrServer := range b.serverList {
		if swrrServer.curWeight > selectServer.curWeight {
			selectServer = swrrServer
		}
	}

	selectServer.curWeight -= b.totalWeight

	return &selectServer.Server, nil
}

func testNaiveBalancer() {
	serverList := creatServerList(20)
	var balancer Balancer
	balancer = NewNaiveWeightedRoundRobinBalancer(serverList)
	fmt.Println("============================")
	fmt.Println("Server list: ")

	for _, server := range serverList {
		fmt.Printf("Name: %s, URL: %s, weight: %d \n", server.Name, server.URL, server.weight)
	}
	fmt.Println("============================")
	fmt.Println("Start Round Robin")
	for i := range 50 {
		server, err := balancer.GetServer()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Round: %d, Name: %s, URL: %s \n", i, server.Name, server.URL)
	}

	fmt.Println("============================")
	fmt.Println("Done.")
}

func testGcdBalancer() {
	serverList := []*Server{
		{Name: "A", URL: "1", weight: 2},
		{Name: "B", URL: "2", weight: 4},
		{Name: "C", URL: "3", weight: 2},
	}

	var balancer Balancer
	balancer = NewGcdWeightedRoundRobinBalancer(serverList)
	fmt.Println("============================")
	fmt.Println("Server list: ")

	for _, server := range serverList {
		fmt.Printf("Name: %s, URL: %s, weight: %d \n", server.Name, server.URL, server.weight)
	}
	fmt.Println("============================")
	fmt.Println("Start Round Robin")
	for i := range 50 {
		server, err := balancer.GetServer()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Round: %d, Name: %s, URL: %s \n", i, server.Name, server.URL)
	}

	fmt.Println("============================")
	fmt.Println("Done.")
}

func testSwrrBalancer() {
	serverList := []*Server{
		{Name: "A", URL: "1", weight: 2},
		{Name: "B", URL: "2", weight: 5},
		{Name: "C", URL: "3", weight: 1},
	}

	var balancer Balancer
	balancer = NewSwrrWeightedRoundRobinBalancer(serverList)
	fmt.Println("============================")
	fmt.Println("Server list: ")

	for _, server := range serverList {
		fmt.Printf("Name: %s, URL: %s, weight: %d \n", server.Name, server.URL, server.weight)
	}
	fmt.Println("============================")
	fmt.Println("Start Round Robin")
	for i := range 50 {
		server, err := balancer.GetServer()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Round: %d, Name: %s, URL: %s \n", i, server.Name, server.URL)
	}

	fmt.Println("============================")
	fmt.Println("Done.")
}

func main() {
	testNaiveBalancer()
	testGcdBalancer()
	testSwrrBalancer()
}

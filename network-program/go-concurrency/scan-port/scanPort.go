package main

import (
	"flag"
	"fmt"
	"go-concurrency/logger"
	"net"
	"strconv"
	"sync"
	"time"
)

var (
	address = "127.0.0.1"
	result  []string
	resMux  sync.Mutex
)

func scanPort(port int, wg *sync.WaitGroup, ch chan bool) {
	defer wg.Done()
	fullAddr := address + ":" + strconv.Itoa(port)
	logger.Debug("Start scan: %s", fullAddr)

	conn, err := net.DialTimeout("tcp", fullAddr, 2*time.Second)
	if err != nil {
		<-ch
		return
	}
	defer conn.Close()

	resMux.Lock()
	result = append(result, fmt.Sprintf("port: %d is open", port))
	resMux.Unlock()

	<-ch
}

func main() {

	var startPort, endPort, workers int
	flag.IntVar(&startPort, "start_port", 0, "start port")
	flag.IntVar(&endPort, "end_port", 1024, "end port")
	flag.IntVar(&workers, "num_workers", 32, "num of workers")
	flag.Parse()

	var wg sync.WaitGroup
	concurr_chan := make(chan bool, workers)

	for port := startPort; port <= endPort; port++ {
		wg.Add(1)
		concurr_chan <- true
		go scanPort(port, &wg, concurr_chan)
	}

	wg.Wait()
	for _, res := range result {
		logger.Info(res)
	}
}

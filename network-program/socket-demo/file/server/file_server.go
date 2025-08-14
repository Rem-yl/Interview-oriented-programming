package main

import (
	"net"
	logger "sockect-demo/logger"
	"sync"
	"time"
)

var (
	clients   = make(map[string]net.Conn) // 记录哪些客户端连接
	clientMux sync.Mutex
)

func main() {
	address := "127.0.0.1:2121"

	l, err := net.Listen("tcp", address)
	if err != nil {
		logger.Errorf("Listen on %s error: %s", address, err)
		return
	}

	defer l.Close()
	logger.Info("Listen on", address)

	for {
		conn, err := l.Accept()

		clientMux.Lock()
		clients[conn.RemoteAddr().String()] = conn
		clientMux.Unlock()

		if err != nil {
			logger.Errorf("Connect on %s error: %s", conn.RemoteAddr(), err)
			continue
		}

		go fileServer(conn)
	}
}

func fileServer(conn net.Conn) {
	defer func() {
		defer conn.Close()
		clientMux.Lock()
		delete(clients, conn.RemoteAddr().String())
		clientMux.Unlock()

		logger.Infof("Connect %s closed.", conn.RemoteAddr())
	}()

	logger.Infof("Connect %s will be closed.", conn.RemoteAddr())
	time.Sleep(2 * time.Second)
}

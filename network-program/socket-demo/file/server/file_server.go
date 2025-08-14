package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	logger "sockect-demo/logger"
	"strings"
	"sync"
)

const ROOTDIR = "./data"

var (
	clients   = make(map[string]net.Conn) // 记录哪些客户端连接
	clientMux sync.Mutex
)

func main() {
	logger.SetLevel(logger.INFO)

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

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		text := scanner.Text()

		logger.Debug("Get text:", text)

		if strings.HasPrefix(text, "list") {
			logger.Debug("Should run listServe")
			listServe(ROOTDIR, "")
		}
	}
}

func listServe(path string, prefix string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		logger.Errorf("Read dir %s error: %s \n", path, err)
		return
	}

	for i, entry := range entries {
		var connector string
		if i == len(entries)-1 {
			connector = "└── "
		} else {
			connector = "├── "
		}

		fmt.Println(prefix + connector + entry.Name())

		if entry.IsDir() {
			newPrefix := prefix
			if i == len(entries)-1 {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			listServe(filepath.Join(path, entry.Name()), newPrefix)
		}
	}
}

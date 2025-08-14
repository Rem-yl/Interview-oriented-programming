package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	clients   = make(map[net.Conn]string)
	clientMux sync.Mutex
)

func main() {
	address := "127.0.0.1:8801"

	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("[Main] Listen on %s error: %s \n", address, err)
		return
	}
	defer l.Close()
	fmt.Println("[Main] Listen on ", address)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Printf("[Main] Error: %s \n", err)
			continue
		}

		clientMux.Lock()
		clients[conn] = conn.RemoteAddr().String()
		clientMux.Unlock()

		go handlerServe(conn)
	}
}

func handlerServe(conn net.Conn) {
	defer func() {
		conn.Close()
		clientMux.Lock()
		delete(clients, conn)
		clientMux.Unlock()
	}()

	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println(conn.RemoteAddr(), "Disconnect")
				return
			}
			fmt.Println("Read Error: ", err)
			return
		}
		msg = strings.TrimRight(msg, "\r\n")
		fmt.Println("[Server] Recived message: \n", msg)

		t := time.Now()
		myTime := t.Format(time.RFC3339)
		msg = fmt.Sprintf("[%s]: %s\n", myTime, msg)

		broadcast(msg)
	}
}

func broadcast(msg string) {
	clientMux.Lock()
	defer clientMux.Unlock()

	for conn := range clients {
		if _, err := conn.Write([]byte(msg)); err != nil {
			fmt.Println("Write error to", conn.RemoteAddr(), err)
		}

		fmt.Printf("[BroadCast] %s to %s \n", msg, conn.RemoteAddr())
	}
}

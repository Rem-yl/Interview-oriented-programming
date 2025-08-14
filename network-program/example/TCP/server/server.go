package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8001")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("Server listening on 127.0.0.1:8001")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn) // 包装一个 io.Reader（比如文件、网络连接、标准输入等）并提供带缓冲的读取功能

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected: ", err)
			return
		}

		msg = strings.TrimSpace(msg)

		fmt.Println("Received: ", msg)

		t := time.Now()
		myTime := t.Format(time.RFC3339)
		msg = fmt.Sprintf("[%s]: %s\n", myTime, msg)

		_, err = conn.Write([]byte(msg))

		if err != nil {
			fmt.Println("Write Error: ", err)
			return
		}
	}
}

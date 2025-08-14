// first: go run ../server/server.go
// then: go run client.go 127.0.0.1:8001
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	args := os.Args

	if len(args) == 1 {
		fmt.Println("Please provide host:port")
		return
	}

	address := args[1]
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n') // 从终端读取文本
		fmt.Fprintf(conn, text)            // 发送到服务器

		message, _ := bufio.NewReader(conn).ReadString('\n') // 等待服务器返回消息
		fmt.Print("->: " + message)
		if strings.TrimSpace(string(text)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}

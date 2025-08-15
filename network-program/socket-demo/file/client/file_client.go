package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	logger "sockect-demo/logger"
	"strings"
)

func main() {
	address := "127.0.0.1:2121"

	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.Errorf("Connect on %s error: %s", address, err)
		return
	}

	defer conn.Close()

	stdScanner := bufio.NewScanner(os.Stdin)
	connScanner := bufio.NewScanner(conn)
	fmt.Print(">>> ")

	go func() {
		for connScanner.Scan() {
			text := connScanner.Text()
			fmt.Println(text)

			fmt.Print(">>> ")
		}
	}()

	// 主线程 用于监控终端输入
	for stdScanner.Scan() {
		text := strings.ToLower(stdScanner.Text())
		logger.Info("CMD: ", text)

		if _, err := conn.Write([]byte(text + "\n")); err != nil {
			logger.Errorf("Write to connect %s error: %s \n", conn.RemoteAddr(), err)
		}
	}
}

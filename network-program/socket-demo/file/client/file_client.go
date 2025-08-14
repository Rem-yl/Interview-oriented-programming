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

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(">>> ")

	for scanner.Scan() {
		text := strings.ToLower(scanner.Text())
		logger.Info("CMD: ", text)

		if _, err := conn.Write([]byte(text + "\n")); err != nil {
			logger.Errorf("Write to connect %s error: %s \n", conn.RemoteAddr(), err)
		}
		fmt.Print(">>> ")
	}
}

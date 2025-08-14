package main

import (
	"net"
	logger "sockect-demo/logger"
)

func main() {
	address := "127.0.0.1:2121"

	conn, err := net.Dial("tcp", address)
	if err != nil {
		logger.Errorf("Connect on %s error: %s", address, err)
		return
	}

	defer conn.Close()
}

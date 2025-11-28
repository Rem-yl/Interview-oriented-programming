package main

import (
	"flag"
	"fmt"
	"go-redis/logger"
	"go-redis/server"
	"go-redis/store"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	var port int
	var logLevel string

	flag.IntVar(&port, "port", 16379, "端口")
	flag.StringVar(&logLevel, "loglevel", "info", "日志级别: debug | info | warn | error")
	flag.Parse()

	level, err := logrus.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		logger.Errorf("非法日志级别: %s", logLevel)
	}

	logger.SetLevel(level)

	s := store.NewStore()

	srv := server.NewServer(fmt.Sprintf(":%d", port), s)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		logger.Info("Received shutdown signal")
		srv.Stop()
		os.Exit(0)
	}()

	logger.Info("Starting Go-Redis server on :", port)
	if err := srv.Start(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}

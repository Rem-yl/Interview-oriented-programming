package main

import (
	"go-redis/logger"
	"go-redis/server"
	"go-redis/store"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	logger.SetLevel(logrus.InfoLevel)

	s := store.NewStore()

	srv := server.NewServer(":16379", s)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		logger.Info("Received shutdown signal")
		srv.Stop()
		os.Exit(0)
	}()

	logger.Info("Starting Go-Redis server on :16379")
	if err := srv.Start(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}

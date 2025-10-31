package test

import (
	"testing"

	"github.com/rem/load-balancer/internal/server"
)

func TestHelloServer(t *testing.T) {
	server.HelloServer("8089")
}

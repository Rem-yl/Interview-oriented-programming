package test

import (
	"testing"

	"github.com/rem/load-balancer/pkg/algo"
	"github.com/rem/load-balancer/pkg/backend"
)

func createBackEnd() []backend.BackEnd {
	serverList := make([]backend.BackEnd, 0, 3)
	ports := []string{"8088", "8089", "8090"}

	for _, port := range ports {
		url := "localhost:" + port
		name := "server-" + port
		backend := backend.NewSimpleBackEnd(url, name)
		serverList = append(serverList, backend)
	}

	return serverList
}

func TestRoundRobin(t *testing.T) {
	serverList := createBackEnd()
	balancer := algo.NewRoundRobinLoadBalancer(serverList)
	for _ = range 10 {
		backend, err := balancer.GetBackEnd()
		if err != nil {
			t.Error(err)
		}

		t.Logf("Name: %s, URL: %s", backend.GetName(), backend.GetURL())
	}
}

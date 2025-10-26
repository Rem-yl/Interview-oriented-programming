package main

import (
	"fmt"
	"testing"
)

func TestSwrrBalancer(t *testing.T) {
	resultMap := make(map[string]int)
	totalWeight := 0
	resultMap["total"] = 0

	serverList := []*Server{
		{Name: "A", URL: "1", weight: 2},
		{Name: "B", URL: "2", weight: 5},
		{Name: "C", URL: "3", weight: 1},
	}

	for _, server := range serverList {
		resultMap[server.Name] = 0
		totalWeight += server.weight
	}

	var balancer Balancer
	totalRequest := totalWeight * 100
	resultMap["total"] = totalRequest
	balancer = NewSwrrWeightedRoundRobinBalancer(serverList)

	for range totalRequest {
		server, err := balancer.GetServer()
		if err != nil {
			panic(err)
		}

		resultMap[server.Name] += 1
	}

	fmt.Println(resultMap)
}

func TestGcdBalancer(t *testing.T) {
	resultMap := make(map[string]int)
	resultMap["total"] = 0
	serverList := []*Server{
		{Name: "A", URL: "1", weight: 2},
		{Name: "B", URL: "2", weight: 4},
		{Name: "C", URL: "3", weight: 2},
	}
	for _, server := range serverList {
		resultMap[server.Name] = 0
	}

	var balancer Balancer
	balancer = NewGcdWeightedRoundRobinBalancer(serverList)
	totalRequest := 8 * 100
	resultMap["total"] = totalRequest

	for range totalRequest {
		server, err := balancer.GetServer()
		if err != nil {
			panic(err)
		}

		resultMap[server.Name] += 1
	}

	fmt.Println(resultMap)
}

package algo

import "github.com/rem/load-balancer/pkg/backend"

func NewRoundRobinLoadBalancer(serverList []backend.BackEnd) *RoundRobinLoadBalancer {
	var roundServerList []*backend.SimpleBackEnd

	for _, server := range serverList {
		roundServer := backend.NewSimpleBackEnd(server.GetURL(), server.GetName(), server.GetWeight())
		roundServerList = append(roundServerList, roundServer)
	}

	return &RoundRobinLoadBalancer{
		serverList: roundServerList,
		idx:        0,
	}
}

func NewSwrrRobinLoadBalancer(serverList []backend.BackEnd) *SwrrRobinLoadBalancer {
	var swrrServerList []*backend.SwrrBackEnd
	totalWeight := 0

	for _, server := range serverList {
		totalWeight += server.GetWeight()
		swrrServer := backend.NewSwrrBackEnd(server.GetURL(), server.GetName(), server.GetWeight())
		swrrServerList = append(swrrServerList, swrrServer)
	}
	loadBalancer := &SwrrRobinLoadBalancer{
		serverList:  swrrServerList,
		totalWeight: totalWeight,
	}

	return loadBalancer
}

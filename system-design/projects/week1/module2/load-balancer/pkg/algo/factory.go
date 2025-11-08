package algo

import "github.com/rem/load-balancer/pkg/backend"

func NewRoundRobinLoadBalancer(serverList []backend.BackEnd) *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{
		serverList: serverList,
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

func NewConsistenceHashLoadBalancer(serverList []backend.BackEnd) *ConsistenceHashLoadBalancer {

	return nil
}

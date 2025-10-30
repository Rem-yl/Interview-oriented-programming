package algo

import "github.com/rem/load-balancer/pkg/backend"

func NewRoundRobinLoadBalancer(serverList []backend.BackEnd) *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{
		serverList: serverList,
		idx:        0,
	}
}

package algo

import (
	"sync"

	"github.com/rem/load-balancer/pkg/backend"
)

type ConsistenceHashLoadBalancer struct {
	serverList []*backend.HashBackend
	mutex      sync.Mutex
}

func (b *ConsistenceHashLoadBalancer) GetBackEnd() (backend.BackEnd, error) {
	return nil, nil
}

package algo

import (
	"errors"
	"sync"

	"github.com/rem/load-balancer/pkg/backend"
)

type RoundRobinLoadBalancer struct {
	serverList []backend.BackEnd
	idx        int
	mutex      sync.Mutex
}

func (b *RoundRobinLoadBalancer) GetBackEnd() (backend.BackEnd, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if len(b.serverList) <= 0 {
		return nil, errors.New("No server list!")
	}

	backend := b.serverList[b.idx]
	b.idx = (b.idx + 1) % len(b.serverList)

	return backend, nil
}

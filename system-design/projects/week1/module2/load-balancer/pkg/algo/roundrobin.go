package algo

import (
	"sync"

	"github.com/rem/load-balancer/pkg/backend"
	"github.com/rem/load-balancer/pkg/errs"
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
		return nil, errs.ErrNoServerList
	}

	backend := b.serverList[b.idx]
	b.idx = (b.idx + 1) % len(b.serverList)

	return backend, nil
}

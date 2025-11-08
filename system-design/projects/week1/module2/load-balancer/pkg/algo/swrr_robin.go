package algo

import (
	"sync"

	"github.com/rem/load-balancer/pkg/backend"
	"github.com/rem/load-balancer/pkg/errs"
)

type SwrrRobinLoadBalancer struct {
	serverList  []*backend.SwrrBackEnd
	mutex       sync.Mutex
	totalWeight int
}

func (b *SwrrRobinLoadBalancer) GetBackEnd() (backend.BackEnd, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(b.serverList) <= 0 {
		return nil, errs.ErrNoServerList
	}

	selectServer := b.serverList[0]

	// 充电
	for _, server := range b.serverList {
		server.CurWeight += server.Weight
	}

	for _, server := range b.serverList {
		if server.CurWeight > selectServer.CurWeight {
			selectServer = server
		}
	}

	selectServer.CurWeight -= b.totalWeight

	return selectServer, nil
}

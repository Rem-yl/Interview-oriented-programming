package tester

import (
	"context"

	"github.com/rem/load-balancer/internal/clients"
)

type ConcurrentTester struct {
	loadBalanceClient *clients.LoadBalanceClient
	backendClient     *clients.BackEndClient
	totalCnt          int //总请求数
	concurrent        int // 并发数
}

func NewConcurrentTester(loadBalanceClient *clients.LoadBalanceClient, backendClient *clients.BackEndClient, totalCnt int, concurrent int) *ConcurrentTester {
	return &ConcurrentTester{
		loadBalanceClient: loadBalanceClient,
		backendClient:     backendClient,
		totalCnt:          totalCnt,
		concurrent:        concurrent,
	}
}

func (t *ConcurrentTester) Run(ctx context.Context) ([]RequestResult, error) {

	return nil, nil
}

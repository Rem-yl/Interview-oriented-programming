package tester

import (
	"context"
	"fmt"
	"time"

	"github.com/rem/load-balancer/internal/clients"
)

type SequentialTester struct {
	loadBalanceClient *clients.LoadBalanceClient
	backendClient     *clients.BackEndClient
	count             int
}

func NewSequentialTester(loadBalanceClient *clients.LoadBalanceClient, backendClient *clients.BackEndClient, count int) *SequentialTester {
	return &SequentialTester{
		loadBalanceClient: loadBalanceClient,
		backendClient:     backendClient,
		count:             count,
	}
}

func (t *SequentialTester) Run(ctx context.Context) ([]RequestResult, error) {
	results := make([]RequestResult, 0, t.count)

	for i := 0; i < t.count; i++ {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		result := t.Request(i)

		results = append(results, result)
	}

	return results, nil
}

func (t *SequentialTester) Request(requestID int) RequestResult {
	result := RequestResult{
		RequestID: requestID,
		Timestamp: time.Now(),
	}

	start := time.Now()
	backend, err := t.loadBalanceClient.GetBackend()
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("获取后端失败: %w", err)
		result.Latency = time.Since(start)
		return result
	}

	_, err = t.backendClient.Request(backend.GetURL())
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("请求后端失败: %w", err)
		result.Backend = backend.GetURL()
		result.Latency = time.Since(start)
		return result
	}

	result.Success = true
	result.Backend = backend.GetURL()
	result.Latency = time.Since(start)

	return result
}

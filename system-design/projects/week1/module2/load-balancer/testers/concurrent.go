package testers

import (
	"context"
	"fmt"
	"sync"
	"time"

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
	requestsPerWorker := t.totalCnt / t.concurrent
	remainder := t.totalCnt % t.concurrent

	resultCh := make(chan RequestResult, t.totalCnt)

	var wg sync.WaitGroup

	for i := 0; i < t.concurrent; i++ {
		count := requestsPerWorker
		if i < remainder {
			count++
		}

		wg.Add(1)
		go func(workID, count int) {
			defer wg.Done()
			t.RunWorker(ctx, workID, count, resultCh)
		}(i, count)
	}

	// 启动一个goroutine在所有worker完成后关闭channel
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// 主goroutine从channel读取结果
	results := make([]RequestResult, 0, t.totalCnt)
	for result := range resultCh {
		results = append(results, result)
	}

	return results, nil
}

func (t *ConcurrentTester) RunWorker(ctx context.Context, workID, count int, resultCh chan<- RequestResult) {
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return
		default:
		}
		result := t.Request(workID, i)
		resultCh <- result
	}
}

func (t *ConcurrentTester) Request(workID, i int) RequestResult {
	result := RequestResult{
		WorkerID:  workID,
		RequestID: workID*1000 + i,
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

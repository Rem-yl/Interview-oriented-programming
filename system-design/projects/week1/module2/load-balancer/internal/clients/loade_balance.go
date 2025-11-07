package clients

import (
	"encoding/json"
	"fmt"

	"github.com/rem/load-balancer/pkg/backend"
)

// LoadBalanceClient 仅负责获取后端 url
type LoadBalanceClient struct {
	client  HttpClient
	baseURL string
}

func NewLoadBalanceClient(client HttpClient, url string) *LoadBalanceClient {
	return &LoadBalanceClient{
		client:  client,
		baseURL: url,
	}
}

func (c *LoadBalanceClient) GetBackend() (backend.BackEnd, error) {
	body, err := c.client.Get(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("获取后端失败: %w", err)
	}

	var resp SimpleBackEndResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp.Data, nil
}

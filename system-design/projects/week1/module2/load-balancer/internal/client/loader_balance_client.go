package client

import (
	"encoding/json"
	"time"

	"github.com/rem/load-balancer/pkg/backend"
)

type LoadBalanceClient struct {
	client  *DefaultHttpClient
	url     string
	timeout time.Duration
}

func NewLoadBalanceClient(url string, timeout time.Duration) *LoadBalanceClient {
	httpClient := NewDefaultHttpClient(timeout)

	return &LoadBalanceClient{
		client:  httpClient,
		url:     url,
		timeout: timeout,
	}
}

func (c *LoadBalanceClient) Get(url string) ([]byte, error) {
	data, err := c.client.Get(url)

	return data, err
}

func (c *LoadBalanceClient) Post(url string, body []byte) ([]byte, error) {
	data, err := c.client.Post(url, body)

	return data, err
}

type SimpleBackEnd struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

func (b *SimpleBackEnd) GetURL() string {
	return b.URL
}

func (b *SimpleBackEnd) GetName() string {
	return b.Name
}

type BackEndResponse struct {
	Data SimpleBackEnd `json:"data"`
}

func (c *LoadBalanceClient) GetBackend() (backend.BackEnd, error) {
	body, err := c.client.Get(c.url)
	if err != nil {
		return nil, err
	}

	var resp BackEndResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

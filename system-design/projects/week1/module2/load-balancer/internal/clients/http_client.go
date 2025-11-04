package clients

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rem/load-balancer/pkg/errs"
)

type DefaultHttpClient struct {
	client  *http.Client
	timeout time.Duration
}

func NewDefaultHttpClient(timeout time.Duration) *DefaultHttpClient {
	return &DefaultHttpClient{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

func (c *DefaultHttpClient) Get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errs.ErrCreateGetRequest
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errs.ErrSendRequest
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("POST %s failed: status=%d, body=%s", url, resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}

func (c *DefaultHttpClient) Post(url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create POST request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send POST request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("POST %s failed: status=%d, body=%s", url, resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}

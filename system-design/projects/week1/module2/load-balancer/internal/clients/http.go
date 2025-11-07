package clients

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type DefaultHttpClient struct {
	client *http.Client
}

func NewDefaultHttpClient(timeout time.Duration) *DefaultHttpClient {
	return &DefaultHttpClient{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				// ===== 连接池配置 =====
				MaxIdleConns:        100,              // 最大空闲连接数
				MaxIdleConnsPerHost: 10,               // 每个 host 的最大空闲连接
				IdleConnTimeout:     90 * time.Second, // 空闲连接超时

				// ===== 连接超时配置 =====
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,  // 连接超时
					KeepAlive: 30 * time.Second, // Keep-Alive 间隔
				}).DialContext,

				// ===== TLS 配置 =====
				TLSHandshakeTimeout: 5 * time.Second, // TLS 握手超时

				// ===== 其他超时 =====
				ResponseHeaderTimeout: 10 * time.Second, // 响应头超时
				ExpectContinueTimeout: 1 * time.Second,  // 100-continue 超时

				// ===== Keep-Alive 配置 =====
				DisableKeepAlives: false, // 启用 Keep-Alive

				// ===== 压缩配置 =====
				DisableCompression: false, // 启用压缩

				// ===== 连接限制 =====
				MaxConnsPerHost: 0, // 无限制（根据需要调整）
			},
		},
	}
}

func (c *DefaultHttpClient) Get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create GET request failed: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send GET request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GET %s failed: status=%d, body=%s", url, resp.StatusCode, string(respBody))
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
		return nil, fmt.Errorf("send POST request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("POST %s failed: status=%d, body=%s", url, resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}

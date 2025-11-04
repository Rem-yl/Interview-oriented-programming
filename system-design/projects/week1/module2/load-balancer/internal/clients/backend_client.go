package clients

import (
	"encoding/json"
)

type BackEndClient struct {
	client HttpClient
}

func NewBackEndClient(client HttpClient) *BackEndClient {
	return &BackEndClient{
		client: client,
	}
}

func (c *BackEndClient) Request(url string) (string, error) {
	body, err := c.client.Get(url)
	if err != nil {
		return "", err
	}

	var resp HelloServerResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}

	return resp.Data, nil
}

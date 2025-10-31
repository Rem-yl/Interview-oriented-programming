package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"log"
)

type Response struct {
	Data BackendInfo `json:"data"`
}

type HelloServerResponse struct {
	Data string `json:"data"`
}

type BackendInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type HttpClient struct {
	client  *http.Client
	timeout time.Duration
	logger  log.Logger
}

func NewHttpClient(timeout time.Duration) *HttpClient {

}

func request[T any](url string) (*T, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func main() {
	var balanceResult *Response
	balanceResult, err := request[Response]("http://127.0.0.1:8187/balancer")
	if err != nil {
		log.Println(err.Error())
		return
	}

	var serverResult *HelloServerResponse
	serverResult, err = request[HelloServerResponse](balanceResult.Data.URL)
	if err != nil {
		log.Println(err.Error())
		return
	}

	fmt.Println(serverResult.Data)
}

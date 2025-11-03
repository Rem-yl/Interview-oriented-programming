package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rem/load-balancer/internal/client"
)

// 单独定义 Data 结构体
type DataResponse struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// 再定义 Response 结构体，包含 Data
type BalancerResponse struct {
	Data DataResponse `json:"data"`
}

type HelloResponse struct {
	Data string `json:"data"`
}

func main() {
	client := client.NewDefaultHttpClient(5 * time.Second)
	url := "127.0.0.1"
	port := "8187"
	addr := fmt.Sprintf("http://%s:%s/balancer", url, port)

	// 发送 GET 请求
	body, err := client.Get(addr)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}

	// 解析 JSON 响应
	var resp BalancerResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		fmt.Println("JSON 解析失败:", err)
		fmt.Println("原始响应:", string(body))
		return
	}

	// 打印结果
	fmt.Println("Name:", resp.Data.Name)
	fmt.Println("URL:", resp.Data.URL)

	body, err = client.Get(resp.Data.URL)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}

	var hResp HelloResponse
	if err := json.Unmarshal(body, &hResp); err != nil {
		fmt.Println("JSON 解析失败:", err)
		fmt.Println("原始响应:", string(body))
		return
	}

	fmt.Println("data: ", hResp.Data)
}

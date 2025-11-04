package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rem/load-balancer/internal/clients"
)

var (
	url = "127.0.0.1"
)

func main() {
	port := flag.String("port", "8187", "客户端启动端口")
	flag.Parse()

	// 验证端口参数
	if *port == "" {
		fmt.Println("错误: 端口号不能为空")
		flag.Usage()
		os.Exit(1)
	}

	addr := fmt.Sprintf("http://%s:%s/balancer", url, *port)
	httpClient := clients.NewDefaultHttpClient(5 * time.Second)
	loadBalanceClient := clients.NewLoadBalanceClient(httpClient, addr)
	backendClient := clients.NewBackEndClient(httpClient)

	backend, err := loadBalanceClient.GetBackend()
	if err != nil {
		fmt.Println("获取后端链接失败: ", err)
		return
	}

	msg, err := backendClient.Request(backend.GetURL())
	if err != nil {
		fmt.Printf("访问后端链接: %s 失败, err: %v \n", backend.GetURL(), err)
		return
	}

	fmt.Println("访问服务成功: ", msg)
}

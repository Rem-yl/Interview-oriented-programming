package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rem/load-balancer/internal/clients"
	"github.com/rem/load-balancer/testers"
)

var (
	url = "127.0.0.1"
)

func main() {
	var (
		port        string
		count       int
		mode        string
		concurrency int
	)
	flag.StringVar(&port, "port", "8187", "客户端启动端口")
	flag.StringVar(&mode, "mode", "sequential", "测试器验证模式")
	flag.IntVar(&count, "count", 100, "总请求数")
	flag.IntVar(&concurrency, "concurrence", 10, "并发数")
	flag.Parse()

	// 验证端口参数
	if port == "" {
		fmt.Println("错误: 端口号不能为空")
		flag.Usage()
		os.Exit(1)
	}

	// 1. 创建客户端
	addr := fmt.Sprintf("http://%s:%s/balancer", url, port)
	httpClient := clients.NewDefaultHttpClient(5 * time.Second)
	loadBalanceClient := clients.NewLoadBalanceClient(httpClient, addr)
	backendClient := clients.NewBackEndClient(httpClient)

	// 2. 创建测试器
	var tester testers.Tester
	switch mode {
	case "sequential":
		tester = testers.NewSequentialTester(loadBalanceClient, backendClient, count)
	case "concurrence":
		tester = testers.NewConcurrentTester(loadBalanceClient, backendClient, count, concurrency)
	default:
		panic("No tester.")
	}

	// 3. 运行测试
	ctx := context.Background()
	results, err := tester.Run(ctx)
	if err != nil {
		fmt.Printf("测试失败: %v \n", err)
		os.Exit(1)
	}

	reporter := testers.NewStatistics()
	reporter.Analyse(results)
	err = reporter.GenReport("load_balancer_report.json")
	if err != nil {
		fmt.Println(err)
	}
}

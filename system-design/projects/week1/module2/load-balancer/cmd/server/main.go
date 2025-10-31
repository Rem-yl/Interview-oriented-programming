package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rem/load-balancer/internal/server"
)

func main() {
	// 解析命令行参数
	port := flag.String("port", "8080", "服务器监听端口")
	flag.Parse()

	// 验证端口参数
	if *port == "" {
		fmt.Println("错误: 端口号不能为空")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("启动服务器，监听端口: %s\n", *port)

	// 启动服务器
	server.HelloServer(*port)
}

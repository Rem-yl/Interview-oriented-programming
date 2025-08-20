package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

// 模拟泄漏：不断启动 goroutine，但 goroutine 永远阻塞
func leakGoroutines() {
	for {
		go func() {
			ch := make(chan struct{})
			<-ch // 永远不会关闭，导致 goroutine 泄漏
		}()
		time.Sleep(10 * time.Millisecond)
	}
}

func main() {
	// 启动 pprof 服务器
	go func() {
		log.Println("pprof server start at :6060")
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	// 模拟 goroutine 泄漏
	go leakGoroutines()

	// 主线程保持运行
	for {
		fmt.Println("程序运行中...")
		time.Sleep(5 * time.Second)
	}
}

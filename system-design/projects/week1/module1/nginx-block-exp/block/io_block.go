package main

import (
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	requestCount int64
	activeConns  int64
)

func doWork(n int) {
	time.Sleep(time.Duration(n) * time.Millisecond)
}

func countMiddleWare(c *gin.Context) {
	atomic.AddInt64(&activeConns, 1)
	defer atomic.AddInt64(&activeConns, -1)
	atomic.AddInt64(&requestCount, 1)
	c.Next()
}

func testHandler(c *gin.Context) {
	doWork(50)
	c.JSON(http.StatusOK, gin.H{
		"msg":   "Hello from Blocking IO",
		"model": "每个请求一个 Goroutine（阻塞）",
	})
}

func statsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"goroutines":     runtime.NumGoroutine(),
		"total_requests": atomic.LoadInt64(&requestCount),
		"active_conns":   atomic.LoadInt64(&activeConns),
	})
}

func main() {
	r := gin.Default()

	r.GET("/test", countMiddleWare, testHandler)
	r.GET("/stats", statsHandler)

	fmt.Println("━━━━━━━━ 阻塞 I/O 服务器 (Gin) ━━━━━━━━")
	fmt.Println("🎯 模型: 每个请求一个 Goroutine（同步处理）")
	fmt.Println("📍 端口: 8001")
	fmt.Println("📊 统计: http://localhost:8001/stats")
	fmt.Println()
	r.Run(":8001")
}

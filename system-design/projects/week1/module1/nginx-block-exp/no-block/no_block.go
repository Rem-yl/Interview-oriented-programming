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
	workerCount  = 100
)

// Task 表示一个异步任务
type Task struct {
	ResultChan chan Result
}

type Result struct {
	Message string
	Model   string
}

// 工作池：固定数量的 Worker Goroutine
var taskQueue = make(chan Task, 1000) // 任务队列

func doWork(n int) {
	time.Sleep(time.Duration(n) * time.Millisecond)
}

func countMiddleWare(c *gin.Context) {
	atomic.AddInt64(&activeConns, 1)
	defer atomic.AddInt64(&activeConns, -1)
	atomic.AddInt64(&requestCount, 1)
	c.Next()
}

// Worker 函数：类似 NGINX 的 Worker 进程
func worker(id int) {
	for task := range taskQueue { // taskQueue为空时会阻塞, 同时多个 Goroutine 读写同一channel是并发安全的
		// 模拟阻塞操作（数据库查询等）
		doWork(50)

		// 发送结果
		task.ResultChan <- Result{
			Message: "Hello from Non-Blocking IO",
			Model:   fmt.Sprintf("工作池模式 (Worker %d)", id),
		}
	}
}

func testHandler(c *gin.Context) {
	task := Task{
		ResultChan: make(chan Result, 1),
	}

	select {
	case taskQueue <- task:
		result := <-task.ResultChan
		c.JSON(http.StatusOK, result)
	default:
		// 队列满了，拒绝请求
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "服务器繁忙，请稍后重试",
		})
	}
}

func statsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"goroutines":     runtime.NumGoroutine(),
		"total_requests": atomic.LoadInt64(&requestCount),
		"active_conns":   atomic.LoadInt64(&activeConns),
		"worker_count":   workerCount,
		"queue_len":      len(taskQueue),
	})
}

// 初始化 Worker Pool
func initWorkerPool() {
	for i := 0; i < workerCount; i++ {
		go worker(i)
	}
	fmt.Printf("✅ 启动了 %d 个 Worker Goroutine（类似 NGINX worker_processes）\n", workerCount)
}

func main() {
	initWorkerPool()

	r := gin.Default()

	r.GET("/test", countMiddleWare, testHandler)
	r.GET("/stats", statsHandler)

	r.Run(":8002")
}

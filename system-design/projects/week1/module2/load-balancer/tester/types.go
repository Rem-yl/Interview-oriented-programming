package tester

import "time"

type RequestResult struct {
	RequestID int
	Timestamp time.Time // 请求时间戳

	// 请求结果
	Success bool
	Backend string        // 请求到的后端url
	Latency time.Duration // 请求延迟
	Error   error

	WorkerID int // 记录是哪个worker发送的
}

package testers

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

type Statistics struct {
	// 基础统计信息
	TotalRequest int
	SuccessCnt   int
	FailureCnt   int
	SuccessRate  float64

	// 延迟统计
	TotalDuration time.Duration
	AvgLatency    time.Duration
	MinLatency    time.Duration
	MaxLatency    time.Duration
	P50Latency    time.Duration
	P90Latency    time.Duration
	P95Latency    time.Duration
	P99Latency    time.Duration

	// 性能指标
	QPS float64

	// 后端分布情况
	BackendDistribution map[string]int
	BackendPercentage   map[string]float64

	// 错误统计
	ErrDistribution map[string]int
}

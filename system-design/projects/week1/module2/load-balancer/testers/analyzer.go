package testers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"
)

// percentile returns the approximate pth percentile from sorted durations.
//
// 例如：
//
//	sorted = [10ms, 20ms, 30ms, 40ms, 50ms]   // 共 5 个样本
//	len(sorted) = 5
//
//	要计算 P90：
//	  pos = 0.90 * (5 - 1) = 3.6
//	  lower = 3 → 40ms
//	  upper = 4 → 50ms
//	  frac  = 0.6
//	  P90 = 40ms + (50ms - 40ms) * 0.6 = 46ms
//
// 也就是说，第 90 百分位处在第 3 个样本（40ms）和第 4 个样本（50ms）之间，
// 位于两者之间大约 60% 的位置，插值得到更平滑的分位估计。
//
// P = y_lower + (y_upper - y_lower) * fraction
func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}

	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}

	// 用线性插值计算分位位置
	pos := (p / 100.0) * float64(len(sorted)-1)
	lower := int(pos)
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[lower]
	}

	frac := pos - float64(lower)
	return sorted[lower] + time.Duration(frac*float64(sorted[upper]-sorted[lower]))
}

func NewStatistics() *Statistics {
	return &Statistics{
		BackendDistribution: make(map[string]int),
		BackendPercentage:   make(map[string]float64),
		ErrDistribution:     make(map[string]int),
	}
}

func (s *Statistics) Analyse(results []RequestResult) error {
	if len(results) <= 0 {
		return errors.New("No results to analyse.")
	}

	s.calcBaseStatic(results)
	s.calcDuration(results) // 计算总耗时
	s.calcLatency(results)  // 计算延迟分布
	s.calcQPS()
	s.calcDistribution(results)
	return nil
}

func (s *Statistics) GenReport(filename string) error {
	// 创建一个用于 JSON 输出的结构体，将 time.Duration 转换为字符串便于阅读
	report := struct {
		// 基础统计信息
		TotalRequest int     `json:"total_request"`
		SuccessCnt   int     `json:"success_count"`
		FailureCnt   int     `json:"failure_count"`
		SuccessRate  float64 `json:"success_rate"`

		// 延迟统计 (转换为字符串便于阅读)
		TotalDuration string `json:"total_duration"`
		AvgLatency    string `json:"avg_latency"`
		MinLatency    string `json:"min_latency"`
		MaxLatency    string `json:"max_latency"`
		P50Latency    string `json:"p50_latency"`
		P90Latency    string `json:"p90_latency"`
		P95Latency    string `json:"p95_latency"`
		P99Latency    string `json:"p99_latency"`

		// 性能指标
		QPS string `json:"qps"`

		// 后端分布情况
		BackendDistribution map[string]int     `json:"backend_distribution"`
		BackendPercentage   map[string]float64 `json:"backend_percentage"`

		// 错误统计
		ErrDistribution map[string]int `json:"error_distribution"`
	}{
		TotalRequest: s.TotalRequest,
		SuccessCnt:   s.SuccessCnt,
		FailureCnt:   s.FailureCnt,
		SuccessRate:  s.SuccessRate,

		TotalDuration: s.TotalDuration.String(),
		AvgLatency:    s.AvgLatency.String(),
		MinLatency:    s.MinLatency.String(),
		MaxLatency:    s.MaxLatency.String(),
		P50Latency:    s.P50Latency.String(),
		P90Latency:    s.P90Latency.String(),
		P95Latency:    s.P95Latency.String(),
		P99Latency:    s.P99Latency.String(),

		QPS: fmt.Sprintf("%.2f requests/sec", s.QPS),

		BackendDistribution: s.BackendDistribution,
		BackendPercentage:   s.BackendPercentage,
		ErrDistribution:     s.ErrDistribution,
	}

	// 使用 MarshalIndent 生成格式化的 JSON，便于阅读
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal statistics: %w", err)
	}

	// 打印到控制台
	fmt.Println("========== Load Balancer Test Report ==========")
	fmt.Println(string(jsonData))
	fmt.Println("================================================")

	// 保存到文件
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("Warning: failed to save report to file: %v\n", err)
	}

	fmt.Printf("Report saved to: %s\n", filename)
	return nil
}

// 计算基础统计信息
func (s *Statistics) calcBaseStatic(results []RequestResult) {
	s.TotalRequest = len(results)
	for _, result := range results {
		if result.Success {
			s.SuccessCnt++
		} else {
			s.FailureCnt++
		}
	}

	s.SuccessRate = float64(s.SuccessCnt) / float64(s.TotalRequest)
}

// 总耗时计算, 只计算访问后端成功的请求
// duration =  最大时间戳的请求 + 请求延迟 - 最小时间戳的请求
// 模拟从最早的请求到最后一个请求收到时的总耗时
func (s *Statistics) calcDuration(results []RequestResult) {
	// 初始化比较的变量
	minTimeStamp, maxTimeStamp := results[0].Timestamp, results[0].Timestamp
	maxIdx := 0

	for i, result := range results {
		if !result.Success || i == 0 {
			continue
		}

		if result.Timestamp.Before(minTimeStamp) {
			minTimeStamp = result.Timestamp
		} else if result.Timestamp.After(maxTimeStamp) {
			maxTimeStamp = result.Timestamp
			maxIdx = i
		}

	}

	s.TotalDuration = maxTimeStamp.Add(results[maxIdx].Latency).Sub(minTimeStamp)
}

// 延迟分布指标计算, 只计算访问后端成功的请求
func (s *Statistics) calcLatency(results []RequestResult) {
	successLatencys := make([]time.Duration, 0, len(results))

	for _, result := range results {
		if !result.Success {
			continue
		}

		successLatencys = append(successLatencys, result.Latency)
	}

	if len(successLatencys) == 0 {
		s.AvgLatency = 0
		s.MinLatency = 0
		s.MaxLatency = 0
		s.P50Latency = 0
		s.P90Latency = 0
		s.P90Latency = 0
		s.P95Latency = 0
		s.P99Latency = 0

		return
	}

	// 对延迟进行从小到大排序
	sort.Slice(successLatencys, func(i, j int) bool {
		return successLatencys[i] < successLatencys[j]
	})

	n := len(successLatencys)
	s.MinLatency = successLatencys[0]
	s.MaxLatency = successLatencys[n-1]

	total := time.Duration(0)
	for _, duration := range successLatencys {
		total += duration
	}
	s.AvgLatency = total / time.Duration(n)

	s.P50Latency = percentile(successLatencys, 50)
	s.P90Latency = percentile(successLatencys, 90)
	s.P95Latency = percentile(successLatencys, 95)
	s.P99Latency = percentile(successLatencys, 99)
}

func (s *Statistics) calcQPS() {
	if s.TotalDuration == 0 {
		s.QPS = 0
		return
	}

	s.QPS = float64(s.TotalRequest) / s.TotalDuration.Seconds()
}

func (s *Statistics) calcDistribution(results []RequestResult) {
	totalBackend := 0

	for _, result := range results {
		if result.Backend != "" {
			s.BackendDistribution[result.Backend]++
			totalBackend++
		}
		if !result.Success {
			s.ErrDistribution[result.Error.Error()]++
		}
	}

	for key, value := range s.BackendDistribution {
		s.BackendPercentage[key] = float64(value) / float64(totalBackend)
	}
}

package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

/*
并发连接测试客户端 - NGINX 事件循环演示

使用方法：
  go run client_concurrent.go [选项]

测试模式：
  --mode slow        慢速发送模式（默认）- 5个连接，每字节间隔100ms
  --mode keepalive   保持连接模式 - 10个连接保持30秒不关闭
  --mode burst       突发模式 - 同时建立20个连接
  --mode stress      压力测试模式 - 持续30秒发送100并发请求
  --mode pipeline    流水线模式 - 10个连接每个发送10个请求
  --mode gradual     渐进模式 - 每秒增加5个连接，观察扩容

高级选项：
  --host string      服务器地址 (默认 "localhost:8080")
  --connections int  并发连接数 (覆盖模式默认值)
  --duration int     测试持续时间（秒）(仅部分模式)
  --delay int        慢速模式字节间隔（毫秒）(默认 100)
  --verbose          详细输出

示例：
  # 慢速发送模式（观察多个连接同时存在）
  go run client_concurrent.go --mode slow

  # 突发模式测试批量接受能力
  go run client_concurrent.go --mode burst

  # 压力测试，100个并发连接
  go run client_concurrent.go --mode stress --connections 100

  # 自定义服务器地址
  go run client_concurrent.go --mode burst --host 192.168.1.100:8080

提示：
  - 使用 slow 模式观察服务器同时处理多个连接
  - 使用 burst 模式测试 event_loop_improved.go 的批量接受能力
  - 使用 stress 模式对比 simple 和 improved 版本的性能差异
  - 使用 keepalive 模式测试服务器的超时机制
*/

// 配置参数
var (
	mode        = flag.String("mode", "slow", "测试模式: slow, keepalive, burst, stress, pipeline, gradual")
	host        = flag.String("host", "localhost:8080", "服务器地址")
	connections = flag.Int("connections", 0, "并发连接数（0=使用模式默认值）")
	duration    = flag.Int("duration", 0, "测试持续时间（秒）（0=使用模式默认值）")
	delay       = flag.Int("delay", 100, "慢速模式字节间隔（毫秒）")
	verbose     = flag.Bool("verbose", false, "详细输出")
)

func main() {
	flag.Parse()

	fmt.Println("🚀 并发连接测试客户端 - NGINX 事件循环演示")
	fmt.Printf("🎯 目标服务器: %s\n", *host)
	fmt.Printf("📋 测试模式: %s\n\n", *mode)

	switch *mode {
	case "slow":
		slowSendMode()
	case "keepalive":
		keepAliveMode()
	case "burst":
		burstMode()
	case "stress":
		stressMode()
	case "pipeline":
		pipelineMode()
	case "gradual":
		gradualMode()
	default:
		fmt.Printf("❌ 未知模式: %s\n", *mode)
		fmt.Println("可用模式: slow, keepalive, burst, stress, pipeline, gradual")
		flag.Usage()
	}
}

// ==================== 模式 1: 慢速发送 ====================

func slowSendMode() {
	connCount := getConnCount(5)
	byteDelay := time.Duration(*delay) * time.Millisecond

	fmt.Println("━━━━━━━━ 慢速发送模式 ━━━━━━━━")
	fmt.Printf("📝 策略：同时建立 %d 个连接，每个连接慢慢发送HTTP请求\n", connCount)
	fmt.Printf("⏱️  每个字节间隔 %dms，让连接保持活跃更长时间\n", *delay)
	fmt.Println("💡 这样你就能在服务器看到多个连接同时存在！\n")

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 1; i <= connCount; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()
			slowSendRequest(connID, byteDelay)
		}(i)

		// 每个连接间隔200ms启动
		time.Sleep(200 * time.Millisecond)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	fmt.Printf("\n✅ 所有连接完成！\n")
	fmt.Printf("📊 统计: 总耗时 %v, 平均每连接 %v\n",
		elapsed.Round(time.Millisecond),
		(elapsed / time.Duration(connCount)).Round(time.Millisecond))
}

func slowSendRequest(connID int, byteDelay time.Duration) {
	conn, err := net.Dial("tcp", *host)
	if err != nil {
		fmt.Printf("❌ [连接 %d] 失败: %v\n", connID, err)
		return
	}
	defer conn.Close()

	if *verbose {
		fmt.Printf("🔵 [连接 %d] 已建立\n", connID)
	}

	request := "GET / HTTP/1.1\r\n" +
		"Host: localhost\r\n" +
		"User-Agent: SlowClient\r\n" +
		"Connection: close\r\n" +
		"\r\n"

	fmt.Printf("📤 [连接 %d] 开始慢速发送请求（每字节间隔 %dms）...\n", connID, byteDelay.Milliseconds())

	for i, char := range []byte(request) {
		_, err := conn.Write([]byte{char})
		if err != nil {
			fmt.Printf("❌ [连接 %d] 发送失败: %v\n", connID, err)
			return
		}

		if *verbose && (i+1)%10 == 0 {
			fmt.Printf("   [连接 %d] 已发送 %d/%d 字节...\n", connID, i+1, len(request))
		}

		time.Sleep(byteDelay)
	}

	fmt.Printf("✅ [连接 %d] 请求发送完成！总计 %d 字节\n", connID, len(request))

	// 接收响应
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	response := make([]byte, 4096)
	n, err := conn.Read(response)
	if err != nil && err != io.EOF {
		fmt.Printf("❌ [连接 %d] 读取响应失败: %v\n", connID, err)
		return
	}

	fmt.Printf("✅ [连接 %d] 收到响应 %d 字节\n", connID, n)
	if *verbose {
		fmt.Printf("🔴 [连接 %d] 连接关闭\n", connID)
	}
}

// ==================== 模式 2: 保持连接 ====================

func keepAliveMode() {
	connCount := getConnCount(10)
	testDuration := getDuration(30)

	fmt.Println("━━━━━━━━ 保持连接模式 ━━━━━━━━")
	fmt.Printf("📝 策略：同时建立 %d 个连接，建立后保持不关闭\n", connCount)
	fmt.Printf("⏱️  持续 %d 秒，让你观察服务器如何管理长连接\n", testDuration)
	fmt.Println("💡 注意观察服务器的定时器如何检查超时连接！\n")

	connections := make([]net.Conn, connCount)

	// 建立连接
	for i := 0; i < connCount; i++ {
		conn, err := net.Dial("tcp", *host)
		if err != nil {
			fmt.Printf("❌ [连接 %d] 失败: %v\n", i+1, err)
			continue
		}
		connections[i] = conn
		fmt.Printf("🔵 [连接 %d] 已建立并保持\n", i+1)
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\n✅ 已建立 %d 个连接，保持 %d 秒...\n", connCount, testDuration)
	fmt.Println("💡 观察服务器端的连接数和超时检查！\n")

	// 定时打印状态
	ticker := time.NewTicker(5 * time.Second)
	done := time.After(time.Duration(testDuration) * time.Second)

	for {
		select {
		case <-ticker.C:
			activeCount := 0
			for i, conn := range connections {
				if conn != nil {
					conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
					_, err := conn.Write([]byte{})
					if err == nil {
						activeCount++
					} else {
						fmt.Printf("🔴 [连接 %d] 已断开\n", i+1)
						connections[i] = nil
					}
				}
			}
			fmt.Printf("📊 当前活跃连接数: %d/%d\n", activeCount, connCount)

		case <-done:
			fmt.Println("\n⏰ 时间到，关闭所有连接...")
			for i, conn := range connections {
				if conn != nil {
					conn.Close()
					if *verbose {
						fmt.Printf("🔴 [连接 %d] 已关闭\n", i+1)
					}
				}
			}
			ticker.Stop()
			fmt.Println("\n✅ 测试完成！")
			return
		}
	}
}

// ==================== 模式 3: 突发 ====================

func burstMode() {
	connCount := getConnCount(20)

	fmt.Println("━━━━━━━━ 突发模式 ━━━━━━━━")
	fmt.Printf("📝 策略：同时建立 %d 个连接并立即发送请求\n", connCount)
	fmt.Println("💡 观察服务器如何一次性处理多个并发事件！\n")

	var wg sync.WaitGroup
	var successCount, failCount int32
	startTime := time.Now()

	fmt.Printf("🔥 开始突发攻击！同时建立 %d 个连接...\n\n", connCount)

	for i := 1; i <= connCount; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()

			connStart := time.Now()
			conn, err := net.Dial("tcp", *host)
			if err != nil {
				fmt.Printf("❌ [连接 %d] 失败: %v\n", connID, err)
				atomic.AddInt32(&failCount, 1)
				return
			}
			defer conn.Close()

			if *verbose {
				fmt.Printf("🔵 [连接 %d] 已建立\n", connID)
			}

			// 立即发送完整请求
			request := "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"
			conn.Write([]byte(request))

			// 读取响应
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			response := make([]byte, 4096)
			n, err := conn.Read(response)
			if err != nil && err != io.EOF {
				fmt.Printf("❌ [连接 %d] 读取失败: %v\n", connID, err)
				atomic.AddInt32(&failCount, 1)
				return
			}

			latency := time.Since(connStart)
			atomic.AddInt32(&successCount, 1)
			fmt.Printf("✅ [连接 %d] 完成 (响应: %d 字节, 延迟: %v)\n",
				connID, n, latency.Round(time.Millisecond))
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	fmt.Printf("\n━━━━━━━━ 统计结果 ━━━━━━━━\n")
	fmt.Printf("✅ 成功: %d\n", successCount)
	fmt.Printf("❌ 失败: %d\n", failCount)
	fmt.Printf("⏱️  总耗时: %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("📊 平均每连接: %v\n", (elapsed / time.Duration(connCount)).Round(time.Millisecond))
	fmt.Printf("🚀 QPS: %.0f\n", float64(successCount)/elapsed.Seconds())
}

// ==================== 模式 4: 压力测试 ====================

func stressMode() {
	connCount := getConnCount(100)
	testDuration := getDuration(30)

	fmt.Println("━━━━━━━━ 压力测试模式 ━━━━━━━━")
	fmt.Printf("📝 策略：持续 %d 秒发送 %d 个并发请求\n", testDuration, connCount)
	fmt.Println("💡 观察服务器在持续高负载下的表现！\n")

	var wg sync.WaitGroup
	var successCount, failCount int64
	var totalLatency int64 // 毫秒
	stopChan := make(chan struct{})
	startTime := time.Now()

	// 启动统计协程
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(startTime).Seconds()
				success := atomic.LoadInt64(&successCount)
				fail := atomic.LoadInt64(&failCount)
				avgLatency := time.Duration(atomic.LoadInt64(&totalLatency)/max(success, 1)) * time.Millisecond

				fmt.Printf("📊 [%.0fs] 成功: %d, 失败: %d, QPS: %.0f, 平均延迟: %v\n",
					elapsed, success, fail, float64(success)/elapsed, avgLatency)
			case <-stopChan:
				return
			}
		}
	}()

	// 发送请求
	for i := 1; i <= connCount; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()

			reqStart := time.Now()
			conn, err := net.Dial("tcp", *host)
			if err != nil {
				if *verbose {
					fmt.Printf("❌ [连接 %d] 失败: %v\n", connID, err)
				}
				atomic.AddInt64(&failCount, 1)
				return
			}
			defer conn.Close()

			request := "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"
			_, err = conn.Write([]byte(request))
			if err != nil {
				atomic.AddInt64(&failCount, 1)
				return
			}

			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			response := make([]byte, 4096)
			_, err = conn.Read(response)
			if err != nil && err != io.EOF {
				if *verbose {
					fmt.Printf("❌ [连接 %d] 读取失败: %v\n", connID, err)
				}
				atomic.AddInt64(&failCount, 1)
				return
			}

			latency := time.Since(reqStart)
			atomic.AddInt64(&successCount, 1)
			atomic.AddInt64(&totalLatency, latency.Milliseconds())

			if *verbose {
				fmt.Printf("✅ [连接 %d] 完成 (延迟: %v)\n", connID, latency.Round(time.Millisecond))
			}
		}(i)
	}

	// 等待指定时间
	time.Sleep(time.Duration(testDuration) * time.Second)
	close(stopChan)
	wg.Wait()
	elapsed := time.Since(startTime)

	// 最终统计
	success := atomic.LoadInt64(&successCount)
	fail := atomic.LoadInt64(&failCount)
	avgLatency := time.Duration(atomic.LoadInt64(&totalLatency)/max(success, 1)) * time.Millisecond

	fmt.Printf("\n━━━━━━━━ 最终统计 ━━━━━━━━\n")
	fmt.Printf("✅ 成功: %d\n", success)
	fmt.Printf("❌ 失败: %d\n", fail)
	fmt.Printf("⏱️  总耗时: %v\n", elapsed.Round(time.Second))
	fmt.Printf("🚀 平均 QPS: %.0f\n", float64(success)/elapsed.Seconds())
	fmt.Printf("📊 平均延迟: %v\n", avgLatency)
	fmt.Printf("📈 成功率: %.2f%%\n", float64(success)/float64(success+fail)*100)
}

// ==================== 模式 5: 流水线 ====================

func pipelineMode() {
	connCount := getConnCount(10)
	requestsPerConn := 10

	fmt.Println("━━━━━━━━ 流水线模式 ━━━━━━━━")
	fmt.Printf("📝 策略：%d 个连接，每个连接发送 %d 个请求\n", connCount, requestsPerConn)
	fmt.Println("💡 观察同一连接上的多个请求处理！\n")

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 1; i <= connCount; i++ {
		wg.Add(1)
		go func(connID int) {
			defer wg.Done()

			conn, err := net.Dial("tcp", *host)
			if err != nil {
				fmt.Printf("❌ [连接 %d] 失败: %v\n", connID, err)
				return
			}
			defer conn.Close()

			fmt.Printf("🔵 [连接 %d] 已建立，准备发送 %d 个请求\n", connID, requestsPerConn)

			for req := 1; req <= requestsPerConn; req++ {
				request := fmt.Sprintf("GET /?req=%d HTTP/1.1\r\nHost: localhost\r\nConnection: keep-alive\r\n\r\n", req)
				_, err := conn.Write([]byte(request))
				if err != nil {
					fmt.Printf("❌ [连接 %d] 请求 %d 发送失败: %v\n", connID, req, err)
					return
				}

				// 读取响应
				conn.SetReadDeadline(time.Now().Add(2 * time.Second))
				response := make([]byte, 4096)
				n, err := conn.Read(response)
				if err != nil && err != io.EOF {
					fmt.Printf("❌ [连接 %d] 请求 %d 读取失败: %v\n", connID, req, err)
					return
				}

				if *verbose {
					fmt.Printf("   [连接 %d] 请求 %d/%d 完成 (%d 字节)\n", connID, req, requestsPerConn, n)
				}

				time.Sleep(100 * time.Millisecond) // 请求间隔
			}

			fmt.Printf("✅ [连接 %d] 所有 %d 个请求完成\n", connID, requestsPerConn)
		}(i)

		time.Sleep(100 * time.Millisecond) // 连接间隔
	}

	wg.Wait()
	elapsed := time.Since(startTime)
	totalRequests := connCount * requestsPerConn

	fmt.Printf("\n━━━━━━━━ 统计结果 ━━━━━━━━\n")
	fmt.Printf("📊 总连接数: %d\n", connCount)
	fmt.Printf("📊 总请求数: %d\n", totalRequests)
	fmt.Printf("⏱️  总耗时: %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("🚀 平均 QPS: %.0f\n", float64(totalRequests)/elapsed.Seconds())
}

// ==================== 模式 6: 渐进增加 ====================

func gradualMode() {
	increment := 5                   // 每秒增加的连接数
	maxConns := getConnCount(50)     // 最大连接数
	testDuration := getDuration(20)  // 测试时长

	fmt.Println("━━━━━━━━ 渐进模式 ━━━━━━━━")
	fmt.Printf("📝 策略：每秒增加 %d 个连接，最多 %d 个，持续 %d 秒\n", increment, maxConns, testDuration)
	fmt.Println("💡 观察服务器如何应对逐渐增长的负载！\n")

	var wg sync.WaitGroup
	var currentConns int32
	stopChan := make(chan struct{})
	startTime := time.Now()

	// 渐进建立连接
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		connID := 1
		for {
			select {
			case <-ticker.C:
				for i := 0; i < increment && int(atomic.LoadInt32(&currentConns)) < maxConns; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()
						atomic.AddInt32(&currentConns, 1)
						defer atomic.AddInt32(&currentConns, -1)

						conn, err := net.Dial("tcp", *host)
						if err != nil {
							if *verbose {
								fmt.Printf("❌ [连接 %d] 失败: %v\n", id, err)
							}
							return
						}
						defer conn.Close()

						if *verbose {
							fmt.Printf("🔵 [连接 %d] 已建立\n", id)
						}

						// 发送请求
						request := "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"
						conn.Write([]byte(request))

						// 读取响应
						response := make([]byte, 4096)
						conn.Read(response)

						// 保持连接一段时间
						time.Sleep(5 * time.Second)
					}(connID)
					connID++
				}

				fmt.Printf("📊 [%2.0fs] 当前连接数: %d\n",
					time.Since(startTime).Seconds(),
					atomic.LoadInt32(&currentConns))

			case <-stopChan:
				return
			}
		}
	}()

	// 等待测试时长
	time.Sleep(time.Duration(testDuration) * time.Second)
	close(stopChan)
	wg.Wait()

	fmt.Println("\n✅ 渐进测试完成！")
}

// ==================== 辅助函数 ====================

func getConnCount(defaultValue int) int {
	if *connections > 0 {
		return *connections
	}
	return defaultValue
}

func getDuration(defaultValue int) int {
	if *duration > 0 {
		return *duration
	}
	return defaultValue
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

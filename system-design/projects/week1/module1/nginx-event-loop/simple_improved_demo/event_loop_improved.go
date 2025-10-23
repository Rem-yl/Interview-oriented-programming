package main

import (
	"fmt"
	"net"
	"time"
)

/*
改进版事件循环 - 解决新连接处理问题

主要改进：
1. 每次循环接受多个新连接（而不是只接受1个）
2. 限制每次最多接受10个（避免饥饿）
3. 添加详细的统计信息

对比原始版本：
- 原始版：每次只接受1个新连接 → 20个并发需要20圈
- 改进版：每次接受最多10个   → 20个并发只需要2圈
*/

func main() {
	fmt.Println("🚀 改进版事件循环演示")
	fmt.Println("✨ 改进：批量接受新连接，减少延迟")
	fmt.Println("访问 http://localhost:8080")
	fmt.Println()

	// 创建监听器
	listener, _ := net.Listen("tcp", ":8080")
	defer listener.Close()

	// 存储所有活跃连接
	connections := make(map[int]*SimpleConnection)
	nextID := 1

	// 上次检查超时的时间
	lastTimeoutCheck := time.Now()

	// 统计信息
	stats := &Stats{
		totalAccepted:   0,
		totalProcessed:  0,
		totalClosed:     0,
		maxConcurrent:   0,
		acceptPerLoop:   make([]int, 0),
	}

	fmt.Println("✅ 事件循环已启动\n")

	// ========== 核心：无限事件循环 ==========
	loopCount := 0
	for {
		loopCount++
		fmt.Printf("━━━━━━━━ 事件循环第 %d 圈 ━━━━━━━━\n", loopCount)

		// ========== 步骤 1: 处理定时器 ==========
		if time.Since(lastTimeoutCheck) > 5*time.Second {
			fmt.Println("⏰ 定时器触发: 检查超时连接")
			checkTimeouts(connections, stats)
			lastTimeoutCheck = time.Now()
		}

		// ========== 步骤 2: 批量接受新连接（改进点！）==========
		maxAcceptPerLoop := 10  // 每次循环最多接受10个新连接
		acceptedCount := 0

		fmt.Printf("📥 尝试接受新连接（最多 %d 个）...\n", maxAcceptPerLoop)

		for i := 0; i < maxAcceptPerLoop; i++ {
			// 设置一个很短的超时，避免阻塞
			listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Millisecond))
			conn, err := listener.Accept()

			if err != nil {
				// 没有更多新连接了
				break
			}

			// 接受新连接
			acceptedCount++
			stats.totalAccepted++

			fmt.Printf("   🔵 新连接到达 (ID=%d)\n", nextID)
			connections[nextID] = &SimpleConnection{
				id:        nextID,
				conn:      conn,
				state:     "reading",
				buffer:    make([]byte, 0),
				createdAt: time.Now(),
			}
			nextID++
		}

		// 更新统计
		stats.acceptPerLoop = append(stats.acceptPerLoop, acceptedCount)

		if acceptedCount > 0 {
			fmt.Printf("   ✅ 本轮接受了 %d 个新连接\n", acceptedCount)
		} else {
			fmt.Printf("   ⏭️  无新连接\n")
		}

		// 更新最大并发数
		currentConcurrent := len(connections)
		if currentConcurrent > stats.maxConcurrent {
			stats.maxConcurrent = currentConcurrent
		}

		// ========== 步骤 3: 处理所有现有连接 ==========
		fmt.Printf("📌 处理现有连接（共 %d 个）...\n", len(connections))

		processedCount := 0
		for id, sc := range connections {
			fmt.Printf("   [连接 %d] 状态: %s\n", id, sc.state)

			switch sc.state {
			case "reading":
				// 尝试读取数据（非阻塞）
				sc.conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
				buffer := make([]byte, 1024)
				n, err := sc.conn.Read(buffer)

				if err != nil {
					// 暂时没数据，跳过
					continue
				}

				if n > 0 {
					sc.buffer = append(sc.buffer, buffer[:n]...)
					fmt.Printf("      📖 读取 %d 字节 (总计: %d)\n", n, len(sc.buffer))

					// 检查是否读完（简化：以 \r\n\r\n 结尾）
					if len(sc.buffer) >= 4 {
						tail := sc.buffer[len(sc.buffer)-4:]
						if string(tail) == "\r\n\r\n" {
							sc.state = "processing"
							fmt.Printf("      ✅ 请求读取完成\n")
							processedCount++
						}
					}
				}

			case "processing":
				// 处理请求（这里简化：直接生成响应）
				fmt.Printf("      ⚙️  处理请求...\n")
				sc.response = []byte("HTTP/1.1 200 OK\r\n" +
					"Content-Type: text/plain\r\n" +
					"Content-Length: 23\r\n\r\n" +
					"Hello from event loop!\n")
				sc.state = "writing"
				fmt.Printf("      ✅ 响应已准备\n")
				stats.totalProcessed++

			case "writing":
				// 尝试写入响应（非阻塞）
				sc.conn.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))
				n, err := sc.conn.Write(sc.response)

				if err != nil {
					continue
				}

				if n > 0 {
					sc.response = sc.response[n:]
					fmt.Printf("      📝 写入 %d 字节 (剩余: %d)\n", n, len(sc.response))

					if len(sc.response) == 0 {
						sc.state = "closed"
						fmt.Printf("      ✅ 响应发送完成\n")
					}
				}

			case "closed":
				// 关闭连接
				sc.conn.Close()
				delete(connections, id)
				stats.totalClosed++
				fmt.Printf("      🔴 连接已关闭\n")
			}
		}

		if processedCount > 0 {
			fmt.Printf("   ✅ 本轮处理了 %d 个连接\n", processedCount)
		}

		// ========== 步骤 4: 显示统计信息 ==========
		fmt.Printf("\n📊 统计信息:\n")
		fmt.Printf("   当前连接数: %d\n", len(connections))
		fmt.Printf("   累计接受: %d, 累计处理: %d, 累计关闭: %d\n",
			stats.totalAccepted, stats.totalProcessed, stats.totalClosed)
		fmt.Printf("   最大并发: %d\n", stats.maxConcurrent)

		// 显示最近5圈的接受数量
		if len(stats.acceptPerLoop) > 0 {
			recentAccepts := stats.acceptPerLoop
			if len(recentAccepts) > 5 {
				recentAccepts = recentAccepts[len(recentAccepts)-5:]
			}
			fmt.Printf("   最近接受: %v\n", recentAccepts)
		}

		// ========== 步骤 5: 短暂休眠 ==========
		fmt.Printf("\n💤 休眠 500ms...\n\n")
		time.Sleep(500 * time.Millisecond)
	}
}

// SimpleConnection 简化的连接对象
type SimpleConnection struct {
	id        int
	conn      net.Conn
	state     string // "reading", "processing", "writing", "closed"
	buffer    []byte
	response  []byte
	createdAt time.Time
}

// Stats 统计信息
type Stats struct {
	totalAccepted  int
	totalProcessed int
	totalClosed    int
	maxConcurrent  int
	acceptPerLoop  []int
}

// checkTimeouts 检查并关闭超时连接
func checkTimeouts(connections map[int]*SimpleConnection, stats *Stats) {
	now := time.Now()
	timeoutCount := 0

	for id, sc := range connections {
		if now.Sub(sc.createdAt) > 30*time.Second {
			fmt.Printf("   ⏱️  连接 %d 超时，关闭\n", id)
			sc.conn.Close()
			delete(connections, id)
			stats.totalClosed++
			timeoutCount++
		}
	}

	if timeoutCount > 0 {
		fmt.Printf("   ⏱️  本次超时检查关闭了 %d 个连接\n", timeoutCount)
	} else {
		fmt.Printf("   ✅ 无超时连接\n")
	}
}

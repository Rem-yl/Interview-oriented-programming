package main

import (
	"fmt"
	"net"
	"time"
)

/*
这是一个**简化版**的事件循环演示，不使用 epoll，便于理解核心概念。

NGINX 事件循环的核心思想：
1. 一个循环不停运行
2. 每次循环检查哪些连接"准备好了"
3. 只处理准备好的连接，从不等待
4. 用状态机跟踪每个连接的状态
*/

func main() {
	fmt.Println("🚀 简化版事件循环演示")
	fmt.Println("访问 http://localhost:8080")
	fmt.Println()

	// 创建监听器
	listener, _ := net.Listen("tcp", ":8080")
	defer listener.Close()

	// 设置为非阻塞模式（这样 Accept 不会阻塞）
	listener.(*net.TCPListener).SetDeadline(time.Now().Add(100 * time.Millisecond))

	// 存储所有活跃连接
	connections := make(map[int]*SimpleConnection)
	nextID := 1

	// 上次检查超时的时间
	lastTimeoutCheck := time.Now()

	fmt.Println("✅ 事件循环已启动\n")

	// ========== 核心：无限事件循环 ==========
	loopCount := 0
	for {
		loopCount++
		fmt.Printf("━━━━━━━━ 事件循环第 %d 圈 ━━━━━━━━\n", loopCount)

		// ========== 步骤 1: 处理定时器 ==========
		if time.Since(lastTimeoutCheck) > 5*time.Second {
			fmt.Println("⏰ 定时器触发: 检查超时连接")
			checkTimeouts(connections)
			lastTimeoutCheck = time.Now()
		}

		// ========== 步骤 2: 尝试接受新连接 ==========
		// 使用非阻塞 Accept，如果没有新连接就立即返回
		listener.(*net.TCPListener).SetDeadline(time.Now().Add(10 * time.Millisecond))
		if conn, err := listener.Accept(); err == nil {
			fmt.Printf("🔵 新连接到达 (ID=%d)\n", nextID)
			connections[nextID] = &SimpleConnection{
				id:        nextID,
				conn:      conn,
				state:     "reading",
				buffer:    make([]byte, 0),
				createdAt: time.Now(),
			}
			nextID++
		}

		// ========== 步骤 3: 处理所有现有连接 ==========
		for id, sc := range connections {
			fmt.Printf("📌 检查连接 %d (状态: %s)\n", id, sc.state)

			switch sc.state {
			case "reading":
				// 尝试读取数据（非阻塞）
				sc.conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
				buffer := make([]byte, 1024)
				n, err := sc.conn.Read(buffer)

				if err != nil {
					// 暂时没数据，跳过
					fmt.Printf("   ⏭️  暂无数据可读\n")
					continue
				}

				if n > 0 {
					sc.buffer = append(sc.buffer, buffer[:n]...)
					fmt.Printf("   📖 读取 %d 字节 (总计: %d)\n", n, len(sc.buffer))

					// 检查是否读完（简化：看是否有 \n）
					if len(sc.buffer) > 0 && sc.buffer[len(sc.buffer)-1] == '\n' {
						sc.state = "processing"
						fmt.Printf("   ✅ 请求读取完成，切换到处理状态\n")
					}
				}

			case "processing":
				// 处理请求（这里简化：直接生成响应）
				fmt.Printf("   ⚙️  处理请求...\n")
				sc.response = []byte("HTTP/1.1 200 OK\r\n" +
					"Content-Type: text/plain\r\n" +
					"Content-Length: 23\r\n\r\n" +
					"Hello from event loop!\n")
				sc.state = "writing"
				fmt.Printf("   ✅ 响应已准备，切换到写入状态\n")

			case "writing":
				// 尝试写入响应（非阻塞）
				sc.conn.SetWriteDeadline(time.Now().Add(10 * time.Millisecond))
				n, err := sc.conn.Write(sc.response)

				if err != nil {
					fmt.Printf("   ⏭️  暂时无法写入\n")
					continue
				}

				if n > 0 {
					sc.response = sc.response[n:]
					fmt.Printf("   📝 写入 %d 字节 (剩余: %d)\n", n, len(sc.response))

					if len(sc.response) == 0 {
						sc.state = "closed"
						fmt.Printf("   ✅ 响应发送完成\n")
					}
				}

			case "closed":
				// 关闭连接
				sc.conn.Close()
				delete(connections, id)
				fmt.Printf("   🔴 连接已关闭 (剩余连接: %d)\n", len(connections))
			}
		}

		// ========== 步骤 4: 短暂休眠（模拟 epoll_wait）==========
		fmt.Printf("💤 当前连接数: %d, 休眠 500ms...\n\n", len(connections))
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

// checkTimeouts 检查并关闭超时连接
func checkTimeouts(connections map[int]*SimpleConnection) {
	now := time.Now()
	for id, sc := range connections {
		// 从创建到现在超过 30s 就关闭连接
		if now.Sub(sc.createdAt) > 30*time.Second {
			fmt.Printf("   ⏱️  连接 %d 超时，关闭\n", id)
			sc.conn.Close()
			delete(connections, id)
		}
	}
}

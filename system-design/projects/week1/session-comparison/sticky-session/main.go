/*
Sticky Session 实现 - 基于 Nginx ip_hash 的会话管理方案

项目背景:
  这是系统设计课程 Week1 Module1.4 会话管理对比实验的第一个方案。
  通过实现三种不同的会话管理方式（Sticky Session、Redis Session、JWT Token），
  深入理解它们的工作原理、性能差异和适用场景。

方案特点:
  - Session 存储在服务器本地内存中 (sync.Map)
  - 依赖 Nginx ip_hash 确保同一客户端总是路由到同一台服务器
  - 优点: 简单、高性能、无需外部依赖
  - 缺点: 服务器宕机会丢失 Session、扩展性差

环境变量:
  SERVER_ID - 服务器唯一标识 (默认: "server-default")
  PORT      - 监听端口 (默认: "8080")

示例:
  PORT=8081 SERVER_ID=server-1 go run main.go
  PORT=8082 SERVER_ID=server-2 go run main.go
  PORT=8083 SERVER_ID=server-3 go run main.go

API 端点:
  POST /login           - 登录，创建 Session
    Request:  {"username": "alice", "password": "123456"}
    Response: {"status": "ok", "data": "get user: alice"}
    Cookie:   session_id=<uuid>

  GET /profile          - 获取用户信息 (需要 Cookie)
    Response: {"username": "alice", "login_time": "...", "server_id": "server-1"}

  GET /debug/sessions   - 查看服务器上的所有 Session (调试用)
    Response: {"server_id": "server-1", "session_count": 3, "sessions": [...]}

如何测试:
  方法 1: 使用 pytest 测试套件 (推荐)
    cd ../test-scripts
    pytest test_sticky_session.py -v

  方法 2: 手动测试
    # 登录
    curl -c cookies.txt -X POST http://localhost:8080/login \
      -H "Content-Type: application/json" \
      -d '{"username":"alice","password":"123456"}'

    # 访问 profile
    curl -b cookies.txt http://localhost:8080/profile

    # 查看 Session
    curl http://localhost:8081/debug/sessions

Nginx 配置:
  使用 docker/nginx-sticky.conf (ip_hash 算法)

  启动 Nginx:
    docker run -d --name nginx-sticky -p 8080:80 \
      -v $(pwd)/docker/nginx-sticky.conf:/etc/nginx/conf.d/default.conf:ro \
      nginx:alpine

实验要点:
  1. 观察 ip_hash 如何实现 Sticky Session (同一客户端总是路由到同一台服务器)
  2. 对比 round_robin 算法 (请求会分散到不同服务器，导致 Session 失效)
  3. 测试服务器宕机场景 (Session 丢失)
  4. 理解为什么需要 ip_hash 或共享存储 (Redis)

相关文件:
  - ../test-scripts/test_sticky_session.py     - pytest 测试套件
  - ../test-scripts/verify_nginx.py            - Nginx 验证脚本
  - ../docker/nginx-sticky.conf                - ip_hash 配置
  - ../docker/nginx-round-robin.conf           - round_robin 对比配置
  - ../EXPERIMENT_GUIDE.md                     - 完整实验指南
*/

package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getEnv 从环境变量读取配置，如果不存在则使用默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

// 服务器配置
var serverID = getEnv("SERVER_ID", "server-default") // 服务器唯一标识
var port = getEnv("PORT", "8080")                    // 监听端口

// sessionStore 本地内存存储，使用 sync.Map 保证并发安全
// 关键点: 每个服务器实例独立存储，不共享
var sessionStore sync.Map

// Session 会话数据结构
type Session struct {
	UserName  string    // 用户名
	LoginTime time.Time // 登录时间
	ServerID  string    // 创建此 Session 的服务器 ID
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	UserName string `json:"username"`
	PassWord string `json:"password"`
}

// loginHandler 处理登录请求
// 1. 生成唯一 Session ID (UUID)
// 2. 创建 Session 对象并存储到本地内存
// 3. 设置 Cookie (session_id)，客户端后续请求会携带此 Cookie
func loginHandler(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成唯一的 Session ID
	sessionID := uuid.New().String()

	// 创建 Session 对象
	session := Session{
		UserName:  req.UserName,
		LoginTime: time.Now(),
		ServerID:  serverID, // 记录是哪个服务器创建的
	}

	// 存储到本地内存
	sessionStore.Store(sessionID, session)

	// 设置 Cookie
	// 参数: name, value, maxAge(秒), path, domain, secure, httpOnly
	// domain="" 表示自动匹配当前域名，适用于 localhost 和 127.0.0.1
	c.SetCookie("session_id", sessionID, 3600, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": fmt.Sprintf("get user: %s", req.UserName)})
}

// profileHandler 获取用户信息
// 1. 从 Cookie 中读取 session_id
// 2. 从本地内存查找对应的 Session
// 3. 如果找到，返回用户信息；否则返回 401 未认证
//
// 关键点: 只能找到本服务器上的 Session，其他服务器的 Session 找不到
// 这就是为什么需要 Nginx ip_hash 确保同一客户端总是路由到同一台服务器
func profileHandler(c *gin.Context) {
	// 1. 从 Cookie 读取 session_id
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		// 没有 Cookie，返回 401
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// 2. 从本地内存查找 Session
	value, ok := sessionStore.Load(sessionID)
	if !ok {
		// Session 不存在（可能是：其他服务器创建的、已过期、无效的 session_id）
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found"})
		return
	}

	// 3. 类型断言，获取 Session 对象
	session := value.(Session)

	// 4. 返回用户信息
	c.JSON(http.StatusOK, gin.H{
		"username":   session.UserName,
		"login_time": session.LoginTime,
		"server_id":  session.ServerID, // 客户端可以看到是哪个服务器处理的
	})

}

// debugSessionHandler 调试接口 - 查看当前服务器上的所有 Session
// 用于实验观察：
// 1. 验证 Session 存储在本地内存
// 2. 观察不同服务器上的 Session 是隔离的
// 3. 测试 Nginx 负载均衡效果
func debugSessionHandler(c *gin.Context) {
	sessions := []map[string]interface{}{}

	// 遍历本服务器上的所有 Session
	sessionStore.Range(func(key, value any) bool {
		session := value.(Session)
		sessions = append(sessions, map[string]interface{}{
			"session_id": key.(string),
			"username":   session.UserName,
			"login_time": session.LoginTime,
			"server_id":  session.ServerID,
		})
		return true // 继续遍历
	})

	// 返回统计信息
	c.JSON(http.StatusOK, gin.H{
		"server_id":     serverID,      // 当前服务器 ID
		"session_count": len(sessions), // Session 数量
		"sessions":      sessions,      // Session 列表
	})
}

func main() {
	r := gin.Default()

	// 注册路由
	r.POST("/login", loginHandler)                // 登录
	r.GET("/profile", profileHandler)             // 获取用户信息
	r.GET("/debug/sessions", debugSessionHandler) // 调试接口

	// 启动服务器
	fmt.Printf("🚀 Server %s starting on port %s...\n", serverID, port)
	r.Run(":" + port)
}

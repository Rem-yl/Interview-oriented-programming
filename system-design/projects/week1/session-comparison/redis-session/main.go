/*
Redis Session 实现 - 基于共享存储的会话管理方案

项目背景:
  这是系统设计课程 Week1 Module1.4 会话管理对比实验的第二个方案。
  与 Sticky Session 不同，Redis Session 将会话数据存储在集中式的 Redis 中，
  所有服务器共享同一份 Session 数据，实现真正的无状态服务。

方案特点:
  - Session 存储在 Redis 集中式存储中（所有服务器共享）
  - 支持 Nginx Round Robin 负载均衡（无需 ip_hash）
  - 优点: 高可用、可扩展、服务器宕机不影响 Session
  - 缺点: 依赖 Redis、每次请求需要网络 I/O（增加延迟 ~1-2ms）

环境变量:
  PORT      - 监听端口 (默认: "8091")
  SERVERID  - 服务器唯一标识 (默认: "server-default")

依赖服务:
  Redis - 必须先启动 Redis 服务
    docker run -d --name redis -p 6379:6379 redis:alpine

示例:
  # 启动 Redis
  docker run -d --name redis -p 6379:6379 redis:alpine

  # 启动 3 个后端服务器
  PORT=8091 SERVERID=server-1 go run main.go
  PORT=8092 SERVERID=server-2 go run main.go
  PORT=8093 SERVERID=server-3 go run main.go

  # 启动 Nginx (Round Robin)
  docker run -d --name nginx-redis -p 8090:80 \
    -v $(pwd)/../docker/nginx-redis.conf:/etc/nginx/conf.d/default.conf:ro \
    nginx:alpine

API 端点:
  POST /login           - 登录，创建 Session 并存储到 Redis
    Request:  {"username": "alice", "password": "123456"}
    Response: {"data": "user: alice login success!"}
    Cookie:   sessionID=<uuid>
    Redis:    SET session:<uuid> <session_data> EX 1800

  GET /profile          - 获取用户信息（从 Redis 读取）
    Request:  Cookie: sessionID=<uuid>
    Response: {"sessionID": "...", "loginTime": "...", "serverID": "server-1"}
    Redis:    GET session:<uuid> + EXPIRE session:<uuid> 1800

如何测试:
  方法 1: 使用 pytest 测试套件 (推荐)
    cd ../test-scripts
    pytest test_redis_session.py -v

  方法 2: 手动测试 - 验证跨服务器 Session 共享
    # 登录到 server-1
    curl -c cookies.txt -X POST http://localhost:8091/login \
      -H "Content-Type: application/json" \
      -d '{"username":"alice","password":"123456"}'

    # 访问 server-2（应该也能获取到 Session）
    curl -b cookies.txt http://localhost:8092/profile

    # 通过 Nginx (Round Robin) 访问
    curl -b cookies.txt http://localhost:8090/profile

  方法 3: 使用 Redis CLI 查看数据
    redis-cli
    KEYS session:*              # 查看所有 Session
    GET session:<uuid>          # 查看具体 Session 内容
    TTL session:<uuid>          # 查看剩余过期时间

Nginx 配置:
  使用 docker/nginx-redis.conf (Round Robin 算法)

  关键区别:
    - 不使用 ip_hash（因为所有服务器共享 Redis）
    - 请求可以路由到任意服务器
    - 每个服务器都能从 Redis 获取 Session

实验要点:
  1. 观察 Session 如何存储在 Redis 中（Key-Value 格式）
  2. 验证跨服务器 Session 共享（登录 server-1，访问 server-2）
  3. 对比 Round Robin 和 ip_hash 的区别
  4. 测试服务器宕机场景（Session 不会丢失）
  5. 观察 Redis 网络延迟对性能的影响

Redis 数据结构:
  Key:   session:<session_id>
  Value: {"session_id":"...","login_time":"...","server_id":"..."}
  TTL:   1800 秒 (30 分钟)

对比 Sticky Session:
  ┌────────────────┬─────────────────┬─────────────────┐
  │ 特性           │ Sticky Session  │ Redis Session   │
  ├────────────────┼─────────────────┼─────────────────┤
  │ 存储位置       │ 服务器本地内存  │ Redis 集中存储  │
  │ 跨服务器共享   │ ❌ 不支持       │ ✅ 支持         │
  │ Nginx 算法     │ ip_hash         │ round_robin     │
  │ 服务器宕机     │ Session 丢失    │ Session 保留    │
  │ 网络延迟       │ ~0.1ms          │ ~1-2ms          │
  │ 依赖外部服务   │ ❌ 无           │ ✅ Redis        │
  └────────────────┴─────────────────┴─────────────────┘

相关文件:
  - ../test-scripts/test_redis_session.py    - pytest 测试套件
  - ../docker/nginx-redis.conf               - Round Robin 配置
  - ../EXPERIMENT_GUIDE.md                   - 完整实验指南 (阶段三)
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
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
var port = getEnv("PORT", "8091")                       // 监听端口（默认 8091）
var serverID = getEnv("SERVERID", "server-default")     // 服务器唯一标识

// redis_client Redis 客户端
// 关键点: 所有服务器共享同一个 Redis 实例
var redis_client = redis.NewClient(&redis.Options{
	Addr: "localhost:6379", // Redis 地址
	DB:   0,                // 数据库编号
})

// Session 会话数据结构
// 与 Sticky Session 不同，这个结构会被序列化后存储到 Redis
type Session struct {
	SessionID string    `json:"session_id"` // Session ID
	LoginTime time.Time `json:"login_time"` // 登录时间
	ServerID  string    `json:"server_id"`  // 创建此 Session 的服务器 ID
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	UserName string `json:"username"`
	PassWord string `json:"password"`
}

// saveSessionToRedis 将 Session 保存到 Redis
// 1. 将 Session 对象序列化为 JSON
// 2. 以 "session:<session_id>" 为 Key 存储到 Redis
// 3. 设置 TTL 为 30 分钟
func saveSessionToRedis(session *Session) error {
	sessionID := session.SessionID
	// 1. 序列化 Session 为 JSON
	data, _ := json.Marshal(session)

	// 2. 构造 Redis Key
	key := "session:" + sessionID

	// 3. 存储到 Redis，设置过期时间 30 分钟
	ctx := context.Background()
	return redis_client.Set(ctx, key, data, 30*time.Minute).Err()
}

// getSessionFromRedis 从 Redis 获取 Session
// 1. 从 Redis 读取 Session 数据
// 2. 反序列化 JSON 为 Session 对象
// 3. 续期: 重新设置 TTL 为 30 分钟（每次访问延长过期时间）
//
// 关键点: 任何服务器都可以读取到 Session（因为存储在共享的 Redis）
func getSessionFromRedis(sessionID string) (*Session, error) {
	// 1. 构造 Redis Key
	key := "session:" + sessionID
	ctx := context.Background()

	// 2. 从 Redis 获取数据
	data, err := redis_client.Get(ctx, key).Result()
	if err != nil {
		// Session 不存在或已过期
		return nil, err
	}

	// 3. 反序列化 JSON 为 Session 对象
	var session Session
	json.Unmarshal([]byte(data), &session)

	// 4. 续期: 重新设置 TTL（类似"活跃保持"）
	redis_client.Expire(ctx, key, 30*time.Minute)

	return &session, nil
}

// loginHandler 处理登录请求
// 1. 生成唯一 Session ID (UUID)
// 2. 创建 Session 对象并保存到 Redis（不是本地内存）
// 3. 设置 Cookie (sessionID)，客户端后续请求会携带此 Cookie
//
// 与 Sticky Session 的区别: Session 存储在 Redis，所有服务器都能访问
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 生成唯一的 Session ID
	sessionID := uuid.New().String()

	// 2. 创建 Session 对象
	session := &Session{
		SessionID: sessionID,
		LoginTime: time.Now(),
		ServerID:  serverID, // 记录是哪个服务器创建的（用于对比实验）
	}

	// 3. 保存到 Redis（而不是本地内存）
	if err := saveSessionToRedis(session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	// 4. 设置 Cookie
	// 注意: Cookie 名称是 sessionID（与 Sticky Session 的 session_id 不同）
	c.SetCookie("sessionID", sessionID, 3600, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"data": fmt.Sprintf("user: %s login success!", req.UserName)})
}

// profileHandler 获取用户信息
// 1. 从 Cookie 中读取 sessionID
// 2. 从 Redis 查找对应的 Session
// 3. 如果找到，返回用户信息；否则返回 401 未认证
//
// 关键点: 可以从 Redis 获取到任何服务器创建的 Session
// 这就是为什么可以使用 Round Robin 负载均衡（无需 ip_hash）
func profileHandler(c *gin.Context) {
	// 1. 从 Cookie 读取 sessionID
	sessionID, err := c.Cookie("sessionID")
	if err != nil {
		// 没有 Cookie，返回 401
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// 2. 从 Redis 读取 Session
	session, err := getSessionFromRedis(sessionID)
	if err != nil {
		// Session 不存在或已过期
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found or expired"})
		return
	}

	// 3. 返回用户信息
	c.JSON(http.StatusOK, gin.H{
		"sessionID": session.SessionID,
		"loginTime": session.LoginTime,
		"serverID":  session.ServerID, // 显示创建 Session 的服务器
		"handledBy": serverID,          // 显示处理此请求的服务器（可能不同）
	})
}

func main() {
	r := gin.Default()

	// 注册路由
	r.POST("/login", loginHandler)    // 登录
	r.GET("/profile", profileHandler)  // 获取用户信息

	// 启动服务器
	fmt.Printf("🚀 Redis Session Server %s starting on port %s...\n", serverID, port)
	fmt.Printf("📦 Redis: localhost:6379\n")
	r.Run(":" + port)
}

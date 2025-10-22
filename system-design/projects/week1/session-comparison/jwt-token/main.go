/*
JWT Token 实现 - 无状态认证方案

项目背景:
  这是系统设计课程 Week1 Module1.4 会话管理对比实验的第三个方案。
  与 Sticky Session 和 Redis Session 不同，JWT Token 是完全无状态的认证方案。
  服务器不需要存储任何会话信息，所有状态都编码在 Token 中。

方案特点:
  - Token 自包含用户信息（无需服务器存储）
  - 完全无状态，易于水平扩展
  - 优点: 无需外部依赖、性能极高、天然支持分布式
  - 缺点: Token 无法主动失效、Token 体积较大、安全性要求高

环境变量:
  PORT      - 监听端口 (默认: "8010")
  SERVERID  - 服务器唯一标识 (默认: "server-default")

依赖服务:
  无 - 这是 JWT 的最大优势，完全无状态

示例:
  # 启动单个服务器
  PORT=8010 SERVERID=server-1 go run main.go

  # 启动多个服务器（演示无状态特性）
  PORT=8010 SERVERID=server-1 go run main.go
  PORT=8011 SERVERID=server-2 go run main.go
  PORT=8012 SERVERID=server-3 go run main.go

API 端点:
  POST /login           - 登录，生成并返回 JWT Token
    Request:  {"username": "alice", "password": "123456"}
    Response: {"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
    说明: Token 中包含 userID, username, 过期时间等信息

  GET /profile          - 获取用户信息（需要 JWT Token）
    Request:  Header: Authorization: Bearer <token>
    Response: {"data": "get user: alice jwt token success, userID: xxx"}
    说明: 通过中间件验证 Token 并解析用户信息

JWT Token 结构:
  Header:  {"alg": "HS256", "typ": "JWT"}
  Payload: {"userid": "...", "username": "...", "exp": ..., "iat": ..., "nbf": ..., "iss": "jwt-session"}
  Signature: HMACSHA256(base64UrlEncode(header) + "." + base64UrlEncode(payload), secretKey)

如何测试:
  方法 1: 使用 pytest 测试套件 (推荐)
    cd ../test-scripts
    pytest test_jwt_token.py -v

  方法 2: 手动测试 - 验证 JWT 认证流程
    # 1. 登录获取 Token
    curl -X POST http://localhost:8010/login \
      -H "Content-Type: application/json" \
      -d '{"username":"alice","password":"123456"}'
    # 响应: {"token": "eyJhbGc..."}

    # 2. 使用 Token 访问 /profile
    curl http://localhost:8010/profile \
      -H "Authorization: Bearer eyJhbGc..."
    # 响应: {"data": "get user: alice jwt token success, userID: xxx"}

    # 3. 验证跨服务器无状态（登录 server-1，访问 server-2）
    # 登录 server-1
    TOKEN=$(curl -s -X POST http://localhost:8010/login \
      -H "Content-Type: application/json" \
      -d '{"username":"alice","password":"123456"}' | jq -r '.token')

    # 访问 server-2（应该也能成功，因为是无状态的）
    curl http://localhost:8011/profile \
      -H "Authorization: Bearer $TOKEN"

  方法 3: 在线解码 JWT Token
    访问 https://jwt.io/
    粘贴 Token，查看 Payload 内容

实验要点:
  1. 理解 JWT 的三部分结构（Header.Payload.Signature）
  2. 验证无状态特性（登录一台服务器，访问另一台服务器）
  3. 观察 Token 过期行为（2 小时后 Token 失效）
  4. 对比与 Session 方案的区别（无需存储、无法主动失效）
  5. 理解安全性要求（secretKey 必须保密、HTTPS 传输）

对比其他方案:
  ┌────────────────┬─────────────────┬─────────────────┬─────────────────┐
  │ 特性           │ Sticky Session  │ Redis Session   │ JWT Token       │
  ├────────────────┼─────────────────┼─────────────────┼─────────────────┤
  │ 存储位置       │ 服务器本地内存  │ Redis 集中存储  │ 客户端（Token） │
  │ 状态           │ 有状态          │ 有状态          │ 无状态          │
  │ 跨服务器共享   │ ❌ 不支持       │ ✅ 支持         │ ✅ 天然支持     │
  │ 依赖外部服务   │ ❌ 无           │ ✅ Redis        │ ❌ 无           │
  │ 可主动失效     │ ✅ 可以         │ ✅ 可以         │ ❌ 不可以       │
  │ 性能           │ 极高            │ 高（网络I/O）   │ 极高            │
  │ 扩展性         │ 差              │ 好              │ 极好            │
  └────────────────┴─────────────────┴─────────────────┴─────────────────┘

安全注意事项:
  1. secretKey 必须足够复杂，生产环境应使用环境变量
  2. Token 必须通过 HTTPS 传输，避免被窃取
  3. Token 过期时间不宜过长（本例 2 小时）
  4. 敏感操作应要求重新认证，不能完全依赖 Token

相关文件:
  - ../test-scripts/test_jwt_token.py      - pytest 测试套件
  - ../test-scripts/verify_jwt_token.py    - 验证脚本
  - ../EXPERIMENT_GUIDE.md                 - 完整实验指南 (阶段四)
*/

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
var port = getEnv("PORT", "8010")           // 监听端口（默认 8010）
var serverID = getEnv("SERVERID", "server-default") // 服务器唯一标识

// JWT 配置常量
const (
	secretKey    = "rem"          // JWT 签名密钥（生产环境应使用环境变量）
	TokenExpiren = 2 * time.Hour  // Token 过期时间（2 小时）
)

// Claims JWT 载荷（Payload）结构
// 包含自定义字段和标准的 JWT 声明
type Claims struct {
	UserID   string `json:"userid"`   // 用户唯一标识
	UserName string `json:"username"` // 用户名
	jwt.RegisteredClaims              // JWT 标准字段（exp, iat, nbf, iss 等）
}

// genJWTToken 生成 JWT Token
// 1. 创建包含用户信息的 Claims
// 2. 设置过期时间、签发时间、生效时间、签发者
// 3. 使用 HS256 算法和 secretKey 签名
// 返回: 完整的 JWT Token 字符串（Header.Payload.Signature）
func genJWTToken(userid, username string) (string, error) {
	claims := &Claims{
		UserID:   userid,
		UserName: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiren)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                   // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                   // 生效时间
			Issuer:    "jwt-session",
		},
	}

	// 创建 Token 对象并签名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// validateJWTToken 验证并解析 JWT Token
// 1. 解析 Token 字符串
// 2. 验证签名是否正确（使用 secretKey）
// 3. 验证 Token 是否过期、是否生效
// 返回: 解析出的 Claims 或错误
func validateJWTToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证 Token 并提取 Claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// JWTAuthMiddleWare JWT 认证中间件
// 工作流程:
// 1. 从 HTTP Header 的 Authorization 字段获取 Token
// 2. 验证 Token 格式（Bearer <token>）
// 3. 验证 Token 签名和有效性
// 4. 将解析出的用户信息存入 gin.Context，供后续 Handler 使用
// 5. 如果验证失败，返回 401/406 错误并终止请求
//
// 关键点: 这是 JWT 方案的核心，所有需要认证的路由都应用此中间件
func JWTAuthMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 1. 获取 Authorization Header
		authHeader := ctx.Request.Header.Get("Authorization")
		if authHeader == "" {
			// 没有提供 Token
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Need jwt token"})
			ctx.Abort()
			return
		}

		// 2. 提取 Token（去掉 "Bearer " 前缀）
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 3. 验证 Token
		claims, err := validateJWTToken(tokenString)
		if err != nil {
			// Token 无效或过期
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}

		// 4. 将用户信息存入 Context，供后续 Handler 使用
		ctx.Set("userID", claims.UserID)
		ctx.Set("username", claims.UserName)

		// 5. 继续执行后续的 Handler
		ctx.Next()
	}
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

// loginHandler 处理登录请求
// 1. 解析登录请求参数
// 2. 生成唯一的 userID (UUID)
// 3. 生成 JWT Token（包含 userID 和 username）
// 4. 返回 Token 给客户端
//
// 与 Session 方案的区别:
// - 不需要存储任何数据（无状态）
// - Token 中包含所有必要信息
// - 客户端后续请求携带 Token 即可
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成唯一的用户 ID
	userID := uuid.New().String()

	// 生成 JWT Token
	token, err := genJWTToken(userID, req.UserName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回 Token（客户端需要保存并在后续请求中携带）
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// profileHandler 获取用户信息
// 1. 从 gin.Context 中获取中间件解析的用户信息
// 2. 返回用户信息
//
// 关键点:
// - 此 Handler 受 JWTAuthMiddleWare 保护
// - userID 和 username 由中间件从 Token 中解析并存入 Context
// - 理论上不会获取失败（中间件已验证），但做了防御性检查
func profileHandler(c *gin.Context) {
	// 1. 从 Context 获取 userID
	userID, exists := c.Get("userID")
	if !exists {
		// 理论上不会发生（中间件已设置）
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get userID failed."})
		return
	}

	// 2. 从 Context 获取 username
	userName, exists := c.Get("username")
	if !exists {
		// 理论上不会发生（中间件已设置）
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get username failed."})
		return
	}

	// 3. 返回用户信息
	c.JSON(http.StatusOK, gin.H{
		"data":     fmt.Sprintf("get user: %s jwt token success, userID: %s", userName, userID),
		"userID":   userID,
		"username": userName,
		"serverID": serverID, // 显示处理此请求的服务器（演示无状态特性）
	})
}

func main() {
	r := gin.Default()

	// 注册路由
	r.POST("/login", loginHandler)                          // 登录（无需认证）
	r.GET("/profile", JWTAuthMiddleWare(), profileHandler)  // 获取用户信息（需要 JWT 认证）

	// 启动服务器
	fmt.Printf("🚀 JWT Token Server %s starting on port %s...\n", serverID, port)
	fmt.Printf("📝 JWT Secret: %s (仅用于演示，生产环境应使用环境变量)\n", secretKey)
	fmt.Printf("⏰ Token 过期时间: %v\n", TokenExpiren)
	r.Run(":" + port)
}

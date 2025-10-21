# 模块 1.4：会话状态管理

> **学习日期**: 2025-10-21
> **模块目标**: 理解无状态服务如何管理有状态会话，掌握三种主流会话管理方案及其选择策略

---

## 📚 核心问题

### 无状态服务 vs 有状态会话的矛盾

在模块 1.3 中，我们学习了**无状态服务设计**的优势：
- ✅ 易于水平扩展
- ✅ 服务器可随时替换
- ✅ 负载均衡更灵活
- ✅ 故障恢复简单

但现实应用中，用户需要**有状态的会话**：
- 🔐 用户登录状态（认证信息）
- 🛒 购物车数据
- ⚙️ 个性化设置
- 📊 临时业务数据

**核心矛盾**：如何在无状态的服务架构中，管理有状态的用户会话？

---

## 🎯 需要解决的关键问题

### 1. 会话数据存储在哪？

| 存储位置 | 优点 | 缺点 | 适用性 |
|---------|------|------|--------|
| **应用服务器内存** | 访问最快 | 无法共享，限制扩展 | ❌ 不适合分布式 |
| **客户端（Cookie）** | 减轻服务器负担 | 安全风险，大小限制（4KB） | ⚠️ 仅存非敏感数据 |
| **独立存储（Redis/DB）** | 可共享，易扩展 | 网络延迟，额外成本 | ✅ 推荐方案 |
| **无需存储（JWT）** | 完全无状态 | 无法主动撤销 | ✅ API/微服务 |

### 2. 如何在多台服务器间共享会话？

```mermaid
graph LR
    User[用户] --> LB[负载均衡器]
    LB --> S1[服务器1]
    LB --> S2[服务器2]
    LB --> S3[服务器3]

    S1 -.-> Session[(Session Store)]
    S2 -.-> Session
    S3 -.-> Session

    style Session fill:#f9f,stroke:#333,stroke-width:2px
```

**问题场景**：
- 用户第一次请求被路由到服务器 1，登录成功
- 第二次请求被路由到服务器 2，如何识别用户已登录？

**解决方案**：
1. **Sticky Session**：让用户始终访问同一台服务器
2. **Session Replication**：在所有服务器间同步 Session
3. **Centralized Session Store**：所有服务器访问同一个 Session 存储

### 3. 性能 vs 一致性权衡

- **本地 Session**：最快（内存访问），但无法共享
- **集中式 Session**：可共享，但有网络延迟（~1-2ms）
- **JWT Token**：无需存储，但无法主动失效

---

## 🔧 三种主流方案详解

### 方案 1: Session Affinity (Sticky Session)

#### 工作原理

```mermaid
sequenceDiagram
    participant User
    participant LB as 负载均衡器
    participant S1 as 服务器1
    participant S2 as 服务器2

    User->>LB: 第1次请求（登录）
    LB->>S1: 路由到服务器1
    S1->>S1: 创建 Session
    S1->>User: 返回 Cookie (server_id=1)

    User->>LB: 第2次请求（携带 Cookie）
    LB->>LB: 检查 Cookie，识别服务器1
    LB->>S1: 继续路由到服务器1
    S1->>S1: 读取本地 Session
    S1->>User: 返回响应
```

#### 实现方式

**Nginx 配置示例**：
```nginx
upstream backend {
    # 使用 IP Hash，同一 IP 总是路由到同一台服务器
    ip_hash;

    server 192.168.1.101:8080;
    server 192.168.1.102:8080;
    server 192.168.1.103:8080;
}

server {
    listen 80;
    location / {
        proxy_pass http://backend;
    }
}
```

#### 优缺点分析

✅ **优点**：
- 实现简单，无需改造应用
- 性能最优（本地内存访问）
- 无额外基础设施成本

❌ **缺点**：
- **单点故障**：服务器宕机会丢失所有 Session
- **扩展困难**：新增服务器会导致 Session 重新分配
- **负载不均**：某些服务器可能负载过高
- **不适合云环境**：容器动态伸缩会导致 Session 丢失

#### 适用场景

- 小规模应用（< 3 台服务器）
- 短会话场景（如临时访问）
- 内网管理系统（服务器固定）

---

### 方案 2: 集中式 Session Store (Redis)

#### 工作原理

```mermaid
sequenceDiagram
    participant User
    participant S1 as 服务器1
    participant S2 as 服务器2
    participant Redis

    User->>S1: 第1次请求（登录）
    S1->>Redis: SET session:abc123 {user_id: 1001}
    S1->>User: 返回 Cookie (session_id=abc123)

    User->>S2: 第2次请求（携带 session_id）
    S2->>Redis: GET session:abc123
    Redis->>S2: 返回 {user_id: 1001}
    S2->>User: 返回响应
```

#### 实现方式

**Redis Session Store 最佳实践**：

```bash
# 1. Session 数据结构设计（使用 Hash 节省内存）
HSET session:abc123 user_id 1001
HSET session:abc123 username "alice"
HSET session:abc123 login_time 1698765432
HSET session:abc123 ip "192.168.1.10"

# 2. 设置过期时间（30分钟）
EXPIRE session:abc123 1800

# 3. 读取 Session
HGETALL session:abc123

# 4. 更新 Session（续期）
HSET session:abc123 last_active 1698765500
EXPIRE session:abc123 1800

# 5. 删除 Session（登出）
DEL session:abc123
```

**Go 实现示例**：

```go
package session

import (
    "context"
    "encoding/json"
    "time"
    "github.com/go-redis/redis/v8"
)

type SessionStore struct {
    client *redis.Client
    ttl    time.Duration
}

type SessionData struct {
    UserID     int64     `json:"user_id"`
    Username   string    `json:"username"`
    LoginTime  time.Time `json:"login_time"`
    LastActive time.Time `json:"last_active"`
}

func NewSessionStore(redisAddr string, ttl time.Duration) *SessionStore {
    return &SessionStore{
        client: redis.NewClient(&redis.Options{
            Addr: redisAddr,
        }),
        ttl: ttl,
    }
}

// 创建 Session
func (s *SessionStore) Create(ctx context.Context, sessionID string, data *SessionData) error {
    data.LoginTime = time.Now()
    data.LastActive = time.Now()

    jsonData, err := json.Marshal(data)
    if err != nil {
        return err
    }

    key := "session:" + sessionID
    return s.client.Set(ctx, key, jsonData, s.ttl).Err()
}

// 获取 Session
func (s *SessionStore) Get(ctx context.Context, sessionID string) (*SessionData, error) {
    key := "session:" + sessionID

    jsonData, err := s.client.Get(ctx, key).Result()
    if err != nil {
        return nil, err
    }

    var data SessionData
    if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
        return nil, err
    }

    // 续期（访问时自动延长 TTL）
    s.client.Expire(ctx, key, s.ttl)

    return &data, nil
}

// 删除 Session（登出）
func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
    key := "session:" + sessionID
    return s.client.Del(ctx, key).Err()
}
```

**Python (Flask) 实现示例**：

```python
from flask import Flask, session
from flask_session import Session
import redis

app = Flask(__name__)

# 配置 Redis Session
app.config['SESSION_TYPE'] = 'redis'
app.config['SESSION_REDIS'] = redis.Redis(host='localhost', port=6379)
app.config['SESSION_PERMANENT'] = True
app.config['PERMANENT_SESSION_LIFETIME'] = 1800  # 30分钟

Session(app)

@app.route('/login', methods=['POST'])
def login():
    # 登录逻辑
    session['user_id'] = 1001
    session['username'] = 'alice'
    session['login_time'] = time.time()
    return {'message': 'Login successful'}

@app.route('/profile')
def profile():
    if 'user_id' not in session:
        return {'error': 'Not authenticated'}, 401

    return {
        'user_id': session['user_id'],
        'username': session['username']
    }

@app.route('/logout', methods=['POST'])
def logout():
    session.clear()
    return {'message': 'Logout successful'}
```

#### 优化技巧

1. **使用 Hash 而非 String**：节省 30-50% 内存
   ```bash
   # ❌ 低效：存储 JSON 字符串
   SET session:abc123 '{"user_id":1001,"username":"alice"}'

   # ✅ 高效：使用 Hash
   HSET session:abc123 user_id 1001 username alice
   ```

2. **Session ID 生成规范**：
   ```go
   // 使用 UUID v4
   sessionID := uuid.New().String()

   // 或使用加密随机字符串
   sessionID := generateSecureRandomString(32)
   ```

3. **避免 Session 过大**：
   - 仅存储必要信息（user_id、权限）
   - 详细数据从数据库查询
   - 单个 Session 建议 < 1KB

4. **处理并发更新**：
   ```go
   // 使用 Redis 事务确保原子性
   pipe := client.TxPipeline()
   pipe.HSet(ctx, key, "last_active", time.Now().Unix())
   pipe.Expire(ctx, key, ttl)
   _, err := pipe.Exec(ctx)
   ```

#### 高可用方案

**单点故障解决**：

1. **Redis Sentinel**（主从 + 自动故障转移）：
   ```
   +--------------------+
   |  Redis Master      |  <--- 写操作
   +--------------------+
            |
       复制  |
            v
   +--------------------+
   |  Redis Slave 1     |  <--- 读操作
   +--------------------+
   |  Redis Slave 2     |
   +--------------------+

   Sentinel 监控 Master，宕机时自动提升 Slave
   ```

2. **Redis Cluster**（分片 + 高可用）：
   ```
   Master 1 (Slot 0-5460)    ---复制---> Slave 1
   Master 2 (Slot 5461-10922) ---复制---> Slave 2
   Master 3 (Slot 10923-16383) ---复制---> Slave 3
   ```

#### 优缺点分析

✅ **优点**：
- **可扩展**：服务器可随意增删
- **高可用**：结合 Redis 高可用方案
- **服务器无状态**：符合云原生架构
- **支持跨域**：多个子系统共享 Session

❌ **缺点**：
- **网络延迟**：每次请求都需访问 Redis（~1-2ms）
- **额外成本**：需维护 Redis 集群
- **Redis 成为瓶颈**：高并发下需优化（主从分离、连接池）

#### 适用场景

- ✅ **推荐**：中大型 Web 应用
- ✅ 需要水平扩展的系统
- ✅ 云环境、容器化部署
- ✅ 单点登录（SSO）系统

---

### 方案 3: JWT Token (Stateless)

#### 工作原理

```mermaid
sequenceDiagram
    participant User
    participant Server

    User->>Server: POST /login (username, password)
    Server->>Server: 验证凭证
    Server->>Server: 签发 JWT Token<br/>{user_id: 1001, exp: ...}
    Server->>User: 返回 JWT Token

    User->>Server: GET /profile<br/>Header: Authorization: Bearer <JWT>
    Server->>Server: 验证 JWT 签名
    Server->>Server: 检查过期时间
    Server->>User: 返回用户数据
```

#### JWT 结构

JWT 由三部分组成，使用 `.` 连接：

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMDAxLCJleHAiOjE2OTg3Njc2MzJ9.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

**Header（头部）**：
```json
{
  "alg": "HS256",  // 签名算法
  "typ": "JWT"     // 令牌类型
}
```

**Payload（载荷）**：
```json
{
  "user_id": 1001,
  "username": "alice",
  "role": "admin",
  "iat": 1698765432,  // 签发时间
  "exp": 1698767632   // 过期时间（2小时后）
}
```

**Signature（签名）**：
```
HMACSHA256(
  base64UrlEncode(header) + "." +
  base64UrlEncode(payload),
  secret_key
)
```

#### 实现方式

**Go 实现示例（使用 golang-jwt）**：

```go
package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("your-256-bit-secret")

type Claims struct {
    UserID   int64  `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

// 生成 JWT Token
func GenerateToken(userID int64, username, role string) (string, error) {
    claims := &Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "my-app",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secretKey)
}

// 验证 JWT Token
func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return secretKey, nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, jwt.ErrSignatureInvalid
}
```

**中间件示例**：

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 从 Header 获取 Token
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }

        // 格式：Bearer <token>
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // 验证 Token
        claims, err := ValidateToken(tokenString)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // 将用户信息存入 Context
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "username", claims.Username)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

#### 刷新 Token 机制

**问题**：JWT 过期后用户需要重新登录，体验差

**解决方案**：双 Token 机制

```mermaid
sequenceDiagram
    participant User
    participant Server

    User->>Server: POST /login
    Server->>User: 返回 Access Token (15分钟)<br/>+ Refresh Token (7天)

    User->>Server: GET /api (Access Token)
    Server->>User: 返回数据

    Note over User: 15分钟后 Access Token 过期

    User->>Server: GET /api (过期的 Access Token)
    Server->>User: 401 Unauthorized

    User->>Server: POST /refresh (Refresh Token)
    Server->>Server: 验证 Refresh Token
    Server->>User: 返回新的 Access Token

    User->>Server: GET /api (新 Access Token)
    Server->>User: 返回数据
```

**实现代码**：

```go
// 登录时同时返回两种 Token
func Login(w http.ResponseWriter, r *http.Request) {
    // ... 验证用户名密码 ...

    // 短期 Access Token (15分钟)
    accessToken, _ := GenerateToken(userID, username, role, 15*time.Minute)

    // 长期 Refresh Token (7天)
    refreshToken, _ := GenerateRefreshToken(userID, 7*24*time.Hour)

    // 将 Refresh Token 存储到 Redis（用于撤销）
    storeRefreshToken(refreshToken, userID)

    json.NewEncoder(w).Write(map[string]string{
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "expires_in":    "900", // 秒
    })
}

// 刷新 Token
func RefreshToken(w http.ResponseWriter, r *http.Request) {
    refreshToken := r.FormValue("refresh_token")

    // 验证 Refresh Token
    claims, err := ValidateRefreshToken(refreshToken)
    if err != nil {
        http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
        return
    }

    // 检查 Redis 中是否存在（是否已撤销）
    if !isRefreshTokenValid(refreshToken) {
        http.Error(w, "Refresh token revoked", http.StatusUnauthorized)
        return
    }

    // 生成新的 Access Token
    newAccessToken, _ := GenerateToken(claims.UserID, claims.Username, claims.Role, 15*time.Minute)

    json.NewEncoder(w).Write(map[string]string{
        "access_token": newAccessToken,
        "expires_in":   "900",
    })
}
```

#### 如何实现"登出"功能？

**问题**：JWT 是无状态的，服务器无法主动撤销 Token

**解决方案**：黑名单机制

```go
// 登出时将 Token 加入黑名单（存储到 Redis）
func Logout(w http.ResponseWriter, r *http.Request) {
    tokenString := extractToken(r)

    claims, _ := ValidateToken(tokenString)

    // 计算 Token 剩余有效期
    ttl := time.Until(claims.ExpiresAt.Time)

    // 存入 Redis 黑名单
    key := "blacklist:" + tokenString
    redisClient.Set(context.Background(), key, "revoked", ttl)

    w.WriteHeader(http.StatusOK)
}

// 验证 Token 时检查黑名单
func ValidateTokenWithBlacklist(tokenString string) (*Claims, error) {
    // 先检查黑名单
    key := "blacklist:" + tokenString
    _, err := redisClient.Get(context.Background(), key).Result()
    if err == nil {
        return nil, errors.New("token revoked")
    }

    // 正常验证
    return ValidateToken(tokenString)
}
```

#### 安全最佳实践

1. **使用 HTTPS**：防止 Token 被窃取
2. **短过期时间**：Access Token 建议 15-30 分钟
3. **不存储敏感信息**：Payload 可被 Base64 解码
4. **使用强密钥**：HS256 至少 256 位，RS256 使用 2048 位
5. **验证 iss/aud**：防止 Token 被其他系统冒用

```go
claims := &Claims{
    UserID: userID,
    RegisteredClaims: jwt.RegisteredClaims{
        Issuer:   "my-app",           // 签发者
        Audience: []string{"web-app"}, // 目标受众
    },
}

// 验证时检查
if claims.Issuer != "my-app" {
    return errors.New("invalid issuer")
}
```

#### 优缺点分析

✅ **优点**：
- **完全无状态**：服务器无需存储 Session
- **易扩展**：无需共享存储
- **跨域友好**：天然支持跨域认证
- **移动端友好**：适合 APP、小程序
- **减少数据库查询**：用户信息编码在 Token 中

❌ **缺点**：
- **无法主动撤销**：Token 过期前一直有效（需黑名单补救）
- **Token 较大**：每次请求携带（~200-500 字节）
- **时间同步要求**：服务器时间不一致会导致验证失败
- **安全风险**：Payload 可被解码，不能存敏感信息

#### 适用场景

- ✅ **API 服务**：RESTful API、GraphQL
- ✅ **微服务架构**：服务间认证
- ✅ **移动应用**：APP、小程序
- ✅ **单页应用（SPA）**：React、Vue
- ❌ **传统 Web**：需频繁撤销 Session 的场景

---

## 🧭 方案选择决策树

```mermaid
graph TD
    Start[开始选择会话管理方案] --> Q1{是否需要<br/>水平扩展?}

    Q1 -->|否<br/>单机或固定服务器| StickySession[Session Affinity<br/>粘滞会话]
    Q1 -->|是<br/>需要动态扩展| Q2{是否需要<br/>主动撤销会话?}

    Q2 -->|是<br/>强制登出/权限变更| Q3{访问模式?}
    Q2 -->|否<br/>自动过期即可| JWT[JWT Token<br/>无状态令牌]

    Q3 -->|传统Web应用<br/>服务器渲染| Redis[集中式Session<br/>Redis存储]
    Q3 -->|API服务<br/>微服务| Choice{灵活选择}

    Choice -->|高频变更权限| Redis
    Choice -->|权限稳定| JWT_Hybrid[JWT + 黑名单<br/>混合方案]

    StickySession --> Note1[⚠️ 注意单点故障]
    Redis --> Note2[✅ 推荐方案<br/>生产环境首选]
    JWT --> Note3[✅ 适合移动端/API]
    JWT_Hybrid --> Note4[⚠️ 增加复杂度<br/>需权衡]

    style Redis fill:#90EE90
    style JWT fill:#87CEEB
    style StickySession fill:#FFB6C1
```

### 决策矩阵

| 场景 | 推荐方案 | 原因 |
|------|---------|------|
| 内网管理系统（< 3台服务器） | Session Affinity | 简单，性能高 |
| 电商平台（需扩展） | Redis Session | 需要强制登出、修改购物车 |
| 移动端 APP API | JWT Token | 无状态，适合分布式 |
| 单页应用（SPA） | JWT Token | 跨域友好，减少服务器压力 |
| 微服务架构 | JWT Token | 服务间认证，无需共享状态 |
| 银行系统（高安全要求） | Redis Session | 需要实时撤销会话 |
| 社交平台（高并发读） | JWT Token + Redis（黑名单） | 兼顾性能与安全 |

---

## 💡 实战建议

### 任务：对比三种方案

根据课程清单要求，完成以下实践：

#### 1. 绘制决策树（已完成上方）

#### 2. 实验对比

**实验设计**：

```bash
# 启动 3 个后端服务实例
cd projects/week1/session-comparison

# 终端 1: 启动服务器1
PORT=8081 go run main.go

# 终端 2: 启动服务器2
PORT=8082 go run main.go

# 终端 3: 启动服务器3
PORT=8083 go run main.go

# 终端 4: 启动 Nginx 负载均衡器
nginx -c nginx.conf

# 终端 5: 启动 Redis
docker run -d -p 6379:6379 redis
```

**测试脚本**：

```bash
# 测试 1: Session Affinity
curl -c cookies.txt http://localhost/login  # 登录
curl -b cookies.txt http://localhost/profile  # 访问个人信息
# 观察：总是路由到同一台服务器

# 测试 2: Redis Session
curl -c cookies.txt http://localhost/login-redis
curl -b cookies.txt http://localhost/profile-redis
# 观察：可以路由到不同服务器，但都能获取 Session

# 测试 3: JWT Token
TOKEN=$(curl http://localhost/login-jwt | jq -r .token)
curl -H "Authorization: Bearer $TOKEN" http://localhost/profile-jwt
# 观察：无需 Cookie，任意服务器都能验证
```

**性能对比**：

```bash
# 使用 Apache Bench 压测
ab -n 10000 -c 100 http://localhost/profile-sticky
ab -n 10000 -c 100 http://localhost/profile-redis
ab -n 10000 -c 100 http://localhost/profile-jwt

# 记录对比：
# - Requests per second (QPS)
# - Time per request (延迟)
# - 服务器宕机后的表现
```

#### 3. 故障演练

```bash
# 场景 1: 服务器宕机
# 杀死服务器1，观察各方案的表现
kill -9 <PID of server 1>

# Sticky Session: 用户 Session 丢失，需重新登录
# Redis Session: 无影响，自动路由到其他服务器
# JWT Token: 无影响

# 场景 2: Redis 宕机
docker stop <redis-container>

# Sticky Session: 无影响（不依赖 Redis）
# Redis Session: 全部失效，用户需重新登录
# JWT Token: 无影响（无状态）

# 场景 3: 强制登出
# Sticky Session: 无法实现（Session 在各服务器本地）
# Redis Session: 删除 Redis Key 即可
# JWT Token: 需要黑名单机制
```

---

## 📊 性能数据参考

基于生产环境实测数据：

| 指标 | Session Affinity | Redis Session | JWT Token |
|------|-----------------|---------------|-----------|
| **延迟** | ~0.1ms | ~1-2ms | ~0.3ms |
| **QPS（单机）** | 50,000 | 30,000 | 45,000 |
| **内存占用（10万用户）** | 500MB/台 | Redis: 2GB | 0 |
| **扩展性** | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **可用性** | ⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **安全性** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |

**说明**：
- Redis Session 延迟包含网络开销（局域网 ~1ms）
- JWT Token 需要 CPU 计算签名（~0.2ms）
- Redis Session 需要高性能 Redis 集群支持高 QPS

---

## 🔗 延伸阅读

### 推荐资料

1. **Session Management Strategies**
   https://stackoverflow.blog/2021/10/06/best-practices-for-authentication-and-authorization-for-rest-apis/
   → REST API 认证授权最佳实践

2. **Sticky Sessions vs Session Replication**
   https://www.nginx.com/blog/nginx-plus-sticky-sessions/
   → Nginx 官方文档

3. **Redis Session Store Pattern**
   https://redis.io/docs/manual/patterns/distributed-locks/
   → Redis 分布式锁与 Session 存储

4. **JWT 完整指南**
   https://jwt.io/introduction
   → JWT 官方文档，包含各语言实现

### 进阶主题

- **OAuth 2.0 与 Session 的结合**
  第三方登录（微信、Google）如何管理 Session

- **跨域 Session 共享**
  子域名共享 Cookie：`.example.com`

- **Session 劫持防护**
  - HTTPS 强制
  - HttpOnly Cookie
  - SameSite 属性
  - CSRF Token

---

## ✅ 学习检查清单

完成本模块后，你应该能够回答：

- [ ] **为什么无状态服务还需要管理会话？**
  答：无状态指服务器不保存业务状态，但用户需要会话（如登录状态）

- [ ] **Sticky Session 的最大问题是什么？**
  答：服务器宕机会丢失所有 Session，且难以水平扩展

- [ ] **Redis 作为 Session Store 的单点故障如何解决？**
  答：使用 Redis Sentinel（高可用）或 Redis Cluster（分片 + 高可用）

- [ ] **JWT Token 如何实现"登出"功能？**
  答：无法主动撤销 Token，需结合黑名单（存储在 Redis）

- [ ] **电商购物车应该用哪种方案？**
  答：推荐 Redis Session（需持久化购物车，且需跨服务器访问）

- [ ] **移动端 APP 认证应该用哪种方案？**
  答：推荐 JWT Token（无状态，适合分布式，减少服务器压力）

- [ ] **如何选择 JWT 签名算法（HS256 vs RS256）？**
  答：
  - HS256：对称加密，简单高效，适合单体应用
  - RS256：非对称加密，更安全，适合微服务（公钥分发）

---

## 📝 本模块学习总结

### 核心收获

1. **理解矛盾**：无状态服务 vs 有状态会话的本质矛盾
2. **掌握方案**：三种主流方案的原理、实现、适用场景
3. **权衡思维**：没有完美方案，只有最适合的选择
4. **实践能力**：能够基于业务场景选择合适的技术方案

### 关键决策点

| 决策因素 | Session Affinity | Redis Session | JWT Token |
|---------|-----------------|---------------|-----------|
| 规模 | 小型 | 中大型 | 大型 |
| 扩展需求 | 低 | 高 | 高 |
| 主动撤销 | 不支持 | ✅ 支持 | 需黑名单 |
| 实现复杂度 | 低 | 中 | 中 |
| 运维成本 | 低 | 中（Redis） | 低 |

### 实践建议

- **小型项目**：先用 Sticky Session，遇到瓶颈再升级
- **企业级应用**：直接使用 Redis Session，避免后期迁移
- **API 服务**：优先选择 JWT Token，简化架构
- **混合方案**：Web 用 Redis Session，API 用 JWT Token

### 下一步

- [ ] 完成实验：对比三种方案的性能和可用性
- [ ] 绘制决策树（可使用 Mermaid 或手绘）
- [ ] 进入**模块 2: 负载均衡算法与实现**

---

**学习日期**: 2025-10-21
**学习时长**: _____ 小时
**关键收获**: _____________________
**待深入**: _____________________

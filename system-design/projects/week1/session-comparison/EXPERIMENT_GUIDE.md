# 会话管理三种方案对比实验指南

> **实验目标**: 通过动手实现三种会话管理方案，深入理解它们的工作原理、性能差异和适用场景

---

## 📋 实验概览

### 实验架构

```
┌─────────────────────────────────────────────────────────┐
│                    客户端测试脚本                          │
│            (Python: requests + 性能分析)                  │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                   负载均衡层                              │
│         (Nginx 或 自实现的简单 Load Balancer)              │
└─────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
   ┌─────────┐        ┌─────────┐        ┌─────────┐
   │ Server1 │        │ Server2 │        │ Server3 │
   │ :8081   │        │ :8082   │        │ :8083   │
   └─────────┘        └─────────┘        └─────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            ▼
                    ┌───────────────┐
                    │  Redis :6379  │
                    │ (方案2使用)    │
                    └───────────────┘
```

### 三种方案实现

| 方案 | 后端语言 | 端口 | 特点 |
|------|---------|------|------|
| **方案1: Sticky Session** | Go | 8081-8083 | Nginx IP Hash 路由 |
| **方案2: Redis Session** | Go + Python 示例 | 8091-8093 | 集中式存储 |
| **方案3: JWT Token** | Go | 8101-8103 | 无状态认证 |

---

## 🎯 实验步骤详解

### 阶段一：准备工作

#### Step 1.1: 创建项目结构

```bash
cd projects/week1/session-comparison

# 创建目录
mkdir -p sticky-session redis-session jwt-token test-scripts docker

# 目录结构
session-comparison/
├── EXPERIMENT_GUIDE.md          # 本文件
├── README.md                     # 项目说明
├── docker/                       # Docker 配置
│   ├── docker-compose.yml
│   └── nginx.conf
├── sticky-session/               # 方案1: 粘滞会话
│   ├── main.go
│   ├── go.mod
│   └── README.md
├── redis-session/                # 方案2: Redis Session
│   ├── go-server/
│   │   ├── main.go
│   │   └── go.mod
│   └── python-server/
│       ├── app.py
│       └── requirements.txt
├── jwt-token/                    # 方案3: JWT Token
│   ├── main.go
│   ├── go.mod
│   └── README.md
└── test-scripts/                 # 测试脚本
    ├── test_sticky.py
    ├── test_redis.py
    ├── test_jwt.py
    ├── performance_compare.py
    └── fault_injection.py
```

#### Step 1.2: 安装依赖

```bash
# 安装 Go (如果未安装)
# macOS: brew install go
# 验证: go version

# 安装 Python 依赖
pip install requests redis pyjwt flask flask-session

# 安装 Docker (用于运行 Redis 和 Nginx)
# macOS: brew install docker
# 启动 Docker Desktop

# 启动 Redis
docker run -d --name redis -p 6379:6379 redis:alpine

# mac安装redis-cli
brew install redis

# 连接redis 
brew services start redis
redis-cli -h 127.0.0.1 -p 6379

# 验证 Redis
redis-cli ping  # 应返回 PONG
```

---

### 阶段二：实现方案1 - Sticky Session

#### Step 2.1: Go 服务器实现要点

**核心功能**：
1. ✅ 在本地内存中存储 Session (使用 `sync.Map`)
2. ✅ 登录时生成 Session ID，存储在 Cookie 中
3. ✅ 每个服务器实例有唯一标识（Server ID）
4. ✅ 提供 API：`/login`, `/profile`, `/logout`

**关键代码提示**：

```go
// 全局 Session 存储（每个服务器独立）
var sessionStore sync.Map

// Session 数据结构
type Session struct {
    UserID    int64
    Username  string
    LoginTime time.Time
    ServerID  string  // 标识是哪个服务器创建的
}

// 生成 Session ID
sessionID := uuid.New().String()

// 存储 Session
sessionStore.Store(sessionID, session)

// 设置 Cookie
http.SetCookie(w, &http.Cookie{
    Name:     "session_id",
    Value:    sessionID,
    Path:     "/",
    MaxAge:   3600, // 1小时
    HttpOnly: true,
})
```

**重点观察**：
- 每个服务器的 Session 是独立的
- 服务器重启后 Session 丢失
- 记录日志：哪个服务器处理了哪个请求

#### Step 2.2: Nginx 配置要点

**配置文件位置**: `docker/nginx-sticky.conf`

```nginx
upstream backend_sticky {
    # IP Hash: 同一 IP 始终路由到同一台服务器
    ip_hash;

    server host.docker.internal:8081;
    server host.docker.internal:8082;
    server host.docker.internal:8083;
}

server {
    listen 80;
    server_name localhost;

    location / {
        proxy_pass http://backend_sticky;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;

        # 传递客户端 IP 用于 ip_hash
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

**启动命令**：

```bash
# 启动 3 个 Go 服务器
PORT=8081 SERVER_ID=server-1 go run sticky-session/main.go &
PORT=8082 SERVER_ID=server-2 go run sticky-session/main.go &
PORT=8083 SERVER_ID=server-3 go run sticky-session/main.go &

# 启动 Nginx (Docker 方式)
docker run -d --name nginx-sticky \
  -p 8080:80 \
  -v $(pwd)/docker/nginx-sticky.conf:/etc/nginx/nginx.conf:ro \
  nginx:alpine
```

#### Step 2.3: 测试脚本要点

**文件**: `test-scripts/test_sticky.py`

**测试场景**：

1. **基本功能测试**：
   ```python
   import requests

   session = requests.Session()  # 自动管理 Cookie

   # 登录
   resp = session.post('http://localhost:8080/login',
                       json={'username': 'alice', 'password': '123456'})
   print(f"Login: {resp.json()}")

   # 多次访问，观察是否总是同一台服务器
   for i in range(10):
       resp = session.get('http://localhost:8080/profile')
       print(f"Request {i+1}: Server={resp.json()['server_id']}")
   ```

2. **多客户端测试**：
   ```python
   # 模拟 5 个不同用户（不同 Session）
   for user_id in range(5):
       session = requests.Session()
       session.post('http://localhost:8080/login',
                   json={'username': f'user{user_id}'})
       # 观察是否被分配到不同服务器
   ```

3. **故障注入测试**：
   ```python
   # 测试步骤：
   # 1. 登录成功
   # 2. 手动杀死处理该用户的服务器（kill -9 PID）
   # 3. 再次请求，观察是否 Session 丢失
   ```

**期望结果**：
- ✅ 同一客户端的请求总是路由到同一台服务器
- ❌ 服务器宕机后，Session 丢失，需要重新登录

---

### 阶段三：实现方案2 - Redis Session

#### Step 3.1: Go 服务器实现要点

**核心功能**：
1. ✅ 连接 Redis 客户端 (使用 `go-redis/redis`)
2. ✅ Session 存储在 Redis，Key 格式: `session:{session_id}`
3. ✅ 所有服务器共享同一个 Redis

**关键代码提示**：

```go
import "github.com/go-redis/redis/v8"

// 初始化 Redis 客户端
var rdb = redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
    DB:   0,
})

// 存储 Session 到 Redis
func saveSession(sessionID string, session *Session) error {
    data, _ := json.Marshal(session)
    key := "session:" + sessionID
    return rdb.Set(ctx, key, data, 30*time.Minute).Err()
}

// 从 Redis 读取 Session
func getSession(sessionID string) (*Session, error) {
    key := "session:" + sessionID
    data, err := rdb.Get(ctx, key).Result()
    if err != nil {
        return nil, err
    }

    var session Session
    json.Unmarshal([]byte(data), &session)

    // 续期：每次访问延长 TTL
    rdb.Expire(ctx, key, 30*time.Minute)

    return &session, nil
}

// 删除 Session (登出)
func deleteSession(sessionID string) error {
    key := "session:" + sessionID
    return rdb.Del(ctx, key).Err()
}
```

**依赖安装**：

```bash
cd redis-session/go-server
go mod init session-redis-demo
go get github.com/go-redis/redis/v8
go get github.com/google/uuid
```

#### Step 3.2: Python Flask 服务器实现要点 (可选)

**目的**：演示跨语言共享 Session

**核心代码提示**：

```python
from flask import Flask, request, jsonify, session
from flask_session import Session
import redis
import json
import uuid

app = Flask(__name__)

# 配置 Redis Session
app.config['SESSION_TYPE'] = 'redis'
app.config['SESSION_REDIS'] = redis.Redis(host='localhost', port=6379)
app.config['SESSION_PERMANENT'] = True
app.config['PERMANENT_SESSION_LIFETIME'] = 1800  # 30分钟

Session(app)

@app.route('/login', methods=['POST'])
def login():
    data = request.json
    session['user_id'] = 1001
    session['username'] = data['username']
    session['server_id'] = 'python-server'
    return jsonify({'message': 'Login successful'})

@app.route('/profile')
def profile():
    if 'user_id' not in session:
        return jsonify({'error': 'Not authenticated'}), 401

    return jsonify({
        'user_id': session['user_id'],
        'username': session['username'],
        'server_id': session['server_id']
    })

if __name__ == '__main__':
    app.run(port=8092)
```

#### Step 3.3: Nginx 配置要点

**配置文件**: `docker/nginx-redis.conf`

```nginx
upstream backend_redis {
    # 使用轮询，不需要 ip_hash
    server host.docker.internal:8091;
    server host.docker.internal:8092;
    server host.docker.internal:8093;
}

server {
    listen 81;
    location / {
        proxy_pass http://backend_redis;
    }
}
```

**关键区别**：
- ❌ 不使用 `ip_hash`
- ✅ 使用默认的轮询（Round Robin）
- ✅ 请求可以被路由到任意服务器

#### Step 3.4: 测试脚本要点

**文件**: `test-scripts/test_redis.py`

**测试场景**：

1. **跨服务器 Session 共享**：
   ```python
   session = requests.Session()

   # 登录
   resp = session.post('http://localhost:8081/login', json={'username': 'alice'})

   # 多次请求，观察是否路由到不同服务器，但都能获取 Session
   for i in range(10):
       resp = session.get('http://localhost:8081/profile')
       print(f"Request {i+1}: Server={resp.json()['server_id']}, User={resp.json()['username']}")
   ```

2. **Redis 中的数据查看**：
   ```bash
   redis-cli
   KEYS session:*          # 查看所有 Session Key
   GET session:abc123      # 查看具体 Session 内容
   TTL session:abc123      # 查看剩余过期时间
   ```

3. **服务器宕机测试**：
   ```python
   # 测试步骤：
   # 1. 登录，Session 存储到 Redis
   # 2. 杀死服务器1
   # 3. 请求被路由到服务器2，仍能获取 Session
   ```

**期望结果**：
- ✅ 请求可以路由到不同服务器，但都能访问 Session
- ✅ 服务器宕机不影响 Session（数据在 Redis 中）
- ✅ 跨语言服务器（Go + Python）可以共享 Session

---

### 阶段四：实现方案3 - JWT Token

#### Step 4.1: Go 服务器实现要点

**核心功能**：
1. ✅ 使用 `golang-jwt/jwt` 库生成和验证 JWT
2. ✅ 登录返回 Token，不存储 Session
3. ✅ 验证 Token 签名和过期时间
4. ✅ 实现刷新 Token 机制（可选）

**关键代码提示**：

```go
import "github.com/golang-jwt/jwt/v5"

var secretKey = []byte("your-256-bit-secret-key")

// Claims 结构
type Claims struct {
    UserID   int64  `json:"user_id"`
    Username string `json:"username"`
    jwt.RegisteredClaims
}

// 生成 JWT Token
func generateToken(userID int64, username string) (string, error) {
    claims := &Claims{
        UserID:   userID,
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "session-demo",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(secretKey)
}

// 验证 JWT Token
func validateToken(tokenString string) (*Claims, error) {
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

// 认证中间件
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 从 Header 获取 Token
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }

        // 格式: Bearer <token>
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // 验证 Token
        claims, err := validateToken(tokenString)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // 将用户信息存入 Context
        ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
        ctx = context.WithValue(ctx, "username", claims.Username)

        next.ServeHTTP(w, r.WithContext(ctx))
    }
}
```

**API 设计**：

```go
// POST /login
// Request: {"username": "alice", "password": "123456"}
// Response: {"token": "eyJhbGc...", "expires_in": 7200}

// GET /profile
// Header: Authorization: Bearer <token>
// Response: {"user_id": 1001, "username": "alice", "server_id": "server-1"}
```

#### Step 4.2: 黑名单实现（可选高级功能）

**目的**：实现"登出"功能

```go
var blacklist sync.Map  // 或使用 Redis

// 登出：将 Token 加入黑名单
func logout(tokenString string) error {
    claims, _ := validateToken(tokenString)

    // 计算 Token 剩余有效期
    ttl := time.Until(claims.ExpiresAt.Time)

    // 存入黑名单（Redis 版本）
    key := "blacklist:" + tokenString
    return rdb.Set(ctx, key, "revoked", ttl).Err()
}

// 验证时检查黑名单
func validateTokenWithBlacklist(tokenString string) (*Claims, error) {
    // 先检查黑名单
    key := "blacklist:" + tokenString
    _, err := rdb.Get(ctx, key).Result()
    if err == nil {
        return nil, errors.New("token revoked")
    }

    return validateToken(tokenString)
}
```

#### Step 4.3: 测试脚本要点

**文件**: `test-scripts/test_jwt.py`

**测试场景**：

1. **基本认证流程**：
   ```python
   import requests

   # 登录获取 Token
   resp = requests.post('http://localhost:8101/login',
                       json={'username': 'alice', 'password': '123456'})
   token = resp.json()['token']
   print(f"Token: {token[:50]}...")

   # 使用 Token 访问 API
   headers = {'Authorization': f'Bearer {token}'}
   resp = requests.get('http://localhost:8101/profile', headers=headers)
   print(f"Profile: {resp.json()}")
   ```

2. **Token 解码（不验证签名）**：
   ```python
   import jwt

   # 解码 Payload（不验证签名，只查看内容）
   payload = jwt.decode(token, options={"verify_signature": False})
   print(f"User ID: {payload['user_id']}")
   print(f"Expires At: {payload['exp']}")
   ```

3. **过期 Token 测试**：
   ```python
   import time

   # 使用短过期时间的 Token (5秒)
   resp = requests.post('http://localhost:8101/login-short')
   token = resp.json()['token']

   # 5秒后再请求
   time.sleep(6)
   resp = requests.get('http://localhost:8101/profile',
                      headers={'Authorization': f'Bearer {token}'})
   print(f"Status: {resp.status_code}")  # 应该是 401
   ```

4. **黑名单测试**：
   ```python
   # 登出
   requests.post('http://localhost:8101/logout',
                headers={'Authorization': f'Bearer {token}'})

   # 再次使用 Token（应该失败）
   resp = requests.get('http://localhost:8101/profile',
                      headers={'Authorization': f'Bearer {token}'})
   print(f"Status: {resp.status_code}")  # 应该是 401
   ```

**期望结果**：
- ✅ 服务器无需存储 Token，完全无状态
- ✅ 请求可以路由到任意服务器
- ✅ Token 过期后自动失效
- ✅ 黑名单可以实现主动登出

---

### 阶段五：性能对比测试

#### Step 5.1: 编写性能测试脚本

**文件**: `test-scripts/performance_compare.py`

**测试指标**：
1. **延迟（Latency）**：单次请求的响应时间
2. **吞吐量（Throughput）**：每秒处理的请求数（QPS）
3. **内存占用**：服务器的内存使用情况

**测试方法**：

```python
import requests
import time
import statistics
from concurrent.futures import ThreadPoolExecutor

def test_latency(url, headers=None, cookies=None):
    """测试单次请求延迟"""
    latencies = []

    for _ in range(100):
        start = time.time()
        requests.get(url, headers=headers, cookies=cookies)
        latency = (time.time() - start) * 1000  # 转换为毫秒
        latencies.append(latency)

    return {
        'p50': statistics.median(latencies),
        'p95': statistics.quantiles(latencies, n=20)[18],
        'p99': statistics.quantiles(latencies, n=100)[98],
        'avg': statistics.mean(latencies)
    }

def test_throughput(url, duration=10, concurrency=50):
    """测试吞吐量"""
    request_count = 0

    def make_request():
        nonlocal request_count
        requests.get(url)
        request_count += 1

    start = time.time()
    end_time = start + duration

    with ThreadPoolExecutor(max_workers=concurrency) as executor:
        while time.time() < end_time:
            executor.submit(make_request)

    elapsed = time.time() - start
    qps = request_count / elapsed

    return {
        'total_requests': request_count,
        'qps': qps,
        'duration': elapsed
    }

# 对比三种方案
schemes = {
    'Sticky Session': {'url': 'http://localhost:8080/profile', 'cookies': ...},
    'Redis Session': {'url': 'http://localhost:8081/profile', 'cookies': ...},
    'JWT Token': {'url': 'http://localhost:8101/profile', 'headers': ...}
}

for name, config in schemes.items():
    print(f"\n=== {name} ===")
    latency = test_latency(config['url'], ...)
    print(f"P50: {latency['p50']:.2f}ms")
    print(f"P99: {latency['p99']:.2f}ms")

    throughput = test_throughput(config['url'])
    print(f"QPS: {throughput['qps']:.0f}")
```

**使用 Apache Bench（更专业）**：

```bash
# 安装
brew install apache2  # macOS

# 测试 Sticky Session
ab -n 10000 -c 100 -C "session_id=abc123" http://localhost:8080/profile

# 测试 Redis Session
ab -n 10000 -c 100 -C "session_id=xyz789" http://localhost:8081/profile

# 测试 JWT Token
ab -n 10000 -c 100 -H "Authorization: Bearer <token>" http://localhost:8101/profile
```

**期望结果示例**：

| 方案 | P50 延迟 | P99 延迟 | QPS | 内存占用(10万用户) |
|------|---------|---------|-----|------------------|
| Sticky Session | ~0.1ms | ~0.5ms | 50,000 | 500MB/台 |
| Redis Session | ~1.5ms | ~3ms | 30,000 | Redis: 2GB |
| JWT Token | ~0.3ms | ~1ms | 45,000 | ~0 |

---

### 阶段六：故障注入测试

#### Step 6.1: 服务器宕机测试

**测试脚本**: `test-scripts/fault_injection.py`

**场景1: 服务器宕机**

```python
import requests
import subprocess
import time

# 1. 启动所有服务器，记录 PID
# 2. 登录，获取 Session/Token
session = requests.Session()
resp = session.post('http://localhost:8080/login', json={'username': 'alice'})

# 3. 查看当前请求到哪个服务器
resp = session.get('http://localhost:8080/profile')
server_id = resp.json()['server_id']
print(f"Current Server: {server_id}")

# 4. 杀死该服务器
# 手动操作: kill -9 <PID>
input("请手动杀死服务器，然后按回车继续...")

# 5. 再次请求，观察结果
try:
    resp = session.get('http://localhost:8080/profile')
    print(f"Result: {resp.json()}")
except Exception as e:
    print(f"Error: {e}")
```

**观察点**：
- **Sticky Session**: Session 丢失，401 错误
- **Redis Session**: 自动路由到其他服务器，正常返回
- **JWT Token**: 自动路由到其他服务器，正常返回

**场景2: Redis 宕机**

```bash
# 停止 Redis
docker stop redis

# 观察各方案表现
# - Sticky Session: 不受影响
# - Redis Session: 全部失效
# - JWT Token: 不受影响
```

#### Step 6.2: 网络延迟模拟

**使用 tc (Linux) 或 pfctl (macOS) 模拟网络延迟**

```bash
# macOS 模拟延迟 (需要 sudo)
sudo dnctl pipe 1 config delay 100  # 100ms 延迟
sudo ipfw add pipe 1 ip from any to 127.0.0.1 dst-port 6379

# 测试 Redis Session 的延迟增加
```

---

## 📊 实验记录表格

### 功能对比

| 功能 | Sticky Session | Redis Session | JWT Token |
|------|---------------|---------------|-----------|
| 跨服务器共享 | ❌ | ✅ | ✅ |
| 服务器宕机恢复 | ❌ | ✅ | ✅ |
| 主动登出 | ✅ | ✅ | ⚠️ 需黑名单 |
| 水平扩展 | ⚠️ 困难 | ✅ | ✅ |
| 依赖外部服务 | ❌ | ✅ Redis | ❌ |

### 性能数据（自己测试填写）

| 方案 | P50 延迟(ms) | P99 延迟(ms) | QPS | 内存占用 |
|------|-------------|-------------|-----|----------|
| Sticky Session | _____ | _____ | _____ | _____ |
| Redis Session | _____ | _____ | _____ | _____ |
| JWT Token | _____ | _____ | _____ | _____ |

---

## 🎓 实验学习要点

### 关键理解

1. **Sticky Session**：
   - ✅ 理解 Nginx `ip_hash` 的工作原理
   - ✅ 明白为什么服务器宕机会导致 Session 丢失
   - ✅ 体会"有状态服务"的扩展困难

2. **Redis Session**：
   - ✅ 理解"集中式存储"的概念
   - ✅ 观察 Redis 中 Session 的存储格式
   - ✅ 体验跨服务器共享的优势
   - ✅ 理解 Redis 成为单点故障的风险

3. **JWT Token**：
   - ✅ 理解"无状态"的真正含义
   - ✅ 掌握 JWT 的三部分结构（Header.Payload.Signature）
   - ✅ 理解为什么 JWT 无法主动撤销
   - ✅ 体验黑名单方案的折衷

### 常见问题 FAQ

**Q1: Sticky Session 下，如何保证负载均衡？**
- A: 使用 `ip_hash` 会导致负载不均，改进方案是使用 `consistent hash`

**Q2: Redis Session 的性能瓶颈在哪？**
- A: 网络延迟（~1-2ms），高并发下需要 Redis 主从分离 + 连接池

**Q3: JWT Token 太大怎么办？**
- A: 只存储必要字段（user_id），详细信息从数据库查询

**Q4: 如何选择适合的方案？**
- A: 参考笔记中的决策树，主要看：规模、扩展需求、主动登出需求

---

## 🚀 进阶实验（可选）

### 高级实验1: 混合方案

**场景**: Web 用 Redis Session，API 用 JWT Token

```
前端 Web 应用 ──(Cookie)──> Redis Session 服务器
移动端 APP   ──(Token)──> JWT 认证服务器
```

### 高级实验2: Redis Session 高可用

**搭建 Redis Sentinel**：

```bash
docker run -d --name redis-master redis:alpine
docker run -d --name redis-slave redis:alpine --slaveof redis-master 6379
docker run -d --name redis-sentinel redis:alpine --sentinel
```

### 高级实验3: JWT 刷新 Token

**实现双 Token 机制**：
- Access Token: 15分钟
- Refresh Token: 7天

---

## ✅ 实验完成检查清单

- [ ] 成功启动 3 个后端服务器实例
- [ ] 配置并启动 Nginx 负载均衡器
- [ ] 启动 Redis 容器
- [ ] 实现方案1: Sticky Session 的登录、查询、登出
- [ ] 实现方案2: Redis Session 的跨服务器共享
- [ ] 实现方案3: JWT Token 的认证和验证
- [ ] 编写测试脚本，对比三种方案的功能
- [ ] 进行性能测试，记录延迟和 QPS 数据
- [ ] 进行故障注入测试，观察不同方案的表现
- [ ] 填写实验记录表格
- [ ] 撰写实验总结报告

---

## 📝 实验报告模板

完成实验后，在 `notes/week1/` 目录创建实验报告：

```markdown
# 会话管理方案对比实验报告

## 实验时间
- 开始: ___________
- 结束: ___________
- 总时长: _____ 小时

## 实验环境
- 操作系统: _____
- Go 版本: _____
- Python 版本: _____
- Redis 版本: _____

## 实验结果

### 功能测试
(记录各方案的功能测试结果)

### 性能测试
(粘贴性能测试数据表格)

### 故障测试
(描述故障场景下的表现)

## 关键发现
1. _____
2. _____
3. _____

## 遇到的问题与解决
1. 问题: _____
   解决: _____

## 实验总结
(你对三种方案的理解和体会)

## 方案选择建议
(基于实验结果，给出选择建议)
```

---

**实验愉快！动手实践是最好的学习方式！** 🎉

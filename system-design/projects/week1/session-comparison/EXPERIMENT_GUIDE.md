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

| 方案                            | 后端语言         | 端口      | 特点               |
| ------------------------------- | ---------------- | --------- | ------------------ |
| **方案1: Sticky Session** | Go               | 8081-8083 | Nginx IP Hash 路由 |
| **方案2: Redis Session**  | Go + Python 示例 | 8091-8093 | 集中式存储         |
| **方案3: JWT Token**      | Go               | 8101-8103 | 无状态认证         |

---

## 🚀 快速开始（5分钟）

如果你想快速启动所有服务器并开始实验，按照以下步骤操作：

### 1. 一键启动所有服务器

```bash
cd /Users/yule/Desktop/opera/2_code/Interview-oriented-programming/system-design/projects/week1/session-comparison

# 启动所有服务器 (Sticky Session + Redis Session + JWT Token)
./start_all_servers.sh
```

这个脚本会自动：
- 启动 9 个 Go 服务器（每种方案 3 个）
- 启动 Redis Docker 容器（如果需要）
- 检查端口可用性
- 等待服务器完全启动
- 显示所有服务器的状态

### 2. 检查服务器状态

```bash
./check_servers.sh
```

你应该看到所有服务器都显示为"✅ 监听中"和"✅ 正常"。

### 3. 运行性能测试

```bash
cd test-scripts

# 完整性能测试 (延迟 + 吞吐量)
python performance_compare.py

# 或者只测试特定方案
python performance_compare.py --schemes sticky jwt --test latency
```

### 4. 停止所有服务器

```bash
cd ..
./stop_all_servers.sh
```

### 启动/停止脚本选项

```bash
# 只启动特定方案
./start_all_servers.sh sticky    # 只启动 Sticky Session
./start_all_servers.sh redis     # 只启动 Redis Session
./start_all_servers.sh jwt       # 只启动 JWT Token

# 只停止特定方案
./stop_all_servers.sh sticky
./stop_all_servers.sh redis
./stop_all_servers.sh jwt

# 强制清理所有 Go 进程（慎用）
./stop_all_servers.sh force
```

### 性能测试选项

```bash
cd test-scripts

# 只测试延迟
python performance_compare.py --test latency --requests 100

# 只测试吞吐量
python performance_compare.py --test throughput --duration 30 --concurrency 100

# 测试并发扩展性
python performance_compare.py --test concurrency

# 自定义参数
python performance_compare.py --test latency --requests 1000 --schemes redis
```

### 查看日志

```bash
# 实时查看单个日志
tail -f logs/sticky-server-1.log
tail -f logs/redis-server-1.log
tail -f logs/jwt-server-1.log

# 查看所有日志
tail -f logs/*.log
```

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
docker run -d --name nginx-sticky -p 8080:80 -v $(pwd)/docker/nginx-sticky.conf:/etc/nginx/conf.d/default.conf:ro nginx:alpine
```

#### Step 2.3: Nginx 负载均衡验证

**为什么需要验证 Nginx 在工作？**

在本地测试时，你可能会疑惑：
- Nginx 是否真的在转发请求？
- `ip_hash` 算法是否在工作？
- 如何证明负载均衡真的在分散请求？

##### 验证方法 1: 对比不同负载均衡算法

**创建两个 Nginx 配置**：

1. **ip_hash 配置** (`docker/nginx-sticky.conf`) - 已有
2. **round_robin 配置** (`docker/nginx-round-robin.conf`) - 新建

```nginx
# docker/nginx-round-robin.conf
upstream backend_round_robin {
    # Round Robin: 轮询分配（默认算法，不使用 ip_hash）
    server host.docker.internal:8081;
    server host.docker.internal:8082;
    server host.docker.internal:8083;
}

server {
    listen 80;
    server_name localhost;

    location / {
        proxy_pass http://backend_round_robin;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

**切换配置并对比**：

```bash
# 使用 ip_hash (Sticky Session)
docker rm -f nginx-sticky
docker run -d --name nginx-sticky -p 8080:80 \
  -v $(pwd)/docker/nginx-sticky.conf:/etc/nginx/conf.d/default.conf:ro \
  nginx:alpine

# 使用 round_robin (轮询)
docker rm -f nginx-sticky
docker run -d --name nginx-round-robin -p 8080:80 \
  -v $(pwd)/docker/nginx-round-robin.conf:/etc/nginx/conf.d/default.conf:ro \
  nginx:alpine
```

##### 验证方法 2: 使用测试脚本观察

**文件**: `test-scripts/verify_nginx.py`

```python
import requests

def test_nginx_algorithm():
    """检测当前 Nginx 使用的负载均衡算法"""
    nginx_url = "http://localhost:8080"

    # 登录并获取 Session
    session = requests.Session()
    session.post(f"{nginx_url}/login",
                json={"username": "test", "password": "123456"})

    # 连续发送 10 个请求
    success_count = 0
    servers = []

    for i in range(10):
        resp = session.get(f"{nginx_url}/profile")
        if resp.status_code == 200:
            success_count += 1
            servers.append(resp.json()['server_id'])

    # 判断算法
    unique_servers = set(servers)

    print(f"总请求: 10")
    print(f"成功: {success_count}")
    print(f"访问的服务器数: {len(unique_servers)}")

    if len(unique_servers) == 1:
        print("✅ 当前算法: ip_hash (Sticky Session)")
        print("   特征: 所有请求都路由到同一台服务器")
    else:
        print("✅ 当前算法: round_robin (轮询)")
        print("   特征: 请求分散到多台服务器")
        print(f"   成功率: {success_count * 10}%")
```

**运行测试**：

```bash
# 使用 ip_hash 时
python verify_nginx.py
# 输出:
# 成功: 10
# 访问的服务器数: 1
# ✅ 当前算法: ip_hash

# 使用 round_robin 时
python verify_nginx.py
# 输出:
# 成功: 3
# 访问的服务器数: 1 (但实际访问了3台)
# ✅ 当前算法: round_robin
# 成功率: 30% (约 1/3，因为 Session 只在1台服务器)
```

##### 验证方法 3: 可视化轮询过程

**文件**: `test-scripts/visualize_routing.py`

```python
def visualize_round_robin():
    """可视化 Round Robin 的轮询模式"""
    nginx_url = "http://localhost:8080"

    session = requests.Session()
    session.post(f"{nginx_url}/login",
                json={"username": "alice", "password": "123456"})

    print(f"{'序号':<6} {'路由到':<15} {'状态':<10}")
    print("-" * 40)

    for i in range(12):
        resp = session.get(f"{nginx_url}/profile")

        if resp.status_code == 200:
            server = resp.json()['server_id']
            status = "✅ 成功"
        else:
            server = "未知"
            status = "❌ 401"

        print(f"{i+1:<6} {server:<15} {status:<10}")

# 运行后看到周期性模式:
# 1      未知            ❌ 401     ← server-1
# 2      未知            ❌ 401     ← server-2
# 3      server-3       ✅ 成功     ← Session 在这里
# 4      未知            ❌ 401     ← server-1
# 5      未知            ❌ 401     ← server-2
# 6      server-3       ✅ 成功     ← 轮询回到 server-3
```

**关键观察点**：

- **ip_hash**: 成功率 100%，所有请求去同一台服务器
- **round_robin**: 成功率 ≈ 33%（3台服务器中1台有Session）
- **周期性模式**: 失败→失败→成功→失败→失败→成功（循环）

##### 验证方法 4: 直接访问 vs 通过 Nginx

```python
def compare_direct_vs_nginx():
    """对比直接访问后端和通过 Nginx"""

    # 通过 Nginx (自动路由)
    session_nginx = requests.Session()
    session_nginx.post("http://localhost:8080/login",
                      json={"username": "alice", "password": "123456"})
    resp = session_nginx.get("http://localhost:8080/profile")
    nginx_server = resp.json()['server_id']

    print(f"通过 Nginx (8080)     → {nginx_server}")

    # 直接访问各个后端
    for port in [8081, 8082, 8083]:
        session_direct = requests.Session()
        resp = session_direct.post(f"http://localhost:{port}/login",
                                   json={"username": "alice", "password": "123456"})
        resp = session_direct.get(f"http://localhost:{port}/profile")
        server_id = resp.json()['server_id']

        marker = "← Nginx 选择的" if server_id == nginx_server else ""
        print(f"直接访问 ({port})     → {server_id} {marker}")
```

**输出示例**：
```
通过 Nginx (8080)     → server-3
直接访问 (8081)       → server-1
直接访问 (8082)       → server-2
直接访问 (8083)       → server-3 ← Nginx 选择的
```

##### 两种算法对比总结

| 特性 | ip_hash | round_robin |
|------|---------|-------------|
| **路由依据** | 客户端 IP 地址哈希 | 轮询顺序 |
| **Sticky Session** | ✅ 自动保证 | ❌ 不保证 |
| **同一客户端** | 总是去同一台服务器 | 轮流访问各服务器 |
| **单机测试表现** | 成功率 100% | 成功率 ≈ 33% (3台) |
| **适用场景** | Session 存本地内存 | 无状态服务/共享存储 |
| **配置** | `ip_hash;` | 默认（无需配置） |

**为什么本地测试 ip_hash 所有请求都去同一台？**

```
本地测试: 所有请求来自 127.0.0.1 (同一 IP)
    ↓
ip_hash 计算: hash("127.0.0.1") % 3 = 固定值
    ↓
总是路由到: 同一台服务器 (如 server-3)
```

**生产环境**：用户来自不同 IP，会自动分散到不同服务器。

##### 快速验证命令

```bash
# 1. 启动 3 个后端服务器
cd sticky-session
PORT=8081 SERVER_ID=server-1 go run main.go &
PORT=8082 SERVER_ID=server-2 go run main.go &
PORT=8083 SERVER_ID=server-3 go run main.go &

# 2. 启动 Nginx (ip_hash)
cd ..
docker run -d --name nginx-sticky -p 8080:80 \
  -v $(pwd)/docker/nginx-sticky.conf:/etc/nginx/conf.d/default.conf:ro \
  nginx:alpine

# 3. 验证 ip_hash
cd test-scripts
python verify_nginx.py
# 期望: 成功率 100%，所有请求去同一台服务器

# 4. 切换到 round_robin
docker rm -f nginx-sticky
docker run -d --name nginx-round-robin -p 8080:80 \
  -v $(pwd)/../docker/nginx-round-robin.conf:/etc/nginx/conf.d/default.conf:ro \
  nginx:alpine

# 5. 验证 round_robin
python verify_round_robin.py
# 期望: 成功率 ≈ 33%，请求轮询到 3 台服务器

# 6. 查看 Nginx 日志
docker logs nginx-round-robin

# 7. 清理
docker rm -f nginx-round-robin
killall main  # 停止所有 Go 服务器
```

#### Step 2.4: 完整测试脚本

**文件**: `test-scripts/test_sticky_session.py` (使用 pytest)

查看文件顶部的文档字符串了解如何运行：

```bash
# 查看运行说明
head -n 63 test-scripts/test_sticky_session.py

# 运行所有测试
pytest test_sticky_session.py -v

# 只运行基础功能测试
pytest test_sticky_session.py::TestBasicFunctionality -v
```

**期望结果**：

使用 **ip_hash** 时：
- ✅ 同一客户端的请求总是路由到同一台服务器
- ✅ 所有测试通过 (13/13)
- ❌ 服务器宕机后，Session 丢失，需要重新登录

使用 **round_robin** 时：
- ✅ 请求被轮询分配到不同服务器
- ❌ 大部分测试失败（约 70% 失败率）
- ❌ 证明 Session 隔离问题（需要 Redis 或 JWT 解决）

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

运行`nginx-redis`
```bash
docker run -d --name nginx-redis -p 81:81 -v $(pwd)/docker/nginx-redis.conf:/etc/nginx/conf.d/default.conf:ro nginx:alpine
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

> **目标**: 通过科学的性能测试，量化对比三种会话管理方案的性能差异

#### 📊 测试指标说明

##### 1. 延迟 (Latency)

**定义**: 从发送请求到收到响应的时间

**关键指标**:
- **P50 (中位数)**: 50% 的请求延迟低于此值
- **P95**: 95% 的请求延迟低于此值 (常用于 SLA)
- **P99**: 99% 的请求延迟低于此值 (尾部延迟)
- **平均值**: 所有请求的平均延迟
- **最小值/最大值**: 最快和最慢的请求

**为什么重要**:
- 直接影响用户体验
- P99 比平均值更能反映实际体验（避免被平均值掩盖的慢请求）

##### 2. 吞吐量 (Throughput)

**定义**: 系统每秒能处理的请求数 (QPS - Queries Per Second)

**关键指标**:
- **QPS**: 每秒请求数
- **成功率**: 成功请求占总请求的比例
- **并发数**: 同时发起请求的线程/连接数

**为什么重要**:
- 反映系统的处理能力
- 决定系统能支撑的用户规模

##### 3. 并发扩展性 (Concurrency Scalability)

**定义**: 系统在不同并发数下的性能表现

**观察点**:
- QPS 随并发数的变化趋势
- 什么并发数下达到峰值 QPS
- 高并发下的错误率变化

**为什么重要**:
- 帮助确定系统的性能瓶颈
- 指导水平扩展决策

#### Step 5.1: 测试准备

确保所有服务器已启动：

```bash
# 使用一键启动脚本
./start_all_servers.sh

# 检查状态
./check_servers.sh

# 所有服务器应该显示为"✅ 监听中"和"✅ 正常"
```

#### Step 5.2: 执行性能测试

##### 测试 1: 延迟对比测试

**目的**: 对比三种方案的响应延迟

```bash
cd test-scripts

# 运行延迟测试 (100 个请求)
python performance_compare.py --test latency --requests 100
```

**预期输出示例**:
```
======================================================================
测试延迟: Sticky Session
======================================================================
发送 100 个请求...
  进度: 20/100 (20%)
  进度: 40/100 (40%)
  进度: 60/100 (60%)
  进度: 80/100 (80%)
  进度: 100/100 (100%)

结果:
  总请求数: 100
  成功: 100, 失败: 0
  成功率: 100.00%

  延迟统计:
    最小值:  0.35 ms
    平均值:  1.52 ms
    P50:     1.45 ms
    P95:     2.80 ms
    P99:     3.50 ms
    最大值:  5.20 ms

======================================================================
对比汇总: LATENCY
======================================================================

方案                 P50 (ms)     P95 (ms)     P99 (ms)     平均 (ms)
----------------------------------------------------------------------
Sticky Session       1.45         2.80         3.50         1.52
Redis Session        2.85         5.20         6.80         3.10
JWT Token            1.20         2.50         3.20         1.35

推荐:
  延迟最低: JWT Token (P50: 1.20 ms)
```

**如何解读**:
- **P50 < 2ms**: 延迟很低，用户体验好
- **P95 < 5ms**: 95% 的用户体验好
- **P99 > 10ms**: 需要关注，可能有性能问题

**差异原因分析**:
- **JWT Token 最快**: 只需验证签名，无存储访问
- **Sticky Session 较快**: 本地内存读取，无网络 I/O
- **Redis Session 较慢**: 每次请求需要访问 Redis (~1-2ms 网络延迟)

##### 测试 2: 吞吐量对比测试

**目的**: 对比三种方案的 QPS (每秒请求数)

```bash
# 运行吞吐量测试 (持续 10 秒，50 并发)
python performance_compare.py --test throughput --duration 10 --concurrency 50
```

**预期输出示例**:
```
======================================================================
测试吞吐量: Sticky Session
======================================================================
并发数: 50, 持续时间: 10 秒
✅ 设置了 50 个会话

开始压测...
  进度: 2/10 秒, 当前 QPS: 5243, 已提交: 10486
  进度: 4/10 秒, 当前 QPS: 5180, 已提交: 20720
  进度: 6/10 秒, 当前 QPS: 5210, 已提交: 31260
  进度: 8/10 秒, 当前 QPS: 5195, 已提交: 41560
  等待所有任务完成...

结果:
  持续时间: 10.02 秒
  总请求数: 52050
  成功: 52050, 失败: 0
  成功率: 100.00%
  QPS: 5194 请求/秒

======================================================================
对比汇总: THROUGHPUT
======================================================================

方案                 QPS             成功率          并发数
-----------------------------------------------------------------
Sticky Session       5194            100.00%         50
Redis Session        3520            100.00%         50
JWT Token            4850            100.00%         50

推荐:
  吞吐量最高: Sticky Session (QPS: 5194)
```

**如何解读**:
- **QPS > 5000**: 性能优秀
- **QPS 2000-5000**: 性能良好
- **QPS < 1000**: 可能有性能瓶颈

**差异原因分析**:
- **Sticky Session 最高**: 无网络 I/O，纯内存操作
- **Redis Session 最低**: Redis I/O 是瓶颈
- **JWT Token 较高**: 只需 CPU 验证签名，无 I/O

##### 测试 3: 并发扩展性测试

**目的**: 观察不同并发数下的性能变化

```bash
# 测试不同并发数 (10, 50, 100, 200)
python performance_compare.py --test concurrency
```

**预期输出示例**:
```
======================================================================
并发扩展性汇总: Sticky Session
======================================================================
并发数       QPS             成功率
---------------------------------------------
10           4850            100.00%
50           5194            100.00%
100          5380            100.00%
200          5420            99.95%

======================================================================
并发扩展性汇总: Redis Session
======================================================================
并发数       QPS             成功率
---------------------------------------------
10           3200            100.00%
50           3520            100.00%
100          3650            100.00%
200          3700            98.50%

======================================================================
并发扩展性汇总: JWT Token
======================================================================
并发数       QPS             成功率
---------------------------------------------
10           4500            100.00%
50           4850            100.00%
100          5050            100.00%
200          5180            99.80%
```

**如何解读**:

1. **线性扩展**: QPS 随并发数增加而增加
   - 说明系统未达到瓶颈
   - 可以继续增加并发

2. **达到平台期**: QPS 不再明显增加
   - 说明达到了系统瓶颈
   - 进一步增加并发没有意义

3. **成功率下降**: 高并发下错误率增加
   - 说明系统过载
   - 需要优化或水平扩展

**观察重点**:
- Sticky Session: 在并发 100 时达到峰值，说明本地内存访问已接近极限
- Redis Session: 扩展性较差，瓶颈在 Redis 网络 I/O
- JWT Token: 扩展性好，CPU 验证签名的性能瓶颈较高

##### 测试 4: 自定义参数测试

```bash
# 只测试 Sticky Session 和 JWT Token
python performance_compare.py --schemes sticky jwt --test latency

# 更长时间的吞吐量测试
python performance_compare.py --test throughput --duration 30 --concurrency 100

# 更多请求的延迟测试
python performance_compare.py --test latency --requests 1000
```

#### Step 5.3: 使用 Apache Bench (ab) 进行测试

##### 安装 Apache Bench

```bash
# macOS
brew install apache2

# Ubuntu/Debian
sudo apt-get install apache2-utils

# 验证
ab -V
```

##### 测试 Sticky Session

```bash
# 先登录获取 Cookie
SESSION_ID=$(curl -s -c - -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"123456"}' \
  | grep session_id | awk '{print $7}')

echo "Session ID: $SESSION_ID"

# 使用 ab 测试
ab -n 10000 -c 100 -C "session_id=$SESSION_ID" http://localhost:8081/profile
```

**输出解读**:
```
Server Software:
Server Hostname:        localhost
Server Port:            8081

Document Path:          /profile
Document Length:        XXX bytes

Concurrency Level:      100
Time taken for tests:   1.923 seconds
Complete requests:      10000
Failed requests:        0
Total transferred:      XXX bytes
Requests per second:    5200.10 [#/sec] (mean)    ← QPS
Time per request:       19.230 [ms] (mean)        ← 平均延迟
Time per request:       0.192 [ms] (mean, across all concurrent requests)

Percentage of the requests served within a certain time (ms)
  50%    18        ← P50
  66%    20
  75%    21
  80%    22
  90%    25
  95%    28        ← P95
  98%    32
  99%    35        ← P99
 100%    45 (longest request)
```

##### 测试 Redis Session

```bash
# 登录获取 Cookie
SESSION_ID=$(curl -s -c - -X POST http://localhost:8091/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"123456"}' \
  | grep sessionID | awk '{print $7}')

# 测试
ab -n 10000 -c 100 -C "sessionID=$SESSION_ID" http://localhost:8091/profile
```

##### 测试 JWT Token

```bash
# 登录获取 Token
TOKEN=$(curl -s -X POST http://localhost:8010/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"123456"}' \
  | jq -r '.token')

echo "Token: ${TOKEN:0:50}..."

# 测试
ab -n 10000 -c 100 -H "Authorization: Bearer $TOKEN" http://localhost:8010/profile
```

#### Step 5.4: 高级测试场景

##### 场景 1: 模拟真实流量模式

**目的**: 模拟用户登录 → 多次访问 → 登出的真实流程

```python
# test_realistic_traffic.py
import requests
import time
import random

def simulate_user_session(scheme_url, num_requests=10):
    """模拟一个用户会话"""
    session = requests.Session()

    # 1. 登录
    resp = session.post(f"{scheme_url}/login",
                       json={"username": f"user_{random.randint(1,1000)}",
                             "password": "123456"})

    if resp.status_code != 200:
        return 0

    # 2. 多次访问
    success_count = 0
    for i in range(num_requests):
        resp = session.get(f"{scheme_url}/profile")
        if resp.status_code == 200:
            success_count += 1

        # 模拟用户思考时间
        time.sleep(random.uniform(0.1, 0.5))

    return success_count

# 模拟 100 个用户并发访问
from concurrent.futures import ThreadPoolExecutor

with ThreadPoolExecutor(max_workers=100) as executor:
    futures = [executor.submit(simulate_user_session, "http://localhost:8081")
               for _ in range(100)]

    total_success = sum(f.result() for f in futures)
    print(f"成功请求: {total_success}/1000")
```

##### 场景 2: 长时间稳定性测试

**目的**: 测试长时间运行下的性能稳定性

```bash
# 运行 5 分钟的压测
python performance_compare.py --test throughput --duration 300 --concurrency 50

# 观察:
# - QPS 是否稳定
# - 错误率是否增加
# - 内存是否泄漏 (使用 top/htop 监控)
```

##### 场景 3: Redis 连接池测试

**目的**: 观察 Redis 连接池对性能的影响

修改 Redis Session 服务器代码：

```go
// 增加连接池大小
var rdb = redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    PoolSize:     100,  // 从默认 10 增加到 100
    MinIdleConns: 10,
})
```

重新测试并对比 QPS 变化。

#### Step 5.5: 性能测试结果记录

将测试结果填入下表，方便对比分析：

##### 延迟测试结果

| 方案           | P50 (ms) | P95 (ms) | P99 (ms) | 平均 (ms) | 最大 (ms) |
| -------------- | -------- | -------- | -------- | --------- | --------- |
| Sticky Session | ____     | ____     | ____     | ____      | ____      |
| Redis Session  | ____     | ____     | ____     | ____      | ____      |
| JWT Token      | ____     | ____     | ____     | ____      | ____      |

##### 吞吐量测试结果

| 方案           | 并发数 | QPS    | 成功率 (%) | CPU 占用 (%) | 内存占用 (MB) |
| -------------- | ------ | ------ | ---------- | ------------ | ------------- |
| Sticky Session | 50     | ____   | ____       | ____         | ____          |
| Redis Session  | 50     | ____   | ____       | ____         | ____          |
| JWT Token      | 50     | ____   | ____       | ____         | ____          |

##### 并发扩展性测试结果

**Sticky Session**:

| 并发数 | QPS  | 成功率 (%) | P99 延迟 (ms) |
| ------ | ---- | ---------- | ------------- |
| 10     | ____ | ____       | ____          |
| 50     | ____ | ____       | ____          |
| 100    | ____ | ____       | ____          |
| 200    | ____ | ____       | ____          |

**Redis Session**:

| 并发数 | QPS  | 成功率 (%) | P99 延迟 (ms) |
| ------ | ---- | ---------- | ------------- |
| 10     | ____ | ____       | ____          |
| 50     | ____ | ____       | ____          |
| 100    | ____ | ____       | ____          |
| 200    | ____ | ____       | ____          |

**JWT Token**:

| 并发数 | QPS  | 成功率 (%) | P99 延迟 (ms) |
| ------ | ---- | ---------- | ------------- |
| 10     | ____ | ____       | ____          |
| 50     | ____ | ____       | ____          |
| 100    | ____ | ____       | ____          |
| 200    | ____ | ____       | ____          |

#### Step 5.6: 性能分析与结论

##### 预期性能排名

**延迟 (越低越好)**:
1. **JWT Token** (~1.2ms P50) - 只需验证签名，无 I/O
2. **Sticky Session** (~1.5ms P50) - 本地内存访问
3. **Redis Session** (~2.8ms P50) - Redis 网络 I/O

**吞吐量 (越高越好)**:
1. **Sticky Session** (~5200 QPS) - 纯内存操作
2. **JWT Token** (~4850 QPS) - CPU 验证签名
3. **Redis Session** (~3520 QPS) - Redis 网络 I/O 瓶颈

##### 性能差异原因分析

**Sticky Session 性能最高的原因**:
- ✅ 本地内存访问，延迟极低 (~0.1ms)
- ✅ 无网络 I/O
- ✅ sync.Map 并发读取性能好
- ❌ 但扩展性差，服务器宕机丢失 Session

**Redis Session 性能相对较低的原因**:
- ❌ 每次请求需要 Redis I/O (~1-2ms)
- ❌ 网络延迟累积
- ❌ Redis 连接池可能成为瓶颈
- ✅ 但可扩展性强，高可用

**JWT Token 性能中等的原因**:
- ✅ 无存储访问，完全无状态
- ✅ 只需 CPU 验证签名 (~0.3ms)
- ❌ Token 体积大，网络传输开销
- ✅ 扩展性最好，天然支持分布式

##### 性能优化建议

**Sticky Session**:
- 使用更高效的哈希表 (如 `sync.Map` 已经很好)
- 定期清理过期 Session
- 考虑 Session 大小，避免存储大对象

**Redis Session**:
- 增加 Redis 连接池大小
- 使用 Redis Cluster 提高吞吐量
- 开启 Redis 持久化 (RDB/AOF)
- 考虑使用本地缓存 (L1 Cache)

**JWT Token**:
- 减小 Token 体积 (只存储必要字段)
- 使用更快的签名算法 (HS256 已经很快)
- 考虑 Token 压缩

#### Step 5.7: 故障排查

##### 问题 1: QPS 远低于预期

**可能原因**:
- 服务器 CPU/内存不足
- 网络延迟过高
- 数据库/Redis 连接池不足

**排查方法**:
```bash
# 查看 CPU 占用
top -o cpu

# 查看网络延迟
ping localhost
ping 127.0.0.1

# 查看 Redis 连接数
redis-cli
CLIENT LIST | wc -l
```

##### 问题 2: 高并发下错误率增加

**可能原因**:
- 连接池耗尽
- 超时设置过短
- 服务器过载

**排查方法**:
```bash
# 查看错误日志
# 检查是否有 "connection refused" 或 "timeout" 错误

# 增加连接池大小
# 增加超时时间
```

##### 问题 3: 延迟不稳定

**可能原因**:
- GC 导致延迟尖刺
- 网络抖动
- 磁盘 I/O (如果有日志写入)

**排查方法**:
```bash
# Go 程序开启 pprof
import _ "net/http/pprof"

# 访问 pprof
go tool pprof http://localhost:6060/debug/pprof/heap
```

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

| 功能           | Sticky Session | Redis Session | JWT Token     |
| -------------- | -------------- | ------------- | ------------- |
| 跨服务器共享   | ❌             | ✅            | ✅            |
| 服务器宕机恢复 | ❌             | ✅            | ✅            |
| 主动登出       | ✅             | ✅            | ⚠️ 需黑名单 |
| 水平扩展       | ⚠️ 困难      | ✅            | ✅            |
| 依赖外部服务   | ❌             | ✅ Redis      | ❌            |

### 性能数据（自己测试填写）

| 方案           | P50 延迟(ms) | P99 延迟(ms) | QPS   | 内存占用 |
| -------------- | ------------ | ------------ | ----- | -------- |
| Sticky Session | _____        | _____        | _____ | _____    |
| Redis Session  | _____        | _____        | _____ | _____    |
| JWT Token      | _____        | _____        | _____ | _____    |

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

## 📚 相关文档

- **[README.md](./README.md)** - 项目总览
- **[SCRIPTS_README.md](./SCRIPTS_README.md)** - 批量管理脚本详细说明
- **[PERFORMANCE_FIX_NOTES.md](./PERFORMANCE_FIX_NOTES.md)** - 性能脚本修复说明
- **[sticky-session/README.md](./sticky-session/README.md)** - Sticky Session 实现细节
- **[redis-session/README.md](./redis-session/README.md)** - Redis Session 实现细节
- **[jwt-token/README.md](./jwt-token/README.md)** - JWT Token 实现细节

---

**实验愉快！动手实践是最好的学习方式！** 🎉

---

> **注**: 本文档整合了原 `QUICK_START.md` 和 `PERFORMANCE_TESTING_GUIDE.md` 的内容，提供从快速开始到深入性能测试的完整实验指南。

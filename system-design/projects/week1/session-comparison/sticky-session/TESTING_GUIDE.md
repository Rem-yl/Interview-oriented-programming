# Sticky Session 实验 - 测试驱动学习指南

> **学习方法**: 通过运行测试代码，观察实际输出与预期输出的差异，理解你的代码需要改进的地方

---

## 🎯 实验目标

通过测试理解 Sticky Session 的三个核心问题：
1. **Session ID 如何传递给客户端？** → Cookie
2. **服务器如何识别客户端？** → 读取 Cookie 中的 Session ID
3. **为什么 Session 无法跨服务器共享？** → 每个服务器独立存储

---

## 测试 1: 登录功能测试

### 🔍 问题：你的代码缺少什么？

**当前你的代码**：
```go
// loginHandler 最后一行
c.JSON(http.StatusOK, gin.H{"status": "ok", "data": fmt.Sprintf("get user: %s", req.UserName)})
```

**问题**：虽然生成了 `sessionID`，但**没有告诉客户端**这个 ID 是什么！

---

### ✅ 测试代码 1.1：登录后是否返回 Cookie？

创建测试文件：`test_login.py`

```python
import requests

# 测试：登录
resp = requests.post('http://localhost:8080/login',
                     json={'username': 'alice', 'password': '123456'})

print("=== 测试 1.1: 登录响应 ===")
print(f"状态码: {resp.status_code}")
print(f"响应体: {resp.json()}")
print(f"\n=== 关键点：Cookie ===")
print(f"Cookies: {resp.cookies}")
print(f"是否有 session_id Cookie: {'session_id' in resp.cookies}")

if 'session_id' in resp.cookies:
    print(f"✅ 测试通过：Session ID = {resp.cookies['session_id']}")
else:
    print(f"❌ 测试失败：没有返回 session_id Cookie")
    print(f"\n💡 你需要做什么？")
    print(f"   在 loginHandler 中添加：")
    print(f"   c.SetCookie('session_id', sessionID, 3600, '/', '', false, true)")
```

**运行你的代码**：
```bash
cd sticky-session
go run main.go
```

**运行测试**：
```bash
python test_login.py
```

---

### 📊 预期输出对比

#### ❌ 你当前代码的输出：
```
=== 测试 1.1: 登录响应 ===
状态码: 200
响应体: {'status': 'ok', 'data': 'get user: alice'}

=== 关键点：Cookie ===
Cookies: <RequestsCookieJar[]>
是否有 session_id Cookie: False
❌ 测试失败：没有返回 session_id Cookie

💡 你需要做什么？
   在 loginHandler 中添加：
   c.SetCookie('session_id', sessionID, 3600, '/', '', false, true)
```

#### ✅ 正确代码的输出：
```
=== 测试 1.1: 登录响应 ===
状态码: 200
响应体: {'message': 'Login successful', 'server_id': 'server-1'}

=== 关键点：Cookie ===
Cookies: <RequestsCookieJar[Cookie(name='session_id', value='a1b2c3d4-...')]>
是否有 session_id Cookie: True
✅ 测试通过：Session ID = a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

---

### 💡 你需要添加的代码

在 `loginHandler` 函数中，`sessionStore.Store()` 之后添加：

```go
// 【关键】设置 Cookie，将 Session ID 返回给客户端
c.SetCookie(
    "session_id",  // Cookie 名称
    sessionID,     // Cookie 值
    3600,          // 过期时间（秒）
    "/",           // Path
    "",            // Domain
    false,         // Secure
    true,          // HttpOnly
)
```

**为什么需要这个？**
- 客户端不会记住 Session ID，需要通过 Cookie 保存
- 后续请求时，浏览器会自动携带这个 Cookie

---

## 测试 2: 获取用户信息功能

### 🔍 问题：你的代码缺少什么？

**当前你的代码**：
- 只有 `/login` 接口
- 没有 `/profile` 接口来验证 Session 是否有效

---

### ✅ 测试代码 2.1：使用 Session 获取用户信息

创建测试文件：`test_profile.py`

```python
import requests

# Step 1: 登录
session = requests.Session()  # Session 对象会自动管理 Cookie
resp = session.post('http://localhost:8080/login',
                    json={'username': 'alice', 'password': '123456'})

print("=== Step 1: 登录 ===")
print(f"状态码: {resp.status_code}")
print(f"Cookies: {session.cookies}")

# Step 2: 访问 /profile
print("\n=== Step 2: 访问 /profile ===")
try:
    resp = session.get('http://localhost:8080/profile')
    print(f"状态码: {resp.status_code}")
    print(f"响应体: {resp.json()}")

    if resp.status_code == 200:
        print("✅ 测试通过：能够获取用户信息")
    else:
        print("❌ 测试失败：无法获取用户信息")
except Exception as e:
    print(f"❌ 测试失败：接口不存在 - {e}")
    print("\n💡 你需要做什么？")
    print("   添加 GET /profile 接口")
```

---

### 📊 预期输出对比

#### ❌ 你当前代码的输出：
```
=== Step 1: 登录 ===
状态码: 200
Cookies: <RequestsCookieJar[]>

=== Step 2: 访问 /profile ===
❌ 测试失败：接口不存在 - 404 page not found

💡 你需要做什么？
   添加 GET /profile 接口
```

#### ✅ 正确代码的输出：
```
=== Step 1: 登录 ===
状态码: 200
Cookies: <RequestsCookieJar[Cookie(name='session_id', value='abc123')]>

=== Step 2: 访问 /profile ===
状态码: 200
响应体: {
  'user_id': 1001,
  'username': 'alice',
  'login_time': '2025-10-21 16:30:15',
  'server_id': 'server-1',
  'created_by': 'server-1',
  'session_match': true
}
✅ 测试通过：能够获取用户信息
```

---

### 💡 你需要添加的代码

在 `main()` 函数中注册新路由：
```go
r.GET("/profile", profileHandler)
```

创建 `profileHandler` 函数：
```go
func profileHandler(c *gin.Context) {
    // 1. 从 Cookie 获取 Session ID
    sessionID, err := c.Cookie("session_id")
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
        return
    }

    // 2. 从 sessionStore 查找
    value, exists := sessionStore.Load(sessionID)
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found"})
        return
    }

    // 3. 类型转换
    session := value.(Session)

    // 4. 返回用户信息
    c.JSON(http.StatusOK, gin.H{
        "username": session.UserName,
        "server_id": session.ServerID,
    })
}
```

**为什么需要这个？**
- 验证 Session 是否有效
- 让客户端能获取登录用户的信息

---

## 测试 3: 多服务器 Session 隔离

### 🔍 核心问题：为什么需要 ServerID？

这是 **Sticky Session 实验的关键**！

**问题**：
- 如果有 3 台服务器（Server-1, Server-2, Server-3）
- 用户在 Server-1 登录，Session 存储在 Server-1 的内存中
- 如果请求被路由到 Server-2，Server-2 能找到这个 Session 吗？

**答案**：❌ **不能！** 因为每个服务器的内存是独立的。

---

### ✅ 测试代码 3.1：启动多个服务器实例

**测试场景**：
1. 启动 3 个服务器（不同端口）
2. 在 Server-1 登录
3. 直接访问 Server-2 和 Server-3 的 `/profile`
4. 观察是否能找到 Session

---

**Step 1: 启动 3 个服务器**

```bash
# 终端 1
PORT=8081 SERVER_ID=server-1 go run main.go

# 终端 2
PORT=8082 SERVER_ID=server-2 go run main.go

# 终端 3
PORT=8083 SERVER_ID=server-3 go run main.go
```

**问题**：你的代码中 `ServerID` 是硬编码的 `"0"`，无法区分服务器。

---

**Step 2: 运行测试** - `test_multi_server.py`

```python
import requests

# 在 Server-1 (8081) 登录
print("=== Step 1: 在 Server-1 登录 ===")
resp = requests.post('http://localhost:8081/login',
                     json={'username': 'alice', 'password': '123456'})
print(f"状态码: {resp.status_code}")

# 获取 Cookie
session_cookie = resp.cookies.get('session_id')
print(f"Session ID: {session_cookie}")
cookies = {'session_id': session_cookie}

# 测试 Server-1 (应该成功)
print("\n=== Step 2: 访问 Server-1 的 /profile ===")
resp1 = requests.get('http://localhost:8081/profile', cookies=cookies)
print(f"Server-1 响应: {resp1.status_code}")
if resp1.status_code == 200:
    print(f"  用户信息: {resp1.json()}")
    print(f"  ✅ Server-1 找到了 Session")

# 测试 Server-2 (应该失败)
print("\n=== Step 3: 访问 Server-2 的 /profile ===")
resp2 = requests.get('http://localhost:8082/profile', cookies=cookies)
print(f"Server-2 响应: {resp2.status_code}")
if resp2.status_code == 401:
    print(f"  错误信息: {resp2.json()}")
    print(f"  ✅ Server-2 找不到 Session (符合预期)")
elif resp2.status_code == 200:
    print(f"  ❌ Server-2 找到了 Session (不应该!)")

# 测试 Server-3 (应该失败)
print("\n=== Step 4: 访问 Server-3 的 /profile ===")
resp3 = requests.get('http://localhost:8083/profile', cookies=cookies)
print(f"Server-3 响应: {resp3.status_code}")
if resp3.status_code == 401:
    print(f"  错误信息: {resp3.json()}")
    print(f"  ✅ Server-3 找不到 Session (符合预期)")
```

---

### 📊 预期输出

#### ❌ 你当前代码的问题：

```
=== Step 1: 在 Server-1 登录 ===
状态码: 200
Session ID: None
❌ 没有返回 Cookie！

(后续测试无法继续)
```

#### ✅ 修复后的预期输出：

```
=== Step 1: 在 Server-1 登录 ===
状态码: 200
Session ID: abc123-def456-...

=== Step 2: 访问 Server-1 的 /profile ===
Server-1 响应: 200
  用户信息: {'username': 'alice', 'server_id': 'server-1'}
  ✅ Server-1 找到了 Session

=== Step 3: 访问 Server-2 的 /profile ===
Server-2 响应: 401
  错误信息: {'error': 'Session not found', 'hint': 'Session created by another server'}
  ✅ Server-2 找不到 Session (符合预期)

=== Step 4: 访问 Server-3 的 /profile ===
Server-3 响应: 401
  错误信息: {'error': 'Session not found'}
  ✅ Server-3 找不到 Session (符合预期)
```

**这就是 Sticky Session 的核心问题！**

---

### 💡 你需要修改的代码

#### 1. 从环境变量读取 ServerID

在 `main.go` 文件开头添加：
```go
import "os"

// 获取环境变量
func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}

var serverID = getEnv("SERVER_ID", "server-default")
var port = getEnv("PORT", "8080")
```

#### 2. 在 Session 中记录 ServerID

修改 `loginHandler`：
```go
session := Session{
    UserName:  req.UserName,
    LoginTime: time.Now(),
    ServerID:  serverID, // 使用全局变量，不是硬编码 "0"
}
```

#### 3. main() 中使用动态端口

```go
func main() {
    r := gin.Default()
    r.POST("/login", loginHandler)
    r.GET("/profile", profileHandler)

    r.Run(":" + port) // 使用环境变量中的端口
}
```

---

## 测试 4: 调试接口 - 查看服务器的 Session 列表

### 🔍 问题：如何验证每个服务器存储了哪些 Session？

**测试场景**：
- 创建 5 个用户，登录到不同服务器
- 查看每个服务器存储了多少 Session

---

### ✅ 测试代码 4.1：Session 分布查看

`test_debug.py`

```python
import requests

# 创建 5 个用户
print("=== 创建 5 个用户 ===")
for i in range(5):
    # 随机路由到不同端口（模拟负载均衡）
    port = 8081 + (i % 3)
    resp = requests.post(f'http://localhost:{port}/login',
                         json={'username': f'user{i}', 'password': '123456'})
    print(f"User{i} -> Server on port {port}: {resp.status_code}")

# 查看每个服务器的 Session
print("\n=== 查看各服务器的 Session ===")
for port in [8081, 8082, 8083]:
    resp = requests.get(f'http://localhost:{port}/debug/sessions')
    data = resp.json()
    print(f"\nServer {port}:")
    print(f"  Server ID: {data['server_id']}")
    print(f"  Session 数量: {data['session_count']}")
    print(f"  Sessions: {data['sessions']}")
```

---

### 📊 预期输出

```
=== 创建 5 个用户 ===
User0 -> Server on port 8081: 200
User1 -> Server on port 8082: 200
User2 -> Server on port 8083: 200
User3 -> Server on port 8081: 200
User4 -> Server on port 8082: 200

=== 查看各服务器的 Session ===

Server 8081:
  Server ID: server-1
  Session 数量: 2
  Sessions: [
    {'session_id': 'abc123', 'username': 'user0', 'server_id': 'server-1'},
    {'session_id': 'def456', 'username': 'user3', 'server_id': 'server-1'}
  ]

Server 8082:
  Server ID: server-2
  Session 数量: 2
  Sessions: [
    {'session_id': 'ghi789', 'username': 'user1', 'server_id': 'server-2'},
    {'session_id': 'jkl012', 'username': 'user4', 'server_id': 'server-2'}
  ]

Server 8083:
  Server ID: server-3
  Session 数量: 1
  Sessions: [
    {'session_id': 'mno345', 'username': 'user2', 'server_id': 'server-3'}
  ]
```

**观察点**：
- ✅ 每个服务器只存储自己创建的 Session
- ✅ Session 总数 = 5，分布在 3 个服务器上

---

### 💡 你需要添加的代码

添加调试接口：

```go
func debugSessionsHandler(c *gin.Context) {
    sessions := []map[string]interface{}{}

    sessionStore.Range(func(key, value interface{}) bool {
        session := value.(Session)
        sessions = append(sessions, map[string]interface{}{
            "session_id": key.(string),
            "username":   session.UserName,
            "server_id":  session.ServerID,
        })
        return true
    })

    c.JSON(http.StatusOK, gin.H{
        "server_id":     serverID,
        "session_count": len(sessions),
        "sessions":      sessions,
    })
}
```

在 `main()` 中注册：
```go
r.GET("/debug/sessions", debugSessionsHandler)
```

---

## 📝 完整的测试清单

### ✅ 必须实现的功能

- [ ] **测试 1**: 登录后返回 `session_id` Cookie
- [ ] **测试 2**: `/profile` 接口能读取 Cookie 并返回用户信息
- [ ] **测试 3**: 不同服务器的 Session 是隔离的
- [ ] **测试 4**: 调试接口能查看服务器的 Session 列表

### ✅ 代码检查清单

- [ ] `loginHandler` 中调用了 `c.SetCookie()`
- [ ] 添加了 `profileHandler` 函数
- [ ] 添加了 `debugSessionsHandler` 函数
- [ ] `ServerID` 从环境变量读取，不是硬编码
- [ ] `port` 从环境变量读取
- [ ] `main()` 中注册了所有路由

---

## 🚀 完整的测试流程

### Step 1: 启动服务器

```bash
# 终端 1
cd sticky-session
PORT=8081 SERVER_ID=server-1 go run main.go

# 终端 2
PORT=8082 SERVER_ID=server-2 go run main.go

# 终端 3
PORT=8083 SERVER_ID=server-3 go run main.go
```

### Step 2: 运行所有测试

```bash
# 终端 4
python test_login.py
python test_profile.py
python test_multi_server.py
python test_debug.py
```

### Step 3: 验证日志输出

观察服务器终端的日志，应该看到：
```
[server-1] User 'alice' logged in, Session ID: abc123...
[server-1] User 'alice' accessed profile
[server-2] Profile request failed: Session 'abc123' not found
```

---

## 💡 关键理解

通过这些测试，你应该理解：

1. **Cookie 的作用**：
   - 客户端无法"记住" Session ID
   - Cookie 是服务器和客户端之间传递 Session ID 的桥梁

2. **Session 的本质**：
   - Session 数据存储在服务器内存中（`sync.Map`）
   - 每个服务器实例的内存是独立的

3. **Sticky Session 的局限**：
   - Session 无法跨服务器共享
   - 如果服务器宕机，Session 会丢失

4. **为什么需要 Nginx ip_hash**：
   - 保证同一客户端的请求总是路由到同一台服务器
   - 否则客户端会频繁遇到 "Session not found" 错误

---

## 🎯 现在开始修改你的代码吧！

按照测试的预期输出，逐步完善你的代码。每完成一个功能，就运行对应的测试验证。

有任何问题随时问我！🚀

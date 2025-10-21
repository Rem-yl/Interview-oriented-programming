# 会话管理方案对比实验 - 问题驱动式实验设计

> **实验日期**: 2025-10-21
> **实验目标**: 通过科学的实验方法，深入理解三种会话管理方案的工作原理、性能特性和适用场景
> **实验方法**: 问题驱动 + 假设验证 + 数据对比

---

## 🔬 实验方法论

### 科学实验流程

每个实验遵循以下步骤：

1. **提出问题** - 明确要解决的核心问题
2. **建立假设** - 基于理论推导实验假设
3. **设计实验** - 确定实验步骤和测试方案
4. **执行实验** - 实施测试并收集数据
5. **分析结果** - 对比预期与实际结果
6. **得出结论** - 回答问题并总结规律

### 实验环境要求

```
操作系统: macOS / Linux
Go 版本: >= 1.21
Python 版本: >= 3.8
Docker: 已安装并运行
Redis: 7.x (Docker 容器)
Nginx: 1.25+ (Docker 容器)
```

---

## 实验组一：Session Affinity (粘滞会话)

### 实验 1.1：验证 IP Hash 路由一致性

#### 🎯 核心问题
**Q: Nginx 的 `ip_hash` 策略能否保证同一客户端的请求总是路由到同一台服务器？**

#### 💡 实验假设
- **H1**: 使用 `ip_hash` 后，同一客户端的所有请求会路由到同一台后端服务器
- **H2**: 不同客户端会被分配到不同的服务器（负载均衡）
- **H3**: 客户端 IP 改变后，会被路由到不同的服务器

#### 📋 实验设计

**前置条件**：
- 启动 3 个 Go 后端服务器（端口 8081, 8082, 8083）
- 每个服务器有唯一 ID（环境变量 `SERVER_ID`）
- Nginx 配置 `ip_hash` 策略

**测试步骤**：

1. **单客户端连续请求测试**
   ```python
   # test_sticky.py
   import requests

   session = requests.Session()

   # 发送 20 次请求
   servers = []
   for i in range(20):
       resp = session.get('http://localhost:8080/health')
       server_id = resp.json()['server_id']
       servers.append(server_id)

   # 验证：所有请求是否路由到同一服务器
   unique_servers = set(servers)
   print(f"访问的服务器: {servers}")
   print(f"唯一服务器数: {len(unique_servers)}")
   ```

2. **多客户端分布测试**
   ```python
   # 模拟 10 个不同客户端
   server_distribution = {}

   for client_id in range(10):
       session = requests.Session()
       resp = session.get('http://localhost:8080/health')
       server_id = resp.json()['server_id']

       server_distribution[client_id] = server_id

   # 统计每个服务器处理的客户端数
   from collections import Counter
   print(Counter(server_distribution.values()))
   ```

3. **IP 变化测试**（高级）
   ```bash
   # 使用不同代理 IP 请求
   curl --interface eth0 http://localhost:8080/health
   curl --interface eth1 http://localhost:8080/health
   # 观察是否路由到不同服务器
   ```

#### 📊 数据收集

| 测试场景 | 请求次数 | 访问的服务器 | 服务器切换次数 | 结论 |
|---------|---------|-------------|---------------|------|
| 单客户端连续请求 | 20 | [填写] | [填写] | ✅/❌ |
| 10个不同客户端 | 10 | [填写分布] | - | ✅/❌ |
| IP变化测试 | 2 | [填写] | [填写] | ✅/❌ |

#### ✅ 预期结果
- 单客户端的所有请求路由到**同一台**服务器（服务器切换次数 = 0）
- 多客户端被**分散**到不同服务器（理想情况：每台服务器 3-4 个客户端）
- IP 变化后路由到**不同**服务器

#### 🔍 结论验证
- [ ] H1 验证：单客户端路由一致性 ________
- [ ] H2 验证：负载均衡效果 ________
- [ ] H3 验证：IP 变化影响 ________

---

### 实验 1.2：Session 数据隔离性验证

#### 🎯 核心问题
**Q: 每台服务器的 Session 数据是否完全隔离？服务器 A 能否访问服务器 B 的 Session？**

#### 💡 实验假设
- **H1**: 每台服务器的 Session 存储在本地内存，相互隔离
- **H2**: 客户端携带 Session ID，但只有创建该 Session 的服务器能识别
- **H3**: 如果 Nginx 路由到错误的服务器，会返回"Session 不存在"

#### 📋 实验设计

**测试步骤**：

1. **登录并获取 Session**
   ```python
   session = requests.Session()

   # 登录（假设路由到 Server-1）
   resp = session.post('http://localhost:8080/login',
                       json={'username': 'alice', 'password': '123456'})

   session_id = session.cookies.get('session_id')
   print(f"Session ID: {session_id}")

   # 验证登录成功
   resp = session.get('http://localhost:8080/profile')
   print(f"Login Server: {resp.json()['server_id']}")
   ```

2. **直接访问其他服务器（绕过 Nginx）**
   ```python
   # 手动携带 Session Cookie 访问不同服务器
   cookies = {'session_id': session_id}

   # 访问 Server-1 (应该成功)
   resp1 = requests.get('http://localhost:8081/profile', cookies=cookies)
   print(f"Server-1: {resp1.status_code}, {resp1.json()}")

   # 访问 Server-2 (应该失败 - Session 不存在)
   resp2 = requests.get('http://localhost:8082/profile', cookies=cookies)
   print(f"Server-2: {resp2.status_code}, {resp2.json()}")

   # 访问 Server-3 (应该失败)
   resp3 = requests.get('http://localhost:8083/profile', cookies=cookies)
   print(f"Server-3: {resp3.status_code}, {resp3.json()}")
   ```

3. **Session 数据内容验证**
   ```go
   // 在 Go 服务器中添加调试接口
   http.HandleFunc("/debug/sessions", func(w http.ResponseWriter, r *http.Request) {
       sessions := []string{}
       sessionStore.Range(func(key, value interface{}) bool {
           sessions = append(sessions, key.(string))
           return true
       })
       json.NewEncoder(w).Encode(map[string]interface{}{
           "server_id": serverID,
           "session_count": len(sessions),
           "session_ids": sessions,
       })
   })
   ```

   ```python
   # 查看每台服务器的 Session 列表
   for port in [8081, 8082, 8083]:
       resp = requests.get(f'http://localhost:{port}/debug/sessions')
       print(f"Server {port}: {resp.json()}")
   ```

#### 📊 数据收集

| 测试项 | Server-1 (创建者) | Server-2 | Server-3 | 结论 |
|-------|-----------------|----------|----------|------|
| 携带 Session Cookie 访问 | [状态码] | [状态码] | [状态码] | ✅/❌ |
| Session 是否存在 | ✅ | ❌ | ❌ | ✅/❌ |
| Session 数据内容 | [user_id] | null | null | ✅/❌ |

#### ✅ 预期结果
- Server-1 返回 200，能正确获取用户信息
- Server-2 和 Server-3 返回 401 (Unauthorized)，提示"Session 不存在"
- 调试接口显示只有 Server-1 存储了该 Session

#### 🔍 结论验证
- [ ] H1 验证：Session 数据隔离 ________
- [ ] H2 验证：只有创建者能识别 ________
- [ ] H3 验证：错误路由导致失败 ________

---

### 实验 1.3：服务器宕机的 Session 丢失测试

#### 🎯 核心问题
**Q: 服务器宕机后，该服务器上的所有 Session 是否会丢失？系统能否自动恢复？**

#### 💡 实验假设
- **H1**: 服务器宕机后，其本地 Session 全部丢失
- **H2**: 客户端再次请求时，会被路由到其他存活的服务器
- **H3**: 由于 Session 不存在，客户端需要重新登录

#### 📋 实验设计

**测试步骤**：

1. **建立多个用户 Session**
   ```python
   # 创建 5 个用户的 Session
   sessions = []
   for i in range(5):
       s = requests.Session()
       resp = s.post('http://localhost:8080/login',
                     json={'username': f'user{i}', 'password': '123456'})

       profile = s.get('http://localhost:8080/profile').json()
       sessions.append({
           'session': s,
           'username': f'user{i}',
           'server_id': profile['server_id']
       })
       print(f"User{i} -> {profile['server_id']}")
   ```

2. **记录 Session 分布**
   ```python
   from collections import Counter
   server_distribution = Counter([s['server_id'] for s in sessions])
   print(f"Session 分布: {dict(server_distribution)}")

   # 示例输出:
   # {'server-1': 2, 'server-2': 2, 'server-3': 1}
   ```

3. **杀死负载最高的服务器**
   ```bash
   # 手动操作：找到负载最高的服务器进程并杀死
   ps aux | grep "SERVER_ID=server-1"
   kill -9 <PID>

   # 或使用脚本
   # pkill -f "PORT=8081"
   ```

4. **验证 Session 状态**
   ```python
   # 等待 2 秒让 Nginx 检测到服务器下线
   time.sleep(2)

   results = []
   for s in sessions:
       try:
           resp = s['session'].get('http://localhost:8080/profile')
           results.append({
               'username': s['username'],
               'original_server': s['server_id'],
               'status': resp.status_code,
               'new_server': resp.json().get('server_id', 'N/A') if resp.status_code == 200 else 'N/A'
           })
       except Exception as e:
           results.append({
               'username': s['username'],
               'original_server': s['server_id'],
               'status': 'error',
               'new_server': 'N/A'
           })

   # 打印结果
   for r in results:
       print(f"{r['username']}: {r['original_server']} -> {r['new_server']} (Status: {r['status']})")
   ```

5. **统计 Session 丢失情况**
   ```python
   lost_sessions = [r for r in results if r['status'] != 200]
   print(f"丢失的 Session 数量: {len(lost_sessions)} / {len(sessions)}")
   ```

#### 📊 数据收集

| 用户 | 原服务器 | 宕机后状态码 | 新服务器 | Session 状态 |
|------|---------|------------|---------|-------------|
| user0 | [填写] | [填写] | [填写] | ✅保留 / ❌丢失 |
| user1 | [填写] | [填写] | [填写] | ✅保留 / ❌丢失 |
| user2 | [填写] | [填写] | [填写] | ✅保留 / ❌丢失 |
| user3 | [填写] | [填写] | [填写] | ✅保留 / ❌丢失 |
| user4 | [填写] | [填写] | [填写] | ✅保留 / ❌丢失 |

**统计**：
- 原本在宕机服务器上的 Session: _____ 个
- 丢失的 Session: _____ 个
- 保留的 Session: _____ 个

#### ✅ 预期结果
- 原本在宕机服务器上的 Session **全部丢失**（返回 401）
- 原本在其他服务器上的 Session **保持正常**（返回 200）
- Nginx 自动将请求路由到存活的服务器

#### 🔍 结论验证
- [ ] H1 验证：宕机服务器的 Session 丢失 ________
- [ ] H2 验证：自动路由到其他服务器 ________
- [ ] H3 验证：客户端需要重新登录 ________

**关键发现**：
- Session Affinity 的最大问题：________
- 对用户体验的影响：________
- 生产环境的风险：________

---

## 实验组二：Redis Session (集中式存储)

### 实验 2.1：跨服务器 Session 共享验证

#### 🎯 核心问题
**Q: 使用 Redis 存储 Session 后，不同服务器能否访问同一个 Session？**

#### 💡 实验假设
- **H1**: Session 存储在 Redis，所有服务器共享
- **H2**: 请求可以路由到任意服务器，都能正确获取 Session
- **H3**: Nginx 可以使用轮询（Round Robin）策略，不需要 `ip_hash`

#### 📋 实验设计

**前置条件**：
- Redis 已启动（`docker run -d -p 6379:6379 redis:alpine`）
- 3 个 Go 服务器连接同一个 Redis（端口 8091, 8092, 8093）
- Nginx 使用默认的轮询策略（不使用 `ip_hash`）

**测试步骤**：

1. **登录并观察 Redis 数据**
   ```python
   import requests
   import redis

   # 登录
   session = requests.Session()
   resp = session.post('http://localhost:8081/login',
                       json={'username': 'alice', 'password': '123456'})

   session_id = session.cookies.get('session_id')
   print(f"Session ID: {session_id}")

   # 查看 Redis 中的数据
   r = redis.Redis(host='localhost', port=6379, decode_responses=True)
   session_key = f"session:{session_id}"
   session_data = r.get(session_key)
   print(f"Redis Data: {session_data}")

   # 查看 TTL
   ttl = r.ttl(session_key)
   print(f"TTL: {ttl} 秒 ({ttl/60:.1f} 分钟)")
   ```

2. **直接访问不同服务器**
   ```python
   # 访问 Server-1
   resp1 = requests.get('http://localhost:8091/profile',
                        cookies={'session_id': session_id})
   print(f"Server-1: {resp1.json()}")

   # 访问 Server-2
   resp2 = requests.get('http://localhost:8092/profile',
                        cookies={'session_id': session_id})
   print(f"Server-2: {resp2.json()}")

   # 访问 Server-3
   resp3 = requests.get('http://localhost:8093/profile',
                        cookies={'session_id': session_id})
   print(f"Server-3: {resp3.json()}")
   ```

3. **通过 Nginx 轮询访问**
   ```python
   # 连续请求 20 次，观察请求分布
   servers = []
   for i in range(20):
       resp = session.get('http://localhost:8081/profile')
       server_id = resp.json()['server_id']
       servers.append(server_id)

   # 统计分布
   from collections import Counter
   print(f"服务器分布: {dict(Counter(servers))}")
   # 期望: 均匀分布，如 {'server-1': 7, 'server-2': 6, 'server-3': 7}
   ```

#### 📊 数据收集

| 测试项 | Server-1 | Server-2 | Server-3 | 结论 |
|-------|----------|----------|----------|------|
| 直接访问状态码 | [填写] | [填写] | [填写] | ✅/❌ |
| 返回的用户信息 | [填写] | [填写] | [填写] | ✅/❌ |
| Session 来源 | Redis | Redis | Redis | ✅/❌ |

**轮询分布**：
- Server-1 请求数: _____
- Server-2 请求数: _____
- Server-3 请求数: _____
- 分布是否均匀: ✅ / ❌

**Redis 数据**：
```json
{
  "session_id": "_____",
  "user_id": _____,
  "username": "_____",
  "ttl": _____ 秒
}
```

#### ✅ 预期结果
- 所有服务器都返回 200，用户信息一致
- Redis 中存在该 Session 数据
- 通过 Nginx 的请求均匀分布到三台服务器

#### 🔍 结论验证
- [ ] H1 验证：Session 存储在 Redis ________
- [ ] H2 验证：所有服务器能访问 ________
- [ ] H3 验证：轮询策略有效 ________

**对比 Sticky Session**：
- 优势：________
- 新增依赖：________

---

### 实验 2.2：Session 续期机制验证

#### 🎯 核心问题
**Q: 用户每次访问时，Session 的过期时间是否会自动延长？如何防止活跃用户被强制登出？**

#### 💡 实验假设
- **H1**: 每次访问 Session 时，Redis TTL 会被重置（续期）
- **H2**: 如果用户持续活跃，Session 永远不会过期
- **H3**: 用户停止访问后，Session 会在 TTL 到期后自动删除

#### 📋 实验设计

**测试步骤**：

1. **设置短 TTL 用于测试**
   ```go
   // 在 Go 代码中设置 TTL 为 30 秒（便于测试）
   const sessionTTL = 30 * time.Second

   func saveSession(sessionID string, session *Session) error {
       key := "session:" + sessionID
       data, _ := json.Marshal(session)
       return rdb.Set(ctx, key, data, sessionTTL).Err()
   }
   ```

2. **观察初始 TTL**
   ```python
   import redis
   import time

   r = redis.Redis(host='localhost', port=6379)

   # 登录
   session = requests.Session()
   session.post('http://localhost:8091/login', json={'username': 'alice'})
   session_id = session.cookies.get('session_id')

   key = f"session:{session_id}"

   # 记录初始 TTL
   ttl_initial = r.ttl(key)
   print(f"初始 TTL: {ttl_initial} 秒")
   ```

3. **等待后观察 TTL 减少**
   ```python
   # 等待 10 秒
   time.sleep(10)
   ttl_after_10s = r.ttl(key)
   print(f"10秒后 TTL: {ttl_after_10s} 秒 (减少了 {ttl_initial - ttl_after_10s} 秒)")
   ```

4. **访问后观察 TTL 重置**
   ```python
   # 发起一次请求（应触发续期）
   session.get('http://localhost:8091/profile')

   # 立即查看 TTL
   ttl_after_request = r.ttl(key)
   print(f"请求后 TTL: {ttl_after_request} 秒")

   # 验证是否接近初始值
   if abs(ttl_after_request - ttl_initial) <= 2:
       print("✅ TTL 已重置（续期成功）")
   else:
       print(f"❌ TTL 未重置（期望 ~{ttl_initial}，实际 {ttl_after_request}）")
   ```

5. **测试自动过期**
   ```python
   # 停止访问，等待 TTL 耗尽
   print(f"等待 {ttl_initial + 5} 秒让 Session 过期...")
   time.sleep(ttl_initial + 5)

   # 尝试访问
   resp = session.get('http://localhost:8091/profile')
   print(f"过期后访问: {resp.status_code}")
   # 期望: 401 Unauthorized

   # 验证 Redis 中已删除
   exists = r.exists(key)
   print(f"Redis 中是否存在: {'是' if exists else '否'}")
   ```

#### 📊 数据收集

| 时间点 | TTL (秒) | 操作 | 预期 TTL | 实际 TTL | 结论 |
|-------|---------|------|---------|---------|------|
| T0 | [填写] | 登录 | 30 | [填写] | ✅/❌ |
| T0+10s | [填写] | 等待 | ~20 | [填写] | ✅/❌ |
| T0+10s | [填写] | 访问(续期) | 30 | [填写] | ✅/❌ |
| T0+35s | [填写] | 等待过期 | 0 (删除) | [填写] | ✅/❌ |

**时间序列图**：
```
TTL (秒)
30 |●                    ●
25 | \                  /
20 |  \                /
15 |   \              /
10 |    \            /
 5 |     \          /
 0 |      ●________/
   +-----|-----|-----|-----> 时间
        10s   请求   35s
```

#### ✅ 预期结果
- 初始 TTL = 30 秒
- 10 秒后 TTL ≈ 20 秒
- 访问后 TTL 重置为 ≈ 30 秒
- 35 秒后 Session 被自动删除

#### 🔍 结论验证
- [ ] H1 验证：访问时 TTL 重置 ________
- [ ] H2 验证：持续活跃不过期 ________
- [ ] H3 验证：停止访问自动删除 ________

**实现要点**：
- Go 代码是否在每次 `getSession` 时调用 `EXPIRE`：________
- 生产环境建议 TTL：________

---

### 实验 2.3：Redis 宕机的影响测试

#### 🎯 核心问题
**Q: Redis 宕机后，所有服务器的 Session 是否全部失效？系统还能提供服务吗？**

#### 💡 实验假设
- **H1**: Redis 宕机后，所有 Session 立即不可用
- **H2**: 用户所有请求都会返回 401（未认证）
- **H3**: Redis 恢复后，旧的 Session 已丢失，用户需要重新登录

#### 📋 实验设计

**测试步骤**：

1. **建立多个用户 Session**
   ```python
   sessions = []
   for i in range(5):
       s = requests.Session()
       s.post('http://localhost:8091/login',
              json={'username': f'user{i}', 'password': '123456'})
       sessions.append(s)

   # 验证所有 Session 有效
   for i, s in enumerate(sessions):
       resp = s.get('http://localhost:8091/profile')
       print(f"User{i}: {resp.status_code}")
   ```

2. **停止 Redis**
   ```bash
   docker stop redis
   # 或
   docker kill redis
   ```

3. **立即测试 Session 可用性**
   ```python
   import time
   time.sleep(1)  # 等待 1 秒

   results = []
   for i, s in enumerate(sessions):
       try:
           resp = s.get('http://localhost:8091/profile', timeout=2)
           results.append({
               'user': f'user{i}',
               'status': resp.status_code,
               'error': None
           })
       except Exception as e:
           results.append({
               'user': f'user{i}',
               'status': 'timeout/error',
               'error': str(e)
           })

   for r in results:
       print(f"{r['user']}: {r['status']} ({r['error']})")
   ```

4. **重启 Redis**
   ```bash
   docker start redis
   # 或重新创建
   docker run -d --name redis -p 6379:6379 redis:alpine
   ```

5. **验证旧 Session 是否恢复**
   ```python
   time.sleep(2)  # 等待 Redis 启动

   for i, s in enumerate(sessions):
       resp = s.get('http://localhost:8091/profile')
       print(f"User{i} (Redis 恢复后): {resp.status_code}")

   # 尝试重新登录
   new_session = requests.Session()
   resp = new_session.post('http://localhost:8091/login',
                          json={'username': 'test', 'password': '123456'})
   print(f"新登录: {resp.status_code}")
   ```

#### 📊 数据收集

| 阶段 | User0 | User1 | User2 | User3 | User4 |
|------|-------|-------|-------|-------|-------|
| Redis 正常 | [状态码] | [状态码] | [状态码] | [状态码] | [状态码] |
| Redis 宕机 | [状态码] | [状态码] | [状态码] | [状态码] | [状态码] |
| Redis 恢复 | [状态码] | [状态码] | [状态码] | [状态码] | [状态码] |

**系统影响**：
- 宕机期间能否登录: ✅ / ❌
- 宕机期间能否访问 API: ✅ / ❌
- 恢复后旧 Session 是否有效: ✅ / ❌

#### ✅ 预期结果
- Redis 宕机后，所有请求返回 **500 或 401**（取决于错误处理）
- 用户无法登录（Session 无法存储）
- Redis 恢复后，旧 Session 丢失，需要重新登录

#### 🔍 结论验证
- [ ] H1 验证：Redis 宕机导致 Session 不可用 ________
- [ ] H2 验证：所有请求失败 ________
- [ ] H3 验证：恢复后旧 Session 丢失 ________

**关键风险**：
- Redis 成为**单点故障**
- 缓解方案：Redis Sentinel 或 Redis Cluster

---

## 实验组三：JWT Token (无状态认证)

### 实验 3.1：Token 无状态性验证

#### 🎯 核心问题
**Q: JWT Token 是否真的无状态？服务器能否在不访问数据库/Redis 的情况下验证 Token？**

#### 💡 实验假设
- **H1**: JWT Token 包含所有必要信息（user_id、过期时间）
- **H2**: 服务器只需验证签名，无需查询存储
- **H3**: 服务器重启不影响 Token 有效性

#### 📋 实验设计

**测试步骤**：

1. **登录并获取 Token**
   ```python
   import requests
   import jwt
   import json

   # 登录
   resp = requests.post('http://localhost:8101/login',
                        json={'username': 'alice', 'password': '123456'})

   token = resp.json()['token']
   print(f"Token: {token[:50]}...")
   ```

2. **解码 Token（不验证签名）**
   ```python
   # 使用 jwt.decode 查看 Payload
   payload = jwt.decode(token, options={"verify_signature": False})
   print(json.dumps(payload, indent=2))

   # 期望输出:
   # {
   #   "user_id": 1001,
   #   "username": "alice",
   #   "exp": 1698774032,
   #   "iat": 1698766832,
   #   "iss": "session-demo"
   # }
   ```

3. **验证服务器是否查询存储**
   ```python
   # 关闭 Redis（如果 JWT 真正无状态，应不受影响）
   # docker stop redis

   # 访问 API
   headers = {'Authorization': f'Bearer {token}'}
   resp = requests.get('http://localhost:8101/profile', headers=headers)
   print(f"Redis 关闭时: {resp.status_code}, {resp.json()}")

   # 期望: 200 OK（不依赖 Redis）
   ```

4. **重启服务器后测试 Token**
   ```bash
   # 重启 Go 服务器
   pkill -f "PORT=8101"
   PORT=8101 SERVER_ID=server-1 go run jwt-token/main.go &

   sleep 2  # 等待启动
   ```

   ```python
   # 使用旧 Token 访问
   resp = requests.get('http://localhost:8101/profile', headers=headers)
   print(f"服务器重启后: {resp.status_code}, {resp.json()}")

   # 期望: 200 OK（Token 仍然有效）
   ```

5. **直接访问不同服务器**
   ```python
   # 使用同一个 Token 访问三台服务器
   for port in [8101, 8102, 8103]:
       resp = requests.get(f'http://localhost:{port}/profile',
                          headers=headers)
       print(f"Server {port}: {resp.json()}")

   # 期望: 所有服务器都返回相同的用户信息
   ```

#### 📊 数据收集

| 测试场景 | 是否依赖存储 | 状态码 | 结论 |
|---------|------------|--------|------|
| 正常访问 | ❌ | [填写] | ✅/❌ |
| Redis 关闭 | ❌ | [填写] | ✅/❌ |
| 服务器重启 | ❌ | [填写] | ✅/❌ |
| 跨服务器访问 | ❌ | [填写] | ✅/❌ |

**Token Payload**：
```json
{
  "user_id": _____,
  "username": "_____",
  "exp": _____ (Unix 时间戳),
  "iat": _____ (Unix 时间戳)
}
```

#### ✅ 预期结果
- Token Payload 包含用户信息
- Redis 关闭不影响 Token 验证
- 服务器重启后 Token 仍然有效
- 所有服务器都能验证同一个 Token

#### 🔍 结论验证
- [ ] H1 验证：Token 包含完整信息 ________
- [ ] H2 验证：无需查询存储 ________
- [ ] H3 验证：服务器重启不影响 ________

**对比 Redis Session**：
- 优势：________
- 劣势：________

---

### 实验 3.2：Token 过期机制测试

#### 🎯 核心问题
**Q: Token 过期后是否会自动失效？服务器如何处理过期的 Token？**

#### 💡 实验假设
- **H1**: Token 包含过期时间（`exp` 字段）
- **H2**: 服务器验证时会检查当前时间是否超过 `exp`
- **H3**: 过期的 Token 无法通过验证，返回 401

#### 📋 实验设计

**测试步骤**：

1. **生成短期 Token（5秒过期）**
   ```go
   // 在 Go 代码中添加测试接口
   http.HandleFunc("/login-short", func(w http.ResponseWriter, r *http.Request) {
       claims := &Claims{
           UserID:   1001,
           Username: "test",
           RegisteredClaims: jwt.RegisteredClaims{
               ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Second)),
               IssuedAt:  jwt.NewNumericDate(time.Now()),
           },
       }
       token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secretKey)
       json.NewEncoder(w).Encode(map[string]string{"token": token})
   })
   ```

   ```python
   import time
   import jwt

   # 获取短期 Token
   resp = requests.post('http://localhost:8101/login-short')
   token = resp.json()['token']

   # 解码查看过期时间
   payload = jwt.decode(token, options={"verify_signature": False})
   exp_timestamp = payload['exp']
   exp_time = time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(exp_timestamp))
   print(f"过期时间: {exp_time}")
   ```

2. **在有效期内访问**
   ```python
   headers = {'Authorization': f'Bearer {token}'}
   resp = requests.get('http://localhost:8101/profile', headers=headers)
   print(f"有效期内: {resp.status_code}, {resp.json()}")
   ```

3. **等待 Token 过期**
   ```python
   print("等待 6 秒让 Token 过期...")
   time.sleep(6)

   # 尝试使用过期 Token
   resp = requests.get('http://localhost:8101/profile', headers=headers)
   print(f"过期后: {resp.status_code}")

   if resp.status_code == 401:
       print(f"错误信息: {resp.json()}")
   ```

4. **测试时间篡改（高级）**
   ```python
   # 手动修改 Token 的 exp 字段（会导致签名验证失败）
   import base64

   parts = token.split('.')
   payload_base64 = parts[1]

   # 添加 padding
   padding = len(payload_base64) % 4
   if padding:
       payload_base64 += '=' * (4 - padding)

   payload_json = base64.urlsafe_b64decode(payload_base64)
   payload_dict = json.loads(payload_json)

   # 篡改过期时间为未来
   payload_dict['exp'] = int(time.time()) + 3600

   # 重新编码
   new_payload = base64.urlsafe_b64encode(json.dumps(payload_dict).encode()).decode().rstrip('=')
   tampered_token = f"{parts[0]}.{new_payload}.{parts[2]}"

   # 尝试使用篡改的 Token
   headers_tampered = {'Authorization': f'Bearer {tampered_token}'}
   resp = requests.get('http://localhost:8101/profile', headers=headers_tampered)
   print(f"篡改 Token: {resp.status_code}")
   # 期望: 401 (签名验证失败)
   ```

#### 📊 数据收集

| 时间点 | Token 状态 | 状态码 | 响应 | 结论 |
|-------|-----------|--------|------|------|
| T0 (生成) | 有效 | [填写] | [填写] | ✅/❌ |
| T0+3s | 有效 | [填写] | [填写] | ✅/❌ |
| T0+6s | 过期 | [填写] | [填写] | ✅/❌ |
| 篡改 exp | 无效签名 | [填写] | [填写] | ✅/❌ |

**时间轴**：
```
时间 ->  T0      T0+3s    T0+5s    T0+6s
Token:   [生成]  [有效]   [到期]   [过期]
访问:      ✅      ✅       -        ❌
```

#### ✅ 预期结果
- 有效期内（< 5秒）: 返回 200
- 过期后（> 5秒）: 返回 401, {"error": "token expired"}
- 篡改 Token: 返回 401, {"error": "invalid signature"}

#### 🔍 结论验证
- [ ] H1 验证：Token 包含过期时间 ________
- [ ] H2 验证：服务器检查过期时间 ________
- [ ] H3 验证：过期 Token 自动失效 ________

**安全性分析**：
- JWT 签名是否防止篡改: ✅ / ❌
- 时间同步的重要性: ________

---

### 实验 3.3：黑名单实现"登出"功能

#### 🎯 核心问题
**Q: JWT Token 本身无法撤销，如何实现"登出"功能？黑名单方案是否有效？**

#### 💡 实验假设
- **H1**: 登出时将 Token 加入 Redis 黑名单
- **H2**: 验证 Token 时先检查黑名单
- **H3**: 黑名单 Key 的 TTL 等于 Token 的剩余有效期

#### 📋 实验设计

**前置条件**：
- Redis 已启动
- Go 服务器实现黑名单检查逻辑

**测试步骤**：

1. **登录并获取 Token**
   ```python
   import redis

   resp = requests.post('http://localhost:8101/login',
                        json={'username': 'alice', 'password': '123456'})
   token = resp.json()['token']

   # 验证 Token 有效
   headers = {'Authorization': f'Bearer {token}'}
   resp = requests.get('http://localhost:8101/profile', headers=headers)
   print(f"登录后: {resp.status_code}")
   ```

2. **登出并观察黑名单**
   ```python
   r = redis.Redis(host='localhost', port=6379, decode_responses=True)

   # 登出
   resp = requests.post('http://localhost:8101/logout', headers=headers)
   print(f"登出: {resp.json()}")

   # 查看 Redis 黑名单
   blacklist_key = f"blacklist:{token}"
   exists = r.exists(blacklist_key)
   print(f"黑名单中存在: {exists}")

   if exists:
       ttl = r.ttl(blacklist_key)
       print(f"黑名单 TTL: {ttl} 秒")
   ```

3. **尝试使用已登出的 Token**
   ```python
   resp = requests.get('http://localhost:8101/profile', headers=headers)
   print(f"登出后访问: {resp.status_code}")
   # 期望: 401

   if resp.status_code == 401:
       print(f"错误信息: {resp.json()}")
       # 期望: {"error": "token revoked"}
   ```

4. **测试黑名单自动过期**
   ```python
   # 登录获取短期 Token (30秒过期)
   # 立即登出
   # 等待 35 秒
   # 验证黑名单 Key 是否已被 Redis 自动删除

   # ... (代码同上)

   import time
   time.sleep(35)

   exists_after_ttl = r.exists(blacklist_key)
   print(f"35秒后黑名单存在: {exists_after_ttl}")
   # 期望: False
   ```

5. **性能影响测试**
   ```python
   # 对比启用/禁用黑名单的性能
   import time

   # 测试 1: 无黑名单检查（纯 JWT 验证）
   start = time.time()
   for _ in range(100):
       requests.get('http://localhost:8101/profile', headers=headers)
   duration_no_blacklist = time.time() - start

   # 测试 2: 有黑名单检查（JWT + Redis 查询）
   # (需要在代码中切换开关)
   start = time.time()
   for _ in range(100):
       requests.get('http://localhost:8101/profile-with-blacklist', headers=headers)
   duration_with_blacklist = time.time() - start

   print(f"无黑名单: {duration_no_blacklist:.2f}s")
   print(f"有黑名单: {duration_with_blacklist:.2f}s")
   print(f"性能损失: {(duration_with_blacklist - duration_no_blacklist) / duration_no_blacklist * 100:.1f}%")
   ```

#### 📊 数据收集

| 测试项 | 结果 | 预期 | 结论 |
|-------|------|------|------|
| 登出后黑名单存在 | [填写] | True | ✅/❌ |
| 登出后访问状态码 | [填写] | 401 | ✅/❌ |
| 黑名单 TTL | [填写秒] | ~7200s | ✅/❌ |
| 35秒后黑名单自动删除 | [填写] | False | ✅/❌ |

**性能对比**：
- 纯 JWT 验证: _____ 秒 (100次请求)
- JWT + 黑名单: _____ 秒 (100次请求)
- 性能损失: _____ %

#### ✅ 预期结果
- 登出后 Token 加入黑名单
- 黑名单中的 Token 返回 401
- 黑名单 TTL = Token 剩余有效期
- Token 过期后黑名单自动删除
- 性能损失约 10-20%（Redis 查询开销）

#### 🔍 结论验证
- [ ] H1 验证：登出时加入黑名单 ________
- [ ] H2 验证：黑名单检查有效 ________
- [ ] H3 验证：TTL 设置正确 ________

**黑名单方案评估**：
- 优点：________
- 缺点：________
- 是否违背"无状态"原则：________

---

## 实验组四：综合性能对比

### 实验 4.1：延迟对比测试

#### 🎯 核心问题
**Q: 三种方案的响应延迟有多大差异？哪种方案最快？**

#### 💡 实验假设
- **H1**: Sticky Session 最快（本地内存访问）
- **H2**: JWT Token 次之（CPU 计算签名）
- **H3**: Redis Session 最慢（网络 I/O）

#### 📋 实验设计

**测试方法**：单客户端串行请求，测量每次请求的延迟

**测试代码**：

```python
import requests
import time
import statistics

def measure_latency(url, headers=None, cookies=None, iterations=100):
    latencies = []

    for _ in range(iterations):
        start = time.perf_counter()
        resp = requests.get(url, headers=headers, cookies=cookies)
        latency = (time.perf_counter() - start) * 1000  # 毫秒

        if resp.status_code == 200:
            latencies.append(latency)

    return {
        'min': min(latencies),
        'max': max(latencies),
        'mean': statistics.mean(latencies),
        'median': statistics.median(latencies),
        'p95': statistics.quantiles(latencies, n=20)[18],
        'p99': statistics.quantiles(latencies, n=100)[98]
    }

# 测试 Sticky Session
sticky_session = requests.Session()
sticky_session.post('http://localhost:8080/login', json={'username': 'alice'})
sticky_result = measure_latency('http://localhost:8080/profile', cookies=sticky_session.cookies)

# 测试 Redis Session
redis_session = requests.Session()
redis_session.post('http://localhost:8081/login', json={'username': 'alice'})
redis_result = measure_latency('http://localhost:8081/profile', cookies=redis_session.cookies)

# 测试 JWT Token
jwt_resp = requests.post('http://localhost:8101/login', json={'username': 'alice'})
jwt_token = jwt_resp.json()['token']
jwt_headers = {'Authorization': f'Bearer {jwt_token}'}
jwt_result = measure_latency('http://localhost:8101/profile', headers=jwt_headers)

# 打印结果
print("延迟对比 (单位: 毫秒)")
print(f"{'指标':<10} {'Sticky':<10} {'Redis':<10} {'JWT':<10}")
print(f"{'平均':<10} {sticky_result['mean']:<10.2f} {redis_result['mean']:<10.2f} {jwt_result['mean']:<10.2f}")
print(f"{'中位数':<10} {sticky_result['median']:<10.2f} {redis_result['median']:<10.2f} {jwt_result['median']:<10.2f}")
print(f"{'P95':<10} {sticky_result['p95']:<10.2f} {redis_result['p95']:<10.2f} {jwt_result['p95']:<10.2f}")
print(f"{'P99':<10} {sticky_result['p99']:<10.2f} {redis_result['p99']:<10.2f} {jwt_result['p99']:<10.2f}")
```

#### 📊 数据收集

| 指标 | Sticky Session | Redis Session | JWT Token |
|------|---------------|---------------|-----------|
| 平均延迟 (ms) | [填写] | [填写] | [填写] |
| 中位数 (P50) | [填写] | [填写] | [填写] |
| P95 延迟 | [填写] | [填写] | [填写] |
| P99 延迟 | [填写] | [填写] | [填写] |
| 最小延迟 | [填写] | [填写] | [填写] |
| 最大延迟 | [填写] | [填写] | [填写] |

**延迟分布图**：
```
延迟 (ms)
  5 |           ●
  4 |     ●     |
  3 |     |     |     ●
  2 | ●   |     |     |
  1 | |   |     |     |
  0 +-----------------------
    Sticky  Redis  JWT
```

#### ✅ 预期结果
- Sticky Session: ~0.1-0.5ms
- JWT Token: ~0.3-1ms
- Redis Session: ~1-3ms

#### 🔍 结论验证
- [ ] H1 验证：Sticky Session 最快 ________
- [ ] H2 验证：JWT 次之 ________
- [ ] H3 验证：Redis 最慢 ________

**延迟来源分析**：
- Sticky Session: ________
- Redis Session: ________
- JWT Token: ________

---

### 实验 4.2：吞吐量对比测试

#### 🎯 核心问题
**Q: 在高并发场景下，哪种方案能处理更多的请求？QPS 差异有多大？**

#### 💡 实验假设
- **H1**: Sticky Session 的 QPS 最高
- **H2**: Redis Session 受 Redis 性能限制
- **H3**: JWT Token 受 CPU 签名验证限制

#### 📋 实验设计

**测试方法**：使用 Apache Bench 并发压测

**测试步骤**：

1. **准备测试 Token/Cookie**
   ```python
   # 获取各方案的认证凭据

   # Sticky Session Cookie
   sticky_resp = requests.post('http://localhost:8080/login', json={'username': 'test'})
   sticky_cookie = sticky_resp.cookies.get('session_id')

   # Redis Session Cookie
   redis_resp = requests.post('http://localhost:8081/login', json={'username': 'test'})
   redis_cookie = redis_resp.cookies.get('session_id')

   # JWT Token
   jwt_resp = requests.post('http://localhost:8101/login', json={'username': 'test'})
   jwt_token = jwt_resp.json()['token']

   print(f"Sticky Cookie: {sticky_cookie}")
   print(f"Redis Cookie: {redis_cookie}")
   print(f"JWT Token: {jwt_token[:50]}...")
   ```

2. **使用 Apache Bench 压测**
   ```bash
   # 测试 Sticky Session
   ab -n 10000 -c 100 \
      -C "session_id=${sticky_cookie}" \
      http://localhost:8080/profile

   # 测试 Redis Session
   ab -n 10000 -c 100 \
      -C "session_id=${redis_cookie}" \
      http://localhost:8081/profile

   # 测试 JWT Token
   ab -n 10000 -c 100 \
      -H "Authorization: Bearer ${jwt_token}" \
      http://localhost:8101/profile
   ```

3. **或使用 Python 并发测试**
   ```python
   from concurrent.futures import ThreadPoolExecutor
   import time

   def test_throughput(url, duration=10, concurrency=100, **kwargs):
       request_count = 0
       errors = 0

       def make_request():
           nonlocal request_count, errors
           try:
               resp = requests.get(url, timeout=5, **kwargs)
               if resp.status_code == 200:
                   request_count += 1
               else:
                   errors += 1
           except Exception:
               errors += 1

       start = time.time()
       end_time = start + duration

       with ThreadPoolExecutor(max_workers=concurrency) as executor:
           while time.time() < end_time:
               executor.submit(make_request)

       elapsed = time.time() - start
       qps = request_count / elapsed

       return {
           'total_requests': request_count,
           'errors': errors,
           'qps': qps,
           'duration': elapsed
       }

   # 测试三种方案
   sticky_tp = test_throughput('http://localhost:8080/profile',
                                cookies=sticky_session.cookies)
   redis_tp = test_throughput('http://localhost:8081/profile',
                               cookies=redis_session.cookies)
   jwt_tp = test_throughput('http://localhost:8101/profile',
                            headers=jwt_headers)

   print(f"Sticky Session: {sticky_tp['qps']:.0f} QPS ({sticky_tp['errors']} errors)")
   print(f"Redis Session:  {redis_tp['qps']:.0f} QPS ({redis_tp['errors']} errors)")
   print(f"JWT Token:      {jwt_tp['qps']:.0f} QPS ({jwt_tp['errors']} errors)")
   ```

#### 📊 数据收集

**Apache Bench 结果**：

| 指标 | Sticky Session | Redis Session | JWT Token |
|------|---------------|---------------|-----------|
| 总请求数 | 10,000 | 10,000 | 10,000 |
| 并发数 | 100 | 100 | 100 |
| 总耗时 (s) | [填写] | [填写] | [填写] |
| QPS | [填写] | [填写] | [填写] |
| 平均延迟 (ms) | [填写] | [填写] | [填写] |
| 失败请求 | [填写] | [填写] | [填写] |

**QPS 对比图**：
```
QPS
50K |  ●
45K |  |     ●
40K |  |     |
30K |  |     |     ●
20K |  |     |     |
10K |  |     |     |
  0 +-----------------
   Sticky Redis JWT
```

#### ✅ 预期结果
- Sticky Session: 40,000 - 50,000 QPS
- JWT Token: 35,000 - 45,000 QPS
- Redis Session: 25,000 - 35,000 QPS

#### 🔍 结论验证
- [ ] H1 验证：Sticky QPS 最高 ________
- [ ] H2 验证：Redis 受限于网络 ________
- [ ] H3 验证：JWT 受限于 CPU ________

**瓶颈分析**：
- Sticky Session 瓶颈: ________
- Redis Session 瓶颈: ________
- JWT Token 瓶颈: ________

---

### 实验 4.3：资源消耗对比

#### 🎯 核心问题
**Q: 三种方案分别消耗多少内存和 CPU？哪种方案更节省资源？**

#### 💡 实验假设
- **H1**: Sticky Session 消耗服务器内存（存储 Session）
- **H2**: Redis Session 消耗 Redis 内存
- **H3**: JWT Token 消耗 CPU（签名计算）

#### 📋 实验设计

**测试步骤**：

1. **创建大量 Session**
   ```python
   # 创建 10,000 个 Session
   for i in range(10000):
       if i % 1000 == 0:
           print(f"创建 {i} 个 Session...")

       # Sticky Session
       requests.post('http://localhost:8081/login',
                    json={'username': f'user{i}', 'password': '123'})
   ```

2. **测量服务器内存占用**
   ```bash
   # 使用 ps 查看 Go 进程内存
   ps aux | grep "PORT=8081" | awk '{print $6/1024 " MB"}'

   # 或使用 Go pprof
   curl http://localhost:8081/debug/pprof/heap > heap.prof
   go tool pprof -top heap.prof
   ```

3. **测量 Redis 内存占用**
   ```bash
   redis-cli info memory | grep used_memory_human

   # 查看 Session 数量
   redis-cli DBSIZE
   ```

4. **测量 CPU 占用**
   ```bash
   # 压测时观察 CPU
   top -p <go-pid>

   # 或使用 Go pprof
   curl http://localhost:8101/debug/pprof/profile?seconds=30 > cpu.prof
   go tool pprof -top cpu.prof
   ```

#### 📊 数据收集

**10,000 个活跃 Session 的资源消耗**：

| 资源 | Sticky Session | Redis Session | JWT Token |
|------|---------------|---------------|-----------|
| 服务器内存 | [填写] MB | [填写] MB | ~0 MB |
| Redis 内存 | 0 MB | [填写] MB | 0 MB (黑名单除外) |
| CPU 使用率 (空闲) | [填写]% | [填写]% | [填写]% |
| CPU 使用率 (压测) | [填写]% | [填写]% | [填写]% |

**单个 Session 的内存占用**：
```
Sticky Session: _____ bytes/session
Redis Session:  _____ bytes/session
JWT Token:      0 bytes (无状态)
```

#### ✅ 预期结果
- Sticky Session: 每个 Session ~500 bytes，10K Session ≈ 5MB
- Redis Session: 每个 Session ~300 bytes，10K Session ≈ 3MB
- JWT Token: 0 字节（服务器无存储）

#### 🔍 结论验证
- [ ] H1 验证：Sticky 消耗服务器内存 ________
- [ ] H2 验证：Redis 消耗 Redis 内存 ________
- [ ] H3 验证：JWT 消耗 CPU ________

**成本分析**：
- 100 万用户的内存成本: ________
- 推荐方案: ________

---

## 实验总结与报告

### 实验数据汇总表

#### 功能对比

| 功能特性 | Sticky Session | Redis Session | JWT Token |
|---------|---------------|---------------|-----------|
| 跨服务器共享 | ❌ | ✅ | ✅ |
| 服务器宕机恢复 | ❌ | ✅ | ✅ |
| 主动登出 | ✅ | ✅ | ⚠️ 需黑名单 |
| 水平扩展 | ⚠️ 困难 | ✅ | ✅ |
| 服务器重启影响 | ❌ Session丢失 | ✅ 无影响 | ✅ 无影响 |
| 依赖外部服务 | ❌ | ✅ Redis | ❌ |

#### 性能对比

| 性能指标 | Sticky Session | Redis Session | JWT Token |
|---------|---------------|---------------|-----------|
| P50 延迟 (ms) | [填写] | [填写] | [填写] |
| P99 延迟 (ms) | [填写] | [填写] | [填写] |
| QPS | [填写] | [填写] | [填写] |
| 内存占用 (10K用户) | [填写] | [填写] | [填写] |
| CPU 使用率 | [填写]% | [填写]% | [填写]% |

#### 适用场景

| 场景 | 推荐方案 | 理由 |
|------|---------|------|
| 小型应用 (< 3台服务器) | Sticky Session | 简单，性能高 |
| 电商平台 | Redis Session | 需要强制登出、修改购物车 |
| 移动端 API | JWT Token | 无状态，适合分布式 |
| 单页应用 (SPA) | JWT Token | 跨域友好 |
| 微服务架构 | JWT Token | 服务间认证 |
| 高安全要求系统 | Redis Session | 需要实时撤销 Session |

---

## 📝 实验报告模板

完成所有实验后，填写以下报告：

### 1. 实验环境

```
操作系统: _____
Go 版本: _____
Python 版本: _____
Redis 版本: _____
Nginx 版本: _____
```

### 2. 关键发现

**实验 1.1 发现**：
- Sticky Session 路由一致性: ________
- 负载均衡效果: ________

**实验 1.3 发现**：
- 服务器宕机影响: ________
- Session 丢失比例: ________

**实验 2.1 发现**：
- 跨服务器共享是否成功: ________
- 性能开销: ________

**实验 3.1 发现**：
- JWT 是否真正无状态: ________
- Token 大小: ________ bytes

**实验 4.1 发现**：
- 延迟排序: ________ < ________ < ________
- 延迟差异原因: ________

### 3. 遇到的问题与解决

| 问题 | 解决方案 |
|------|---------|
| [填写问题] | [填写解决方法] |
| [填写问题] | [填写解决方法] |

### 4. 性能数据对比

(粘贴上方的性能对比表)

### 5. 方案选择建议

基于实验结果，针对不同场景的建议：

- **小型项目**：________，因为 ________
- **中大型项目**：________，因为 ________
- **API 服务**：________，因为 ________

### 6. 实验总结

**最大收获**：
1. ________
2. ________
3. ________

**理论与实践的差异**：
- ________

**未来探索方向**：
- ________

### 7. 实验时长统计

- 准备工作: _____ 小时
- 实验组一 (Sticky Session): _____ 小时
- 实验组二 (Redis Session): _____ 小时
- 实验组三 (JWT Token): _____ 小时
- 实验组四 (性能对比): _____ 小时
- 报告撰写: _____ 小时
- **总计**: _____ 小时

---

## ✅ 实验完成检查清单

### 实验组一：Sticky Session
- [ ] 实验 1.1: IP Hash 路由一致性
- [ ] 实验 1.2: Session 数据隔离性
- [ ] 实验 1.3: 服务器宕机测试

### 实验组二：Redis Session
- [ ] 实验 2.1: 跨服务器共享验证
- [ ] 实验 2.2: Session 续期机制
- [ ] 实验 2.3: Redis 宕机影响

### 实验组三：JWT Token
- [ ] 实验 3.1: 无状态性验证
- [ ] 实验 3.2: Token 过期机制
- [ ] 实验 3.3: 黑名单实现登出

### 实验组四：性能对比
- [ ] 实验 4.1: 延迟对比测试
- [ ] 实验 4.2: 吞吐量对比测试
- [ ] 实验 4.3: 资源消耗对比

### 文档输出
- [ ] 所有数据表格已填写
- [ ] 实验报告已完成
- [ ] 代码已提交到 Git
- [ ] 笔记已更新

---

## 🎓 学习建议

### 实验顺序
1. 先完成**实验组一**（最简单，建立基础理解）
2. 再完成**实验组二**（理解集中式存储）
3. 然后完成**实验组三**（理解无状态）
4. 最后完成**实验组四**（综合对比）

### 时间分配建议
- **第 1 天**：准备环境 + 实验组一 (3-4 小时)
- **第 2 天**：实验组二 (3-4 小时)
- **第 3 天**：实验组三 (3-4 小时)
- **第 4 天**：实验组四 + 报告 (3-4 小时)
- **总计**：12-16 小时

### 关键提示
- ✅ 每个实验都要**亲自运行代码**，不要只看结果
- ✅ **记录实际数据**，不要凭空猜测
- ✅ **对比预期与实际**，分析差异原因
- ✅ **遇到问题先调试**，理解背后原理
- ✅ **写实验报告时**，总结自己的理解和发现

---

**实验愉快！通过这些实验，你将深刻理解三种会话管理方案的本质区别！** 🚀

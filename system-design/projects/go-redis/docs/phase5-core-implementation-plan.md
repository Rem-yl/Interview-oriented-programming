# 阶段五核心功能实现计划

## 🎯 目标

**快速让 go-redis 具备核心功能可用，为后续架构设计学习打好基础**

- 实现 6 个核心命令（INCR/DECR/INCRBY/DECRBY/MGET/MSET）
- 实现过期机制（EXPIRE/TTL + 懒删除 + 定期删除）
- 预计总耗时：**3-4 小时**

---

## 📋 Part 1: 核心命令实现（2-3 小时）

### 命令清单

| 序号 | 命令 | 预计耗时 | 核心功能 |
|------|------|---------|---------|
| 1 | **INCR** | 30分钟 | 原子自增，最常用 |
| 2 | **DECR** | 10分钟 | 原子自减 |
| 3 | **INCRBY** | 10分钟 | 增加指定值 |
| 4 | **DECRBY** | 10分钟 | 减少指定值 |
| 5 | **MSET** | 20分钟 | 批量设置 |
| 6 | **MGET** | 30分钟 | 批量获取 |

---

### 1. INCR 命令

**文件**：`handler/incr.go`

**核心逻辑**：
```go
type IncrHandler struct {
    db *store.Store
}

func (h *IncrHandler) Handle(args []protocol.Value) *protocol.Value {
    // 1. 参数验证：必须是 1 个参数
    // 2. 获取 key 的值
    //    - 不存在：初始化为 0
    //    - 存在：尝试解析为 int64
    //      - string "123" → 123（使用 strconv.ParseInt）
    //      - int64 → 直接使用
    //      - 其他 → 返回错误
    // 3. 自增 1
    // 4. 保存回 Store
    // 5. 返回新值（Integer 类型）
}
```

**关键点**：
- 类型处理：支持 `string` 和 `int64` 两种类型
- 错误信息：`"ERR value is not an integer or out of range"`
- 原子性：Store 的 RWMutex 已经保证（Get + Set 在锁内）

**测试用例**：
```go
// handler/router_test.go 中添加
func TestIncrHandler(t *testing.T) {
    // TC1: 键不存在 → 返回 1
    // TC2: 键存在，值为 "10" → 返回 11
    // TC3: 键存在，值为 "abc" → 返回错误
}
```

---

### 2. DECR 命令

**文件**：`handler/decr.go`

**核心逻辑**：
```go
// 几乎和 INCR 一样，只是改为 -1
func (h *DecrHandler) Handle(args []protocol.Value) *protocol.Value {
    // 复用 INCR 的逻辑，改为 currentNum - 1
}
```

**提示**：可以提取公共函数 `incrementBy(key string, delta int64)` 供 INCR/DECR 共用。

---

### 3. INCRBY 命令

**文件**：`handler/incrby.go`

**核心逻辑**：
```go
func (h *IncrbyHandler) Handle(args []protocol.Value) *protocol.Value {
    // 1. 参数验证：2 个参数（key, increment）
    // 2. 解析 increment 为 int64
    //    args[1].Str → strconv.ParseInt()
    // 3. 复用 INCR 逻辑，增量为 increment
    // 4. 返回新值
}
```

**示例**：
```bash
INCRBY counter 5   → 增加 5
INCRBY counter -3  → 减少 3（允许负数）
```

---

### 4. DECRBY 命令

**文件**：`handler/decrby.go`

**核心逻辑**：
```go
func (h *DecrbyHandler) Handle(args []protocol.Value) *protocol.Value {
    // 和 INCRBY 一样，只是增量改为负数
    // 或者直接调用 INCRBY(-delta)
}
```

---

### 5. MSET 命令

**文件**：`handler/mset.go`

**核心逻辑**：
```go
func (h *MsetHandler) Handle(args []protocol.Value) *protocol.Value {
    // 1. 参数验证：
    //    - 至少 2 个参数
    //    - 必须是偶数个（key-value 对）
    // 2. 遍历参数，每 2 个一组
    //    for i := 0; i < len(args); i += 2 {
    //        key := args[i].Str
    //        value := args[i+1].Str
    //        h.db.Set(key, value)
    //    }
    // 3. 返回 SimpleString("OK")
}
```

**示例**：
```bash
MSET k1 v1 k2 v2 k3 v3
→ 设置 3 个键值对
```

---

### 6. MGET 命令

**文件**：`handler/mget.go`

**核心逻辑**：
```go
func (h *MgetHandler) Handle(args []protocol.Value) *protocol.Value {
    // 1. 参数验证：至少 1 个 key
    // 2. 创建结果数组
    //    results := make([]protocol.Value, len(args))
    // 3. 遍历每个 key
    //    for i, keyVal := range args {
    //        value, exists := h.db.Get(keyVal.Str)
    //        if !exists {
    //            results[i] = protocol.NullBulkString()  // NULL
    //        } else {
    //            // 只返回字符串类型，其他类型返回 NULL
    //            if strVal, ok := value.(string); ok {
    //                results[i] = protocol.BulkString(strVal)
    //            } else {
    //                results[i] = protocol.NullBulkString()
    //            }
    //        }
    //    }
    // 4. 返回数组
    //    return &protocol.Value{
    //        Type:  protocol.ArrayType,
    //        Array: results,
    //    }
}
```

**关键点**：
- 返回类型是 **Array**
- 不存在的 key 返回 **NULL**（`IsNull: true`）

---

### 注册命令

**文件**：`handler/router.go`

在 `registerDefaultHandlers()` 中添加：
```go
func (r *Router) registerDefaultHandlers() {
    // ... 现有命令

    // 数值操作
    r.Register("INCR", NewIncrHandler(r.db))
    r.Register("DECR", NewDecrHandler(r.db))
    r.Register("INCRBY", NewIncrbyHandler(r.db))
    r.Register("DECRBY", NewDecrbyHandler(r.db))

    // 批量操作
    r.Register("MGET", NewMgetHandler(r.db))
    r.Register("MSET", NewMsetHandler(r.db))
}
```

---

### 验证方法

#### 1. 单元测试

在 `handler/router_test.go` 中添加测试：

```go
func TestIncrHandler(t *testing.T) {
    s := store.NewStore()
    r := NewRouter(s)

    // TC1: 键不存在
    resp := executeCommand(r, "*2\r\n$4\r\nINCR\r\n$7\r\ncounter\r\n")
    assert.Equal(t, ":1\r\n", resp)

    // TC2: 键存在
    s.Set("counter", int64(10))
    resp = executeCommand(r, "*2\r\n$4\r\nINCR\r\n$7\r\ncounter\r\n")
    assert.Equal(t, ":11\r\n", resp)

    // TC3: 非整数
    s.Set("mykey", "hello")
    resp = executeCommand(r, "*2\r\n$4\r\nINCR\r\n$5\r\nmykey\r\n")
    assert.Contains(t, resp, "ERR")
}

func TestMgetMsetHandler(t *testing.T) {
    s := store.NewStore()
    r := NewRouter(s)

    // MSET
    resp := executeCommand(r, "*7\r\n$4\r\nMSET\r\n$2\r\nk1\r\n$2\r\nv1\r\n$2\r\nk2\r\n$2\r\nv2\r\n$2\r\nk3\r\n$2\r\nv3\r\n")
    assert.Equal(t, "+OK\r\n", resp)

    // MGET
    resp = executeCommand(r, "*5\r\n$4\r\nMGET\r\n$2\r\nk1\r\n$2\r\nk2\r\n$2\r\nk3\r\n$2\r\nk4\r\n")
    // 验证返回数组，包含 v1, v2, v3, nil
}
```

#### 2. 集成测试（redis-cli）

```bash
# 启动服务
go run main.go

# 另一个终端
redis-cli -p 16379

# 测试 INCR
127.0.0.1:16379> SET counter 0
OK
127.0.0.1:16379> INCR counter
(integer) 1
127.0.0.1:16379> INCRBY counter 10
(integer) 11

# 测试 MSET/MGET
127.0.0.1:16379> MSET k1 v1 k2 v2 k3 v3
OK
127.0.0.1:16379> MGET k1 k2 k3 k4
1) "v1"
2) "v2"
3) "v3"
4) (nil)
```

---

## 📋 Part 2: 过期机制实现（1-2 小时）

### 目标

实现 Redis 的过期机制，让缓存可以自动过期。

---

### 实现步骤

#### 1. 扩展 Store 数据结构

**文件**：`store/store.go`

```go
type Store struct {
    mu      sync.RWMutex
    data    map[string]interface{}
    expires map[string]time.Time  // 新增：过期时间映射
    stopCh  chan struct{}          // 新增：停止清理信号
}

func NewStore() *Store {
    s := &Store{
        data:    make(map[string]interface{}),
        expires: make(map[string]time.Time),
        stopCh:  make(chan struct{}),
    }
    go s.cleanupExpiredKeys()  // 启动后台清理
    return s
}
```

#### 2. 添加过期相关方法

```go
// SetWithExpire 设置带过期时间的键值
func (s *Store) SetWithExpire(key string, value interface{}, expireAt time.Time) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value
    s.expires[key] = expireAt
}

// Expire 设置键的过期时间（秒）
func (s *Store) Expire(key string, seconds int) bool {
    s.mu.Lock()
    defer s.mu.Unlock()

    // 检查键是否存在
    if _, exists := s.data[key]; !exists {
        return false
    }

    s.expires[key] = time.Now().Add(time.Duration(seconds) * time.Second)
    return true
}

// TTL 获取键的剩余生存时间（秒）
func (s *Store) TTL(key string) int64 {
    s.mu.RLock()
    defer s.mu.RUnlock()

    // 键不存在
    if _, exists := s.data[key]; !exists {
        return -2
    }

    // 没有过期时间
    expireAt, hasExpire := s.expires[key]
    if !hasExpire {
        return -1
    }

    // 已经过期
    remaining := time.Until(expireAt).Seconds()
    if remaining <= 0 {
        return -2
    }

    return int64(remaining)
}

// isExpired 检查键是否过期（内部方法，不加锁）
func (s *Store) isExpired(key string) bool {
    expireAt, exists := s.expires[key]
    if !exists {
        return false
    }
    return time.Now().After(expireAt)
}
```

#### 3. 修改 Get 方法（懒删除）

```go
func (s *Store) Get(key string) (interface{}, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()

    // 检查是否过期
    if s.isExpired(key) {
        // 删除过期键
        delete(s.data, key)
        delete(s.expires, key)
        return nil, false
    }

    value, exists := s.data[key]
    return value, exists
}
```

#### 4. 实现定期删除

```go
// cleanupExpiredKeys 后台定期清理过期键
func (s *Store) cleanupExpiredKeys() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            s.deleteExpiredKeys()
        case <-s.stopCh:
            return
        }
    }
}

// deleteExpiredKeys 删除过期的键（随机抽样）
func (s *Store) deleteExpiredKeys() {
    s.mu.Lock()
    defer s.mu.Unlock()

    // 随机抽样 20 个键检查
    const sampleSize = 20
    count := 0

    for key, expireAt := range s.expires {
        if count >= sampleSize {
            break
        }

        if time.Now().After(expireAt) {
            delete(s.data, key)
            delete(s.expires, key)
        }

        count++
    }
}

// Close 停止清理 goroutine
func (s *Store) Close() {
    close(s.stopCh)
}
```

#### 5. 实现 EXPIRE 命令

**文件**：`handler/expire.go`

```go
type ExpireHandler struct {
    db *store.Store
}

func (h *ExpireHandler) Handle(args []protocol.Value) *protocol.Value {
    // 1. 参数验证：2 个参数（key, seconds）
    // 2. 解析 seconds 为整数
    //    seconds, err := strconv.Atoi(args[1].Str)
    // 3. 调用 Store.Expire
    //    success := h.db.Expire(key, seconds)
    // 4. 返回 0（失败）或 1（成功）
    //    return protocol.Integer(0 or 1)
}
```

#### 6. 实现 TTL 命令

**文件**：`handler/ttl.go`

```go
type TtlHandler struct {
    db *store.Store
}

func (h *TtlHandler) Handle(args []protocol.Value) *protocol.Value {
    // 1. 参数验证：1 个参数（key）
    // 2. 调用 Store.TTL
    //    ttl := h.db.TTL(key)
    // 3. 返回剩余秒数
    //    -2: 键不存在
    //    -1: 键存在但没有设置过期时间
    //    >=0: 剩余秒数
}
```

#### 7. 注册命令

```go
// handler/router.go
r.Register("EXPIRE", NewExpireHandler(r.db))
r.Register("TTL", NewTtlHandler(r.db))
```

---

### 验证方法

#### 单元测试

```go
func TestExpiration(t *testing.T) {
    s := store.NewStore()
    defer s.Close()

    // 设置键并过期
    s.Set("mykey", "value")
    s.Expire("mykey", 1)

    // 立即获取应该存在
    val, exists := s.Get("mykey")
    assert.True(t, exists)

    // 等待 2 秒后应该过期
    time.Sleep(2 * time.Second)
    val, exists = s.Get("mykey")
    assert.False(t, exists)
}

func TestTTL(t *testing.T) {
    s := store.NewStore()
    defer s.Close()

    // 键不存在
    assert.Equal(t, int64(-2), s.TTL("nonexistent"))

    // 键存在但无过期时间
    s.Set("mykey", "value")
    assert.Equal(t, int64(-1), s.TTL("mykey"))

    // 设置过期时间
    s.Expire("mykey", 10)
    ttl := s.TTL("mykey")
    assert.True(t, ttl > 0 && ttl <= 10)
}
```

#### 集成测试

```bash
redis-cli -p 16379

127.0.0.1:16379> SET mykey "Hello"
OK
127.0.0.1:16379> EXPIRE mykey 10
(integer) 1
127.0.0.1:16379> TTL mykey
(integer) 9
127.0.0.1:16379> TTL mykey
(integer) 8

# 等待 10 秒后
127.0.0.1:16379> GET mykey
(nil)
```

---

## ✅ 验收标准

### 功能验收

- [ ] 6 个核心命令全部实现且测试通过
- [ ] INCR/DECR 的原子性保证
- [ ] MGET/MSET 批量操作正确
- [ ] 过期机制正常工作（懒删除 + 定期删除）
- [ ] EXPIRE/TTL 命令正确
- [ ] 所有单元测试通过
- [ ] redis-cli 集成测试通过

### 性能验收

- [ ] INCR 并发性能测试（可选）
- [ ] 过期清理不影响正常操作

### 代码质量

- [ ] 代码格式化（`go fmt`）
- [ ] 静态检查通过（`go vet`）
- [ ] 测试覆盖率 > 80%

---

## 📊 进度跟踪

| 任务 | 预计耗时 | 实际耗时 | 状态 |
|------|---------|---------|------|
| INCR 实现 | 30分钟 | | ⬜ |
| DECR 实现 | 10分钟 | | ⬜ |
| INCRBY 实现 | 10分钟 | | ⬜ |
| DECRBY 实现 | 10分钟 | | ⬜ |
| MSET 实现 | 20分钟 | | ⬜ |
| MGET 实现 | 30分钟 | | ⬜ |
| 命令注册和测试 | 20分钟 | | ⬜ |
| **Part 1 小计** | **2-3小时** | | |
| Store 扩展 | 20分钟 | | ⬜ |
| EXPIRE 实现 | 20分钟 | | ⬜ |
| TTL 实现 | 10分钟 | | ⬜ |
| 懒删除机制 | 10分钟 | | ⬜ |
| 定期删除机制 | 30分钟 | | ⬜ |
| 过期测试 | 20分钟 | | ⬜ |
| **Part 2 小计** | **1-2小时** | | |
| **总计** | **3-4小时** | | |

---

## 🚀 实现建议

### 顺序

1. **Part 1 优先**：先实现 6 个核心命令
   - 按顺序实现：INCR → DECR → INCRBY → DECRBY → MSET → MGET
   - 每实现 1-2 个就测试一次

2. **Part 2 其次**：再实现过期机制
   - 先扩展 Store
   - 再实现命令
   - 最后测试整体

### 调试技巧

- 使用 `redis-cli` 的 `--raw` 模式查看原始输出
- 使用 Go 的 `testing.T.Log()` 打印调试信息
- 遇到问题先检查 RESP 协议格式是否正确

### 常见坑

1. **INCR/DECR**：注意类型转换，字符串 "10" 和 int64(10) 都要支持
2. **MGET**：返回的是 Array，注意构造 `protocol.Value`
3. **过期机制**：`time.After()` vs `time.Until()`，注意时区
4. **并发安全**：Store 的锁已经保证，但要注意不要死锁

---

## 📚 参考资料

- Redis 命令文档：https://redis.io/commands/
- RESP 协议规范：https://redis.io/docs/reference/protocol-spec/
- 现有代码参考：
  - `handler/get.go` - GET 命令实现
  - `handler/set.go` - SET 命令实现
  - `handler/keys.go` - KEYS 命令（处理数组返回）

---

**祝实现顺利！完成后就可以开始架构设计学习了！** 🎉

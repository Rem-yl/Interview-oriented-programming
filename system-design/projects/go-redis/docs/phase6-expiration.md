# 第六阶段：过期时间支持

## 1. 需求概述

实现 Redis 的键过期功能，这是 Redis 最重要的特性之一。支持为键设置过期时间，并在过期后自动删除，广泛应用于缓存、会话管理、限流等场景。

### 1.1 业务背景

过期功能的典型应用场景：
- **缓存**：自动清理过期的缓存数据
- **会话管理**：用户会话自动过期
- **验证码**：验证码有效期控制
- **限流**：时间窗口内的请求计数

### 1.2 核心目标

- 实现过期时间设置和查询
- 实现后台过期键清理机制
- 保证过期检查的性能
- 支持懒删除和定期删除策略

---

## 2. 命令清单

| 命令 | 语法 | 功能 | 返回值 |
|------|------|------|--------|
| **EXPIRE** | `EXPIRE key seconds` | 设置过期时间（秒） | 1/0 |
| **EXPIREAT** | `EXPIREAT key timestamp` | 设置过期时间戳 | 1/0 |
| **TTL** | `TTL key` | 查看剩余时间（秒） | 秒数/-1/-2 |
| **PERSIST** | `PERSIST key` | 移除过期时间 | 1/0 |
| **SETEX** | `SETEX key seconds value` | 设置值并指定过期时间 | OK |
| **PEXPIRE** | `PEXPIRE key milliseconds` | 设置过期时间（毫秒） | 1/0 |
| **PTTL** | `PTTL key` | 查看剩余时间（毫秒） | 毫秒数/-1/-2 |

---

## 3. 架构设计

### 3.1 数据结构设计

#### 扩展 Store 结构

```go
package store

import (
    "sync"
    "time"
)

// Store 扩展支持过期时间
type Store struct {
    mu         sync.RWMutex
    data       map[string]interface{}
    expires    map[string]time.Time  // ← 新增：过期时间映射
    stopClean  chan struct{}         // ← 新增：停止清理信号
}

// NewStore 创建带过期功能的 Store
func NewStore() *Store {
    s := &Store{
        data:      make(map[string]interface{}),
        expires:   make(map[string]time.Time),
        stopClean: make(chan struct{}),
    }

    // 启动后台清理 goroutine
    go s.cleanupExpiredKeys()

    return s
}
```

### 3.2 核心方法

```go
// SetWithExpire 设置键值对并指定过期时间
func (s *Store) SetWithExpire(key string, value interface{}, expiration time.Duration) {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.data[key] = value
    if expiration > 0 {
        s.expires[key] = time.Now().Add(expiration)
    }
}

// Get 获取值（自动检查过期）
func (s *Store) Get(key string) (interface{}, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    // 检查是否过期
    if s.isExpired(key) {
        return nil, false
    }

    value, exists := s.data[key]
    return value, exists
}

// Expire 设置过期时间
func (s *Store) Expire(key string, seconds int64) bool {
    s.mu.Lock()
    defer s.mu.Unlock()

    // 键必须存在
    if _, exists := s.data[key]; !exists {
        return false
    }

    s.expires[key] = time.Now().Add(time.Duration(seconds) * time.Second)
    return true
}

// TTL 获取剩余时间（秒）
func (s *Store) TTL(key string) int64 {
    s.mu.RLock()
    defer s.mu.RUnlock()

    // 键不存在
    if _, exists := s.data[key]; !exists {
        return -2
    }

    // 没有设置过期时间
    expireTime, hasExpire := s.expires[key]
    if !hasExpire {
        return -1
    }

    // 计算剩余时间
    ttl := time.Until(expireTime)
    if ttl <= 0 {
        return -2 // 已过期
    }

    return int64(ttl.Seconds())
}

// Persist 移除过期时间
func (s *Store) Persist(key string) bool {
    s.mu.Lock()
    defer s.mu.Unlock()

    // 键必须存在且有过期时间
    if _, exists := s.data[key]; !exists {
        return false
    }

    if _, hasExpire := s.expires[key]; !hasExpire {
        return false
    }

    delete(s.expires, key)
    return true
}

// isExpired 检查键是否过期（内部方法，调用前需加锁）
func (s *Store) isExpired(key string) bool {
    expireTime, hasExpire := s.expires[key]
    if !hasExpire {
        return false
    }

    return time.Now().After(expireTime)
}
```

### 3.3 过期键清理策略

Redis 使用两种策略清理过期键：

#### 策略 1：懒删除（Lazy Deletion）
访问键时检查是否过期，过期则删除。

```go
// Get 方法中的懒删除
func (s *Store) Get(key string) (interface{}, bool) {
    s.mu.Lock() // 需要写锁，可能删除
    defer s.mu.Unlock()

    // 懒删除：访问时检查过期
    if s.isExpiredNoLock(key) {
        delete(s.data, key)
        delete(s.expires, key)
        return nil, false
    }

    value, exists := s.data[key]
    return value, exists
}
```

#### 策略 2：定期删除（Periodic Deletion）
后台 goroutine 定期随机抽查并删除过期键。

```go
// cleanupExpiredKeys 后台清理过期键
func (s *Store) cleanupExpiredKeys() {
    ticker := time.NewTicker(100 * time.Millisecond) // 每 100ms 检查一次
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            s.cleanupRound()
        case <-s.stopClean:
            return
        }
    }
}

// cleanupRound 一轮清理
func (s *Store) cleanupRound() {
    s.mu.Lock()
    defer s.mu.Unlock()

    now := time.Now()
    expiredKeys := make([]string, 0)

    // 随机抽查一定数量的键
    maxCheck := 20
    checked := 0

    for key, expireTime := range s.expires {
        if checked >= maxCheck {
            break
        }
        checked++

        if now.After(expireTime) {
            expiredKeys = append(expiredKeys, key)
        }
    }

    // 删除过期键
    for _, key := range expiredKeys {
        delete(s.data, key)
        delete(s.expires, key)
        logger.Debugf("Expired key deleted: %s", key)
    }
}

// Stop 停止后台清理
func (s *Store) Stop() {
    close(s.stopClean)
}
```

---

## 4. Handler 实现

### 4.1 EXPIRE Handler

```go
package handler

import (
    "go-redis/protocol"
    "go-redis/store"
    "strconv"
)

type ExpireHandler struct {
    db *store.Store
}

func NewExpireHandler(db *store.Store) *ExpireHandler {
    return &ExpireHandler{db: db}
}

func (h *ExpireHandler) Handle(args []protocol.Value) *protocol.Value {
    if len(args) != 2 {
        return protocol.Error("ERR wrong number of arguments for 'expire' command")
    }

    key := args[0].Str
    secondsStr := args[1].Str

    seconds, err := strconv.ParseInt(secondsStr, 10, 64)
    if err != nil || seconds < 0 {
        return protocol.Error("ERR invalid expire time in 'expire' command")
    }

    success := h.db.Expire(key, seconds)
    if success {
        return protocol.Integer(1)
    }
    return protocol.Integer(0)
}
```

### 4.2 TTL Handler

```go
type TTLHandler struct {
    db *store.Store
}

func NewTTLHandler(db *store.Store) *TTLHandler {
    return &TTLHandler{db: db}
}

func (h *TTLHandler) Handle(args []protocol.Value) *protocol.Value {
    if len(args) != 1 {
        return protocol.Error("ERR wrong number of arguments for 'ttl' command")
    }

    key := args[0].Str
    ttl := h.db.TTL(key)

    return protocol.Integer(ttl)
}
```

### 4.3 SETEX Handler

```go
type SetexHandler struct {
    db *store.Store
}

func NewSetexHandler(db *store.Store) *SetexHandler {
    return &SetexHandler{db: db}
}

func (h *SetexHandler) Handle(args []protocol.Value) *protocol.Value {
    if len(args) != 3 {
        return protocol.Error("ERR wrong number of arguments for 'setex' command")
    }

    key := args[0].Str
    secondsStr := args[1].Str
    value := args[2].Str

    seconds, err := strconv.ParseInt(secondsStr, 10, 64)
    if err != nil || seconds <= 0 {
        return protocol.Error("ERR invalid expire time in 'setex' command")
    }

    h.db.SetWithExpire(key, value, time.Duration(seconds)*time.Second)

    return protocol.SimpleString("OK")
}
```

### 4.4 PERSIST Handler

```go
type PersistHandler struct {
    db *store.Store
}

func NewPersistHandler(db *store.Store) *PersistHandler {
    return &PersistHandler{db: db}
}

func (h *PersistHandler) Handle(args []protocol.Value) *protocol.Value {
    if len(args) != 1 {
        return protocol.Error("ERR wrong number of arguments for 'persist' command")
    }

    key := args[0].Str
    success := h.db.Persist(key)

    if success {
        return protocol.Integer(1)
    }
    return protocol.Integer(0)
}
```

---

## 5. 测试

### 5.1 单元测试

```go
func TestExpire(t *testing.T) {
    s := store.NewStore()
    defer s.Stop()

    // 设置键
    s.Set("mykey", "value")

    // 设置过期时间
    success := s.Expire("mykey", 1) // 1秒过期
    if !success {
        t.Error("Expected expire to succeed")
    }

    // 立即检查 TTL
    ttl := s.TTL("mykey")
    if ttl <= 0 || ttl > 1 {
        t.Errorf("Expected TTL around 1 second, got %d", ttl)
    }

    // 等待过期
    time.Sleep(1100 * time.Millisecond)

    // 检查键是否过期
    _, exists := s.Get("mykey")
    if exists {
        t.Error("Key should have expired")
    }

    // TTL 应该返回 -2
    ttl = s.TTL("mykey")
    if ttl != -2 {
        t.Errorf("Expected TTL -2 for expired key, got %d", ttl)
    }
}

func TestPersist(t *testing.T) {
    s := store.NewStore()
    defer s.Stop()

    // 设置键并添加过期时间
    s.SetWithExpire("mykey", "value", 10*time.Second)

    // 移除过期时间
    success := s.Persist("mykey")
    if !success {
        t.Error("Expected persist to succeed")
    }

    // TTL 应该返回 -1（永不过期）
    ttl := s.TTL("mykey")
    if ttl != -1 {
        t.Errorf("Expected TTL -1 after persist, got %d", ttl)
    }
}

func TestSetex(t *testing.T) {
    s := store.NewStore()
    r := handler.NewRouter(s)
    defer s.Stop()

    // SETEX mykey 1 "value"
    cmdResp := "*4\r\n$5\r\nSETEX\r\n$5\r\nmykey\r\n$1\r\n1\r\n$5\r\nvalue\r\n"
    reader := strings.NewReader(cmdResp)
    p := protocol.NewParser(reader)
    cmd, _ := p.Parse()

    resp := r.Route(cmd)
    if resp.Str != "OK" {
        t.Error("Expected OK from SETEX")
    }

    // 检查 TTL
    ttl := s.TTL("mykey")
    if ttl <= 0 || ttl > 1 {
        t.Errorf("Expected TTL around 1 second, got %d", ttl)
    }

    // 等待过期
    time.Sleep(1100 * time.Millisecond)

    _, exists := s.Get("mykey")
    if exists {
        t.Error("Key should have expired")
    }
}
```

### 5.2 并发测试

```go
func TestExpireConcurrent(t *testing.T) {
    s := store.NewStore()
    defer s.Stop()

    s.Set("counter", int64(0))

    var wg sync.WaitGroup

    // 并发设置过期时间
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            s.Expire("counter", 10)
        }()
    }

    // 并发检查 TTL
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            s.TTL("counter")
        }()
    }

    wg.Wait()

    // 验证键仍然存在且有过期时间
    ttl := s.TTL("counter")
    if ttl <= 0 || ttl > 10 {
        t.Errorf("Expected valid TTL, got %d", ttl)
    }
}
```

### 5.3 集成测试（redis-cli）

```bash
# 启动服务器
go run main.go

# EXPIRE 测试
127.0.0.1:16379> SET mykey "Hello"
OK
127.0.0.1:16379> EXPIRE mykey 10
(integer) 1
127.0.0.1:16379> TTL mykey
(integer) 10

# SETEX 测试
127.0.0.1:16379> SETEX session:user1 60 "session_data"
OK
127.0.0.1:16379> TTL session:user1
(integer) 60

# PERSIST 测试
127.0.0.1:16379> PERSIST session:user1
(integer) 1
127.0.0.1:16379> TTL session:user1
(integer) -1

# 过期验证
127.0.0.1:16379> SETEX temp 2 "will_expire"
OK
127.0.0.1:16379> GET temp
"will_expire"
# 等待 2 秒
127.0.0.1:16379> GET temp
(nil)
127.0.0.1:16379> TTL temp
(integer) -2
```

---

## 6. 性能优化

### 6.1 优化清理频率

```go
// 根据键数量动态调整清理频率
func (s *Store) adaptiveCleanup() {
    for {
        keyCount := len(s.data)

        // 键少时降低清理频率
        var interval time.Duration
        if keyCount < 100 {
            interval = 1 * time.Second
        } else if keyCount < 1000 {
            interval = 500 * time.Millisecond
        } else {
            interval = 100 * time.Millisecond
        }

        time.Sleep(interval)
        s.cleanupRound()
    }
}
```

### 6.2 使用最小堆优化

```go
// 使用最小堆存储过期时间，优先清理最早过期的键
type expiryHeap []expiryItem

type expiryItem struct {
    key        string
    expireTime time.Time
}

func (h expiryHeap) Len() int           { return len(h) }
func (h expiryHeap) Less(i, j int) bool { return h[i].expireTime.Before(h[j].expireTime) }
func (h expiryHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

// 每次清理时只检查堆顶的键
func (s *Store) cleanupHeapBased() {
    if len(s.expiryHeap) == 0 {
        return
    }

    now := time.Now()
    for len(s.expiryHeap) > 0 && s.expiryHeap[0].expireTime.Before(now) {
        item := heap.Pop(&s.expiryHeap).(expiryItem)
        delete(s.data, item.key)
        delete(s.expires, item.key)
    }
}
```

---

## 7. 验收标准

### 7.1 功能验收

- [ ] EXPIRE 命令正确设置过期时间
- [ ] TTL 命令正确返回剩余时间
- [ ] 键过期后自动删除
- [ ] PERSIST 正确移除过期时间
- [ ] SETEX 原子操作成功
- [ ] 后台清理机制正常工作
- [ ] 并发访问安全

### 7.2 性能验收

- [ ] 过期检查不影响 GET 性能（< 10% 性能下降）
- [ ] 后台清理 CPU 占用 < 5%
- [ ] 支持至少 10,000 个过期键

### 7.3 边界情况

- [ ] 过期时间为 0 或负数的处理
- [ ] 过期时间溢出的处理
- [ ] 不存在的键设置过期时间
- [ ] 已过期键的 GET 操作

---

## 8. 下一步

完成过期功能后，可以：

1. **实现持久化**（Phase 7）
   - RDB 快照需要保存过期时间
   - AOF 日志需要记录 EXPIRE 命令

2. **实现淘汰策略**
   - LRU（最近最少使用）
   - LFU（最不经常使用）
   - Random（随机淘汰）

3. **优化内存使用**
   - 过期字典的内存优化
   - 使用位图存储过期标记

---

**过期功能是 Redis 的核心特性，完成后你的 Redis 将更加实用！** 🎯

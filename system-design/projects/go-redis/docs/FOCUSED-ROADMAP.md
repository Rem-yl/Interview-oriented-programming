# Go-Redis 核心系统设计学习路线

## 🎯 核心理念

**不追求命令数量，聚焦系统设计核心概念**

```
❌ 错误思路：实现 200+ 个命令
✅ 正确思路：掌握分布式系统核心设计
```

---

## 📍 当前状态

✅ **已完成的核心能力**：
- 网络通信（TCP Server）
- 协议解析（RESP）
- 命令路由（Handler Pattern）
- 并发控制（Goroutine + Lock）
- 基础存储（In-Memory Map）

**结论**：基础架构已完备，可以直接进入系统设计学习。

---

## 🎓 精简学习路线（推荐）

### Phase A: 数据可靠性 ⭐⭐⭐（核心）

**学习目标**：理解分布式系统中的数据持久化

#### A1: RDB 快照持久化（1 周）

**为什么重要**：
- 理解快照机制（Point-in-Time Snapshot）
- 学习 Fork + Copy-on-Write
- 掌握序列化/反序列化

**核心实现**：
```go
// 1. 数据快照
type Snapshot struct {
    Timestamp  time.Time
    Data       map[string]interface{}
    Expires    map[string]time.Time
}

// 2. 保存（使用 encoding/gob 或 JSON）
func (s *Store) SaveSnapshot(filename string) error

// 3. 恢复
func (s *Store) LoadSnapshot(filename string) error

// 4. 后台保存（避免阻塞）
func (s *Store) BackgroundSave() error
```

**学到的概念**：
- 数据一致性
- 崩溃恢复
- 性能与可靠性权衡

**不需要实现**：
- ❌ 压缩算法（LZF）
- ❌ CRC 校验和
- ❌ 增量快照

---

#### A2: AOF 日志持久化（1 周）

**为什么重要**：
- 理解 Write-Ahead Logging
- 学习日志回放机制
- 掌握 fsync 和数据安全性

**核心实现**：
```go
// 1. 日志记录
type AOF struct {
    file   *os.File
    buffer *bufio.Writer
}

func (a *AOF) AppendCommand(cmd string) error {
    a.buffer.WriteString(cmd + "\n")
    // 根据策略 fsync
    if syncPolicy == "always" {
        return a.file.Sync()
    }
}

// 2. 日志重放
func (a *AOF) Replay(router *handler.Router) error {
    // 读取文件，逐行执行命令
}

// 3. AOF 重写（压缩）
func (a *AOF) Rewrite() error {
    // 将当前内存状态转为命令序列
}
```

**学到的概念**：
- 日志结构存储（Log-Structured Storage）
- Durability vs Performance
- 日志压缩（Compaction）

**不需要实现**：
- ❌ 混合持久化（RDB + AOF）
- ❌ AOF 格式的 RESP 优化

---

#### A3: 过期机制（1 周）

**为什么重要**：
- 理解缓存淘汰
- 学习后台任务调度
- 掌握时间轮算法

**核心实现**：
```go
// 1. 扩展数据结构
type Store struct {
    data    map[string]interface{}
    expires map[string]time.Time  // 过期时间
}

// 2. 懒删除（被动删除）
func (s *Store) Get(key string) (interface{}, bool) {
    if s.isExpired(key) {
        s.deleteExpired(key)
        return nil, false
    }
    return s.data[key], true
}

// 3. 定期删除（主动删除）
func (s *Store) cleanupLoop() {
    ticker := time.NewTicker(100 * time.Millisecond)
    for range ticker.C {
        s.sampleAndDelete(20) // 随机抽样 20 个键
    }
}
```

**学到的概念**：
- 懒删除 vs 主动删除
- 时间轮（Time Wheel）
- 资源回收策略

**不需要实现**：
- ❌ 复杂的淘汰策略（LRU, LFU）
- ❌ 最大内存限制（maxmemory）

---

### Phase B: 分布式协作 ⭐⭐⭐（核心）

**学习目标**：理解分布式系统中的数据一致性

#### B1: 主从复制（2 周）

**为什么重要**：
- 理解数据复制原理
- 学习最终一致性
- 掌握全量同步 + 增量同步

**核心实现**：
```go
// 1. Master 角色
type Master struct {
    replicas []*Replica
    backlog  *ReplicationBacklog // 复制积压缓冲区
}

func (m *Master) PropagateCommand(cmd string) {
    for _, replica := range m.replicas {
        replica.SendCommand(cmd)
    }
}

// 2. Replica 角色
type Replica struct {
    masterAddr string
    offset     int64 // 复制偏移量
}

func (r *Replica) Sync() error {
    // 1. 发送 PSYNC offset
    // 2. 接收 RDB 快照
    // 3. 应用增量命令
}

// 3. 复制协议
// PSYNC <replication-id> <offset>
// +FULLRESYNC <replication-id> <offset>
// +CONTINUE
```

**学到的概念**：
- 全量同步（Full Sync）
- 增量同步（Partial Sync）
- 复制偏移量（Offset）
- 主从延迟（Replication Lag）

**不需要实现**：
- ❌ 无盘复制（Diskless Replication）
- ❌ 链式复制（Cascading Replication）

---

#### B2: 高可用：简化版哨兵（1 周）

**为什么重要**：
- 理解故障检测
- 学习自动故障转移
- 掌握分布式共识基础

**核心实现**：
```go
// 简化版：单个哨兵监控主节点
type Sentinel struct {
    masterAddr string
    replicas   []string
}

// 1. 心跳检测
func (s *Sentinel) monitorMaster() {
    ticker := time.NewTicker(1 * time.Second)
    for range ticker.C {
        if !s.ping(s.masterAddr) {
            s.failoverCount++
            if s.failoverCount >= 3 {
                s.doFailover() // 3 次失败则故障转移
            }
        }
    }
}

// 2. 故障转移
func (s *Sentinel) doFailover() {
    // 1. 选择最佳从节点（复制偏移量最大）
    // 2. 提升为主节点
    // 3. 通知其他从节点
}
```

**学到的概念**：
- 心跳检测（Heartbeat）
- 故障检测（Failure Detection）
- 自动故障转移（Failover）
- 选主算法（简化版）

**不需要实现**：
- ❌ 完整的 Raft/Paxos 共识算法
- ❌ 脑裂处理
- ❌ 多哨兵投票

---

### Phase C: 并发与性能 ⭐⭐（进阶）

**学习目标**：理解高性能系统设计

#### C1: 事务支持（1 周）

**为什么重要**：
- 理解 ACID 特性
- 学习乐观锁（WATCH）
- 掌握命令队列

**核心实现**：
```go
type Transaction struct {
    commands []Command
    watching map[string]uint64 // 监视的键及版本号
}

// MULTI
func (c *Client) StartTransaction() {
    c.inTransaction = true
    c.txn = &Transaction{}
}

// WATCH key
func (c *Client) Watch(key string) {
    version := c.store.GetVersion(key)
    c.txn.watching[key] = version
}

// EXEC
func (c *Client) ExecTransaction() []*Value {
    // 1. 检查 WATCH 的键是否被修改
    for key, oldVersion := range c.txn.watching {
        if c.store.GetVersion(key) != oldVersion {
            return nil // 事务失败
        }
    }

    // 2. 原子执行所有命令
    results := make([]*Value, len(c.txn.commands))
    for i, cmd := range c.txn.commands {
        results[i] = c.router.Route(cmd)
    }
    return results
}
```

**学到的概念**：
- 事务隔离
- 乐观锁（Optimistic Locking）
- CAS（Compare-And-Swap）

**不需要实现**：
- ❌ MVCC（多版本并发控制）
- ❌ 完整的 ACID 保证

---

#### C2: 发布/订阅（1 周）

**为什么重要**：
- 理解消息队列模式
- 学习观察者模式
- 掌握异步通信

**核心实现**：
```go
type PubSub struct {
    mu          sync.RWMutex
    channels    map[string]map[*Client]struct{}
    patterns    map[string]map[*Client]struct{}
}

// SUBSCRIBE channel
func (ps *PubSub) Subscribe(client *Client, channel string) {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    if ps.channels[channel] == nil {
        ps.channels[channel] = make(map[*Client]struct{})
    }
    ps.channels[channel][client] = struct{}{}
}

// PUBLISH channel message
func (ps *PubSub) Publish(channel string, message string) int {
    ps.mu.RLock()
    defer ps.mu.RUnlock()

    count := 0
    for client := range ps.channels[channel] {
        client.SendMessage(channel, message)
        count++
    }
    return count
}
```

**学到的概念**：
- 发布订阅模式
- 广播机制
- 消息路由

**不需要实现**：
- ❌ 持久化消息队列
- ❌ 消息 ACK

---

### Phase D: 可观测性 ⭐（运维）

**学习目标**：理解生产系统监控

#### D1: 监控和统计（1 周）

**核心实现**：
```go
type Metrics struct {
    Commands      int64         // 命令总数
    Connections   int64         // 连接总数
    KeysCount     int64         // 键总数
    Memory        int64         // 内存使用
    HitRate       float64       // 缓存命中率
    ExpiredKeys   int64         // 过期键数量
}

// INFO 命令
func (s *Server) Info() map[string]interface{} {
    return map[string]interface{}{
        "version":       "1.0.0",
        "uptime":        time.Since(s.startTime).Seconds(),
        "commands":      s.metrics.Commands,
        "connections":   s.metrics.Connections,
        "keys":          s.store.KeyCount(),
        "memory":        s.metrics.Memory,
        "hit_rate":      s.metrics.HitRate,
    }
}

// 暴露 Prometheus Metrics
http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    metrics := s.collectMetrics()
    fmt.Fprintf(w, "redis_commands_total %d\n", metrics.Commands)
    fmt.Fprintf(w, "redis_connections_total %d\n", metrics.Connections)
    // ...
})
```

**学到的概念**：
- 指标收集（Metrics）
- Prometheus 集成
- 可观测性（Observability）

---

## 🎯 推荐学习路径

### 路径 1：数据可靠性优先（推荐）

```
当前状态
    ↓
Phase A3: 过期机制（1 周）
    ↓
Phase A1: RDB 持久化（1 周）
    ↓
Phase A2: AOF 持久化（1 周）
    ↓
Phase B1: 主从复制（2 周）
    ↓
Phase C1: 事务（1 周）
    ↓
Phase D1: 监控（1 周）
```

**总计**：8 周完成核心系统设计学习

---

### 路径 2：分布式系统优先

```
当前状态
    ↓
Phase A3: 过期机制（1 周）
    ↓
Phase B1: 主从复制（2 周）
    ↓
Phase B2: 哨兵（1 周）
    ↓
Phase A1: RDB 持久化（1 周）
    ↓
Phase C1: 事务（1 周）
    ↓
Phase C2: 发布订阅（1 周）
```

**总计**：7 周完成分布式核心学习

---

## 📊 学习价值对比

| 学习内容 | 命令数量 | 系统设计价值 |
|---------|---------|-------------|
| **再实现 20 个字符串命令** | +20 | ⭐ 低（重复劳动） |
| **实现 RDB 持久化** | +2 | ⭐⭐⭐⭐⭐ 极高 |
| **实现主从复制** | +3 | ⭐⭐⭐⭐⭐ 极高 |
| **实现事务** | +4 | ⭐⭐⭐⭐ 高 |
| **实现发布订阅** | +5 | ⭐⭐⭐⭐ 高 |

---

## 🎓 每个阶段的学习重点

### Phase A: 数据可靠性
**理论学习**：
- 《Designing Data-Intensive Applications》第 3 章（存储与检索）
- Redis 持久化文档
- WAL（Write-Ahead Logging）原理

**实践重点**：
- 数据一致性保证
- 性能与可靠性权衡
- 崩溃恢复机制

---

### Phase B: 分布式协作
**理论学习**：
- 《Designing Data-Intensive Applications》第 5 章（复制）
- CAP 定理
- 最终一致性

**实践重点**：
- 主从同步协议
- 故障检测和恢复
- 数据一致性保证

---

### Phase C: 并发与性能
**理论学习**：
- 事务隔离级别
- 乐观锁 vs 悲观锁
- 发布订阅模式

**实践重点**：
- 并发控制
- 异步通信
- 性能优化

---

## 💡 最终建议

### 精简命令集（足够用）

保留核心命令即可：
```bash
# 字符串（5 个足够）
SET, GET, DEL, INCR, EXPIRE

# 列表（可选，选 4 个）
LPUSH, RPUSH, LPOP, LRANGE

# 哈希（可选，选 4 个）
HSET, HGET, HDEL, HGETALL

# 通用（2 个）
KEYS, EXISTS
```

**总计**：11-15 个命令足够支撑所有系统设计学习。

---

### 不要实现的功能（性价比低）

❌ **命令数量堆砌**
- APPEND, STRLEN, GETRANGE 等（边际价值低）
- MSETNX, SETRANGE 等（用得少）

❌ **复杂数据类型**
- Sorted Set（跳表实现复杂，价值不大）
- Stream（太新，概念复杂）
- Bitmap, HyperLogLog（特殊场景）

❌ **高级特性**
- Lua 脚本（需要嵌入脚本引擎）
- Redis Cluster（分片太复杂）
- Redis Modules（API 设计复杂）

---

### 应该深入的功能（高价值）

✅ **持久化**（核心中的核心）
- RDB：理解快照
- AOF：理解 WAL

✅ **复制**（分布式基础）
- 主从复制：理解数据同步
- 哨兵：理解故障转移

✅ **并发控制**（性能关键）
- 事务：理解隔离
- 发布订阅：理解异步

✅ **可观测性**（生产必备）
- 监控指标
- 日志记录

---

## 🎯 8 周学习计划（推荐）

| 周次 | 内容 | 产出 |
|------|------|------|
| 第 1 周 | 过期机制 | EXPIRE, TTL, 后台清理 |
| 第 2 周 | RDB 持久化 | SAVE, BGSAVE, 加载恢复 |
| 第 3 周 | AOF 持久化 | AOF 记录、重放、重写 |
| 第 4-5 周 | 主从复制 | REPLICAOF, PSYNC |
| 第 6 周 | 事务 | MULTI, EXEC, WATCH |
| 第 7 周 | 发布订阅 | PUBLISH, SUBSCRIBE |
| 第 8 周 | 监控运维 | INFO, MONITOR, Metrics |

**完成后**：
- 掌握分布式系统核心设计
- 理解数据一致性和可靠性
- 拥有完整的生产级思维
- 可以写出高质量的系统设计文档

---

## 📚 配套学习资源

### 书籍（必读）
1. **《Designing Data-Intensive Applications》** - Martin Kleppmann
   - 第 3 章：存储与检索
   - 第 5 章：复制
   - 第 7 章：事务

2. **《Redis 设计与实现》** - 黄健宏
   - 第 9-11 章：持久化
   - 第 15-16 章：复制和哨兵

### 论文（选读）
- Raft 共识算法（简化版哨兵可以参考）
- The Log-Structured Merge-Tree (LSM-Tree)

### 源码（参考）
- [Redis 官方源码](https://github.com/redis/redis)
  - `rdb.c` - RDB 实现
  - `aof.c` - AOF 实现
  - `replication.c` - 复制实现

---

**总结：用 20% 的命令实现，学习 80% 的系统设计核心知识！** 🚀

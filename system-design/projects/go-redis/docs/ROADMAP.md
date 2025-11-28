# Go-Redis 扩展开发路线图

## 📍 当前状态

✅ **已完成**：
- Phase 1: 存储层（Store）
- Phase 2: 协议层（RESP Protocol）
- Phase 3: 命令处理层（Handler）
- Phase 4: 服务器层（TCP Server）

**当前功能**：
- 6 个基础命令（PING, SET, GET, DEL, EXISTS, KEYS）
- 完整的 RESP 协议支持
- 并发客户端处理
- 可使用 `redis-cli` 连接

---

## 🎯 扩展路线图总览

```
┌─────────────────────────────────────────────────────────┐
│                   已完成 (Phase 1-4)                     │
│        基础存储 + 协议 + 命令处理 + 服务器               │
└───────────────────────┬─────────────────────────────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
┌───────▼────────┐ ┌───▼────────┐ ┌───▼────────┐
│   短期目标      │ │  中期目标   │ │  长期目标   │
│  (1-2 周)      │ │ (3-4 周)   │ │  (5-8 周)  │
└───────┬────────┘ └───┬────────┘ └───┬────────┘
        │              │              │
  Phase 5-6       Phase 7-8      Phase 9-12
```

---

## 📅 详细规划

### Phase 5: 扩展命令 ⭐⭐⭐（优先级最高）

**预计时间**：1 周
**文档**：[phase5-advanced-commands.md](./phase5-advanced-commands.md)

#### 目标
实现更多实用命令，提升系统可用性。

#### 命令列表

**字符串操作**（4 个命令）：
- `APPEND key value` - 追加字符串
- `STRLEN key` - 获取长度
- `GETRANGE key start end` - 获取子串
- `SETRANGE key offset value` - 设置子串

**数值操作**（4 个命令）：
- `INCR key` - 自增 1
- `DECR key` - 自减 1
- `INCRBY key increment` - 增加指定值
- `DECRBY key decrement` - 减少指定值

**批量操作**（3 个命令）：
- `MGET key [key ...]` - 批量获取
- `MSET key value [key value ...]` - 批量设置
- `MSETNX key value [key value ...]` - 批量设置（不存在时）

#### 学习收获
- 原子操作实现（INCR/DECR）
- 并发安全保证
- 批量操作优化
- 类型检查和转换

#### 验收标准
- [ ] 11 个新命令全部实现
- [ ] INCR 并发测试通过
- [ ] 性能不低于基础命令
- [ ] 所有测试覆盖率 > 85%

---

### Phase 6: 过期时间支持 ⭐⭐⭐（优先级最高）

**预计时间**：1 周
**文档**：[phase6-expiration.md](./phase6-expiration.md)

#### 目标
实现 Redis 最重要的特性之一：键过期。

#### 命令列表

- `EXPIRE key seconds` - 设置过期时间（秒）
- `EXPIREAT key timestamp` - 设置过期时间戳
- `TTL key` - 查看剩余时间（秒）
- `PTTL key` - 查看剩余时间（毫秒）
- `PERSIST key` - 移除过期时间
- `SETEX key seconds value` - 设置值并指定过期时间
- `PEXPIRE key milliseconds` - 设置过期时间（毫秒）

#### 核心实现
1. **数据结构扩展**
   - 添加 `expires map[string]time.Time`
   - 修改 Get 方法支持过期检查

2. **清理策略**
   - 懒删除：访问时检查
   - 定期删除：后台 goroutine

3. **性能优化**
   - 使用最小堆存储过期键
   - 动态调整清理频率

#### 学习收获
- 时间管理
- 后台任务实现
- 内存优化技巧
- 缓存淘汰策略

#### 验收标准
- [ ] 7 个过期相关命令实现
- [ ] 后台清理机制正常
- [ ] 过期检查性能影响 < 10%
- [ ] 支持 10,000+ 过期键

---

### Phase 7: 持久化 ⭐⭐（中期目标）

**预计时间**：2 周

#### 7.1 RDB 快照持久化

**原理**：定期将内存数据保存到磁盘。

**实现要点**：
```go
// RDB 文件格式（简化）
type RDBSnapshot struct {
    Version    uint8
    Data       map[string]interface{}
    Expires    map[string]time.Time
    Checksum   uint64
}

// 保存快照
func (s *Store) SaveRDB(filename string) error {
    snapshot := RDBSnapshot{
        Version: 1,
        Data:    s.data,
        Expires: s.expires,
    }

    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := gob.NewEncoder(file)
    return encoder.Encode(snapshot)
}

// 加载快照
func (s *Store) LoadRDB(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    var snapshot RDBSnapshot
    decoder := gob.NewDecoder(file)
    if err := decoder.Decode(&snapshot); err != nil {
        return err
    }

    s.data = snapshot.Data
    s.expires = snapshot.Expires
    return nil
}
```

**新增命令**：
- `SAVE` - 同步保存快照
- `BGSAVE` - 后台保存快照
- `LASTSAVE` - 上次保存时间

**配置选项**：
```ini
# redis.conf
save 900 1      # 900秒内至少1次修改
save 300 10     # 300秒内至少10次修改
save 60 10000   # 60秒内至少10000次修改
```

#### 7.2 AOF 日志持久化

**原理**：记录每个写命令，崩溃恢复时重放。

**实现要点**：
```go
type AOFLog struct {
    file     *os.File
    mu       sync.Mutex
    commands []string
}

// 记录命令
func (a *AOFLog) Append(cmd string) error {
    a.mu.Lock()
    defer a.mu.Unlock()

    _, err := a.file.WriteString(cmd + "\n")
    if err != nil {
        return err
    }

    // 根据策略刷盘
    return a.file.Sync() // fsync
}

// 重放日志
func (a *AOFLog) Replay(router *handler.Router) error {
    file, err := os.Open("appendonly.aof")
    if err != nil {
        return err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        cmdStr := scanner.Text()
        cmd := parseCommand(cmdStr)
        router.Route(cmd) // 重放命令
    }

    return scanner.Err()
}
```

**新增命令**：
- `BGREWRITEAOF` - 后台重写 AOF

**配置选项**：
```ini
appendonly yes
appendfsync always    # 每次写入都刷盘
appendfsync everysec  # 每秒刷盘
appendfsync no        # 由操作系统决定
```

#### 学习收获
- 文件 I/O 操作
- 数据序列化/反序列化
- Fsync 和数据安全性
- Fork 和写时复制（COW）
- AOF 重写机制

---

### Phase 8: 复杂数据类型 ⭐⭐（中期目标）

**预计时间**：2 周

#### 8.1 List 列表

**底层实现**：双向链表或压缩列表

**命令**：
- `LPUSH key value [value ...]` - 左侧插入
- `RPUSH key value [value ...]` - 右侧插入
- `LPOP key` - 左侧弹出
- `RPOP key` - 右侧弹出
- `LLEN key` - 列表长度
- `LRANGE key start stop` - 获取范围
- `LINDEX key index` - 获取索引元素
- `LSET key index value` - 设置索引元素

**应用场景**：
- 消息队列
- 最新消息列表
- 时间线

#### 8.2 Hash 哈希表

**底层实现**：哈希表或压缩列表

**命令**：
- `HSET key field value` - 设置字段
- `HGET key field` - 获取字段
- `HDEL key field [field ...]` - 删除字段
- `HGETALL key` - 获取所有字段
- `HKEYS key` - 获取所有字段名
- `HVALS key` - 获取所有值
- `HEXISTS key field` - 字段是否存在
- `HLEN key` - 字段数量

**应用场景**：
- 用户信息存储
- 配置管理
- 购物车

#### 8.3 Set 集合

**底层实现**：哈希表或整数集合

**命令**：
- `SADD key member [member ...]` - 添加成员
- `SREM key member [member ...]` - 删除成员
- `SMEMBERS key` - 获取所有成员
- `SISMEMBER key member` - 成员是否存在
- `SCARD key` - 集合大小
- `SUNION key [key ...]` - 并集
- `SINTER key [key ...]` - 交集
- `SDIFF key [key ...]` - 差集

**应用场景**：
- 标签系统
- 好友关系
- 共同关注

---

### Phase 9: 发布/订阅 ⭐（长期目标）

**预计时间**：1 周

#### 命令
- `PUBLISH channel message` - 发布消息
- `SUBSCRIBE channel [channel ...]` - 订阅频道
- `UNSUBSCRIBE [channel ...]` - 取消订阅
- `PSUBSCRIBE pattern [pattern ...]` - 模式订阅
- `PUNSUBSCRIBE [pattern ...]` - 取消模式订阅

#### 实现要点
```go
type PubSub struct {
    mu          sync.RWMutex
    subscribers map[string]map[*Client]struct{} // 频道 -> 订阅者集合
}

func (ps *PubSub) Subscribe(channel string, client *Client) {
    ps.mu.Lock()
    defer ps.mu.Unlock()

    if ps.subscribers[channel] == nil {
        ps.subscribers[channel] = make(map[*Client]struct{})
    }
    ps.subscribers[channel][client] = struct{}{}
}

func (ps *PubSub) Publish(channel string, message string) int {
    ps.mu.RLock()
    defer ps.mu.RUnlock()

    count := 0
    for client := range ps.subscribers[channel] {
        client.SendMessage(message)
        count++
    }
    return count
}
```

---

### Phase 10: 事务 ⭐（长期目标）

**预计时间**：1 周

#### 命令
- `MULTI` - 开始事务
- `EXEC` - 执行事务
- `DISCARD` - 放弃事务
- `WATCH key [key ...]` - 监视键
- `UNWATCH` - 取消监视

#### 实现要点
```go
type Transaction struct {
    commands [][]protocol.Value
    watching map[string]uint64 // 监视的键及其版本号
}

func (t *Transaction) Watch(key string, version uint64) {
    t.watching[key] = version
}

func (t *Transaction) Exec(router *Router) []*protocol.Value {
    // 检查监视的键是否被修改
    for key, oldVersion := range t.watching {
        if currentVersion := getVersion(key); currentVersion != oldVersion {
            return nil // 事务失败
        }
    }

    // 执行所有命令
    results := make([]*protocol.Value, len(t.commands))
    for i, cmd := range t.commands {
        results[i] = router.Route(cmd)
    }

    return results
}
```

---

### Phase 11: 性能优化 ⭐

**预计时间**：持续进行

#### 11.1 内存优化
- 使用对象池减少 GC 压力
- 字符串 interning
- 压缩数据结构（ziplist, intset）

#### 11.2 并发优化
- 分段锁（Sharding）
- 无锁数据结构
- Goroutine 池

#### 11.3 网络优化
- 连接池
- Pipeline 支持
- 零拷贝

#### 11.4 I/O 优化
- 批量写入
- 缓冲区优化
- mmap 文件映射

---

### Phase 12: 监控和运维 ⭐

**预计时间**：1 周

#### 12.1 监控命令
- `INFO [section]` - 服务器信息
- `MONITOR` - 实时监控命令
- `CLIENT LIST` - 客户端列表
- `SLOWLOG GET [count]` - 慢查询日志
- `CONFIG GET/SET parameter` - 配置管理

#### 12.2 Metrics 暴露
```go
// 使用 Prometheus 格式暴露指标
type Metrics struct {
    Commands       int64  // 命令总数
    Connections    int64  // 连接总数
    Keys           int64  // 键总数
    Memory         int64  // 内存使用
    HitRate        float64 // 命中率
}

// HTTP 端点暴露 Metrics
http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    metrics := collectMetrics()
    w.Write([]byte(formatPrometheus(metrics)))
})
```

#### 12.3 日志系统
- 结构化日志（JSON）
- 日志级别控制
- 日志轮转

---

## 🎓 学习建议

### 推荐学习路径

```
Phase 5 (扩展命令)
    ↓
Phase 6 (过期时间)
    ↓
Phase 7.1 (RDB 持久化)
    ↓
Phase 8.1 (List 列表)
    ↓
Phase 7.2 (AOF 持久化)
    ↓
Phase 8.2 (Hash 哈希表)
    ↓
Phase 8.3 (Set 集合)
    ↓
Phase 11 (性能优化)
    ↓
Phase 9 (发布/订阅)
    ↓
Phase 10 (事务)
    ↓
Phase 12 (监控运维)
```

### 每个阶段的学习方法

1. **阅读文档**（20%）
   - 官方 Redis 文档
   - 相关博客和书籍

2. **设计方案**（20%）
   - 画架构图
   - 设计数据结构
   - 确定接口

3. **编写代码**（40%）
   - 先写测试
   - 再写实现
   - 重构优化

4. **测试验证**（20%）
   - 单元测试
   - 集成测试
   - 性能测试
   - redis-cli 验证

---

## 📊 技能提升树

完成所有阶段后，你将掌握：

### 编程技能
- ✅ Go 语言精通
- ✅ 并发编程
- ✅ 网络编程
- ✅ 文件 I/O
- ✅ 测试驱动开发

### 系统设计
- ✅ 分层架构
- ✅ 数据结构设计
- ✅ 协议设计
- ✅ 性能优化
- ✅ 可扩展性设计

### Redis 知识
- ✅ RESP 协议
- ✅ 数据类型实现
- ✅ 持久化机制
- ✅ 过期策略
- ✅ 事务实现
- ✅ 发布订阅模式

### 工程实践
- ✅ 代码规范
- ✅ 单元测试
- ✅ 性能测试
- ✅ 文档编写
- ✅ 版本控制

---

## 🎯 里程碑

| 里程碑 | 完成标志 | 预计时间 |
|-------|---------|---------|
| **M1: 基础完成** | Phase 1-4 完成 | ✅ 已完成 |
| **M2: 实用化** | Phase 5-6 完成 | 2 周 |
| **M3: 生产级** | Phase 7-8 完成 | 4 周 |
| **M4: 高级特性** | Phase 9-10 完成 | 2 周 |
| **M5: 完整系统** | Phase 11-12 完成 | 持续进行 |

---

## 📚 推荐资源

### 书籍
- 《Redis 设计与实现》- 黄健宏
- 《Redis 深度历险》- 钱文品
- Designing Data-Intensive Applications - Martin Kleppmann

### 源码
- [Redis 官方源码](https://github.com/redis/redis)（C）
- [Godis](https://github.com/HDT3213/godis)（Go 实现参考）

### 文档
- [Redis 官方文档](https://redis.io/docs/)
- [Redis 命令参考](https://redis.io/commands/)
- [RESP 协议规范](https://redis.io/docs/reference/protocol-spec/)

---

## 💡 最后的建议

1. **循序渐进**
   - 不要跳跃，按顺序完成
   - 每个阶段都要写测试
   - 每个阶段都要文档化

2. **深度优于广度**
   - 理解每个设计决策的原因
   - 不要只是模仿，要理解原理
   - 思考还有哪些实现方式

3. **实践第一**
   - 动手写代码比看文档重要
   - 遇到问题先自己思考
   - 用真实场景测试你的实现

4. **持续优化**
   - 性能测试很重要
   - 代码重构不可少
   - 关注代码可读性

---

**祝你的 Go-Redis 之旅顺利！记住：每一行代码都是学习的机会。** 🚀

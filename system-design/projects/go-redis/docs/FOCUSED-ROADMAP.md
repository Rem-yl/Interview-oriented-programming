# 系统架构设计学习路线

## 🎯 核心理念

**本项目是系统架构设计学习工具，不是 Redis 克隆**

```
❌ 错误目标：实现完整的 Redis 功能
✅ 正确目标：掌握分布式系统架构设计思维

学习方式：
  通过 Redis 案例 → 学习通用架构模式 → 应用到其他系统
```

**为什么选择 Redis 作为学习案例？**
- 架构演进清晰：单机 → 持久化 → 主从 → 集群
- 设计权衡明显：性能 vs 可靠性，简单 vs 复杂
- 应用广泛：架构模式可迁移到缓存、消息队列、数据库等系统

---

## 📍 当前状态

✅ **已完成：Stage 0 - 单机内存存储架构**

**架构能力**：
- 分层架构设计（Server → Protocol → Handler → Store）
- 并发安全设计（RWMutex）
- 协议解析（RESP）
- 命令模式（Handler Pattern）

**架构局限**：
- ❌ 无持久化：数据易失
- ❌ 单点故障：无容错能力
- ❌ 无法扩展：单机瓶颈

**下一步**：解决数据可靠性问题

---

## 🏗️ 架构演进路线

### Stage 1: 持久化架构设计 ⭐⭐⭐

**架构主题**：数据可靠性（Reliability）

#### 📋 学习路线

**推荐路径**：先实现一种持久化方案，深入理解，再对比另一种

```
路径 A（推荐）：RDB → AOF → 对比分析
路径 B：       AOF → RDB → 对比分析
```

#### Phase 1.1: RDB 快照架构（1-2 周）

**架构模式**：**Snapshot Pattern**（快照模式）

##### 问题与方案

**问题**：如何在不阻塞服务的情况下保存数据？

**方案对比**：
| 方案 | 优点 | 缺点 | 适用性 |
|------|------|------|-------|
| **停服保存** | 简单 | 服务中断❌ | 不可接受 |
| **加锁复制** | 可行 | 长时间锁，内存翻倍 | 小数据集 |
| **Fork + COW** | 不阻塞，高效 | 需要OS支持 | ✅ 最优 |

##### 核心架构图

```
写入流程（正常运行）：
Client → SET key value → Memory (快速完成)

后台保存流程（BGSAVE）：
┌────────────────────────────────┐
│  Main Process (主进程)          │
│  ┌──────────────────────────┐  │
│  │  In-Memory Data          │  │
│  │  {key1: val1, ...}       │  │
│  └──────────────────────────┘  │
│           │                     │
│           │ Fork()              │
│           ├──────────────┐      │
│           │              │      │
│  ┌────────▼──────┐  ┌────▼─────────────┐
│  │ 继续处理请求   │  │ Child Process    │
│  │ (写时复制)     │  │ 序列化数据        │
│  └───────────────┘  │ → dump.rdb       │
│                     └──────────────────┘
└────────────────────────────────────────┘
```

##### 关键设计决策

**决策 1：序列化格式**

```go
// 方案 A：JSON (易读，体积大，慢)
{
  "version": 1,
  "data": {"key1": "value1", "key2": 123}
}

// 方案 B：Gob (Go原生，紧凑，快) ✅ 选择此方案
// 二进制格式，直接编码 Go 数据结构

// 方案 C：自定义二进制 (最优性能，维护成本高)
// Redis 使用此方案
```

**决策 2：原子性保证**

```bash
# 问题：保存过程中崩溃怎么办？
# 解决：先写临时文件，成功后原子替换

1. 写入 → temp-12345.rdb
2. 成功 → rename(temp-12345.rdb, dump.rdb)  # 原子操作
3. 失败 → 保留旧的 dump.rdb，删除 temp 文件
```

##### 实现范围

**必须实现**：
```go
// 1. 数据结构
type Snapshot struct {
    Version   int
    Timestamp time.Time
    Data      map[string]interface{}
}

// 2. 核心接口（Store 层）
func (s *Store) Save(filename string) error    // 阻塞保存
func (s *Store) Load(filename string) error    // 加载

// 3. 命令（Handler 层）
SAVE      // 前台保存（会阻塞）
BGSAVE    // 后台保存（简化版，goroutine 实现即可，不用真 Fork）
```

**不必实现**：
- ❌ 真实的 Fork（Go 不支持，用 goroutine 模拟即可）
- ❌ LZF 压缩（性价比低）
- ❌ CRC 校验（可选优化）

**验证方法**：
```bash
# 1. 功能测试
SET key1 value1
SET key2 value2
BGSAVE
# 重启服务
GET key1  # 应该返回 value1

# 2. 性能测试
# 写入 10万 键值对，测试保存和加载时间
```

##### 学习重点

**理解的架构概念**：
- ✅ **Snapshot Pattern**：虚拟机快照、Docker镜像也用此模式
- ✅ **Fork + COW**：操作系统层面的优化技巧
- ✅ **原子操作**：如何保证数据一致性

**Trade-offs 分析**：
| 维度 | 选择 | 原因 |
|------|------|------|
| **性能 vs 可靠性** | 可靠性优先 | 允许短暂性能下降 |
| **空间 vs 时间** | 时间优先（紧凑格式） | 节省磁盘空间 |
| **一致性 vs 可用性** | 可用性优先（后台保存） | 不阻塞服务 |

**可迁移应用**：
- 数据库备份（PostgreSQL pg_dump）
- 容器镜像（Docker commit）
- 游戏存档（定期快照）

---

#### Phase 1.2: AOF 日志架构（1-2 周）

**架构模式**：**Write-Ahead Logging**（预写日志）

##### 问题与方案

**问题**：RDB丢失数据太多（最后一次快照后的数据），如何改进？

**对比**：
| 持久化方式 | RPO（丢失数据） | RTO（恢复时间） | 性能影响 |
|-----------|---------------|---------------|---------|
| **RDB** | 分钟级 | 快（秒级） | 小 |
| **AOF** | 秒级或0 | 慢（需重放） | 中-大 |

##### 核心架构图

```
写入流程（WAL）：
┌────────────────────────────────────────┐
│  Write Operation                       │
│                                        │
│  1. 写入 AOF 文件 ───┐                  │
│     appendonly.aof   │                 │
│     ↓                │                 │
│  2. fsync()  ←───────┘                 │
│     (根据策略决定)                       │
│     ↓                                  │
│  3. 修改内存数据                        │
│     ↓                                  │
│  4. 返回客户端                          │
└────────────────────────────────────────┘

AOF 文件内容（RESP 格式）：
*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
*2\r\n$3\r\nDEL\r\n$3\r\nkey\r\n

恢复流程：
启动 → 读取 AOF → 逐条执行命令 → 恢复完成
```

##### 关键设计决策

**决策 1：fsync 策略** （最重要！）

```go
type SyncPolicy string

const (
    AlwaysSync  SyncPolicy = "always"   // 每次写入立即 fsync
    EverySecond SyncPolicy = "everysec" // 每秒 fsync 一次
    NoSync      SyncPolicy = "no"       // 由操作系统决定
)
```

| 策略 | 丢失风险 | 性能 | QPS | 适用场景 |
|------|---------|------|-----|---------|
| `always` | 0-1条命令 | 极慢 | ~200 | 金融系统 |
| `everysec` | 最多1秒 | 中等 | ~50K | **推荐** |
| `no` | 数十秒 | 极快 | ~100K | 不重要数据 |

**决策 2：AOF Rewrite（重写机制）**

```
问题：AOF 文件无限增长
  SET counter 1
  INCR counter  # 100万次
  结果：AOF 有 100万+1 条命令

解决：定期重写
  压缩为：SET counter 1000000

触发条件：
  - 文件大小 > 64MB
  - 文件大小 > 上次重写的2倍
```

##### 实现范围

**必须实现**：
```go
// 1. AOF 管理器
type AOF struct {
    file       *os.File
    buffer     *bufio.Writer
    syncPolicy SyncPolicy
}

// 2. 核心接口
func (a *AOF) Append(cmd *protocol.Value) error  // 追加命令
func (a *AOF) Sync() error                       // 执行 fsync
func (a *AOF) Replay(router *Router) error       // 重放日志
func (a *AOF) Rewrite(store *Store) error        // AOF 重写

// 3. 命令
BGREWRITEAOF  // 后台重写 AOF
```

**实现要点**：
```go
// 写入时：
1. 将命令序列化为 RESP 格式
2. 写入 buffer
3. 根据 syncPolicy 决定是否立即 fsync

// 重写时：
1. 遍历当前内存数据
2. 为每个 key 生成 SET 命令
3. 写入新 AOF 文件
4. 原子替换旧文件
```

**不必实现**：
- ❌ RDB + AOF 混合持久化
- ❌ AOF 并发重写优化

**验证方法**：
```bash
# 1. 功能测试
SET key1 value1
# 重启服务
GET key1  # 应该返回 value1

# 2. fsync 策略测试
# 设置 always/everysec/no，对比性能差异

# 3. 重写测试
SET counter 0
INCR counter  # 1000次
BGREWRITEAOF
# 查看 AOF 文件大小变化
```

##### 学习重点

**架构模式**：
- ✅ **WAL**：MySQL binlog、PostgreSQL WAL、Kafka log
- ✅ **Log Compaction**：Kafka、LSM-Tree 都用此思想
- ✅ **Durability 保证**：理解 fsync 对性能的影响

**Trade-offs**：
```
可靠性 vs 性能：
  AOF always  → 几乎不丢数据，但慢
  AOF everysec → 最多丢1秒，性能尚可  ← Redis默认
  RDB         → 可能丢分钟级数据，但快

空间 vs 复杂度：
  不重写 → 简单，但文件巨大
  重写   → 复杂，但节省空间
```

---

#### 阶段总结：Stage 1

**学到的架构模式**：
1. ✅ **Snapshot Pattern**（RDB）
2. ✅ **Write-Ahead Logging**（AOF）
3. ✅ **Log Compaction**（AOF Rewrite）

**学到的设计原则**：
1. ✅ **Durability vs Performance** - 没有完美方案，只有最适合的
2. ✅ **分层设计** - 持久化层独立，不影响上层逻辑
3. ✅ **可配置性** - 让用户根据场景选择策略

**可迁移的知识**：
- 数据库持久化（PostgreSQL、MySQL）
- 消息队列（Kafka、RabbitMQ）
- 文件系统（Journaling FS）
- 版本控制系统（Git）

**下一阶段预告**：
- 问题：单点故障，无高可用
- 方案：主从复制 + 故障转移

---

## 📚 学习资源

### 必读

**书籍**：
- 《Designing Data-Intensive Applications》- Martin Kleppmann
  - 第3章：Storage and Retrieval（RDB、AOF原理）
  - 第5章：Replication（主从复制）

**Redis 官方文档**：
- [Redis Persistence](https://redis.io/docs/management/persistence/)
- [Replication](https://redis.io/docs/management/replication/)

### 源码阅读（选读）

**Redis 源码**（C语言，但逻辑清晰）：
- `rdb.c` - RDB 实现（2000行）
- `aof.c` - AOF 实现（1500行）
- `replication.c` - 复制协议（3000行）

**阅读方法**：
1. 不要逐行读，先看主流程
2. 关注核心数据结构
3. 理解设计决策（注释会说明）

---

## 🎯 下一步建议

### 推荐学习路径

```
当前 (Stage 0: 单机架构)
  ↓
Stage 1.1: RDB 快照架构 (1-2周)
  ↓
Stage 1.2: AOF 日志架构 (1-2周)
  ↓
对比分析：RDB vs AOF (写设计文档)
  ↓
Stage 2: 主从复制架构 (2-3周)
  ↓
Stage 3: 并发控制 (事务、Pub/Sub) (2周)
  ↓
总结：架构设计思维提炼
```

### 学习建议

**不要**：
- ❌ 追求功能完整性（不要实现所有命令）
- ❌ 过早优化（先实现基本功能，再优化）
- ❌ 孤立学习（每个阶段都要思考可迁移性）

**应该**：
- ✅ 先画架构图，再写代码
- ✅ 每个阶段写设计文档（记录 Trade-offs）
- ✅ 性能测试（验证设计假设）
- ✅ 对比其他系统（如何应用到 Kafka、PostgreSQL）

### 每个阶段的产出

**代码**：
- 核心功能实现
- 单元测试 + 集成测试
- 性能基准测试

**文档**：
- 架构设计文档（问题、方案、Trade-offs）
- 性能测试报告
- 学习笔记（可迁移的知识点）

---

## 💡 架构设计思维框架

### 每次设计时问自己

**1. 问题定义**
- 当前架构的局限是什么？
- 要解决什么具体问题？
- 非功能性需求是什么（性能、可靠性、一致性）？

**2. 方案对比**
- 有哪些可选方案？
- 每个方案的 Trade-offs 是什么？
- 为什么选择这个方案？

**3. 设计决策**
- 关键设计点有哪些？
- 每个决策的理由是什么？
- 如何验证设计是正确的？

**4. 学习反思**
- 学到了什么架构模式？
- 如何应用到其他系统？
- 如果重新设计，会怎么改进？

---

## 📊 进度跟踪

| 阶段 | 状态 | 学到的架构模式 | 耗时 |
|------|------|---------------|------|
| Stage 0: 单机架构 | ✅ 已完成 | 分层架构、命令模式 | - |
| Stage 1.1: RDB | ⬜ 待开始 | Snapshot Pattern | - |
| Stage 1.2: AOF | ⬜ 待开始 | WAL, Log Compaction | - |
| Stage 2: 主从复制 | ⬜ 待规划 | Replication | - |

---

**记住**：
> **目标不是实现 Redis，而是通过 Redis 学习系统架构设计**
>
> 衡量学习成果的标准：
> - 能独立设计一个分布式系统吗？
> - 能分析和权衡不同架构方案吗？
> - 能将学到的模式应用到其他系统吗？

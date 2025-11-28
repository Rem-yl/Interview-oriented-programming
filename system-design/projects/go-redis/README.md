# 系统架构设计学习项目

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![Learning](https://img.shields.io/badge/purpose-architecture%20learning-orange.svg)
![Status](https://img.shields.io/badge/stage-0%20completed-green.svg)

**通过 Redis 案例学习分布式系统架构设计**

[核心理念](#核心理念) • [快速开始](#快速开始) • [架构演进](#架构演进路线) • [学习指南](#学习方法)

</div>

---

## 🎯 核心理念

> **本项目不是 Redis 克隆，而是系统架构设计学习工具**

### 学习目标

```
❌ 错误目标：实现 200+ 个 Redis 命令
✅ 正确目标：掌握分布式系统架构设计思维

学习路径：
  分析需求 → 架构方案对比 → 设计 Trade-offs → 实现验证 → 总结迁移
```

### 为什么用 Redis 作为案例？

| 优势 | 说明 |
|------|------|
| **架构演进清晰** | 单机 → 持久化 → 主从 → 集群，层层递进 |
| **设计权衡明显** | 性能 vs 可靠性、简单 vs 复杂 |
| **模式可迁移** | 学到的架构模式可应用于缓存、消息队列、数据库 |
| **源码质量高** | Redis 源码清晰，是学习分布式系统的绝佳教材 |

### 学习成果标准

**不是**：能写多少代码、实现多少功能
**而是**：
- ✅ 能独立设计一个分布式系统吗？
- ✅ 能分析和权衡不同架构方案吗？
- ✅ 能将架构思维应用到其他系统吗？
- ✅ 能读懂并评价一个系统的架构设计吗？

---

## 🚀 快速开始

### 前置要求

- Go 1.23 或更高版本
- Redis CLI（用于测试，可选）

### 安装

```bash
# 克隆项目
git clone <repository-url>
cd go-redis

# 安装依赖
go mod download

# 构建项目
go build -o go-redis main.go
```

### 运行服务器

```bash
# 方式 1：直接运行
go run main.go

# 方式 2：使用构建的二进制文件
./go-redis
```

服务器默认监听在 `localhost:16379`（避免与系统 Redis 冲突）。

你应该看到如下输出：

```
INFO[2024-11-28 14:00:00] Starting Go-Redis server on :16379
INFO[2024-11-28 14:00:00] Redis server listening on :16379
```

### 连接到服务器

使用 Redis CLI 连接：

```bash
redis-cli -p 16379
```

或使用任何支持 Redis 协议的客户端。

### 基本使用示例

```bash
# PING 测试连接
127.0.0.1:16379> PING
PONG

# SET/GET 基本操作
127.0.0.1:16379> SET name "Alice"
OK

127.0.0.1:16379> GET name
"Alice"

# 查看所有键
127.0.0.1:16379> KEYS *
1) "name"

# 检查键是否存在
127.0.0.1:16379> EXISTS name
(integer) 1

# 删除键
127.0.0.1:16379> DEL name
(integer) 1

# 模式匹配
127.0.0.1:16379> SET user:1:name "Alice"
OK
127.0.0.1:16379> SET user:2:name "Bob"
OK
127.0.0.1:16379> KEYS user:*
1) "user:1:name"
2) "user:2:name"
```

---

## 🏗️ 架构演进路线

### 当前状态：Stage 0 - 单机内存存储架构 ✅

**已实现的架构能力**：

| 架构层次 | 设计模式 | 核心能力 |
|---------|---------|---------|
| **Server 层** | 多路复用 | TCP 服务、并发连接处理 |
| **Protocol 层** | 解析器模式 | RESP 协议解析和序列化 |
| **Handler 层** | 命令模式 | 可扩展的命令路由 |
| **Store 层** | 并发安全 | RWMutex 保护的 Map |

**支持的命令**（6 个核心命令即可）：
- `PING` `SET` `GET` `DEL` `EXISTS` `KEYS`

**架构局限**：
- ❌ 无持久化（数据易失）
- ❌ 单点故障（无容错）
- ❌ 单机瓶颈（无法扩展）

### 下一阶段：Stage 1 - 持久化架构设计

**架构主题**：数据可靠性（Reliability）

#### Phase 1.1: RDB 快照架构
- **架构模式**：Snapshot Pattern
- **核心概念**：Fork + COW、原子操作
- **学习重点**：理解快照机制、性能 vs 可靠性权衡
- **可迁移应用**：数据库备份、容器镜像、游戏存档

#### Phase 1.2: AOF 日志架构
- **架构模式**：Write-Ahead Logging
- **核心概念**：WAL、Log Compaction、fsync 策略
- **学习重点**：理解日志持久化、Durability vs Performance
- **可迁移应用**：MySQL binlog、Kafka log、文件系统

### 未来规划：Stage 2+

**Stage 2**：分布式高可用架构（主从复制 + 故障转移）
**Stage 3**：并发控制（事务、Pub/Sub）

详见 **[FOCUSED-ROADMAP.md](docs/FOCUSED-ROADMAP.md)** 和 **[ARCHITECTURE-GUIDE.md](docs/ARCHITECTURE-GUIDE.md)**

---

## 🏗️ 架构设计

### 系统架构

```
┌─────────────────────────────────────────────┐
│         客户端 (redis-cli / 应用)            │
└───────────────────┬─────────────────────────┘
                    │ TCP 连接
┌───────────────────▼─────────────────────────┐
│              服务器层 (Server)               │
│  ┌────────────────────────────────────────┐ │
│  │         TCP Listener (监听器)          │ │
│  └──────┬───────────────────────┬─────────┘ │
│         │                       │            │
│  ┌──────▼─────┐          ┌──────▼─────┐     │
│  │ Client 1   │   ...    │ Client N   │     │
│  │ (goroutine)│          │ (goroutine)│     │
│  └──────┬─────┘          └──────┬─────┘     │
└─────────┼────────────────────────┼───────────┘
          │                        │
┌─────────▼────────────────────────▼───────────┐
│             协议层 (Protocol)                │
│      Parser ←→ Router ←→ Serializer          │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│            命令处理层 (Handler)              │
│         Router + Command Handlers            │
└──────────────────┬───────────────────────────┘
                   │
┌──────────────────▼───────────────────────────┐
│              存储层 (Store)                  │
│         map[string]interface{}               │
└──────────────────────────────────────────────┘
```

### 分层架构

#### 1. 存储层 (Store)

**职责**：提供线程安全的键值存储

**核心组件**：
- `Store` 结构：使用 `sync.RWMutex` 保证并发安全
- 支持任意类型值（`interface{}`）

**关键方法**：
```go
Set(key string, value interface{})
Get(key string) (interface{}, bool)
Delete(key string) bool
Exists(key string) bool
Keys() []string
```

**位置**：`store/store.go`

#### 2. 协议层 (Protocol)

**职责**：RESP 协议的解析和序列化

**核心组件**：
- `Parser`：从 TCP 流中解析 RESP 数据
- `Serializer`：将数据结构序列化为 RESP 格式
- `Value`：RESP 数据类型的统一表示

**关键方法**：
```go
Parse() (*Value, error)           // 解析 RESP 请求
Serialize(value *Value) string    // 序列化响应
```

**位置**：`protocol/`

#### 3. 命令处理层 (Handler)

**职责**：命令路由和业务逻辑处理

**核心组件**：
- `Router`：命令分发器
- `Handler` 接口：命令处理器抽象
- 各种命令处理器：`PingHandler`, `SetHandler`, `GetHandler` 等

**关键方法**：
```go
Route(cmd *Value) *Value                    // 路由命令
Register(command string, handler Handler)   // 注册处理器
Handle(args []Value) *Value                 // 处理命令
```

**位置**：`handler/`

#### 4. 服务器层 (Server)

**职责**：TCP 网络服务和连接管理

**核心组件**：
- `Server`：TCP 服务器
- `Client`：客户端连接会话

**关键方法**：
```go
Start() error     // 启动服务器
Stop() error      // 优雅关闭
Serve()           // 客户端请求循环
```

**位置**：`server/`

### 数据流

```
客户端发送请求:
  SET name Alice
       ↓
TCP 连接 (Server.Accept)
       ↓
Client.Serve() 循环
       ↓
Parser.Parse() 解析 RESP
       ↓
Router.Route() 路由命令
       ↓
SetHandler.Handle() 执行命令
       ↓
Store.Set() 存储数据
       ↓
Serializer.Serialize() 序列化响应
       ↓
conn.Write() 发送响应
       ↓
客户端收到: +OK
```

---

## 📁 项目结构

```
go-redis/
├── main.go                    # 程序入口
├── go.mod                     # Go 模块定义
├── go.sum                     # 依赖校验和
├── README.md                  # 项目文档
│
├── docs/                      # 需求文档
│   ├── phase1-store.md        # 存储层需求
│   ├── phase2-protocol.md     # 协议层需求
│   ├── phase3-handler.md      # 处理层需求
│   └── phase4-server.md       # 服务器层需求
│
├── store/                     # 存储层
│   ├── store.go               # Store 实现
│   └── store_test.go          # 存储层测试
│
├── protocol/                  # 协议层
│   ├── parser.go              # RESP 解析器
│   ├── serializer.go          # RESP 序列化器
│   ├── types.go               # 数据类型定义
│   ├── helpers.go             # 辅助函数
│   ├── parser_test.go         # 解析器测试
│   └── serializer_test.go     # 序列化器测试
│
├── handler/                   # 命令处理层
│   ├── router.go              # 命令路由器
│   ├── ping.go                # PING 命令
│   ├── set.go                 # SET 命令
│   ├── get.go                 # GET 命令
│   ├── del.go                 # DEL 命令
│   ├── exists.go              # EXISTS 命令
│   ├── keys.go                # KEYS 命令
│   └── router_test.go         # 集成测试
│
├── server/                    # 服务器层
│   ├── server.go              # TCP 服务器
│   ├── client.go              # 客户端会话
│   ├── server_test.go         # 服务器测试 (待完善)
│   └── client_test.go         # 客户端测试 (待完善)
│
├── logger/                    # 日志工具
│   └── logger.go              # 日志封装
│
└── types/                     # 公共类型
    └── handler.go             # Handler 接口定义
```

---

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
go test ./... -v

# 运行特定包的测试
go test ./store -v
go test ./protocol -v
go test ./handler -v
go test ./server -v

# 查看测试覆盖率
go test ./... -cover

# 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 测试统计

| 模块 | 测试数量 | 覆盖率 |
|------|---------|--------|
| Store | 6+ | >90% |
| Protocol | 15+ | >85% |
| Handler | 27+ | >90% |
| Server | 待完善 | - |
| **总计** | **48+** | **>85%** |

### 使用 redis-benchmark 压测

```bash
# 安装 redis-benchmark（通常随 redis 安装）
brew install redis

# 压力测试（10 万请求，50 并发）
redis-benchmark -p 16379 -t set,get -n 100000 -c 50

# 测试特定命令
redis-benchmark -p 16379 -t ping -n 100000
```

---

## 📊 性能

### 基准测试结果

> 测试环境：MacBook Pro M1, 16GB RAM, Go 1.23

| 操作 | 吞吐量 (ops/sec) | 延迟 (ms) |
|------|-----------------|-----------|
| PING | ~80,000 | < 0.1 |
| SET | ~70,000 | < 0.2 |
| GET | ~75,000 | < 0.15 |
| DEL | ~70,000 | < 0.2 |

**注**：以上为初步测试数据，实际性能取决于硬件和工作负载。

### 性能优化点

- ✅ 使用 `sync.RWMutex` 提升读并发性能
- ✅ 每个客户端独立 goroutine，避免阻塞
- ✅ 零拷贝序列化（直接操作字节）
- 📋 待优化：连接池、内存池、批处理

---

## 🛠️ 开发指南

### 添加新命令

1. **创建 Handler 文件** (`handler/mycommand.go`)：

```go
package handler

import (
    "go-redis/protocol"
    "go-redis/store"
)

type MyCommandHandler struct {
    db *store.Store
}

func NewMyCommandHandler(db *store.Store) *MyCommandHandler {
    return &MyCommandHandler{db: db}
}

func (h *MyCommandHandler) Handle(args []protocol.Value) *protocol.Value {
    // 参数验证
    if len(args) != 1 {
        return protocol.Error("ERR wrong number of arguments for 'mycommand' command")
    }

    // 业务逻辑
    // ...

    return protocol.SimpleString("OK")
}
```

2. **注册到 Router** (`handler/router.go`):

```go
func (r *Router) registerDefaultHandlers() {
    // ... 其他命令
    r.Register("MYCOMMAND", NewMyCommandHandler(r.db))
}
```

3. **编写测试** (`handler/router_test.go`):

```go
func TestMyCommandHandler(t *testing.T) {
    s := store.NewStore()
    r := NewRouter(s)

    cmdResp := "*2\r\n$9\r\nMYCOMMAND\r\n$3\r\narg\r\n"
    reader := strings.NewReader(cmdResp)
    p := protocol.NewParser(reader)
    cmd, _ := p.Parse()

    resp := r.Route(cmd)
    if resp.Str != "OK" {
        t.Error("Expected OK")
    }
}
```

### 代码规范

- 使用 `go fmt` 格式化代码
- 使用 `go vet` 进行静态检查
- 遵循 Go 命名约定
- 为所有公开函数添加文档注释
- 单元测试覆盖率 > 80%

### 提交规范

```bash
# 提交格式
<type>(<scope>): <subject>

# 示例
feat(handler): add INCR command
fix(server): fix connection leak on shutdown
docs(readme): update installation guide
test(protocol): add parser edge case tests
```

类型：
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `test`: 测试相关
- `refactor`: 重构
- `perf`: 性能优化
- `chore`: 构建/工具相关

---

## 📖 学习方法

### 架构设计思维框架

每个阶段学习时，遵循这个思维框架：

```
1. 问题定义
   - 当前架构的局限是什么？
   - 要解决什么具体问题？

2. 方案对比
   - 有哪些可选方案？
   - 每个方案的 Trade-offs 是什么？

3. 设计决策
   - 为什么选择这个方案？
   - 关键设计点有哪些？

4. 实现验证
   - 如何验证设计是正确的？
   - 性能测试、故障演练

5. 总结迁移
   - 学到了什么架构模式？
   - 如何应用到其他系统？
```

### 推荐学习流程

**不要**：
- ❌ 直接看代码就开始写
- ❌ 追求功能完整性
- ❌ 孤立学习（不思考可迁移性）

**应该**：
- ✅ 先读 [ARCHITECTURE-GUIDE.md](docs/ARCHITECTURE-GUIDE.md) 理解架构思维
- ✅ 再读 [FOCUSED-ROADMAP.md](docs/FOCUSED-ROADMAP.md) 规划学习路径
- ✅ 每个阶段先画架构图，再写代码
- ✅ 写设计文档记录 Trade-offs
- ✅ 性能测试验证设计假设
- ✅ 对比其他系统（如何应用到 Kafka、PostgreSQL）

### 每个阶段的产出

**代码**：
- 核心功能实现（最小化，够用即可）
- 单元测试 + 集成测试
- 性能基准测试

**文档**：
- 架构设计文档（问题、方案、Trade-offs）
- 性能测试报告
- 学习笔记（可迁移的知识点）

---

## 📚 学习资源

### 必读文档

1. **[ARCHITECTURE-GUIDE.md](docs/ARCHITECTURE-GUIDE.md)** - 系统架构设计思维框架
   - 架构模式详解（Snapshot、WAL、Replication）
   - Trade-offs 分析方法
   - CAP 定理、设计原则

2. **[FOCUSED-ROADMAP.md](docs/FOCUSED-ROADMAP.md)** - 架构演进学习路线
   - Stage 1: 持久化架构（RDB、AOF）
   - Stage 2: 分布式高可用架构
   - Stage 3: 并发控制

### 实现参考文档

- [Phase 1: 存储层](docs/phase1-store.md)
- [Phase 2: 协议层](docs/phase2-protocol.md)
- [Phase 3: 命令处理层](docs/phase3-handler.md)
- [Phase 4: 服务器层](docs/phase4-server.md)

### 外部资源

**书籍**（强烈推荐）：
- 《Designing Data-Intensive Applications》- Martin Kleppmann
  - 第 3 章：Storage and Retrieval
  - 第 5 章：Replication
  - 第 7 章：Transactions

- 《Redis 设计与实现》- 黄健宏
  - 第 9-11 章：持久化
  - 第 15-16 章：复制和哨兵

**Redis 官方文档**：
- [Redis Persistence](https://redis.io/docs/management/persistence/)
- [Redis Replication](https://redis.io/docs/management/replication/)
- [Redis Protocol](https://redis.io/docs/reference/protocol-spec/)

**源码阅读**（选读）：
- [Redis 源码](https://github.com/redis/redis)
  - `rdb.c` - RDB 实现
  - `aof.c` - AOF 实现
  - `replication.c` - 复制协议

---

## 🙏 致谢

- 感谢 [Redis](https://redis.io/) 项目提供的优秀设计
- 感谢所有贡献者和问题报告者

---

## 📄 许可证

本项目采用 MIT 许可证

---

<div align="center">

**⭐️ 如果这个项目对你有帮助，请给个 Star！**

Made with ❤️ for Learning

</div>

# Go-Redis：简易 Redis 实现学习项目

## 项目概述

这是一个用于学习的简易 Redis 实现项目，采用测试驱动开发（TDD）方式，帮助你理解 Redis 的核心设计思想和实现原理。

### 学习目标

- 理解 Redis 的基本数据结构和操作
- 掌握 TCP 服务器的设计与实现
- 学习 RESP 协议（Redis 序列化协议）
- 理解并发安全的数据存储
- 实践测试驱动开发（TDD）

### 项目特点

- 简化版实现，专注核心概念
- 测试先行，每个功能都有对应测试
- 渐进式开发，分阶段完成
- 代码清晰，注重可读性

## 核心设计思路

### 1. 整体架构

```
┌─────────────────┐
│   Client        │
└────────┬────────┘
         │ TCP + RESP Protocol
┌────────▼────────────────────┐
│   Protocol Layer            │
│  (RESP Parser/Serializer)   │
└────────┬────────────────────┘
         │
┌────────▼────────────────────┐
│   Command Handler           │
│  (Route & Execute)          │
└────────┬────────────────────┘
         │
┌────────▼────────────────────┐
│   Data Store                │
│  (Thread-safe Storage)      │
└─────────────────────────────┘
```

### 2. 核心模块

#### 模块 1：数据存储层 (Store)
- **职责**：存储键值对数据
- **设计要点**：
  - 使用 `map[string]interface{}` 存储数据
  - 使用 `sync.RWMutex` 保证并发安全
  - 支持多种数据类型（String, List, Hash, Set）

#### 模块 2：协议层 (Protocol)
- **职责**：解析和序列化 RESP 协议
- **设计要点**：
  - RESP 是基于文本的协议，易于调试
  - 支持简单字符串、错误、整数、批量字符串、数组
  - 使用 `bufio.Reader` 进行高效解析

#### 模块 3：命令处理层 (Handler)
- **职责**：处理各种 Redis 命令
- **设计要点**：
  - 使用命令模式，每个命令一个处理器
  - 命令路由：根据命令名称分发到对应处理器
  - 返回统一的响应格式

#### 模块 4：服务器层 (Server)
- **职责**：TCP 服务器，处理客户端连接
- **设计要点**：
  - 使用 `net.Listen` 监听端口
  - 为每个客户端连接启动 goroutine
  - 优雅关闭机制

## TDD 开发路线图

### 第一阶段：数据存储层（Day 1-2）

**目标**：实现线程安全的内存数据库

**测试先行**：
```go
// store_test.go 中应该测试的功能：
- TestSetAndGet：基本的设置和获取
- TestGetNonExistent：获取不存在的键
- TestDelete：删除键
- TestConcurrentAccess：并发读写安全性
- TestExpiration：键过期功能（可选）
```

**设计提示**：
- 创建 `Store` 结构体
- 实现方法：`Set(key, value)`, `Get(key)`, `Delete(key)`, `Exists(key)`
- 注意：先写测试，看测试失败，再实现代码让测试通过

---

### 第二阶段：RESP 协议解析（Day 3-4）

**目标**：实现 Redis 序列化协议的解析和序列化

**RESP 协议简介**：
- 简单字符串：`+OK\r\n`
- 错误：`-Error message\r\n`
- 整数：`:1000\r\n`
- 批量字符串：`$6\r\nfoobar\r\n`
- 数组：`*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n`

**测试先行**：
```go
// resp_test.go 中应该测试的功能：
- TestParseSimpleString：解析简单字符串
- TestParseError：解析错误信息
- TestParseInteger：解析整数
- TestParseBulkString：解析批量字符串
- TestParseArray：解析数组
- TestSerializeResponse：序列化响应
```

**设计提示**：
- 创建 `RESPParser` 结构体
- 实现 `Parse(reader)` 方法返回解析后的值
- 实现 `Serialize(value)` 方法将值序列化为 RESP 格式
- 使用递归处理嵌套数组

---

### 第三阶段：命令处理器（Day 5-6）

**目标**：实现 Redis 基本命令处理

**支持的命令**：
- `PING`：测试连接
- `SET key value`：设置键值
- `GET key`：获取值
- `DEL key [key ...]`：删除键
- `EXISTS key`：检查键是否存在
- `KEYS pattern`：查找键（简化版，只支持 `*`）

**测试先行**：
```go
// handler_test.go 中应该测试的功能：
- TestPingCommand：测试 PING 命令
- TestSetCommand：测试 SET 命令
- TestGetCommand：测试 GET 命令
- TestDelCommand：测试 DEL 命令
- TestUnknownCommand：测试未知命令处理
```

**设计提示**：
- 创建 `Handler` 结构体，持有 `Store` 引用
- 实现 `Execute(cmd []string)` 方法
- 使用 `switch` 或 `map` 进行命令路由
- 返回 RESP 格式的响应

---

### 第四阶段：TCP 服务器（Day 7-8）

**目标**：实现完整的 TCP 服务器，接受客户端连接

**测试先行**：
```go
// server_test.go 中应该测试的功能：
- TestServerStartStop：服务器启动和停止
- TestClientConnection：客户端连接
- TestCommandExecution：通过网络执行命令
- TestMultipleClients：多个客户端并发连接
```

**设计提示**：
- 创建 `Server` 结构体
- 实现 `Start(address)` 方法监听端口
- 为每个连接创建 goroutine 处理请求
- 实现 `handleConnection(conn)` 方法
- 实现优雅关闭：使用 context 或 channel

---

### 第五阶段：高级数据结构（Day 9-10，可选）

**目标**：支持 List 和 Hash 数据类型

**List 命令**：
- `LPUSH key value [value ...]`：从左侧插入
- `RPUSH key value [value ...]`：从右侧插入
- `LPOP key`：从左侧弹出
- `RPOP key`：从右侧弹出
- `LRANGE key start stop`：获取范围内的元素

**Hash 命令**：
- `HSET key field value`：设置字段
- `HGET key field`：获取字段
- `HDEL key field [field ...]`：删除字段
- `HGETALL key`：获取所有字段

**测试先行**：
```go
// list_test.go 和 hash_test.go
- 测试各个命令的基本功能
- 测试边界条件
- 测试类型错误（对 String 执行 List 命令）
```

**设计提示**：
- 扩展 `Store` 以支持类型检查
- 内部使用 `[]interface{}` 表示 List
- 内部使用 `map[string]string` 表示 Hash

---

## 项目结构

```
go-redis/
├── README.md                 # 本文档
├── go.mod                    # Go 模块定义
├── main.go                   # 程序入口
├── store/
│   ├── store.go             # 数据存储实现
│   └── store_test.go        # 数据存储测试
├── protocol/
│   ├── resp.go              # RESP 协议实现
│   └── resp_test.go         # RESP 协议测试
├── handler/
│   ├── handler.go           # 命令处理器
│   ├── string_commands.go   # String 命令实现
│   ├── list_commands.go     # List 命令实现（可选）
│   ├── hash_commands.go     # Hash 命令实现（可选）
│   └── handler_test.go      # 命令处理器测试
└── server/
    ├── server.go            # TCP 服务器
    └── server_test.go       # 服务器测试
```

## TDD 开发流程

每个阶段遵循以下流程：

### 🔴 Red（红灯）
1. 写一个失败的测试
2. 运行测试，确保它失败
3. 思考：这个测试验证了什么？

### 🟢 Green（绿灯）
4. 写最简单的代码让测试通过
5. 运行测试，确保它通过
6. 不要过度设计，只需让测试通过

### 🔵 Refactor（重构）
7. 改进代码质量
8. 消除重复
9. 运行测试，确保仍然通过

**重复这个循环，直到功能完成！**

## 快速开始

### 1. 初始化项目

```bash
# 创建 Go 模块
go mod init github.com/yourusername/go-redis

# 创建目录结构
mkdir -p store protocol handler server
```

### 2. 开始第一阶段

```bash
# 创建测试文件
touch store/store_test.go

# 编写第一个测试（参考第一阶段的测试提示）
# 运行测试
go test ./store -v

# 测试失败后，创建实现文件
touch store/store.go

# 实现代码，让测试通过
go test ./store -v
```

### 3. 测试你的 Redis

完成所有阶段后，可以使用官方 Redis 客户端测试：

```bash
# 启动你的 Redis 服务器
go run main.go

# 在另一个终端，使用 redis-cli 连接
redis-cli -p 6380

# 测试命令
127.0.0.1:6380> PING
PONG
127.0.0.1:6380> SET name "GoRedis"
OK
127.0.0.1:6380> GET name
"GoRedis"
```

## 学习建议

1. **严格遵循 TDD**：先写测试再写实现，不要跳过
2. **小步前进**：每次只实现一个小功能
3. **频繁运行测试**：确保每次改动都不破坏已有功能
4. **理解设计**：每个模块都有明确的职责，保持代码简洁
5. **参考资料**：
   - [Redis 协议规范](https://redis.io/docs/reference/protocol-spec/)
   - [Redis 命令参考](https://redis.io/commands/)
   - Go 标准库文档：`net`, `bufio`, `sync`

## 扩展思考

完成基础功能后，可以思考以下问题：

- 如何实现数据持久化（AOF 或 RDB）？
- 如何实现键过期机制？
- 如何实现事务（MULTI/EXEC）？
- 如何优化内存使用？
- 如何实现 Pub/Sub？
- 如何进行性能测试和优化？

## 总结

这个项目通过渐进式开发，让你深入理解 Redis 的核心设计。记住：

- **测试是你的安全网**：测试给你重构的信心
- **简单优于复杂**：先让它工作，再让它优雅
- **理解优于记忆**：理解每个设计决策的原因

祝你学习愉快！🚀

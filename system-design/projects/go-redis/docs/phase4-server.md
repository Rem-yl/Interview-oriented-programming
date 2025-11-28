# 第四阶段：服务器层需求文档

## 1. 需求概述

实现 Redis 服务器层（Server），负责监听 TCP 连接、管理客户端会话、协调协议层和命令处理层的工作。该层是整个 Redis 服务的入口，实现了完整的客户端-服务器通信。

### 1.1 业务背景

在完成了存储层（Phase 1）、协议层（Phase 2）和命令处理层（Phase 3）后，我们需要一个服务器层来：
- 监听 TCP 端口，接受客户端连接
- 为每个客户端创建独立的会话
- 读取客户端请求，调用 Parser 解析
- 将解析后的命令路由到 Handler 执行
- 将执行结果通过 Serializer 序列化后返回客户端
- 处理并发连接和优雅关闭

### 1.2 核心目标

- 实现基于 TCP 的 Redis 服务器
- 支持多客户端并发连接
- 实现完整的请求-响应循环
- 优雅的启动和关闭机制
- 完善的错误处理和日志记录
- 可以被真实的 `redis-cli` 客户端连接

---

## 2. 系统架构

### 2.1 整体架构

```
┌─────────────────────────────────────────────┐
│         客户端 (redis-cli / 应用)            │
└───────────────────┬─────────────────────────┘
                    │ TCP 连接
┌───────────────────▼─────────────────────────┐
│              服务器层 (Server)    ← 本阶段   │
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
└──────────────────────────────────────────────┘
```

### 2.2 核心组件

#### Server 结构

```go
// Server Redis 服务器
type Server struct {
    addr     string           // 监听地址，如 ":6379"
    listener net.Listener     // TCP 监听器
    router   *handler.Router  // 命令路由器
    store    *store.Store     // 数据存储
    clients  sync.Map         // 客户端连接映射
    shutdown chan struct{}    // 关闭信号
    wg       sync.WaitGroup   // 等待所有连接关闭
}

// NewServer 创建新的服务器
func NewServer(addr string, s *store.Store) *Server

// Start 启动服务器
func (s *Server) Start() error

// Stop 优雅关闭服务器
func (s *Server) Stop() error
```

#### Client 结构

```go
// Client 客户端连接
type Client struct {
    id       string           // 客户端 ID
    conn     net.Conn         // TCP 连接
    parser   *protocol.Parser // RESP 解析器
    router   *handler.Router  // 命令路由器
    shutdown chan struct{}    // 关闭信号
}

// NewClient 创建新的客户端
func NewClient(conn net.Conn, router *handler.Router) *Client

// Serve 处理客户端请求
func (c *Client) Serve()

// Close 关闭客户端连接
func (c *Client) Close() error
```

---

## 3. 功能需求

### 3.1 核心功能清单

| 功能 | 描述 | 优先级 |
|------|------|--------|
| TCP 监听 | 监听指定端口，接受客户端连接 | P0 |
| 并发处理 | 每个客户端连接独立的 goroutine | P0 |
| 请求解析 | 使用 Parser 解析 RESP 请求 | P0 |
| 命令执行 | 调用 Router 执行命令 | P0 |
| 响应序列化 | 使用 Serializer 序列化响应 | P0 |
| 错误处理 | 捕获并返回错误信息 | P0 |
| 优雅关闭 | 正确关闭所有连接和监听器 | P0 |
| 日志记录 | 记录连接、请求、错误等信息 | P1 |
| 连接超时 | 支持读写超时设置 | P1 |
| 最大连接数 | 限制并发连接数量 | P2 |

### 3.2 详细功能规格

#### 3.2.1 服务器启动流程

```
1. 创建 Store 实例
2. 创建 Router 实例并绑定 Store
3. 创建 Server 实例，指定监听地址
4. 调用 server.Start()
   ├─ 启动 TCP 监听器
   ├─ 记录启动日志
   ├─ 进入 Accept 循环
   └─ 等待客户端连接
```

**示例代码**：

```go
func main() {
    // 1. 创建存储
    s := store.NewStore()

    // 2. 创建服务器
    server := NewServer(":6379", s)

    // 3. 启动服务器
    logger.Info("Starting Redis server on :6379")
    if err := server.Start(); err != nil {
        logger.Fatalf("Failed to start server: %v", err)
    }
}
```

#### 3.2.2 客户端连接处理流程

```
1. Accept 客户端连接
2. 创建 Client 实例
   ├─ 生成唯一 ID
   ├─ 创建 Parser (基于连接的 Reader)
   └─ 绑定 Router
3. 启动 goroutine 执行 client.Serve()
4. 将 Client 添加到 clients 映射
5. 继续 Accept 下一个连接
```

#### 3.2.3 请求-响应循环

```
Client.Serve():
  Loop:
    1. 调用 parser.Parse() 读取并解析请求
       ├─ 如果 EOF → 客户端断开 → break Loop
       ├─ 如果 ParseError → 返回错误响应 → continue
       └─ 解析成功 → 继续

    2. 调用 router.Route(cmd) 执行命令
       └─ 返回 RESP Value

    3. 调用 protocol.Serialize(value) 序列化响应

    4. 写入响应到 conn.Write()
       ├─ 如果写入失败 → 记录错误 → break Loop
       └─ 写入成功 → continue Loop

  End Loop:
    5. 从 clients 映射中移除
    6. 关闭连接
    7. 记录日志
```

#### 3.2.4 优雅关闭流程

```
Server.Stop():
  1. 关闭 shutdown channel
  2. 停止接受新连接 (listener.Close())
  3. 遍历所有 clients，调用 client.Close()
     ├─ 发送关闭信号
     └─ 关闭 TCP 连接
  4. 等待所有客户端 goroutine 退出 (wg.Wait())
  5. 记录关闭日志
```

---

## 4. 架构设计

### 4.1 目录结构

```
server/
├── server.go           # Server 结构和方法
├── client.go           # Client 结构和方法
├── server_test.go      # 服务器测试
└── client_test.go      # 客户端测试

main.go                 # 程序入口
```

### 4.2 核心实现

#### 4.2.1 Server 实现

```go
package server

import (
    "fmt"
    "go-redis/handler"
    "go-redis/logger"
    "go-redis/store"
    "net"
    "sync"
    "sync/atomic"
)

// Server Redis 服务器
type Server struct {
    addr       string
    listener   net.Listener
    router     *handler.Router
    store      *store.Store
    clients    sync.Map // map[string]*Client
    shutdown   chan struct{}
    wg         sync.WaitGroup
    clientID   int64 // 原子计数器，用于生成客户端 ID
}

// NewServer 创建新的服务器
func NewServer(addr string, s *store.Store) *Server {
    router := handler.NewRouter(s)

    return &Server{
        addr:     addr,
        router:   router,
        store:    s,
        shutdown: make(chan struct{}),
    }
}

// Start 启动服务器
func (s *Server) Start() error {
    // 监听 TCP 端口
    listener, err := net.Listen("tcp", s.addr)
    if err != nil {
        return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
    }
    s.listener = listener

    logger.Infof("Redis server listening on %s", s.addr)

    // Accept 循环
    for {
        conn, err := listener.Accept()
        if err != nil {
            // 检查是否是因为关闭导致的错误
            select {
            case <-s.shutdown:
                logger.Info("Server is shutting down")
                return nil
            default:
                logger.Errorf("Failed to accept connection: %v", err)
                continue
            }
        }

        // 为每个连接创建客户端并启动 goroutine
        client := NewClient(conn, s.router, s.nextClientID())
        s.clients.Store(client.id, client)

        s.wg.Add(1)
        go func() {
            defer s.wg.Done()
            client.Serve()
            s.clients.Delete(client.id)
        }()
    }
}

// Stop 优雅关闭服务器
func (s *Server) Stop() error {
    logger.Info("Stopping server...")

    // 1. 关闭 shutdown channel
    close(s.shutdown)

    // 2. 停止接受新连接
    if s.listener != nil {
        s.listener.Close()
    }

    // 3. 关闭所有客户端连接
    s.clients.Range(func(key, value interface{}) bool {
        client := value.(*Client)
        client.Close()
        return true
    })

    // 4. 等待所有客户端 goroutine 退出
    s.wg.Wait()

    logger.Info("Server stopped")
    return nil
}

// nextClientID 生成下一个客户端 ID
func (s *Server) nextClientID() string {
    id := atomic.AddInt64(&s.clientID, 1)
    return fmt.Sprintf("client-%d", id)
}
```

#### 4.2.2 Client 实现

```go
package server

import (
    "go-redis/handler"
    "go-redis/logger"
    "go-redis/protocol"
    "io"
    "net"
)

// Client 客户端连接
type Client struct {
    id       string
    conn     net.Conn
    parser   *protocol.Parser
    router   *handler.Router
    shutdown chan struct{}
}

// NewClient 创建新的客户端
func NewClient(conn net.Conn, router *handler.Router, id string) *Client {
    return &Client{
        id:       id,
        conn:     conn,
        parser:   protocol.NewParser(conn),
        router:   router,
        shutdown: make(chan struct{}),
    }
}

// Serve 处理客户端请求
func (c *Client) Serve() {
    logger.Infof("[%s] Client connected from %s", c.id, c.conn.RemoteAddr())
    defer logger.Infof("[%s] Client disconnected", c.id)
    defer c.conn.Close()

    for {
        // 检查是否收到关闭信号
        select {
        case <-c.shutdown:
            return
        default:
        }

        // 1. 解析请求
        cmd, err := c.parser.Parse()
        if err != nil {
            if err == io.EOF {
                // 客户端正常断开
                return
            }

            // 解析错误，返回错误响应
            logger.Errorf("[%s] Parse error: %v", c.id, err)
            errorResp := protocol.Error(fmt.Sprintf("ERR parse error: %v", err))
            c.sendResponse(errorResp)
            continue
        }

        logger.Debugf("[%s] Received command: %+v", c.id, cmd)

        // 2. 执行命令
        response := c.router.Route(cmd)

        // 3. 发送响应
        if err := c.sendResponse(response); err != nil {
            logger.Errorf("[%s] Failed to send response: %v", c.id, err)
            return
        }
    }
}

// sendResponse 发送响应到客户端
func (c *Client) sendResponse(value *protocol.Value) error {
    // 序列化响应
    data := protocol.Serialize(value)

    // 写入连接
    _, err := c.conn.Write([]byte(data))
    if err != nil {
        return err
    }

    logger.Debugf("[%s] Sent response: %s", c.id, data)
    return nil
}

// Close 关闭客户端连接
func (c *Client) Close() error {
    close(c.shutdown)
    return c.conn.Close()
}
```

#### 4.2.3 Main 入口

```go
package main

import (
    "go-redis/logger"
    "go-redis/server"
    "go-redis/store"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // 设置日志级别
    logger.SetLevel(logger.InfoLevel)

    // 创建存储
    s := store.NewStore()

    // 创建服务器
    srv := server.NewServer(":6379", s)

    // 优雅关闭处理
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh

        logger.Info("Received shutdown signal")
        srv.Stop()
        os.Exit(0)
    }()

    // 启动服务器
    logger.Info("Starting Redis server on :6379")
    if err := srv.Start(); err != nil {
        logger.Fatalf("Server error: %v", err)
    }
}
```

---

## 5. 测试计划

### 5.1 单元测试

#### 5.1.1 服务器测试

```go
package server

import (
    "go-redis/store"
    "net"
    "testing"
    "time"
)

func TestServerStartStop(t *testing.T) {
    s := store.NewStore()
    srv := NewServer(":16379", s)

    // 启动服务器
    go func() {
        if err := srv.Start(); err != nil {
            t.Errorf("Failed to start server: %v", err)
        }
    }()

    // 等待服务器启动
    time.Sleep(100 * time.Millisecond)

    // 停止服务器
    if err := srv.Stop(); err != nil {
        t.Errorf("Failed to stop server: %v", err)
    }
}

func TestServerAcceptConnection(t *testing.T) {
    s := store.NewStore()
    srv := NewServer(":16380", s)

    // 启动服务器
    go srv.Start()
    time.Sleep(100 * time.Millisecond)
    defer srv.Stop()

    // 连接到服务器
    conn, err := net.Dial("tcp", ":16380")
    if err != nil {
        t.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    // 验证连接成功
    if conn == nil {
        t.Error("Connection is nil")
    }
}
```

### 5.2 集成测试

#### 5.2.1 完整请求-响应测试

```go
func TestServerPingPong(t *testing.T) {
    s := store.NewStore()
    srv := NewServer(":16381", s)

    go srv.Start()
    time.Sleep(100 * time.Millisecond)
    defer srv.Stop()

    // 连接
    conn, err := net.Dial("tcp", ":16381")
    if err != nil {
        t.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    // 发送 PING 命令
    request := "*1\r\n$4\r\nPING\r\n"
    _, err = conn.Write([]byte(request))
    if err != nil {
        t.Fatalf("Failed to write: %v", err)
    }

    // 读取响应
    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        t.Fatalf("Failed to read: %v", err)
    }

    response := string(buffer[:n])
    expected := "+PONG\r\n"
    if response != expected {
        t.Errorf("Expected %q, got %q", expected, response)
    }
}
```

#### 5.2.2 SET/GET 测试

```go
func TestServerSetGet(t *testing.T) {
    s := store.NewStore()
    srv := NewServer(":16382", s)

    go srv.Start()
    time.Sleep(100 * time.Millisecond)
    defer srv.Stop()

    conn, _ := net.Dial("tcp", ":16382")
    defer conn.Close()

    // SET key value
    setCmd := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAlice\r\n"
    conn.Write([]byte(setCmd))

    buffer := make([]byte, 1024)
    n, _ := conn.Read(buffer)
    if string(buffer[:n]) != "+OK\r\n" {
        t.Error("SET failed")
    }

    // GET key
    getCmd := "*2\r\n$3\r\nGET\r\n$4\r\nname\r\n"
    conn.Write([]byte(getCmd))

    n, _ = conn.Read(buffer)
    expected := "$5\r\nAlice\r\n"
    if string(buffer[:n]) != expected {
        t.Errorf("Expected %q, got %q", expected, string(buffer[:n]))
    }
}
```

### 5.3 并发测试

```go
func TestServerConcurrentClients(t *testing.T) {
    s := store.NewStore()
    srv := NewServer(":16383", s)

    go srv.Start()
    time.Sleep(100 * time.Millisecond)
    defer srv.Stop()

    // 启动 10 个并发客户端
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            conn, err := net.Dial("tcp", ":16383")
            if err != nil {
                t.Errorf("Client %d: failed to connect: %v", id, err)
                return
            }
            defer conn.Close()

            // 发送 PING
            conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))

            buffer := make([]byte, 1024)
            n, _ := conn.Read(buffer)

            if string(buffer[:n]) != "+PONG\r\n" {
                t.Errorf("Client %d: unexpected response", id)
            }
        }(i)
    }

    wg.Wait()
}
```

### 5.4 使用真实 redis-cli 测试

```bash
# 启动服务器
go run main.go

# 在另一个终端，使用 redis-cli 连接
redis-cli -p 6379

# 测试命令
127.0.0.1:6379> PING
PONG

127.0.0.1:6379> SET name Alice
OK

127.0.0.1:6379> GET name
"Alice"

127.0.0.1:6379> KEYS *
1) "name"

127.0.0.1:6379> DEL name
(integer) 1

127.0.0.1:6379> EXISTS name
(integer) 0
```

---

## 6. 验收标准

### 6.1 功能验收

- [ ] 服务器能够成功启动并监听指定端口
- [ ] 能够接受多个客户端并发连接
- [ ] 正确处理 PING、SET、GET、DEL、EXISTS、KEYS 命令
- [ ] 错误命令返回正确的错误信息
- [ ] 客户端断开连接时正确清理资源
- [ ] 服务器能够优雅关闭
- [ ] 所有单元测试通过
- [ ] 所有集成测试通过

### 6.2 质量验收

- [ ] 代码通过 `go fmt` 格式化
- [ ] 代码通过 `go vet` 静态检查
- [ ] 无 goroutine 泄漏
- [ ] 无 TCP 连接泄漏
- [ ] 完整的日志记录
- [ ] 完整的文档注释

### 6.3 兼容性验收

- [ ] 可以被 `redis-cli` 客户端连接
- [ ] RESP 协议完全兼容
- [ ] 响应格式符合 Redis 标准
- [ ] 与前三个阶段正确集成

---

## 7. 实现提示

### 7.1 开发顺序建议

1. **创建基础结构**
   - `server/server.go` - Server 结构定义
   - `server/client.go` - Client 结构定义

2. **实现服务器核心**
   - Server.Start() - 监听和 Accept 循环
   - Server.Stop() - 优雅关闭

3. **实现客户端处理**
   - Client.Serve() - 请求-响应循环
   - Client.Close() - 关闭连接

4. **集成测试**
   - 单元测试
   - 集成测试
   - redis-cli 测试

5. **实现 main 入口**
   - main.go - 程序入口
   - 信号处理

### 7.2 关键技术点

#### 7.2.1 优雅关闭

```go
// 使用 sync.WaitGroup 等待所有 goroutine 退出
s.wg.Add(1)
go func() {
    defer s.wg.Done()
    client.Serve()
}()

// 关闭时等待
s.wg.Wait()
```

#### 7.2.2 并发安全的客户端映射

```go
// 使用 sync.Map 存储客户端
var clients sync.Map

// 添加
clients.Store(id, client)

// 删除
clients.Delete(id)

// 遍历
clients.Range(func(key, value interface{}) bool {
    // ...
    return true
})
```

#### 7.2.3 信号处理

```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
<-sigCh

// 优雅关闭
server.Stop()
```

#### 7.2.4 错误处理

```go
// 区分不同类型的错误
if err == io.EOF {
    // 客户端正常断开
    return
}

// 检查是否是关闭导致的错误
select {
case <-s.shutdown:
    return nil
default:
    logger.Error(err)
}
```

### 7.3 常见陷阱

1. **Goroutine 泄漏**
   - 每个客户端 goroutine 必须正确退出
   - 使用 defer 和 sync.WaitGroup

2. **连接泄漏**
   - 每个 Accept 的连接必须在某处 Close
   - 使用 defer conn.Close()

3. **关闭顺序**
   - 先停止 listener（不接受新连接）
   - 再关闭所有客户端
   - 最后等待所有 goroutine 退出

4. **并发写入同一连接**
   - net.Conn 的 Write 不是并发安全的
   - 每个客户端单独的 goroutine 避免了这个问题

5. **阻塞读取**
   - Parser.Parse() 会阻塞等待数据
   - 关闭连接会导致 io.EOF 错误，正确处理

---

## 8. 调试技巧

### 8.1 日志调试

```go
// 在关键位置添加日志
logger.Infof("[%s] Client connected from %s", c.id, c.conn.RemoteAddr())
logger.Debugf("[%s] Received command: %+v", c.id, cmd)
logger.Debugf("[%s] Sent response: %s", c.id, data)
logger.Infof("[%s] Client disconnected", c.id)
```

### 8.2 网络调试

```bash
# 查看服务器是否在监听
netstat -an | grep 6379

# 使用 telnet 测试
telnet localhost 6379
*1
$4
PING

# 使用 nc 测试
echo -e "*1\r\n\$4\r\nPING\r\n" | nc localhost 6379

# 使用 redis-cli
redis-cli -p 6379 PING
```

### 8.3 并发调试

```bash
# 运行测试时检查竞态条件
go test ./server -race

# 查看 goroutine 数量
import "runtime"
fmt.Println("Goroutines:", runtime.NumGoroutine())
```

### 8.4 性能分析

```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# 内存 profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# 使用 pprof HTTP 服务
import _ "net/http/pprof"
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

# 访问 http://localhost:6060/debug/pprof/
```

---

## 9. 扩展思考

完成基础功能后，可以思考：

1. **连接超时**
   - 如何实现读写超时？
   - 如何处理空闲连接？

2. **最大连接数限制**
   - 如何限制并发连接数？
   - 如何拒绝超出限制的连接？

3. **连接池**
   - 是否需要实现连接池？
   - 如何复用连接？

4. **TLS 支持**
   - 如何支持加密连接？
   - 如何配置证书？

5. **性能优化**
   - 如何减少内存分配？
   - 如何优化序列化/反序列化？
   - 是否需要零拷贝？

6. **监控指标**
   - 如何统计连接数、请求数、错误数？
   - 如何暴露 metrics 端点？

---

## 10. 参考资料

- [Go net 包文档](https://pkg.go.dev/net)
- [Redis Protocol Specification](https://redis.io/docs/reference/protocol-spec/)
- [Go 并发编程](https://go.dev/doc/effective_go#concurrency)
- [TCP Socket 编程](https://pkg.go.dev/net#TCPConn)

---

## 11. 交付物

完成本阶段后，应该交付：

1. [ ] `server/server.go` - Server 实现
2. [ ] `server/client.go` - Client 实现
3. [ ] `server/server_test.go` - 服务器测试
4. [ ] `server/client_test.go` - 客户端测试
5. [ ] `main.go` - 程序入口
6. [ ] 所有测试通过的截图或日志
7. [ ] 使用 redis-cli 测试的截图或日志
8. [ ] README.md 更新（包含启动和使用说明）

完成后，你将拥有一个**完整可用的 Redis 服务器**，可以：
- 使用 `go run main.go` 启动
- 使用 `redis-cli -p 6379` 连接
- 执行所有已实现的命令（PING, SET, GET, DEL, EXISTS, KEYS）
- 支持多客户端并发连接
- 优雅关闭和资源清理

---

## 附录：完整 main.go 示例

```go
package main

import (
    "flag"
    "fmt"
    "go-redis/logger"
    "go-redis/server"
    "go-redis/store"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // 命令行参数
    port := flag.Int("port", 6379, "Port to listen on")
    logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
    flag.Parse()

    // 设置日志级别
    switch *logLevel {
    case "debug":
        logger.SetLevel(logger.DebugLevel)
    case "info":
        logger.SetLevel(logger.InfoLevel)
    case "warn":
        logger.SetLevel(logger.WarnLevel)
    case "error":
        logger.SetLevel(logger.ErrorLevel)
    default:
        logger.SetLevel(logger.InfoLevel)
    }

    // 打印启动信息
    logger.Info("========================================")
    logger.Info("        Go-Redis Server")
    logger.Info("========================================")
    logger.Infof("Port: %d", *port)
    logger.Infof("Log Level: %s", *logLevel)
    logger.Info("========================================")

    // 创建存储
    s := store.NewStore()

    // 创建服务器
    addr := fmt.Sprintf(":%d", *port)
    srv := server.NewServer(addr, s)

    // 优雅关闭处理
    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        sig := <-sigCh

        logger.Infof("Received signal: %v", sig)
        logger.Info("Shutting down server...")

        if err := srv.Stop(); err != nil {
            logger.Errorf("Error stopping server: %v", err)
        }

        os.Exit(0)
    }()

    // 启动服务器
    logger.Infof("Starting Redis server on %s", addr)
    logger.Info("Press Ctrl+C to stop")

    if err := srv.Start(); err != nil {
        logger.Fatalf("Server error: %v", err)
    }
}
```

准备好开始实现了吗？🚀

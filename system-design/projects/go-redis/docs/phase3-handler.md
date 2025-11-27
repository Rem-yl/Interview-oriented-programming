# 第三阶段：命令处理层需求文档

## 1. 需求概述

实现 Redis 命令处理层（Handler），负责接收解析后的 RESP 协议数据，执行具体的命令逻辑，并返回符合 RESP 格式的响应。该层是连接协议层和存储层的桥梁，实现了 Redis 的核心业务逻辑。

### 1.1 业务背景

在完成了存储层（Phase 1）和协议层（Phase 2）后，我们需要一个中间层来：
- 将解析后的命令数组转换为具体的操作
- 调用存储层执行数据操作
- 将操作结果序列化为 RESP 响应
- 处理命令验证和错误情况

### 1.2 核心目标

- 实现基础的 Redis 命令集（PING, SET, GET, DEL, EXISTS, KEYS）
- 建立可扩展的命令注册和路由机制
- 实现命令参数验证
- 提供统一的错误处理
- 与已有的 Store 和 Protocol 层无缝集成

---

## 2. 系统架构

### 2.1 整体架构

```
┌─────────────────────────────────────────────┐
│            客户端 (redis-cli)               │
└───────────────────┬─────────────────────────┘
                    │ RESP 协议
┌───────────────────▼─────────────────────────┐
│              协议层 (Protocol)              │
│         Parser ──────────── Serializer      │
└───────────────────┬─────────────────────────┘
                    │ Value 对象
┌───────────────────▼─────────────────────────┐
│            命令处理层 (Handler)   ← 本阶段  │
│  ┌────────────────────────────────────────┐ │
│  │         命令路由器 (Router)            │ │
│  └──────┬──────────┬──────────┬──────────┘ │
│         │          │          │             │
│  ┌──────▼─────┐ ┌──▼─────┐ ┌──▼──────┐    │
│  │ PingHandler│ │SetHandler│ │GetHandler│   │
│  └────────────┘ └─────┬───┘ └─────┬────┘   │
└────────────────────────┼──────────┼─────────┘
                         │          │
┌────────────────────────▼──────────▼─────────┐
│              存储层 (Store)                 │
│         map[string]interface{}              │
└─────────────────────────────────────────────┘
```

### 2.2 核心组件

#### Handler 接口
```go
// Handler 命令处理器接口
type Handler interface {
    // Handle 处理命令并返回响应
    Handle(args []Value) *Value
}
```

#### Router 路由器
```go
// Router 命令路由器
type Router struct {
    handlers map[string]Handler
    store    *store.Store
}

// Register 注册命令处理器
func (r *Router) Register(command string, handler Handler)

// Route 路由命令到对应的处理器
func (r *Router) Route(cmd *Value) *Value
```

---

## 3. 功能需求

### 3.1 核心命令清单

| 命令 | 格式 | 功能 | 返回值 | 优先级 |
|------|------|------|--------|--------|
| PING | `PING [message]` | 测试连接 | `+PONG\r\n` 或回显消息 | P0 |
| SET | `SET key value` | 设置键值 | `+OK\r\n` | P0 |
| GET | `GET key` | 获取值 | Bulk String 或 NULL | P0 |
| DEL | `DEL key [key ...]` | 删除键 | Integer（删除数量） | P0 |
| EXISTS | `EXISTS key` | 检查键是否存在 | `:0\r\n` 或 `:1\r\n` | P1 |
| KEYS | `KEYS pattern` | 查找匹配的键 | Array of Bulk Strings | P1 |

### 3.2 详细命令规格

#### PING 命令

**语法**：
```
PING [message]
```

**描述**：
- 无参数：返回 `+PONG\r\n`
- 有参数：返回参数内容（Bulk String）

**示例**：
```bash
# 客户端
*1\r\n$4\r\nPING\r\n

# 服务器
+PONG\r\n

# 客户端
*2\r\n$4\r\nPING\r\n$5\r\nhello\r\n

# 服务器
$5\r\nhello\r\n
```

**测试用例**：
- TC1: `PING` → `+PONG\r\n`
- TC2: `PING hello` → `$5\r\nhello\r\n`
- TC3: `PING "hello world"` → `$11\r\nhello world\r\n`

---

#### SET 命令

**语法**：
```
SET key value
```

**描述**：
- 设置键的值
- 如果键已存在，覆盖旧值
- 总是返回 `+OK\r\n`

**示例**：
```bash
# 客户端
*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAlice\r\n

# 服务器
+OK\r\n
```

**测试用例**：
- TC1: `SET key value` → `+OK\r\n`
- TC2: `SET key newvalue` → `+OK\r\n`（覆盖）
- TC3: `SET "hello world" value` → `+OK\r\n`（键包含空格）
- TC4: `SET key` → Error（参数不足）

---

#### GET 命令

**语法**：
```
GET key
```

**描述**：
- 获取键的值
- 键不存在时返回 NULL Bulk String

**示例**：
```bash
# 键存在
# 客户端
*2\r\n$3\r\nGET\r\n$4\r\nname\r\n

# 服务器
$5\r\nAlice\r\n

# 键不存在
# 客户端
*2\r\n$3\r\nGET\r\n$10\r\nnonexistent\r\n

# 服务器
$-1\r\n
```

**测试用例**：
- TC1: `GET existingkey` → Bulk String
- TC2: `GET nonexistent` → `$-1\r\n`
- TC3: `GET` → Error（参数不足）

---

#### DEL 命令

**语法**：
```
DEL key [key ...]
```

**描述**：
- 删除一个或多个键
- 返回实际删除的键数量
- 不存在的键被忽略

**示例**：
```bash
# 客户端
*4\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n$4\r\nkey3\r\n

# 服务器（假设删除了 2 个）
:2\r\n
```

**测试用例**：
- TC1: `DEL key1` → `:1\r\n`
- TC2: `DEL key1 key2 key3` → `:3\r\n`
- TC3: `DEL nonexistent` → `:0\r\n`
- TC4: `DEL` → Error（参数不足）

---

#### EXISTS 命令

**语法**：
```
EXISTS key
```

**描述**：
- 检查键是否存在
- 存在返回 `:1\r\n`
- 不存在返回 `:0\r\n`

**示例**：
```bash
# 客户端
*2\r\n$6\r\nEXISTS\r\n$4\r\nname\r\n

# 服务器（键存在）
:1\r\n

# 服务器（键不存在）
:0\r\n
```

**测试用例**：
- TC1: `EXISTS existingkey` → `:1\r\n`
- TC2: `EXISTS nonexistent` → `:0\r\n`

---

#### KEYS 命令

**语法**：
```
KEYS pattern
```

**描述**：
- 查找所有匹配模式的键
- 支持简单的 `*` 通配符
- 返回键的数组

**模式匹配规则**：
- `*` 匹配任意字符（包括空）
- 其他字符精确匹配

**示例**：
```bash
# 客户端
*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n

# 服务器（假设有 3 个键）
*3\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n$4\r\nkey3\r\n

# 没有匹配
*0\r\n
```

**测试用例**：
- TC1: `KEYS *` → 所有键
- TC2: `KEYS user:*` → 匹配 `user:` 前缀的键
- TC3: `KEYS *name` → 匹配以 `name` 结尾的键
- TC4: `KEYS nonexistent*` → `*0\r\n`（空数组）

---

## 4. 架构设计

### 4.1 目录结构

```
handler/
├── handler.go          # Handler 接口定义
├── router.go           # Router 路由器实现
├── ping.go             # PING 命令处理器
├── set.go              # SET 命令处理器
├── get.go              # GET 命令处理器
├── del.go              # DEL 命令处理器
├── exists.go           # EXISTS 命令处理器
├── keys.go             # KEYS 命令处理器
├── handler_test.go     # 接口测试
├── router_test.go      # 路由器测试
├── ping_test.go        # PING 测试
├── set_test.go         # SET 测试
├── get_test.go         # GET 测试
├── del_test.go         # DEL 测试
├── exists_test.go      # EXISTS 测试
└── keys_test.go        # KEYS 测试
```

### 4.2 核心接口设计

#### Handler 接口

```go
package handler

import (
    "go-redis/protocol"
)

// Handler 命令处理器接口
type Handler interface {
    // Handle 处理命令并返回 RESP Value
    // args: 命令参数（不包括命令名本身）
    // 返回: RESP 响应 Value
    Handle(args []protocol.Value) *protocol.Value
}
```

#### Router 结构

```go
package handler

import (
    "go-redis/protocol"
    "go-redis/store"
    "strings"
)

// Router 命令路由器
type Router struct {
    handlers map[string]Handler
    store    *store.Store
}

// NewRouter 创建新的路由器
func NewRouter(s *store.Store) *Router {
    r := &Router{
        handlers: make(map[string]Handler),
        store:    s,
    }
    r.registerDefaultHandlers()
    return r
}

// Register 注册命令处理器
func (r *Router) Register(command string, handler Handler) {
    r.handlers[strings.ToUpper(command)] = handler
}

// Route 路由命令到对应的处理器
// cmd: 解析后的命令（数组类型）
// 返回: RESP 响应
func (r *Router) Route(cmd *protocol.Value) *protocol.Value {
    // 1. 验证命令格式（必须是数组）
    if cmd.Type != protocol.ArrayType {
        return protocol.Error("ERR expected array")
    }

    // 2. 验证数组不为空
    if len(cmd.Array) == 0 {
        return protocol.Error("ERR empty command")
    }

    // 3. 提取命令名（第一个元素）
    commandName := strings.ToUpper(cmd.Array[0].Str)

    // 4. 查找处理器
    handler, exists := r.handlers[commandName]
    if !exists {
        return protocol.Error("ERR unknown command '" + commandName + "'")
    }

    // 5. 提取参数（剩余元素）
    args := cmd.Array[1:]

    // 6. 调用处理器
    return handler.Handle(args)
}

// registerDefaultHandlers 注册默认命令处理器
func (r *Router) registerDefaultHandlers() {
    r.Register("PING", NewPingHandler())
    r.Register("SET", NewSetHandler(r.store))
    r.Register("GET", NewGetHandler(r.store))
    r.Register("DEL", NewDelHandler(r.store))
    r.Register("EXISTS", NewExistsHandler(r.store))
    r.Register("KEYS", NewKeysHandler(r.store))
}
```

### 4.3 命令处理器实现示例

#### PING Handler

```go
package handler

import "go-redis/protocol"

// PingHandler PING 命令处理器
type PingHandler struct{}

func NewPingHandler() *PingHandler {
    return &PingHandler{}
}

func (h *PingHandler) Handle(args []protocol.Value) *protocol.Value {
    // PING 支持 0 或 1 个参数
    if len(args) == 0 {
        // 无参数：返回 PONG
        return &protocol.Value{
            Type: protocol.StringType,
            Str:  "PONG",
        }
    }

    if len(args) == 1 {
        // 有参数：回显参数
        return &protocol.Value{
            Type: protocol.BulkStringType,
            Str:  args[0].Str,
        }
    }

    // 参数过多
    return &protocol.Value{
        Type: protocol.ErrorType,
        Str:  "ERR wrong number of arguments for 'ping' command",
    }
}
```

#### SET Handler

```go
package handler

import (
    "go-redis/protocol"
    "go-redis/store"
)

// SetHandler SET 命令处理器
type SetHandler struct {
    store *store.Store
}

func NewSetHandler(s *store.Store) *SetHandler {
    return &SetHandler{store: s}
}

func (h *SetHandler) Handle(args []protocol.Value) *protocol.Value {
    // SET 需要恰好 2 个参数：key value
    if len(args) != 2 {
        return &protocol.Value{
            Type: protocol.ErrorType,
            Str:  "ERR wrong number of arguments for 'set' command",
        }
    }

    key := args[0].Str
    value := args[1].Str

    // 调用 Store 设置值
    h.store.Set(key, value)

    // 返回 OK
    return &protocol.Value{
        Type: protocol.StringType,
        Str:  "OK",
    }
}
```

#### GET Handler

```go
package handler

import (
    "go-redis/protocol"
    "go-redis/store"
)

// GetHandler GET 命令处理器
type GetHandler struct {
    store *store.Store
}

func NewGetHandler(s *store.Store) *GetHandler {
    return &GetHandler{store: s}
}

func (h *GetHandler) Handle(args []protocol.Value) *protocol.Value {
    // GET 需要恰好 1 个参数：key
    if len(args) != 1 {
        return &protocol.Value{
            Type: protocol.ErrorType,
            Str:  "ERR wrong number of arguments for 'get' command",
        }
    }

    key := args[0].Str

    // 调用 Store 获取值
    value, exists := h.store.Get(key)

    if !exists {
        // 键不存在：返回 NULL
        return &protocol.Value{
            Type:   protocol.BulkStringType,
            IsNull: true,
        }
    }

    // 返回值（假设存储的是字符串）
    return &protocol.Value{
        Type: protocol.BulkStringType,
        Str:  value.(string),
    }
}
```

---

## 5. 辅助函数

为了方便创建 RESP Value，建议在 `protocol/helpers.go` 中添加以下辅助函数（如果还没有）：

```go
package protocol

// SimpleString 创建简单字符串
func SimpleString(s string) *Value {
    return &Value{
        Type: StringType,
        Str:  s,
    }
}

// Error 创建错误
func Error(msg string) *Value {
    return &Value{
        Type: ErrorType,
        Str:  msg,
    }
}

// Integer 创建整数
func Integer(n int64) *Value {
    return &Value{
        Type: IntType,
        Int:  n,
    }
}

// BulkString 创建批量字符串
func BulkString(s string) *Value {
    return &Value{
        Type: BulkStringType,
        Str:  s,
    }
}

// NullBulkString 创建 NULL 批量字符串
func NullBulkString() *Value {
    return &Value{
        Type:   BulkStringType,
        IsNull: true,
    }
}

// Array 创建数组
func Array(values []Value) *Value {
    return &Value{
        Type:  ArrayType,
        Array: values,
    }
}

// EmptyArray 创建空数组
func EmptyArray() *Value {
    return &Value{
        Type:  ArrayType,
        Array: []Value{},
    }
}
```

---

## 6. 测试计划

### 6.1 单元测试策略

每个命令处理器都需要独立的测试文件，采用**表驱动测试**模式。

#### 测试结构示例

```go
func TestPingHandler(t *testing.T) {
    tests := []struct {
        name     string
        args     []protocol.Value
        expected *protocol.Value
    }{
        {
            name:     "no arguments",
            args:     []protocol.Value{},
            expected: &protocol.Value{Type: protocol.StringType, Str: "PONG"},
        },
        {
            name: "with message",
            args: []protocol.Value{
                {Type: protocol.BulkStringType, Str: "hello"},
            },
            expected: &protocol.Value{Type: protocol.BulkStringType, Str: "hello"},
        },
        {
            name: "too many arguments",
            args: []protocol.Value{
                {Type: protocol.BulkStringType, Str: "hello"},
                {Type: protocol.BulkStringType, Str: "world"},
            },
            expected: &protocol.Value{
                Type: protocol.ErrorType,
                Str:  "ERR wrong number of arguments for 'ping' command",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            handler := NewPingHandler()
            result := handler.Handle(tt.args)

            if !compareValues(result, tt.expected) {
                t.Errorf("expected %+v, got %+v", tt.expected, result)
            }
        })
    }
}
```

### 6.2 集成测试

测试完整的请求-响应流程：

```go
func TestRouterIntegration(t *testing.T) {
    // 1. 创建 Store
    s := store.NewStore()

    // 2. 创建 Router
    router := handler.NewRouter(s)

    // 3. 模拟客户端命令
    setCmd := &protocol.Value{
        Type: protocol.ArrayType,
        Array: []protocol.Value{
            {Type: protocol.BulkStringType, Str: "SET"},
            {Type: protocol.BulkStringType, Str: "name"},
            {Type: protocol.BulkStringType, Str: "Alice"},
        },
    }

    // 4. 执行命令
    response := router.Route(setCmd)

    // 5. 验证响应
    if response.Type != protocol.StringType || response.Str != "OK" {
        t.Errorf("expected +OK, got %+v", response)
    }

    // 6. 验证数据已存储
    getCmd := &protocol.Value{
        Type: protocol.ArrayType,
        Array: []protocol.Value{
            {Type: protocol.BulkStringType, Str: "GET"},
            {Type: protocol.BulkStringType, Str: "name"},
        },
    }

    response = router.Route(getCmd)

    if response.Str != "Alice" {
        t.Errorf("expected 'Alice', got %s", response.Str)
    }
}
```

### 6.3 测试用例清单

| 测试文件 | 测试场景 | 测试数量 |
|---------|---------|---------|
| `ping_test.go` | 无参、有参、多参 | 3+ |
| `set_test.go` | 正常、覆盖、参数错误 | 4+ |
| `get_test.go` | 存在、不存在、参数错误 | 3+ |
| `del_test.go` | 单键、多键、不存在、参数错误 | 4+ |
| `exists_test.go` | 存在、不存在 | 2+ |
| `keys_test.go` | 全匹配、模式匹配、无匹配 | 3+ |
| `router_test.go` | 路由正确性、错误命令、空命令 | 5+ |
| **总计** | | **24+** |

---

## 7. 验收标准

### 7.1 功能验收

- [ ] 所有 6 个基础命令正确实现
- [ ] 命令路由器正确分发命令
- [ ] 参数验证正确处理
- [ ] 错误情况返回正确的错误信息
- [ ] 所有单元测试通过
- [ ] 集成测试通过
- [ ] 测试覆盖率 ≥ 90%

### 7.2 质量验收

- [ ] 代码通过 `go fmt` 格式化
- [ ] 代码通过 `go vet` 静态检查
- [ ] 无明显性能问题
- [ ] 完整的文档注释
- [ ] 良好的错误处理

### 7.3 兼容性验收

- [ ] 与 Store 层正确集成
- [ ] 与 Protocol 层正确集成
- [ ] 响应格式符合 RESP 规范
- [ ] 可以被真实的 redis-cli 调用（下一阶段验证）

---

## 8. 实现提示

### 8.1 开发顺序建议

1. **创建目录和基础文件**
   - `handler/handler.go` - 接口定义
   - `protocol/helpers.go` - 辅助函数

2. **实现 Router**
   - `handler/router.go`
   - `handler/router_test.go`

3. **实现简单命令（从易到难）**
   - PING → SET → GET → EXISTS → DEL → KEYS

4. **集成测试**
   - 完整的请求-响应流程测试

5. **性能测试**
   - Benchmark 测试

### 8.2 关键技术点

#### 命令大小写处理

```go
// 始终转换为大写
commandName := strings.ToUpper(cmd.Array[0].Str)
```

#### 参数数量验证

```go
if len(args) != expectedCount {
    return protocol.Error("ERR wrong number of arguments for '" +
        strings.ToLower(commandName) + "' command")
}
```

#### 类型断言

```go
// Store 中存储的是 interface{}，需要类型断言
value, exists := h.store.Get(key)
if exists {
    strValue := value.(string)  // 假设存储的是字符串
}
```

#### 模式匹配（KEYS 命令）

```go
// 简单的通配符匹配
func matchPattern(pattern, str string) bool {
    if pattern == "*" {
        return true
    }

    // 可以使用 path.Match 或自己实现
    // 这里简化处理
    if strings.HasPrefix(pattern, "*") {
        suffix := pattern[1:]
        return strings.HasSuffix(str, suffix)
    }

    if strings.HasSuffix(pattern, "*") {
        prefix := pattern[:len(pattern)-1]
        return strings.HasPrefix(str, prefix)
    }

    return pattern == str
}
```

### 8.3 常见陷阱

1. **忘记大小写转换**
   - Redis 命令不区分大小写
   - 总是转换为大写后再匹配

2. **参数索引错误**
   - `cmd.Array[0]` 是命令名
   - `cmd.Array[1:]` 才是参数

3. **NULL 值处理**
   - GET 不存在的键要返回 NULL，不是空字符串
   - 用 `IsNull: true`

4. **Store 的类型**
   - Store 存储 `interface{}`
   - 需要类型断言

5. **错误消息格式**
   - 遵循 Redis 的错误格式
   - `ERR <message>`

---

## 9. 调试技巧

### 9.1 单元测试调试

```bash
# 运行单个测试
go test ./handler -v -run TestPingHandler

# 运行所有测试
go test ./handler -v

# 查看覆盖率
go test ./handler -cover

# 生成覆盖率报告
go test ./handler -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 9.2 日志调试

```go
import "go-redis/logger"

func (h *SetHandler) Handle(args []protocol.Value) *protocol.Value {
    logger.Debugf("SET command: key=%s, value=%s", args[0].Str, args[1].Str)
    // ...
}
```

### 9.3 打印 Value 结构

```go
import "encoding/json"

func printValue(v *protocol.Value) {
    data, _ := json.MarshalIndent(v, "", "  ")
    fmt.Println(string(data))
}
```

---

## 10. 扩展思考

完成基础功能后，可以思考：

1. **如何支持更多命令？**
   - INCR, DECR（整数操作）
   - LPUSH, RPUSH（列表操作）
   - HSET, HGET（哈希表操作）

2. **如何优化性能？**
   - 命令处理器池
   - 减少内存分配
   - 并发处理

3. **如何处理命令别名？**
   - 例如：`P` 作为 `PING` 的别名

4. **如何支持事务？**
   - MULTI, EXEC, DISCARD

5. **如何实现中间件？**
   - 日志中间件
   - 权限验证中间件
   - 性能监控中间件

---

## 11. 参考资料

- [Redis 命令参考](https://redis.io/commands/)
- [RESP 协议规范](https://redis.io/docs/reference/protocol-spec/)
- [Go 接口设计](https://go.dev/doc/effective_go#interfaces)
- [表驱动测试](https://go.dev/wiki/TableDrivenTests)

---

## 12. 交付物

完成本阶段后，应该交付：

1. [ ] `handler/handler.go` - Handler 接口定义
2. [ ] `handler/router.go` - Router 实现
3. [ ] `handler/ping.go` - PING 命令实现
4. [ ] `handler/set.go` - SET 命令实现
5. [ ] `handler/get.go` - GET 命令实现
6. [ ] `handler/del.go` - DEL 命令实现
7. [ ] `handler/exists.go` - EXISTS 命令实现
8. [ ] `handler/keys.go` - KEYS 命令实现
9. [ ] `protocol/helpers.go` - 辅助函数
10. [ ] 所有对应的测试文件
11. [ ] 所有测试通过的截图或日志
12. [ ] 覆盖率报告（≥ 90%）

完成后，即可进入**第四阶段：服务器层（Server）**，实现 TCP 服务器和客户端连接处理。

---

## 附录：完整示例

### A.1 完整的 DEL Handler 实现

```go
package handler

import (
    "go-redis/protocol"
    "go-redis/store"
)

// DelHandler DEL 命令处理器
type DelHandler struct {
    store *store.Store
}

func NewDelHandler(s *store.Store) *DelHandler {
    return &DelHandler{store: s}
}

func (h *DelHandler) Handle(args []protocol.Value) *protocol.Value {
    // DEL 至少需要 1 个参数
    if len(args) < 1 {
        return protocol.Error("ERR wrong number of arguments for 'del' command")
    }

    // 删除所有指定的键
    deletedCount := int64(0)
    for _, arg := range args {
        key := arg.Str
        if h.store.Delete(key) {
            deletedCount++
        }
    }

    // 返回删除的数量
    return protocol.Integer(deletedCount)
}
```

### A.2 完整的测试示例

```go
package handler

import (
    "go-redis/protocol"
    "go-redis/store"
    "testing"
)

func TestDelHandler(t *testing.T) {
    tests := []struct {
        name        string
        setupKeys   map[string]string  // 预设的键值
        args        []protocol.Value
        expectedInt int64
        expectedErr bool
    }{
        {
            name: "delete single existing key",
            setupKeys: map[string]string{
                "key1": "value1",
            },
            args: []protocol.Value{
                {Type: protocol.BulkStringType, Str: "key1"},
            },
            expectedInt: 1,
            expectedErr: false,
        },
        {
            name: "delete multiple keys",
            setupKeys: map[string]string{
                "key1": "value1",
                "key2": "value2",
                "key3": "value3",
            },
            args: []protocol.Value{
                {Type: protocol.BulkStringType, Str: "key1"},
                {Type: protocol.BulkStringType, Str: "key2"},
                {Type: protocol.BulkStringType, Str: "key3"},
            },
            expectedInt: 3,
            expectedErr: false,
        },
        {
            name:      "delete non-existent key",
            setupKeys: map[string]string{},
            args: []protocol.Value{
                {Type: protocol.BulkStringType, Str: "nonexistent"},
            },
            expectedInt: 0,
            expectedErr: false,
        },
        {
            name: "delete mix of existing and non-existing",
            setupKeys: map[string]string{
                "key1": "value1",
            },
            args: []protocol.Value{
                {Type: protocol.BulkStringType, Str: "key1"},
                {Type: protocol.BulkStringType, Str: "key2"},
            },
            expectedInt: 1,
            expectedErr: false,
        },
        {
            name:        "no arguments",
            setupKeys:   map[string]string{},
            args:        []protocol.Value{},
            expectedInt: 0,
            expectedErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 创建新的 Store
            s := store.NewStore()

            // 预设键值
            for key, value := range tt.setupKeys {
                s.Set(key, value)
            }

            // 创建处理器
            handler := NewDelHandler(s)

            // 执行命令
            result := handler.Handle(tt.args)

            // 验证结果
            if tt.expectedErr {
                if result.Type != protocol.ErrorType {
                    t.Errorf("expected error, got %+v", result)
                }
            } else {
                if result.Type != protocol.IntType {
                    t.Errorf("expected integer type, got %v", result.Type)
                }
                if result.Int != tt.expectedInt {
                    t.Errorf("expected %d deleted, got %d", tt.expectedInt, result.Int)
                }
            }
        })
    }
}
```

准备好开始实现了吗？🚀

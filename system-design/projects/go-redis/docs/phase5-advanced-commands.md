# 第五阶段：扩展命令实现

## 1. 需求概述

在完成了基础的 6 个命令（PING, SET, GET, DEL, EXISTS, KEYS）后，本阶段将实现更多实用的 Redis 命令，包括字符串操作、数值操作和批量操作等，让你的 Redis 服务器更加完善。

### 1.1 业务背景

实际应用中，Redis 不仅仅是简单的键值存储，还需要支持：
- 原子操作（计数器、统计）
- 字符串操作（追加、截取）
- 批量操作（提高效率）
- 多类型值处理（字符串、整数）

### 1.2 核心目标

- 实现 10+ 个常用 Redis 命令
- 掌握原子操作的实现
- 学习类型检查和转换
- 提升命令处理的健壮性

---

## 2. 命令清单

### 2.1 字符串操作命令（优先级 P0）

| 命令 | 语法 | 功能 | 返回值 |
|------|------|------|--------|
| **APPEND** | `APPEND key value` | 追加字符串 | 追加后的长度 |
| **STRLEN** | `STRLEN key` | 获取字符串长度 | 字符串长度 |
| **GETRANGE** | `GETRANGE key start end` | 获取子串 | 子串内容 |
| **SETRANGE** | `SETRANGE key offset value` | 设置子串 | 修改后长度 |

### 2.2 数值操作命令（优先级 P0）

| 命令 | 语法 | 功能 | 返回值 |
|------|------|------|--------|
| **INCR** | `INCR key` | 自增 1 | 增加后的值 |
| **DECR** | `DECR key` | 自减 1 | 减少后的值 |
| **INCRBY** | `INCRBY key increment` | 增加指定值 | 增加后的值 |
| **DECRBY** | `DECRBY key decrement` | 减少指定值 | 减少后的值 |

### 2.3 批量操作命令（优先级 P1）

| 命令 | 语法 | 功能 | 返回值 |
|------|------|------|--------|
| **MGET** | `MGET key [key ...]` | 批量获取 | 值数组 |
| **MSET** | `MSET key value [key value ...]` | 批量设置 | OK |
| **MSETNX** | `MSETNX key value [key value ...]` | 批量设置（不存在时） | 0/1 |

---

## 3. 详细命令规格

### 3.1 APPEND 命令

**语法**：
```
APPEND key value
```

**描述**：
- 如果键已存在，追加值到现有值的末尾
- 如果键不存在，相当于 SET
- 返回追加后字符串的长度

**示例**：
```bash
redis> SET mykey "Hello"
OK

redis> APPEND mykey " World"
(integer) 11

redis> GET mykey
"Hello World"

# 键不存在的情况
redis> APPEND newkey "First"
(integer) 5

redis> GET newkey
"First"
```

**实现要点**：
```go
func (h *AppendHandler) Handle(args []protocol.Value) *protocol.Value {
    if len(args) != 2 {
        return protocol.Error("ERR wrong number of arguments for 'append' command")
    }

    key := args[0].Str
    appendValue := args[1].Str

    // 获取现有值
    existingValue, exists := h.db.Get(key)

    var newValue string
    if exists {
        // 类型检查
        strValue, ok := existingValue.(string)
        if !ok {
            return protocol.Error("WRONGTYPE Operation against a key holding the wrong kind of value")
        }
        newValue = strValue + appendValue
    } else {
        newValue = appendValue
    }

    h.db.Set(key, newValue)
    return protocol.Integer(int64(len(newValue)))
}
```

**测试用例**：
- TC1: 键存在，追加成功
- TC2: 键不存在，相当于 SET
- TC3: 键存在但不是字符串类型，返回错误

---

### 3.2 INCR 命令

**语法**：
```
INCR key
```

**描述**：
- 将键的值自增 1
- 如果键不存在，初始化为 0 然后自增
- 值必须是整数字符串
- **原子操作**，线程安全

**示例**：
```bash
redis> SET counter "10"
OK

redis> INCR counter
(integer) 11

redis> INCR counter
(integer) 12

# 键不存在
redis> INCR newcounter
(integer) 1

# 错误情况
redis> SET mykey "hello"
OK

redis> INCR mykey
(error) ERR value is not an integer or out of range
```

**实现要点**：
```go
func (h *IncrHandler) Handle(args []protocol.Value) *protocol.Value {
    if len(args) != 1 {
        return protocol.Error("ERR wrong number of arguments for 'incr' command")
    }

    key := args[0].Str

    // 加锁保证原子性（Store 内部已有锁）
    existingValue, exists := h.db.Get(key)

    var currentNum int64
    if exists {
        // 尝试转换为整数
        switch v := existingValue.(type) {
        case string:
            num, err := strconv.ParseInt(v, 10, 64)
            if err != nil {
                return protocol.Error("ERR value is not an integer or out of range")
            }
            currentNum = num
        case int64:
            currentNum = v
        default:
            return protocol.Error("WRONGTYPE Operation against a key holding the wrong kind of value")
        }
    }

    // 自增
    newNum := currentNum + 1
    h.db.Set(key, newNum)

    return protocol.Integer(newNum)
}
```

**关键点**：
- **原子性**：整个读-改-写操作必须原子执行
- **类型处理**：支持字符串和 int64 两种存储格式
- **边界检查**：检查溢出（可选）

**测试用例**：
- TC1: 键存在且为整数字符串
- TC2: 键存在且为 int64
- TC3: 键不存在，初始化为 0
- TC4: 键存在但不是数字，返回错误
- TC5: 并发自增测试（验证原子性）

---

### 3.3 INCRBY 命令

**语法**：
```
INCRBY key increment
```

**描述**：
- 将键的值增加指定的整数
- increment 可以是负数（相当于减法）
- 其他特性同 INCR

**示例**：
```bash
redis> SET counter "10"
OK

redis> INCRBY counter 5
(integer) 15

redis> INCRBY counter -3
(integer) 12

redis> INCRBY newcounter 100
(integer) 100
```

**实现提示**：
- 可以复用 INCR 的逻辑，只是增量可变

---

### 3.4 STRLEN 命令

**语法**：
```
STRLEN key
```

**描述**：
- 返回键的字符串长度
- 键不存在返回 0
- 非字符串类型返回错误

**示例**：
```bash
redis> SET mykey "Hello World"
OK

redis> STRLEN mykey
(integer) 11

redis> STRLEN nonexistent
(integer) 0
```

**实现**：
```go
func (h *StrlenHandler) Handle(args []protocol.Value) *protocol.Value {
    if len(args) != 1 {
        return protocol.Error("ERR wrong number of arguments for 'strlen' command")
    }

    key := args[0].Str
    value, exists := h.db.Get(key)

    if !exists {
        return protocol.Integer(0)
    }

    strValue, ok := value.(string)
    if !ok {
        return protocol.Error("WRONGTYPE Operation against a key holding the wrong kind of value")
    }

    return protocol.Integer(int64(len(strValue)))
}
```

---

### 3.5 MGET 命令

**语法**：
```
MGET key [key ...]
```

**描述**：
- 批量获取多个键的值
- 不存在的键返回 NULL
- 返回数组，顺序与参数顺序一致

**示例**：
```bash
redis> SET key1 "Hello"
OK

redis> SET key2 "World"
OK

redis> MGET key1 key2 key3
1) "Hello"
2) "World"
3) (nil)
```

**实现**：
```go
func (h *MgetHandler) Handle(args []protocol.Value) *protocol.Value {
    if len(args) < 1 {
        return protocol.Error("ERR wrong number of arguments for 'mget' command")
    }

    results := make([]protocol.Value, len(args))

    for i, keyVal := range args {
        key := keyVal.Str
        value, exists := h.db.Get(key)

        if !exists {
            results[i] = protocol.Value{
                Type:   protocol.BulkStringType,
                IsNull: true,
            }
        } else {
            strValue, ok := value.(string)
            if !ok {
                // 非字符串类型也返回 NULL（Redis 行为）
                results[i] = protocol.Value{
                    Type:   protocol.BulkStringType,
                    IsNull: true,
                }
            } else {
                results[i] = protocol.Value{
                    Type: protocol.BulkStringType,
                    Str:  strValue,
                }
            }
        }
    }

    return &protocol.Value{
        Type:  protocol.ArrayType,
        Array: results,
    }
}
```

---

### 3.6 MSET 命令

**语法**：
```
MSET key value [key value ...]
```

**描述**：
- 批量设置多个键值对
- 总是成功，返回 OK
- 是原子操作

**示例**：
```bash
redis> MSET key1 "Hello" key2 "World" key3 "!"
OK

redis> GET key1
"Hello"

redis> GET key2
"World"
```

**实现**：
```go
func (h *MsetHandler) Handle(args []protocol.Value) *protocol.Value {
    if len(args) < 2 || len(args)%2 != 0 {
        return protocol.Error("ERR wrong number of arguments for 'mset' command")
    }

    // 批量设置
    for i := 0; i < len(args); i += 2 {
        key := args[i].Str
        value := args[i+1].Str
        h.db.Set(key, value)
    }

    return protocol.SimpleString("OK")
}
```

**注意**：
- 参数必须是偶数个（key-value 对）
- 原子性：所有设置操作要么全部成功，要么全部失败（当前简化实现已满足）

---

## 4. 实现建议

### 4.1 开发顺序

推荐按以下顺序实现：

1. **STRLEN**（最简单）
   - 只需获取字符串长度
   - 适合热身

2. **APPEND**（中等）
   - 字符串拼接
   - 需要类型检查

3. **INCR / DECR**（重要）
   - 原子操作
   - 类型转换
   - 并发测试

4. **INCRBY / DECRBY**（扩展）
   - 基于 INCR/DECR

5. **MGET**（批量读）
   - 处理数组返回
   - NULL 值处理

6. **MSET**（批量写）
   - 参数解析
   - 批量操作

### 4.2 目录结构

```
handler/
├── router.go
├── ping.go
├── set.go
├── get.go
├── del.go
├── exists.go
├── keys.go
├── append.go         # ← 新增
├── strlen.go         # ← 新增
├── incr.go           # ← 新增
├── decr.go           # ← 新增
├── incrby.go         # ← 新增
├── decrby.go         # ← 新增
├── mget.go           # ← 新增
├── mset.go           # ← 新增
└── router_test.go    # 更新测试
```

### 4.3 通用工具函数

建议在 `handler/utils.go` 中添加通用函数：

```go
package handler

import (
    "go-redis/protocol"
    "strconv"
)

// ParseInt 从 Value 中解析整数
func ParseInt(value interface{}) (int64, error) {
    switch v := value.(type) {
    case string:
        return strconv.ParseInt(v, 10, 64)
    case int64:
        return v, nil
    case int:
        return int64(v), nil
    default:
        return 0, errors.New("value is not an integer")
    }
}

// ToString 将值转换为字符串
func ToString(value interface{}) (string, bool) {
    switch v := value.(type) {
    case string:
        return v, true
    case int64:
        return strconv.FormatInt(v, 10), true
    default:
        return "", false
    }
}

// WrongTypeError 返回类型错误
func WrongTypeError() *protocol.Value {
    return protocol.Error("WRONGTYPE Operation against a key holding the wrong kind of value")
}

// NotIntegerError 返回非整数错误
func NotIntegerError() *protocol.Value {
    return protocol.Error("ERR value is not an integer or out of range")
}
```

---

## 5. 测试策略

### 5.1 单元测试示例

```go
func TestIncrHandler(t *testing.T) {
    tests := []struct {
        name        string
        setupKey    string
        setupValue  interface{}
        args        []protocol.Value
        expectedInt int64
        expectError bool
    }{
        {
            name:        "incr existing integer string",
            setupKey:    "counter",
            setupValue:  "10",
            args:        []protocol.Value{{Type: protocol.BulkStringType, Str: "counter"}},
            expectedInt: 11,
            expectError: false,
        },
        {
            name:        "incr non-existent key",
            setupKey:    "",
            setupValue:  nil,
            args:        []protocol.Value{{Type: protocol.BulkStringType, Str: "newcounter"}},
            expectedInt: 1,
            expectError: false,
        },
        {
            name:        "incr non-integer value",
            setupKey:    "mykey",
            setupValue:  "hello",
            args:        []protocol.Value{{Type: protocol.BulkStringType, Str: "mykey"}},
            expectedInt: 0,
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            s := store.NewStore()
            if tt.setupKey != "" {
                s.Set(tt.setupKey, tt.setupValue)
            }

            handler := NewIncrHandler(s)
            result := handler.Handle(tt.args)

            if tt.expectError {
                if result.Type != protocol.ErrorType {
                    t.Errorf("expected error, got %v", result)
                }
            } else {
                if result.Type != protocol.IntType {
                    t.Errorf("expected IntType, got %v", result.Type)
                }
                if result.Int != tt.expectedInt {
                    t.Errorf("expected %d, got %d", tt.expectedInt, result.Int)
                }
            }
        })
    }
}
```

### 5.2 并发测试（INCR）

```go
func TestIncrConcurrent(t *testing.T) {
    s := store.NewStore()
    s.Set("counter", int64(0))

    handler := NewIncrHandler(s)

    // 启动 100 个 goroutine 并发自增
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            args := []protocol.Value{{Type: protocol.BulkStringType, Str: "counter"}}
            handler.Handle(args)
        }()
    }

    wg.Wait()

    // 验证最终结果
    value, _ := s.Get("counter")
    finalCount := value.(int64)

    if finalCount != 100 {
        t.Errorf("expected 100, got %d", finalCount)
    }
}
```

### 5.3 集成测试（redis-cli）

```bash
# 启动服务器
go run main.go

# 在另一个终端测试
redis-cli -p 16379

# INCR 测试
127.0.0.1:16379> SET counter 0
OK
127.0.0.1:16379> INCR counter
(integer) 1
127.0.0.1:16379> INCR counter
(integer) 2
127.0.0.1:16379> INCRBY counter 10
(integer) 12

# APPEND 测试
127.0.0.1:16379> SET msg "Hello"
OK
127.0.0.1:16379> APPEND msg " World"
(integer) 11
127.0.0.1:16379> GET msg
"Hello World"

# MGET/MSET 测试
127.0.0.1:16379> MSET k1 v1 k2 v2 k3 v3
OK
127.0.0.1:16379> MGET k1 k2 k3 k4
1) "v1"
2) "v2"
3) "v3"
4) (nil)
```

---

## 6. 性能优化

### 6.1 INCR 原子性优化

如果 Store 的锁粒度太粗，可以考虑：

```go
// 选项 1：Store 提供原子操作方法
func (s *Store) AtomicIncr(key string, delta int64) (int64, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    value, exists := s.data[key]
    var current int64

    if exists {
        // 类型转换...
    }

    newValue := current + delta
    s.data[key] = newValue
    return newValue, nil
}

// 选项 2：使用 sync.Map（Go 1.9+）
// 天然支持原子操作
```

### 6.2 批量操作优化

MGET/MSET 可以减少网络往返：

```go
// 性能对比
// 方式 1：逐个 GET（需要 N 次网络往返）
GET key1
GET key2
GET key3

// 方式 2：MGET（只需 1 次网络往返）
MGET key1 key2 key3
```

---

## 7. 验收标准

### 7.1 功能验收

- [ ] 所有新命令正确实现
- [ ] 参数验证完整
- [ ] 错误处理正确
- [ ] 类型检查严格
- [ ] 原子操作保证（INCR/DECR）
- [ ] 所有单元测试通过
- [ ] 并发测试通过（INCR）
- [ ] redis-cli 测试通过

### 7.2 性能验收

- [ ] INCR 并发性能 > 50,000 ops/sec
- [ ] MGET 性能优于单独 GET
- [ ] 无明显内存泄漏

### 7.3 代码质量

- [ ] 代码格式化（go fmt）
- [ ] 静态检查通过（go vet）
- [ ] 测试覆盖率 > 85%
- [ ] 完整的文档注释

---

## 8. 扩展思考

完成这些命令后，思考：

1. **如何实现 GETSET**？
   - 原子地设置新值并返回旧值

2. **如何优化字符串存储**？
   - 使用 []byte 而不是 string 减少拷贝

3. **如何实现 SETEX**（带过期时间的 SET）？
   - 需要先实现过期机制（Phase 6）

4. **如何实现 SETNX**（仅在不存在时设置）？
   - 可用于分布式锁

---

## 9. 下一步

完成本阶段后，你可以：

1. **实现过期时间支持**（Phase 6）
   - EXPIRE, TTL, PERSIST
   - 后台清理机制

2. **实现持久化**（Phase 7）
   - RDB 快照
   - AOF 日志

3. **实现列表、哈希、集合**（Phase 8）
   - 更复杂的数据结构

---

## 10. 参考资料

- [Redis 字符串命令](https://redis.io/commands/?group=string)
- [Redis INCR 原子性保证](https://redis.io/commands/incr/)
- [Go 原子操作](https://pkg.go.dev/sync/atomic)

---

**准备好开始实现了吗？建议从 STRLEN 开始，逐步实现所有命令！** 🚀

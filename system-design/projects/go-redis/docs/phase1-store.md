# 第一阶段：数据存储层需求文档

## 1. 需求概述

实现一个线程安全的内存键值存储引擎，作为 Redis 的核心数据层。该模块需要支持基本的 CRUD 操作，并保证在高并发场景下的数据一致性和访问安全性。

### 1.1 业务背景

Redis 的核心是一个高性能的内存数据库，所有数据操作都基于内存存储。本阶段需要实现这个基础存储引擎，为后续的协议解析、命令处理等模块提供可靠的数据访问接口。

### 1.2 核心目标

- 提供高性能的内存键值存储
- 保证多线程并发访问的安全性
- 设计简洁易用的 API 接口
- 支持未来的功能扩展（过期、持久化等）

---

## 2. 功能需求

### 2.1 核心功能清单

| 功能 | 优先级 | 说明 |
|------|--------|------|
| Set 操作 | P0 | 设置键值对，键为 string，值为任意类型 |
| Get 操作 | P0 | 根据键获取值，返回值和存在标识 |
| Delete 操作 | P0 | 删除指定的键 |
| Exists 操作 | P0 | 检查键是否存在 |
| Keys 操作 | P1 | 获取所有键的列表 |
| Clear 操作 | P1 | 清空所有数据 |

### 2.2 功能详细说明

#### F1: Set 操作
**功能描述**：设置键值对到存储中

**输入**：
- `key`: string - 键名
- `value`: interface{} - 任意类型的值

**输出**：无

**行为**：
- 如果键不存在，创建新的键值对
- 如果键已存在，覆盖原有值
- 操作必须是原子性的

**示例**：
```go
store.Set("name", "Alice")
store.Set("age", 25)
store.Set("scores", []int{90, 85, 88})
```

---

#### F2: Get 操作
**功能描述**：根据键获取对应的值

**输入**：
- `key`: string - 键名

**输出**：
- `value`: interface{} - 对应的值
- `exists`: bool - 键是否存在的标识

**行为**：
- 如果键存在，返回值和 true
- 如果键不存在，返回 nil 和 false
- 必须能区分"值为 nil"和"键不存在"两种情况

**示例**：
```go
value, exists := store.Get("name")  // "Alice", true
value, exists := store.Get("missing")  // nil, false
```

---

#### F3: Delete 操作
**功能描述**：删除指定的键

**输入**：
- `key`: string - 要删除的键名

**输出**：无

**行为**：
- 如果键存在，删除该键值对
- 如果键不存在，操作无效果（不报错）
- 删除后，该键的 Get 操作应返回不存在

**示例**：
```go
store.Delete("name")
```

---

#### F4: Exists 操作
**功能描述**：检查键是否存在

**输入**：
- `key`: string - 键名

**输出**：
- `exists`: bool - 是否存在

**行为**：
- 键存在返回 true
- 键不存在返回 false
- 不修改存储状态

**示例**：
```go
exists := store.Exists("name")  // true 或 false
```

---

#### F5: Keys 操作
**功能描述**：获取所有键的列表

**输入**：无

**输出**：
- `keys`: []string - 所有键的切片

**行为**：
- 返回当前存储中所有的键
- 空存储返回空切片（不是 nil）
- 返回的顺序不保证（map 的迭代顺序）

**示例**：
```go
keys := store.Keys()  // ["name", "age", "scores"]
```

---

#### F6: Clear 操作
**功能描述**：清空所有数据

**输入**：无

**输出**：无

**行为**：
- 删除所有键值对
- 清空后，Keys() 应返回空切片
- 操作必须是原子性的

**示例**：
```go
store.Clear()
```

---

## 3. 架构设计

### 3.1 整体架构

```
┌─────────────────────────────────────┐
│         Store 接口层                │
│  (Public API: Set/Get/Delete...)    │
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│       并发控制层                    │
│    (sync.RWMutex 读写锁)            │
└─────────────┬───────────────────────┘
              │
┌─────────────▼───────────────────────┐
│        数据存储层                   │
│   (map[string]interface{})          │
└─────────────────────────────────────┘
```

### 3.2 数据结构设计

#### Store 结构体

```go
type Store struct {
    mu   sync.RWMutex          // 读写锁
    data map[string]interface{} // 数据字典
}
```

**设计要点**：

1. **为什么使用 `sync.RWMutex`？**
   - 读多写少的场景：Redis 的读操作通常远多于写操作
   - 允许多个读操作并发执行
   - 写操作时独占访问
   - 性能优于普通的 `sync.Mutex`

2. **为什么值类型是 `interface{}`？**
   - 需要支持多种数据类型（String、List、Hash、Set 等）
   - 提供灵活性，为后续扩展预留空间
   - Go 的类型系统允许安全的类型断言

3. **为什么不使用 `sync.Map`？**
   - `sync.Map` 适合"键集合稳定"的场景
   - Redis 的键频繁增删，不适合 `sync.Map`
   - `RWMutex + map` 在这个场景下性能更好

### 3.3 并发控制策略

#### 读写锁使用规则

| 操作类型 | 锁类型 | 说明 |
|---------|--------|------|
| Get | RLock | 只读操作，允许并发 |
| Exists | RLock | 只读操作，允许并发 |
| Keys | RLock | 只读操作，允许并发 |
| Set | Lock | 写操作，独占访问 |
| Delete | Lock | 写操作，独占访问 |
| Clear | Lock | 写操作，独占访问 |

#### 锁的使用模式

```go
func (s *Store) Get(key string) (interface{}, bool) {
    s.mu.RLock()         // 获取读锁
    defer s.mu.RUnlock() // 确保释放

    value, exists := s.data[key]
    return value, exists
}
```

**注意事项**：
- 必须使用 `defer` 确保锁被释放
- 避免在持有锁时进行耗时操作
- 防止死锁：不要在持有锁时调用其他需要锁的方法

### 3.4 接口设计

```go
// Store 接口（可选，便于后续扩展和测试）
type Storer interface {
    Set(key string, value interface{})
    Get(key string) (interface{}, bool)
    Delete(key string)
    Exists(key string) bool
    Keys() []string
    Clear()
}
```

---

## 4. 测试计划

### 4.1 测试策略

采用**测试驱动开发（TDD）**方式：
1. 先编写测试用例
2. 运行测试，看到失败
3. 编写实现代码
4. 运行测试，确保通过
5. 重构代码
6. 重复以上步骤

### 4.2 测试用例清单

#### TC1: 基本读写测试 (TestSetAndGet)

**测试目标**：验证基本的 Set 和 Get 功能

**测试步骤**：
1. 创建 Store 实例
2. 使用 Set 设置键值对
3. 使用 Get 获取值
4. 验证值正确且 exists 为 true

**预期结果**：
- Get 返回的值与 Set 设置的值相同
- exists 标识为 true

**边界条件**：
- 不同类型的值（string, int, []int, map 等）
- 重复设置同一个键（覆盖）

---

#### TC2: 获取不存在的键 (TestGetNonExistent)

**测试目标**：验证获取不存在的键的行为

**测试步骤**：
1. 创建 Store 实例
2. 直接 Get 一个未设置的键
3. 验证返回值

**预期结果**：
- value 为 nil
- exists 为 false

**重要性**：区分"值为 nil"和"键不存在"

---

#### TC3: 删除操作测试 (TestDelete)

**测试目标**：验证删除功能

**测试步骤**：
1. 创建 Store 并设置键
2. 验证键存在
3. 删除键
4. 再次 Get，验证键不存在

**预期结果**：
- 删除前 exists 为 true
- 删除后 exists 为 false

**边界条件**：
- 删除不存在的键（不应报错）

---

#### TC4: 键存在性检查 (TestExists)

**测试目标**：验证 Exists 方法

**测试步骤**：
1. 检查不存在的键
2. 设置键
3. 再次检查

**预期结果**：
- 设置前返回 false
- 设置后返回 true

---

#### TC5: 并发安全测试 (TestConcurrentAccess)

**测试目标**：验证多线程并发访问的安全性

**测试步骤**：
1. 启动 N 个 goroutine 并发写入
2. 启动 M 个 goroutine 并发读取
3. 等待所有 goroutine 完成
4. 验证无 panic 或数据竞争

**预期结果**：
- 没有 panic
- 使用 `go test -race` 无竞态警告
- 数据一致性

**测试参数**：
- 写 goroutine 数：10-100
- 读 goroutine 数：10-100
- 每个 goroutine 操作次数：100-1000

---

#### TC6: 获取所有键测试 (TestKeys)

**测试目标**：验证 Keys 方法

**测试步骤**：
1. 空存储时调用 Keys
2. 添加多个键
3. 再次调用 Keys
4. 验证所有键都存在

**预期结果**：
- 空存储返回空切片
- 返回的键集合包含所有已设置的键

**注意**：不验证顺序，因为 map 迭代顺序不确定

---

#### TC7: 清空测试 (TestClear)

**测试目标**：验证 Clear 方法

**测试步骤**：
1. 添加多个键
2. 调用 Clear
3. 验证所有键都被删除

**预期结果**：
- Clear 后 Keys() 返回空切片
- 所有键的 Exists 都返回 false

---

### 4.3 测试覆盖率要求

- **代码覆盖率**：≥ 95%
- **分支覆盖率**：≥ 90%
- **并发测试**：必须通过 `-race` 检测

### 4.4 性能测试（可选）

使用 Go 的 benchmark 测试性能：

```go
func BenchmarkSet(b *testing.B)
func BenchmarkGet(b *testing.B)
func BenchmarkConcurrentGet(b *testing.B)
```

**性能目标**：
- Set 操作：< 100 ns/op
- Get 操作：< 50 ns/op

---

## 5. 验收标准

### 5.1 功能验收

- [ ] 所有 P0 功能已实现（Set, Get, Delete, Exists）
- [ ] 所有 P1 功能已实现（Keys, Clear）
- [ ] 所有测试用例通过
- [ ] 测试覆盖率 ≥ 95%

### 5.2 质量验收

- [ ] 代码通过 `go fmt` 格式化
- [ ] 代码通过 `go vet` 静态检查
- [ ] 代码通过 `go test -race` 竞态检测
- [ ] 所有公开函数有完整的文档注释
- [ ] 无明显性能问题

### 5.3 代码规范

- [ ] 变量命名清晰有意义
- [ ] 每个函数职责单一
- [ ] 适当的错误处理
- [ ] 合理的代码注释

---

## 6. 技术约束

### 6.1 开发环境

- Go 版本：≥ 1.23.0
- 依赖：仅标准库（`sync` 包）

### 6.2 性能约束

- 内存占用：合理使用，避免内存泄漏
- 并发性能：读操作允许并发
- 响应时间：纳秒级别（内存操作）

### 6.3 限制条件

- 不需要实现持久化
- 不需要实现键过期
- 不需要实现 LRU 等淘汰策略
- 不需要考虑分布式

---

## 7. 实现提示

### 7.1 开发顺序建议

按照以下顺序实现，每个功能完成后确保测试通过：

1. **基础结构** → NewStore, Set, Get
2. **删除功能** → Delete
3. **检查功能** → Exists
4. **列表功能** → Keys
5. **清空功能** → Clear
6. **并发测试** → 验证线程安全

### 7.2 常见陷阱

1. **忘记加锁** → 使用 `-race` 检测
2. **死锁** → 避免嵌套锁调用
3. **锁粒度过大** → 在锁内避免耗时操作
4. **返回值设计** → Get 必须返回两个值

### 7.3 调试技巧

```bash
# 运行测试
go test ./store -v

# 竞态检测
go test ./store -race

# 覆盖率
go test ./store -cover

# 详细覆盖率报告
go test ./store -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## 8. 扩展思考

完成基础功能后，可以思考以下问题：

1. **如何实现键过期？**
   - 为每个键添加过期时间字段
   - 实现惰性删除或定期删除策略

2. **如何优化内存使用？**
   - 考虑使用更紧凑的数据结构
   - 实现 LRU 淘汰策略

3. **如何支持更多数据类型？**
   - 设计类型系统
   - 为不同类型设计不同的存储结构

4. **如何提升性能？**
   - 考虑分片（sharding）减少锁竞争
   - 使用更高效的并发数据结构

---

## 9. 参考资料

- [Go sync 包文档](https://pkg.go.dev/sync)
- [Go map 数据结构](https://go.dev/blog/maps)
- [Effective Go - 并发](https://go.dev/doc/effective_go#concurrency)
- [Redis 内部数据结构](https://redis.io/docs/data-types/)

---

## 10. 交付物

完成本阶段后，应该交付：

1. ✅ `store/store.go` - 实现文件
2. ✅ `store/store_test.go` - 测试文件
3. ✅ 所有测试通过的截图或日志
4. ✅ 覆盖率报告（≥ 95%）

准备好后，进入第二阶段：RESP 协议实现。

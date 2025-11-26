# 第二阶段：RESP 协议层需求文档

## 1. 需求概述

实现 Redis 序列化协议（RESP - REdis Serialization Protocol）的解析器和序列化器，作为客户端与服务器之间的通信桥梁。该模块需要能够正确解析客户端发送的命令，并将服务器响应序列化为符合 RESP 规范的格式。

### 1.1 业务背景

RESP 是 Redis 客户端和服务器之间通信使用的协议。它是一个基于文本的协议，具有以下特点：
- 实现简单，易于调试
- 解析速度快
- 人类可读（便于使用 telnet/netcat 调试）
- 支持二进制安全

### 1.2 核心目标

- 实现完整的 RESP 协议解析器
- 支持所有 RESP 数据类型
- 实现高效的序列化器
- 为后续的命令处理层提供标准化接口
- 保证协议的正确性和健壮性

---

## 2. RESP 协议规范

### 2.1 协议格式

RESP 协议使用不同的前缀字符来标识数据类型：

| 类型         | 前缀 | 格式                         | 示例                           |
| ------------ | ---- | ---------------------------- | ------------------------------ |
| 简单字符串   | `+`  | `+内容\r\n`                  | `+OK\r\n`                      |
| 错误         | `-`  | `-错误信息\r\n`              | `-ERR unknown command\r\n`     |
| 整数         | `:`  | `:数字\r\n`                  | `:1000\r\n`                    |
| 批量字符串   | `$`  | `$长度\r\n内容\r\n`          | `$6\r\nfoobar\r\n`             |
| 数组         | `*`  | `*元素数量\r\n[元素...]\r\n` | `*2\r\n$3\r\nfoo\r\n:123\r\n` |
| 空批量字符串 | `$`  | `$-1\r\n`                    | 表示 NULL                      |
| 空数组       | `*`  | `*-1\r\n`                    | 表示 NULL 数组                 |

### 2.2 数据类型详解

#### 简单字符串 (Simple String)

**格式**：`+内容\r\n`

**特点**：
- 不能包含 `\r` 或 `\n` 字符
- 用于简单的状态回复
- 如：`+OK\r\n`、`+PONG\r\n`

**示例**：
```
+OK\r\n
+PONG\r\n
+QUEUED\r\n
```

---

#### 错误 (Error)

**格式**：`-错误类型 错误信息\r\n`

**特点**：
- 第一个单词是错误类型（惯例大写）
- 后面是错误描述
- 客户端应将其视为异常

**示例**：
```
-ERR unknown command 'foobar'\r\n
-WRONGTYPE Operation against a key holding the wrong kind of value\r\n
```

---

#### 整数 (Integer)

**格式**：`:数字\r\n`

**特点**：
- 用于返回整数结果
- 可以是负数
- 如计数、布尔值（0/1）

**示例**：
```
:0\r\n
:1000\r\n
:-123\r\n
```

---

#### 批量字符串 (Bulk String)

**格式**：`$长度\r\n内容\r\n`

**特点**：
- 二进制安全（可以包含任何字节）
- 长度以字节为单位
- NULL 表示为 `$-1\r\n`

**示例**：
```
$6\r\nfoobar\r\n          # 字符串 "foobar"
$0\r\n\r\n                # 空字符串
$-1\r\n                   # NULL
$11\r\nhello\r\nworld\r\n # 包含 \r\n 的字符串
```

---

#### 数组 (Array)

**格式**：`*元素数量\r\n[元素1][元素2]...\r\n`

**特点**：
- 可以包含任意类型的元素
- 元素可以是不同类型
- 可以嵌套
- NULL 数组表示为 `*-1\r\n`
- 空数组表示为 `*0\r\n`

**示例**：
```
# 数组 ["foo", "bar"]
*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n

# 数组 [1, 2, 3]
*3\r\n:1\r\n:2\r\n:3\r\n

# 混合类型数组 [1, "foo", ["bar", 5]]
*3\r\n:1\r\n$3\r\nfoo\r\n*2\r\n$3\r\nbar\r\n:5\r\n

# 空数组
*0\r\n

# NULL 数组
*-1\r\n
```

### 2.3 客户端命令格式

客户端发送的命令使用**数组格式**：

**示例**：
```bash
# SET name "Alice"
*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAlice\r\n

# GET name
*2\r\n$3\r\nGET\r\n$4\r\nname\r\n

# DEL key1 key2 key3
*4\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n$4\r\nkey3\r\n
```

### 2.4 服务器响应格式

服务器根据命令结果返回不同类型：

| 场景           | 返回类型   | 示例                  |
| -------------- | ---------- | --------------------- |
| 成功状态       | 简单字符串 | `+OK\r\n`             |
| 错误           | 错误       | `-ERR ...\r\n`        |
| 整数结果       | 整数       | `:1\r\n`              |
| 字符串结果     | 批量字符串 | `$5\r\nAlice\r\n`     |
| 多个结果       | 数组       | `*2\r\n...\r\n`       |
| 键不存在/NULL  | NULL       | `$-1\r\n`             |

---

## 3. 功能需求

### 3.1 核心功能清单

| 功能                 | 优先级 | 说明                         |
| -------------------- | ------ | ---------------------------- |
| 解析简单字符串       | P0     | 解析 `+` 开头的简单字符串    |
| 解析错误             | P0     | 解析 `-` 开头的错误信息      |
| 解析整数             | P0     | 解析 `:` 开头的整数          |
| 解析批量字符串       | P0     | 解析 `$` 开头的批量字符串    |
| 解析数组             | P0     | 解析 `*` 开头的数组          |
| 解析 NULL 值         | P0     | 解析 `$-1` 和 `*-1`          |
| 序列化简单字符串     | P0     | 将字符串序列化为 `+...`      |
| 序列化错误           | P0     | 将错误序列化为 `-...`        |
| 序列化整数           | P0     | 将整数序列化为 `:...`        |
| 序列化批量字符串     | P0     | 将字符串序列化为 `$len...`   |
| 序列化数组           | P0     | 将数组序列化为 `*len...`     |
| 序列化 NULL          | P0     | 将 nil 序列化为 `$-1`        |

### 3.2 数据结构设计

#### Value 类型定义

```go
// Value 表示 RESP 协议中的一个值
type Value struct {
    Type   string      // "string", "error", "integer", "bulk", "array", "null"
    Str    string      // 用于 string, error, bulk
    Int    int64       // 用于 integer
    Array  []Value     // 用于 array
    IsNull bool        // 标识是否为 NULL
}
```

**设计说明**：
- 使用统一的 `Value` 结构体表示所有 RESP 类型
- `Type` 字段区分不同类型
- 不同字段用于存储不同类型的数据
- `IsNull` 明确标识 NULL 值

#### 解析器结构

```go
// Parser RESP 协议解析器
type Parser struct {
    reader *bufio.Reader
}

// NewParser 创建新的解析器
func NewParser(reader io.Reader) *Parser

// Parse 解析一个 RESP 值
func (p *Parser) Parse() (*Value, error)
```

#### 序列化器

```go
// Serialize 将 Value 序列化为 RESP 格式
func Serialize(v *Value) string

// 便捷函数
func SimpleString(s string) *Value
func Error(msg string) *Value
func Integer(n int64) *Value
func BulkString(s string) *Value
func Array(values []Value) *Value
func Null() *Value
```

---

## 4. 架构设计

### 4.1 整体架构

```
┌────────────────────────────────────┐
│       应用层 (Handler)             │
└────────────┬───────────────────────┘
             │ Value 对象
┌────────────▼───────────────────────┐
│       协议层 (Protocol)            │
│  ┌──────────────┬──────────────┐   │
│  │   Parser     │  Serializer  │   │
│  │  (解析器)    │  (序列化器)  │   │
│  └──────┬───────┴───────┬──────┘   │
│         │               │          │
└─────────┼───────────────┼──────────┘
          │               │
┌─────────▼───────────────▼──────────┐
│     网络层 (TCP Reader/Writer)     │
└────────────────────────────────────┘
```

### 4.2 解析流程

```
输入字节流
    │
    ▼
读取首字符 (前缀)
    │
    ├─ '+' → 解析简单字符串
    ├─ '-' → 解析错误
    ├─ ':' → 解析整数
    ├─ '$' → 解析批量字符串
    │         ├─ 长度 = -1 → NULL
    │         └─ 长度 >= 0 → 读取内容
    └─ '*' → 解析数组
              ├─ 长度 = -1 → NULL 数组
              ├─ 长度 = 0  → 空数组
              └─ 长度 > 0  → 递归解析每个元素
    │
    ▼
返回 Value 对象
```

### 4.3 核心算法

#### 解析批量字符串

```go
func (p *Parser) parseBulkString() (*Value, error) {
    // 1. 读取长度行：$6\r\n
    line, err := p.readLine()
    if err != nil {
        return nil, err
    }

    // 2. 解析长度
    length, err := strconv.Atoi(line[1:]) // 跳过 '$'
    if err != nil {
        return nil, err
    }

    // 3. 处理 NULL 情况
    if length == -1 {
        return &Value{Type: "null", IsNull: true}, nil
    }

    // 4. 读取指定长度的内容
    content := make([]byte, length)
    _, err = io.ReadFull(p.reader, content)
    if err != nil {
        return nil, err
    }

    // 5. 读取并验证结尾的 \r\n
    p.reader.ReadByte() // \r
    p.reader.ReadByte() // \n

    return &Value{Type: "bulk", Str: string(content)}, nil
}
```

#### 解析数组（递归）

```go
func (p *Parser) parseArray() (*Value, error) {
    // 1. 读取长度行
    line, err := p.readLine()
    if err != nil {
        return nil, err
    }

    // 2. 解析元素数量
    count, err := strconv.Atoi(line[1:])
    if err != nil {
        return nil, err
    }

    // 3. 处理特殊情况
    if count == -1 {
        return &Value{Type: "null", IsNull: true}, nil
    }
    if count == 0 {
        return &Value{Type: "array", Array: []Value{}}, nil
    }

    // 4. 递归解析每个元素
    array := make([]Value, count)
    for i := 0; i < count; i++ {
        value, err := p.Parse() // 递归调用
        if err != nil {
            return nil, err
        }
        array[i] = *value
    }

    return &Value{Type: "array", Array: array}, nil
}
```

---

## 5. 测试计划

### 5.1 测试策略

继续采用 **TDD** 方式：
1. 为每种 RESP 类型编写测试
2. 测试正常情况和边界情况
3. 测试错误处理
4. 测试嵌套结构

### 5.2 测试用例清单

#### TC1: 解析简单字符串 (TestParseSimpleString)

**测试目标**：验证简单字符串的解析

**测试数据**：
```
输入："+OK\r\n"
期望：Value{Type: "string", Str: "OK"}

输入："+PONG\r\n"
期望：Value{Type: "string", Str: "PONG"}
```

**边界条件**：
- 空字符串：`+\r\n`
- 包含空格：`+Hello World\r\n`

---

#### TC2: 解析错误 (TestParseError)

**测试目标**：验证错误信息的解析

**测试数据**：
```
输入："-ERR unknown command\r\n"
期望：Value{Type: "error", Str: "ERR unknown command"}

输入："-WRONGTYPE Operation against a key\r\n"
期望：Value{Type: "error", Str: "WRONGTYPE Operation against a key"}
```

---

#### TC3: 解析整数 (TestParseInteger)

**测试目标**：验证整数的解析

**测试数据**：
```
输入：":0\r\n"
期望：Value{Type: "integer", Int: 0}

输入：":1000\r\n"
期望：Value{Type: "integer", Int: 1000}

输入：":-123\r\n"
期望：Value{Type: "integer", Int: -123}
```

**边界条件**：
- 负数
- 零
- 大数值

---

#### TC4: 解析批量字符串 (TestParseBulkString)

**测试目标**：验证批量字符串的解析

**测试数据**：
```
输入：
$6\r\n
foobar\r\n
期望：Value{Type: "bulk", Str: "foobar"}

输入：
$0\r\n
\r\n
期望：Value{Type: "bulk", Str: ""}

输入：
$-1\r\n
期望：Value{Type: "null", IsNull: true}
```

**边界条件**：
- 空字符串（长度为 0）
- NULL（长度为 -1）
- 包含 \r\n 的字符串
- 二进制数据

---

#### TC5: 解析数组 (TestParseArray)

**测试目标**：验证数组的解析

**测试数据**：
```
# 简单数组 ["foo", "bar"]
输入：
*2\r\n
$3\r\n
foo\r\n
$3\r\n
bar\r\n
期望：Value{
    Type: "array",
    Array: []Value{
        {Type: "bulk", Str: "foo"},
        {Type: "bulk", Str: "bar"},
    }
}

# 混合类型数组 [1, "foo"]
输入：
*2\r\n
:1\r\n
$3\r\n
foo\r\n
期望：Value{
    Type: "array",
    Array: []Value{
        {Type: "integer", Int: 1},
        {Type: "bulk", Str: "foo"},
    }
}

# 空数组
输入：*0\r\n
期望：Value{Type: "array", Array: []Value{}}

# NULL 数组
输入：*-1\r\n
期望：Value{Type: "null", IsNull: true}
```

**边界条件**：
- 空数组
- NULL 数组
- 嵌套数组
- 包含 NULL 元素的数组

---

#### TC6: 解析嵌套数组 (TestParseNestedArray)

**测试目标**：验证嵌套数组的解析

**测试数据**：
```
# 数组 [[1, 2], [3, 4]]
输入：
*2\r\n
*2\r\n
:1\r\n
:2\r\n
*2\r\n
:3\r\n
:4\r\n

期望：嵌套的 Value 结构
```

---

#### TC7: 解析命令 (TestParseCommand)

**测试目标**：验证实际命令的解析

**测试数据**：
```
# SET name Alice
输入：
*3\r\n
$3\r\nSET\r\n
$4\r\nname\r\n
$5\r\nAlice\r\n

期望：Value{
    Type: "array",
    Array: []Value{
        {Type: "bulk", Str: "SET"},
        {Type: "bulk", Str: "name"},
        {Type: "bulk", Str: "Alice"},
    }
}
```

---

#### TC8: 序列化测试 (TestSerialize)

**测试目标**：验证各种类型的序列化

**测试数据**：
```go
输入：SimpleString("OK")
期望："+OK\r\n"

输入：Error("ERR unknown")
期望："-ERR unknown\r\n"

输入：Integer(123)
期望：":123\r\n"

输入：BulkString("foobar")
期望："$6\r\nfoobar\r\n"

输入：Null()
期望："$-1\r\n"

输入：Array([]Value{
    *BulkString("SET"),
    *BulkString("key"),
    *BulkString("value"),
})
期望："*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
```

---

#### TC9: 往返测试 (TestRoundTrip)

**测试目标**：验证解析和序列化的一致性

**测试逻辑**：
```go
原始字符串 → 解析 → Value → 序列化 → 新字符串
验证：新字符串 == 原始字符串
```

---

#### TC10: 错误处理测试 (TestParseErrors)

**测试目标**：验证错误输入的处理

**测试数据**：
```
- 不完整的输入（缺少 \r\n）
- 无效的前缀字符
- 批量字符串长度不匹配
- 数组元素数量不匹配
- 无效的整数格式
- EOF 错误
```

---

### 5.3 性能测试

```go
func BenchmarkParseSimpleString(b *testing.B)
func BenchmarkParseBulkString(b *testing.B)
func BenchmarkParseArray(b *testing.B)
func BenchmarkSerialize(b *testing.B)
```

**性能目标**：
- 解析简单类型：< 100 ns/op
- 解析复杂类型：< 500 ns/op
- 序列化：< 200 ns/op

---

## 6. 验收标准

### 6.1 功能验收

- [ ] 所有 RESP 数据类型都能正确解析
- [ ] 所有 RESP 数据类型都能正确序列化
- [ ] 支持嵌套数组
- [ ] 正确处理 NULL 值
- [ ] 所有测试用例通过
- [ ] 测试覆盖率 ≥ 95%

### 6.2 质量验收

- [ ] 代码通过 `go fmt` 格式化
- [ ] 代码通过 `go vet` 静态检查
- [ ] 无明显性能问题
- [ ] 完整的文档注释
- [ ] 良好的错误处理

### 6.3 正确性验收

- [ ] 能够解析 redis-cli 发送的命令
- [ ] 序列化的响应能被 redis-cli 正确显示
- [ ] 往返测试（parse → serialize）结果一致
- [ ] 二进制安全（能处理任意字节）

---

## 7. 实现提示

### 7.1 开发顺序建议

1. **定义数据结构** → Value, Parser
2. **实现辅助函数** → readLine, readBytes
3. **实现简单类型解析** → SimpleString, Error, Integer
4. **实现批量字符串** → BulkString (含 NULL)
5. **实现数组解析** → Array (递归)
6. **实现序列化** → Serialize 函数
7. **测试验证** → 所有测试通过

### 7.2 关键技术点

#### 使用 bufio.Reader

```go
import "bufio"

// 创建带缓冲的读取器
reader := bufio.NewReader(conn)

// 读取一行（直到 \r\n）
line, err := reader.ReadString('\n')
line = strings.TrimSuffix(line, "\r\n")

// 精确读取 N 个字节
buf := make([]byte, n)
_, err := io.ReadFull(reader, buf)
```

#### 处理 \r\n

RESP 协议使用 `\r\n` 作为行分隔符，需要正确处理：

```go
// 读取一行并去除 \r\n
func (p *Parser) readLine() (string, error) {
    line, err := p.reader.ReadString('\n')
    if err != nil {
        return "", err
    }
    // 去除 \r\n
    return strings.TrimSuffix(line, "\r\n"), nil
}
```

#### 二进制安全

批量字符串必须是二进制安全的：

```go
// 错误做法：使用 ReadString
content, _ := reader.ReadString('\n') // ❌ 如果内容包含 \n 会提前结束

// 正确做法：使用 ReadFull
content := make([]byte, length)
io.ReadFull(reader, content) // ✅ 精确读取指定字节数
```

### 7.3 常见陷阱

1. **忘记处理 \r\n**
   - 批量字符串内容后有 `\r\n` 需要读取并丢弃

2. **长度计算错误**
   - 批量字符串的长度是**字节数**，不是字符数
   - 对于 UTF-8 字符串，`len("中文")` = 6（字节），不是 2

3. **递归解析错误**
   - 解析数组时，要为每个元素递归调用 `Parse()`
   - 注意处理嵌套数组的情况

4. **NULL 值处理**
   - `$-1\r\n` 表示 NULL 批量字符串
   - `*-1\r\n` 表示 NULL 数组
   - 需要与空字符串/空数组区分

### 7.4 调试技巧

```bash
# 使用 telnet 测试
telnet localhost 6379
*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n

# 使用 printf 生成测试数据
printf "*3\r\n\$3\r\nSET\r\n\$3\r\nkey\r\n\$5\r\nvalue\r\n" | nc localhost 6379

# 查看字节
echo -n "+OK\r\n" | od -c
echo -n "+OK\r\n" | xxd

# 运行测试
go test ./protocol -v
go test ./protocol -cover
```

### 7.5 代码示例结构

```
protocol/
├── value.go           # Value 结构体定义
├── parser.go          # Parser 实现
├── serializer.go      # Serialize 函数实现
├── helpers.go         # 辅助函数（SimpleString, Error 等）
├── parser_test.go     # 解析器测试
└── serializer_test.go # 序列化器测试
```

---

## 8. 扩展思考

完成基础功能后，可以思考：

1. **如何优化性能？**
   - 使用对象池减少内存分配
   - 缓冲区复用
   - 避免不必要的字符串拷贝

2. **如何处理流式数据？**
   - 当前实现一次解析一个完整 Value
   - 如何处理网络分包？
   - 如何实现管道（pipelining）？

3. **如何支持 RESP3？**
   - RESP3 增加了新的数据类型
   - Map、Set、Boolean 等
   - 如何兼容 RESP2？

4. **错误恢复**
   - 解析失败后如何恢复？
   - 如何跳过损坏的数据？

---

## 9. 参考资料

- [RESP 协议规范](https://redis.io/docs/reference/protocol-spec/)
- [Go bufio 包文档](https://pkg.go.dev/bufio)
- [Go io 包文档](https://pkg.go.dev/io)
- [Redis 命令参考](https://redis.io/commands/)

---

## 10. 交付物

完成本阶段后，应该交付：

1. [ ] `protocol/value.go` - Value 结构体定义
2. [ ] `protocol/parser.go` - 解析器实现
3. [ ] `protocol/serializer.go` - 序列化器实现
4. [ ] `protocol/helpers.go` - 辅助函数
5. [ ] `protocol/parser_test.go` - 解析器测试
6. [ ] `protocol/serializer_test.go` - 序列化器测试
7. [ ] 所有测试通过的截图或日志
8. [ ] 覆盖率报告（≥ 95%）

准备好后，进入第三阶段：命令处理层实现。

---

## 附录：完整示例

### A.1 客户端命令示例

```
命令：PING
RESP：
*1\r\n
$4\r\n
PING\r\n

命令：SET mykey "Hello World"
RESP：
*3\r\n
$3\r\n
SET\r\n
$5\r\n
mykey\r\n
$11\r\n
Hello World\r\n

命令：GET mykey
RESP：
*2\r\n
$3\r\n
GET\r\n
$5\r\n
mykey\r\n
```

### A.2 服务器响应示例

```
PING → +PONG\r\n
SET  → +OK\r\n
GET  → $11\r\nHello World\r\n (键存在)
GET  → $-1\r\n (键不存在)
DEL  → :2\r\n (删除了 2 个键)
错误 → -ERR unknown command 'FOOBAR'\r\n
```

### A.3 测试用例示例

```go
func TestParseCommand(t *testing.T) {
    // 模拟客户端发送 SET name Alice
    input := "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nAlice\r\n"

    parser := NewParser(strings.NewReader(input))
    value, err := parser.Parse()

    if err != nil {
        t.Fatalf("Parse error: %v", err)
    }

    if value.Type != "array" {
        t.Errorf("Expected array, got %s", value.Type)
    }

    if len(value.Array) != 3 {
        t.Errorf("Expected 3 elements, got %d", len(value.Array))
    }

    if value.Array[0].Str != "SET" {
        t.Errorf("Expected SET, got %s", value.Array[0].Str)
    }

    // ... 更多断言
}
```

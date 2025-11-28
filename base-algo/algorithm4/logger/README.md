# Logger 日志包

基于 logrus 封装的日志库，提供结构化日志、文件名/行号显示、彩色输出等功能。

## 功能特性

✅ **自动显示调用位置**：文件名 + 行号
✅ **时间戳**：完整的日期时间格式
✅ **彩色输出**：不同日志级别使用不同颜色
✅ **结构化日志**：支持添加字段（Fields）
✅ **多日志级别**：Debug、Info、Warn、Error、Fatal
✅ **线程安全**：可在并发环境中安全使用
✅ **路径简化**：只显示相对于项目的路径

## 快速开始

### 1. 基本使用

```go
import "go-redis/logger"

func main() {
    // 不同级别的日志
    logger.Debug("这是调试信息")
    logger.Info("这是普通信息")
    logger.Warn("这是警告信息")
    logger.Error("这是错误信息")

    // 格式化输出
    logger.Debugf("用户 %s 登录成功", "Alice")
    logger.Infof("处理请求耗时 %d ms", 150)
}
```

### 2. 结构化日志（推荐）

```go
import (
    "go-redis/logger"
    "github.com/sirupsen/logrus"
)

// 单个字段
logger.WithField("user_id", "123").Info("用户登录")

// 多个字段
logger.WithFields(logrus.Fields{
    "method":  "GET",
    "path":    "/api/users",
    "status":  200,
    "latency": "15ms",
}).Info("HTTP 请求完成")

// 链式调用
logger.WithField("operation", "SET").
    WithField("key", "user:456").
    WithField("value", "data").
    Debug("执行存储操作")
```

### 3. 设置日志级别

```go
import (
    "go-redis/logger"
    "github.com/sirupsen/logrus"
)

// 生产环境：只显示 Info 及以上级别
logger.SetLevel(logrus.InfoLevel)

// 开发环境：显示所有日志（包括 Debug）
logger.SetLevel(logrus.DebugLevel)

// 只显示错误
logger.SetLevel(logrus.ErrorLevel)
```

## 日志级别说明

| 级别  | 用途                     | 示例场景                    |
| ----- | ------------------------ | --------------------------- |
| Debug | 调试信息，开发时使用     | 函数入口/出口、变量值       |
| Info  | 普通信息，记录关键操作   | 服务启动、请求处理、业务操作 |
| Warn  | 警告信息，需要注意的问题 | 磁盘空间不足、配置问题       |
| Error | 错误信息，操作失败       | 数据库连接失败、文件读取失败 |
| Fatal | 致命错误，程序退出       | 无法启动服务、关键资源缺失   |

## 日志输出示例

运行 `go test ./logger -v` 查看实际输出：

```
DEBU[2025-11-27 10:30:45] 这是一条 Debug 日志                           func=logger.TestLoggerOutput file="logger/logger_test.go:10"
INFO[2025-11-27 10:30:45] 这是一条 Info 日志                            func=logger.TestLoggerOutput file="logger/logger_test.go:11"
WARN[2025-11-27 10:30:45] 这是一条 Warn 日志                            func=logger.TestLoggerOutput file="logger/logger_test.go:12"
ERRO[2025-11-27 10:30:45] 这是一条 Error 日志                           func=logger.TestLoggerOutput file="logger/logger_test.go:13"
DEBU[2025-11-27 10:30:45] 用户 Alice 执行了 SET 操作                    func=logger.TestLoggerOutput file="logger/logger_test.go:16"
INFO[2025-11-27 10:30:45] 设置键值对: key=count, value=42               func=logger.TestLoggerOutput file="logger/logger_test.go:17"
INFO[2025-11-27 10:30:45] 用户登录                                      func=logger.TestLoggerWithFields file="logger/logger_test.go:24" user=Bob
INFO[2025-11-27 10:30:45] 设置键值对成功                                func=logger.TestLoggerWithFields file="logger/logger_test.go:30" key=user:123 ttl=3600 value=data
```

## 在 Store 中的使用

查看 `store/store.go` 了解实际使用案例：

```go
func (s *Store) Set(key string, value interface{}) {
    logger.WithFields(logrus.Fields{
        "operation": "SET",
        "key":       key,
    }).Debug("执行 Set 操作")

    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[key] = value

    logger.WithField("key", key).Debug("Set 操作完成")
}
```

## 高级用法

### 错误处理

```go
if err := someOperation(); err != nil {
    logger.WithFields(logrus.Fields{
        "error":     err.Error(),
        "operation": "someOperation",
    }).Error("操作失败")
}
```

### 性能统计

```go
start := time.Now()
// ... 执行操作 ...
duration := time.Since(start)

logger.WithFields(logrus.Fields{
    "operation":  "query",
    "latency_ms": duration.Milliseconds(),
}).Debug("性能统计")
```

### 并发日志

```go
// logrus 是线程安全的
for i := 0; i < 10; i++ {
    go func(id int) {
        logger.WithField("worker_id", id).Info("Worker 开始工作")
    }(i)
}
```

## 测试时控制日志输出

### 方法 1：调整日志级别

```go
import "github.com/sirupsen/logrus"

func TestSomething(t *testing.T) {
    // 测试时只显示错误日志
    logger.SetLevel(logrus.ErrorLevel)

    // ... 测试代码 ...

    // 恢复到 Debug 级别
    logger.SetLevel(logrus.DebugLevel)
}
```

### 方法 2：完全禁用日志输出

```go
import (
    "io"
    "go-redis/logger"
)

func TestSomething(t *testing.T) {
    // 禁用日志输出
    logger.SetOutput(io.Discard)

    // ... 测试代码 ...

    // 恢复到标准输出
    logger.SetOutput(os.Stdout)
}
```

## 运行测试

```bash
# 运行日志测试，查看所有日志输出
go test ./logger -v

# 运行 Store 测试，查看带日志的操作
go test ./store -v

# 只运行特定测试
go test ./logger -v -run TestLoggerWithFields
```

## 性能说明

- **Debug 日志**：只在开发/调试时启用，生产环境应设置为 Info 级别
- **结构化日志**：使用 `WithFields` 比字符串拼接更高效
- **Benchmark 测试**：建议在 benchmark 测试时禁用日志或设置为 Error 级别

## 最佳实践

### ✅ 推荐做法

```go
// 1. 使用结构化日志
logger.WithFields(logrus.Fields{
    "user_id": userID,
    "action":  "login",
}).Info("用户操作")

// 2. 关键操作记录 Info 级别
logger.Info("服务启动成功")

// 3. 调试信息使用 Debug 级别
logger.Debug("变量值检查")

// 4. 错误包含上下文信息
logger.WithField("error", err.Error()).Error("操作失败")
```

### ❌ 不推荐做法

```go
// 1. 避免过度日志
for i := 0; i < 10000; i++ {
    logger.Debug("循环", i)  // ❌ 会产生大量日志
}

// 2. 避免敏感信息
logger.Infof("用户密码: %s", password)  // ❌ 安全问题

// 3. 避免在高频操作中使用日志
func Get(key string) {
    logger.Debug("get")  // ❌ 每次Get都记录影响性能
}
```

## 常见问题

### Q: 为什么看不到 Debug 日志？
A: 检查日志级别是否设置为 `DebugLevel`：
```go
logger.SetLevel(logrus.DebugLevel)
```

### Q: 如何在测试时隐藏日志？
A: 使用 `logger.SetLevel(logrus.ErrorLevel)` 或 `logger.SetOutput(io.Discard)`

### Q: 日志会影响性能吗？
A: Debug 日志有一定影响，生产环境建议设置为 Info 级别

### Q: 可以输出到文件吗？
A: 可以，使用 `SetOutput`：
```go
file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
logger.SetOutput(file)
```

## 参考资料

- [logrus 官方文档](https://github.com/sirupsen/logrus)
- [Go 标准库 log](https://pkg.go.dev/log)

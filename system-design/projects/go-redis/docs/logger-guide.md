# Logger 快速使用指南

## 已完成的工作

✅ 安装 logrus 日志库
✅ 创建封装的 logger 包
✅ 集成到 Store 模块
✅ 创建完整的测试用例
✅ 创建使用示例和文档

## 日志输出特性

你现在的日志包含以下信息：

1. **时间戳**：`2025-11-27 10:19:32`
2. **文件位置**：`go-redis/store/store.go:31`
3. **函数名**：`go-redis/store.(*Store).Set()`
4. **日志级别**：`DEBUG/INFO/WARN/ERROR`（带颜色）
5. **结构化字段**：`key=user:1 operation=SET`

## 使用方式

### 在你的代码中使用

```go
import "go-redis/logger"

// 1. 基本使用
logger.Debug("调试信息")
logger.Info("普通信息")
logger.Warn("警告信息")
logger.Error("错误信息")

// 2. 格式化输出
logger.Infof("用户 %s 执行了 %s 操作", username, operation)

// 3. 结构化日志（推荐）
logger.WithFields(logrus.Fields{
    "user_id": "123",
    "action":  "login",
}).Info("用户操作")
```

### 运行演示程序

```bash
# 查看完整的日志演示
go run examples/logger_demo.go
```

### 测试时查看日志

```bash
# 运行 logger 测试
go test ./logger -v

# 运行 Store 测试（会显示操作日志）
go test ./store -v -run TestSetAndGet
```

### 控制日志级别

#### 开发环境（显示所有日志）
```go
logger.SetLevel(logrus.DebugLevel)
```

#### 生产环境（只显示重要日志）
```go
logger.SetLevel(logrus.InfoLevel)
```

#### 测试时禁用日志
```go
import "io"

func TestSomething(t *testing.T) {
    // 禁用日志输出
    logger.SetOutput(io.Discard)

    // 测试代码...

    // 恢复输出
    logger.SetOutput(os.Stdout)
}
```

## 文件结构

```
go-redis/
├── logger/
│   ├── logger.go          # 日志实现
│   ├── logger_test.go     # 测试文件
│   ├── example.go         # 使用示例
│   └── README.md          # 详细文档
├── store/
│   └── store.go           # 已集成日志
├── examples/
│   └── logger_demo.go     # 演示程序
└── docs/
    └── logger-guide.md    # 本文档
```

## 日志级别说明

| 级别  | 用途                 | 使用场景                     |
| ----- | -------------------- | ---------------------------- |
| Debug | 详细调试信息         | 开发时追踪程序执行流程       |
| Info  | 关键信息             | 记录重要的业务操作           |
| Warn  | 警告信息             | 可恢复的错误或需要注意的情况 |
| Error | 错误信息             | 操作失败，但程序可以继续运行 |
| Fatal | 致命错误（程序退出） | 无法继续运行的严重错误       |

## 常用命令

```bash
# 1. 查看日志功能演示
go run examples/logger_demo.go

# 2. 运行所有日志测试
go test ./logger -v

# 3. 运行带日志的 Store 测试
go test ./store -v

# 4. 只看 INFO 级别日志（需在代码中设置）
# logger.SetLevel(logrus.InfoLevel)

# 5. 测试并发安全时的日志输出
go test ./store -v -run TestConcurrentAccess
```

## 性能建议

### ✅ 推荐

- 生产环境使用 `Info` 级别
- 使用结构化日志（`WithFields`）
- 关键操作记录日志
- Benchmark 测试时禁用日志

### ❌ 避免

- 在高频循环中打日志
- 记录敏感信息（密码、密钥等）
- 过度使用 Debug 日志
- 记录大量数据

## 下一步

1. 根据需要调整日志级别
2. 在新功能中添加日志记录
3. 查看 `logger/README.md` 了解更多高级用法
4. 在开发时使用日志辅助调试

## 参考文档

- 详细文档：`logger/README.md`
- 使用示例：`logger/example.go`
- 演示程序：`examples/logger_demo.go`
- logrus 官方文档：https://github.com/sirupsen/logrus

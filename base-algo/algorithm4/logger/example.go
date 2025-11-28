package logger

import "github.com/sirupsen/logrus"

// ExampleUsage 展示如何使用 logger 包
func ExampleUsage() {
	// 1. 基本使用
	Debug("应用启动")
	Info("Redis 服务器监听在 :6379")
	Warn("磁盘使用率超过 80%")
	Error("连接数据库失败")

	// 2. 格式化输出
	username := "Alice"
	operation := "SET"
	Debugf("用户 %s 执行了 %s 操作", username, operation)

	key := "user:123"
	value := 42
	Infof("设置键值对: %s = %d", key, value)

	// 3. 带字段的结构化日志（推荐）
	WithField("user_id", "123").Info("用户登录成功")

	WithFields(logrus.Fields{
		"method":  "POST",
		"path":    "/api/users",
		"status":  200,
		"latency": "15ms",
	}).Info("HTTP 请求处理完成")

	// 4. 链式调用
	WithField("operation", "GET").
		WithField("key", "user:456").
		WithField("hit", true).
		Debug("缓存命中")

	// 5. 设置日志级别
	SetLevel(logrus.InfoLevel) // 只显示 Info 及以上级别
	Debug("这条日志不会显示")
	Info("这条日志会显示")

	// 恢复 Debug 级别
	SetLevel(logrus.DebugLevel)

	// 6. 错误处理示例
	if err := someOperation(); err != nil {
		WithField("error", err.Error()).Error("操作失败")
		// 注意：Fatal 会终止程序，慎用
		// Fatal("致命错误，程序退出")
	}
}

// someOperation 模拟一个操作
func someOperation() error {
	// 模拟业务逻辑
	return nil
}

// StoreOperationExample 展示在 Store 操作中使用日志
func StoreOperationExample() {
	key := "user:100"
	value := "Alice"

	// 操作开始
	WithFields(logrus.Fields{
		"operation": "SET",
		"key":       key,
	}).Debug("开始执行 SET 操作")

	// 模拟存储操作
	// store.Set(key, value)

	// 操作成功
	WithFields(logrus.Fields{
		"key":   key,
		"value": value,
	}).Info("键值对设置成功")

	// 性能日志
	WithFields(logrus.Fields{
		"operation":  "SET",
		"latency_ns": 1500,
	}).Debug("操作性能统计")
}

// ConcurrentLoggingExample 展示并发环境下的日志使用
func ConcurrentLoggingExample() {
	// logrus 是线程安全的，可以在多个 goroutine 中安全使用

	// 模拟并发操作
	for i := 0; i < 5; i++ {
		WithFields(logrus.Fields{
			"worker_id": i,
			"task":      "process_data",
		}).Info("Worker 开始处理任务")
	}
}

// ErrorRecoveryExample 展示错误恢复和日志记录
func ErrorRecoveryExample() {
	defer func() {
		if r := recover(); r != nil {
			WithFields(logrus.Fields{
				"panic": r,
			}).Error("捕获到 panic，已恢复")
		}
	}()

	// 模拟可能 panic 的代码
	// ...
}

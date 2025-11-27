package main

import (
	"go-redis/logger"
	"go-redis/store"

	"github.com/sirupsen/logrus"
)

func main() {
	logger.Info("========== Go-Redis 日志演示程序 ==========")

	// 1. 基本日志使用
	demonstrateBasicLogging()

	// 2. 结构化日志
	demonstrateStructuredLogging()

	// 3. Store 操作日志
	demonstrateStoreLogging()

	// 4. 日志级别控制
	demonstrateLogLevels()

	logger.Info("========== 演示完成 ==========")
}

// demonstrateBasicLogging 演示基本日志功能
func demonstrateBasicLogging() {
	logger.Info("\n--- 1. 基本日志功能 ---")

	logger.Debug("这是 Debug 级别日志，用于调试")
	logger.Info("这是 Info 级别日志，用于记录关键信息")
	logger.Warn("这是 Warn 级别日志，用于警告")
	logger.Error("这是 Error 级别日志，用于错误")

	// 格式化输出
	username := "Alice"
	operation := "LOGIN"
	logger.Infof("用户 %s 执行了 %s 操作", username, operation)
}

// demonstrateStructuredLogging 演示结构化日志
func demonstrateStructuredLogging() {
	logger.Info("\n--- 2. 结构化日志（推荐方式）---")

	// 单个字段
	logger.WithField("user_id", "12345").Info("用户登录")

	// 多个字段
	logger.WithFields(logrus.Fields{
		"method":     "POST",
		"path":       "/api/users",
		"status":     200,
		"latency_ms": 15,
	}).Info("HTTP 请求处理完成")

	// 链式调用
	logger.WithField("operation", "QUERY").
		WithField("table", "users").
		WithField("rows", 100).
		Debug("数据库查询完成")
}

// demonstrateStoreLogging 演示 Store 操作日志
func demonstrateStoreLogging() {
	logger.Info("\n--- 3. Store 操作日志 ---")

	// 创建 Store 实例（会记录日志）
	s := store.NewStore()

	// 执行一些操作，观察日志输出
	logger.Info("执行一系列 Store 操作...")

	s.Set("user:1", "Alice")
	s.Set("user:2", "Bob")
	s.Set("count", 42)

	value, exists := s.Get("user:1")
	if exists {
		logger.WithFields(logrus.Fields{
			"key":   "user:1",
			"value": value,
		}).Info("成功获取值")
	}

	keys := s.Keys()
	logger.WithField("total_keys", len(keys)).Info("获取所有键")

	s.Delete("count")
	logger.Info("删除键: count")
}

// demonstrateLogLevels 演示日志级别控制
func demonstrateLogLevels() {
	logger.Info("\n--- 4. 日志级别控制 ---")

	logger.Info("当前日志级别：Debug（显示所有日志）")
	logger.Debug("Debug 日志可见")
	logger.Info("Info 日志可见")

	logger.Info("\n切换到 Info 级别（隐藏 Debug 日志）...")
	logger.SetLevel(logrus.InfoLevel)

	logger.Debug("这条 Debug 日志不会显示")
	logger.Info("Info 日志仍然可见")

	logger.Info("\n切换回 Debug 级别...")
	logger.SetLevel(logrus.DebugLevel)
	logger.Debug("Debug 日志又可见了")
}

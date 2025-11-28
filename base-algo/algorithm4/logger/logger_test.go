package logger

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestLoggerOutput(t *testing.T) {
	// 测试不同级别的日志输出
	Debug("这是一条 Debug 日志")
	Info("这是一条 Info 日志")
	Warn("这是一条 Warn 日志")
	Error("这是一条 Error 日志")

	// 测试格式化输出
	Debugf("用户 %s 执行了 %s 操作", "Alice", "SET")
	Infof("存储了键值对: key=%s, value=%d", "count", 42)
	Warnf("键 %s 即将过期", "session:123")
	Errorf("操作失败: %s", "连接超时")
}

func TestLoggerWithFields(t *testing.T) {
	// 测试带字段的日志
	WithField("user", "Bob").Info("用户登录")

	WithFields(logrus.Fields{
		"key":   "user:123",
		"value": "data",
		"ttl":   3600,
	}).Info("设置键值对成功")

	// 链式调用
	WithField("operation", "GET").
		WithField("key", "user:456").
		WithField("latency_ms", 2.5).
		Debug("执行查询操作")
}

func TestLoggerLevels(t *testing.T) {
	// 测试设置不同的日志级别

	// 设置为 Info 级别（不显示 Debug）
	SetLevel(logrus.InfoLevel)
	Debug("这条 Debug 不会显示")
	Info("这条 Info 会显示")

	// 设置回 Debug 级别
	SetLevel(logrus.DebugLevel)
	Debug("现在 Debug 又可以显示了")
}

func TestGetCallerInfo(t *testing.T) {
	// 测试获取调用者信息
	file, line, function := GetCallerInfo(1)

	t.Logf("文件: %s", file)
	t.Logf("行号: %d", line)
	t.Logf("函数: %s", function)

	if file == "" || line == 0 || function == "" {
		t.Error("获取调用者信息失败")
	}
}

func TestLoggerInFunction(t *testing.T) {
	// 模拟在函数中使用日志
	simulateStoreOperation("user:100", "Alice")
}

func simulateStoreOperation(key string, value interface{}) {
	Debug("开始执行 Set 操作")

	WithFields(logrus.Fields{
		"key":   key,
		"value": value,
	}).Info("设置键值对")

	Debugf("键 %s 已成功存储", key)
}

// 示例：模拟错误处理
func TestErrorHandling(t *testing.T) {
	key := "test:key"

	// 模拟一个错误场景
	if err := simulateError(key); err != nil {
		WithField("key", key).
			WithField("error", err.Error()).
			Error("操作失败")
	}
}

func simulateError(key string) error {
	Warnf("键 %s 不存在", key)
	// 模拟返回错误（实际项目中这里会是真实的错误）
	return nil
}

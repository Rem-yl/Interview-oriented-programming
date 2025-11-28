package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	// Log 是全局的日志实例
	Log *logrus.Logger
)

// CallerHook 用于修正调用者信息，跳过 logger 包装层
type CallerHook struct{}

func (hook *CallerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *CallerHook) Fire(entry *logrus.Entry) error {
	// 向上查找调用栈，跳过 logger 包内的函数
	pcs := make([]uintptr, 10)
	depth := runtime.Callers(1, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for {
		frame, more := frames.Next()
		// 跳过 logger 包和 logrus 包的函数
		if !strings.Contains(frame.File, "/logger/logger.go") &&
			!strings.Contains(frame.File, "/sirupsen/logrus") {
			entry.Caller = &frame
			break
		}
		if !more {
			break
		}
	}
	return nil
}

func init() {
	Log = logrus.New()

	// 设置输出到标准输出
	Log.SetOutput(os.Stdout)

	// 设置日志级别为 Debug（开发时显示所有日志）
	Log.SetLevel(logrus.DebugLevel)

	// 使用自定义的格式化器，显示文件名和行号
	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,                  // 显示完整时间戳
		TimestampFormat: "2006-01-02 15:04:05", // 时间格式
		ForceColors:     true,                  // 强制使用彩色输出
		DisableQuote:    true,                  // 禁用引号包裹
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			// 简化文件路径并添加行号
			filename := fmt.Sprintf("%s:%d", shortenFilePath(f.File), f.Line)
			// 简化函数名
			funcName := f.Function
			if idx := strings.LastIndex(funcName, "/"); idx != -1 {
				funcName = funcName[idx+1:]
			}
			return funcName, filename
		},
	})

	// 启用调用者报告
	Log.SetReportCaller(true)

	// 添加 Hook 来修正调用者信息
	Log.AddHook(&CallerHook{})
}

// shortenFilePath 缩短文件路径，只保留相对路径
func shortenFilePath(file string) string {
	// 查找 "go-redis" 字符串的位置
	idx := strings.LastIndex(file, "go-redis")
	if idx != -1 {
		// 从 go-redis 开始截取
		return file[idx:]
	}

	// 如果没找到，就返回文件名和上一级目录
	dir := filepath.Dir(file)
	parentDir := filepath.Base(dir)
	fileName := filepath.Base(file)
	return filepath.Join(parentDir, fileName)
}

// SetLevel 设置日志级别
func SetLevel(level logrus.Level) {
	Log.SetLevel(level)
}

// SetOutput 设置日志输出位置
func SetOutput(output io.Writer) {
	Log.SetOutput(output)
}

// Debug 输出 Debug 级别日志
func Debug(args ...interface{}) {
	Log.Debug(args...)
}

// Debugf 格式化输出 Debug 级别日志
func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

// Info 输出 Info 级别日志
func Info(args ...interface{}) {
	Log.Info(args...)
}

// Infof 格式化输出 Info 级别日志
func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

// Warn 输出 Warn 级别日志
func Warn(args ...interface{}) {
	Log.Warn(args...)
}

// Warnf 格式化输出 Warn 级别日志
func Warnf(format string, args ...interface{}) {
	Log.Warnf(format, args...)
}

// Error 输出 Error 级别日志
func Error(args ...interface{}) {
	Log.Error(args...)
}

// Errorf 格式化输出 Error 级别日志
func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

// Fatal 输出 Fatal 级别日志并退出程序
func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

// Fatalf 格式化输出 Fatal 级别日志并退出程序
func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

// WithField 添加单个字段
func WithField(key string, value interface{}) *logrus.Entry {
	return Log.WithField(key, value)
}

// WithFields 添加多个字段
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}

// GetCallerInfo 获取调用者信息（文件名、行号、函数名）
func GetCallerInfo(skip int) (file string, line int, function string) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown", 0, "unknown"
	}

	function = runtime.FuncForPC(pc).Name()
	// 简化函数名（去掉包路径）
	if idx := strings.LastIndex(function, "/"); idx != -1 {
		function = function[idx+1:]
	}

	// 简化文件路径
	file = shortenFilePath(file)

	return file, line, function
}

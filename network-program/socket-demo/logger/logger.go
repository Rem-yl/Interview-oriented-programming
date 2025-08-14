package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// 日志等级
const (
	DEBUG = iota
	INFO
	ERROR
)

var (
	logger   *log.Logger
	logLevel = INFO // 默认等级
)

// SetLevel 设置日志等级
func SetLevel(level int) {
	logLevel = level
}

func init() {
	logger = log.New(os.Stdout, "", 0) // 不要 log 包自带时间戳
}

// 获取日志前缀（时间 + 文件 + 行号 + 函数名）
func logPrefix() string {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return "[Unknown]"
	}

	funcName := runtime.FuncForPC(pc).Name()
	funcName = funcName[strings.LastIndex(funcName, ".")+1:]

	shortFile := file[strings.LastIndex(file, "/")+1:]
	return fmt.Sprintf("[%s] [%s:%d] %s",
		time.Now().Format("2006-01-02 15:04:05"),
		shortFile,
		line,
		funcName,
	)
}

// Info 输出普通日志
func Info(v ...interface{}) {
	if logLevel <= INFO {
		logger.Println(append([]interface{}{logPrefix(), "[INFO] - "}, v...)...)
	}
}

// Infof 输出格式化日志
func Infof(format string, v ...interface{}) {
	if logLevel <= INFO {
		logger.Printf("%s [INFO] - "+format, append([]interface{}{logPrefix()}, v...)...)
	}
}

// Debug 输出调试日志
func Debug(v ...interface{}) {
	if logLevel <= DEBUG {
		logger.Println(append([]interface{}{logPrefix(), "[DEBUG] - "}, v...)...)
	}
}

// Debugf 输出格式化调试日志
func Debugf(format string, v ...interface{}) {
	if logLevel <= DEBUG {
		logger.Printf("%s [DEBUG] - "+format, append([]interface{}{logPrefix()}, v...)...)
	}
}

// Error 输出错误日志（始终打印）
func Error(v ...interface{}) {
	logger.Println(append([]interface{}{logPrefix(), "[ERROR] - "}, v...)...)
}

// Errorf 输出格式化错误日志（始终打印）
func Errorf(format string, v ...interface{}) {
	logger.Printf("%s [ERROR] - "+format, append([]interface{}{logPrefix()}, v...)...)
}

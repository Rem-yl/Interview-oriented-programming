package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

var logger *log.Logger

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
	logger.Println(append([]interface{}{logPrefix(), "[INFO] - "}, v...)...)
}

// Infof 输出格式化日志
func Infof(format string, v ...interface{}) {
	logger.Printf("%s [INFO] - "+format, append([]interface{}{logPrefix()}, v...)...)
}

func Debug(v ...interface{}) {
	logger.Println(append([]interface{}{logPrefix(), "[DEBUG] - "}, v...)...)
}

// Infof 输出格式化日志
func Debugf(format string, v ...interface{}) {
	logger.Printf("%s [DEBUG] - "+format, append([]interface{}{logPrefix()}, v...)...)
}

// Error 输出错误日志
func Error(v ...interface{}) {
	logger.Println(append([]interface{}{logPrefix(), "[ERROR] - "}, v...)...)
}

// Errorf 输出格式化错误日志
func Errorf(format string, v ...interface{}) {
	logger.Printf("%s [ERROR] - "+format, append([]interface{}{logPrefix()}, v...)...)
}

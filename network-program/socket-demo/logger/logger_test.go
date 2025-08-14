package logger

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"
)

// 捕获 logger 输出的辅助函数
func captureOutput(f func()) string {
	var buf bytes.Buffer
	oldOutput := logger.Writer() // 保存原来的输出
	logger.SetOutput(&buf)       // 重定向到内存
	defer logger.SetOutput(oldOutput)

	f()
	return buf.String()
}

// 用正则检查日志前缀格式
var prefixRegex = regexp.MustCompile(`^$begin:math:display$\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}$end:math:display$ \S+:\d+ \S+$begin:math:text$$end:math:text$`)

func TestInfo(t *testing.T) {
	output := captureOutput(func() {
		Info("Hello", "World")
	})

	if !prefixRegex.MatchString(output) {
		t.Errorf("日志前缀格式不正确: %s", output)
	}
	if !strings.Contains(output, "Hello World") {
		t.Errorf("日志内容不正确: %s", output)
	}
}

func TestInfof(t *testing.T) {
	output := captureOutput(func() {
		Infof("Hello %s", "Go")
	})

	if !prefixRegex.MatchString(output) {
		t.Errorf("日志前缀格式不正确: %s", output)
	}
	if !strings.Contains(output, "Hello Go") {
		t.Errorf("日志内容不正确: %s", output)
	}
}

func TestError(t *testing.T) {
	output := captureOutput(func() {
		Error("Oops")
	})

	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("错误日志缺少 [ERROR] 标记: %s", output)
	}
	if !strings.Contains(output, "Oops") {
		t.Errorf("错误日志缺少内容: %s", output)
	}
}

func TestErrorf(t *testing.T) {
	output := captureOutput(func() {
		Errorf("Oops: %s", "Disk full")
	})

	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("错误日志缺少 [ERROR] 标记: %s", output)
	}
	if !strings.Contains(output, "Oops: Disk full") {
		t.Errorf("错误日志缺少内容: %s", output)
	}
}

// 额外测试：多 goroutine 安全性（并发输出）
func TestLoggerConcurrency(t *testing.T) {
	done := make(chan struct{})
	go func() {
		Info("goroutine log")
		close(done)
	}()
	<-done
}

func TestMain(m *testing.M) {
	// 保证测试时不污染真实 stdout
	logger.SetOutput(os.Stdout)
	os.Exit(m.Run())
}

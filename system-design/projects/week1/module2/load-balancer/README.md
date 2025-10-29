# 生产级负载均衡器开发指南

> 从零到生产级：循序渐进构建高性能、可扩展的负载均衡器

---

## 目录

- [项目概述](#项目概述)
- [架构设计](#架构设计)
- [开发路线图](#开发路线图)
- [迭代开发步骤](#迭代开发步骤)
  - [迭代 0：项目初始化](#迭代-0项目初始化)
  - [迭代 1：基础版本 - Round Robin](#迭代-1基础版本---round-robin)
  - [迭代 2：添加加权轮询](#迭代-2添加加权轮询)
  - [迭代 3：添加最少连接算法](#迭代-3添加最少连接算法)
  - [迭代 4：算法可扩展设计](#迭代-4算法可扩展设计)
  - [迭代 5：添加健康检查](#迭代-5添加健康检查)
  - [迭代 6：添加熔断器](#迭代-6添加熔断器)
  - [迭代 7：一致性哈希算法](#迭代-7一致性哈希算法)
  - [迭代 8：性能优化](#迭代-8性能优化)
  - [迭代 9：监控与可观测性](#迭代-9监控与可观测性)
  - [迭代 10：测试与压测](#迭代-10测试与压测)
- [最终项目结构](#最终项目结构)
- [运行与部署](#运行与部署)

---

## 项目概述

### 目标

构建一个生产级的 HTTP 负载均衡器，具备以下特性：

**核心功能**：
- ✅ 多种负载均衡算法（Round Robin、加权轮询、最少连接、一致性哈希）
- ✅ 健康检查（主动 + 被动）
- ✅ 熔断保护
- ✅ 配置热更新
- ✅ 监控与指标

**技术要求**：
- 语言：Go 1.23+
- 性能：支持 10K+ QPS，P99 延迟 < 10ms
- 可扩展：易于添加新的负载均衡算法
- 生产级：完善的错误处理、日志、监控

### 学习目标

通过这个项目，你将掌握：
1. 负载均衡算法的实现
2. Go 并发编程（goroutine、channel、sync）
3. 接口设计与抽象
4. 健康检查与故障检测
5. 性能优化技巧
6. 测试驱动开发

---

## 架构设计

### 核心架构图

```
                         ┌─────────────────────┐
                         │   Load Balancer     │
                         │                     │
    Client Request ────► │  ┌───────────────┐  │
                         │  │  HTTP Server  │  │
                         │  └───────┬───────┘  │
                         │          │          │
                         │  ┌───────▼───────┐  │
                         │  │   Algorithm   │  │ ◄── 可扩展
                         │  │    Selector   │  │
                         │  └───────┬───────┘  │
                         │          │          │
                         │  ┌───────▼───────┐  │
                         │  │ Health Check  │  │
                         │  │   Manager     │  │
                         │  └───────┬───────┘  │
                         │          │          │
                         │  ┌───────▼───────┐  │
                         │  │Circuit Breaker│  │
                         │  └───────┬───────┘  │
                         └──────────┼──────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
                    ▼               ▼               ▼
              ┌──────────┐    ┌──────────┐    ┌──────────┐
              │Backend 1 │    │Backend 2 │    │Backend 3 │
              └──────────┘    └──────────┘    └──────────┘
```

### 模块划分

| 模块 | 职责 | 位置 |
|------|------|------|
| **cmd/lb** | 程序入口 | `cmd/lb/main.go` |
| **internal/server** | HTTP 服务器 | `internal/server/` |
| **internal/proxy** | 反向代理 | `internal/proxy/` |
| **pkg/algorithm** | 负载均衡算法 | `pkg/algorithm/` |
| **pkg/healthcheck** | 健康检查 | `pkg/healthcheck/` |
| **pkg/circuitbreaker** | 熔断器 | `pkg/circuitbreaker/` |
| **pkg/backend** | 后端服务器模型 | `pkg/backend/` |
| **pkg/config** | 配置管理 | `pkg/config/` |
| **pkg/metrics** | 监控指标 | `pkg/metrics/` |

### 可扩展设计

**核心接口**：

```go
// 负载均衡算法接口
type Algorithm interface {
    Select(backends []*Backend) (*Backend, error)
    Name() string
}

// 后端服务器接口
type Backend interface {
    GetURL() string
    IsHealthy() bool
    GetWeight() int
    IncrementConnections()
    DecrementConnections()
}

// 健康检查接口
type HealthChecker interface {
    Check(backend *Backend) error
    Start(ctx context.Context)
    Stop()
}
```

---

## 开发路线图

### 迭代计划

| 迭代 | 目标 | 耗时 | 复杂度 |
|------|------|------|--------|
| 0 | 项目初始化 | 30min | ⭐ |
| 1 | 基础 Round Robin | 2h | ⭐⭐ |
| 2 | 加权轮询 | 1h | ⭐⭐ |
| 3 | 最少连接 | 1h | ⭐⭐ |
| 4 | 算法可扩展设计 | 2h | ⭐⭐⭐ |
| 5 | 健康检查 | 3h | ⭐⭐⭐⭐ |
| 6 | 熔断器 | 3h | ⭐⭐⭐⭐ |
| 7 | 一致性哈希 | 2h | ⭐⭐⭐ |
| 8 | 性能优化 | 3h | ⭐⭐⭐⭐ |
| 9 | 监控指标 | 2h | ⭐⭐⭐ |
| 10 | 测试与压测 | 3h | ⭐⭐⭐ |

**总计**：约 22 小时（可分 3-4 天完成）

### 迭代原则

1. **增量开发**：每个迭代都是在上一个迭代的基础上添加功能
2. **可运行**：每个迭代结束都有一个可运行的版本
3. **测试驱动**：每个功能都要有对应的测试
4. **文档同步**：代码和文档同步更新

---

## 迭代开发步骤

## 迭代 0：项目初始化

### 目标

- 创建项目结构
- 初始化 Go 模块
- 设置基础配置

### 步骤

#### 1. 创建项目结构

```bash
# 进入项目目录
cd /Users/yule/Desktop/opera/2_code/Interview-oriented-programming/system-design/projects/week1/module2

# 创建负载均衡器目录
mkdir -p load-balancer

cd load-balancer

# 创建项目结构
mkdir -p cmd/lb
mkdir -p internal/{server,proxy}
mkdir -p pkg/{algorithm,backend,config,healthcheck,circuitbreaker,metrics}
mkdir -p test/{mock,integration,benchmark}
mkdir -p configs

# 创建基础文件
touch cmd/lb/main.go
touch README.md
touch Makefile
touch .gitignore
```

#### 2. 初始化 Go 模块

```bash
go mod init github.com/yourusername/load-balancer
```

#### 3. 创建 .gitignore

```gitignore
# 编译产物
/bin/
*.exe
*.test
*.prof

# IDE
.idea/
.vscode/
*.swp

# 配置文件（敏感信息）
configs/local.yaml

# 日志
*.log

# 测试覆盖率
coverage.out
```

#### 4. 创建 Makefile

```makefile
.PHONY: build run test clean bench lint

# 变量
BINARY_NAME=lb
BUILD_DIR=./bin
MAIN_PATH=./cmd/lb

# 构建
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# 运行
run: build
	@$(BUILD_DIR)/$(BINARY_NAME)

# 测试
test:
	@echo "Running tests..."
	@go test -v -race ./...

# 基准测试
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# 覆盖率
coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out

# 代码检查
lint:
	@golangci-lint run

# 格式化
fmt:
	@go fmt ./...

# 安装依赖
deps:
	@go mod tidy
	@go mod download
```

#### 5. 创建 README.md

```markdown
# 生产级负载均衡器

一个用 Go 实现的高性能、可扩展的 HTTP 负载均衡器。

## 功能特性

- ✅ 多种负载均衡算法（Round Robin、加权轮询、最少连接、一致性哈希）
- ✅ 健康检查（主动 + 被动）
- ✅ 熔断保护
- ✅ 监控与指标
- ✅ 配置热更新

## 快速开始

\`\`\`bash
# 构建
make build

# 运行
make run

# 测试
make test
\`\`\`

## 开发进度

- [x] 迭代 0：项目初始化
- [ ] 迭代 1：基础 Round Robin
- [ ] 迭代 2：加权轮询
- [ ] ...
```

### 验证

```bash
# 检查目录结构
tree -L 3

# 初始化模块
go mod init github.com/yourusername/load-balancer
go mod tidy
```

### 检查点 ✅

- [ ] 项目结构创建完成
- [ ] Go 模块初始化成功
- [ ] Makefile 可以执行
- [ ] README 文档创建

---

## 迭代 1：基础版本 - Round Robin

### 目标

实现一个最简单的负载均衡器，支持基础的 Round Robin 算法。

### 功能需求

- HTTP 服务器监听请求
- 轮询方式选择后端服务器
- 反向代理请求到后端

### 步骤

#### 1. 定义后端服务器模型

**文件**：`pkg/backend/backend.go`

```go
package backend

import (
	"net/url"
	"sync"
	"sync/atomic"
)

// Backend 表示一个后端服务器
type Backend struct {
	URL     *url.URL
	Alive   bool
	mu      sync.RWMutex

	// 连接计数（用于最少连接算法）
	connections int64
}

// NewBackend 创建新的后端服务器
func NewBackend(urlStr string) (*Backend, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return &Backend{
		URL:   parsedURL,
		Alive: true,
	}, nil
}

// GetURL 返回后端 URL
func (b *Backend) GetURL() string {
	return b.URL.String()
}

// IsAlive 检查后端是否存活
func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Alive
}

// SetAlive 设置后端存活状态
func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Alive = alive
}

// GetConnections 获取当前连接数
func (b *Backend) GetConnections() int64 {
	return atomic.LoadInt64(&b.connections)
}

// IncrementConnections 增加连接数
func (b *Backend) IncrementConnections() {
	atomic.AddInt64(&b.connections, 1)
}

// DecrementConnections 减少连接数
func (b *Backend) DecrementConnections() {
	atomic.AddInt64(&b.connections, -1)
}
```

#### 2. 实现 Round Robin 算法

**文件**：`pkg/algorithm/roundrobin.go`

```go
package algorithm

import (
	"errors"
	"sync"

	"github.com/yourusername/load-balancer/pkg/backend"
)

// RoundRobin 轮询算法
type RoundRobin struct {
	backends []*backend.Backend
	current  int
	mu       sync.Mutex
}

// NewRoundRobin 创建轮询算法实例
func NewRoundRobin(backends []*backend.Backend) *RoundRobin {
	return &RoundRobin{
		backends: backends,
		current:  0,
	}
}

// Select 选择下一个后端服务器
func (rr *RoundRobin) Select() (*backend.Backend, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if len(rr.backends) == 0 {
		return nil, errors.New("no backends available")
	}

	// 跳过不存活的后端
	tried := 0
	for {
		backend := rr.backends[rr.current]
		rr.current = (rr.current + 1) % len(rr.backends)

		if backend.IsAlive() {
			return backend, nil
		}

		tried++
		if tried >= len(rr.backends) {
			return nil, errors.New("no healthy backends available")
		}
	}
}

// Name 返回算法名称
func (rr *RoundRobin) Name() string {
	return "RoundRobin"
}
```

#### 3. 实现反向代理

**文件**：`internal/proxy/proxy.go`

```go
package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/yourusername/load-balancer/pkg/backend"
)

// ReverseProxy 反向代理
type ReverseProxy struct {
	client *http.Client
}

// NewReverseProxy 创建反向代理
func NewReverseProxy() *ReverseProxy {
	return &ReverseProxy{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// ServeHTTP 处理反向代理请求
func (p *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request, backend *backend.Backend) {
	// 构造目标 URL
	targetURL := fmt.Sprintf("%s%s", backend.GetURL(), r.RequestURI)

	// 创建新请求
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		log.Printf("Error creating proxy request: %v", err)
		return
	}

	// 复制请求头
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// 添加 X-Forwarded-* 头
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
	proxyReq.Header.Set("X-Forwarded-Proto", r.URL.Scheme)
	proxyReq.Header.Set("X-Forwarded-Host", r.Host)

	// 发送请求
	resp, err := p.client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		log.Printf("Error proxying request to %s: %v", backend.GetURL(), err)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}
}
```

#### 4. 实现负载均衡服务器

**文件**：`internal/server/server.go`

```go
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/yourusername/load-balancer/internal/proxy"
	"github.com/yourusername/load-balancer/pkg/algorithm"
	"github.com/yourusername/load-balancer/pkg/backend"
)

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	port      int
	algorithm algorithm.Selector
	proxy     *proxy.ReverseProxy
	server    *http.Server
}

// Selector 算法选择器接口
type Selector interface {
	Select() (*backend.Backend, error)
	Name() string
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(port int, algo Selector) *LoadBalancer {
	lb := &LoadBalancer{
		port:      port,
		algorithm: algo,
		proxy:     proxy.NewReverseProxy(),
	}

	// 创建 HTTP 服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/", lb.handleRequest)
	mux.HandleFunc("/health", lb.handleHealth)

	lb.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return lb
}

// handleRequest 处理客户端请求
func (lb *LoadBalancer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// 选择后端服务器
	backend, err := lb.algorithm.Select()
	if err != nil {
		http.Error(w, "No backend available", http.StatusServiceUnavailable)
		log.Printf("Failed to select backend: %v", err)
		return
	}

	// 增加连接计数
	backend.IncrementConnections()
	defer backend.DecrementConnections()

	// 记录日志
	log.Printf("[%s] Forwarding request to %s: %s %s",
		lb.algorithm.Name(), backend.GetURL(), r.Method, r.URL.Path)

	// 代理请求
	lb.proxy.ServeHTTP(w, r, backend)
}

// handleHealth 健康检查端点
func (lb *LoadBalancer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Start 启动服务器
func (lb *LoadBalancer) Start() error {
	log.Printf("Load Balancer [%s] starting on port %d", lb.algorithm.Name(), lb.port)
	return lb.server.ListenAndServe()
}

// Shutdown 优雅关闭
func (lb *LoadBalancer) Shutdown(ctx context.Context) error {
	log.Println("Shutting down load balancer...")
	return lb.server.Shutdown(ctx)
}
```

#### 5. 创建主程序

**文件**：`cmd/lb/main.go`

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/load-balancer/internal/server"
	"github.com/yourusername/load-balancer/pkg/algorithm"
	"github.com/yourusername/load-balancer/pkg/backend"
)

func main() {
	// 创建后端服务器列表
	backends := []*backend.Backend{}
	backendURLs := []string{
		"http://localhost:8081",
		"http://localhost:8082",
		"http://localhost:8083",
	}

	for _, url := range backendURLs {
		b, err := backend.NewBackend(url)
		if err != nil {
			log.Fatalf("Failed to create backend %s: %v", url, err)
		}
		backends = append(backends, b)
		log.Printf("Added backend: %s", url)
	}

	// 创建 Round Robin 算法
	algo := algorithm.NewRoundRobin(backends)

	// 创建负载均衡器
	lb := server.NewLoadBalancer(8080, algo)

	// 启动服务器（在 goroutine 中）
	go func() {
		if err := lb.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := lb.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown gracefully: %v", err)
	}

	log.Println("Server stopped")
}
```

#### 6. 创建测试后端服务器

**文件**：`test/mock/backend.go`

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: backend <port>")
	}

	port := os.Args[1]

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf("Response from backend on port %s\nPath: %s\n", port, r.URL.Path)
		w.Write([]byte(response))
		log.Printf("Handled request: %s %s", r.Method, r.URL.Path)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Backend server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

### 测试

#### 1. 启动测试后端

```bash
# 终端 1
go run test/mock/backend.go 8081

# 终端 2
go run test/mock/backend.go 8082

# 终端 3
go run test/mock/backend.go 8083
```

#### 2. 启动负载均衡器

```bash
# 终端 4
make run
```

#### 3. 测试负载均衡

```bash
# 发送多个请求，观察轮询效果
for i in {1..9}; do
  curl http://localhost:8080/
  echo "---"
done
```

**预期输出**：
```
Response from backend on port 8081
---
Response from backend on port 8082
---
Response from backend on port 8083
---
Response from backend on port 8081
---
...
```

### 检查点 ✅

- [ ] Round Robin 算法正确实现
- [ ] 请求能够轮询到不同后端
- [ ] 日志输出正确
- [ ] 优雅关闭功能正常

---

## 迭代 2：添加加权轮询

### 目标

支持为后端服务器设置权重，按权重分配流量。

### 步骤

#### 1. 更新 Backend 模型

**文件**：`pkg/backend/backend.go`（添加权重字段）

```go
type Backend struct {
	URL     *url.URL
	Alive   bool
	Weight  int  // 新增：权重
	mu      sync.RWMutex

	connections int64
}

func NewBackend(urlStr string, weight int) (*Backend, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	if weight <= 0 {
		weight = 1  // 默认权重为 1
	}

	return &Backend{
		URL:    parsedURL,
		Alive:  true,
		Weight: weight,
	}, nil
}

// GetWeight 获取权重
func (b *Backend) GetWeight() int {
	return b.Weight
}
```

#### 2. 实现加权轮询算法（平滑加权轮询）

**文件**：`pkg/algorithm/weighted_roundrobin.go`

```go
package algorithm

import (
	"errors"
	"sync"

	"github.com/yourusername/load-balancer/pkg/backend"
)

// WeightedRoundRobin 平滑加权轮询算法
type WeightedRoundRobin struct {
	backends      []*backend.Backend
	currentWeight map[*backend.Backend]int  // 当前权重
	mu            sync.Mutex
}

// NewWeightedRoundRobin 创建加权轮询算法实例
func NewWeightedRoundRobin(backends []*backend.Backend) *WeightedRoundRobin {
	wrr := &WeightedRoundRobin{
		backends:      backends,
		currentWeight: make(map[*backend.Backend]int),
	}

	// 初始化当前权重
	for _, b := range backends {
		wrr.currentWeight[b] = 0
	}

	return wrr
}

// Select 选择下一个后端服务器（平滑加权轮询）
func (wrr *WeightedRoundRobin) Select() (*backend.Backend, error) {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	if len(wrr.backends) == 0 {
		return nil, errors.New("no backends available")
	}

	var selected *backend.Backend
	totalWeight := 0

	// 遍历所有后端
	for _, b := range wrr.backends {
		if !b.IsAlive() {
			continue
		}

		// 累加当前权重
		wrr.currentWeight[b] += b.GetWeight()
		totalWeight += b.GetWeight()

		// 选择当前权重最大的
		if selected == nil || wrr.currentWeight[b] > wrr.currentWeight[selected] {
			selected = b
		}
	}

	if selected == nil {
		return nil, errors.New("no healthy backends available")
	}

	// 减去总权重
	wrr.currentWeight[selected] -= totalWeight

	return selected, nil
}

// Name 返回算法名称
func (wrr *WeightedRoundRobin) Name() string {
	return "WeightedRoundRobin"
}
```

#### 3. 更新主程序支持权重

**文件**：`cmd/lb/main.go`

```go
func main() {
	// 创建后端服务器列表（带权重）
	backends := []*backend.Backend{}
	backendConfigs := []struct {
		URL    string
		Weight int
	}{
		{"http://localhost:8081", 5},  // 权重 5
		{"http://localhost:8082", 1},  // 权重 1
		{"http://localhost:8083", 1},  // 权重 1
	}

	for _, cfg := range backendConfigs {
		b, err := backend.NewBackend(cfg.URL, cfg.Weight)
		if err != nil {
			log.Fatalf("Failed to create backend %s: %v", cfg.URL, err)
		}
		backends = append(backends, b)
		log.Printf("Added backend: %s (weight: %d)", cfg.URL, cfg.Weight)
	}

	// 创建加权轮询算法
	algo := algorithm.NewWeightedRoundRobin(backends)

	// ... 其余代码不变
}
```

### 测试

```bash
# 发送 14 个请求（5+1+1 的两个周期）
for i in {1..14}; do
  curl http://localhost:8080/ 2>/dev/null | grep "port"
done
```

**预期输出**（符合 5:1:1 的权重比例）：
```
port 8081
port 8081
port 8082
port 8081
port 8083
port 8081
port 8081
（重复）
```

### 检查点 ✅

- [ ] 加权轮询算法正确实现
- [ ] 流量分配符合权重比例
- [ ] 平滑加权轮询效果良好（避免突发流量）

---

## 迭代 3：添加最少连接算法

### 目标

实现最少连接算法，优先选择连接数最少的后端。

### 步骤

#### 实现最少连接算法

**文件**：`pkg/algorithm/leastconnection.go`

```go
package algorithm

import (
	"errors"
	"sync"

	"github.com/yourusername/load-balancer/pkg/backend"
)

// LeastConnection 最少连接算法
type LeastConnection struct {
	backends []*backend.Backend
	mu       sync.RWMutex
}

// NewLeastConnection 创建最少连接算法实例
func NewLeastConnection(backends []*backend.Backend) *LeastConnection {
	return &LeastConnection{
		backends: backends,
	}
}

// Select 选择连接数最少的后端服务器
func (lc *LeastConnection) Select() (*backend.Backend, error) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	if len(lc.backends) == 0 {
		return nil, errors.New("no backends available")
	}

	var selected *backend.Backend
	minConnections := int64(-1)

	for _, b := range lc.backends {
		if !b.IsAlive() {
			continue
		}

		connections := b.GetConnections()
		if minConnections == -1 || connections < minConnections {
			selected = b
			minConnections = connections
		}
	}

	if selected == nil {
		return nil, errors.New("no healthy backends available")
	}

	return selected, nil
}

// Name 返回算法名称
func (lc *LeastConnection) Name() string {
	return "LeastConnection"
}
```

### 测试

创建测试文件：`pkg/algorithm/leastconnection_test.go`

```go
package algorithm

import (
	"testing"

	"github.com/yourusername/load-balancer/pkg/backend"
)

func TestLeastConnection(t *testing.T) {
	// 创建后端服务器
	backends := []*backend.Backend{
		mustCreateBackend("http://server1.com", 1),
		mustCreateBackend("http://server2.com", 1),
		mustCreateBackend("http://server3.com", 1),
	}

	lc := NewLeastConnection(backends)

	// 模拟连接
	backends[0].IncrementConnections()
	backends[0].IncrementConnections()  // server1: 2 connections
	backends[1].IncrementConnections()  // server2: 1 connection
	// server3: 0 connections

	// 应该选择 server3（连接数最少）
	selected, err := lc.Select()
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	if selected.GetURL() != "http://server3.com" {
		t.Errorf("Expected server3, got %s", selected.GetURL())
	}
}

func mustCreateBackend(url string, weight int) *backend.Backend {
	b, err := backend.NewBackend(url, weight)
	if err != nil {
		panic(err)
	}
	return b
}
```

运行测试：
```bash
go test ./pkg/algorithm/... -v
```

### 检查点 ✅

- [ ] 最少连接算法正确实现
- [ ] 测试通过
- [ ] 连接计数正确维护

---

## 迭代 4：算法可扩展设计

### 目标

设计一个灵活的架构，支持：
1. 轻松添加新的负载均衡算法
2. 运行时切换算法
3. 配置文件驱动

### 步骤

#### 1. 定义算法接口

**文件**：`pkg/algorithm/interface.go`

```go
package algorithm

import "github.com/yourusername/load-balancer/pkg/backend"

// Algorithm 负载均衡算法接口
type Algorithm interface {
	Select() (*backend.Backend, error)
	Name() string
}

// AlgorithmType 算法类型
type AlgorithmType string

const (
	AlgoRoundRobin         AlgorithmType = "round-robin"
	AlgoWeightedRoundRobin AlgorithmType = "weighted-round-robin"
	AlgoLeastConnection    AlgorithmType = "least-connection"
	AlgoConsistentHash     AlgorithmType = "consistent-hash"
)
```

#### 2. 实现算法工厂

**文件**：`pkg/algorithm/factory.go`

```go
package algorithm

import (
	"fmt"

	"github.com/yourusername/load-balancer/pkg/backend"
)

// Factory 算法工厂
type Factory struct{}

// NewFactory 创建算法工厂
func NewFactory() *Factory {
	return &Factory{}
}

// Create 创建指定类型的算法实例
func (f *Factory) Create(algoType AlgorithmType, backends []*backend.Backend) (Algorithm, error) {
	switch algoType {
	case AlgoRoundRobin:
		return NewRoundRobin(backends), nil

	case AlgoWeightedRoundRobin:
		return NewWeightedRoundRobin(backends), nil

	case AlgoLeastConnection:
		return NewLeastConnection(backends), nil

	case AlgoConsistentHash:
		// TODO: 迭代 7 实现
		return nil, fmt.Errorf("consistent hash not implemented yet")

	default:
		return nil, fmt.Errorf("unknown algorithm type: %s", algoType)
	}
}

// Available 返回所有可用的算法类型
func (f *Factory) Available() []AlgorithmType {
	return []AlgorithmType{
		AlgoRoundRobin,
		AlgoWeightedRoundRobin,
		AlgoLeastConnection,
	}
}
```

#### 3. 添加配置管理

**文件**：`pkg/config/config.go`

```go
package config

import (
	"encoding/json"
	"os"

	"github.com/yourusername/load-balancer/pkg/algorithm"
)

// Config 负载均衡器配置
type Config struct {
	Port      int                    `json:"port"`
	Algorithm algorithm.AlgorithmType `json:"algorithm"`
	Backends  []BackendConfig        `json:"backends"`
}

// BackendConfig 后端服务器配置
type BackendConfig struct {
	URL    string `json:"url"`
	Weight int    `json:"weight"`
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	// 设置默认值
	if config.Port == 0 {
		config.Port = 8080
	}

	if config.Algorithm == "" {
		config.Algorithm = algorithm.AlgoRoundRobin
	}

	return &config, nil
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
```

#### 4. 创建配置文件

**文件**：`configs/config.json`

```json
{
  "port": 8080,
  "algorithm": "weighted-round-robin",
  "backends": [
    {
      "url": "http://localhost:8081",
      "weight": 5
    },
    {
      "url": "http://localhost:8082",
      "weight": 1
    },
    {
      "url": "http://localhost:8083",
      "weight": 1
    }
  ]
}
```

#### 5. 更新主程序使用配置

**文件**：`cmd/lb/main.go`

```go
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/load-balancer/internal/server"
	"github.com/yourusername/load-balancer/pkg/algorithm"
	"github.com/yourusername/load-balancer/pkg/backend"
	"github.com/yourusername/load-balancer/pkg/config"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "configs/config.json", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Loaded config from %s", *configPath)
	log.Printf("Algorithm: %s", cfg.Algorithm)

	// 创建后端服务器列表
	backends := []*backend.Backend{}
	for _, bcfg := range cfg.Backends {
		b, err := backend.NewBackend(bcfg.URL, bcfg.Weight)
		if err != nil {
			log.Fatalf("Failed to create backend %s: %v", bcfg.URL, err)
		}
		backends = append(backends, b)
		log.Printf("Added backend: %s (weight: %d)", bcfg.URL, bcfg.Weight)
	}

	// 使用工厂创建算法
	factory := algorithm.NewFactory()
	algo, err := factory.Create(cfg.Algorithm, backends)
	if err != nil {
		log.Fatalf("Failed to create algorithm: %v", err)
	}

	log.Printf("Using algorithm: %s", algo.Name())

	// 创建负载均衡器
	lb := server.NewLoadBalancer(cfg.Port, algo)

	// 启动服务器
	go func() {
		if err := lb.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := lb.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown gracefully: %v", err)
	}

	log.Println("Server stopped")
}
```

### 测试

```bash
# 使用默认配置运行
make run

# 使用自定义配置运行
go run cmd/lb/main.go -config configs/custom.json
```

### 检查点 ✅

- [ ] 算法工厂正确实现
- [ ] 配置文件能够加载
- [ ] 可以通过配置切换算法
- [ ] 代码结构清晰，易于扩展

---

由于篇幅限制，我将后续迭代（5-10）的详细内容继续在下一部分。你想让我：

1. 继续完成剩余的迭代（5-10）？
2. 还是先让你完成前 4 个迭代，再继续？

现在你已经有：
- ✅ 完整的项目结构
- ✅ 三种负载均衡算法（Round Robin、加权轮询、最少连接）
- ✅ 可扩展的架构设计（接口 + 工厂模式）
- ✅ 配置文件驱动

选择哪个方向？
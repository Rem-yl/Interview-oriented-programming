# 阻塞 vs 非阻塞 I/O 性能实验

## 实验目的

对比三种并发模型在**相同工作负载**下的性能表现：
1. **Blocking I/O** (每请求一个 Goroutine)
2. **Worker Pool** (固定数量 Worker，原始版本)
3. **Optimized Worker Pool** (优化版 Worker Pool)

## 代码结构

```
nginx-block-exp/
├── block/
│   └── io_block.go          # 阻塞模型 (端口 8001)
├── no-block/
│   ├── no_block.go          # Worker Pool 原始版 (端口 8002, workers=100)
│   └── no_block_v2.go       # Worker Pool 优化版 (端口 8003, workers=50)
└── README.md                # 本文档
```

## 快速开始

### 1. 启动三个服务器

```bash
# 终端1 - Blocking I/O
cd block
go run io_block.go

# 终端2 - Worker Pool (原始)
cd no-block
go run no_block.go

# 终端3 - Worker Pool (优化)
cd no-block
go run no_block_v2.go
```

### 2. 运行压测

```bash
# 安装 wrk (macOS)
brew install wrk

# 测试 Blocking I/O (8001)
wrk -t4 -c100 -d30s --latency http://localhost:8001/test

# 测试 Worker Pool 原始版 (8002)
wrk -t4 -c100 -d30s --latency http://localhost:8002/test

# 测试 Worker Pool 优化版 (8003)
wrk -t4 -c100 -d30s --latency http://localhost:8003/test
```

### 3. 查看实时统计

```bash
# 在压测期间，另开终端查看统计
watch -n 1 'curl -s http://localhost:8001/stats | jq'
watch -n 1 'curl -s http://localhost:8002/stats | jq'
watch -n 1 'curl -s http://localhost:8003/stats | jq'
```

## 性能分析

### 为什么 Worker Pool 可能更慢？

#### 问题 1: Worker 数量不足

**场景**: 100 并发请求，但只有 10 个 Worker

```
Block 模型:
  100 请求 → 100 Goroutine → 同时处理
  延迟 = 50ms

Worker Pool (10 workers):
  100 请求 → 10 Worker → 排队处理
  第1批: 10个 (50ms)
  第2批: 10个 (50ms)
  ...
  第10批: 10个 (50ms)
  延迟 = 50ms × 10 = 500ms (最坏情况)
```

**修复**: 增加 Worker 数量到 50-100

#### 问题 2: Channel 开销

```go
// 每个请求的 Channel 操作:
task := Task{ResultChan: make(chan Result, 1)}  // 分配 Chan
taskQueue <- task                                // 写入 (锁竞争)
result := <-task.ResultChan                     // 读取 (锁竞争)
```

**开销**:
- Channel 创建: ~100ns
- Channel 发送/接收: ~50ns (有锁竞争时更高)
- 总额外开销: ~200ns per request

**优化**: 使用 `sync.Pool` 复用 Channel (见 no_block_v2.go)

#### 问题 3: 额外的上下文切换

```
Block:    HTTP → Handler Goroutine → 返回
          (1次Goroutine切换)

Worker:   HTTP → Handler Goroutine → Worker Goroutine → 返回
          (3次Goroutine切换)
```

**开销**: 每次切换 ~1-2μs (高并发时)

### Worker Pool 的真正优势场景

Worker Pool **不适合**当前场景，但适合：

#### ✅ 场景 1: 外部资源限制

```go
// 数据库连接池: 最多50个连接
var dbPool = 50

// Worker Pool 保护数据库
workerCount = 50  // 匹配数据库连接数
```

**优势**: 防止数据库过载 (OOM/连接耗尽)

#### ✅ 场景 2: CPU 密集型任务

```go
func worker(id int) {
    for task := range taskQueue {
        // 大量CPU计算 (图片处理、加密等)
        result := expensiveComputation(task)
        task.ResultChan <- result
    }
}

workerCount = runtime.NumCPU()  // 8核 = 8 Worker
```

**优势**: 避免过多 Goroutine 竞争 CPU，减少上下文切换

#### ✅ 场景 3: 限流/流控

```go
// 第三方 API 限制: 100 QPS
workerCount = 10
rateLimiter := time.NewTicker(10 * time.Millisecond)

func worker(id int) {
    for task := range taskQueue {
        <-rateLimiter.C  // 限流
        callExternalAPI(task)
    }
}
```

**优势**: 精确控制并发数，避免超出限制

### 预期性能对比

| 指标 | Block (8001) | Worker Pool 原始 (8002) | Worker Pool 优化 (8003) |
|------|--------------|-------------------------|-------------------------|
| **QPS** | 1,800 | 1,600 | 1,750 |
| **P50 延迟** | 52ms | 60ms | 54ms |
| **P99 延迟** | 65ms | 120ms | 70ms |
| **Goroutine 数** | ~105 | ~105 | ~55 |
| **内存** | 45MB | 48MB | 42MB |

**结论**:
- **Block 模型最快** (Go 的 Goroutine 已经足够轻量)
- **Worker Pool 原始版慢** (10 workers 成为瓶颈)
- **Worker Pool 优化版接近 Block** (100 workers + sync.Pool)

## 实验步骤

### 实验 1: Worker 数量对性能的影响

```bash
# 修改 no_block.go 中的 workerCount
workerCount = 10    # 压测记录结果
workerCount = 50    # 压测记录结果
workerCount = 100   # 压测记录结果
workerCount = 200   # 压测记录结果
```

**观察指标**:
- QPS 如何变化？
- 延迟分布 (P50/P99) 如何变化？
- Goroutine 数量变化？

### 实验 2: 队列长度影响

```bash
# 修改 taskQueue 缓冲大小
taskQueue = make(chan Task, 10)      # 小队列
taskQueue = make(chan Task, 1000)    # 中队列
taskQueue = make(chan Task, 10000)   # 大队列
```

**观察**:
- 队列满时的拒绝率
- 延迟抖动

### 实验 3: 不同并发级别

```bash
# 低并发 (10)
wrk -t2 -c10 -d30s --latency http://localhost:8001/test

# 中并发 (100)
wrk -t4 -c100 -d30s --latency http://localhost:8001/test

# 高并发 (1000)
wrk -t8 -c1000 -d30s --latency http://localhost:8001/test
```

**观察**: 哪种模型在高并发下更稳定？

## 关键代码对比

### Block 模型 (io_block.go)

```go
func testHandler(c *gin.Context) {
    doWork(50)  // 直接阻塞当前 Goroutine
    c.JSON(http.StatusOK, response)
}
```

**特点**:
- ✅ 简单直接
- ✅ 低延迟 (无额外开销)
- ❌ Goroutine 数量不可控
- ❌ 资源消耗随并发线性增长

### Worker Pool 原始版 (no_block.go)

```go
func testHandler(c *gin.Context) {
    task := Task{ResultChan: make(chan Result, 1)}
    taskQueue <- task
    result := <-task.ResultChan
    c.JSON(http.StatusOK, result)
}

func worker(id int) {
    for task := range taskQueue {
        doWork(50)
        task.ResultChan <- Result{}
    }
}
```

**特点**:
- ✅ Goroutine 数量可控 (workerCount)
- ❌ 额外 Channel 开销
- ❌ Worker 数量不足时成为瓶颈

### Worker Pool 优化版 (no_block_v2.go)

```go
var taskPool = sync.Pool{
    New: func() interface{} {
        return &Task{ResultChan: make(chan Result, 1)}
    },
}

func testHandler(c *gin.Context) {
    task := taskPool.Get().(*Task)  // 复用 Task
    taskQueue <- task
    result := <-task.ResultChan
    taskPool.Put(task)              // 归还
    c.JSON(http.StatusOK, result)
}
```

**特点**:
- ✅ 减少 Channel 分配开销
- ✅ 更高的 Worker 数量 (50)
- ✅ 性能接近 Block 模型
- ❌ 代码复杂度增加

## 监控和调试

### 查看 Goroutine 数量

```bash
# 使用 /stats 接口
curl http://localhost:8001/stats | jq '.goroutines'

# 或使用 pprof
go tool pprof http://localhost:8001/debug/pprof/goroutine
```

### 查看内存使用

```bash
# pprof heap
go tool pprof http://localhost:8001/debug/pprof/heap

# 查看实时分配
go tool pprof http://localhost:8001/debug/pprof/allocs
```

### CPU 分析

```bash
# 压测时采集 CPU profile
go tool pprof http://localhost:8001/debug/pprof/profile?seconds=30

# 分析结果
(pprof) top10
(pprof) list testHandler
```

## 学习要点

### 1. Go 的 Goroutine 已经很轻量

- 栈大小: 2KB 起步 (动态增长)
- 切换成本: 纳秒级
- 100 并发 → 100 Goroutine ≈ 200KB 内存

**结论**: 对于纯 I/O 操作，直接用 Goroutine 就好

### 2. Worker Pool 不是银弹

**适用场景**:
- 外部资源有限制 (数据库连接池)
- CPU 密集型任务 (限制并发数 = CPU核心数)
- 需要精确限流

**不适用场景**:
- 简单的 I/O 等待 (time.Sleep)
- Go 原生的异步操作 (网络I/O)

### 3. 性能优化的权衡

```
性能提升 = 收益 - 成本

收益: 降低延迟、提高吞吐
成本: 代码复杂度、维护成本、调试难度

Worker Pool 优化版:
  收益: +5% QPS
  成本: +50% 代码复杂度
  → ROI 可能不划算
```

### 4. NGINX vs Go 的差异

| 特性 | NGINX (C) | Go (Gin) |
|------|-----------|----------|
| **并发原语** | Event Loop | Goroutine |
| **Worker 数量** | 固定 (worker_processes) | 动态 (无限) |
| **内存模型** | 手动管理 | GC 自动管理 |
| **适用场景** | 静态文件、反向代理 | 业务逻辑、API 服务 |

**为什么 NGINX 需要 Worker Pool**:
- C 没有 Goroutine，线程成本高 (~8MB 栈)
- 手动事件循环，必须用状态机
- 静态文件服务，CPU 密集度低

**为什么 Go 不一定需要**:
- Goroutine 成本低，可以动态创建
- Runtime 自动调度，无需手动管理
- 业务逻辑复杂，阻塞式代码更清晰

## 结论

1. **对于简单的 Web 服务** (如当前实验):
   - **Block 模型最优** (代码简单 + 性能好)

2. **Worker Pool 的价值**:
   - 不在于提高性能，而在于**资源保护**和**限流**

3. **Go vs NGINX**:
   - NGINX 的 Worker Pool 是**必需品** (C 语言限制)
   - Go 的 Worker Pool 是**可选优化** (特定场景)

4. **性能优化原则**:
   - 先测量，后优化
   - 理解瓶颈在哪 (CPU? I/O? 锁竞争?)
   - 评估复杂度 vs 收益

## 下一步实验

1. **真实数据库场景**: 集成 PostgreSQL/MySQL，观察连接池限制
2. **CPU 密集型任务**: 实现图片处理/加密，对比 Worker Pool 效果
3. **混合场景**: 同时有 I/O 和 CPU 任务，如何分配 Worker
4. **动态扩缩容**: 根据负载自动调整 Worker 数量

## 参考资料

- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [NGINX Event Loop](https://www.nginx.com/blog/inside-nginx-how-we-designed-for-performance-scale/)
- [Effective Go - Concurrency](https://go.dev/doc/effective_go#concurrency)

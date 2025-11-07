# Tester 实现思路梳理

## 一、现状分析

### 1.1 已有的基础设施

```
✅ 已完成的部分：
├── HttpClient (interface + implementation)
│   └── 负责底层 HTTP 请求（带超时、连接池配置）
├── LoadBalanceClient
│   └── 负责从负载均衡器获取后端地址
├── BackEndClient
│   └── 负责请求后端服务器
└── Tester (interface only)
    └── 只有接口定义，没有实现

❌ 缺失的部分：
├── RequestResult 类型（记录单次请求结果）
├── SequentialTester 实现（顺序测试）
├── ConcurrentTester 实现（并发测试）
└── 修改 main.go 使用 Tester
```

### 1.2 当前 main.go 的问题

```go
// cmd/client/main.go 当前逻辑
backend, err := loadBalanceClient.GetBackend()  // 只请求1次
msg, err := backendClient.Request(backend.GetURL())
fmt.Println("访问服务成功: ", msg)

问题：
❌ 只能发送 1 次请求
❌ 无法验证负载均衡算法是否正确（需要多次请求观察分布）
❌ 没有统计功能（成功率、延迟、分布）
❌ 无法并发测试
```

## 二、实现目标

实现两种测试器，验证负载均衡器的：
1. **正确性**（SequentialTester）：算法分布是否符合预期（如 Round Robin 应该均匀）
2. **并发安全性**（ConcurrentTester）：高并发下是否线程安全
3. **性能**（ConcurrentTester）：QPS、延迟等指标

## 三、核心数据结构设计

### 3.1 RequestResult - 记录单次请求的结果

```go
// 文件位置：pkg/tester/types.go

type RequestResult struct {
    // 基本信息
    RequestID int       // 请求编号（第几次请求）
    Timestamp time.Time // 请求时间戳

    // 请求结果
    Success   bool          // 是否成功
    Backend   string        // 请求到的后端（名称或URL）
    Latency   time.Duration // 请求延迟
    Error     error         // 错误信息（如果失败）

    // 可选：并发测试时使用
    WorkerID  int  // 哪个 worker 发送的（并发测试用）
}
```

**设计要点**：
- `RequestID`：用于排序和追踪
- `Backend`：统计分布的关键字段（如 "server-8180", "server-8181"）
- `Latency`：性能分析的关键指标
- `Success + Error`：区分成功和失败，单次失败不中断整个测试

### 3.2 Tester Interface（已存在）

```go
// pkg/tester/interface.go（已有）

type Tester interface {
    Run(ctx context.Context) ([]RequestResult, error)
}
```

**为什么需要 context？**
- ✅ 超时控制：整个测试的总超时
- ✅ 取消信号：用户 Ctrl+C 可以优雅退出
- ✅ 传递元数据：如 trace ID、request ID

**为什么返回 []RequestResult？**
- ✅ 后续统计分析需要完整数据（成功率、分布、延迟）
- ✅ 可以生成报告（表格、JSON、图表）

## 四、实现步骤

### Step 1: 创建 RequestResult 类型 ✅

```bash
# 文件：pkg/tester/types.go

创建内容：
1. RequestResult 结构体
2. 辅助方法（可选）：
   - IsSuccess() bool
   - GetBackend() string
   - String() string（用于打印）
```

### Step 2: 实现 SequentialTester（顺序测试器）

```bash
# 文件：pkg/tester/sequential.go
```

#### 2.1 结构体设计

```go
type SequentialTester struct {
    lbClient      *clients.LoadBalanceClient  // 负载均衡器客户端
    backendClient *clients.BackEndClient      // 后端客户端
    count         int                          // 总请求次数
}

构造函数：
func NewSequentialTester(
    lbClient *clients.LoadBalanceClient,
    backendClient *clients.BackEndClient,
    count int,
) *SequentialTester {
    return &SequentialTester{
        lbClient:      lbClient,
        backendClient: backendClient,
        count:         count,
    }
}
```

#### 2.2 核心逻辑：Run 方法

```go
func (t *SequentialTester) Run(ctx context.Context) ([]RequestResult, error) {
    // 1. 预分配结果切片（性能优化：避免多次扩容）
    results := make([]RequestResult, 0, t.count)

    // 2. 顺序发送 N 次请求
    for i := 0; i < t.count; i++ {
        // 2.1 检查 context 是否取消（支持优雅退出）
        select {
        case <-ctx.Done():
            return results, ctx.Err()  // 用户取消或超时
        default:
        }

        // 2.2 执行单次请求
        result := t.doSingleRequest(i)

        // 2.3 收集结果（即使失败也继续）
        results = append(results, result)
    }

    return results, nil
}
```

**关键设计点**：
| 设计点 | 代码 | 原因 |
|--------|------|------|
| 预分配切片 | `make([]RequestResult, 0, count)` | 避免多次扩容，提高性能 |
| context 检查 | `select { case <-ctx.Done() }` | 支持用户 Ctrl+C 或超时退出 |
| 错误不中断 | 单次失败继续循环 | 收集完整数据，统计成功率 |

#### 2.3 单次请求逻辑

```go
func (t *SequentialTester) doSingleRequest(requestID int) RequestResult {
    result := RequestResult{
        RequestID: requestID,
        Timestamp: time.Now(),
    }

    start := time.Now()

    // Step 1: 获取后端地址
    backend, err := t.lbClient.GetBackend()
    if err != nil {
        result.Success = false
        result.Error = fmt.Errorf("获取后端失败: %w", err)
        result.Latency = time.Since(start)
        return result
    }

    // Step 2: 请求后端服务器
    _, err = t.backendClient.Request(backend.GetURL())
    if err != nil {
        result.Success = false
        result.Error = fmt.Errorf("请求后端失败: %w", err)
        result.Backend = backend.GetURL()  // 记录哪个后端
        result.Latency = time.Since(start)
        return result
    }

    // Step 3: 成功
    result.Success = true
    result.Backend = backend.GetURL()  // 关键：记录访问了哪个后端
    result.Latency = time.Since(start)

    return result
}
```

**测试场景**：
```go
// 测试用例 1：Round Robin 算法验证
// 配置：3 个后端，发送 9 次请求
// 预期分布：每个后端 3 次（均匀分布）

tester := NewSequentialTester(lbClient, backendClient, 9)
results, _ := tester.Run(context.Background())

统计分布：
{
    "server-8180": 3,
    "server-8181": 3,
    "server-8182": 3
}
```

### Step 3: 实现 ConcurrentTester（并发测试器）

```bash
# 文件：pkg/tester/concurrent.go
```

#### 3.1 结构体设计

```go
type ConcurrentTester struct {
    lbClient      *clients.LoadBalanceClient
    backendClient *clients.BackEndClient
    totalCount    int  // 总请求数
    concurrent    int  // 并发数（goroutine 数量）
}

构造函数：
func NewConcurrentTester(
    lbClient *clients.LoadBalanceClient,
    backendClient *clients.BackEndClient,
    totalCount int,
    concurrent int,
) *ConcurrentTester {
    return &ConcurrentTester{
        lbClient:      lbClient,
        backendClient: backendClient,
        totalCount:    totalCount,
        concurrent:    concurrent,
    }
}
```

#### 3.2 核心逻辑：并发模型

```
并发模型（Worker Pool）：

假设：
- 总请求数：100
- 并发数：10

执行方式：
┌─────────┐ ┌─────────┐     ┌─────────┐
│Worker 0 │ │Worker 1 │ ... │Worker 9 │
│ 10 req  │ │ 10 req  │     │ 10 req  │
└────┬────┘ └────┬────┘     └────┬────┘
     │           │                │
     └───────────┼────────────────┘
                 ↓
         Result Channel (buffered)
                 ↓
         Main Goroutine Collects
```

#### 3.3 Run 方法实现

```go
func (t *ConcurrentTester) Run(ctx context.Context) ([]RequestResult, error) {
    // 1. 计算每个 worker 的请求数
    requestsPerWorker := t.totalCount / t.concurrent
    remainder := t.totalCount % t.concurrent

    // 例如：100 / 10 = 10（每个），余数 0
    // 例如：105 / 10 = 10（每个），余数 5（前5个worker多发1次）

    // 2. 创建结果 channel（buffered，避免 worker 阻塞）
    resultCh := make(chan RequestResult, t.totalCount)

    // 3. 创建 WaitGroup（等待所有 worker 完成）
    var wg sync.WaitGroup

    // 4. 启动 workers
    for i := 0; i < t.concurrent; i++ {
        count := requestsPerWorker
        if i < remainder {
            count++  // 余数分配给前几个 worker
        }

        wg.Add(1)
        go func(workerID, count int) {
            defer wg.Done()
            t.runWorker(ctx, workerID, count, resultCh)
        }(i, count)
    }

    // 5. 等待所有 worker 完成，然后关闭 channel
    go func() {
        wg.Wait()
        close(resultCh)  // 关闭后，range 循环会退出
    }()

    // 6. 收集结果
    results := make([]RequestResult, 0, t.totalCount)
    for result := range resultCh {
        results = append(results, result)
    }

    return results, nil
}
```

**关键设计点**：
| 设计点 | 代码 | 原因 |
|--------|------|------|
| Buffered Channel | `make(chan RequestResult, totalCount)` | 避免 worker 阻塞等待接收 |
| WaitGroup | `wg.Wait()` | 等待所有 worker 完成 |
| 关闭 Channel | `close(resultCh)` | 通知收集 goroutine 结束 |
| 余数分配 | 前几个 worker 多分配 | 确保总数正确 |

#### 3.4 Worker 逻辑

```go
func (t *ConcurrentTester) runWorker(
    ctx context.Context,
    workerID int,
    count int,
    resultCh chan<- RequestResult,
) {
    for i := 0; i < count; i++ {
        // 检查 context 是否取消
        select {
        case <-ctx.Done():
            return  // 优雅退出
        default:
        }

        // 执行单次请求
        result := t.doSingleRequest(workerID, i)

        // 发送结果到 channel
        resultCh <- result
    }
}

func (t *ConcurrentTester) doSingleRequest(workerID, requestSeq int) RequestResult {
    // 逻辑与 SequentialTester 类似，但多记录 WorkerID
    result := RequestResult{
        WorkerID:  workerID,
        RequestID: workerID*1000 + requestSeq,  // 全局唯一 ID
        Timestamp: time.Now(),
    }

    // ... 其余逻辑同 SequentialTester.doSingleRequest

    return result
}
```

### Step 4: 修改 main.go 使用 Tester

```go
// cmd/client/main.go

func main() {
    // ... 参数解析

    // 1. 创建客户端（已有）
    httpClient := clients.NewDefaultHttpClient(5 * time.Second)
    loadBalanceClient := clients.NewLoadBalanceClient(httpClient, addr)
    backendClient := clients.NewBackEndClient(httpClient)

    // 2. 创建测试器
    count := 100  // 发送100次请求

    // 选择测试模式
    mode := "sequential"  // 或 "concurrent"

    var tester tester.Tester
    if mode == "sequential" {
        tester = tester.NewSequentialTester(loadBalanceClient, backendClient, count)
    } else {
        concurrent := 10  // 10个并发
        tester = tester.NewConcurrentTester(loadBalanceClient, backendClient, count, concurrent)
    }

    // 3. 运行测试
    ctx := context.Background()
    results, err := tester.Run(ctx)
    if err != nil {
        fmt.Printf("测试失败: %v\n", err)
        os.Exit(1)
    }

    // 4. 简单输出结果（后续迭代会改进）
    fmt.Printf("总请求数: %d\n", len(results))

    // 统计成功率
    successCount := 0
    for _, r := range results {
        if r.Success {
            successCount++
        }
    }
    fmt.Printf("成功率: %.2f%%\n", float64(successCount)/float64(len(results))*100)

    // 统计分布
    distribution := make(map[string]int)
    for _, r := range results {
        if r.Success {
            distribution[r.Backend]++
        }
    }
    fmt.Println("后端分布:")
    for backend, count := range distribution {
        fmt.Printf("  %s: %d 次\n", backend, count)
    }
}
```

## 五、实现文件清单

### 需要创建的文件

```
pkg/tester/
├── interface.go          ✅ 已存在
├── types.go              ❌ 需要创建（RequestResult）
├── sequential.go         ❌ 需要创建（SequentialTester）
└── concurrent.go         ❌ 需要创建（ConcurrentTester）
```

### 需要修改的文件

```
cmd/client/main.go        ❌ 需要修改（使用 Tester）
```

## 六、测试策略

### 6.1 SequentialTester 测试

```bash
# 测试用例 1：Round Robin 均匀分布
./client --port 8187 --mode sequential --count 9

预期输出：
总请求数: 9
成功率: 100.00%
后端分布:
  http://localhost:8180: 3 次
  http://localhost:8181: 3 次
  http://localhost:8182: 3 次
```

### 6.2 ConcurrentTester 测试

```bash
# 测试用例 2：并发测试
./client --port 8187 --mode concurrent --count 100 --concurrent 10

预期输出：
总请求数: 100
成功率: 100.00%
后端分布:
  http://localhost:8180: 33 次
  http://localhost:8181: 33 次
  http://localhost:8182: 34 次

验证：分布大致均匀（±5% 波动正常）
```

### 6.3 负载均衡器未启动

```bash
# 测试用例 3：错误处理
# 停止负载均衡器
./client --port 8187 --mode sequential --count 10

预期输出：
总请求数: 10
成功率: 0.00%
错误: connection refused
```

## 七、常见问题和设计权衡

### Q1: 为什么单次请求失败不中断整个测试？

**原因**：
- ✅ 收集完整数据（需要计算成功率）
- ✅ 测试稳定性（某个后端挂了，其他后端应该继续服务）
- ✅ 真实场景模拟（生产环境部分请求失败很常见）

### Q2: 为什么并发测试要用 Channel 而不是 Mutex？

**对比**：

```go
// 方案 1：Mutex（不推荐）
var results []RequestResult
var mu sync.Mutex

mu.Lock()
results = append(results, result)
mu.Unlock()

问题：锁竞争严重，性能差

// 方案 2：Channel（推荐）
resultCh <- result

优势：
- 无锁设计
- 符合 Go 的并发哲学："不要通过共享内存来通信，而要通过通信来共享内存"
- 性能更好
```

### Q3: 如何选择并发数？

**建议**：
- 顺序测试：验证算法正确性（Round Robin 分布）
- 10-50 并发：模拟正常流量
- 100-500 并发：压力测试
- 1000+ 并发：极端情况测试（需要调整系统参数）

### Q4: Buffered Channel 的大小如何选择？

```go
// 选项 1：等于总请求数（推荐）
resultCh := make(chan RequestResult, totalCount)

// 选项 2：固定大小（如 100）
resultCh := make(chan RequestResult, 100)

// 选项 3：Unbuffered
resultCh := make(chan RequestResult)
```

**推荐选项 1**：
- ✅ Worker 永远不会阻塞
- ✅ 逻辑简单
- ❌ 内存占用：100 次请求约 10KB（可接受）

## 八、实现顺序建议

```
Day 1: 基础实现
├── 1. 创建 types.go（RequestResult）         30 min
├── 2. 实现 SequentialTester                  1 hour
└── 3. 简单修改 main.go 测试                  30 min

Day 2: 并发实现
├── 4. 实现 ConcurrentTester                  1.5 hours
└── 5. 测试并发正确性                         1 hour

Day 3: 改进和测试
├── 6. 添加命令行参数（--mode, --count等）   30 min
├── 7. 完善错误处理                           30 min
└── 8. 编写单元测试                           1 hour
```

## 九、验收标准

### 必须完成：
- [ ] RequestResult 类型定义
- [ ] SequentialTester 实现并通过测试
- [ ] ConcurrentTester 实现并通过测试
- [ ] main.go 能够使用两种测试器
- [ ] 能够统计成功率和后端分布

### 验证方式：
```bash
# 1. 启动后端服务器（3个）
cd scripts && ./start_server.sh

# 2. 启动负载均衡器
./bin/lb --port 8187

# 3. 运行顺序测试
./bin/client --mode sequential --count 9
# 验证：3个后端各3次

# 4. 运行并发测试
./bin/client --mode concurrent --count 100 --concurrent 10
# 验证：分布大致均匀

# 5. 停止负载均衡器，测试错误处理
./bin/client --mode sequential --count 5
# 验证：成功率 0%，有错误信息
```

## 十、下一步（迭代 4-5）

完成 Tester 后，后续迭代将实现：
1. **Statistics**（统计分析）：成功率、平均延迟、P99延迟、标准差
2. **Reporter**（报告输出）：表格、JSON、图表
3. **高级功能**：重试、超时、熔断器

---

**总结**：

核心思路就是：
1. **RequestResult**：记录单次请求的所有信息
2. **SequentialTester**：for 循环 + context 检查
3. **ConcurrentTester**：Worker Pool + Channel 收集结果
4. **main.go**：创建 Tester → 运行 → 统计输出

开始实现时，建议**先完成 SequentialTester**，因为：
- 逻辑简单，易于调试
- 可以验证负载均衡算法正确性
- ConcurrentTester 的单次请求逻辑可以复用

有任何实现中的问题随时问我！

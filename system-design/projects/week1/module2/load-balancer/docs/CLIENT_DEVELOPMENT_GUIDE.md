# 负载均衡器客户端开发指南

## 文档说明

**本指南的特点**：

- ❌ **不提供**完整代码实现
- ✅ **提供**深度原理解析
- ✅ **提供**设计思路和权衡分析
- ✅ **提供**测试策略和验证方法
- ✅ **提供**企业级最佳实践
- ✅ **引导**你自己思考和实现

**学习方法**：

1. 先理解每个迭代的**问题场景**
2. 思考**为什么**需要这样设计
3. 分析**权衡取舍**
4. 自己**编写实现**
5. 通过**测试验证**
6. 思考**思考题**加深理解

---

## 目录

- [当前代码问题分析](#当前代码问题分析)
- [设计哲学](#设计哲学)
- [迭代 0：技术债清理](#迭代-0技术债清理)
- [迭代 1：客户端抽象设计](#迭代-1客户端抽象设计)
- [迭代 2：配置管理](#迭代-2配置管理)
- [迭代 3：测试器设计](#迭代-3测试器设计)
- [迭代 4：统计分析](#迭代-4统计分析)
- [迭代 5：报告输出](#迭代-5报告输出)
- [迭代 6：并发测试](#迭代-6并发测试)
- [迭代 7：性能基准测试](#迭代-7性能基准测试)

---

## 当前代码问题分析

### 你的实现（简化版）

```go
func main() {
    // 1. 请求负载均衡器获取后端
    resp, err := http.Get("http://127.0.0.1:8187/balancer")
    if err != nil {
        panic(err)  // ❌ 问题 1
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)  // ❌ 问题 2
    if err != nil {
        panic(err)
    }

    var result Response
    if err := json.Unmarshal(body, &result); err != nil {
        panic(err)
    }

    // 2. 请求后端服务器
    resp, err = http.Get(result.Data.URL)  // ❌ 问题 3
    if err != nil {
        panic(err)
    }

    body, err = ioutil.ReadAll(resp.Body)  // ❌ 问题 4
    // ...
}
```

### 问题清单

| 问题             | 描述                          | 影响                           | 改进方向             |
| ---------------- | ----------------------------- | ------------------------------ | -------------------- |
| **问题 1** | 错误处理使用 `panic`        | 程序崩溃，不够健壮             | 返回 error，优雅处理 |
| **问题 2** | 使用废弃的 `ioutil.ReadAll` | Go 1.16+ 应该用 `io.ReadAll` | 更新 API             |
| **问题 3** | 代码重复                      | 两次 HTTP 请求逻辑几乎相同     | 提取公共函数         |
| **问题 4** | 硬编码 URL                    | 不够灵活                       | 配置化               |
| **问题 5** | 没有超时控制                  | 可能无限等待                   | 设置 timeout         |
| **问题 6** | 只能发送 1 次请求             | 无法测试负载均衡效果           | 支持多次请求         |
| **问题 7** | 没有统计功能                  | 无法验证算法正确性             | 添加统计分析         |

### 思考题

**Q1**：为什么 `panic` 不适合用于错误处理？什么时候可以用 `panic`？

<details>
<summary>提示</summary>

- `panic` 适用于：程序无法继续运行的致命错误（如配置文件必需但缺失）
- `error` 适用于：可以恢复的错误（如网络请求失败可以重试）
- 客户端请求失败应该属于可恢复错误

</details>

**Q2**：为什么要避免代码重复？DRY 原则（Don't Repeat Yourself）的本质是什么？

<details>
<summary>提示</summary>

代码重复的问题：

1. 修改困难：改一处要改多处
2. 容易出错：可能改漏了某处
3. 维护成本高

DRY 的本质：**一个知识点在系统中只有唯一的、明确的、权威的表示**

</details>

**Q3**：当前代码的功能是什么？测试负载均衡器需要什么功能？

<details>
<summary>提示</summary>

当前功能：发送 1 次请求，打印结果

测试需求：

- 发送 N 次请求
- 统计每个后端的访问次数
- 验证分布是否符合预期（Round Robin 应该均匀分布）
- 并发测试（模拟多个客户端）
- 性能测试（QPS、延迟）

</details>

---

## 设计哲学

### 核心原则

```
┌────────────────────────────────────────────┐
│  企业级客户端设计原则                       │
├────────────────────────────────────────────┤
│  1. 单一职责 (SRP)                         │
│     - 每个模块只做一件事                   │
│     - Client 负责请求                      │
│     - Tester 负责测试逻辑                  │
│     - Statistics 负责统计                  │
│     - Reporter 负责输出                    │
│                                            │
│  2. 开闭原则 (OCP)                         │
│     - 对扩展开放：可以添加新的测试模式      │
│     - 对修改关闭：不修改已有代码           │
│                                            │
│  3. 依赖倒置 (DIP)                         │
│     - 依赖接口而非实现                     │
│     - 便于 Mock 测试                       │
│                                            │
│  4. 组合优于继承                           │
│     - 通过组合不同的组件实现功能           │
└────────────────────────────────────────────┘
```

### 架构设计

```
                     ┌─────────────┐
                     │   main.go   │
                     │  (入口点)    │
                     └──────┬──────┘
                            │
              ┌─────────────┼─────────────┐
              │             │             │
              ↓             ↓             ↓
      ┌──────────┐  ┌──────────┐  ┌──────────┐
      │  Config  │  │  Client  │  │  Tester  │
      │  (配置)  │  │ (HTTP客户端)│ │(测试逻辑)│
      └──────────┘  └──────────┘  └─────┬────┘
                                        │
                           ┌────────────┼────────────┐
                           │                         │
                           ↓                         ↓
                   ┌──────────────┐         ┌──────────────┐
                   │  Statistics  │         │   Reporter   │
                   │   (统计分析)  │         │  (报告输出)   │
                   └──────────────┘         └──────────────┘
```

### 职责划分

| 模块                 | 职责                          | 不负责          |
| -------------------- | ----------------------------- | --------------- |
| **Config**     | 管理配置（URL、次数、超时等） | ❌ 不发送请求   |
| **Client**     | 发送 HTTP 请求、解析响应      | ❌ 不做统计分析 |
| **Tester**     | 控制测试流程（顺序、并发）    | ❌ 不格式化输出 |
| **Statistics** | 计算统计数据（成功率、延迟）  | ❌ 不发送请求   |
| **Reporter**   | 格式化输出（表格、JSON）      | ❌ 不计算统计   |

---

## 迭代 0：技术债清理

### 目标

清理现有代码的技术债务，为后续重构打好基础。

### 问题 1：废弃 API

**现状**：

```go
body, err := ioutil.ReadAll(resp.Body)  // Go 1.16+ 已废弃
```

**为什么废弃**？

- Go 1.16 开始，`ioutil` 包的功能被分散到其他包
- `ioutil.ReadAll` → `io.ReadAll`
- `ioutil.WriteFile` → `os.WriteFile`

**改进方向**：

```go
import "io"  // 而不是 "io/ioutil"

body, err := io.ReadAll(resp.Body)
```

**思考题**：

1. 为什么 Go 团队要废弃 `ioutil` 包？
2. 如何保持代码与 Go 版本的兼容性？

---

### 问题 2：错误处理策略

**场景分析**：

```
用户场景 1：网络抖动导致请求失败
期望行为：打印错误信息，程序继续运行

用户场景 2：负载均衡器地址配置错误
期望行为：提示用户检查配置，优雅退出

当前实现：panic → 程序崩溃，用户体验差
```

**错误处理分类**：

| 错误类型               | 示例               | 处理策略                       |
| ---------------------- | ------------------ | ------------------------------ |
| **可恢复错误**   | 网络超时、连接失败 | 返回 error，调用者决定是否重试 |
| **用户错误**     | 配置错误、参数错误 | 打印友好提示，退出码非 0       |
| **不可恢复错误** | 内存溢出、栈溢出   | panic（极少使用）              |

**设计思路**：

```go
核心原则：让错误可见，让程序健壮

1. 函数返回 error
   func doRequest() error {
       // ...
       if err != nil {
           return fmt.Errorf("请求失败: %w", err)
       }
       return nil
   }

2. main 函数处理错误
   func main() {
       if err := run(); err != nil {
           fmt.Fprintf(os.Stderr, "错误: %v\n", err)
           os.Exit(1)
       }
   }

3. 提供上下文信息
   错误消息应该包含：
   - 发生了什么
   - 在哪个步骤
   - 如何解决（可选）
```

**实现提示**：

```go
错误包装模式（Go 1.13+）：

// 包装错误，保留原始错误信息
if err != nil {
    return fmt.Errorf("获取负载均衡器响应失败: %w", err)
}

// 使用时可以检查底层错误类型
if errors.Is(err, context.DeadlineExceeded) {
    // 处理超时错误
}
```

**测试策略**：

如何测试错误处理？

```
测试用例 1：模拟网络错误
- 启动客户端
- 停止负载均衡器
- 发送请求
- 验证：返回友好的错误信息，程序不崩溃

测试用例 2：模拟超时
- 设置超时 1ms
- 发送请求
- 验证：返回超时错误

测试用例 3：模拟解析错误
- Mock 一个返回非 JSON 的服务器
- 发送请求
- 验证：返回解析错误
```

**思考题**：

**Q1**：什么时候应该用 `panic`？

<details>
<summary>提示</summary>

panic 的适用场景（非常少）：

1. 程序初始化失败（如必需的配置文件缺失）
2. 检测到不可能发生的错误（编程错误）
3. 第三方库要求（如 `must` 系列函数）

原则：**panic 只用于 bug，不用于预期的错误**

</details>

**Q2**：如何设计友好的错误消息？

<details>
<summary>提示</summary>

好的错误消息应该包含：

1. **What**：发生了什么错误
2. **Where**：在哪个步骤
3. **Why**：可能的原因（如果能推断）
4. **How**：如何解决（可选）

示例：

```
❌ 不好：error
✅ 好：请求负载均衡器失败: Get "http://localhost:8187/balancer": dial tcp 127.0.0.1:8187: connect: connection refused

提示：请检查负载均衡器是否正在运行
```

</details>

---

### 问题 3：代码重复

**重复代码分析**：

```go
// 第一次请求（负载均衡器）
resp, err := http.Get(balancerURL)
if err != nil {
    panic(err)
}
defer resp.Body.Close()
body, err := ioutil.ReadAll(resp.Body)
if err != nil {
    panic(err)
}
var result Response
if err := json.Unmarshal(body, &result); err != nil {
    panic(err)
}

// 第二次请求（后端服务器）
resp, err = http.Get(backendURL)
if err != nil {
    panic(err)
}
defer resp.Body.Close()
body, err = ioutil.ReadAll(resp.Body)
if err != nil {
    panic(err)
}
var serverResult HelloServerResponse
if err := json.Unmarshal(body, &serverResult); err != nil {
    panic(err)
}
```

**重复模式识别**：

```
共同步骤：
1. 发送 HTTP GET 请求
2. 读取响应体
3. 解析 JSON
4. 错误处理

不同点：
- URL 不同
- 解析的结构体类型不同
```

**提取公共逻辑的思路**：

```go
设计目标：提取通用的 HTTP 请求函数

挑战：如何处理不同的响应类型？

方案 1：使用泛型（Go 1.18+）
func request[T any](url string) (*T, error)

方案 2：传入结果指针
func request(url string, result interface{}) error

方案 3：返回原始字节，由调用者解析
func request(url string) ([]byte, error)
```

**设计权衡**：

| 方案                  | 优点           | 缺点               | 适用场景          |
| --------------------- | -------------- | ------------------ | ----------------- |
| **泛型**        | 类型安全、简洁 | 需要 Go 1.18+      | 现代项目          |
| **interface{}** | 灵活、兼容性好 | 类型不安全         | 需要兼容老版本 Go |
| **返回字节**    | 简单、职责清晰 | 调用者需要处理解析 | 通用 HTTP 客户端  |

**实现提示**：

```go
方向 1：通用请求函数

考虑因素：
1. 超时控制：应该可以自定义超时
2. 错误分类：网络错误 vs HTTP 错误 vs 解析错误
3. 日志记录：可选的请求日志

函数签名示例：
func doRequest(url string, timeout time.Duration, result interface{}) error

内部步骤：
1. 创建带超时的 HTTP Client
2. 发送请求
3. 检查 HTTP 状态码
4. 读取响应体
5. 解析 JSON
6. 错误包装（提供上下文）
```

**测试策略**：

```
测试用例 1：正常请求
- 启动 mock HTTP 服务器
- 返回正常 JSON 响应
- 验证：解析成功

测试用例 2：HTTP 错误（404、500）
- Mock 服务器返回 404
- 验证：返回 HTTP 错误

测试用例 3：JSON 解析错误
- Mock 服务器返回非 JSON
- 验证：返回解析错误

测试用例 4：超时
- Mock 服务器延迟响应
- 设置短超时
- 验证：返回超时错误
```

**思考题**：

**Q1**：什么时候应该提取公共函数？

<details>
<summary>提示</summary>

提取的时机（Rule of Three）：

- 第一次：写代码
- 第二次：注意到重复，但先不提取
- 第三次：确认模式，提取公共函数

原则：**不要过早抽象**，等模式稳定后再提取

</details>

**Q2**：如何处理 JSON 解析到不同类型的问题？

<details>
<summary>提示</summary>

选项 1：使用 Go 1.18 泛型

```go
func request[T any](url string) (*T, error) {
    var result T
    // ... 请求和解析
    return &result, nil
}

// 使用
backend, err := request[BackendInfo]("http://...")
```

选项 2：传入结果指针

```go
func request(url string, result interface{}) error {
    // ... 请求
    return json.Unmarshal(body, result)
}

// 使用
var backend BackendInfo
err := request("http://...", &backend)
```

选项 3：分离关注点

```go
// HTTP 层：只负责获取字节
func doHTTPRequest(url string) ([]byte, error)

// 业务层：负责解析
func getBackend(url string) (*BackendInfo, error) {
    body, err := doHTTPRequest(url)
    if err != nil {
        return nil, err
    }
    var result BackendInfo
    err = json.Unmarshal(body, &result)
    return &result, err
}
```

推荐：选项 3，职责更清晰

</details>

---

### 问题 4：HTTP 客户端最佳实践

**当前问题**：

```go
resp, err := http.Get(url)  // 使用默认 HTTP Client
```

**默认 Client 的问题**：

| 问题                   | 影响             | 生产环境风险      |
| ---------------------- | ---------------- | ----------------- |
| **无超时**       | 可能永久阻塞     | 🔴 高：资源泄漏   |
| **无连接池配置** | 性能不可控       | ⚠️ 中：并发受限 |
| **无重试**       | 网络抖动导致失败 | ⚠️ 中：不够健壮 |

**企业级 HTTP Client 设计**：

```go
生产级 HTTP Client 应该配置：

1. 超时控制
   - 总超时：整个请求的最大时间
   - 连接超时：建立连接的最大时间
   - 读取超时：读取响应的最大时间

2. 连接池
   - MaxIdleConns：总的空闲连接数
   - MaxIdleConnsPerHost：每个 host 的空闲连接数
   - IdleConnTimeout：空闲连接保持时间

3. 重试机制
   - 最大重试次数
   - 重试策略（指数退避）
   - 可重试的错误类型

4. 监控和日志
   - 请求日志（可选）
   - 性能指标（延迟、成功率）
```

**配置示例思路**：

```go
创建自定义 HTTP Client 的考虑因素：

1. 合理的超时时间
   - 太短：正常请求也会超时
   - 太长：失败请求占用资源太久
   - 建议：根据实际情况设置（通常 5-30 秒）

2. 连接池大小
   - 太小：并发受限
   - 太大：浪费资源
   - 建议：
     * MaxIdleConns: 100（总共）
     * MaxIdleConnsPerHost: 10-20

3. Keep-Alive
   - 复用 TCP 连接，减少握手开销
   - 建议：启用，设置 30-90 秒

结构设计提示：
type HTTPClient struct {
    client  *http.Client
    timeout time.Duration
    logger  Logger  // 可选
}

func NewHTTPClient(timeout time.Duration) *HTTPClient {
    // 创建自定义 Transport
    // 设置超时、连接池等
}

func (c *HTTPClient) Get(url string) ([]byte, error) {
    // 发送请求
    // 错误处理
    // 日志记录
}
```

**测试策略**：

```
测试超时配置：

测试用例 1：正常请求不超时
- 设置超时 5 秒
- Mock 服务器 100ms 后响应
- 验证：成功

测试用例 2：慢请求超时
- 设置超时 100ms
- Mock 服务器 1 秒后响应
- 验证：返回超时错误

测试连接池：

测试用例 3：并发请求复用连接
- 发送 100 个并发请求
- 监控建立的连接数
- 验证：连接数 <= MaxIdleConnsPerHost
```

**思考题**：

**Q1**：为什么需要超时控制？

<details>
<summary>提示</summary>

没有超时的风险：

1. **资源泄漏**：goroutine 和内存永久占用
2. **雪崩效应**：下游服务慢导致上游堵塞
3. **用户体验差**：用户无限等待

超时的好处：

- 快速失败（Fail Fast）
- 保护系统资源
- 更好的错误处理机会（可以重试）

推荐实践：

- 总超时：包含整个请求周期
- 使用 context.WithTimeout 控制

</details>

**Q2**：连接池的作用是什么？

<details>
<summary>提示</summary>

TCP 连接建立的开销：

1. 三次握手
2. TLS 握手（HTTPS）
3. 慢启动

连接池的作用：

- **复用连接**：避免重复握手
- **提高性能**：减少延迟
- **控制并发**：限制连接数

注意事项：

- Keep-Alive 要匹配服务器配置
- 空闲连接也占用资源，需要设置超时

</details>

**Q3**：什么错误应该重试？什么不应该？

<details>
<summary>提示</summary>

应该重试的错误（临时性错误）：

- 网络超时
- 连接被拒绝
- 5xx 服务器错误（服务器临时故障）

不应该重试的错误（永久性错误）：

- 4xx 客户端错误（请求本身有问题）
- 解析错误
- 业务逻辑错误

重试策略：

- 指数退避：避免雪崩
- 最大重试次数：防止无限重试
- 幂等性检查：确保重试安全

</details>

---

### 迭代 0 总结

**完成清单**：

- [X] 将 `ioutil.ReadAll` 替换为 `io.ReadAll`
- [X] 将所有 `panic` 改为返回 `error`
- [X] 在 `main` 函数中统一处理错误
- [X] 提取通用的 HTTP 请求函数
- [X] 创建自定义 HTTP Client（配置超时、连接池）
- [X] 编写测试验证错误处理

**验收标准**：

```bash
# 测试错误处理
./client  # 负载均衡器未启动
# 预期输出：友好的错误信息，程序退出码为 1

# 测试超时
./client --timeout 1ms
# 预期输出：超时错误

# 代码质量检查
go vet ./cmd/client/
golangci-lint run ./cmd/client/
```

**下一步**：

完成迭代 0 后，你将拥有：

- ✅ 健壮的错误处理
- ✅ 现代化的 API 使用
- ✅ 可复用的 HTTP 客户端

接下来我们将设计更高级的功能：多次请求、统计分析等。

---

## 迭代 1：客户端抽象设计

### 目标

设计清晰的抽象层，为后续功能扩展打基础。

### 问题场景

**场景 1：当前的问题**

```go
// 现在：所有逻辑都在 main 函数中
func main() {
    // 请求负载均衡器
    // 解析响应
    // 请求后端
    // 解析响应
    // 打印结果
}

问题：
- 逻辑混乱
- 难以测试
- 难以扩展（如何添加新功能？）
```

**场景 2：未来的需求**

```
需求 1：发送 100 次请求，统计分布
需求 2：并发发送请求，测试性能
需求 3：支持不同的负载均衡器（Nginx、HAProxy）
需求 4：输出不同格式（表格、JSON、CSV）

如何设计才能优雅地支持这些需求？
```

### 设计思路：职责分离

```
核心思想：把"做什么"和"怎么做"分开

当前：main 既知道做什么，又知道怎么做（高耦合）
目标：main 只知道做什么，具体怎么做交给专门的模块（低耦合）

类比：
- 你（main）想吃饭
- 不需要知道怎么种菜、怎么做饭（Client）
- 只需要调用外卖平台（抽象接口）
```

### 抽象设计

**层次 1：HTTP 层（最底层）**

```go
职责：发送 HTTP 请求，返回原始数据

为什么需要这一层？
- 隔离 HTTP 细节（超时、重试、连接池）
- 便于 Mock 测试
- 可以替换底层实现（如换成 fasthttp）

设计提示：
type HTTPClient interface {
    Get(url string) ([]byte, error)
    Post(url string, body []byte) ([]byte, error)
}

实现提示：
type DefaultHTTPClient struct {
    client  *http.Client
    timeout time.Duration
}

func (c *DefaultHTTPClient) Get(url string) ([]byte, error) {
    // 使用 c.client 发送请求
    // 检查状态码
    // 读取响应体
    // 错误处理
}
```

**层次 2：业务层**

```go
职责：解析业务数据

LoadBalancerClient：与负载均衡器交互
- GetBackend() (BackendInfo, error)

BackendClient：与后端服务器交互
- Request(url string) (string, error)

为什么要分成两个 Client？
- 单一职责：每个 Client 只负责一个服务
- 便于扩展：可以独立添加功能
- 便于测试：可以单独 Mock
```

**层次 3：应用层**

```go
职责：业务逻辑编排

Tester：控制测试流程
- Run() ([]RequestResult, error)

为什么需要 Tester？
- 封装测试逻辑（发送 N 次请求、并发控制）
- 可以有不同的实现（Sequential、Concurrent）
- main 函数只需要调用 tester.Run()
```

### 接口设计原则

**原则 1：接口隔离（ISP）**

```
不要设计胖接口：
❌ type Client interface {
    Get(url string) ([]byte, error)
    Post(url string, body []byte) ([]byte, error)
    Put(url string, body []byte) ([]byte, error)
    Delete(url string) ([]byte, error)
    // ... 20 个方法
}

应该设计小接口：
✅ type Getter interface {
    Get(url string) ([]byte, error)
}

✅ type Poster interface {
    Post(url string, body []byte) ([]byte, error)
}

原因：客户端只需要它需要的方法
```

**原则 2：依赖倒置（DIP）**

```go
高层模块不应该依赖低层模块，都应该依赖抽象

❌ 不好的设计：
type Tester struct {
    client *DefaultHTTPClient  // 依赖具体实现
}

✅ 好的设计：
type Tester struct {
    client HTTPClient  // 依赖接口
}

好处：
1. 可以替换实现（如测试时用 Mock）
2. 解耦，降低复杂度
```

### 数据结构设计

**RequestResult：请求结果**

```go
设计目标：记录单次请求的所有信息

需要记录什么？
1. 请求是否成功
2. 选择了哪个后端
3. 请求耗时
4. 错误信息（如果失败）
5. 时间戳（用于分析）

结构设计提示：
type RequestResult struct {
    // 核心信息
    Backend   string        // 后端服务器地址
    Success   bool          // 是否成功
    Latency   time.Duration // 延迟

    // 错误信息
    Error     error         // 错误（如果有）

    // 元数据
    Timestamp time.Time     // 请求时间
    RequestID int           // 请求序号（可选）
}

为什么这样设计？
- 包含所有需要统计的信息
- 便于后续分析（成功率、延迟分布、后端分布）
- 失败和成功都有完整记录
```

**BackendInfo：后端信息**

```go
已有的结构（来自负载均衡器响应）：
type BackendInfo struct {
    Name string `json:"name"`
    URL  string `json:"url"`
}

够用吗？需要添加什么？

扩展思路：
- 后端的权重（用于验证加权轮询）
- 后端的健康状态（用于测试健康检查）
- 后端的区域（用于测试地域路由）

建议：暂时不扩展，按需添加
```

### 实现提示

**步骤 1：定义接口**

```go
在 pkg/client/ 目录下创建接口定义

文件结构：
pkg/client/
├── http.go          # HTTP 层接口
├── loadbalancer.go  # 负载均衡器客户端接口
└── backend.go       # 后端服务器客户端接口

每个文件的职责：
- 定义接口
- 定义数据结构
- 提供构造函数
```

**步骤 2：实现 HTTP Client**

```go
关键点：
1. 使用自定义 http.Client（迭代 0 的成果）
2. 统一的错误处理
3. 可选的请求日志

错误分类思路：
- NetworkError：网络错误（连接失败、超时）
- HTTPError：HTTP 错误（4xx、5xx）
- ParseError：解析错误

实现方向：
func (c *DefaultHTTPClient) Get(url string) ([]byte, error) {
    // 1. 发送请求
    resp, err := c.client.Get(url)
    if err != nil {
        return nil, &NetworkError{...}
    }
    defer resp.Body.Close()

    // 2. 检查状态码
    if resp.StatusCode != 200 {
        return nil, &HTTPError{StatusCode: resp.StatusCode}
    }

    // 3. 读取响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, &NetworkError{...}
    }

    return body, nil
}
```

**步骤 3：实现 LoadBalancer Client**

```go
职责：
- 请求负载均衡器
- 解析后端信息
- 错误处理

关键设计：
1. 依赖 HTTPClient 接口（不是具体实现）
2. 封装解析逻辑
3. 返回强类型的 BackendInfo

结构提示：
type LoadBalancerClient struct {
    httpClient HTTPClient
    baseURL    string
}

func (c *LoadBalancerClient) GetBackend() (BackendInfo, error) {
    // 1. 构造 URL
    url := c.baseURL + "/balancer"

    // 2. 发送请求
    body, err := c.httpClient.Get(url)
    if err != nil {
        return BackendInfo{}, fmt.Errorf("获取后端失败: %w", err)
    }

    // 3. 解析 JSON
    var resp struct {
        Data BackendInfo `json:"data"`
    }
    if err := json.Unmarshal(body, &resp); err != nil {
        return BackendInfo{}, fmt.Errorf("解析响应失败: %w", err)
    }

    return resp.Data, nil
}
```

**步骤 4：实现 Backend Client**

```go
职责：
- 请求后端服务器
- 解析响应
- 返回结果

实现思路类似 LoadBalancerClient
```

### 测试策略

**单元测试：使用 Mock**

```go
测试 LoadBalancerClient.GetBackend()：

测试用例 1：正常响应
- Mock HTTPClient 返回正常 JSON
- 验证：解析正确

测试用例 2：HTTP 错误
- Mock HTTPClient 返回错误
- 验证：错误传播正确

测试用例 3：JSON 解析错误
- Mock HTTPClient 返回非法 JSON
- 验证：返回解析错误

Mock 实现提示：
type MockHTTPClient struct {
    Response []byte
    Error    error
}

func (m *MockHTTPClient) Get(url string) ([]byte, error) {
    return m.Response, m.Error
}

// 测试代码
func TestGetBackend_Success(t *testing.T) {
    mockClient := &MockHTTPClient{
        Response: []byte(`{"data":{"name":"s1","url":"http://localhost:8081"}}`),
        Error:    nil,
    }

    lbClient := &LoadBalancerClient{
        httpClient: mockClient,
        baseURL:    "http://test",
    }

    backend, err := lbClient.GetBackend()

    // 断言
    assert.NoError(t, err)
    assert.Equal(t, "s1", backend.Name)
    assert.Equal(t, "http://localhost:8081", backend.URL)
}
```

**集成测试：真实服务器**

```go
测试策略：
1. 启动真实的负载均衡器和后端服务器
2. 创建真实的 HTTP Client
3. 发送请求
4. 验证结果

提示：可以在 CI/CD 中自动化
```

### 思考题

**Q1**：为什么要分 HTTP 层和业务层？

<details>
<summary>提示</summary>

分层的好处：

1. **关注点分离**：

   - HTTP 层只关心发送请求
   - 业务层只关心业务逻辑
2. **可测试性**：

   - HTTP 层可以独立测试
   - 业务层可以 Mock HTTP 层
3. **可替换性**：

   - 可以换用其他 HTTP 库
   - 可以换用其他协议（gRPC）
4. **可复用性**：

   - HTTP Client 可以在其他项目中使用

</details>

**Q2**：接口应该定义在哪里？

<details>
<summary>提示</summary>

Go 的接口设计哲学：**接受接口，返回结构**

选项 1：定义在使用方

```go
// tester/tester.go
type HTTPClient interface {  // Tester 定义它需要的接口
    Get(url string) ([]byte, error)
}
```

选项 2：定义在实现方

```go
// client/http.go
type HTTPClient interface {  // HTTP Client 定义接口
    Get(url string) ([]byte, error)
}
```

推荐：选项 1（Go 惯用法）

- 使用方定义它需要的最小接口
- 实现方实现接口（隐式实现）
- 降低耦合

例外：如果接口会被多处使用，可以放在公共包

</details>

**Q3**：如何设计才能支持未来的扩展？

<details>
<summary>提示</summary>

开闭原则（OCP）：对扩展开放，对修改关闭

示例：支持不同的负载均衡器

❌ 不好的设计：

```go
type Client struct {
    balancerType string  // "nginx", "haproxy", "custom"
}

func (c *Client) GetBackend() (BackendInfo, error) {
    switch c.balancerType {
    case "nginx":
        // nginx 逻辑
    case "haproxy":
        // haproxy 逻辑
    default:
        // 默认逻辑
    }
}
```

问题：每次添加新类型都要修改代码

✅ 好的设计：

```go
type LoadBalancer interface {
    GetBackend() (BackendInfo, error)
}

type CustomLoadBalancer struct { ... }
type NginxLoadBalancer struct { ... }

// Tester 依赖接口
type Tester struct {
    balancer LoadBalancer
}
```

好处：添加新类型只需要实现接口

</details>

---

### 迭代 1 总结

**完成清单**：

- [ ] 定义 HTTPClient 接口
- [ ] 实现 DefaultHTTPClient
- [ ] 定义 LoadBalancerClient 接口并实现
- [ ] 定义 BackendClient 接口并实现
- [ ] 定义 RequestResult 数据结构
- [ ] 编写单元测试（使用 Mock）
- [ ] 编写集成测试（使用真实服务器）

**目录结构**：

```
cmd/client/
└── main.go

pkg/client/
├── http.go           # HTTP Client 接口和实现
├── loadbalancer.go   # LoadBalancer Client
├── backend.go        # Backend Client
└── types.go          # 数据结构定义

test/client/
├── http_test.go
├── loadbalancer_test.go
└── backend_test.go
```

**验收标准**：

```bash
# 运行单元测试
go test ./pkg/client/... -v

# 运行集成测试（需要启动服务器）
./start_server.sh  # 启动后端服务器
go run cmd/lb/main.go &  # 启动负载均衡器
go test ./test/client/... -v -tags=integration
```

---

## 迭代 2：配置管理

### 目标

实现灵活的配置管理，支持命令行参数、环境变量、配置文件。

### 问题场景

**场景 1：硬编码的问题**

```go
// 当前代码
const balancerURL = "http://127.0.0.1:8187/balancer"
const requestCount = 1

问题：
- 修改配置需要重新编译
- 不同环境（开发、测试、生产）无法灵活切换
- 无法通过命令行快速测试
```

**场景 2：用户使用场景**

```bash
# 场景 1：快速测试，使用默认配置
./client

# 场景 2：指定负载均衡器地址
./client --lb-url http://192.168.1.100:8187

# 场景 3：发送 100 次请求
./client --count 100

# 场景 4：并发测试
./client --count 1000 --concurrent 10

# 场景 5：设置超时
./client --timeout 5s

# 场景 6：输出 JSON 格式
./client --count 100 --output json
```

### 设计思路：配置分层

```
配置优先级（从高到低）：
1. 命令行参数（最高优先级）
2. 环境变量
3. 配置文件
4. 默认值（最低优先级）

原理：
- 命令行：用于临时覆盖，快速测试
- 环境变量：用于环境相关配置（CI/CD）
- 配置文件：用于持久化配置
- 默认值：提供开箱即用的体验
```

### 配置项设计

**核心配置**：

```go
需要哪些配置？

1. 负载均衡器配置
   - URL
   - 超时时间

2. 测试配置
   - 请求次数
   - 并发数
   - 测试持续时间（alternative to count）

3. 输出配置
   - 输出格式（table, json, csv）
   - 是否显示详细信息

4. HTTP 配置
   - 超时
   - 重试次数
   - Keep-Alive

结构设计提示：
type Config struct {
    // 负载均衡器配置
    LoadBalancerURL string

    // 测试配置
    RequestCount int
    Concurrent   int
    Duration     time.Duration

    // HTTP 配置
    Timeout    time.Duration
    MaxRetries int

    // 输出配置
    OutputFormat string  // "table", "json", "csv"
    Verbose      bool
}
```

### 实现方案

**方案 1：使用 flag 包（标准库）**

```go
优点：
+ 无依赖
+ 简单直接
+ 适合简单场景

缺点：
- 功能有限（只支持命令行参数）
- 没有子命令支持
- 错误提示不够友好

适用场景：简单工具

示例：
var (
    lbURL = flag.String("lb-url", "http://localhost:8187", "负载均衡器地址")
    count = flag.Int("count", 10, "请求次数")
)

func main() {
    flag.Parse()
    fmt.Println(*lbURL, *count)
}
```

**方案 2：使用 pflag 包**

```go
优点：
+ 兼容 flag
+ 支持短选项（-c, --count）
+ 更好的类型支持

缺点：
- 需要额外依赖
- 仍然只支持命令行

适用场景：需要更好的命令行体验

import "github.com/spf13/pflag"

var (
    lbURL = pflag.StringP("lb-url", "u", "http://localhost:8187", "负载均衡器地址")
    count = pflag.IntP("count", "c", 10, "请求次数")
)
```

**方案 3：使用 cobra + viper（企业级）**

```go
优点：
+ 支持子命令
+ 支持配置文件
+ 支持环境变量
+ 自动生成帮助文档

缺点：
- 依赖较重
- 学习曲线

适用场景：复杂的 CLI 工具

架构：
- cobra：命令行解析和子命令
- viper：配置管理（文件、环境变量）
```

### 推荐方案：分阶段实现

**第一阶段：使用 flag 包**

```go
目标：快速实现基本功能

步骤：
1. 定义 Config 结构
2. 使用 flag 解析命令行参数
3. 提供合理的默认值
4. 验证配置（如 count > 0）

实现提示：
func parseFlags() *Config {
    cfg := &Config{}

    // 定义 flags
    flag.StringVar(&cfg.LoadBalancerURL, "lb-url", "http://localhost:8187", "负载均衡器URL")
    flag.IntVar(&cfg.RequestCount, "count", 10, "请求次数")
    flag.IntVar(&cfg.Concurrent, "concurrent", 1, "并发数")
    flag.DurationVar(&cfg.Timeout, "timeout", 5*time.Second, "请求超时")
    flag.StringVar(&cfg.OutputFormat, "output", "table", "输出格式")
    flag.BoolVar(&cfg.Verbose, "verbose", false, "详细模式")

    flag.Parse()

    return cfg
}

func (c *Config) Validate() error {
    // 验证配置
    if c.RequestCount <= 0 {
        return fmt.Errorf("请求次数必须大于 0")
    }
    if c.Concurrent <= 0 {
        return fmt.Errorf("并发数必须大于 0")
    }
    if c.Concurrent > c.RequestCount {
        return fmt.Errorf("并发数不能大于请求次数")
    }
    return nil
}
```

**第二阶段：支持配置文件（可选）**

```go
目标：支持持久化配置

配置文件格式（YAML）：
# client.yaml
load_balancer:
  url: http://localhost:8187
  timeout: 5s

test:
  count: 100
  concurrent: 10

output:
  format: table
  verbose: false

读取逻辑提示：
func loadConfig(path string) (*Config, error) {
    // 1. 读取文件
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return defaultConfig(), nil  // 文件不存在使用默认值
        }
        return nil, err
    }

    // 2. 解析 YAML
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("解析配置文件失败: %w", err)
    }

    return &cfg, nil
}

配置合并逻辑：
func mergeConfig(file, flags *Config) *Config {
    // 命令行参数覆盖配置文件
    // 使用反射或手动合并
}
```

**第三阶段：支持环境变量（可选）**

```go
目标：支持环境变量（用于 CI/CD）

环境变量命名规范：
LB_CLIENT_URL=http://localhost:8187
LB_CLIENT_COUNT=100
LB_CLIENT_CONCURRENT=10

读取逻辑提示：
func loadFromEnv(cfg *Config) {
    if url := os.Getenv("LB_CLIENT_URL"); url != "" {
        cfg.LoadBalancerURL = url
    }
    if count := os.Getenv("LB_CLIENT_COUNT"); count != "" {
        cfg.RequestCount, _ = strconv.Atoi(count)
    }
    // ...
}
```

### 帮助文档设计

**设计原则：让用户自助**

```go
好的帮助文档应该包含：
1. 简短描述
2. 使用示例
3. 参数列表和说明
4. 默认值

示例：
负载均衡器测试客户端

用法:
  client [flags]

示例:
  # 发送 10 次请求（默认）
  client

  # 发送 100 次请求到指定地址
  client --lb-url http://192.168.1.100:8187 --count 100

  # 并发测试
  client --count 1000 --concurrent 10

参数:
  --lb-url string      负载均衡器地址 (默认 "http://localhost:8187")
  -c, --count int      请求次数 (默认 10)
  --concurrent int     并发数 (默认 1)
  --timeout duration   请求超时 (默认 5s)
  --output string      输出格式: table, json, csv (默认 "table")
  -v, --verbose        详细模式
  -h, --help           帮助信息

实现：
flag.Usage = func() {
    fmt.Fprintf(os.Stderr, helpText)
    flag.PrintDefaults()
}
```

### 配置验证

**设计思路：快速失败（Fail Fast）**

```go
验证时机：程序启动时立即验证

验证内容：
1. 必需参数是否提供
2. 参数值是否合理
3. 参数之间的关系是否正确

实现提示：
func (c *Config) Validate() error {
    var errs []string

    // 验证 URL
    if c.LoadBalancerURL == "" {
        errs = append(errs, "负载均衡器 URL 不能为空")
    }
    if _, err := url.Parse(c.LoadBalancerURL); err != nil {
        errs = append(errs, "负载均衡器 URL 格式错误")
    }

    // 验证数值范围
    if c.RequestCount <= 0 {
        errs = append(errs, "请求次数必须大于 0")
    }
    if c.RequestCount > 1000000 {
        errs = append(errs, "请求次数过大，建议 <= 1000000")
    }

    // 验证关系
    if c.Concurrent > c.RequestCount {
        errs = append(errs, "并发数不能大于请求次数")
    }

    // 验证枚举值
    validFormats := []string{"table", "json", "csv"}
    if !contains(validFormats, c.OutputFormat) {
        errs = append(errs, "输出格式必须是: "+strings.Join(validFormats, ", "))
    }

    if len(errs) > 0 {
        return fmt.Errorf("配置验证失败:\n  - %s", strings.Join(errs, "\n  - "))
    }

    return nil
}

使用：
cfg := parseFlags()
if err := cfg.Validate(); err != nil {
    fmt.Fprintf(os.Stderr, "%v\n", err)
    os.Exit(1)
}
```

### 测试策略

**单元测试：配置解析**

```go
测试用例 1：默认配置
- 不传任何参数
- 验证：使用默认值

测试用例 2：覆盖默认值
- 传入自定义参数
- 验证：使用传入的值

测试用例 3：配置验证
- 传入非法值（count = -1）
- 验证：返回验证错误

测试用例 4：配置文件读取
- 提供配置文件
- 验证：正确读取并合并

实现提示（使用 flag.FlagSet 便于测试）：
func TestParseFlags(t *testing.T) {
    // 重置 flag
    flags := flag.NewFlagSet("test", flag.ContinueOnError)

    // 定义 flags
    // ...

    // 解析
    flags.Parse([]string{"--count", "100"})

    // 验证
    assert.Equal(t, 100, count)
}
```

### 思考题

**Q1**：为什么要分层配置（命令行、环境变量、配置文件）？

<details>
<summary>提示</summary>

不同来源的用途：

1. **命令行**：

   - 临时覆盖
   - 快速测试
   - 一次性使用
   - 示例：`./client --count 1000`
2. **环境变量**：

   - 环境相关配置
   - CI/CD 集成
   - 容器化部署
   - 示例：`export LB_CLIENT_URL=http://prod-lb:8080`
3. **配置文件**：

   - 持久化配置
   - 团队共享
   - 复杂配置
   - 示例：`client.yaml`
4. **默认值**：

   - 开箱即用
   - 约定优于配置

优先级设计原则：

- 越临时的配置，优先级越高
- 越持久的配置，优先级越低

</details>

**Q2**：如何设计合理的默认值？

<details>
<summary>提示</summary>

好的默认值应该：

1. **安全**：不会造成破坏（如 count=10 而不是 1000000）
2. **常用**：覆盖大部分使用场景
3. **易理解**：符合直觉

示例：

- RequestCount: 10（足够看到分布，又不会太慢）
- Concurrent: 1（顺序测试，简单）
- Timeout: 5s（大部分请求能完成，又不会等太久）
- OutputFormat: "table"（最直观）

避免：

- 没有默认值（用户必须手动指定）
- 默认值太大（如 count=1000000，慢且危险）
- 默认值太小（如 timeout=1ms，几乎总是失败）

</details>

**Q3**：配置验证应该多严格？

<details>
<summary>提示</summary>

验证策略：

**必须验证**：

- 类型正确（已由 flag 保证）
- 业务规则（count > 0）
- 安全限制（count <= 1000000）

**建议验证**：

- URL 格式（可以提前发现错误）
- 文件存在性（配置文件）

**可选验证**：

- 网络可达性（连接负载均衡器）
  - 优点：提前发现问题
  - 缺点：增加启动时间，可能误报

**不应该验证**：

- 负载均衡器的具体实现
- 后端服务器的状态

原则：**验证配置本身，不验证运行时状态**

</details>

---

### 迭代 2 总结

**完成清单**：

- [ ] 定义 Config 结构
- [ ] 使用 flag 包解析命令行参数
- [ ] 实现配置验证
- [ ] 设计帮助文档
- [ ] 编写配置解析测试
- [ ] （可选）支持配置文件
- [ ] （可选）支持环境变量

**文件结构**：

```
cmd/client/
├── main.go
└── config.go      # 配置管理

configs/
└── client.yaml    # 示例配置文件（可选）
```

**验收标准**：

```bash
# 测试默认配置
./client --help
./client

# 测试自定义配置
./client --count 100 --concurrent 10

# 测试配置验证
./client --count -1  # 应该报错
./client --concurrent 100 --count 10  # 应该报错

# 测试配置文件（如果实现）
./client --config client.yaml
```

**下一步**：

有了配置管理，接下来实现：

- 测试器：根据配置发送请求
- 统计分析：分析请求结果
- 报告输出：根据 OutputFormat 输出

---

## 迭代 3：测试器设计

### 目标

实现测试逻辑，支持顺序和并发两种模式。

### 问题场景

**场景 1：当前的局限**

```go
// 当前代码：只能发送 1 次请求
backend, err := lbClient.GetBackend()
// 请求后端
// 打印结果

问题：
- 无法测试负载均衡效果（需要多次请求）
- 无法收集统计数据
- 无法并发测试
```

**场景 2：测试需求**

```
测试场景 1：顺序测试
- 发送 N 次请求
- 记录每次请求的结果
- 用于验证算法正确性（Round Robin 应该均匀）

测试场景 2：并发测试
- M 个并发，每个发送 N/M 次请求
- 用于测试性能和并发安全性

测试场景 3：持续时间测试
- 持续发送请求 D 秒
- 用于稳定性测试
```

### 设计思路：策略模式

```
核心思想：分离测试逻辑和执行方式

┌─────────────────────────────────┐
│  Tester (接口)                  │
│  + Run() []RequestResult        │
└────────────┬────────────────────┘
             │
      ┌──────┴──────┐
      │             │
      ↓             ↓
┌─────────────┐ ┌──────────────┐
│ Sequential  │ │ Concurrent   │
│ Tester      │ │ Tester       │
└─────────────┘ └──────────────┘

好处：
- 可以独立测试每种模式
- 可以轻松添加新模式（如 Duration Tester）
- main 函数不需要关心具体如何测试
```

### 接口设计

**Tester 接口**：

```go
职责：执行测试，返回结果

type Tester interface {
    Run(ctx context.Context) ([]RequestResult, error)
}

为什么需要 context？
1. 超时控制：整个测试的超时
2. 取消：用户可以 Ctrl+C 中断
3. 传递元数据：如 tracing ID

为什么返回切片？
- 需要分析所有请求的结果
- 计算统计数据（成功率、延迟、分布）
```

### 顺序测试器设计

**职责**：按顺序发送 N 次请求

```go
结构设计：
type SequentialTester struct {
    lbClient      LoadBalancerClient
    backendClient BackendClient
    count         int
}

核心逻辑：
func (t *SequentialTester) Run(ctx context.Context) ([]RequestResult, error) {
    results := make([]RequestResult, 0, t.count)

    for i := 0; i < t.count; i++ {
        // 检查 context 是否取消
        select {
        case <-ctx.Done():
            return results, ctx.Err()
        default:
        }

        // 执行单次请求
        result := t.doSingleRequest(i)
        results = append(results, result)
    }

    return results, nil
}

单次请求逻辑：
func (t *SequentialTester) doSingleRequest(requestID int) RequestResult {
    result := RequestResult{
        RequestID: requestID,
        Timestamp: time.Now(),
    }

    start := time.Now()

    // 1. 获取后端
    backend, err := t.lbClient.GetBackend()
    if err != nil {
        result.Success = false
        result.Error = fmt.Errorf("获取后端失败: %w", err)
        return result
    }

    // 2. 请求后端
    _, err = t.backendClient.Request(backend.URL)
    if err != nil {
        result.Success = false
        result.Error = fmt.Errorf("请求后端失败: %w", err)
        return result
    }

    // 3. 记录成功
    result.Success = true
    result.Backend = backend.Name  // 或 backend.URL
    result.Latency = time.Since(start)

    return result
}
```

**关键设计点**：

| 设计点                 | 说明                                | 为什么                 |
| ---------------------- | ----------------------------------- | ---------------------- |
| **预分配切片**   | `make([]RequestResult, 0, count)` | 避免多次扩容，提高性能 |
| **context 检查** | 每次循环检查 `ctx.Done()`         | 支持优雅取消           |
| **错误不中断**   | 单次失败不影响后续请求              | 收集完整数据           |
| **记录延迟**     | `time.Since(start)`               | 用于性能分析           |

### 并发测试器设计

**职责**：并发发送请求

**并发模型**：

```
假设：
- 总请求数：100
- 并发数：10

模型：
┌─────────┐ ┌─────────┐     ┌─────────┐
│Worker 1 │ │Worker 2 │ ... │Worker 10│
│ 10 req  │ │ 10 req  │     │ 10 req  │
└────┬────┘ └────┬────┘     └────┬────┘
     │           │                │
     └───────────┼────────────────┘
                 ↓
          Result Channel
                 ↓
         Collect Results
```

**实现挑战**：

| 挑战               | 问题                     | 解决方案                   |
| ------------------ | ------------------------ | -------------------------- |
| **结果收集** | 多个 goroutine 同时写入  | 使用 channel               |
| **等待完成** | 如何知道所有 worker 完成 | 使用 sync.WaitGroup        |
| **错误处理** | 某个 worker panic        | recover + error channel    |
| **均匀分配** | 100 个请求，10 个 worker | 每个 10 个，余数分给前几个 |

**设计思路**：

```go
type ConcurrentTester struct {
    lbClient      LoadBalancerClient
    backendClient BackendClient
    totalCount    int
    concurrent    int
}

func (t *ConcurrentTester) Run(ctx context.Context) ([]RequestResult, error) {
    // 1. 计算每个 worker 的请求数
    requestsPerWorker := t.totalCount / t.concurrent
    remainder := t.totalCount % t.concurrent

    // 2. 创建结果 channel
    resultCh := make(chan RequestResult, t.totalCount)

    // 3. 创建 WaitGroup
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

    // 5. 等待所有 worker 完成
    go func() {
        wg.Wait()
        close(resultCh)  // 关闭 channel
    }()

    // 6. 收集结果
    results := make([]RequestResult, 0, t.totalCount)
    for result := range resultCh {
        results = append(results, result)
    }

    return results, nil
}

func (t *ConcurrentTester) runWorker(
    ctx context.Context,
    workerID int,
    count int,
    resultCh chan<- RequestResult,
) {
    for i := 0; i < count; i++ {
        select {
        case <-ctx.Done():
            return
        default:
        }

        result := t.doSingleRequest(workerID, i)
        resultCh <- result
    }
}
```

**关键设计点**：

| 设计点                       | 代码                                     | 原因                    |
| ---------------------------- | ---------------------------------------- | ----------------------- |
| **buffered channel**   | `make(chan RequestResult, totalCount)` | 避免 worker 阻塞        |
| **close channel**      | `close(resultCh)` after `wg.Wait()`  | 通知收集 goroutine 结束 |
| **均匀分配**           | 余数分给前几个 worker                    | 确保总数正确            |
| **goroutine 安全退出** | 检查 `ctx.Done()`                      | 避免泄漏                |

**思考题**：

**Q1**：为什么要用 buffered channel？

<details>
<summary>提示</summary>

unbuffered vs buffered：

**Unbuffered channel**（`make(chan T)`）：

- 发送方阻塞，直到接收方接收
- 风险：如果收集端慢，worker 会阻塞

**Buffered channel**（`make(chan T, n)`）：

- 缓冲区未满时，发送方不阻塞
- 适合生产者-消费者模式

本场景：

- 生产者：多个 worker（快）
- 消费者：一个收集者（相对慢）
- 使用 buffered channel 避免 worker 等待

</details>

**Q2**：如何处理余数分配？

<details>
<summary>提示</summary>

问题：100 个请求，10 个 worker，如何分配？

方案 1：前几个多分配

```
Worker 0: 10
Worker 1: 10
...
Worker 9: 10
总计: 100
```

方案 2：前几个少分配

```
Worker 0: 9
Worker 1: 9
...
Worker 9: 11
总计: 100
```

推荐：方案 1

- 代码简单
- 差异小（最多差 1）

</details>

### 测试策略

**单元测试：Mock**

```go
测试 SequentialTester：

测试用例 1：全部成功
- Mock LBClient 和 BackendClient 返回成功
- 发送 10 次请求
- 验证：
  * 返回 10 个结果
  * 所有结果 Success = true
  * 调用了 10 次 GetBackend

测试用例 2：部分失败
- Mock 前 5 次成功，后 5 次失败
- 验证：
  * 返回 10 个结果（不中断）
  * 5 个成功，5 个失败

测试用例 3：context 取消
- 启动测试
- 中途调用 cancel()
- 验证：提前返回

Mock 实现示例：
type MockLBClient struct {
    backends []BackendInfo
    index    int
}

func (m *MockLBClient) GetBackend() (BackendInfo, error) {
    backend := m.backends[m.index%len(m.backends)]
    m.index++
    return backend, nil
}
```

**集成测试：真实服务器**

```go
测试 ConcurrentTester：

测试用例 1：验证并发数
- 启动 10 个并发
- 验证：实际创建了 10 个 goroutine

测试用例 2：验证总数
- 100 个请求，10 个并发
- 验证：返回 100 个结果

测试用例 3：验证分布
- 5 个后端，100 次请求，Round Robin
- 验证：每个后端约 20 次

如何验证并发？
- 使用计数器
- 使用 channel 监控
- 使用 pprof 查看 goroutine
```

### 思考题

**Q1**：为什么错误不中断测试？

<details>
<summary>提示</summary>

场景：100 次请求，第 10 次失败

方案 1：失败立即返回

```go
if err != nil {
    return nil, err
}
```

问题：只能看到 10 个结果，无法分析后续

方案 2：记录错误，继续执行

```go
if err != nil {
    result.Success = false
    result.Error = err
}
results = append(results, result)
```

好处：

- 收集完整数据
- 可以分析失败率
- 可以发现间歇性问题

例外：如果所有请求都失败，可以考虑提前退出（优化）

</details>

**Q2**：如何优雅处理 context 取消？

<details>
<summary>提示</summary>

问题：用户按 Ctrl+C，如何停止测试？

解决方案：

```go
// 1. main 函数设置 signal 处理
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigCh
    fmt.Println("\n接收到中断信号，正在停止...")
    cancel()
}()

// 2. Tester 检查 context
select {
case <-ctx.Done():
    return results, ctx.Err()
default:
}

// 3. 报告部分结果
if err == context.Canceled {
    fmt.Printf("测试被中断，已完成 %d/%d 次请求\n", len(results), total)
}
```

</details>

---

### 迭代 3 总结

**完成清单**：

- [ ] 定义 Tester 接口
- [ ] 实现 SequentialTester
- [ ] 实现 ConcurrentTester
- [ ] 实现 context 取消支持
- [ ] 编写单元测试（Mock）
- [ ] 编写集成测试（真实服务器）

**文件结构**：

```
pkg/tester/
├── tester.go         # Tester 接口
├── sequential.go     # 顺序测试器
├── concurrent.go     # 并发测试器
└── types.go          # RequestResult 等

test/tester/
├── sequential_test.go
└── concurrent_test.go
```

**验收标准**：

```bash
# 单元测试
go test ./pkg/tester/... -v

# 集成测试
./start_server.sh
go run cmd/lb/main.go &

# 顺序测试
go run cmd/client/main.go --count 100

# 并发测试
go run cmd/client/main.go --count 1000 --concurrent 10
```

---

## 迭代 4：统计分析

### 目标

从请求结果中提取统计信息，为报告输出提供数据。

### 问题场景

**场景 1：原始数据 vs 统计数据**

```go
// 原始数据：[]RequestResult
// 100 个 RequestResult，如何分析？

问题：
- 成功率是多少？
- 平均延迟？
- 每个后端分配了多少请求？
- 延迟分布（P50, P90, P99）？
```

**场景 2：用户需求**

```
用户场景 1：验证 Round Robin
- 发送 100 次请求到 5 个后端
- 期望：每个后端 20 次
- 需要统计：后端分布

用户场景 2：性能测试
- 发送 1000 次请求
- 需要知道：平均延迟、P99 延迟、QPS

用户场景 3：稳定性测试
- 发送 10000 次请求
- 需要知道：成功率、错误分布
```

### 设计思路：统计对象

```go
设计目标：一个清晰的统计数据结构

type Statistics struct {
    // 基础统计
    TotalRequests int
    SuccessCount  int
    FailureCount  int
    SuccessRate   float64

    // 延迟统计
    TotalDuration time.Duration  // 总耗时
    AvgLatency    time.Duration  // 平均延迟
    MinLatency    time.Duration
    MaxLatency    time.Duration
    P50Latency    time.Duration  // 中位数
    P90Latency    time.Duration
    P95Latency    time.Duration
    P99Latency    time.Duration

    // 性能指标
    QPS float64  // 每秒请求数

    // 后端分布
    BackendDistribution map[string]int      // backend -> count
    BackendPercentage   map[string]float64  // backend -> percentage

    // 错误统计
    ErrorDistribution map[string]int  // error type -> count
}

为什么这样设计？
- 包含所有常见指标
- 便于输出（table, json）
- 便于验证（如检查分布是否均匀）
```

### 核心算法

**算法 1：计算百分位数（Percentile）**

```go
问题：给定延迟列表，计算 P50, P90, P99

原理：
- P50：50% 的请求延迟 <= P50
- P90：90% 的请求延迟 <= P90
- P99：99% 的请求延迟 <= P99

算法步骤：
1. 对延迟列表排序
2. 计算位置：index = len(latencies) * percentage / 100
3. 取对应位置的值

示例：
latencies = [1ms, 2ms, 3ms, 4ms, 5ms]  // 5 个值

P50: index = 5 * 50 / 100 = 2.5 ≈ 2 → latencies[2] = 3ms
P90: index = 5 * 90 / 100 = 4.5 ≈ 4 → latencies[4] = 5ms

实现提示：
func calculatePercentile(latencies []time.Duration, percentile float64) time.Duration {
    if len(latencies) == 0 {
        return 0
    }

    // 1. 排序（注意：会修改原切片，可能需要复制）
    sort.Slice(latencies, func(i, j int) bool {
        return latencies[i] < latencies[j]
    })

    // 2. 计算位置
    index := int(float64(len(latencies)) * percentile / 100.0)
    if index >= len(latencies) {
        index = len(latencies) - 1
    }

    return latencies[index]
}

注意事项：
- 需要复制切片避免修改原始数据
- 边界情况：空切片、单个元素
- 性能：排序的时间复杂度 O(n log n)
```

**算法 2：计算 QPS**

```go
问题：如何计算 QPS（Queries Per Second）？

公式：QPS = 总请求数 / 总耗时（秒）

实现思路：
func calculateQPS(totalRequests int, duration time.Duration) float64 {
    if duration == 0 {
        return 0
    }
    return float64(totalRequests) / duration.Seconds()
}

总耗时如何计算？
方案 1：第一个请求到最后一个请求的时间差
方案 2：所有请求延迟之和

推荐：方案 1（更准确反映实际耗时）

实现：
firstTimestamp := results[0].Timestamp
lastTimestamp := results[len(results)-1].Timestamp
totalDuration := lastTimestamp.Sub(firstTimestamp)
```

**算法 3：计算分布百分比**

```go
问题：每个后端占比是多少？

实现思路：
func calculatePercentage(distribution map[string]int, total int) map[string]float64 {
    percentage := make(map[string]float64)

    for backend, count := range distribution {
        percentage[backend] = float64(count) / float64(total) * 100.0
    }

    return percentage
}

结果示例：
{
    "localhost:8081": 20.0,
    "localhost:8082": 20.0,
    "localhost:8083": 20.0,
    "localhost:8084": 20.0,
    "localhost:8085": 20.0,
}
```

### 实现思路

**分析器设计**：

```go
type Analyzer struct{}

func (a *Analyzer) Analyze(results []RequestResult) (*Statistics, error) {
    if len(results) == 0 {
        return nil, fmt.Errorf("没有结果可分析")
    }

    stats := &Statistics{
        BackendDistribution: make(map[string]int),
        ErrorDistribution:   make(map[string]int),
    }

    // 1. 基础统计
    stats.TotalRequests = len(results)

    // 2. 收集数据（一次遍历）
    latencies := make([]time.Duration, 0, len(results))

    for _, result := range results {
        if result.Success {
            stats.SuccessCount++
            stats.BackendDistribution[result.Backend]++
            latencies = append(latencies, result.Latency)
        } else {
            stats.FailureCount++
            // 错误分类
            errorType := classifyError(result.Error)
            stats.ErrorDistribution[errorType]++
        }
    }

    // 3. 计算派生指标
    stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalRequests) * 100.0

    // 4. 延迟统计（如果有成功的请求）
    if len(latencies) > 0 {
        stats.AvgLatency = calculateAverage(latencies)
        stats.MinLatency = calculateMin(latencies)
        stats.MaxLatency = calculateMax(latencies)
        stats.P50Latency = calculatePercentile(latencies, 50)
        stats.P90Latency = calculatePercentile(latencies, 90)
        stats.P95Latency = calculatePercentile(latencies, 95)
        stats.P99Latency = calculatePercentile(latencies, 99)
    }

    // 5. QPS 计算
    totalDuration := calculateTotalDuration(results)
    stats.TotalDuration = totalDuration
    stats.QPS = calculateQPS(len(results), totalDuration)

    // 6. 后端百分比
    stats.BackendPercentage = calculatePercentage(stats.BackendDistribution, stats.SuccessCount)

    return stats, nil
}

辅助函数：
func calculateAverage(durations []time.Duration) time.Duration {
    var sum time.Duration
    for _, d := range durations {
        sum += d
    }
    return sum / time.Duration(len(durations))
}

func calculateMin(durations []time.Duration) time.Duration {
    min := durations[0]
    for _, d := range durations {
        if d < min {
            min = d
        }
    }
    return min
}

func calculateMax(durations []time.Duration) time.Duration {
    max := durations[0]
    for _, d := range durations {
        if d > max {
            max = d
        }
    }
    return max
}

func calculateTotalDuration(results []RequestResult) time.Duration {
    if len(results) == 0 {
        return 0
    }
    first := results[0].Timestamp
    last := results[len(results)-1].Timestamp
    return last.Sub(first)
}

func classifyError(err error) string {
    if err == nil {
        return "unknown"
    }

    // 简单分类
    errStr := err.Error()
    if strings.Contains(errStr, "timeout") {
        return "timeout"
    }
    if strings.Contains(errStr, "connection refused") {
        return "connection_refused"
    }
    return "other"
}
```

### 性能优化

**优化点 1：避免多次遍历**

```go
❌ 不优雅：
// 遍历 1：计算成功数
for _, result := range results {
    if result.Success {
        successCount++
    }
}

// 遍历 2：收集延迟
for _, result := range results {
    if result.Success {
        latencies = append(latencies, result.Latency)
    }
}

✅ 优雅：一次遍历完成所有统计
for _, result := range results {
    if result.Success {
        successCount++
        latencies = append(latencies, result.Latency)
        backendDistribution[result.Backend]++
    }
}
```

**优化点 2：避免不必要的复制**

```go
问题：calculatePercentile 需要排序，会修改原切片

方案 1：复制后排序
latenciesCopy := make([]time.Duration, len(latencies))
copy(latenciesCopy, latencies)
sort.Slice(latenciesCopy, ...)

方案 2：排序一次，计算所有百分位
// 排序
sort.Slice(latencies, ...)

// 计算所有百分位
p50 := latencies[len(latencies)*50/100]
p90 := latencies[len(latencies)*90/100]
p99 := latencies[len(latencies)*99/100]

推荐：方案 2（性能更好）
```

### 测试策略

**单元测试：边界情况**

```go
测试用例 1：空结果
- 输入：[]RequestResult{}
- 验证：返回错误

测试用例 2：全部成功
- 输入：10 个成功的结果
- 验证：
  * SuccessCount = 10
  * FailureCount = 0
  * SuccessRate = 100.0

测试用例 3：部分失败
- 输入：7 成功，3 失败
- 验证：
  * SuccessCount = 7
  * FailureCount = 3
  * SuccessRate = 70.0

测试用例 4：百分位计算
- 输入：延迟 [1, 2, 3, 4, 5, 6, 7, 8, 9, 10] ms
- 验证：
  * P50 ≈ 5ms
  * P90 ≈ 9ms
  * P99 ≈ 10ms

测试用例 5：后端分布
- 输入：5 个后端，各 20 次
- 验证：
  * 每个后端 count = 20
  * 每个后端 percentage = 20.0
```

### 思考题

**Q1**：为什么需要 P50、P90、P99，不能只看平均值吗？

<details>
<summary>提示</summary>

平均值的问题：被异常值影响

示例：

```
请求延迟：[1ms, 1ms, 1ms, 1ms, 1ms, 1ms, 1ms, 1ms, 1ms, 100ms]

平均值：(1*9 + 100) / 10 = 10.9ms
P50：1ms
P90：1ms
P99：100ms
```

用户体验：

- 90% 的用户：1ms（非常快）
- 10% 的用户：100ms（很慢）
- 平均值：10.9ms（误导性）

结论：

- 平均值隐藏了长尾延迟
- P99 更能反映最差情况
- 生产环境通常关注 P99/P999

</details>

**Q2**：如何高效计算百分位数？

<details>
<summary>提示</summary>

方案对比：

**方案 1：精确计算（排序）**

- 时间复杂度：O(n log n)
- 空间复杂度：O(n)
- 适用：数据量小（< 100万）

**方案 2：近似计算（直方图）**

- 时间复杂度：O(n)
- 空间复杂度：O(buckets)
- 适用：数据量大，可接受误差

**方案 3：流式计算（t-digest）**

- 时间复杂度：O(n)
- 空间复杂度：O(1)
- 适用：流式数据，无法一次性加载

本场景：方案 1（数据量不大，需要精确）

</details>

**Q3**：QPS 应该如何计算才准确？

<details>
<summary>提示</summary>

QPS 的不同定义：

**定义 1：实际吞吐量**

```
QPS = 总请求数 / （最后一个请求时间 - 第一个请求时间）
```

含义：实际完成的 QPS

**定义 2：理论吞吐量**

```
QPS = 总请求数 / 所有请求延迟之和
```

含义：如果没有等待，理论上的 QPS

推荐：定义 1（更符合实际）

注意：

- 顺序测试：QPS ≈ 1 / 平均延迟
- 并发测试：QPS = 并发数 / 平均延迟

</details>

---

### 迭代 4 总结

**完成清单**：

- [ ] 定义 Statistics 结构
- [ ] 实现 Analyzer.Analyze()
- [ ] 实现百分位计算
- [ ] 实现 QPS 计算
- [ ] 实现分布统计
- [ ] 编写单元测试（边界情况）
- [ ] 编写性能测试（大数据量）

**文件结构**：

```
pkg/statistics/
├── statistics.go   # Statistics 结构定义
├── analyzer.go     # Analyzer 实现
└── utils.go        # 辅助函数（percentile, average, etc.）

test/statistics/
└── analyzer_test.go
```

**验收标准**：

```bash
# 单元测试
go test ./pkg/statistics/... -v

# 性能测试（1万条数据）
go test ./pkg/statistics/... -bench=. -benchmem
```

---

## 迭代 5：报告输出

### 目标

将统计数据以友好的格式输出（表格、JSON、CSV）。

### 问题场景

**场景 1：不同的使用场景**

```
场景 1：人类阅读（开发、测试）
- 需要：表格、颜色、对齐
- 示例：控制台输出

场景 2：程序解析（CI/CD、监控）
- 需要：JSON 格式
- 示例：集成到监控系统

场景 3：数据分析（Excel、Tableau）
- 需要：CSV 格式
- 示例：导入 Excel 分析
```

### 设计思路：策略模式

```go
设计目标：支持多种输出格式

type Reporter interface {
    Report(stats *Statistics) error
}

type TableReporter struct {
    writer io.Writer
}

type JSONReporter struct {
    writer io.Writer
}

type CSVReporter struct {
    writer io.Writer
}

好处：
- 易于扩展（添加新格式）
- 易于测试（Mock writer）
- 职责清晰
```

### 表格输出设计

**设计目标：清晰、美观、对齐**

```
示例输出：

========================================
  负载均衡测试报告
========================================
总请求数:     100
成功请求:     95
失败请求:     5
成功率:       95.00%
测试时长:     5.2s
QPS:          19.23

----------------------------------------
延迟统计
----------------------------------------
平均延迟:     52.3ms
最小延迟:     12.1ms
最大延迟:     156.8ms
P50:          48.2ms
P90:          89.5ms
P95:          112.3ms
P99:          145.6ms

----------------------------------------
后端分布
----------------------------------------
后端服务器              请求数    占比
localhost:8081          19        20.00%
localhost:8082          19        20.00%
localhost:8083          19        20.00%
localhost:8084          19        20.00%
localhost:8085          19        20.00%

========================================
```

**实现提示**：

```go
type TableReporter struct {
    writer io.Writer
}

func (r *TableReporter) Report(stats *Statistics) error {
    // 1. 标题
    r.printHeader()

    // 2. 基础统计
    r.printBasicStats(stats)

    // 3. 延迟统计
    r.printLatencyStats(stats)

    // 4. 后端分布
    r.printBackendDistribution(stats)

    // 5. 错误统计（如果有）
    if stats.FailureCount > 0 {
        r.printErrorDistribution(stats)
    }

    // 6. 结尾
    r.printFooter()

    return nil
}

func (r *TableReporter) printHeader() {
    fmt.Fprintln(r.writer, strings.Repeat("=", 40))
    fmt.Fprintln(r.writer, "  负载均衡测试报告")
    fmt.Fprintln(r.writer, strings.Repeat("=", 40))
}

func (r *TableReporter) printBasicStats(stats *Statistics) {
    fmt.Fprintf(r.writer, "总请求数:     %d\n", stats.TotalRequests)
    fmt.Fprintf(r.writer, "成功请求:     %d\n", stats.SuccessCount)
    fmt.Fprintf(r.writer, "失败请求:     %d\n", stats.FailureCount)
    fmt.Fprintf(r.writer, "成功率:       %.2f%%\n", stats.SuccessRate)
    fmt.Fprintf(r.writer, "测试时长:     %v\n", stats.TotalDuration)
    fmt.Fprintf(r.writer, "QPS:          %.2f\n", stats.QPS)
}

// 其他方法类似
```

**格式化技巧**：

| 技巧                 | 代码                                               | 效果                       |
| -------------------- | -------------------------------------------------- | -------------------------- |
| **对齐**       | `fmt.Fprintf(w, "%-20s %10d\n", backend, count)` | 左对齐20字符，右对齐10字符 |
| **浮点数精度** | `fmt.Fprintf(w, "%.2f%%", rate)`                 | 保留2位小数                |
| **分隔线**     | `strings.Repeat("-", 40)`                        | 生成分隔线                 |
| **时间格式**   | `duration.Round(time.Millisecond)`               | 四舍五入到毫秒             |

**颜色输出（可选）**：

```go
使用 ANSI 转义码：

const (
    ColorReset  = "\033[0m"
    ColorRed    = "\033[31m"
    ColorGreen  = "\033[32m"
    ColorYellow = "\033[33m"
)

示例：
if stats.SuccessRate >= 95 {
    fmt.Fprintf(w, "%s成功率: %.2f%%%s\n", ColorGreen, stats.SuccessRate, ColorReset)
} else if stats.SuccessRate >= 80 {
    fmt.Fprintf(w, "%s成功率: %.2f%%%s\n", ColorYellow, stats.SuccessRate, ColorReset)
} else {
    fmt.Fprintf(w, "%s成功率: %.2f%%%s\n", ColorRed, stats.SuccessRate, ColorReset)
}

注意：检测是否是 TTY（终端），避免在文件中输出转义码
```

### JSON 输出设计

**设计目标：结构化、易解析**

```json
{
  "summary": {
    "total_requests": 100,
    "success_count": 95,
    "failure_count": 5,
    "success_rate": 95.0,
    "total_duration_ms": 5200,
    "qps": 19.23
  },
  "latency": {
    "avg_ms": 52.3,
    "min_ms": 12.1,
    "max_ms": 156.8,
    "p50_ms": 48.2,
    "p90_ms": 89.5,
    "p95_ms": 112.3,
    "p99_ms": 156.8
  },
  "backend_distribution": {
    "localhost:8081": {
      "count": 19,
      "percentage": 20.0
    },
    "localhost:8082": {
      "count": 19,
      "percentage": 20.0
    }
  },
  "error_distribution": {
    "timeout": 3,
    "connection_refused": 2
  }
}
```

**实现提示**：

```go
type JSONReporter struct {
    writer io.Writer
}

// JSON 输出结构（与 Statistics 不同，更适合序列化）
type JSONReport struct {
    Summary struct {
        TotalRequests int     `json:"total_requests"`
        SuccessCount  int     `json:"success_count"`
        FailureCount  int     `json:"failure_count"`
        SuccessRate   float64 `json:"success_rate"`
        TotalDurationMS float64 `json:"total_duration_ms"`
        QPS           float64 `json:"qps"`
    } `json:"summary"`

    Latency struct {
        AvgMS float64 `json:"avg_ms"`
        MinMS float64 `json:"min_ms"`
        MaxMS float64 `json:"max_ms"`
        P50MS float64 `json:"p50_ms"`
        P90MS float64 `json:"p90_ms"`
        P95MS float64 `json:"p95_ms"`
        P99MS float64 `json:"p99_ms"`
    } `json:"latency"`

    BackendDistribution map[string]struct {
        Count      int     `json:"count"`
        Percentage float64 `json:"percentage"`
    } `json:"backend_distribution"`

    ErrorDistribution map[string]int `json:"error_distribution,omitempty"`
}

func (r *JSONReporter) Report(stats *Statistics) error {
    // 1. 转换为 JSON 结构
    report := r.convertToJSONReport(stats)

    // 2. 序列化
    encoder := json.NewEncoder(r.writer)
    encoder.SetIndent("", "  ")  // 格式化输出

    return encoder.Encode(report)
}

func (r *JSONReporter) convertToJSONReport(stats *Statistics) *JSONReport {
    report := &JSONReport{}

    // 填充字段
    report.Summary.TotalRequests = stats.TotalRequests
    report.Summary.SuccessCount = stats.SuccessCount
    report.Summary.FailureCount = stats.FailureCount
    report.Summary.SuccessRate = stats.SuccessRate
    report.Summary.TotalDurationMS = stats.TotalDuration.Seconds() * 1000
    report.Summary.QPS = stats.QPS

    // 延迟统计（转换为毫秒）
    report.Latency.AvgMS = stats.AvgLatency.Seconds() * 1000
    report.Latency.MinMS = stats.MinLatency.Seconds() * 1000
    // ...

    // 后端分布
    report.BackendDistribution = make(map[string]struct{...})
    for backend, count := range stats.BackendDistribution {
        report.BackendDistribution[backend] = struct{
            Count:      int
            Percentage: float64
        }{
            Count:      count,
            Percentage: stats.BackendPercentage[backend],
        }
    }

    // 错误分布
    if len(stats.ErrorDistribution) > 0 {
        report.ErrorDistribution = stats.ErrorDistribution
    }

    return report
}
```

### CSV 输出设计

**设计目标：适合 Excel 分析**

```csv
Metric,Value
Total Requests,100
Success Count,95
Failure Count,5
Success Rate,95.00
Total Duration (s),5.2
QPS,19.23

Latency,Value (ms)
Average,52.3
Min,12.1
Max,156.8
P50,48.2
P90,89.5
P95,112.3
P99,145.6

Backend,Count,Percentage
localhost:8081,19,20.00
localhost:8082,19,20.00
localhost:8083,19,20.00
```

**实现提示**：

```go
type CSVReporter struct {
    writer io.Writer
}

func (r *CSVReporter) Report(stats *Statistics) error {
    w := csv.NewWriter(r.writer)
    defer w.Flush()

    // 1. 基础统计
    w.Write([]string{"Metric", "Value"})
    w.Write([]string{"Total Requests", fmt.Sprint(stats.TotalRequests)})
    w.Write([]string{"Success Count", fmt.Sprint(stats.SuccessCount)})
    // ...

    // 2. 空行
    w.Write([]string{})

    // 3. 延迟统计
    w.Write([]string{"Latency", "Value (ms)"})
    w.Write([]string{"Average", fmt.Sprintf("%.2f", stats.AvgLatency.Seconds()*1000)})
    // ...

    // 4. 后端分布
    w.Write([]string{})
    w.Write([]string{"Backend", "Count", "Percentage"})
    for backend, count := range stats.BackendDistribution {
        percentage := stats.BackendPercentage[backend]
        w.Write([]string{backend, fmt.Sprint(count), fmt.Sprintf("%.2f", percentage)})
    }

    return w.Error()
}
```

### 工厂模式创建 Reporter

```go
设计目标：根据配置创建对应的 Reporter

func NewReporter(format string, writer io.Writer) (Reporter, error) {
    switch format {
    case "table":
        return &TableReporter{writer: writer}, nil
    case "json":
        return &JSONReporter{writer: writer}, nil
    case "csv":
        return &CSVReporter{writer: writer}, nil
    default:
        return nil, fmt.Errorf("不支持的输出格式: %s", format)
    }
}

使用：
reporter, err := NewReporter(config.OutputFormat, os.Stdout)
if err != nil {
    return err
}
reporter.Report(stats)
```

### 测试策略

**单元测试：输出验证**

```go
测试用例 1：表格输出
- 输入：完整的 Statistics
- 输出到 bytes.Buffer
- 验证：
  * 包含所有字段
  * 格式正确
  * 对齐正确

测试用例 2：JSON 输出
- 输出到 bytes.Buffer
- 解析 JSON
- 验证：字段值正确

测试用例 3：CSV 输出
- 输出到 bytes.Buffer
- 解析 CSV
- 验证：行数正确，值正确

示例：
func TestTableReporter(t *testing.T) {
    stats := &Statistics{
        TotalRequests: 100,
        SuccessCount:  95,
        // ...
    }

    var buf bytes.Buffer
    reporter := &TableReporter{writer: &buf}

    err := reporter.Report(stats)
    assert.NoError(t, err)

    output := buf.String()
    assert.Contains(t, output, "总请求数:     100")
    assert.Contains(t, output, "成功率:       95.00%")
}
```

### 思考题

**Q1**：为什么需要多种输出格式？

<details>
<summary>提示</summary>

不同场景的需求：

**Table（人类阅读）**：

- 场景：开发调试、手动测试
- 优点：直观、美观
- 缺点：不易解析

**JSON（程序解析）**：

- 场景：CI/CD、监控集成
- 优点：结构化、易解析
- 缺点：不够直观

**CSV（数据分析）**：

- 场景：导入 Excel、数据分析
- 优点：通用、易分析
- 缺点：缺乏层次结构

设计原则：**让工具适应使用场景，而不是反过来**

</details>

**Q2**：如何让表格输出更美观？

<details>
<summary>提示</summary>

技巧：

1. **对齐**：

   - 左对齐：标签、文本
   - 右对齐：数字
   - 使用 `%-20s` 和 `%10d`
2. **分隔线**：

   - 标题和内容之间
   - 不同部分之间
3. **颜色**（可选）：

   - 成功率 >= 95%：绿色
   - 成功率 80-95%：黄色
   - 成功率 < 80%：红色
4. **单位**：

   - 明确标注（ms, s, %）
   - 统一精度（%.2f）
5. **可读性**：

   - 避免过长的行
   - 合理的空白

</details>

---

### 迭代 5 总结

**完成清单**：

- [ ] 定义 Reporter 接口
- [ ] 实现 TableReporter
- [ ] 实现 JSONReporter
- [ ] 实现 CSVReporter
- [ ] 实现工厂函数
- [ ] 编写单元测试
- [ ] （可选）添加颜色输出

**文件结构**：

```
pkg/reporter/
├── reporter.go     # Reporter 接口
├── table.go        # TableReporter
├── json.go         # JSONReporter
├── csv.go          # CSVReporter
└── factory.go      # 工厂函数

test/reporter/
├── table_test.go
├── json_test.go
└── csv_test.go
```

**验收标准**：

```bash
# 单元测试
go test ./pkg/reporter/... -v

# 集成测试
./client --count 100 --output table
./client --count 100 --output json
./client --count 100 --output csv
```

---

## 迭代 6：并发测试

（在迭代 3 中已经实现了 ConcurrentTester，这里补充一些高级主题）

### 高级主题 1：并发安全性验证

**问题**：如何验证负载均衡器的并发安全性？

```go
测试场景：
- 10 个并发，每个发送 100 次请求
- 验证：没有数据竞争

检测方法：
go test -race ./pkg/tester/...

常见问题：
1. 共享变量未加锁
2. map 并发读写
3. slice 并发 append
```

### 高级主题 2：压力测试

**目标**：找到负载均衡器的性能上限

```go
策略：
1. 从低并发开始（10）
2. 逐步增加（20, 50, 100, 200, ...）
3. 观察延迟和成功率
4. 找到临界点（延迟突增或成功率下降）

实现提示：
for concurrent := 10; concurrent <= 1000; concurrent *= 2 {
    tester := NewConcurrentTester(..., concurrent)
    results, _ := tester.Run(ctx)
    stats := analyzer.Analyze(results)

    fmt.Printf("并发数: %d, QPS: %.2f, P99: %v\n",
        concurrent, stats.QPS, stats.P99Latency)

    // 如果成功率下降，停止
    if stats.SuccessRate < 95 {
        break
    }
}
```

---

## 迭代 7：性能基准测试

### 目标

提供标准化的性能测试报告。

### 基准测试设计

**Go Benchmark**：

```go
func BenchmarkLoadBalancer(b *testing.B) {
    // Setup
    client := setupClient()

    b.ResetTimer()  // 重置计时器

    // 运行 b.N 次
    for i := 0; i < b.N; i++ {
        _, err := client.GetBackend()
        if err != nil {
            b.Fatal(err)
        }
    }
}

运行：
go test -bench=. -benchmem

输出示例：
BenchmarkLoadBalancer-8    50000    25432 ns/op    1024 B/op    15 allocs/op

解释：
- 8: GOMAXPROCS
- 50000: 运行了 50000 次
- 25432 ns/op: 每次操作耗时
- 1024 B/op: 每次操作分配内存
- 15 allocs/op: 每次操作分配次数
```

**自定义性能测试**：

```go
func runPerformanceTest() {
    // 预热
    for i := 0; i < 100; i++ {
        client.GetBackend()
    }

    // 正式测试
    start := time.Now()
    count := 10000

    for i := 0; i < count; i++ {
        client.GetBackend()
    }

    duration := time.Since(start)
    qps := float64(count) / duration.Seconds()
    avgLatency := duration / time.Duration(count)

    fmt.Printf("QPS: %.2f, 平均延迟: %v\n", qps, avgLatency)
}
```

---

## 总结与最佳实践

### 完整的客户端架构

```
cmd/client/main.go
    ↓
├─ config.go (配置管理)
├─ pkg/client/ (HTTP 客户端)
├─ pkg/tester/ (测试器)
├─ pkg/statistics/ (统计分析)
└─ pkg/reporter/ (报告输出)
```

### 最佳实践清单

**代码质量**：

- [ ] 没有 panic，优雅处理错误
- [ ] 使用现代 API（io.ReadAll vs ioutil.ReadAll）
- [ ] 提取重复代码
- [ ] 接口驱动设计

**配置管理**：

- [ ] 支持命令行参数
- [ ] 提供合理的默认值
- [ ] 配置验证
- [ ] 友好的帮助文档

**并发安全**：

- [ ] 通过 race detector 检测
- [ ] 使用 channel 传递数据
- [ ] 使用 sync.WaitGroup 等待

**测试**：

- [ ] 单元测试（Mock）
- [ ] 集成测试（真实服务器）
- [ ] 边界测试
- [ ] 性能测试

**用户体验**：

- [ ] 清晰的错误提示
- [ ] 进度显示（可选）
- [ ] 美观的输出
- [ ] 多种输出格式

### 验收标准

**功能完整性**：

```bash
# 基本功能
./client --count 100

# 并发测试
./client --count 1000 --concurrent 10

# 不同输出格式
./client --count 100 --output json
./client --count 100 --output csv

# 自定义配置
./client --lb-url http://192.168.1.100:8187 --timeout 10s
```

**代码质量**：

```bash
# 单元测试
go test ./... -v -cover

# 竞态检测
go test ./... -race

# 静态检查
go vet ./...
golangci-lint run

# 性能测试
go test -bench=. -benchmem
```

**文档完善**：

- [ ] README.md（使用说明）
- [ ] 代码注释
- [ ] 示例

---

## 下一步

完成这个客户端后，你将掌握：

1. **企业级 Go 代码设计**

   - 清晰的架构
   - 接口驱动
   - 可测试、可扩展
2. **HTTP 客户端最佳实践**

   - 超时控制
   - 连接池
   - 错误处理
3. **并发编程**

   - Goroutine 管理
   - Channel 通信
   - 并发安全
4. **测试策略**

   - 单元测试
   - Mock 技术
   - 性能测试
5. **工程实践**

   - 配置管理
   - 日志记录
   - 用户体验

这些技能可以应用到任何 Go 项目中！

开始实现吧！遇到问题随时思考本文档中的**思考题**，它们会引导你找到答案。

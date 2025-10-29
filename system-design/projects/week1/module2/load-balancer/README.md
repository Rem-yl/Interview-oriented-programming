# 生产级负载均衡器开发指南

> 从原理到实践：问题驱动、测试先行、深度思考

---

## 文档说明

**本指南的特点**：

- ❌ **不提供**完整代码实现
- ✅ **提供**深度原理解析
- ✅ **提供**设计思路和权衡分析
- ✅ **提供**测试策略和验证方法
- ✅ **提供**生产实战经验
- ✅ **引导**你自己思考和实现

**学习方法**：

1. 先理解**为什么需要**这个功能（问题场景）
2. 再理解**底层原理**（算法、理论）
3. 对比**不同设计方案**的权衡
4. 思考**如何测试**验证正确性
5. 了解**生产环境**常见问题
6. **动手实现**，遇到问题再回来思考

---

## 目录

- [设计哲学](#设计哲学)
- [架构设计](#架构设计)
- [开发路线图](#开发路线图)
- [迭代 0：项目架构设计](#迭代-0项目架构设计)
- [迭代 1：核心抽象与接口设计](#迭代-1核心抽象与接口设计)
- [迭代 2：负载均衡算法原理](#迭代-2负载均衡算法原理)
- [迭代 3：反向代理原理](#迭代-3反向代理原理)
- [迭代 4：健康检查机制](#迭代-4健康检查机制)
- [迭代 5：熔断器模式](#迭代-5熔断器模式)
- [迭代 6：高级算法实现](#迭代-6高级算法实现)
- [迭代 7：配置管理与热更新](#迭代-7配置管理与热更新)
- [迭代 8：可观测性设计](#迭代-8可观测性设计)
- [迭代 9：性能优化方法论](#迭代-9性能优化方法论)

---

## 设计哲学

### 为什么需要设计原则？

**反面案例**：

```
假设你直接开始写代码，一个大类包含所有功能：
- LoadBalancer 类有 2000 行代码
- 包含算法逻辑、健康检查、熔断、监控...
- 想添加新算法？修改 LoadBalancer
- 想改健康检查？修改 LoadBalancer
- 一个 bug 可能影响所有功能
```

**结果**：

- 代码难以理解（新人要读 2000 行）
- 难以测试（单元测试要 mock 一大堆东西）
- 难以扩展（修改一处影响全局）
- 团队协作困难（多人同时修改同一个文件冲突不断）

### 核心设计原则

#### 1. 单一职责原则（SRP）

**原理**：一个类/模块只做一件事，只有一个变化的理由。

**思考**：

```
问：负载均衡器需要哪些职责？
答：
  - 选择后端（Algorithm）
  - 转发请求（Proxy）
  - 检查健康（HealthChecker）
  - 熔断保护（CircuitBreaker）
  - 监控统计（Metrics）

问：如果都放在一个类里会怎样？
答：
  - 修改算法会影响健康检查吗？不应该！
  - 修改代理逻辑会影响熔断器吗？不应该！
  - 那为什么要放在一起？
```

**设计决策**：

- 每个职责独立成模块
- 模块之间通过**接口**通信
- LoadBalancer 只做**编排**，不做具体实现

#### 2. 开闭原则（OCP）

**原理**：对扩展开放，对修改封闭。

**场景**：

```
需求：支持新的负载均衡算法（一致性哈希）

错误做法：
  修改 LoadBalancer 代码，增加 if-else 分支

正确做法：
  实现 Algorithm 接口，无需修改 LoadBalancer
```

**如何实现**？

- 定义稳定的**接口**
- 新功能通过**实现接口**添加
- 使用**工厂模式**创建具体实现
- LoadBalancer 只依赖接口，不依赖具体实现

**思考题**：

1. 如果有 10 种算法，使用继承好还是接口好？为什么？
2. 如何在运行时动态切换算法？
3. 如何让用户自定义算法而不修改你的代码？

#### 3. 依赖倒置原则（DIP）

**原理**：高层模块不依赖低层模块，都依赖抽象。

**反例**：

```
// 高层模块直接依赖具体实现
type LoadBalancer struct {
    roundRobin *RoundRobin  // 具体实现
}

问题：
  - 想换成 WeightedRoundRobin？修改代码
  - 想测试？必须创建真实的 RoundRobin 对象
  - 想支持多种算法？LoadBalancer 要知道所有算法
```

**正确设计**：

```
// 高层模块依赖抽象
type LoadBalancer struct {
    algorithm Algorithm  // 接口
}

好处：
  - 切换算法？传入不同的 Algorithm 实现
  - 测试？传入 Mock 对象
  - 支持新算法？实现 Algorithm 接口即可
```

**依赖关系图**：

```
高层: LoadBalancer (应用逻辑)
        ↓ 依赖
抽象层: Algorithm, HealthChecker, Proxy (接口)
        ↑ 实现
低层: RoundRobin, ActiveHealthCheck, ReverseProxy (具体实现)
```

#### 4. 组合优于继承

**为什么不用继承**？

**继承的问题**：

```
场景：支持多种功能组合

需求：
  - 负载均衡器 + 健康检查
  - 负载均衡器 + 熔断器
  - 负载均衡器 + 健康检查 + 熔断器
  - 负载均衡器 + 健康检查 + 熔断器 + 限流

继承方案：
BaseLoadBalancer
  ├─ HealthCheckLoadBalancer
  │   └─ HealthCheckCircuitBreakerLoadBalancer
  │       └─ HealthCheckCircuitBreakerRateLimitLoadBalancer
  └─ CircuitBreakerLoadBalancer
      └─ ...（组合爆炸！）
```

**组合方案**：

```
LoadBalancer 内部组合各个组件：
  - algorithm: Algorithm
  - healthChecker: HealthChecker
  - circuitBreaker: CircuitBreaker
  - rateLimiter: RateLimiter

需要什么功能，就组合什么组件
可以在运行时动态替换组件
```

**思考题**：

1. 如果后端类型需要继承（如 HTTPBackend, TCPBackend），如何设计？
2. 组合是否意味着永远不用继承？什么时候继承更合适？

---

## 架构设计

### 整体架构图

```
┌─────────────────────────────────────────────────────┐
│                  Load Balancer                      │
│  ┌──────────────────────────────────────────────┐  │
│  │  HTTP Server (Entry Point)                   │  │
│  │  - 接收客户端请求                              │  │
│  └──────────────────┬───────────────────────────┘  │
│                     │                               │
│  ┌──────────────────▼───────────────────────────┐  │
│  │  Middleware Chain                            │  │
│  │  - Logging (请求日志)                         │  │
│  │  - Metrics (监控指标)                         │  │
│  │  - Rate Limiting (限流，可选)                 │  │
│  └──────────────────┬───────────────────────────┘  │
│                     │                               │
│  ┌──────────────────▼───────────────────────────┐  │
│  │  Backend Selector                            │  │
│  │                                               │  │
│  │  ┌─────────────────┐   ┌──────────────────┐ │  │
│  │  │   Algorithm     │   │ Circuit Breaker  │ │  │
│  │  │   (选择策略)     │←──│  (熔断保护)      │ │  │
│  │  └─────────────────┘   └──────────────────┘ │  │
│  │          ↓                                   │  │
│  │  获取健康的后端列表                           │  │
│  └──────────────────┬───────────────────────────┘  │
│                     │                               │
│  ┌──────────────────▼───────────────────────────┐  │
│  │  Reverse Proxy                               │  │
│  │  - 转发请求到选中的后端                        │  │
│  │  - 管理连接池                                  │  │
│  │  - 处理超时和错误                              │  │
│  └──────────────────┬───────────────────────────┘  │
└─────────────────────┼───────────────────────────────┘
                      │
          ┌───────────┼───────────┐
          │           │           │
          ▼           ▼           ▼
      Backend1    Backend2    Backend3
          ↑           ↑           ↑
          │           │           │
┌─────────┴───────────┴───────────┴────────┐
│       Health Check Manager               │
│  - Active: 定期主动探测后端               │
│  - Passive: 监控真实请求结果              │
│  - 更新后端健康状态                        │
└──────────────────────────────────────────┘
```

### 数据流向

**请求处理流程**：

```
1. 客户端请求 → HTTP Server
   ↓
2. Middleware 处理
   - 记录请求日志
   - 开始计时（用于延迟统计）
   - 检查限流（如果启用）
   ↓
3. Backend Selector 选择后端
   - Algorithm.Select() 选择候选后端
   - 检查后端是否健康
   - 检查熔断器是否允许请求
   - 如果熔断器打开，重试其他后端
   ↓
4. Reverse Proxy 转发请求
   - 从连接池获取连接
   - 构造新的 HTTP 请求
   - 设置超时
   - 转发并等待响应
   ↓
5. 处理响应
   - 复制响应头和 Body
   - 返回给客户端
   - 记录指标（延迟、状态码）
   - 更新后端连接数
   ↓
6. 被动健康检查
   - 如果请求失败，记录失败
   - 计算失败率
   - 如果失败率过高，标记后端不健康
```

**后台健康检查流程**：

```
定时器触发 (如每 10 秒)
   ↓
Active Health Checker
   - 对每个后端发送探测请求 (GET /health)
   - 记录成功/失败
   - 连续失败 >= 阈值 → 标记不健康
   - 连续成功 >= 阈值 → 标记健康
   ↓
更新 Backend 状态
   ↓
影响后续请求的后端选择
```

### 模块职责清单

| 模块                     | 核心职责       | 输入            | 输出                        | 状态      |
| ------------------------ | -------------- | --------------- | --------------------------- | --------- |
| **Backend**        | 后端服务器抽象 | -               | URL, 健康状态, 权重, 连接数 | 有状态    |
| **Algorithm**      | 负载均衡策略   | 后端列表        | 选中的后端                  | 有/无状态 |
| **Proxy**          | 反向代理       | HTTP 请求, 后端 | HTTP 响应                   | 无状态    |
| **HealthChecker**  | 健康检查       | 后端列表        | 更新后端健康状态            | 有状态    |
| **CircuitBreaker** | 熔断保护       | 请求函数        | 执行结果或快速失败          | 有状态    |
| **Metrics**        | 监控指标       | 请求/响应事件   | Prometheus 指标             | 有状态    |
| **Config**         | 配置管理       | 配置文件        | 配置对象                    | 无状态    |

### 关键设计决策

#### 1. 为什么每个后端独立一个熔断器？

**对比方案**：

**方案 A：全局熔断器**

```
优点：实现简单
缺点：一个后端故障，所有后端都被熔断（误伤）
```

**方案 B：每个后端独立熔断器**

```
优点：精细控制，故障隔离
缺点：实现稍复杂，内存占用稍高
```

**选择**：方案 B

**原因**：

- 生产环境中，后端故障通常是个别的（如某台机器宕机）
- 不应因为一个后端故障而拒绝所有请求
- 内存开销可接受（假设 100 个后端，每个熔断器占用 < 1KB）

#### 2. 为什么需要主动+被动两种健康检查？

**单独使用主动检查的问题**：

```
- 检查间隔内，后端可能突然故障
- 有额外的网络开销
- 探测请求可能不代表实际业务流量的健康状态
```

**单独使用被动检查的问题**：

```
- 需要真实流量才能检测
- 低流量后端可能长时间不被检测
- 故障检测有延迟（需要实际请求失败）
```

**组合使用的优势**：

```
- 主动检查：定期探测，提前发现问题
- 被动检查：基于真实流量，快速响应
- 互相补充，提高检测准确性和实时性
```

#### 3. 为什么 Algorithm 接口不包含健康检查逻辑？

**错误设计**：

```
type Algorithm interface {
    Select(backends []Backend) Backend
}

// 实现中：
func (rr *RoundRobin) Select(backends []Backend) Backend {
    // 算法内部判断健康状态
    for {
        backend := backends[rr.current]
        if backend.IsHealthy() {
            return backend
        }
        rr.current++
    }
}
```

**问题**：

- 每个算法都要重复实现健康检查逻辑
- 健康检查策略变化，所有算法都要修改
- 违反单一职责原则

**正确设计**：

```
// LoadBalancer 负责过滤
func (lb *LoadBalancer) selectBackend() Backend {
    // 1. 过滤出健康的后端
    healthy := lb.filterHealthyBackends()

    // 2. 算法只关心选择逻辑
    return lb.algorithm.Select(healthy)
}
```

**好处**：

- 算法专注于选择逻辑
- 健康检查逻辑统一管理
- 易于扩展（如增加熔断器检查）

---

## 开发路线图

### 迭代策略

**原则**：

1. **从简单到复杂**：先实现基础功能，再添加高级特性
2. **可运行的 MVP**：每个迭代都产出可运行的系统
3. **测试先行**：先写测试，再写实现
4. **持续重构**：发现问题及时重构，不要留技术债

### 迭代计划

| 迭代 | 目标           | 输出                   | 可运行？      |
| ---- | -------------- | ---------------------- | ------------- |
| 0    | 项目架构设计   | 目录结构、接口定义     | ❌            |
| 1    | 核心抽象与接口 | 所有接口定义           | ✅ (可编译)   |
| 2    | 基础负载均衡   | Round Robin + Weighted | ✅ (可运行)   |
| 3    | 反向代理       | HTTP 转发              | ✅ (可测试)   |
| 4    | 健康检查       | 主动+被动检查          | ✅ (高可用)   |
| 5    | 熔断器         | 故障隔离               | ✅ (防雪崩)   |
| 6    | 高级算法       | 一致性哈希、最少连接   | ✅ (更多选择) |
| 7    | 配置管理       | 热更新                 | ✅ (可运维)   |
| 8    | 可观测性       | Metrics + Logging      | ✅ (可监控)   |
| 9    | 性能优化       | 优化瓶颈               | ✅ (生产级)   |

### 里程碑

**M1 (迭代 0-1)**: 架构设计完成

- 目录结构清晰
- 接口定义完整
- 依赖关系合理

**M2 (迭代 2-3)**: MVP 可运行

- 基础负载均衡工作
- 可以转发 HTTP 请求
- 有基本的测试覆盖

**M3 (迭代 4-5)**: 高可用特性

- 自动故障检测
- 熔断保护
- 系统稳定性提升

**M4 (迭代 6-9)**: 生产级系统

- 完善的算法支持
- 可运维（配置管理）
- 可观测（监控、日志）
- 高性能（优化完成）

---

## 迭代 0：项目架构设计

### 思考：为什么先设计架构？

**问题**：为什么不直接开始写代码？

**答案**：

```
没有架构的代码库：
├── main.go (1000 行，什么都有)
├── utils.go (500 行，杂乱的工具函数)
└── handler.go (800 行，HTTP 处理逻辑)

6 个月后：
- 新人：这代码怎么这么乱？
- 你：我也不知道为什么当时这么写...
- 老板：添加一个功能要多久？
- 你：不知道，可能会影响其他功能...
```

**好的架构的价值**：

1. **可理解性**：新人能快速理解代码结构
2. **可维护性**：知道在哪里修改代码
3. **可扩展性**：添加新功能不影响现有功能
4. **可测试性**：每个模块可独立测试
5. **团队协作**：多人可以并行开发不同模块

### 目录结构设计

#### Go 项目的标准布局

**问题**：`internal/` 和 `pkg/` 有什么区别？

**答案**：

```
internal/: Go 语言保留目录
  - 里面的包不能被外部项目导入
  - 适合放项目内部实现

pkg/: 约定俗成的目录
  - 可以被外部项目导入
  - 适合放可复用的库

示例：
  如果其他项目想用你的负载均衡算法：
    import "github.com/你/load-balancer/pkg/algorithm"  ✅
    import "github.com/你/load-balancer/internal/server"  ❌ 编译错误
```

#### 推荐的目录结构

```
load-balancer/
├── cmd/                       # 应用程序入口
│   └── lb/
│       └── main.go           # 主程序（< 100 行，只做初始化）
│
├── internal/                  # 内部实现
│   ├── server/               # HTTP 服务器
│   │   ├── server.go        # 服务器实现
│   │   ├── handler.go       # 请求处理
│   │   └── middleware.go    # 中间件
│   │
│   └── proxy/                # 反向代理
│       ├── proxy.go
│       └── transport.go     # 连接池管理
│
├── pkg/                      # 公共库（可导出）
│   ├── algorithm/           # 负载均衡算法
│   │   ├── interface.go    # Algorithm 接口
│   │   ├── factory.go      # 算法工厂
│   │   ├── roundrobin.go
│   │   ├── weighted.go
│   │   ├── leastconn.go
│   │   └── consistent.go
│   │
│   ├── backend/             # 后端抽象
│   │   ├── interface.go
│   │   └── backend.go
│   │
│   ├── healthcheck/         # 健康检查
│   │   ├── interface.go
│   │   ├── active.go
│   │   └── passive.go
│   │
│   ├── circuitbreaker/      # 熔断器
│   │   ├── breaker.go
│   │   └── state.go
│   │
│   ├── config/              # 配置管理
│   │   ├── config.go
│   │   └── watcher.go
│   │
│   └── metrics/             # 监控指标
│       ├── metrics.go
│       └── collector.go
│
├── test/                     # 测试辅助
│   ├── mock/                # Mock 对象
│   ├── integration/         # 集成测试
│   └── benchmark/           # 性能测试
│
├── configs/                  # 配置文件
│   ├── config.yaml
│   └── config.example.yaml
│
├── scripts/                  # 脚本
│   ├── build.sh
│   └── test.sh
│
├── docs/                     # 文档
│   ├── architecture.md
│   └── api.md
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### 思考题

#### Q1: 如何决定一个包应该放在 internal/ 还是 pkg/？

**判断标准**：

```
问自己：这个包有没有可能被其他项目使用？

例子：
  - algorithm 包：其他项目可能需要负载均衡算法 → pkg/
  - server 包：HTTP 服务器是本项目特定的 → internal/
  - backend 包：后端抽象是通用的 → pkg/
  - handler 包：处理器是本项目业务逻辑 → internal/
```

#### Q2: 为什么 test/ 要独立出来，而不是放在每个包里？

**对比**：

```
方案 A：测试文件放在每个包里
  pkg/algorithm/
    ├── roundrobin.go
    └── roundrobin_test.go  ✅ 单元测试

方案 B：测试文件集中放
  test/
    ├── mock/              # Mock 对象
    ├── integration/       # 集成测试
    └── benchmark/         # 性能测试

推荐：两者结合
  - 单元测试放在包里（roundrobin_test.go）
  - 集成测试、Mock、性能测试放在 test/
```

#### Q3: main.go 应该放什么内容？

**原则**：main.go 应该尽量简单，只做初始化和启动

**好的 main.go**：

```
职责：
  1. 加载配置
  2. 初始化日志
  3. 创建各个组件
  4. 启动服务器
  5. 优雅关闭

长度：< 150 行

不应该包含：
  - 业务逻辑
  - 复杂的初始化逻辑（应该封装成函数）
  - 算法实现
```

### 实践任务

#### 任务 1：创建目录结构

```bash
# 思考：为什么要先创建空目录？
# 答：确保架构清晰，避免后期重构

mkdir -p cmd/lb
mkdir -p internal/{server,proxy}
mkdir -p pkg/{algorithm,backend,healthcheck,circuitbreaker,config,metrics}
mkdir -p test/{mock,integration,benchmark}
mkdir -p configs scripts docs
```

#### 任务 2：初始化 Go 模块

```bash
go mod init github.com/你的用户名/load-balancer
```

**思考**：为什么模块名要用完整的 GitHub 路径？

- 方便其他人导入你的包
- Go 模块的命名约定
- 避免模块名冲突

#### 任务 3：创建 Makefile

**问题**：为什么需要 Makefile？

**答案**：

- 统一构建命令（不用记住复杂的 go 命令）
- 方便 CI/CD 集成
- 团队协作标准化

**Makefile 应该包含什么**？

```
常用命令：
  - make build: 构建可执行文件
  - make test: 运行测试
  - make bench: 运行性能测试
  - make lint: 代码检查
  - make run: 运行程序
  - make clean: 清理构建产物
```

**思考**：自己设计一个 Makefile，包含上述命令。

### 验证标准

- [X] 目录结构创建完成
- [X] `go mod init` 成功执行
- [X] Makefile 包含基本命令
- [X] README.md 包含项目说明
- [X] .gitignore 包含 Go 常见忽略项

---

## 迭代 1：核心抽象与接口设计

### 为什么要先设计接口？

**传统开发流程**：

```
1. 直接写实现
2. 发现需要扩展
3. 修改已有代码
4. 影响其他模块
5. 引入 bug
6. 重复 1-5
```

**接口驱动开发**：

```
1. 定义接口（契约）
2. 基于接口编写调用代码
3. 实现接口
4. 测试（可以 mock 接口）
5. 扩展时只需实现新的接口实现
```

**优势**：

- **并行开发**：团队成员基于接口并行工作
- **依赖倒置**：高层不依赖低层具体实现
- **易于测试**：可以 mock 接口
- **面向契约**：接口即文档

### Backend 接口设计

#### 设计思路

**问题 1**：Backend 应该提供哪些能力？

**分析**：

```
谁会使用 Backend？
  - Algorithm：需要知道后端的权重、连接数、健康状态
  - HealthChecker：需要更新健康状态
  - Proxy：需要获取 URL，管理连接数
  - Metrics：需要统计信息

Backend 需要什么信息？
  - 基础信息：URL、名称
  - 健康状态：是否健康（bool）
  - 权重：用于加权算法（int）
  - 连接数：用于最少连接算法（int64）
  - 统计：总请求数、失败数
```

**问题 2**：如何保证并发安全？

**场景分析**：

```
多个 goroutine 可能同时：
  - 读取健康状态（Algorithm.Select）
  - 修改健康状态（HealthChecker.Check）
  - 增加连接数（Proxy.Forward 开始）
  - 减少连接数（Proxy.Forward 结束）

如何保护？
  方案 1：全部加锁（简单但性能差）
  方案 2：读写锁 + 原子操作（复杂但高性能）
  方案 3：只用原子操作（某些场景可行）
```

**设计决策**：

```
健康状态：使用 RWMutex（读多写少）
连接数：使用 atomic（频繁修改）
统计信息：使用 atomic（只增不减）
```

#### 思考题

**Q1：Backend 接口应该包含业务逻辑吗？**

错误设计：

```
type Backend interface {
    GetURL() string
    SendHealthCheck() error  // ❌ 业务逻辑
    ForwardRequest() error   // ❌ 业务逻辑
}

问题：
  - Backend 职责过重
  - 难以替换实现
  - 测试困难
```

正确设计：

```
type Backend interface {
    // 只包含状态查询和简单修改
    GetURL() string
    IsHealthy() bool
    SetHealthy(bool)
    GetWeight() int
    GetActiveConnections() int64
    IncrementConnections()
    DecrementConnections()
}

业务逻辑由其他模块负责：
  - HealthChecker 负责健康检查逻辑
  - Proxy 负责请求转发逻辑
```

**Q2：是否应该暴露 Backend 的内部状态修改方法？**

对比：

```
方案 A：暴露 SetHealthy()
  优点：灵活
  缺点：任何人都可以修改健康状态，容易误用

方案 B：只通过 HealthChecker 修改
  优点：职责清晰，不易误用
  缺点：需要额外的权限控制

选择：方案 A，但在文档中说明只有 HealthChecker 应该调用
原因：
  - Go 没有 friend 类
  - internal 包已经限制了访问范围
  - 实用主义：相信开发者
```

**Q3：如何设计 Backend 的创建方式？**

思考：

```
方案 1：构造函数
  backend := NewBackend("http://localhost:8080", 5)

方案 2：Builder 模式
  backend := NewBackendBuilder().
      URL("http://localhost:8080").
      Weight(5).
      Build()

方案 3：从配置创建
  backend := backend.FromConfig(BackendConfig{
      URL: "http://localhost:8080",
      Weight: 5,
  })

选择：方案 1 + 方案 3
  - 方案 1：简单场景
  - 方案 3：从配置文件加载
  - 方案 2：参数太多时使用
```

### Algorithm 接口设计

#### 设计思路

**核心问题**：算法应该有状态吗？

**无状态算法（如 Random）**：

```
每次 Select() 调用互不影响
优点：线程安全，无需加锁
缺点：无法实现 Round Robin
```

**有状态算法（如 Round Robin）**：

```
需要记住上次选择的位置
优点：可以实现轮询
缺点：需要处理并发访问
```

**设计决策**：

```
Algorithm 接口本身不限制有无状态
由具体实现决定是否需要状态
如果有状态，实现必须保证并发安全
```

#### 接口定义的权衡

**方案 1：简单接口**

```
type Algorithm interface {
    Select(backends []Backend) (Backend, error)
    Name() string
}

优点：简单清晰
缺点：无法支持基于请求内容的选择（如一致性哈希）
```

**方案 2：带上下文的接口**

```
type Algorithm interface {
    Select(backends []Backend) (Backend, error)
    SelectWithContext(backends []Backend, ctx SelectContext) (Backend, error)
    Name() string
}

优点：支持更多场景
缺点：所有算法都要实现 SelectWithContext，即使不需要
```

**方案 3：组合接口**

```
type Algorithm interface {
    Select(backends []Backend) (Backend, error)
    Name() string
}

type ContextAwareAlgorithm interface {
    Algorithm
    SelectWithContext(backends []Backend, ctx SelectContext) (Backend, error)
}

优点：
  - 基础算法只需实现 Select
  - 需要上下文的算法实现 ContextAwareAlgorithm
  - 调用方可以类型断言检查是否支持上下文

缺点：稍微复杂
```

**选择**：方案 3（接口隔离原则）

#### 思考题

**Q1：算法是否应该过滤不健康的后端？**

```
方案 A：算法内部过滤
func (rr *RoundRobin) Select(backends []Backend) Backend {
    for ... {
        if backend.IsHealthy() {
            return backend
        }
    }
}

方案 B：调用方过滤
func (lb *LoadBalancer) selectBackend() Backend {
    healthy := filterHealthy(lb.backends)
    return lb.algorithm.Select(healthy)
}

思考：
  1. 哪种方案更符合单一职责原则？
  2. 如果有 10 种算法，哪种方案重复代码更少？
  3. 如果要增加熔断器检查，哪种方案更容易扩展？
```

**Q2：算法是否应该修改 Backend 状态（如连接数）？**

```
场景：最少连接算法选择后，是否应该立即增加连接数？

方案 A：算法修改状态
  - 优点：逻辑集中
  - 缺点：算法职责过重，如果选择后没有真正连接怎么办？

方案 B：Proxy 修改状态
  - 优点：职责清晰，只有真正连接时才修改
  - 缺点：算法和 Proxy 需要协调

答案：方案 B（谁真正使用资源，谁负责管理）
```

**Q3：如何支持自定义算法？**

```
需求：用户想实现自己的负载均衡算法

设计 1：硬编码所有算法
  - ❌ 无法扩展

设计 2：接口 + 注册机制
  // 用户实现
  type MyAlgorithm struct {}
  func (m *MyAlgorithm) Select(...) Backend { ... }

  // 注册
  algorithm.Register("my_algo", func() Algorithm {
      return &MyAlgorithm{}
  })

  // 使用
  algo := algorithm.Create("my_algo")

思考：
  1. 注册机制如何实现？
  2. 是否需要支持运行时注册？
  3. 如何处理重名问题？
```

### HealthChecker 接口设计

#### 设计思路

**生命周期管理**：

```
健康检查是后台任务，需要：
  1. 启动检查（Start）
  2. 停止检查（Stop）
  3. 优雅关闭（等待正在进行的检查完成）

如何实现？
  - 使用 context.Context 控制生命周期
  - 使用 channel 通信
  - 使用 WaitGroup 等待 goroutine 结束
```

**接口定义的考虑**：

```
问题：应该传入什么参数？

方案 1：Start() 不带参数
  - Backend 列表在创建 HealthChecker 时传入
  - 优点：简单
  - 缺点：无法动态添加/删除后端

方案 2：Start(backends []Backend)
  - 每次启动时传入后端列表
  - 优点：灵活
  - 缺点：每次调用 Start 都重新开始检查

方案 3：AddBackend/RemoveBackend
  - 动态管理后端
  - 优点：最灵活
  - 缺点：实现复杂

选择：方案 1 + 提供 AddBackend/RemoveBackend 方法
```

#### 思考题

**Q1：主动检查和被动检查应该是同一个接口吗？**

```
对比：

方案 A：统一接口
type HealthChecker interface {
    Start()
    Stop()
    Check(backend Backend) error
}

问题：
  - 主动检查自己发起请求
  - 被动检查监听请求结果
  - 行为差异太大，统一接口合适吗？

方案 B：分开接口
type ActiveHealthChecker interface {
    Start()
    Stop()
}

type PassiveHealthChecker interface {
    RecordResult(backend Backend, success bool)
}

思考：
  1. 哪种方案更符合接口隔离原则？
  2. LoadBalancer 应该持有一个还是两个 HealthChecker？
  3. 如何实现组合策略（主动+被动）？
```

**Q2：健康检查失败，应该立即标记后端不健康吗？**

```
场景：网络抖动导致偶尔失败

方案 A：立即标记
  - 优点：响应快
  - 缺点：误判率高，状态抖动

方案 B：连续失败 N 次才标记
  - 优点：稳定
  - 缺点：响应慢

方案 C：基于失败率
  - 最近 10 次检查，失败率 > 50% 才标记
  - 优点：平衡响应速度和稳定性
  - 缺点：实现复杂

思考：
  1. 如何在接口中体现这种策略？
  2. 是配置参数还是不同的实现？
  3. 如何测试这种行为？
```

### Proxy 接口设计

#### 设计思路

**核心问题**：Proxy 的职责是什么？

```
明确的职责：
  ✅ 转发 HTTP 请求到后端
  ✅ 处理连接、超时
  ✅ 复制请求和响应

不应该包含的职责：
  ❌ 选择后端（Algorithm 的职责）
  ❌ 判断后端是否健康（HealthChecker 的职责）
  ❌ 熔断逻辑（CircuitBreaker 的职责）
  ❌ 记录监控指标（Metrics 的职责，但可以作为参数传入）
```

**接口设计**：

```
最简单的接口：
type Proxy interface {
    Forward(w http.ResponseWriter, r *http.Request, backend Backend) error
}

思考：
  1. 需要返回更多信息吗（如响应状态码、延迟）？
  2. 如何集成监控？
  3. 如何处理请求的生命周期（如超时）？
```

#### 思考题

**Q1：Proxy 应该管理连接池吗？**

```
HTTP 连接池：
  - 避免每次请求都建立 TCP 连接
  - 复用连接提高性能

问题：
  - 连接池应该全局唯一还是每个后端独立？
  - 谁负责配置连接池参数？
  - 如何监控连接池状态？

思考：
  1. 如果全局连接池，如何限制每个后端的连接数？
  2. 如果每个后端独立连接池，如何避免总连接数过多？
  3. 连接池配置应该在 Proxy 还是 Backend？
```

**Q2：如何处理请求失败？**

```
失败类型：
  1. 连接失败（后端不可达）
  2. 超时（后端响应慢）
  3. 5xx 错误（后端内部错误）
  4. 4xx 错误（客户端错误）

问题：
  - 哪些错误应该重试？
  - 重试几次？
  - 是否应该尝试其他后端？
  - 如何避免重试导致的雪崩？

思考：
  1. 重试逻辑应该在 Proxy 还是 LoadBalancer？
  2. 如何区分可重试和不可重试的错误？
  3. 如何实现幂等性检查（避免重复提交）？
```

### 模块依赖关系

**依赖层次**：

```
高层
  ↓
LoadBalancer (编排层)
  ↓ 依赖
Algorithm, HealthChecker, CircuitBreaker, Proxy (业务逻辑层)
  ↓ 依赖
Backend (领域模型层)
  ↓
低层
```

**依赖原则**：

1. **高层不依赖低层**：LoadBalancer 依赖接口，不依赖具体实现
2. **同层不互相依赖**：Algorithm 不依赖 HealthChecker
3. **单向依赖**：下层不知道上层的存在

**思考题**：

**Q1：如果 Algorithm 需要使用 Metrics 记录选择结果，如何设计？**

```
错误设计：
type RoundRobin struct {
    metrics *Metrics  // ❌ 算法依赖 Metrics
}

问题：
  - 算法层依赖更高层的 Metrics
  - 难以测试（必须创建 Metrics 对象）

正确设计：
方案 A：依赖注入
  - Algorithm.Select() 返回选择结果
  - LoadBalancer 记录 Metrics

方案 B：事件回调
  - Algorithm 发出事件
  - LoadBalancer 监听事件并记录

方案 C：传递 Metrics 接口
  - 算法依赖 Metrics 接口（不是具体实现）
  - 测试时可以传入 Mock

思考：哪种方案最合适？
```

### 实践任务

#### 任务 1：定义所有接口

在对应的文件中定义接口：

- `pkg/backend/interface.go`: Backend 接口
- `pkg/algorithm/interface.go`: Algorithm 接口
- `pkg/healthcheck/interface.go`: HealthChecker 接口
- `internal/proxy/interface.go`: Proxy 接口

**要求**：

1. 只定义接口，不实现
2. 每个接口写清楚注释（职责、参数、返回值）
3. 思考并发安全性

#### 任务 2：画出依赖关系图

使用 Mermaid 或手绘，画出：

1. 模块依赖关系
2. 数据流向
3. 接口调用关系

#### 任务 3：编写接口文档

为每个接口编写文档，说明：

1. 接口的职责
2. 方法的语义
3. 并发安全性保证
4. 使用示例（伪代码）

### 验证标准

- [ ] 所有接口定义完成
- [ ] 接口注释完整
- [ ] 依赖关系合理（符合 DIP）
- [ ] 代码可以编译通过（go build）
- [ ] 没有循环依赖

---

## 迭代 2：负载均衡算法原理

### 为什么需要负载均衡算法？

**问题场景**：

```
有 3 个后端服务器：
- 后端 A：8 核 CPU，16GB 内存
- 后端 B：4 核 CPU，8GB 内存
- 后端 C：2 核 CPU，4GB 内存

如果平均分配流量（每个 33%）：
  → 后端 C 过载
  → 后端 A 资源浪费
  → 整体性能不是最优

需要：根据后端能力分配流量
```

### Round Robin（轮询）算法

#### 原理

**最简单的负载均衡策略**：

```
依次选择每个后端，循环往复

示例：
后端列表：[A, B, C]

请求 1 → A
请求 2 → B
请求 3 → C
请求 4 → A  (循环)
请求 5 → B
请求 6 → C
```

#### 核心问题

**问题 1：如何记住上次选择的位置？**

```
需要状态：
  - current: 当前位置（int）

操作：
  1. 获取 current 位置的后端
  2. current = (current + 1) % len(backends)

思考：
  1. current 应该是 0-based 还是 1-based？
  2. 如何处理空列表？
  3. 如果后端列表变化了怎么办？
```

**问题 2：并发安全**

**场景分析**：

```
时刻 T1:
  Goroutine 1: 读取 current = 0
  Goroutine 2: 读取 current = 0

时刻 T2:
  Goroutine 1: current = (0 + 1) % 3 = 1
  Goroutine 2: current = (0 + 1) % 3 = 1

结果：
  两个请求都选择了 Backend[1]，跳过了 Backend[0]
```

**解决方案对比**：

| 方案     | 实现         | 性能 | 复杂度 | 正确性    |
| -------- | ------------ | ---- | ------ | --------- |
| 无锁     | 不加锁       | 最高 | 最低   | ❌ 有竞态 |
| 互斥锁   | sync.Mutex   | 中   | 中     | ✅ 正确   |
| 读写锁   | sync.RWMutex | 中   | 中     | ✅ 正确   |
| 原子操作 | atomic       | 高   | 中     | ✅ 正确   |

**思考题**：

```
Q: 为什么读写锁在这里没有优势？
A: Select() 操作既读又写 current，无法利用读写分离

Q: 原子操作 vs 互斥锁，如何选择？
A:
  - 互斥锁：实现简单，代码清晰
  - 原子操作：性能更好，但要注意计数器溢出

建议：先用互斥锁实现，性能测试后再考虑优化
```

**问题 3：跳过不健康的后端**

**场景**：

```
后端列表：[A (健康), B (不健康), C (健康)]
current = 1（指向 B）

期望：跳过 B，选择 C

挑战：
  1. 如果循环查找，如何避免死循环？
  2. 如果所有后端都不健康，应该返回什么？
  3. 查找健康后端的时间复杂度是多少？
```

**设计思路**：

```
方案 1：有限次重试
  tried = 0
  while tried < len(backends):
      backend = backends[current]
      current = (current + 1) % len(backends)
      if backend.IsHealthy():
          return backend
      tried++
  return error("no healthy backend")

时间复杂度：O(n)，n = 后端数量

方案 2：预过滤
  healthy = filter(backends, isHealthy)
  if len(healthy) == 0:
      return error
  return roundRobin(healthy)

思考：
  1. 哪种方案更好？
  2. 如何测试"所有后端不健康"的情况？
```

#### 测试策略

**单元测试清单**：

**Test 1：基础轮询**

```
测试目标：验证轮询顺序正确

测试步骤：
  1. 创建 3 个 Backend：A, B, C
  2. 连续调用 Select() 6 次
  3. 验证顺序：A, B, C, A, B, C

边界条件：
  - 只有 1 个后端
  - 后端列表为空
```

**Test 2：并发安全**

```
测试目标：验证无竞态条件

测试步骤：
  1. 创建 3 个 Backend
  2. 启动 100 个 goroutine
  3. 每个 goroutine 调用 Select() 1000 次
  4. 统计每个后端被选择的次数

验证：
  - 使用 go test -race 检测竞态
  - 每个后端被选择的次数应该接近 33333

如何验证？
  分布应该均匀：
    A: 33000 - 34000 ✅
    B: 33000 - 34000 ✅
    C: 33000 - 34000 ✅

  不均匀的情况：
    A: 50000 ❌ 有问题
    B: 30000 ✅
    C: 20000 ❌ 有问题
```

**Test 3：跳过不健康后端**

```
测试目标：验证只选择健康后端

测试步骤：
  1. 创建 3 个 Backend：A (健康), B (不健康), C (健康)
  2. 调用 Select() 10 次
  3. 验证：
     - 从不选择 B
     - A 和 C 交替出现

边界条件：
  - 所有后端不健康（应该返回错误）
  - 只有一个后端健康
  - 健康后端不连续（如 A, C, E 健康，B, D 不健康）
```

**Test 4：后端列表变化**

```
场景：运行时添加/删除后端

测试步骤：
  1. 初始：A, B, C
  2. 选择几次，观察顺序
  3. 添加 D
  4. 继续选择，验证 D 被包含
  5. 删除 B
  6. 继续选择，验证 B 不再出现

问题：
  1. 添加后端后，从哪个位置开始选择？
  2. 删除后端后，current 指针如何处理？
  3. 如何保证线程安全？
```

**性能测试**：

```
Benchmark：
  - 测试 Select() 的 QPS
  - 测试不同后端数量（10, 100, 1000）的性能
  - 测试不同并发数（1, 10, 100）的性能

期望结果：
  - QPS > 100万（简单操作，应该很快）
  - 时间复杂度 O(1) 或 O(n)，n = 后端数
```

### Weighted Round Robin（加权轮询）

#### 为什么需要加权？

**问题场景**：

```
3 台服务器：
  Server A: 高配置，处理能力强
  Server B: 中配置
  Server C: 低配置

简单 Round Robin：
  → 每台分配 33% 流量
  → Server C 过载
  → Server A 浪费

加权分配：
  Server A: 50% 流量
  Server B: 30% 流量
  Server C: 20% 流量
  → 根据能力分配流量
```

#### 朴素实现：扩展数组

**思路**：

```
权重：A=5, B=1, C=1

扩展数组：
  [A, A, A, A, A, B, C]

然后用简单 Round Robin

结果：
  7 个请求的分配：A, A, A, A, A, B, C
  比例：5:1:1 ✅
```

**问题分析**：

```
❌ 内存浪费：
  如果权重是 A=1000, B=1
  需要 1001 个元素

❌ 不够平滑：
  5 个 A 连续出现
  用户可能感觉负载不均

❌ 权重更新困难：
  修改权重需要重建数组
```

#### Smooth Weighted Round Robin（平滑加权轮询）

**原理**：

**核心思想**：

```
每个后端维护两个权重：
  1. weight：配置的固定权重
  2. current_weight：当前动态权重

算法步骤（每次选择）：
  1. 遍历所有后端：current_weight += weight
  2. 选择 current_weight 最大的后端
  3. 被选中的后端：current_weight -= total_weight
```

**示例**（A=5, B=1, C=1）：

```
初始状态：
  A: current=0, weight=5
  B: current=0, weight=1
  C: current=0, weight=1
  total = 7

第 1 轮：
  步骤1 (累加): A:5, B:1, C:1
  步骤2 (选择): 选择 A（最大）
  步骤3 (减去): A:5-7=-2, B:1, C:1
  → 选择 A

第 2 轮：
  步骤1: A:-2+5=3, B:1+1=2, C:1+1=2
  步骤2: 选择 A（最大）
  步骤3: A:3-7=-4, B:2, C:2
  → 选择 A

第 3 轮：
  步骤1: A:-4+5=1, B:2+1=3, C:2+1=3
  步骤2: 选择 B（最大，平局时选第一个）
  步骤3: A:1, B:3-7=-4, C:3
  → 选择 B

第 4 轮：
  步骤1: A:1+5=6, B:-4+1=-3, C:3+1=4
  步骤2: 选择 A
  步骤3: A:6-7=-1, B:-3, C:4
  → 选择 A

第 5 轮：
  步骤1: A:-1+5=4, B:-3+1=-2, C:4+1=5
  步骤2: 选择 C
  步骤3: A:4, B:-2, C:5-7=-2
  → 选择 C

第 6 轮：
  步骤1: A:4+5=9, B:-2+1=-1, C:-2+1=-1
  步骤2: 选择 A
  步骤3: A:9-7=2, B:-1, C:-1
  → 选择 A

第 7 轮：
  步骤1: A:2+5=7, B:-1+1=0, C:-1+1=0
  步骤2: 选择 A
  步骤3: A:7-7=0, B:0, C:0
  → 选择 A

结果：A, A, B, A, C, A, A
比例：5:1:1 ✅
分布：更均匀（不是连续 5 个 A）
```

**数学证明**（可选）：

**为什么比例是正确的？**

```
定理：在 total_weight 轮后，每个后端被选择的次数 = 其权重

证明思路：
  1. 每轮选择，所有后端的 current_weight 总和增加 total_weight
  2. 被选中的后端减去 total_weight
  3. 因此，每 total_weight 轮后，总和不变
  4. 由于选择最大的，权重大的后端被选中更多次
  5. 长期来看，选择次数与权重成正比

（完整证明需要数学归纳法，这里省略）
```

**为什么是平滑的？**

```
观察：
  - A 的权重虽然是 5，但不是连续出现 5 次
  - A, A, B, A, C, A, A（A 分散在各处）

原因：
  - current_weight 动态变化
  - 被选中后立即减去 total_weight，给其他后端机会
  - 权重小的后端也有机会累积到最大

好处：
  - 用户感觉更均匀
  - 避免某个后端短时间内过载
```

#### 实现挑战

**挑战 1：如何存储 current_weight？**

```
方案 1：在 Backend 中存储
  问题：Backend 是共享的，多个算法实例会互相干扰

方案 2：在 Algorithm 中维护 map[Backend]int
  问题：如何使用 Backend 作为 map 的 key？
    - Backend 是接口，能做 key 吗？
    - 如何保证 Backend 的唯一性？

方案 3：使用 Backend.GetName() 作为 key
  map[string]int
  问题：Name 必须唯一

思考：哪种方案最合适？
```

**挑战 2：并发安全**

```
问题：
  current_weight 是算法的状态，会被修改

解决：
  - 使用 sync.Mutex 保护整个 Select() 方法
  - 不能用原子操作（需要选择最大值，不是简单的加减）

思考：
  1. 能否用读写锁优化？
  2. 能否无锁实现？
```

**挑战 3：权重为 0 的后端**

```
场景：
  A: weight=5
  B: weight=0  （想临时下线，但不删除）
  C: weight=1

问题：
  B 的 current_weight 始终是 0，永远不会被选中 ✅

思考：
  1. 权重为 0 和不健康有什么区别？
  2. 是否应该允许权重为 0？
  3. 权重为负数呢？
```

#### 测试策略

**Test 1：比例正确性**

```
测试目标：验证长期选择比例符合权重

测试步骤：
  1. 创建 Backend：A=5, B=1, C=1
  2. 调用 Select() 700 次（100 个周期）
  3. 统计每个后端被选择的次数

验证：
  A: 约 500 次 (500/700 ≈ 71.4%)
  B: 约 100 次 (100/700 ≈ 14.3%)
  C: 约 100 次 (100/700 ≈ 14.3%)

允许误差：± 5%
```

**Test 2：平滑性**

```
测试目标：验证分布平滑

测试步骤：
  1. 创建 Backend：A=5, B=1, C=1
  2. 调用 Select() 7 次
  3. 记录顺序

验证：
  不应该出现：A, A, A, A, A, B, C（连续）
  应该类似：A, A, B, A, C, A, A（分散）

如何量化平滑性？
  计算最大连续出现次数：
    朴素算法：max_consecutive(A) = 5 ❌
    平滑算法：max_consecutive(A) = 2 ✅
```

**Test 3：权重更新**

```
场景：运行时修改权重

测试步骤：
  1. 初始权重：A=5, B=1, C=1
  2. 选择 100 次，验证比例
  3. 修改权重：A=1, B=5, C=1
  4. 再选择 100 次，验证新比例

问题：
  1. 修改权重后，current_weight 如何处理？
     - 方案 A：重置为 0
     - 方案 B：保持不变
     - 方案 C：按比例调整
  2. 哪种方案过渡最平滑？
```

**Test 4：并发压力测试**

```
测试目标：验证高并发下的正确性和性能

测试步骤：
  1. 100 个 goroutine
  2. 每个调用 Select() 10000 次
  3. 统计总选择次数和比例

验证：
  - go test -race 无竞态
  - 比例正确
  - 性能可接受（QPS > 10万）
```

#### 思考题

**Q1：如何实现加权最少连接算法？**

```
结合两个因素：
  1. 连接数（越少越好）
  2. 权重（越大越好）

公式：
  负载 = 连接数 / 权重

选择负载最小的后端

示例：
  A: 连接数=10, 权重=5 → 负载=2.0
  B: 连接数=5, 权重=1  → 负载=5.0
  C: 连接数=3, 权重=1  → 负载=3.0

选择：A（负载最低）

思考：
  1. 如何实现这个算法？
  2. 需要状态吗？
  3. 时间复杂度是多少？
  4. 如何测试？
```

**Q2：如何处理权重全部为 0 的情况？**

```
场景：所有后端的权重都是 0

问题：total_weight = 0，除以 0 错误

解决方案：
  1. 返回错误
  2. 降级为简单 Round Robin
  3. 拒绝权重为 0 的配置

思考：哪种方案更合理？
```

**Q3：权重范围应该限制吗？**

```
问题：
  - 权重可以是负数吗？
  - 权重可以是小数吗？
  - 权重的最大值应该是多少？

思考：
  1. 负数权重有意义吗？
  2. 如果支持小数，如何处理精度问题？
  3. 如果权重是 [0.1, 0.2, 0.3]，如何处理？
     （提示：可以归一化或缩放）
```

### 实践任务

#### 任务 1：实现 Round Robin

1. 实现 `RoundRobin` 结构体
2. 实现 `Select()` 方法
3. 实现并发安全保护
4. 实现跳过不健康后端的逻辑

**验证**：

- 编写单元测试
- 运行 `go test -race`
- 编写性能测试

#### 任务 2：实现 Weighted Round Robin

1. 实现 `WeightedRoundRobin` 结构体
2. 实现平滑加权算法
3. 处理边界情况（空列表、权重为 0）

**验证**：

- 测试比例正确性
- 测试平滑性（手动验证前几轮的顺序）
- 并发测试

#### 任务 3：性能对比

编写 Benchmark 对比：

1. Round Robin 性能
2. Weighted Round Robin 性能
3. 不同后端数量的影响
4. 不同并发数的影响

**思考**：

- 哪个算法更快？为什么？
- 如何优化性能？

### 验证标准

- [ ] Round Robin 轮询顺序正确
- [ ] Weighted Round Robin 比例正确
- [ ] 平滑性满足要求
- [ ] 并发测试通过（无竞态）
- [ ] 性能满足要求（QPS > 10万）
- [ ] 单元测试覆盖率 > 80%

---

## 迭代 3：反向代理原理

### 为什么需要反向代理？

**正向代理 vs 反向代理**：

```
正向代理：
  客户端 → 代理服务器 → 互联网
  - 客户端知道代理的存在
  - 隐藏客户端身份
  - 例子：VPN

反向代理：
  客户端 → 代理服务器 → 后端服务器
  - 客户端不知道后端的存在
  - 隐藏后端结构
  - 例子：负载均衡器、CDN
```

**反向代理的作用**：

```
1. 负载均衡：分发请求到多个后端
2. 缓存：缓存静态内容
3. SSL 终止：在代理层处理 HTTPS
4. 安全：隐藏后端服务器信息
5. 压缩：压缩响应
6. 统一入口：提供统一的 API 地址
```

### HTTP 反向代理原理

#### 核心流程

```
1. 客户端请求
   ↓
2. 负载均衡器接收请求
   ↓
3. 构造新的 HTTP 请求
   - 目标：后端服务器
   - 复制原始请求的内容
   ↓
4. 发送请求到后端
   ↓
5. 接收后端响应
   ↓
6. 构造响应返回给客户端
   - 复制后端响应的内容
```

#### 设计挑战

**挑战 1：如何构造后端请求？**

**需要复制的内容**：

```
1. HTTP 方法：GET, POST, PUT, DELETE, etc.
2. URL 路径：/api/users
3. Query 参数：?page=1&size=10
4. 请求头：
   - Host
   - User-Agent
   - Content-Type
   - Authorization
   - Cookie
   - ...
5. 请求体：POST/PUT 的 Body
```

**特殊处理的头**：

```
1. Host：
   - 原始值：client.example.com
   - 应该改为：backend.internal

2. X-Forwarded-For：
   - 记录客户端真实 IP
   - 格式：client_ip, proxy1_ip, proxy2_ip

3. X-Forwarded-Proto：
   - 原始协议：http 或 https

4. X-Forwarded-Host：
   - 原始 Host

为什么需要这些头？
  - 后端需要知道客户端真实 IP（日志、安全）
  - 后端需要知道原始协议（生成正确的链接）
```

**问题**：

```
Q1: 应该复制所有请求头吗？
A: 不应该！某些头是 hop-by-hop 的（只对当前连接有效）

需要删除的头：
  - Connection
  - Keep-Alive
  - Proxy-Authenticate
  - Proxy-Authorization
  - TE
  - Trailers
  - Transfer-Encoding (有时需要)
  - Upgrade

为什么？
  这些头是代理层的，不应该转发给后端
```

**Q2: 如何处理请求体？**

```
挑战：
  - 请求体可能很大（文件上传）
  - 请求体是流式的（io.Reader）
  - 只能读一次

方案 1：全部读入内存
  - 简单
  - 但大文件会 OOM

方案 2：流式转发
  - 边读边转发
  - 内存占用小
  - 但无法重试（已经读了）

方案 3：使用 GetBody（如果有）
  - HTTP 请求可能有 GetBody 方法
  - 可以重新获取 Body
  - 支持重试

思考：选择哪种方案？
```

**挑战 2：连接管理**

**为什么需要连接池？**

```
场景：每秒 1000 个请求

不使用连接池：
  - 每个请求建立新的 TCP 连接
  - TCP 三次握手：1-2ms
  - 总耗时：1000-2000ms 仅用于建立连接
  - 系统开销大（socket 资源）

使用连接池：
  - 复用现有连接
  - 减少握手开销
  - HTTP/1.1 Keep-Alive
```

**连接池参数**：

```
关键参数：
  1. MaxIdleConns: 最大空闲连接数
     - 太小：连接不够用，频繁建立新连接
     - 太大：占用系统资源

  2. MaxIdleConnsPerHost: 每个后端的最大空闲连接
     - 控制每个后端的连接数
     - 避免某个后端连接过多

  3. IdleConnTimeout: 空闲连接超时时间
     - 太短：频繁关闭再建立连接
     - 太长：占用资源

  4. MaxConnsPerHost: 每个后端的最大连接数（总数）
     - 限制同时连接数
     - 避免打垮后端

推荐配置：
  MaxIdleConns: 100
  MaxIdleConnsPerHost: 10
  IdleConnTimeout: 90s
  MaxConnsPerHost: 50
```

**思考题**：

```
Q1: 如果有 10 个后端，MaxIdleConnsPerHost=10，理论上最多有多少个连接？
A: 10 * 10 = 100 个

Q2: 如果后端突然下线，连接池中的连接会怎样？
A:
  - 下次使用时会发现连接失败
  - 自动从池中移除
  - 建立新连接

Q3: 连接池是否应该预热（提前建立连接）？
A:
  - 优点：首次请求更快
  - 缺点：可能浪费资源（如果流量低）
  - 通常不预热，按需建立
```

**挑战 3：超时控制**

**多种超时**：

```
1. Dial Timeout (连接超时):
   - 建立 TCP 连接的时间
   - 推荐：5-10s

2. TLS Handshake Timeout:
   - HTTPS 握手时间
   - 推荐：10s

3. Response Header Timeout:
   - 读取响应头的时间
   - 推荐：10-30s
   - 用于快速检测后端无响应

4. Idle Timeout:
   - 空闲连接超时
   - 推荐：90s

5. Total Request Timeout:
   - 整个请求的超时（包括读写 Body）
   - 推荐：30-60s
   - 根据业务场景调整
```

**超时的优先级**：

```
问题：如果设置了多个超时，哪个先生效？

答案：取最小值

示例：
  DialTimeout: 5s
  ResponseHeaderTimeout: 10s
  TotalRequestTimeout: 30s

  如果连接就花了 5s，触发 DialTimeout
  如果连接 1s，等响应头 11s，触发 ResponseHeaderTimeout
  如果总时间 31s，触发 TotalRequestTimeout
```

**思考题**：

```
Q1: 如果后端响应慢，应该设置长超时还是短超时？
A: 两难：
  - 长超时：请求堆积，资源耗尽
  - 短超时：大量超时，后端更慢

  解决：
    - 设置合理的超时（如 P99 延迟 * 2）
    - 使用熔断器保护
    - 监控超时率

Q2: 超时后，后端还在处理请求怎么办？
A:
  - 客户端已经断开
  - 后端继续处理（浪费资源）
  - 最好：后端支持 context 取消
  - 现实：很多后端不支持
```

**挑战 4：错误处理**

**错误分类**：

```
1. 连接错误：
   - Connection Refused: 后端不可达
   - Connection Timeout: 连接超时
   - 处理：标记后端不健康，重试其他后端

2. 超时错误：
   - Request Timeout: 请求超时
   - 处理：记录慢请求，可能重试

3. HTTP 错误：
   - 5xx: 后端内部错误
   - 4xx: 客户端错误
   - 处理：5xx 可能重试，4xx 不重试

4. 网络错误：
   - Connection Reset: 连接被重置
   - EOF: 连接关闭
   - 处理：重试

5. 客户端断开：
   - Context Canceled: 客户端取消请求
   - 处理：记录日志，停止转发
```

**重试策略**：

```
问题：哪些错误应该重试？

可重试：
  ✅ 连接错误（选择其他后端）
  ✅ 超时（选择其他后端）
  ✅ 502, 503, 504（后端临时错误）
  ✅ 网络错误

不应重试：
  ❌ 4xx 错误（客户端问题）
  ❌ POST/PUT（可能不幂等）
  ❌ 已经发送了部分 Body（无法重试）
  ❌ 客户端取消请求

思考：
  1. 如何判断请求是否幂等？
  2. 如何判断 Body 是否已经发送？
  3. 重试次数应该限制吗？
```

#### 响应处理

**复制响应的挑战**：

```
1. 响应头：
   - 复制所有响应头吗？
   - 需要删除 hop-by-hop 头

2. 响应状态码：
   - 直接复制

3. 响应体：
   - 可能很大（文件下载）
   - 流式复制（避免 OOM）
   - 使用 io.Copy

4. Trailer（可选）：
   - HTTP 支持在 Body 后发送额外的头
   - 很少使用
```

**性能优化**：

```
问题：如何高效复制数据？

方案 1：逐字节复制
  - 慢

方案 2：使用缓冲区
  buf := make([]byte, 32*1024)
  io.CopyBuffer(dst, src, buf)
  - 快，但每次分配内存

方案 3：使用对象池
  bufferPool := sync.Pool{
      New: func() interface{} {
          return make([]byte, 32*1024)
      },
  }
  - 更快，复用内存

方案 4：零拷贝（如果支持）
  - 某些系统调用支持（sendfile, splice）
  - 最快，但平台相关
```

### 测试策略

**单元测试**：

```
Test 1：基础转发
  - 创建 mock 后端（httptest.Server）
  - 发送 GET 请求
  - 验证响应正确

Test 2：请求头处理
  - 验证 X-Forwarded-* 头正确设置
  - 验证 hop-by-hop 头被删除

Test 3：请求体转发
  - 发送 POST 请求（带 Body）
  - 验证后端收到完整 Body

Test 4：大文件转发
  - 上传 100MB 文件
  - 验证内存占用稳定（< 50MB）
  - 验证文件完整

Test 5：超时处理
  - Mock 后端延迟 10s
  - 设置超时 5s
  - 验证触发超时错误

Test 6：连接错误处理
  - 后端不可达
  - 验证返回 502 或 503
  - 验证错误日志

Test 7：并发转发
  - 100 个并发请求
  - 验证响应正确
  - 验证无竞态（go test -race）
```

**集成测试**：

```
Test 1：真实 HTTP 服务器
  - 启动真实的后端服务器
  - 通过负载均衡器访问
  - 验证端到端流程

Test 2：连接池行为
  - 发送 100 个请求
  - 验证连接复用（只建立少量连接）
  - 验证空闲连接被关闭

Test 3：长连接场景
  - WebSocket 或 Server-Sent Events
  - 验证长连接正常工作
```

**性能测试**：

```
Benchmark 1：小请求
  - GET / （无 Body）
  - QPS 和延迟

Benchmark 2：大请求
  - POST /upload （1MB Body）
  - 吞吐量和内存占用

Benchmark 3：不同并发数
  - 1, 10, 100, 1000 并发
  - 观察性能曲线
```

### 思考题

**Q1：反向代理应该修改响应吗？**

```
场景：
  - 后端返回 HTML，包含后端的 URL
  - 客户端收到后，直接访问后端（绕过负载均衡器）

方案 1：不修改
  - 简单
  - 但可能暴露后端地址

方案 2：重写 URL
  - 复杂（需要解析 HTML/JSON）
  - 性能开销大

方案 3：要求后端返回相对路径
  - 最佳实践
  - 需要后端配合

思考：你会选择哪种方案？
```

**Q2：如何处理 WebSocket？**

```
WebSocket 特点：
  - 升级协议（从 HTTP 到 WebSocket）
  - 双向通信
  - 长连接

挑战：
  - 如何转发 Upgrade 握手？
  - 如何保持长连接？
  - 如何双向转发数据？

提示：
  - 需要 goroutine 同时读写
  - 需要正确处理 Upgrade 头
  - 需要考虑粘性会话（同一客户端总是路由到同一后端）
```

**Q3：如何实现缓存？**

```
场景：
  - 静态内容（如图片、CSS）
  - 可以缓存减少后端压力

挑战：
  1. 如何判断是否可缓存？
     - 检查 Cache-Control 头
     - 检查 HTTP 方法（只缓存 GET）

  2. 缓存 Key 如何设计？
     - URL + Query 参数？
     - 加上 Headers（如 Accept-Language）？

  3. 缓存多大？
     - LRU 淘汰？
     - 最大大小限制？

  4. 缓存失效？
     - 过期时间？
     - 主动失效？

思考：
  1. 缓存应该在 Proxy 里实现还是独立的中间件？
  2. 如何测试缓存功能？
```

### 实践任务

#### 任务 1：实现基础反向代理

1. 实现 `ReverseProxy` 结构体
2. 实现 `Forward()` 方法
3. 实现请求头处理（X-Forwarded-*）
4. 实现响应复制

**要求**：

- 支持 GET, POST, PUT, DELETE
- 正确处理请求头和响应头
- 流式复制 Body（避免 OOM）

#### 任务 2：配置连接池

1. 配置 `http.Transport`
2. 设置合适的超时参数
3. 设置连接池大小

**验证**：

- 观察连接复用情况
- 压力测试，观察性能

#### 任务 3：错误处理

1. 实现连接错误处理
2. 实现超时错误处理
3. 实现 HTTP 错误处理

**验证**：

- 模拟各种错误场景
- 验证错误日志完整

#### 任务 4：性能测试

1. 编写 Benchmark
2. 对比不同配置的性能
3. 找出性能瓶颈

**思考**：

- 瓶颈在哪里？
- 如何优化？

### 验证标准

- [ ] 能够正确转发 HTTP 请求
- [ ] 请求头和响应头处理正确
- [ ] 支持大文件传输（无 OOM）
- [ ] 超时控制生效
- [ ] 错误处理正确
- [ ] 性能满足要求（QPS > 5000）
- [ ] 测试覆盖率 > 80%

---

## 迭代 4：健康检查机制

### 为什么需要健康检查？

**问题场景**：

```
有 3 个后端：A, B, C

时刻 T1: 所有后端正常
  → 流量均匀分配：33% 每个

时刻 T2: 后端 B 宕机
  → 如果没有健康检查：
    - 继续向 B 发送 33% 的流量
    - 这 33% 的请求全部失败
    - 用户体验很差

  → 有健康检查：
    - 检测到 B 不健康
    - 停止向 B 发送流量
    - 流量重新分配给 A 和 C (50% 每个)
    - 故障影响最小化
```

**健康检查的价值**：

```
1. 故障检测：及时发现后端问题
2. 自动恢复：后端恢复后自动加回
3. 提高可用性：减少失败请求
4. 降低延迟：避免请求发到慢后端
```

### 主动 vs 被动健康检查

#### 原理对比

**主动健康检查（Active Health Check）**：

```
原理：
  定期主动发送探测请求（如 GET /health）
  根据响应判断后端健康状态

流程：
  每隔 10 秒：
    对每个后端：
      发送 GET /health
      if 响应 200 OK:
          记录成功
      else:
          记录失败

      if 连续失败 >= 3 次:
          标记为不健康
      if 连续成功 >= 2 次:
          标记为健康
```

**被动健康检查（Passive Health Check）**：

```
原理：
  监控真实请求的结果
  根据失败率判断后端健康状态

流程：
  每次 Proxy.Forward() 后：
    记录请求结果（成功/失败）
    计算最近 N 次请求的失败率

    if 失败率 > 50%:
        标记为不健康
    if 失败率 < 10%:
        标记为健康
```

#### 对比分析

| 维度                 | 主动检查                         | 被动检查             |
| -------------------- | -------------------------------- | -------------------- |
| **检测延迟**   | 取决于检查间隔（如 10s）         | 立即（请求失败时）   |
| **额外开销**   | 有（探测请求）                   | 无                   |
| **流量要求**   | 不需要真实流量                   | 需要真实流量         |
| **检测准确性** | 可能不准（探测请求 != 真实请求） | 准确（基于真实请求） |
| **误判风险**   | 网络抖动可能误判                 | 低流量可能漏检       |
| **适用场景**   | 高可用系统                       | 高流量系统           |

**思考题**：

```
Q1: 为什么主动检查不够准确？

A: 场景：
  - 健康检查端点 /health 返回 200 OK
  - 但是业务接口 /api/users 有 bug，返回 500

  主动检查：认为健康 ✅
  实际情况：不健康 ❌

Q2: 被动检查为什么需要真实流量？

A: 如果某个后端流量很少（如权重低）：
  - 很长时间才有一个请求
  - 即使不健康也无法及时检测到

Q3: 如何结合两者优势？

A: 组合策略：
  - 主动检查：定期探测，提前发现问题
  - 被动检查：实时监控，快速响应
  - 双重确认：只有两者都认为不健康才真正下线
```

### 主动健康检查设计

#### 核心问题

**问题 1：状态抖动**

**场景**：

```
网络偶尔丢包，后端其实是健康的

时间线：
T1: 检查成功 → Healthy
T2: 网络丢包，检查失败 → Unhealthy?
T3: 检查成功 → Healthy?
T4: 检查失败 → Unhealthy?

结果：
  状态频繁切换 Healthy ↔ Unhealthy
  流量分配不稳定
  用户体验差
```

**解决方案：阈值机制**

**设计**：

```
维护连续成功/失败次数

状态机：
  Healthy 状态：
    检查成功 → consecutive_fails = 0
    检查失败 → consecutive_fails++
    if consecutive_fails >= FailThreshold (如 3):
        转换为 Unhealthy

  Unhealthy 状态：
    检查失败 → consecutive_success = 0
    检查成功 → consecutive_success++
    if consecutive_success >= SuccessThreshold (如 2):
        转换为 Healthy
```

**参数选择**：

```
FailThreshold (失败阈值):
  - 太小（如 1）：误判率高，一次失败就下线
  - 太大（如 10）：检测延迟长，多次失败才下线
  - 推荐：2-3

SuccessThreshold (成功阈值):
  - 太小（如 1）：可能虚假恢复
  - 太大（如 10）：恢复太慢
  - 推荐：2-3

Interval (检查间隔):
  - 太短（如 1s）：额外开销大
  - 太长（如 60s）：检测延迟长
  - 推荐：5-10s
```

**思考题**：

```
Q1: FailThreshold 和 SuccessThreshold 应该相等吗？

A: 不一定！
  - FailThreshold 可以更大（更保守下线）
  - SuccessThreshold 可以更小（更快恢复）
  - 或者反过来（更激进下线，更保守恢复）

  思考：你会如何选择？为什么？

Q2: 如何计算平均检测延迟？

A: 最坏情况：
  - 后端在检查后立即故障
  - 下次检查前都是健康的
  - 检测延迟 = Interval + FailThreshold * Interval

  示例：Interval=10s, FailThreshold=3
  检测延迟 = 10s + 3*10s = 40s

  如何缩短？
    - 减小 Interval（但增加开销）
    - 减小 FailThreshold（但增加误判）
    - 结合被动检查（立即检测）
```

**问题 2：检查风暴**

**场景**：

```
有 100 个后端
检查间隔 5 秒

如果同时检查：
  瞬间 100 个并发请求
  可能打垮后端或网络
  可能自己成为瓶颈
```

**解决方案 1：错峰检查**

**设计**：

```
给每个后端分配不同的初始延迟

示例：100 个后端，间隔 10s
  后端 0: 延迟 0s，然后每 10s 检查
  后端 1: 延迟 0.1s，然后每 10s 检查
  后端 2: 延迟 0.2s，然后每 10s 检查
  ...
  后端 99: 延迟 9.9s，然后每 10s 检查

公式：
  initial_delay = (index / total) * interval

结果：
  检查请求均匀分布在 10s 内
  避免瞬间高并发
```

**解决方案 2：限流**

**设计**：

```
限制同时进行的检查数量

使用信号量（semaphore）：
  max_concurrent_checks = 10

  for each backend:
      acquire_semaphore()
      check_backend(backend)
      release_semaphore()

结果：
  同时最多 10 个检查在进行
  避免过载
```

**对比**：

```
错峰检查：
  优点：实现简单，检查均匀分布
  缺点：固定的检查顺序

限流：
  优点：灵活控制并发数
  缺点：可能出现短时间的高峰

推荐：结合使用
  - 错峰作为主要手段
  - 限流作为保险
```

**问题 3：优雅关闭**

**场景**：

```
负载均衡器需要停止

问题：
  - 健康检查 goroutine 正在运行
  - 如何停止它们？
  - 如何确保正在进行的检查完成？
```

**错误做法**：

```
❌ 直接退出程序
  - 正在进行的检查被中断
  - 可能导致资源泄漏

❌ 使用 time.Sleep 等待
  - 不确定需要等多久
  - 可能等不够或等太久
```

**正确做法：使用 Context**

**设计**：

```
启动健康检查：
  ctx, cancel := context.WithCancel(context.Background())

  for each backend:
      go checkLoop(ctx, backend)

停止健康检查：
  cancel()  // 通知所有 goroutine 停止
  wg.Wait() // 等待所有 goroutine 结束

checkLoop 实现：
  ticker := time.NewTicker(interval)
  defer ticker.Stop()

  for {
      select {
      case <-ticker.C:
          checkOne(backend)
      case <-ctx.Done():
          return  // 收到停止信号，退出
      }
  }
```

**验证**：

```
测试优雅关闭：
  1. 启动健康检查
  2. 等待几次检查完成
  3. 调用 Stop()
  4. 验证：
     - 所有 goroutine 在合理时间内结束（< 2 * interval）
     - 无资源泄漏（goroutine, timer, socket）
```

### 被动健康检查设计

#### 失败率计算

**问题**：如何计算失败率？

**方案 1：固定窗口**

```
维护固定窗口（如最近 1 分钟）的统计

数据结构：
  window_start: timestamp
  success_count: int
  failure_count: int

更新：
  每次请求结果：
    if 窗口过期:
        重置计数
    if 成功:
        success_count++
    else:
        failure_count++

计算失败率：
  failure_count / (success_count + failure_count)
```

**问题**：

```
边界效应：

时间线（窗口 1 分钟）：
  T0-T59: 100 个成功请求
  T60: 窗口重置，计数清零
  T60-T61: 10 个失败请求

T61 时的失败率：
  10 / 10 = 100%  ❌ 误判！

问题：
  - 窗口重置时丢失历史数据
  - 短时间内的少量请求影响过大
```

**方案 2：滑动窗口**

```
维护最近 N 次请求的结果

数据结构：
  results: []bool  // true=成功, false=失败
  size: int        // 窗口大小

更新：
  results.append(success)
  if len(results) > size:
      results = results[1:]  // 移除最旧的

计算失败率：
  failures := count(results, false)
  return failures / len(results)
```

**优点**：

```
✅ 精确：反映最近 N 次的实际情况
✅ 平滑：没有边界效应
✅ 实时：每次请求都更新
```

**缺点**：

```
❌ 内存占用：需要存储 N 个 bool
  - 如果 N=1000，每个后端 1KB
  - 如果 100 个后端，共 100KB
  - 可接受
```

**优化：环形缓冲区**

```
使用固定大小的数组，循环写入

数据结构：
  results: [N]bool
  index: int  // 下次写入的位置
  count: int  // 当前结果数（<= N）

更新：
  results[index] = success
  index = (index + 1) % N
  if count < N:
      count++

计算失败率：
  failures := 0
  for i := 0; i < count; i++:
      if !results[i]:
          failures++
  return failures / count
```

**方案 3：滑动时间窗口（桶）**

```
按时间分桶，记录每个桶的统计

数据结构：
  buckets: []Bucket
  bucket_size: time.Duration  // 如 10s

  Bucket:
      timestamp: time.Time
      success: int
      failure: int

更新：
  current_bucket := get_or_create_bucket(now)
  if success:
      current_bucket.success++
  else:
      current_bucket.failure++

计算失败率：
  window_start := now - window_size  // 如 60s
  success, failure := 0, 0
  for each bucket:
      if bucket.timestamp >= window_start:
          success += bucket.success
          failure += bucket.failure
  return failure / (success + failure)
```

**对比**：

| 方案               | 内存            | 精度           | 实现复杂度 |
| ------------------ | --------------- | -------------- | ---------- |
| 固定窗口           | 低              | 低（边界效应） | 低         |
| 滑动窗口（计数）   | 中 (O(N))       | 高             | 中         |
| 滑动时间窗口（桶） | 低 (O(buckets)) | 中             | 高         |

**推荐**：滑动窗口（计数）

- 实现相对简单
- 精度高
- 内存占用可接受

**参数选择**：

```
window_size (窗口大小):
  - 太小（如 10）：容易波动，少量失败就误判
  - 太大（如 1000）：反应慢，积累很多失败才检测到
  - 推荐：50-100

failure_threshold (失败率阈值):
  - 太低（如 10%）：太敏感
  - 太高（如 80%）：太迟钝
  - 推荐：30%-50%
```

#### 集成到 Proxy

**设计思路**：

```
在 Proxy.Forward() 中记录请求结果

问题 1：哪些错误算"失败"？
  ✅ 连接错误（后端不可达）
  ✅ 超时（后端响应慢）
  ✅ 5xx 错误（后端内部错误）
  ❌ 4xx 错误（客户端错误，不算后端问题）
  ❌ 客户端取消（不算后端问题）

问题 2：成功的定义？
  - 响应状态码 < 500？
  - 还是只有 200-299 算成功？

  推荐：< 500 算成功
  原因：3xx 重定向，4xx 客户端错误，都不是后端问题

问题 3：如何传递结果到 PassiveHealthChecker？
  方案 A：Proxy 持有 PassiveHealthChecker
  方案 B：通过回调函数
  方案 C：通过 channel 传递事件

  推荐：方案 A（简单直接）
```

**实现思路**：

```
type Proxy struct {
    client       *http.Client
    passiveCheck *PassiveHealthChecker
}

func (p *Proxy) Forward(w, r, backend) error {
    err := p.doForward(w, r, backend)

    // 记录结果
    success := p.isSuccess(err)
    p.passiveCheck.RecordResult(backend, success)

    return err
}

func (p *Proxy) isSuccess(err error) bool {
    if err != nil {
        // 检查是否是客户端取消
        if errors.Is(err, context.Canceled) {
            return true  // 不算后端失败
        }
        return false
    }
    return true
}
```

### 组合策略设计

**如何结合主动和被动检查？**

**策略 1：OR 逻辑（激进）**

```
只要主动或被动任一认为不健康，就标记为不健康

优点：快速检测故障
缺点：误判率高
```

**策略 2：AND 逻辑（保守）**

```
只有主动和被动都认为不健康，才标记为不健康

优点：误判率低
缺点：检测延迟长
```

**策略 3：分级响应（推荐）**

```
根据严重程度决定：

状态：Healthy → Degraded → Unhealthy

转换规则：
  Healthy → Degraded:
    - 被动检查失败率 > 30% OR 主动检查失败 1 次
    - 减少流量（降权重 50%）

  Degraded → Unhealthy:
    - 被动检查失败率 > 50% OR 主动检查连续失败 3 次
    - 停止流量

  Unhealthy → Degraded:
    - 主动检查连续成功 2 次

  Degraded → Healthy:
    - 被动检查失败率 < 10% AND 主动检查连续成功 3 次
```

**好处**：

```
1. 灵活：不是简单的健康/不健康二元状态
2. 平滑：逐步减少流量，而不是突然下线
3. 容错：小问题不立即下线
```

### 测试策略

**单元测试**：

**Test 1：主动检查 - 阈值机制**

```
场景：验证连续失败才下线

步骤：
  1. 创建 Backend，初始健康
  2. Mock HTTP 返回 500（失败）
  3. 检查 2 次，验证仍然健康（FailThreshold=3）
  4. 检查第 3 次，验证变为不健康

边界：
  - 连续失败中间有成功（重置计数）
  - 恰好达到阈值
```

**Test 2：主动检查 - 错峰**

```
场景：验证检查时间分布均匀

步骤：
  1. 创建 100 个 Backend
  2. 启动健康检查（Interval=10s）
  3. 记录每个后端第一次检查的时间
  4. 验证时间分布在 [0, 10s] 内
  5. 验证没有同时检查（允许 ±100ms 误差）
```

**Test 3：被动检查 - 滑动窗口**

```
场景：验证失败率计算正确

步骤：
  1. WindowSize=10
  2. 记录：S S S S S F F F F F（5 成功，5 失败）
  3. 失败率应该是 50%
  4. 再记录：S S（2 成功）
  5. 窗口：S S S F F F F F S S（去掉最旧的 2 个 S）
  6. 失败率：5/10 = 50%
```

**Test 4：被动检查 - 集成**

```
场景：验证 Proxy 正确记录结果

步骤：
  1. Mock 后端，返回 500
  2. 调用 Proxy.Forward() 10 次
  3. 验证 PassiveHealthChecker 记录了 10 次失败
  4. 验证 Backend 被标记为不健康
```

**Test 5：优雅关闭**

```
场景：验证健康检查正确停止

步骤：
  1. 启动健康检查
  2. 等待 2 秒
  3. 调用 Stop()
  4. 验证：
     - 在 2 * Interval 内所有 goroutine 结束
     - 使用 runtime.NumGoroutine() 验证无泄漏
```

**集成测试**：

**Test 1：模拟后端故障**

```
场景：后端从健康到不健康再恢复

步骤：
  1. 启动真实的 HTTP 后端
  2. 验证被标记为健康
  3. 停止后端
  4. 等待检测（等待 Interval * FailThreshold）
  5. 验证被标记为不健康
  6. 重启后端
  7. 等待恢复（等待 Interval * SuccessThreshold）
  8. 验证恢复为健康
```

**Test 2：网络抖动**

```
场景：偶尔的网络丢包不应导致下线

步骤：
  1. 创建 Backend
  2. Mock 网络：90% 成功，10% 失败
  3. 运行 100 次检查
  4. 验证：
     - 主动检查：仍然健康（连续失败 < 3）
     - 被动检查：仍然健康（失败率 < 50%）
```

### 思考题

**Q1：健康检查端点应该检查什么？**

```
场景：后端的 /health 端点

简单实现：
  GET /health → 200 OK

问题：
  - 进程还活着，但数据库连接断了？
  - 进程还活着，但内存用完了？
  - 进程还活着，但依赖服务挂了？

深度检查：
  - 检查数据库连接
  - 检查关键依赖
  - 检查资源使用率（CPU, 内存）

权衡：
  - 检查太浅：不准确
  - 检查太深：端点变慢，反而影响健康检查

思考：
  1. /health 应该有超时吗？
  2. /health 应该有缓存吗（避免频繁检查依赖）？
  3. 应该有浅检查（/health）和深检查（/health/deep）两个端点吗？
```

**Q2：如何处理健康检查的健康检查？**

```
问题：
  如果健康检查模块本身有 bug 怎么办？
  如果 goroutine 卡死，不再检查怎么办？

监控健康检查：
  1. 记录最后一次检查时间
  2. 如果超过 2 * Interval 没有检查，告警
  3. 监控检查的延迟（如果越来越慢，可能有问题）

思考：
  谁来监控监控系统？（著名的递归问题）
```

**Q3：如何实现预热（Warm-up）？**

```
场景：
  后端刚启动，还在初始化（加载缓存、预热连接池）
  如果立即分配全流量，可能打挂

预热策略：
  健康状态：Unhealthy → Warming → Healthy

  Warming 阶段：
    - 逐步增加流量
    - 0% → 10% → 30% → 50% → 100%
    - 持续时间：如 60 秒

  转换条件：
    - Unhealthy → Warming: 主动检查成功
    - Warming → Healthy: 预热时间到 且 被动检查失败率 < 10%
    - Warming → Unhealthy: 被动检查失败率 > 50%

思考：
  1. 预热时间如何确定？
  2. 如何逐步增加流量（修改权重？）
  3. 是否所有后端都需要预热？
```

### 实践任务

#### 任务 1：实现主动健康检查

1. 实现阈值机制
2. 实现错峰启动
3. 实现优雅关闭

**验证**：

- 单元测试覆盖所有场景
- 集成测试验证端到端流程
- 无 goroutine 泄漏

#### 任务 2：实现被动健康检查

1. 实现滑动窗口
2. 实现失败率计算
3. 集成到 Proxy

**验证**：

- 测试失败率计算正确
- 测试不同窗口大小的效果
- 压力测试性能

#### 任务 3：实现组合策略

1. 实现分级状态（Healthy, Degraded, Unhealthy）
2. 实现状态转换逻辑
3. 集成主动和被动检查

**验证**：

- 绘制状态转换图
- 测试各种转换场景
- 验证流量分配符合预期

### 验证标准

- [ ] 主动检查正确检测故障
- [ ] 被动检查正确计算失败率
- [ ] 阈值机制避免状态抖动
- [ ] 错峰避免检查风暴
- [ ] 优雅关闭无资源泄漏
- [ ] 组合策略合理
- [ ] 测试覆盖率 > 80%

---

（由于篇幅限制，迭代 5-9 将在后续继续...）

## 继续学习

本指南专注于**教你如何思考**，而不是给你答案。

**下一步**：

1. 完成迭代 0-4 的实现
2. 遇到问题时，回来重新思考设计
3. 不断重构，追求优雅的代码

**学习资源**：

- 参考你已经学习的：
  - `notes/week1/module2/健康检查机制深度解析.md`
  - `notes/week1/module2/故障检测与自动转移深度解析.md`
- 阅读源码：Nginx, HAProxy, Envoy
- 实践：在真实项目中应用

**记住**：

> 好的代码不是一次写成的，是不断重构出来的。
> 理解原理比记住代码更重要。
> 遇到问题，先思考为什么，再思考怎么做。

---

**待续**：迭代 5-9 将继续深入探讨熔断器、高级算法、配置管理、可观测性和性能优化的原理与实践。

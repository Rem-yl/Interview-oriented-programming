# V5: 微服务架构 (Microservices Architecture)

## 架构概述

微服务架构是一种将单体应用拆分为多个**小型、独立部署的服务**的架构风格。每个服务围绕特定业务能力构建，运行在独立进程中，通过轻量级通信机制（通常是HTTP RESTful API）协作。

## 从 V4 到 V5 的演进

### V4 的问题回顾

```
┌────────────────────────────────────────┐
│         Single Application             │
│  ┌──────────┐  ┌──────────┐           │
│  │ Command  │  │  Query   │           │
│  │  Side    │  │  Side    │           │
│  └──────────┘  └──────────┘           │
│  ┌──────────────────────────┐         │
│  │   Shared Database        │         │
│  └──────────────────────────┘         │
└────────────────────────────────────────┘
          单点，无法独立扩展
```

**问题**：
- ❌ 单点故障（应用崩溃 = 系统不可用）
- ❌ 无法独立扩展（Todo和User都要扩展）
- ❌ 技术栈绑定（都要用同一语言）
- ❌ 团队协作困难（共享代码库）
- ❌ 部署风险高（一个bug影响全部）

### V5 的解决方案

```
                 ┌─────────────┐
                 │  API Gateway │
                 └──────┬──────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
   ┌────▼────┐     ┌───▼────┐    ┌────▼─────┐
   │  Todo   │     │  User  │    │  Notify  │
   │ Service │     │Service │    │ Service  │
   └────┬────┘     └───┬────┘    └────┬─────┘
        │              │              │
   ┌────▼────┐    ┌───▼────┐    ┌────▼─────┐
   │ Todo DB │    │User DB │    │ Queue    │
   └─────────┘    └────────┘    └──────────┘

   每个服务：
   - 独立部署
   - 独立数据库
   - 独立扩展
```

## 设计目标

- ✅ 服务自治（独立开发、部署、扩展）
- ✅ 故障隔离（一个服务挂掉不影响其他）
- ✅ 技术异构（不同服务可用不同技术栈）
- ✅ 团队自治（每个团队负责一个服务）
- ✅ 按需扩展（只扩展需要的服务）

## 项目结构

```
v5-microservices/
├── README.md
├── docker-compose.yml          # 本地开发环境
│
├── api-gateway/                # API网关
│   ├── go.mod
│   ├── main.go
│   ├── routes/
│   └── middleware/
│
├── todo-service/               # 待办事项服务
│   ├── go.mod
│   ├── main.go
│   ├── cmd/
│   ├── internal/
│   │   ├── domain/
│   │   ├── application/
│   │   └── infrastructure/
│   ├── api/
│   │   └── http/
│   └── Dockerfile
│
├── user-service/               # 用户服务
│   ├── go.mod
│   ├── main.go
│   ├── internal/
│   ├── api/
│   └── Dockerfile
│
├── notification-service/       # 通知服务
│   ├── go.mod
│   ├── main.go
│   ├── internal/
│   └── Dockerfile
│
├── shared/                     # 共享库（谨慎使用）
│   ├── auth/
│   ├── errors/
│   └── logger/
│
└── infrastructure/             # 基础设施
    ├── postgres/
    ├── redis/
    └── kafka/
```

## 服务拆分原则

### 1. 按业务能力拆分（推荐）

```
Todo Service    - 管理待办事项的所有操作
User Service    - 用户认证、授权、个人信息
Notification    - 发送邮件、短信、推送通知
```

### 2. 按子域拆分（DDD）

```
核心域    - Todo Management（核心业务）
支撑域    - User Management（支撑）
通用域    - Notification（可复用）
```

### 3. 拆分的反模式

❌ **过度拆分**：每个表一个服务
❌ **按技术拆分**：Frontend Service, Backend Service
❌ **按层级拆分**：Controller Service, Service Service

## 核心组件

### 1. API Gateway（API网关）

```go
// api-gateway/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http/httputil"
    "net/url"
)

type Gateway struct {
    todoService  string
    userService  string
    notifyService string
}

func NewGateway() *Gateway {
    return &Gateway{
        todoService:  "http://todo-service:8081",
        userService:  "http://user-service:8082",
        notifyService: "http://notification-service:8083",
    }
}

func (g *Gateway) Setup() *gin.Engine {
    r := gin.Default()

    // 认证中间件
    r.Use(AuthMiddleware())

    // 路由转发
    todos := r.Group("/todos")
    {
        todos.Any("/*path", g.proxyTo(g.todoService))
    }

    users := r.Group("/users")
    {
        users.Any("/*path", g.proxyTo(g.userService))
    }

    return r
}

func (g *Gateway) proxyTo(target string) gin.HandlerFunc {
    targetURL, _ := url.Parse(target)
    proxy := httputil.NewSingleHostReverseProxy(targetURL)

    return func(c *gin.Context) {
        proxy.ServeHTTP(c.Writer, c.Request)
    }
}
```

**职责**：
- 路由转发
- 认证和授权
- 限流和熔断
- 日志和监控

### 2. Todo Service（待办服务）

```go
// todo-service/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "todo-service/internal/application"
    "todo-service/internal/infrastructure/persistence"
    "todo-service/api/http"
)

func main() {
    // 初始化依赖
    db := initDatabase()
    todoRepo := persistence.NewTodoRepository(db)
    todoService := application.NewTodoService(todoRepo)
    handler := http.NewTodoHandler(todoService)

    // HTTP服务器
    r := gin.Default()

    // 健康检查
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // Todo API
    todos := r.Group("/todos")
    {
        todos.POST("", handler.Create)
        todos.GET("", handler.List)
        todos.GET("/:id", handler.Get)
        todos.PUT("/:id", handler.Update)
        todos.DELETE("/:id", handler.Delete)
    }

    r.Run(":8081")
}
```

**特点**：
- 独立的数据库
- 完整的CQRS实现
- 发布领域事件

### 3. User Service（用户服务）

```go
// user-service/internal/application/user_service.go
package application

type UserService struct {
    userRepo UserRepository
    jwtSecret string
}

func (s *UserService) Register(username, password string) (*User, error) {
    // 注册逻辑
    hashedPassword := hashPassword(password)
    user := NewUser(username, hashedPassword)
    return s.userRepo.Save(user)
}

func (s *UserService) Login(username, password string) (string, error) {
    user, err := s.userRepo.FindByUsername(username)
    if err != nil {
        return "", ErrInvalidCredentials
    }

    if !verifyPassword(user.PasswordHash, password) {
        return "", ErrInvalidCredentials
    }

    // 生成JWT
    token := generateJWT(user.ID, s.jwtSecret)
    return token, nil
}

func (s *UserService) ValidateToken(token string) (*Claims, error) {
    return parseJWT(token, s.jwtSecret)
}
```

### 4. Notification Service（通知服务）

```go
// notification-service/internal/application/notifier.go
package application

type NotificationService struct {
    emailSender EmailSender
}

func (s *NotificationService) SendTodoCompletedNotification(userID int, todoTitle string) error {
    user, _ := s.getUserEmail(userID)

    message := EmailMessage{
        To:      user.Email,
        Subject: "Todo Completed",
        Body:    fmt.Sprintf("Your todo '%s' has been completed!", todoTitle),
    }

    return s.emailSender.Send(message)
}
```

## 服务间通信

### 1. 同步通信（HTTP REST）

```go
// todo-service 调用 user-service
type UserServiceClient struct {
    baseURL string
    client  *http.Client
}

func (c *UserServiceClient) GetUser(userID int) (*User, error) {
    url := fmt.Sprintf("%s/users/%d", c.baseURL, userID)

    resp, err := c.client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var user User
    json.NewDecoder(resp.Body).Decode(&user)

    return &user, nil
}

// 使用
func (s *TodoService) CreateTodo(userID int, title string) error {
    // 验证用户存在
    user, err := s.userClient.GetUser(userID)
    if err != nil {
        return ErrUserNotFound
    }

    // 创建待办
    todo := NewTodo(userID, title)
    return s.todoRepo.Save(todo)
}
```

**问题**：
- 服务耦合（User服务挂了，Todo服务也创建失败）
- 性能差（同步等待）

### 2. 异步通信（消息队列）

```go
// todo-service 发布事件
type EventPublisher struct {
    kafka *kafka.Producer
}

func (p *EventPublisher) PublishTodoCompleted(todo *Todo) error {
    event := TodoCompletedEvent{
        TodoID: todo.ID,
        UserID: todo.UserID,
        Title:  todo.Title,
        Time:   time.Now(),
    }

    message, _ := json.Marshal(event)

    return p.kafka.Produce(&kafka.Message{
        Topic: "todo-completed",
        Value: message,
    })
}

// notification-service 订阅事件
type EventConsumer struct {
    kafka    *kafka.Consumer
    notifier *NotificationService
}

func (c *EventConsumer) Start() {
    c.kafka.Subscribe([]string{"todo-completed"})

    for {
        msg := c.kafka.Poll(100)
        if msg == nil {
            continue
        }

        var event TodoCompletedEvent
        json.Unmarshal(msg.Value, &event)

        // 发送通知
        c.notifier.SendTodoCompletedNotification(
            event.UserID,
            event.Title,
        )
    }
}
```

**优点**：
- 解耦（Notification挂了不影响Todo）
- 异步（不阻塞主流程）
- 可靠（消息持久化）

## 数据管理

### 1. 每个服务独立数据库

```yaml
# docker-compose.yml
services:
  todo-db:
    image: postgres:15
    environment:
      POSTGRES_DB: todo_db
    ports:
      - "5432:5432"

  user-db:
    image: postgres:15
    environment:
      POSTGRES_DB: user_db
    ports:
      - "5433:5432"
```

### 2. 跨服务查询问题

**反模式：直接查询其他服务的数据库**
```go
// ❌ 不要这样做
db.Query("SELECT * FROM user_db.users WHERE id = ?", userID)
```

**模式1：API调用**
```go
// ✅ 通过API获取
user := userClient.GetUser(userID)
```

**模式2：数据冗余**
```go
// todos 表中冗余用户信息
type Todo struct {
    ID       int
    UserID   int
    UserName string  // 冗余
    Title    string
}
```

**模式3：CQRS读模型**
```sql
-- 在读服务中创建联合视图
CREATE VIEW todo_with_user AS
SELECT t.*, u.username
FROM todo_read_models t
LEFT JOIN user_read_models u ON t.user_id = u.id;
```

## 服务发现

### 1. 客户端发现（Consul）

```go
// 服务注册
func registerService() {
    consul := api.NewClient(api.DefaultConfig())

    registration := &api.AgentServiceRegistration{
        ID:      "todo-service-1",
        Name:    "todo-service",
        Address: "localhost",
        Port:    8081,
        Check: &api.AgentServiceCheck{
            HTTP:     "http://localhost:8081/health",
            Interval: "10s",
        },
    }

    consul.Agent().ServiceRegister(registration)
}

// 服务发现
func discoverService(serviceName string) (string, error) {
    consul := api.NewClient(api.DefaultConfig())

    services, _, err := consul.Health().Service(serviceName, "", true, nil)
    if err != nil {
        return "", err
    }

    if len(services) == 0 {
        return "", errors.New("service not found")
    }

    // 负载均衡：随机选择一个实例
    instance := services[rand.Intn(len(services))]
    url := fmt.Sprintf("http://%s:%d",
        instance.Service.Address,
        instance.Service.Port,
    )

    return url, nil
}
```

### 2. 服务端发现（Kubernetes）

```yaml
# todo-service-deployment.yaml
apiVersion: v1
kind: Service
metadata:
  name: todo-service
spec:
  selector:
    app: todo-service
  ports:
    - port: 80
      targetPort: 8081
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: todo-service
spec:
  replicas: 3  # 3个实例
  template:
    spec:
      containers:
      - name: todo-service
        image: todo-service:latest
        ports:
        - containerPort: 8081
```

```go
// 通过K8s Service访问
userServiceURL := "http://user-service:80"  // K8s内部DNS
```

## 容错模式

### 1. 熔断器（Circuit Breaker）

```go
import "github.com/sony/gobreaker"

type ResilientUserClient struct {
    baseClient *UserServiceClient
    cb         *gobreaker.CircuitBreaker
}

func NewResilientUserClient(baseClient *UserServiceClient) *ResilientUserClient {
    cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        "user-service",
        MaxRequests: 3,
        Interval:    time.Minute,
        Timeout:     time.Second * 60,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 3 && failureRatio >= 0.6
        },
    })

    return &ResilientUserClient{
        baseClient: baseClient,
        cb:         cb,
    }
}

func (c *ResilientUserClient) GetUser(userID int) (*User, error) {
    result, err := c.cb.Execute(func() (interface{}, error) {
        return c.baseClient.GetUser(userID)
    })

    if err != nil {
        // 熔断打开，返回降级响应
        return &User{ID: userID, Name: "Unknown"}, nil
    }

    return result.(*User), nil
}
```

### 2. 重试（Retry）

```go
import "github.com/avast/retry-go"

func (c *UserClient) GetUserWithRetry(userID int) (*User, error) {
    var user *User

    err := retry.Do(
        func() error {
            var err error
            user, err = c.GetUser(userID)
            return err
        },
        retry.Attempts(3),
        retry.Delay(time.Second),
        retry.DelayType(retry.BackOffDelay),
    )

    return user, err
}
```

### 3. 超时（Timeout）

```go
func (c *UserClient) GetUser(userID int) (*User, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := c.client.Do(req)

    // ...
}
```

## 部署配置

### Docker Compose（开发环境）

```yaml
# docker-compose.yml
version: '3.8'

services:
  api-gateway:
    build: ./api-gateway
    ports:
      - "8080:8080"
    environment:
      TODO_SERVICE_URL: http://todo-service:8081
      USER_SERVICE_URL: http://user-service:8082
    depends_on:
      - todo-service
      - user-service

  todo-service:
    build: ./todo-service
    ports:
      - "8081:8081"
    environment:
      DATABASE_URL: postgres://todo_db:5432/todo
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - todo-db
      - kafka

  user-service:
    build: ./user-service
    ports:
      - "8082:8082"
    environment:
      DATABASE_URL: postgres://user_db:5433/user
    depends_on:
      - user-db

  notification-service:
    build: ./notification-service
    environment:
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - kafka

  todo-db:
    image: postgres:15
    environment:
      POSTGRES_DB: todo
      POSTGRES_PASSWORD: password

  user-db:
    image: postgres:15
    environment:
      POSTGRES_DB: user
      POSTGRES_PASSWORD: password

  kafka:
    image: confluentinc/cp-kafka:latest
    environment:
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
    depends_on:
      - zookeeper

  zookeeper:
    image: confluentinc/cp-zookeeper:latest
```

```bash
# 启动所有服务
docker-compose up

# 查看日志
docker-compose logs -f todo-service

# 扩展服务
docker-compose up --scale todo-service=3
```

## 优点分析

| 优点 | 说明 | 示例 |
|------|------|------|
| **独立部署** | 修改一个服务不影响其他 | 更新Todo服务，User服务无需重启 |
| **技术异构** | 不同服务可用不同技术 | Todo用Go，Notify用Python |
| **故障隔离** | 一个服务挂掉不影响全局 | Notify挂了，Todo仍可用 |
| **按需扩展** | 只扩展需要的服务 | Todo高负载，只扩展Todo服务 |
| **团队自治** | 团队独立开发部署 | Todo团队和User团队并行开发 |

## 缺点分析

| 缺点 | 说明 | 影响 |
|------|------|------|
| **运维复杂** | 管理多个服务 | 需要K8s等工具 |
| **分布式事务** | 跨服务事务困难 | 需要Saga模式 |
| **数据一致性** | 最终一致性 | 可能数据不同步 |
| **网络开销** | 服务间调用 | 性能损耗 |
| **测试复杂** | 集成测试困难 | 需要完整环境 |

## 何时使用微服务

✅ **适合场景**：
- 大型复杂系统
- 需要独立扩展不同功能
- 多团队并行开发
- 需要技术异构

❌ **不适合场景**：
- 小型应用（< 5人团队）
- 简单CRUD
- 团队不熟悉分布式系统
- 运维能力不足

## 演进到 V6 的动机

微服务虽然解决了扩展性问题，但引入了新问题：

### 问题 1: 服务间强耦合
```go
// Todo服务直接调用User服务
user, err := userClient.GetUser(userID)  // 同步调用
if err != nil {
    return err  // User服务挂了，Todo也失败
}
```

### 问题 2: 级联故障
```
User服务挂 → Todo服务失败 → API网关超时 → 整个系统慢
```

### 问题 3: 数据一致性
```
创建Todo成功，但发送通知失败 → 数据不一致
```

**这时候，我们需要事件驱动架构！**

## 练习任务

### 必做任务
1. ✅ 拆分服务（Todo, User, Notification）
2. ✅ 实现API网关
3. ✅ 配置服务发现
4. ✅ 使用Docker Compose部署

### 进阶任务
1. 🔧 实现熔断器和重试
2. 🔧 添加分布式追踪（Jaeger）
3. 🔧 实现Saga模式处理分布式事务
4. 🔧 部署到Kubernetes

### 思考题
1. 💭 如何划分服务边界？
2. 💭 如何处理分布式事务？
3. 💭 微服务的粒度如何把握？

---

**完成 V5 后，继续学习 [V6: 事件驱动架构](./v6-event-driven.md)**

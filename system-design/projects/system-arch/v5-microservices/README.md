# V5: 微服务架构实现

## 开始之前

请先阅读架构设计文档：[V5 微服务架构](../docs/v5-microservices.md)

## 实现目标

将单体应用拆分为多个独立的微服务，每个服务可独立部署和扩展。

## 服务拆分

```
api-gateway          # API网关（路由转发）
todo-service         # 待办服务
user-service         # 用户服务
notification-service # 通知服务
```

## 目录结构

```
v5-microservices/
├── docker-compose.yml
├── api-gateway/
│   ├── go.mod
│   ├── main.go
│   ├── routes/
│   └── middleware/
├── todo-service/
│   ├── go.mod
│   ├── main.go
│   ├── internal/
│   │   ├── domain/
│   │   ├── application/
│   │   └── infrastructure/
│   ├── api/
│   └── Dockerfile
├── user-service/
│   ├── go.mod
│   ├── main.go
│   ├── internal/
│   ├── api/
│   └── Dockerfile
└── notification-service/
    ├── go.mod
    ├── main.go
    ├── internal/
    └── Dockerfile
```

## 实现步骤

### Step 1: 实现 API Gateway

```go
type Gateway struct {
    todoServiceURL string
    userServiceURL string
}

func (g *Gateway) proxyToTodoService(c *gin.Context) {
    // 转发到 todo-service
    proxyRequest(c, g.todoServiceURL)
}
```

功能：
- 路由转发
- 认证中间件
- 限流
- 日志

### Step 2: 实现 Todo Service

使用 V3 或 V4 的架构：
- 独立的数据库
- 完整的业务逻辑
- HTTP API
- 健康检查端点

```go
func main() {
    r := gin.Default()

    r.GET("/health", healthCheck)
    r.POST("/todos", createTodo)
    r.GET("/todos", listTodos)

    r.Run(":8081")
}
```

### Step 3: 实现 User Service

```go
type UserService struct {
    userRepo UserRepository
    jwtSecret string
}

func (s *UserService) Register(username, password string) error
func (s *UserService) Login(username, password string) (token string, error)
func (s *UserService) ValidateToken(token string) (*Claims, error)
```

### Step 4: 实现 Notification Service

```go
type NotificationService struct {
    emailSender EmailSender
}

func (s *NotificationService) SendEmail(to, subject, body string) error
```

## 服务间通信

### 方式1: HTTP REST（同步）

```go
type UserServiceClient struct {
    baseURL string
    client  *http.Client
}

func (c *UserServiceClient) GetUser(userID int) (*User, error) {
    resp, err := c.client.Get(fmt.Sprintf("%s/users/%d", c.baseURL, userID))
    // ...
}
```

### 方式2: 消息队列（异步）

```go
// Todo Service 发布事件
publisher.Publish("todo.completed", TodoCompletedEvent{
    TodoID: todo.ID,
    UserID: todo.UserID,
})

// Notification Service 订阅事件
consumer.Subscribe("todo.completed", func(event TodoCompletedEvent) {
    // 发送通知
})
```

## Docker 部署

### Dockerfile（每个服务）

```dockerfile
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .

EXPOSE 8081
CMD ["./main"]
```

### docker-compose.yml

```yaml
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
      DATABASE_URL: postgres://postgres:password@todo-db:5432/todo
    depends_on:
      - todo-db

  user-service:
    build: ./user-service
    ports:
      - "8082:8082"
    environment:
      DATABASE_URL: postgres://postgres:password@user-db:5432/user
    depends_on:
      - user-db

  notification-service:
    build: ./notification-service
    environment:
      SMTP_HOST: smtp.example.com
      SMTP_PORT: 587

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
```

## 运行和测试

### 启动所有服务

```bash
docker-compose up --build
```

### 测试 API

```bash
# 通过网关访问
curl http://localhost:8080/todos
curl http://localhost:8080/users

# 直接访问服务（开发时）
curl http://localhost:8081/todos
curl http://localhost:8082/users
```

### 扩展服务

```bash
# 扩展 todo-service 到 3 个实例
docker-compose up --scale todo-service=3
```

## 容错模式

### 1. 熔断器（Circuit Breaker）

```go
import "github.com/sony/gobreaker"

cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "user-service",
    MaxRequests: 3,
    Timeout:     time.Second * 60,
})

user, err := cb.Execute(func() (interface{}, error) {
    return userClient.GetUser(userID)
})
```

### 2. 重试（Retry）

```go
import "github.com/avast/retry-go"

err := retry.Do(
    func() error {
        _, err := userClient.GetUser(userID)
        return err
    },
    retry.Attempts(3),
    retry.Delay(time.Second),
)
```

### 3. 超时（Timeout）

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
```

## 服务发现（可选）

### 使用 Consul

```go
// 注册服务
consul := api.NewClient(api.DefaultConfig())
consul.Agent().ServiceRegister(&api.AgentServiceRegistration{
    ID:      "todo-service-1",
    Name:    "todo-service",
    Address: "localhost",
    Port:    8081,
})

// 发现服务
services, _, _ := consul.Health().Service("user-service", "", true, nil)
instance := services[0]
url := fmt.Sprintf("http://%s:%d", instance.Service.Address, instance.Service.Port)
```

## 监控和日志

### 健康检查

每个服务都要实现：
```go
r.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{
        "status": "ok",
        "service": "todo-service",
        "version": "1.0.0",
    })
})
```

### 结构化日志

```go
import "log/slog"

slog.Info("Todo created",
    "todo_id", todo.ID,
    "user_id", userID,
    "service", "todo-service",
)
```

### 分布式追踪（可选）

使用 OpenTelemetry 或 Jaeger

## 进阶任务

### 1. 实现 API Gateway 认证

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")

        claims, err := validateToken(token)
        if err != nil {
            c.AbortWithStatus(401)
            return
        }

        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

### 2. 实现 Saga 模式

处理分布式事务：
```
CreateOrder → ReserveInventory → ProcessPayment
                    ↓ 失败
              CancelInventory ← CancelOrder
```

### 3. 部署到 Kubernetes

编写 K8s 配置文件：
- Deployment
- Service
- Ingress

## 对比单体应用

| 维度 | V4 CQRS单体 | V5 微服务 |
|------|-------------|-----------|
| 部署 | 1个应用 | N个服务 |
| 扩展 | 整体扩展 | 独立扩展 |
| 故障影响 | 全局影响 | 隔离 |
| 开发复杂度 | 中 | 高 |
| 运维复杂度 | 低 | 高 |

## 完成后

思考以下问题：

1. 如何划分服务边界？
2. 如何处理分布式事务？
3. 微服务的粒度如何把握？
4. 服务间通信应该用同步还是异步？

完成后继续学习 [V6: 事件驱动架构](../docs/v6-event-driven.md)

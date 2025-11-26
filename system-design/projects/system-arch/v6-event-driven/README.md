# V6: 事件驱动架构实现

## 开始之前

请先阅读架构设计文档：[V6 事件驱动架构](../docs/v6-event-driven.md)

## 实现目标

使用事件驱动架构，通过异步消息实现服务间的完全解耦。

## 核心概念

- **Event（事件）**：已发生的事实
- **Publisher（发布者）**：发布事件的服务
- **Subscriber（订阅者）**：订阅和处理事件的服务
- **Event Bus（事件总线）**：消息中间件（Kafka/RabbitMQ）

## 目录结构

```
v6-event-driven/
├── docker-compose.yml
├── shared/                    # 共享事件定义
│   ├── events/
│   │   ├── base_event.go
│   │   └── todo_events.go
│   └── eventbus/
│       ├── publisher.go
│       └── subscriber.go
├── todo-service/             # 发布事件
│   ├── internal/
│   │   ├── domain/
│   │   ├── application/
│   │   │   └── event_publisher.go
│   │   └── infrastructure/
│   │       └── kafka/
│   └── main.go
├── notification-service/     # 订阅事件
│   ├── internal/
│   │   └── event_handlers/
│   │       ├── todo_created_handler.go
│   │       └── todo_completed_handler.go
│   └── main.go
└── analytics-service/        # 订阅事件
    ├── internal/
    │   └── event_handlers/
    └── main.go
```

## 实现步骤

### Step 1: 定义事件

```go
// shared/events/base_event.go
type BaseEvent struct {
    EventID       string    `json:"event_id"`
    EventType     string    `json:"event_type"`
    AggregateID   string    `json:"aggregate_id"`
    Timestamp     time.Time `json:"timestamp"`
    CorrelationID string    `json:"correlation_id"`
}

// shared/events/todo_events.go
type TodoCreatedEvent struct {
    BaseEvent
    TodoID      int    `json:"todo_id"`
    UserID      int    `json:"user_id"`
    Title       string `json:"title"`
    Description string `json:"description"`
}

type TodoCompletedEvent struct {
    BaseEvent
    TodoID      int       `json:"todo_id"`
    UserID      int       `json:"user_id"`
    Title       string    `json:"title"`
    CompletedAt time.Time `json:"completed_at"`
}
```

### Step 2: 实现事件发布

```go
// todo-service/internal/infrastructure/kafka/publisher.go
type KafkaPublisher struct {
    writer *kafka.Writer
}

func (p *KafkaPublisher) Publish(ctx context.Context, topic string, event interface{}) error {
    data, _ := json.Marshal(event)

    return p.writer.WriteMessages(ctx, kafka.Message{
        Topic: topic,
        Value: data,
    })
}

// todo-service/internal/application/command_handler.go
func (h *CreateTodoHandler) Handle(cmd CreateTodoCommand) error {
    // 1. 创建领域对象
    todo, _ := domain.NewTodo(cmd.Title, cmd.Description)

    // 2. 持久化
    h.repo.Save(todo)

    // 3. 发布事件
    event := events.NewTodoCreatedEvent(todo.ID(), cmd.UserID, cmd.Title, cmd.Description)
    return h.publisher.Publish(context.Background(), "todo-events", event)
}
```

### Step 3: 实现事件订阅

```go
// notification-service/internal/infrastructure/kafka/consumer.go
type KafkaSubscriber struct {
    reader   *kafka.Reader
    handlers map[string]EventHandler
}

func (s *KafkaSubscriber) Subscribe(eventType string, handler EventHandler) {
    s.handlers[eventType] = handler
}

func (s *KafkaSubscriber) Start(ctx context.Context) error {
    for {
        message, _ := s.reader.ReadMessage(ctx)

        var baseEvent events.BaseEvent
        json.Unmarshal(message.Value, &baseEvent)

        if handler, ok := s.handlers[baseEvent.EventType]; ok {
            handler(ctx, message.Value)
        }
    }
}

// notification-service/internal/event_handlers/todo_completed_handler.go
type TodoCompletedHandler struct {
    emailSender EmailSender
    userRepo    UserRepository
}

func (h *TodoCompletedHandler) Handle(ctx context.Context, eventData []byte) error {
    var event events.TodoCompletedEvent
    json.Unmarshal(eventData, &event)

    user, _ := h.userRepo.FindByID(event.UserID)
    message := fmt.Sprintf("Your todo '%s' has been completed!", event.Title)

    return h.emailSender.Send(user.Email, "Todo Completed", message)
}
```

### Step 4: 组装服务

```go
// notification-service/main.go
func main() {
    // 创建订阅者
    subscriber := kafka.NewKafkaSubscriber(
        []string{"localhost:9092"},
        "notification-group",
        "todo-events",
    )

    // 注册事件处理器
    todoCompletedHandler := &TodoCompletedHandler{...}
    subscriber.Subscribe("todo.completed", todoCompletedHandler.Handle)

    // 启动消费
    subscriber.Start(context.Background())
}
```

## Kafka 部署

### docker-compose.yml

```yaml
version: '3.8'

services:
  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000

  kafka:
    image: confluentinc/cp-kafka:latest
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1

  todo-service:
    build: ./todo-service
    ports:
      - "8081:8081"
    environment:
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - kafka

  notification-service:
    build: ./notification-service
    environment:
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - kafka

  analytics-service:
    build: ./analytics-service
    environment:
      KAFKA_BROKERS: kafka:9092
    depends_on:
      - kafka
```

## 运行和测试

### 启动所有服务

```bash
docker-compose up --build
```

### 测试事件流

```bash
# 1. 创建待办（发布 TodoCreatedEvent）
curl -X POST http://localhost:8081/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"测试事件驱动","description":"学习EDA"}'

# 2. 完成待办（发布 TodoCompletedEvent）
curl -X PUT http://localhost:8081/todos/1/complete

# 3. 查看通知服务日志
docker-compose logs -f notification-service

# 4. 查看分析服务日志
docker-compose logs -f analytics-service
```

## 事件处理模式

### 1. 幂等性处理

```go
type EventHandler struct {
    processedEvents map[string]bool
    mu              sync.Mutex
}

func (h *EventHandler) Handle(event Event) error {
    h.mu.Lock()
    defer h.mu.Unlock()

    // 检查是否已处理
    if h.processedEvents[event.EventID] {
        return nil  // 跳过重复事件
    }

    // 处理事件
    h.process(event)

    // 标记为已处理
    h.processedEvents[event.EventID] = true
    return nil
}
```

### 2. 死信队列（DLQ）

```go
func (h *Handler) Handle(event Event) error {
    // 重试3次
    for i := 0; i < 3; i++ {
        if err := h.process(event); err == nil {
            return nil
        }
        time.Sleep(time.Second * time.Duration(i+1))
    }

    // 仍然失败，发送到死信队列
    h.dlq.Send(event)
    return nil
}
```

### 3. Saga 模式

```go
// 订单创建流程
type OrderSaga struct {
    publisher EventPublisher
}

func (s *OrderSaga) OnOrderCreated(event OrderCreatedEvent) error {
    // 1. 发布库存预留事件
    s.publisher.Publish("inventory.reserve", InventoryReserveEvent{
        OrderID: event.OrderID,
    })
}

func (s *OrderSaga) OnInventoryReserveFailed(event InventoryReserveFailedEvent) error {
    // 补偿：取消订单
    s.publisher.Publish("order.cancel", OrderCancelEvent{
        OrderID: event.OrderID,
    })
}
```

## 事件存储（Event Sourcing）

### 创建事件存储

```sql
CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(255) UNIQUE,
    event_type VARCHAR(100),
    aggregate_id VARCHAR(255),
    aggregate_type VARCHAR(100),
    event_data JSONB,
    metadata JSONB,
    timestamp TIMESTAMP,
    version INTEGER
);

CREATE INDEX idx_aggregate_id ON events(aggregate_id);
CREATE INDEX idx_event_type ON events(event_type);
```

### 实现事件存储

```go
type EventStore struct {
    db *sql.DB
}

func (s *EventStore) Save(event Event) error {
    _, err := s.db.Exec(`
        INSERT INTO events (event_id, event_type, aggregate_id, event_data, timestamp)
        VALUES ($1, $2, $3, $4, $5)
    `, event.EventID, event.EventType, event.AggregateID, event.Data, event.Timestamp)

    return err
}

func (s *EventStore) GetByAggregateID(aggregateID string) ([]Event, error) {
    rows, _ := s.db.Query(`
        SELECT event_id, event_type, event_data, timestamp
        FROM events
        WHERE aggregate_id = $1
        ORDER BY timestamp ASC
    `, aggregateID)

    // 解析事件流
    var events []Event
    for rows.Next() {
        // ...
    }

    return events, nil
}
```

### 从事件流重建状态

```go
func RebuildTodoFromEvents(events []Event) (*Todo, error) {
    var todo *Todo

    for _, event := range events {
        switch event.EventType {
        case "todo.created":
            todo = applyTodoCreated(event)
        case "todo.completed":
            todo.ApplyCompleted(event)
        case "todo.priority_changed":
            todo.ApplyPriorityChanged(event)
        }
    }

    return todo, nil
}
```

## 监控和追踪

### 关联ID

```go
type BaseEvent struct {
    EventID       string `json:"event_id"`
    CorrelationID string `json:"correlation_id"`  // 同一业务流程
    CausationID   string `json:"causation_id"`    // 因果关系
}

// 使用
event := TodoCreatedEvent{
    BaseEvent: BaseEvent{
        EventID:       uuid.New().String(),
        CorrelationID: ctx.Value("request_id").(string),  // 从HTTP请求传递
        CausationID:   ctx.Value("request_id").(string),
    },
}
```

### 事件日志

```go
slog.Info("Event published",
    "event_id", event.EventID,
    "event_type", event.EventType,
    "aggregate_id", event.AggregateID,
    "correlation_id", event.CorrelationID,
)

slog.Info("Event consumed",
    "event_id", event.EventID,
    "event_type", event.EventType,
    "handler", "TodoCompletedHandler",
    "processing_time_ms", duration.Milliseconds(),
)
```

## 进阶任务

### 1. 实现事件版本化

```go
type TodoCreatedEventV1 struct {
    TodoID int
    Title  string
}

type TodoCreatedEventV2 struct {
    TodoID      int
    Title       string
    Description string  // 新增字段
}

// 处理器支持多版本
func (h *Handler) Handle(eventData []byte) error {
    var baseEvent BaseEvent
    json.Unmarshal(eventData, &baseEvent)

    switch baseEvent.EventVersion {
    case "v1":
        var event TodoCreatedEventV1
        // 处理V1
    case "v2":
        var event TodoCreatedEventV2
        // 处理V2
    }
}
```

### 2. 实现 CQRS + Event Sourcing

```
Command → Generate Events → Save to Event Store → Publish
                                    ↓
Query ← Read Model ← Event Projection ← Subscribe
```

### 3. 添加分布式追踪

使用 OpenTelemetry：
```go
import "go.opentelemetry.io/otel"

tracer := otel.Tracer("todo-service")
ctx, span := tracer.Start(ctx, "publish-event")
defer span.End()

publisher.Publish(ctx, topic, event)
```

## 对比微服务架构

| 维度 | V5 微服务 | V6 事件驱动 |
|------|-----------|-------------|
| 通信方式 | 同步HTTP | 异步消息 |
| 耦合度 | 服务间耦合 | 完全解耦 |
| 性能 | 同步等待 | 非阻塞 |
| 故障影响 | 级联故障 | 隔离 |
| 一致性 | 强一致 | 最终一致 |
| 可追踪性 | 简单 | 复杂 |

## 完成后

恭喜！你已经完成了所有6个架构的学习！

### 总结

```
V1: 单体        → 快速开发
V2: 分层        → 关注点分离
V3: 六边形      → 业务与技术隔离
V4: CQRS        → 读写优化
V5: 微服务      → 独立扩展
V6: 事件驱动    → 完全解耦
```

### 思考题

1. 💭 什么时候应该使用事件驱动？
2. 💭 如何保证事件的顺序性？
3. 💭 最终一致性对业务有什么影响？
4. 💭 如何选择合适的架构模式？

### 下一步

- 阅读 [架构对比指南](../docs/architecture-comparison.md)
- 在实际项目中应用所学架构
- 深入学习 DDD、Event Sourcing、CQRS
- 研究大型系统的架构设计

---

**架构学习之路没有终点，持续实践和思考！**

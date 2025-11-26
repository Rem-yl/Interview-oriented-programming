# V6: 事件驱动架构 (Event-Driven Architecture)

## 架构概述

事件驱动架构（EDA）是一种基于**事件的产生、检测、消费和响应**的架构模式。系统通过异步消息传递来解耦组件，每个组件对感兴趣的事件做出响应，而不需要知道事件的产生者。

## 从 V5 到 V6 的演进

### V5 的问题回顾

```go
// V5: 同步调用，紧耦合
func (s *TodoService) CompleteTodo(id int) error {
    todo, _ := s.repo.FindByID(id)
    todo.Complete()
    s.repo.Update(todo)

    // 同步调用其他服务
    user, _ := s.userClient.GetUser(todo.UserID)  // 阻塞
    s.notifyClient.Send(user.Email, "Todo completed")  // 阻塞

    return nil
}
```

**问题**：
- ❌ 服务间紧耦合（User/Notify挂了影响Todo）
- ❌ 性能差（同步等待）
- ❌ 级联故障（一个服务慢导致全部慢）
- ❌ 难以扩展（添加新功能需要修改主流程）

### V6 的解决方案

```go
// V6: 事件驱动，解耦
func (s *TodoService) CompleteTodo(id int) error {
    todo, _ := s.repo.FindByID(id)
    todo.Complete()
    s.repo.Update(todo)

    // 发布事件，立即返回
    event := TodoCompletedEvent{
        TodoID: todo.ID,
        UserID: todo.UserID,
        Title:  todo.Title,
    }
    s.eventBus.Publish("todo.completed", event)

    return nil  // 不等待其他服务
}

// 其他服务订阅事件
func (n *NotificationService) OnTodoCompleted(event TodoCompletedEvent) {
    user := n.userRepo.FindByID(event.UserID)
    n.sendEmail(user.Email, "Todo completed")
}
```

**优点**：
- ✅ 完全解耦（服务不知道彼此存在）
- ✅ 异步处理（不阻塞主流程）
- ✅ 容易扩展（添加订阅者即可）
- ✅ 故障隔离（订阅者失败不影响发布者）

## 设计目标

- ✅ 松耦合（发布者/订阅者互不依赖）
- ✅ 高扩展性（添加新功能无需改动现有代码）
- ✅ 异步处理（提升性能）
- ✅ 事件溯源（可追溯历史）
- ✅ 最终一致性

## 项目结构

```
v6-event-driven/
├── README.md
├── docker-compose.yml
│
├── shared/                     # 共享事件定义
│   ├── events/
│   │   ├── todo_events.go      # Todo相关事件
│   │   ├── user_events.go      # User相关事件
│   │   └── base_event.go       # 基础事件
│   └── eventbus/
│       ├── publisher.go        # 发布者接口
│       └── subscriber.go       # 订阅者接口
│
├── todo-service/               # 待办服务（发布事件）
│   ├── internal/
│   │   ├── domain/
│   │   │   └── todo.go
│   │   ├── application/
│   │   │   ├── command_handler.go
│   │   │   └── event_publisher.go
│   │   └── infrastructure/
│   │       ├── persistence/
│   │       └── kafka/
│   │           └── publisher.go
│   └── main.go
│
├── notification-service/       # 通知服务（订阅事件）
│   ├── internal/
│   │   ├── application/
│   │   │   └── event_handlers/
│   │   │       ├── todo_completed_handler.go
│   │   │       └── todo_created_handler.go
│   │   └── infrastructure/
│   │       └── kafka/
│   │           └── consumer.go
│   └── main.go
│
├── analytics-service/          # 分析服务（订阅事件）
│   ├── internal/
│   │   ├── application/
│   │   │   └── event_handlers/
│   │   │       └── todo_events_handler.go
│   │   └── infrastructure/
│   │       └── timeseries_db/
│   └── main.go
│
└── event-store/                # 事件存储（可选）
    ├── main.go
    └── internal/
        └── eventstore/
            └── store.go
```

## 核心概念

### 1. 事件（Event）

#### 事件定义

```go
// shared/events/base_event.go
package events

import (
    "time"
    "github.com/google/uuid"
)

// BaseEvent 所有事件的基础结构
type BaseEvent struct {
    EventID        string    `json:"event_id"`         // 事件唯一ID
    EventType      string    `json:"event_type"`       // 事件类型
    AggregateID    string    `json:"aggregate_id"`     // 聚合根ID
    AggregateType  string    `json:"aggregate_type"`   // 聚合根类型
    Timestamp      time.Time `json:"timestamp"`        // 发生时间
    CorrelationID  string    `json:"correlation_id"`   // 关联ID（追踪）
    CausationID    string    `json:"causation_id"`     // 因果ID
    Metadata       map[string]string `json:"metadata"` // 元数据
}

func NewBaseEvent(eventType, aggregateID, aggregateType string) BaseEvent {
    return BaseEvent{
        EventID:       uuid.New().String(),
        EventType:     eventType,
        AggregateID:   aggregateID,
        AggregateType: aggregateType,
        Timestamp:     time.Now(),
        Metadata:      make(map[string]string),
    }
}
```

#### 领域事件

```go
// shared/events/todo_events.go
package events

// TodoCreatedEvent 待办创建事件
type TodoCreatedEvent struct {
    BaseEvent
    TodoID      int    `json:"todo_id"`
    UserID      int    `json:"user_id"`
    Title       string `json:"title"`
    Description string `json:"description"`
}

func NewTodoCreatedEvent(todoID, userID int, title, description string) TodoCreatedEvent {
    return TodoCreatedEvent{
        BaseEvent:   NewBaseEvent("todo.created", fmt.Sprintf("%d", todoID), "todo"),
        TodoID:      todoID,
        UserID:      userID,
        Title:       title,
        Description: description,
    }
}

// TodoCompletedEvent 待办完成事件
type TodoCompletedEvent struct {
    BaseEvent
    TodoID       int       `json:"todo_id"`
    UserID       int       `json:"user_id"`
    Title        string    `json:"title"`
    CompletedAt  time.Time `json:"completed_at"`
}

// TodoDeletedEvent 待办删除事件
type TodoDeletedEvent struct {
    BaseEvent
    TodoID int `json:"todo_id"`
    UserID int `json:"user_id"`
}

// TodoPriorityChangedEvent 优先级变更事件
type TodoPriorityChangedEvent struct {
    BaseEvent
    TodoID      int    `json:"todo_id"`
    OldPriority string `json:"old_priority"`
    NewPriority string `json:"new_priority"`
}
```

**事件命名规范**：
- 使用过去式（已发生的事实）
- 格式：`<聚合根>.<动作>Event`
- 示例：`TodoCreated`, `UserRegistered`, `OrderPlaced`

### 2. 事件总线（Event Bus）

#### 发布者接口

```go
// shared/eventbus/publisher.go
package eventbus

import "context"

type EventPublisher interface {
    Publish(ctx context.Context, topic string, event interface{}) error
    PublishBatch(ctx context.Context, topic string, events []interface{}) error
}
```

#### Kafka 发布者实现

```go
// todo-service/internal/infrastructure/kafka/publisher.go
package kafka

import (
    "context"
    "encoding/json"
    "github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
    writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string) *KafkaPublisher {
    return &KafkaPublisher{
        writer: &kafka.Writer{
            Addr:     kafka.TCP(brokers...),
            Balancer: &kafka.LeastBytes{},
        },
    }
}

func (p *KafkaPublisher) Publish(ctx context.Context, topic string, event interface{}) error {
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }

    message := kafka.Message{
        Topic: topic,
        Value: data,
        // 使用事件ID作为Key，保证同一聚合根的事件顺序
        Key: []byte(getAggregateID(event)),
    }

    return p.writer.WriteMessages(ctx, message)
}

func (p *KafkaPublisher) Close() error {
    return p.writer.Close()
}
```

#### 订阅者接口

```go
// shared/eventbus/subscriber.go
package eventbus

import "context"

type EventHandler func(ctx context.Context, event interface{}) error

type EventSubscriber interface {
    Subscribe(topic string, handler EventHandler) error
    Start(ctx context.Context) error
    Stop() error
}
```

#### Kafka 订阅者实现

```go
// notification-service/internal/infrastructure/kafka/consumer.go
package kafka

import (
    "context"
    "encoding/json"
    "github.com/segmentio/kafka-go"
)

type KafkaSubscriber struct {
    reader   *kafka.Reader
    handlers map[string]eventbus.EventHandler
}

func NewKafkaSubscriber(brokers []string, groupID, topic string) *KafkaSubscriber {
    return &KafkaSubscriber{
        reader: kafka.NewReader(kafka.ReaderConfig{
            Brokers:  brokers,
            GroupID:  groupID,
            Topic:    topic,
            MinBytes: 10e3, // 10KB
            MaxBytes: 10e6, // 10MB
        }),
        handlers: make(map[string]eventbus.EventHandler),
    }
}

func (s *KafkaSubscriber) Subscribe(eventType string, handler eventbus.EventHandler) error {
    s.handlers[eventType] = handler
    return nil
}

func (s *KafkaSubscriber) Start(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            message, err := s.reader.ReadMessage(ctx)
            if err != nil {
                return err
            }

            // 解析事件
            var baseEvent events.BaseEvent
            if err := json.Unmarshal(message.Value, &baseEvent); err != nil {
                continue
            }

            // 根据事件类型调用对应的处理器
            if handler, ok := s.handlers[baseEvent.EventType]; ok {
                if err := handler(ctx, message.Value); err != nil {
                    // 处理失败，记录日志或重试
                    log.Printf("Failed to handle event: %v", err)
                }
            }
        }
    }
}
```

### 3. 事件发布（Producer）

```go
// todo-service/internal/application/command_handler.go
package application

import (
    "context"
    "todo-service/internal/domain"
    "shared/events"
    "shared/eventbus"
)

type CreateTodoHandler struct {
    todoRepo  TodoRepository
    publisher eventbus.EventPublisher
}

func (h *CreateTodoHandler) Handle(ctx context.Context, cmd CreateTodoCommand) error {
    // 1. 执行领域逻辑
    todo, err := domain.NewTodo(cmd.Title, cmd.Description)
    if err != nil {
        return err
    }

    // 2. 持久化
    if err := h.todoRepo.Save(todo); err != nil {
        return err
    }

    // 3. 发布事件
    event := events.NewTodoCreatedEvent(
        todo.ID(),
        cmd.UserID,
        todo.Title(),
        todo.Description(),
    )

    return h.publisher.Publish(ctx, "todo-events", event)
}
```

### 4. 事件订阅（Consumer）

```go
// notification-service/internal/application/event_handlers/todo_completed_handler.go
package event_handlers

import (
    "context"
    "encoding/json"
    "shared/events"
)

type TodoCompletedHandler struct {
    emailSender EmailSender
    userRepo    UserRepository
}

func (h *TodoCompletedHandler) Handle(ctx context.Context, eventData interface{}) error {
    // 1. 反序列化事件
    var event events.TodoCompletedEvent
    if err := json.Unmarshal(eventData.([]byte), &event); err != nil {
        return err
    }

    // 2. 获取用户信息
    user, err := h.userRepo.FindByID(event.UserID)
    if err != nil {
        return err
    }

    // 3. 发送通知
    message := fmt.Sprintf("Your todo '%s' has been completed!", event.Title)
    return h.emailSender.Send(user.Email, "Todo Completed", message)
}
```

### 5. 事件存储（Event Store）

#### 事件持久化

```go
// event-store/internal/eventstore/store.go
package eventstore

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"
)

type Event struct {
    ID            int64
    EventID       string
    EventType     string
    AggregateID   string
    AggregateType string
    EventData     json.RawMessage
    Metadata      json.RawMessage
    Timestamp     time.Time
}

type EventStore interface {
    Save(ctx context.Context, event Event) error
    GetByAggregateID(ctx context.Context, aggregateID string) ([]Event, error)
    GetByEventType(ctx context.Context, eventType string, since time.Time) ([]Event, error)
}

type PostgresEventStore struct {
    db *sql.DB
}

func (s *PostgresEventStore) Save(ctx context.Context, event Event) error {
    query := `
        INSERT INTO events (event_id, event_type, aggregate_id, aggregate_type, event_data, metadata, timestamp)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

    _, err := s.db.ExecContext(ctx, query,
        event.EventID,
        event.EventType,
        event.AggregateID,
        event.AggregateType,
        event.EventData,
        event.Metadata,
        event.Timestamp,
    )

    return err
}

func (s *PostgresEventStore) GetByAggregateID(ctx context.Context, aggregateID string) ([]Event, error) {
    query := `
        SELECT event_id, event_type, aggregate_id, aggregate_type, event_data, metadata, timestamp
        FROM events
        WHERE aggregate_id = $1
        ORDER BY timestamp ASC
    `

    rows, err := s.db.QueryContext(ctx, query, aggregateID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var events []Event
    for rows.Next() {
        var event Event
        if err := rows.Scan(
            &event.EventID,
            &event.EventType,
            &event.AggregateID,
            &event.AggregateType,
            &event.EventData,
            &event.Metadata,
            &event.Timestamp,
        ); err != nil {
            return nil, err
        }
        events = append(events, event)
    }

    return events, nil
}
```

#### 事件溯源（Event Sourcing）

```go
// 从事件流重建聚合根状态
func RebuildTodoFromEvents(events []Event) (*domain.Todo, error) {
    var todo *domain.Todo

    for _, event := range events {
        switch event.EventType {
        case "todo.created":
            var e events.TodoCreatedEvent
            json.Unmarshal(event.EventData, &e)
            todo = domain.NewTodoFromEvent(e)

        case "todo.completed":
            var e events.TodoCompletedEvent
            json.Unmarshal(event.EventData, &e)
            todo.ApplyCompleted(e)

        case "todo.priority_changed":
            var e events.TodoPriorityChangedEvent
            json.Unmarshal(event.EventData, &e)
            todo.ApplyPriorityChanged(e)
        }
    }

    return todo, nil
}

// 领域对象应用事件
func (t *Todo) ApplyCompleted(event events.TodoCompletedEvent) {
    t.status = StatusCompleted
    t.completedAt = event.CompletedAt
    t.updatedAt = event.Timestamp
}
```

## 常见模式

### 1. 事件通知（Event Notification）

最简单的模式，只传递"发生了什么"：

```go
type TodoCompletedEvent struct {
    TodoID int       `json:"todo_id"`
    Time   time.Time `json:"time"`
}
```

订阅者需要自己查询详细信息。

### 2. 事件携带数据（Event-Carried State Transfer）

事件包含完整数据，订阅者无需查询：

```go
type TodoCompletedEvent struct {
    TodoID      int       `json:"todo_id"`
    UserID      int       `json:"user_id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    CompletedAt time.Time `json:"completed_at"`
    // 包含所有需要的数据
}
```

### 3. 事件溯源（Event Sourcing）

不存储当前状态，只存储事件流：

```go
// 不存储 Todo 的当前状态
// 而是存储所有发生的事件
events := []Event{
    TodoCreatedEvent{...},
    TodoPriorityChangedEvent{...},
    TodoCompletedEvent{...},
}

// 通过重放事件来重建状态
todo := RebuildFromEvents(events)
```

### 4. CQRS + Event Sourcing

```
Command Side                         Query Side
     │                                   │
     ├──> Execute Command                │
     ├──> Generate Events                │
     ├──> Save to Event Store ───────────┼──> Build Read Model
     └──> Publish Events                 │
                                         ├──> Update Materialized View
                                         └──> Serve Queries
```

## 处理模式

### 1. 至少一次（At Least Once）

```go
// Kafka 默认行为
func (h *Handler) Handle(event Event) error {
    // 处理事件
    err := h.process(event)

    // 只有成功后才提交偏移量
    if err == nil {
        h.consumer.CommitMessages(event.Message)
    }

    return err
}
```

**特点**：
- 可能重复处理
- 需要幂等性保证

### 2. 幂等性处理

```go
type EventHandler struct {
    processedEvents map[string]bool  // 或使用Redis
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
    if err := h.process(event); err != nil {
        return err
    }

    // 标记为已处理
    h.processedEvents[event.EventID] = true
    return nil
}
```

### 3. 消费者组（Consumer Group）

```go
// 多个实例组成一个消费者组
// Kafka 自动分配分区
consumer1 := NewKafkaSubscriber(brokers, "notification-group", "todo-events")
consumer2 := NewKafkaSubscriber(brokers, "notification-group", "todo-events")

// 分区0 → consumer1
// 分区1 → consumer2
// 保证每个事件只被组内一个实例处理
```

## 错误处理

### 1. 死信队列（Dead Letter Queue）

```go
func (h *Handler) Handle(event Event) error {
    if err := h.process(event); err != nil {
        // 重试3次
        for i := 0; i < 3; i++ {
            if err = h.process(event); err == nil {
                return nil
            }
            time.Sleep(time.Second * time.Duration(i+1))
        }

        // 仍然失败，发送到死信队列
        h.dlq.Send(event, err)
        return nil  // 不阻塞后续消息
    }

    return nil
}
```

### 2. 补偿事务（Saga）

```go
// 订单流程
events := []Event{
    OrderCreatedEvent{},
    PaymentProcessedEvent{},  // 如果失败
    PaymentFailedEvent{},     // 触发补偿
    OrderCancelledEvent{},    // 取消订单
    InventoryReleasedEvent{}, // 释放库存
}
```

## 监控和追踪

### 1. 关联ID（Correlation ID）

```go
type BaseEvent struct {
    EventID       string `json:"event_id"`
    CorrelationID string `json:"correlation_id"`  // 同一业务流程
    CausationID   string `json:"causation_id"`    // 因果关系
}

// 用户请求 → RequestID: req-123
// TodoCreatedEvent:      CorrelationID=req-123, CausationID=req-123
// NotificationSentEvent: CorrelationID=req-123, CausationID=todo-created-456

// 可追踪整个业务流程
```

### 2. 事件日志

```go
func (p *KafkaPublisher) Publish(ctx context.Context, topic string, event interface{}) error {
    start := time.Now()

    err := p.writer.WriteMessages(ctx, message)

    // 记录日志
    log.WithFields(log.Fields{
        "event_type":     event.EventType,
        "aggregate_id":   event.AggregateID,
        "topic":          topic,
        "duration_ms":    time.Since(start).Milliseconds(),
        "success":        err == nil,
    }).Info("Event published")

    return err
}
```

## 优点分析

| 优点 | 说明 | 示例 |
|------|------|------|
| **松耦合** | 发布者/订阅者不知道彼此 | Todo服务不知道谁在监听事件 |
| **高扩展性** | 添加订阅者无需改代码 | 新增Analytics服务，只需订阅 |
| **异步处理** | 提升性能 | 发送通知不阻塞主流程 |
| **故障隔离** | 订阅者失败不影响发布者 | 通知失败不影响待办创建 |
| **审计追踪** | 事件存储完整历史 | 可重放所有操作 |

## 缺点分析

| 缺点 | 说明 | 影响 |
|------|------|------|
| **最终一致性** | 数据可能短暂不一致 | 用户可能看到旧数据 |
| **调试困难** | 异步流程难以追踪 | 需要分布式追踪工具 |
| **事件版本管理** | 事件结构变化麻烦 | 需要版本策略 |
| **消息顺序** | 难以保证全局顺序 | 只能保证分区内顺序 |
| **复杂度高** | 理解和实现成本 | 团队学习曲线 |

## 与之前架构对比

| 维度 | V5 微服务 | V6 事件驱动 |
|------|-----------|-------------|
| **通信方式** | 同步HTTP | 异步消息 |
| **耦合度** | 服务间耦合 | 完全解耦 |
| **性能** | 同步等待 | 非阻塞 |
| **故障影响** | 级联故障 | 隔离 |
| **扩展性** | 需修改代码 | 添加订阅者 |
| **一致性** | 强一致性 | 最终一致性 |
| **复杂度** | 高 | 更高 |

## 何时使用事件驱动

✅ **适合场景**：
- 需要高解耦
- 异步处理场景
- 需要完整审计追踪
- 多个系统需要对同一事件响应
- 最终一致性可接受

❌ **不适合场景**：
- 需要强一致性
- 简单的请求-响应模式
- 实时性要求极高
- 团队不熟悉异步编程

## 最佳实践

### 1. 事件命名

```go
// ✅ 好的命名（过去式）
TodoCreatedEvent
UserRegisteredEvent
OrderShippedEvent

// ❌ 不好的命名
CreateTodoEvent      // 命令，不是事件
TodoEvent           // 太泛化
TodoChange          // 不明确
```

### 2. 事件粒度

```go
// ✅ 细粒度事件（推荐）
TodoCreatedEvent
TodoTitleChangedEvent
TodoCompletedEvent

// ❌ 粗粒度事件
TodoUpdatedEvent  // 不知道具体改了什么
```

### 3. 事件版本化

```go
// V1
type TodoCreatedEventV1 struct {
    TodoID int    `json:"todo_id"`
    Title  string `json:"title"`
}

// V2 添加字段
type TodoCreatedEventV2 struct {
    TodoID      int    `json:"todo_id"`
    Title       string `json:"title"`
    Description string `json:"description"`  // 新增
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

## 练习任务

### 必做任务
1. ✅ 定义领域事件
2. ✅ 实现事件发布和订阅
3. ✅ 使用Kafka作为消息中间件
4. ✅ 实现至少3个事件处理器

### 进阶任务
1. 🔧 实现事件存储（Event Store）
2. 🔧 实现事件溯源（Event Sourcing）
3. 🔧 添加死信队列处理
4. 🔧 实现Saga模式处理分布式事务
5. 🔧 使用OpenTelemetry实现分布式追踪

### 思考题
1. 💭 如何保证事件的顺序性？
2. 💭 如何处理事件处理失败的情况？
3. 💭 事件驱动和消息队列的区别？
4. 💭 如何选择事件的粒度？

## 总结

经过6个版本的演进，我们学习了：

```
V1: 单体架构        → 快速开发，适合原型
V2: 分层架构        → 关注点分离，可测试
V3: 六边形架构      → 业务与技术隔离，充血模型
V4: CQRS架构        → 读写分离，性能优化
V5: 微服务架构      → 独立部署，按需扩展
V6: 事件驱动架构    → 完全解耦，异步处理
```

**关键收获**：
- 没有"最好"的架构，只有"合适"的架构
- 架构应该随业务复杂度演进
- 每种架构都有权衡（Trade-offs）
- 理解原理比记住模式更重要

---

**恭喜完成所有架构的学习！现在开始实践吧！**

# V4: CQRS架构实现

## 开始之前

请先阅读架构设计文档：[V4 CQRS架构](../docs/v4-cqrs.md)

## 实现目标

实现命令查询职责分离（CQRS），读写使用不同的模型和存储。

## 核心概念

- **Command Side（命令侧）**：处理写操作，使用领域模型
- **Query Side（查询侧）**：处理读操作，使用简单DTO
- **Synchronizer（同步器）**：同步写库到读库

## 目录结构

```
v4-cqrs/
├── go.mod
├── main.go
├── domain/                    # 领域模型（Command侧用）
│   └── todo.go
├── application/
│   ├── commands/             # 命令处理器
│   │   ├── create_todo.go
│   │   ├── complete_todo.go
│   │   └── handler.go
│   └── queries/              # 查询处理器
│       ├── get_todo.go
│       ├── list_todos.go
│       ├── statistics.go
│       └── handler.go
├── infrastructure/
│   ├── command_store/        # 写存储
│   │   └── memory/
│   ├── query_store/          # 读存储
│   │   └── memory/
│   └── sync/
│       └── synchronizer.go
├── dto/                      # 数据传输对象
│   ├── todo_dto.go
│   └── statistics_dto.go
└── adapters/
    └── http/
        ├── command_handler.go
        └── query_handler.go
```

## 实现步骤

### Step 1: 定义 Command 和 Query

```go
// Command（写操作）
type CreateTodoCommand struct {
    Title       string
    Description string
}

type CompleteTodoCommand struct {
    TodoID int
}

// Query（读操作）
type GetTodoQuery struct {
    TodoID int
}

type ListTodosQuery struct {
    Status *string
    Page   int
}

type StatisticsQuery struct{}
```

### Step 2: 实现写侧（Command Side）

使用 V3 的领域模型：
- 创建领域对象
- 应用业务规则
- 持久化到写库
- **同步到读库**

### Step 3: 实现读侧（Query Side）

直接查询 DTO，无业务逻辑：
```go
func (h *ListTodosHandler) Handle(query Query) (interface{}, error) {
    return h.queryRepo.List(query)
}
```

### Step 4: 实现同步器

```go
type Synchronizer struct {
    queryRepo QueryRepository
}

func (s *Synchronizer) SyncTodoCreated(todo *domain.Todo) error {
    dto := s.toDTO(todo)
    return s.queryRepo.Insert(dto)
}
```

## 数据库设计

### 写库（规范化）

```sql
CREATE TABLE todos (
    id INTEGER PRIMARY KEY,
    title VARCHAR(100),
    description TEXT,
    status INTEGER,      -- 枚举值
    priority INTEGER,
    created_at DATETIME,
    updated_at DATETIME
);
```

### 读库（反规范化）

```sql
CREATE TABLE todo_read_models (
    id INTEGER PRIMARY KEY,
    title VARCHAR(100),
    description TEXT,
    status_text VARCHAR(20),    -- "pending", "completed"
    priority_text VARCHAR(20),  -- "low", "normal", "high"
    completed BOOLEAN,
    created_at_formatted VARCHAR(20),
    -- 查询优化字段
    search_text TEXT
);

CREATE INDEX idx_status ON todo_read_models(status_text);
CREATE INDEX idx_completed ON todo_read_models(completed);
```

## 同步策略

### 方案1: 同步写入

```go
// 命令处理器内同步
func (h *CreateTodoHandler) Handle(cmd Command) error {
    // 1. 写入主库
    h.writeRepo.Save(todo)

    // 2. 同步写入读库（阻塞）
    h.sync.SyncTodoCreated(todo)

    return nil
}
```

### 方案2: 异步同步（推荐）

```go
// 使用 channel 或消息队列
func (h *CreateTodoHandler) Handle(cmd Command) error {
    h.writeRepo.Save(todo)

    // 发送到队列（非阻塞）
    h.syncQueue <- SyncEvent{Type: "created", Todo: todo}

    return nil
}

// 独立的 goroutine 处理同步
func (s *Synchronizer) Start() {
    for event := range s.syncQueue {
        s.sync(event)
    }
}
```

## 测试

### Command 测试

```go
func TestCreateTodoHandler(t *testing.T) {
    mockWriteRepo := new(MockWriteRepository)
    mockSync := new(MockSynchronizer)

    handler := NewCreateTodoHandler(mockWriteRepo, mockSync)

    cmd := CreateTodoCommand{Title: "Test"}
    err := handler.Handle(cmd)

    assert.NoError(t, err)
    mockWriteRepo.AssertCalled(t, "Save")
    mockSync.AssertCalled(t, "SyncTodoCreated")
}
```

### Query 测试

```go
func TestListTodosHandler(t *testing.T) {
    mockQueryRepo := new(MockQueryRepository)
    handler := NewListTodosHandler(mockQueryRepo)

    result, err := handler.Handle(ListTodosQuery{})

    assert.NoError(t, err)
    mockQueryRepo.AssertCalled(t, "List")
}
```

## 进阶任务

### 1. 实现统计查询

```go
type StatisticsDTO struct {
    TotalTodos     int
    CompletedTodos int
    PendingTodos   int
}

// 在读库中预先计算
func (r *QueryRepository) GetStatistics() (*StatisticsDTO, error) {
    // 使用 SQL 聚合查询
}
```

### 2. 添加缓存层

```go
type CachedQueryRepository struct {
    repo  QueryRepository
    cache *redis.Client
}

func (r *CachedQueryRepository) GetByID(id int) (*TodoDTO, error) {
    // 1. 先查缓存
    if cached, ok := r.cache.Get(id); ok {
        return cached, nil
    }

    // 2. 查数据库
    dto, err := r.repo.GetByID(id)

    // 3. 写入缓存
    r.cache.Set(id, dto)

    return dto, err
}
```

### 3. 实现 Event Sourcing

将所有事件存储到事件流：
```go
type EventStore interface {
    Save(event Event) error
    GetByAggregateID(id string) ([]Event, error)
}
```

## 对比 V3

| 维度 | V3 六边形 | V4 CQRS |
|------|-----------|---------|
| 读写模型 | 统一 | 分离 |
| 查询性能 | 一般 | 优化 |
| 复杂查询 | 困难 | 简单 |
| 一致性 | 强一致 | 最终一致 |

## 完成后

思考以下问题：

1. 如何保证读写库的一致性？
2. 最终一致性对用户体验有什么影响？
3. CQRS 和 Event Sourcing 的区别？

完成后继续学习 [V5: 微服务架构](../docs/v5-microservices.md)

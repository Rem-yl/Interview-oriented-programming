# V4: CQRS架构 (Command Query Responsibility Segregation)

## 架构概述

CQRS（命令查询职责分离）是一种将**数据的读取和写入分离**的架构模式。核心思想是：更新数据的"命令"和查询数据的"查询"使用不同的模型和存储。

## 从 V3 到 V4 的演进

### V3 的问题回顾

```go
// V3: 读写使用相同的领域模型
func (s *Service) CompleteTodo(cmd Command) error {
    todo, _ := s.repo.FindByID(id)  // 加载完整领域对象
    todo.Complete()                  // 执行业务逻辑
    s.repo.Update(todo)              // 保存
    return nil
}

func (s *Service) GetTodoList() []TodoDTO {
    todos, _ := s.repo.FindAll()     // 也加载完整领域对象
    return toDTO(todos)               // 但只是展示，不需要业务逻辑
}
```

**问题**：
- ❌ 读操作不需要领域逻辑，却有领域对象开销
- ❌ 复杂查询（统计、聚合）难以用领域模型表达
- ❌ 读写性能优化方向不同，无法各自优化

### V4 的解决方案

```
┌──────────────────────────────────────────────┐
│               API Layer                      │
└──────┬───────────────────────┬───────────────┘
       │                       │
       ▼                       ▼
┌─────────────┐         ┌─────────────┐
│   Command   │         │    Query    │
│    Side     │         │    Side     │
├─────────────┤         ├─────────────┤
│ Domain      │         │ DTOs        │
│ Models      │         │ (简单对象)  │
│             │         │             │
│ Business    │         │ No Logic    │
│ Logic       │         │ (只读)      │
└─────┬───────┘         └──────┬──────┘
      │                        │
      ▼                        ▼
┌─────────────┐         ┌─────────────┐
│ Write DB    │ sync→   │  Read DB    │
│ (规范化)    │ ─────→  │ (反规范化)  │
└─────────────┘         └─────────────┘
```

**核心原则**：
1. **命令（Command）**：修改状态，使用领域模型
2. **查询（Query）**：读取数据，使用简单DTO
3. **数据同步**：写入后同步到读模型

## 设计目标

- ✅ 读写分离，各自优化
- ✅ 读侧无业务逻辑，性能更好
- ✅ 写侧保留领域模型，业务正确
- ✅ 支持复杂查询和统计
- ✅ 读写可独立扩展

## 目录结构

```
v4-cqrs/
├── go.mod
├── go.sum
├── main.go
├── README.md
│
├── domain/                      # 领域层（Command侧使用）
│   ├── todo.go
│   └── errors.go
│
├── application/
│   ├── commands/               # Command侧（写操作）
│   │   ├── create_todo.go      # 命令处理器
│   │   ├── complete_todo.go
│   │   └── handler.go          # CommandHandler接口
│   │
│   └── queries/                # Query侧（读操作）
│       ├── get_todo.go         # 查询处理器
│       ├── list_todos.go
│       ├── statistics.go       # 统计查询
│       └── handler.go          # QueryHandler接口
│
├── infrastructure/
│   ├── command_store/          # 写存储（规范化）
│   │   ├── repository.go
│   │   └── sqlite/
│   │       └── todo_repository.go
│   │
│   ├── query_store/            # 读存储（反规范化）
│   │   ├── repository.go
│   │   └── sqlite/
│   │       └── todo_query_repository.go
│   │
│   └── sync/                   # 数据同步
│       └── synchronizer.go
│
├── adapters/
│   └── http/
│       ├── command_handler.go  # POST, PUT, DELETE
│       └── query_handler.go    # GET
│
└── dto/                        # 数据传输对象
    ├── todo_dto.go             # 读模型DTO
    └── statistics_dto.go
```

## 核心概念

### 1. Command Side（命令侧）

#### 命令定义

```go
// application/commands/create_todo.go
package commands

// Command 接口
type Command interface{}

// CreateTodoCommand 创建待办命令
type CreateTodoCommand struct {
    Title       string
    Description string
}

// CompleteTodoCommand 完成待办命令
type CompleteTodoCommand struct {
    TodoID int
}

// IncreasePriorityCommand 提升优先级命令
type IncreasePriorityCommand struct {
    TodoID int
}
```

#### 命令处理器

```go
// application/commands/handler.go
package commands

type CommandHandler interface {
    Handle(cmd Command) error
}

// application/commands/create_todo.go
type CreateTodoHandler struct {
    writeRepo  WriteRepository     // 写存储
    queryRepo  QueryRepository     // 读存储
    sync       Synchronizer        // 同步器
}

func (h *CreateTodoHandler) Handle(cmd Command) error {
    createCmd := cmd.(CreateTodoCommand)

    // 1. 创建领域对象（包含业务逻辑）
    title, err := domain.NewTitle(createCmd.Title)
    if err != nil {
        return err
    }

    todo, err := domain.NewTodo(title, createCmd.Description)
    if err != nil {
        return err
    }

    // 2. 持久化到写存储
    if err := h.writeRepo.Save(todo); err != nil {
        return err
    }

    // 3. 同步到读存储
    return h.sync.SyncTodoCreated(todo)
}
```

#### 写存储 Repository

```go
// infrastructure/command_store/repository.go
package command_store

import "your-module/domain"

// WriteRepository 写侧仓储（使用领域模型）
type WriteRepository interface {
    Save(todo *domain.Todo) error
    FindByID(id int) (*domain.Todo, error)
    Update(todo *domain.Todo) error
    Delete(id int) error
}

// 规范化存储（第三范式）
// 表结构：todos(id, title, description, status, priority, created_at, updated_at)
```

### 2. Query Side（查询侧）

#### 查询定义

```go
// application/queries/get_todo.go
package queries

// Query 接口
type Query interface{}

// GetTodoQuery 获取单个待办查询
type GetTodoQuery struct {
    TodoID int
}

// ListTodosQuery 列表查询
type ListTodosQuery struct {
    Status    *string
    Priority  *string
    Completed *bool
    Page      int
    PageSize  int
}

// StatisticsQuery 统计查询
type StatisticsQuery struct {
    UserID int
}
```

#### 查询处理器

```go
// application/queries/handler.go
package queries

type QueryHandler interface {
    Handle(query Query) (interface{}, error)
}

// application/queries/list_todos.go
type ListTodosHandler struct {
    queryRepo QueryRepository
}

func (h *ListTodosHandler) Handle(query Query) (interface{}, error) {
    listQuery := query.(ListTodosQuery)

    // 直接从读存储查询（无领域逻辑）
    return h.queryRepo.List(listQuery)
}
```

#### 读存储 Repository

```go
// infrastructure/query_store/repository.go
package query_store

import "your-module/dto"

// QueryRepository 读侧仓储（使用DTO）
type QueryRepository interface {
    GetByID(id int) (*dto.TodoDTO, error)
    List(query queries.ListTodosQuery) ([]*dto.TodoDTO, error)
    GetStatistics(userID int) (*dto.StatisticsDTO, error)
}

// 反规范化存储（为查询优化）
// 表结构：todo_read_models(id, title, description, status_text,
//                          priority_text, completed, created_at, ...)
```

#### DTO（数据传输对象）

```go
// dto/todo_dto.go
package dto

// TodoDTO 读模型（扁平化，便于查询和展示）
type TodoDTO struct {
    ID          int    `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Status      string `json:"status"`       // "pending", "completed"
    Priority    string `json:"priority"`     // "low", "normal", "high"
    Completed   bool   `json:"completed"`
    CreatedAt   string `json:"created_at"`   // 格式化后的时间
    UpdatedAt   string `json:"updated_at"`
}

// StatisticsDTO 统计信息
type StatisticsDTO struct {
    TotalTodos      int `json:"total_todos"`
    CompletedTodos  int `json:"completed_todos"`
    PendingTodos    int `json:"pending_todos"`
    HighPriorityTodos int `json:"high_priority_todos"`
}
```

### 3. 数据同步

#### Synchronizer（同步器）

```go
// infrastructure/sync/synchronizer.go
package sync

import (
    "your-module/domain"
    "your-module/dto"
    "your-module/infrastructure/query_store"
)

type Synchronizer struct {
    queryRepo query_store.QueryRepository
}

func NewSynchronizer(queryRepo query_store.QueryRepository) *Synchronizer {
    return &Synchronizer{queryRepo: queryRepo}
}

// SyncTodoCreated 同步新创建的待办
func (s *Synchronizer) SyncTodoCreated(todo *domain.Todo) error {
    // 将领域对象转换为DTO
    dto := &dto.TodoDTO{
        ID:          todo.ID().Value(),
        Title:       todo.Title().String(),
        Description: todo.Description(),
        Status:      s.statusToString(todo.Status()),
        Priority:    s.priorityToString(todo.Priority()),
        Completed:   todo.Status() == domain.StatusCompleted,
        CreatedAt:   todo.CreatedAt().Format("2006-01-02 15:04:05"),
        UpdatedAt:   todo.UpdatedAt().Format("2006-01-02 15:04:05"),
    }

    // 插入到读存储
    return s.queryRepo.Insert(dto)
}

// SyncTodoUpdated 同步更新
func (s *Synchronizer) SyncTodoUpdated(todo *domain.Todo) error {
    dto := s.toDTO(todo)
    return s.queryRepo.Update(dto)
}

// SyncTodoDeleted 同步删除
func (s *Synchronizer) SyncTodoDeleted(todoID int) error {
    return s.queryRepo.Delete(todoID)
}
```

## HTTP 适配器

### Command Handler（写操作）

```go
// adapters/http/command_handler.go
package http

import (
    "github.com/gin-gonic/gin"
    "your-module/application/commands"
)

type CommandHandlers struct {
    createHandler   *commands.CreateTodoHandler
    completeHandler *commands.CompleteTodoHandler
}

func (h *CommandHandlers) CreateTodo(c *gin.Context) {
    var req struct {
        Title       string `json:"title"`
        Description string `json:"description"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 构建命令
    cmd := commands.CreateTodoCommand{
        Title:       req.Title,
        Description: req.Description,
    }

    // 执行命令
    if err := h.createHandler.Handle(cmd); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 命令执行成功（不返回数据，客户端重新查询）
    c.JSON(201, gin.H{"message": "created"})
}
```

### Query Handler（读操作）

```go
// adapters/http/query_handler.go
package http

import (
    "github.com/gin-gonic/gin"
    "your-module/application/queries"
)

type QueryHandlers struct {
    getHandler  *queries.GetTodoHandler
    listHandler *queries.ListTodosHandler
    statsHandler *queries.StatisticsHandler
}

func (h *QueryHandlers) GetTodo(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))

    // 构建查询
    query := queries.GetTodoQuery{TodoID: id}

    // 执行查询
    result, err := h.getHandler.Handle(query)
    if err != nil {
        c.JSON(404, gin.H{"error": "not found"})
        return
    }

    // 直接返回DTO
    c.JSON(200, result)
}

func (h *QueryHandlers) GetStatistics(c *gin.Context) {
    query := queries.StatisticsQuery{}

    result, err := h.statsHandler.Handle(query)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, result)
}
```

## 数据库设计

### 写存储（规范化）

```sql
-- 规范化设计，便于维护数据完整性
CREATE TABLE todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    status INTEGER NOT NULL,      -- 枚举值
    priority INTEGER NOT NULL,    -- 枚举值
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);
```

### 读存储（反规范化）

```sql
-- 反规范化设计，便于查询性能
CREATE TABLE todo_read_models (
    id INTEGER PRIMARY KEY,
    title VARCHAR(100),
    description TEXT,
    status_text VARCHAR(20),       -- "pending", "completed"
    priority_text VARCHAR(20),     -- "low", "normal", "high"
    completed BOOLEAN,
    created_at_formatted VARCHAR(20),
    updated_at_formatted VARCHAR(20),

    -- 额外的查询优化字段
    search_text TEXT,              -- 全文搜索
    sort_order INTEGER             -- 排序字段
);

-- 查询优化索引
CREATE INDEX idx_status ON todo_read_models(status_text);
CREATE INDEX idx_priority ON todo_read_models(priority_text);
CREATE INDEX idx_completed ON todo_read_models(completed);

-- 统计视图
CREATE VIEW todo_statistics AS
SELECT
    COUNT(*) as total_todos,
    SUM(CASE WHEN completed = 1 THEN 1 ELSE 0 END) as completed_todos,
    SUM(CASE WHEN completed = 0 THEN 1 ELSE 0 END) as pending_todos,
    SUM(CASE WHEN priority_text = 'high' THEN 1 ELSE 0 END) as high_priority_todos
FROM todo_read_models;
```

## 同步策略

### 1. 同步写入（Sync）

```go
func (h *CreateTodoHandler) Handle(cmd Command) error {
    // 写入主存储
    h.writeRepo.Save(todo)

    // 同步写入读存储（事务内）
    h.sync.SyncTodoCreated(todo)  // 阻塞

    return nil
}
```

**优点**：强一致性
**缺点**：性能较差

### 2. 异步同步（Async - 推荐）

```go
func (h *CreateTodoHandler) Handle(cmd Command) error {
    // 写入主存储
    h.writeRepo.Save(todo)

    // 发送事件到队列（非阻塞）
    h.eventBus.Publish(TodoCreatedEvent{Todo: todo})

    return nil
}

// 独立的事件处理器
func (s *Synchronizer) OnTodoCreated(event TodoCreatedEvent) {
    s.queryRepo.Insert(toDTO(event.Todo))
}
```

**优点**：性能好，解耦
**缺点**：最终一致性

## 测试策略

### Command侧测试

```go
// 测试业务逻辑
func TestCompleteTodoHandler(t *testing.T) {
    mockWriteRepo := new(MockWriteRepository)
    mockSync := new(MockSynchronizer)

    handler := NewCompleteTodoHandler(mockWriteRepo, mockSync)

    // 模拟已存在的待办
    todo := &domain.Todo{...}
    mockWriteRepo.On("FindByID", 1).Return(todo, nil)
    mockWriteRepo.On("Update", todo).Return(nil)
    mockSync.On("SyncTodoUpdated", todo).Return(nil)

    cmd := CompleteTodoCommand{TodoID: 1}
    err := handler.Handle(cmd)

    assert.NoError(t, err)
    assert.Equal(t, domain.StatusCompleted, todo.Status())
}
```

### Query侧测试

```go
// 测试查询逻辑
func TestListTodosHandler(t *testing.T) {
    mockQueryRepo := new(MockQueryRepository)
    handler := NewListTodosHandler(mockQueryRepo)

    expectedDTOs := []*dto.TodoDTO{...}
    mockQueryRepo.On("List", mock.Anything).Return(expectedDTOs, nil)

    query := ListTodosQuery{Status: "pending"}
    result, err := handler.Handle(query)

    assert.NoError(t, err)
    assert.Len(t, result, 2)
}
```

## 优点分析

| 优点 | 说明 | 示例 |
|------|------|------|
| **读写分离** | 各自优化 | 写侧规范化，读侧反规范化 |
| **性能优化** | 读侧无业务逻辑 | 直接查询DTO，无ORM映射 |
| **复杂查询** | 支持统计、聚合 | 预先计算统计数据 |
| **独立扩展** | 读写独立扩展 | 读库可以有多个副本 |
| **业务清晰** | 命令明确业务意图 | CreateTodo vs UpdateTodo |

## 缺点分析

| 缺点 | 说明 | 影响 |
|------|------|------|
| **复杂度高** | 两套模型 | 学习和维护成本 |
| **数据同步** | 需要保持一致性 | 可能出现延迟 |
| **最终一致性** | 读写可能不一致 | 用户可能看到旧数据 |
| **代码量大** | 命令/查询处理器 | 开发时间增加 |

## 与 V3 对比

| 维度 | V3 六边形 | V4 CQRS |
|------|-----------|---------|
| **模型** | 统一领域模型 | 读写分离模型 |
| **查询** | 加载领域对象 | 直接查询DTO |
| **复杂查询** | 困难 | 简单（预计算） |
| **一致性** | 强一致性 | 最终一致性 |
| **性能** | 读写相同 | 读写各自优化 |
| **复杂度** | 高 | 更高 |

## 何时使用 CQRS

✅ **适合场景**：
- 读写比例悬殊（读多写少）
- 需要复杂查询和统计
- 读写性能要求不同
- 需要独立扩展读写

❌ **不适合场景**：
- 简单CRUD应用
- 实时一致性要求高
- 团队不熟悉CQRS
- 读写操作相似

## 演进到 V5 的动机

当单体应用遇到以下问题：

### 问题 1: 单点故障
```
应用崩溃 → 整个系统不可用
```

### 问题 2: 扩展性限制
```
某个功能需要更多资源 → 整个应用都要扩展
```

### 问题 3: 技术栈限制
```
所有功能必须用同一种语言和框架
```

### 问题 4: 团队协作
```
多个团队修改同一代码库 → 冲突频繁
```

**这时候，我们需要微服务架构！**

## 练习任务

### 必做任务
1. ✅ 实现 Command 和 Query 分离
2. ✅ 实现读写两套存储
3. ✅ 实现同步机制
4. ✅ 实现统计查询功能

### 进阶任务
1. 🔧 使用消息队列实现异步同步
2. 🔧 实现事件溯源（Event Sourcing）
3. 🔧 添加缓存层（Redis）到读侧
4. 🔧 实现读库的主从复制

### 思考题
1. 💭 如何处理同步失败的情况？
2. 💭 最终一致性对用户体验有什么影响？
3. 💭 CQRS 和 Event Sourcing 的区别？

---

**完成 V4 后，继续学习 [V5: 微服务架构](./v5-microservices.md)**

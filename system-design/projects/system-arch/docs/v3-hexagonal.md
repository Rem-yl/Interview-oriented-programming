# V3: 六边形架构 (Hexagonal Architecture / Ports & Adapters)

## 架构概述

六边形架构（又称端口和适配器架构）由 Alistair Cockburn 提出，核心理念是**业务逻辑与技术实现完全隔离**。应用的核心（业务逻辑）位于中心，通过"端口"与外部世界交互，"适配器"负责具体的技术实现。

## 从 V2 到 V3 的演进

### V2 的问题回顾

```go
// V2: Service 依赖 Repository 接口
type TodoService struct {
    repo repository.TodoRepository  // 依赖技术层接口
}

// Repository 接口受数据库影响
type TodoRepository interface {
    FindByID(id int) (*Todo, error)  // 假设有数据库主键
    Update(todo *Todo) error         // 假设直接保存整个对象
}
```

**问题**：
- ❌ 业务层依赖了技术层的概念（Repository）
- ❌ Model 是贫血对象，只有数据没有行为
- ❌ 业务规则分散在 Service 的各个方法中

### V3 的解决方案

```
        ┌─────────────────────────────┐
        │     Driving Adapters        │
        │  (HTTP, gRPC, CLI, Tests)   │
        └──────────────┬──────────────┘
                       │
        ┌──────────────▼──────────────┐
        │      Driving Ports          │  ← 接口由核心定义
        │   (Application Services)    │
        ├─────────────────────────────┤
        │                             │
        │    Domain Core (核心)       │
        │  - Entities (充血模型)      │
        │  - Value Objects            │
        │  - Domain Services          │
        │                             │
        ├─────────────────────────────┤
        │      Driven Ports           │  ← 接口由核心定义
        │    (Repository, etc)        │
        └──────────────┬──────────────┘
                       │
        ┌──────────────▼──────────────┐
        │     Driven Adapters         │
        │  (Memory, SQLite, Postgres) │
        └─────────────────────────────┘
```

**核心原则**：
1. **依赖倒置**：外层依赖内层，内层不依赖外层
2. **端口由内部定义**：业务逻辑定义需要什么接口
3. **适配器实现端口**：技术细节实现业务接口

## 设计目标

- ✅ 业务逻辑与技术实现完全隔离
- ✅ 领域模型富含行为（充血模型）
- ✅ 可测试性极高（核心逻辑无技术依赖）
- ✅ 技术实现可自由替换
- ✅ 遵循 SOLID 原则

## 目录结构

```
v3-hexagonal/
├── go.mod
├── go.sum
├── main.go                      # 程序入口，组装适配器
├── README.md
│
├── domain/                      # 核心领域（最内层）
│   ├── todo.go                  # 领域实体（充血模型）
│   ├── value_objects.go         # 值对象
│   └── errors.go                # 领域错误
│
├── application/                 # 应用层（编排领域对象）
│   ├── ports/                   # 端口定义（接口）
│   │   ├── input/              # Driving Ports（入口）
│   │   │   └── todo_service.go  # 应用服务接口
│   │   └── output/             # Driven Ports（出口）
│   │       └── todo_repository.go  # 持久化接口
│   │
│   └── services/               # 应用服务实现
│       └── todo_service_impl.go
│
└── adapters/                    # 适配器（最外层）
    ├── input/                   # Driving Adapters（入站）
    │   ├── http/
    │   │   ├── router.go
    │   │   └── todo_handler.go  # HTTP 适配器
    │   └── cli/
    │       └── todo_cli.go      # CLI 适配器
    │
    └── output/                  # Driven Adapters（出站）
        ├── persistence/
        │   ├── memory/
        │   │   └── todo_memory.go
        │   └── sqlite/
        │       └── todo_sqlite.go
        └── notification/
            └── email_notifier.go
```

## 核心概念

### 1. Domain Layer（领域层）- 核心

#### 充血模型（Rich Domain Model）

```go
// domain/todo.go
package domain

import (
    "errors"
    "time"
)

// Todo 领域实体（充血模型，包含业务行为）
type Todo struct {
    id          TodoID
    title       Title
    description string
    status      TodoStatus
    priority    Priority
    createdAt   time.Time
    updatedAt   time.Time
}

// 工厂方法
func NewTodo(title Title, description string) (*Todo, error) {
    return &Todo{
        id:          NewTodoID(),
        title:       title,
        description: description,
        status:      StatusPending,
        priority:    PriorityNormal,
        createdAt:   time.Now(),
        updatedAt:   time.Now(),
    }, nil
}

// 业务方法：完成待办
func (t *Todo) Complete() error {
    if t.status == StatusCompleted {
        return ErrAlreadyCompleted
    }

    // 业务规则：高优先级需要审批
    if t.priority == PriorityHigh {
        return ErrHighPriorityNeedsApproval
    }

    t.status = StatusCompleted
    t.updatedAt = time.Now()
    return nil
}

// 业务方法：提升优先级
func (t *Todo) IncreasePriority() error {
    if t.status == StatusCompleted {
        return ErrCannotChangePriorityWhenCompleted
    }

    switch t.priority {
    case PriorityLow:
        t.priority = PriorityNormal
    case PriorityNormal:
        t.priority = PriorityHigh
    default:
        return ErrAlreadyHighestPriority
    }

    t.updatedAt = time.Now()
    return nil
}

// Getter 方法（封装内部状态）
func (t *Todo) ID() TodoID         { return t.id }
func (t *Todo) Title() Title       { return t.title }
func (t *Todo) Status() TodoStatus { return t.status }
```

#### 值对象（Value Objects）

```go
// domain/value_objects.go
package domain

import (
    "errors"
    "strings"
)

// TodoID 值对象
type TodoID struct {
    value int
}

func NewTodoID() TodoID {
    // 简化版，实际应用中可能用 UUID
    return TodoID{value: generateID()}
}

func (id TodoID) Value() int {
    return id.value
}

// Title 值对象（带业务验证）
type Title struct {
    value string
}

func NewTitle(value string) (Title, error) {
    value = strings.TrimSpace(value)

    if value == "" {
        return Title{}, errors.New("title cannot be empty")
    }

    if len(value) > 100 {
        return Title{}, errors.New("title too long (max 100 characters)")
    }

    return Title{value: value}, nil
}

func (t Title) String() string {
    return t.value
}

// TodoStatus 枚举
type TodoStatus int

const (
    StatusPending TodoStatus = iota
    StatusInProgress
    StatusCompleted
)

// Priority 枚举
type Priority int

const (
    PriorityLow Priority = iota
    PriorityNormal
    PriorityHigh
)
```

#### 领域错误

```go
// domain/errors.go
package domain

import "errors"

var (
    ErrTodoNotFound                      = errors.New("todo not found")
    ErrAlreadyCompleted                  = errors.New("todo already completed")
    ErrHighPriorityNeedsApproval        = errors.New("high priority todo needs approval")
    ErrCannotChangePriorityWhenCompleted = errors.New("cannot change priority when completed")
    ErrAlreadyHighestPriority           = errors.New("already highest priority")
)
```

### 2. Application Layer（应用层）- 编排

#### Driving Port（入站端口）

```go
// application/ports/input/todo_service.go
package input

import "your-module/domain"

// TodoService 应用服务接口（由应用层定义，供外部调用）
type TodoService interface {
    CreateTodo(command CreateTodoCommand) (*domain.Todo, error)
    GetTodo(query GetTodoQuery) (*domain.Todo, error)
    CompleteTodo(command CompleteTodoCommand) error
    IncreasePriority(command IncreasePriorityCommand) error
    ListTodos(query ListTodosQuery) ([]*domain.Todo, error)
}

// 命令对象（Command）
type CreateTodoCommand struct {
    Title       string
    Description string
}

type CompleteTodoCommand struct {
    TodoID int
}

type IncreasePriorityCommand struct {
    TodoID int
}

// 查询对象（Query）
type GetTodoQuery struct {
    TodoID int
}

type ListTodosQuery struct {
    Status *domain.TodoStatus
}
```

#### Driven Port（出站端口）

```go
// application/ports/output/todo_repository.go
package output

import "your-module/domain"

// TodoRepository 由应用层定义，供持久化适配器实现
type TodoRepository interface {
    Save(todo *domain.Todo) error
    FindByID(id domain.TodoID) (*domain.Todo, error)
    FindAll() ([]*domain.Todo, error)
    Update(todo *domain.Todo) error
    Delete(id domain.TodoID) error
}
```

**关键点**：
- 接口由内部（应用层）定义，不受外部技术影响
- 使用领域对象（domain.Todo, domain.TodoID）
- 外部适配器实现这些接口

#### 应用服务实现

```go
// application/services/todo_service_impl.go
package services

import (
    "your-module/application/ports/input"
    "your-module/application/ports/output"
    "your-module/domain"
)

type TodoServiceImpl struct {
    todoRepo output.TodoRepository
    // 可以有其他依赖，如通知服务
}

func NewTodoService(repo output.TodoRepository) input.TodoService {
    return &TodoServiceImpl{
        todoRepo: repo,
    }
}

func (s *TodoServiceImpl) CreateTodo(cmd input.CreateTodoCommand) (*domain.Todo, error) {
    // 1. 创建值对象
    title, err := domain.NewTitle(cmd.Title)
    if err != nil {
        return nil, err
    }

    // 2. 创建领域对象（工厂方法）
    todo, err := domain.NewTodo(title, cmd.Description)
    if err != nil {
        return nil, err
    }

    // 3. 持久化
    if err := s.todoRepo.Save(todo); err != nil {
        return nil, err
    }

    return todo, nil
}

func (s *TodoServiceImpl) CompleteTodo(cmd input.CompleteTodoCommand) error {
    // 1. 查询领域对象
    todoID := domain.TodoID{Value: cmd.TodoID}
    todo, err := s.todoRepo.FindByID(todoID)
    if err != nil {
        return domain.ErrTodoNotFound
    }

    // 2. 调用领域方法（业务逻辑在这里）
    if err := todo.Complete(); err != nil {
        return err
    }

    // 3. 持久化更改
    return s.todoRepo.Update(todo)
}
```

**关键点**：
- 应用服务是"薄"的，只负责编排
- 真正的业务逻辑在领域对象中
- 依赖端口接口，不是具体实现

### 3. Adapters Layer（适配器层）- 技术实现

#### HTTP Adapter（Driving Adapter）

```go
// adapters/input/http/todo_handler.go
package http

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "your-module/application/ports/input"
)

type TodoHandler struct {
    service input.TodoService  // 依赖端口接口
}

func NewTodoHandler(service input.TodoService) *TodoHandler {
    return &TodoHandler{service: service}
}

func (h *TodoHandler) Create(c *gin.Context) {
    var req struct {
        Title       string `json:"title"`
        Description string `json:"description"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 构建命令对象
    cmd := input.CreateTodoCommand{
        Title:       req.Title,
        Description: req.Description,
    }

    // 调用应用服务
    todo, err := h.service.CreateTodo(cmd)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 响应转换（Domain → DTO）
    c.JSON(http.StatusCreated, toDTO(todo))
}

func (h *TodoHandler) Complete(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))

    cmd := input.CompleteTodoCommand{TodoID: id}

    if err := h.service.CompleteTodo(cmd); err != nil {
        // 错误映射
        if err == domain.ErrHighPriorityNeedsApproval {
            c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "completed"})
}
```

#### Repository Adapter（Driven Adapter）

```go
// adapters/output/persistence/memory/todo_memory.go
package memory

import (
    "sync"
    "your-module/application/ports/output"
    "your-module/domain"
)

type MemoryTodoRepository struct {
    todos map[int]*domain.Todo
    mu    sync.RWMutex
}

func NewMemoryTodoRepository() output.TodoRepository {
    return &MemoryTodoRepository{
        todos: make(map[int]*domain.Todo),
    }
}

func (r *MemoryTodoRepository) Save(todo *domain.Todo) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    r.todos[todo.ID().Value()] = todo
    return nil
}

func (r *MemoryTodoRepository) FindByID(id domain.TodoID) (*domain.Todo, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    if todo, ok := r.todos[id.Value()]; ok {
        return todo, nil
    }

    return nil, domain.ErrTodoNotFound
}
```

## 依赖方向

```
┌─────────────────────────────────────┐
│        Adapters (外层)              │
│  ↓ 依赖                             │
│        Ports (接口层)               │
│  ↓ 依赖                             │
│        Domain (核心，不依赖任何外部) │
└─────────────────────────────────────┘
```

**依赖倒置原则（DIP）**：
- 外层依赖内层
- 内层定义接口，外层实现接口
- 核心领域不依赖任何技术框架

## 测试策略

### 1. 领域层测试（最重要）

```go
// 测试业务逻辑，无需任何Mock
func TestTodo_Complete(t *testing.T) {
    // Given
    title, _ := domain.NewTitle("Test Todo")
    todo, _ := domain.NewTodo(title, "Description")

    // When
    err := todo.Complete()

    // Then
    assert.NoError(t, err)
    assert.Equal(t, domain.StatusCompleted, todo.Status())
}

func TestTodo_Complete_HighPriority_ShouldFail(t *testing.T) {
    // Given
    title, _ := domain.NewTitle("High Priority Todo")
    todo, _ := domain.NewTodo(title, "")
    todo.IncreasePriority()  // 提升到 High

    // When
    err := todo.Complete()

    // Then
    assert.ErrorIs(t, err, domain.ErrHighPriorityNeedsApproval)
}
```

### 2. 应用层测试（使用Mock Repository）

```go
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Save(todo *domain.Todo) error {
    args := m.Called(todo)
    return args.Error(0)
}

func TestTodoService_CreateTodo(t *testing.T) {
    // Given
    mockRepo := new(MockRepository)
    mockRepo.On("Save", mock.Anything).Return(nil)

    service := services.NewTodoService(mockRepo)

    // When
    cmd := input.CreateTodoCommand{
        Title: "Test Todo",
    }
    todo, err := service.CreateTodo(cmd)

    // Then
    assert.NoError(t, err)
    assert.NotNil(t, todo)
    mockRepo.AssertExpectations(t)
}
```

### 3. 适配器测试

```go
// HTTP适配器测试
func TestHandler_Create(t *testing.T) {
    mockService := new(MockTodoService)
    handler := http.NewTodoHandler(mockService)

    // ... 测试HTTP逻辑
}

// Repository适配器测试
func TestMemoryRepository_Save(t *testing.T) {
    repo := memory.NewMemoryTodoRepository()
    title, _ := domain.NewTitle("Test")
    todo, _ := domain.NewTodo(title, "")

    err := repo.Save(todo)
    assert.NoError(t, err)
}
```

## 优点分析

| 优点 | 说明 | 示例 |
|------|------|------|
| **业务逻辑独立** | 核心不依赖任何框架 | 领域对象可单独测试 |
| **充血模型** | 业务规则封装在实体中 | `todo.Complete()` 包含业务逻辑 |
| **高度可测试** | 领域层无需Mock | 直接测试业务规则 |
| **技术可替换** | 适配器随意替换 | HTTP → gRPC, Memory → DB |
| **符合SOLID** | 依赖倒置、单一职责 | 每层职责明确 |

## 缺点分析

| 缺点 | 说明 | 影响 |
|------|------|------|
| **学习曲线陡** | 概念复杂 | 团队需要培训 |
| **代码量增加** | 接口、适配器代码多 | 简单功能也复杂 |
| **过度设计风险** | 小项目不适合 | 增加维护成本 |
| **DTO转换繁琐** | Domain ↔ DTO转换 | 模板代码多 |

## 与 V2 对比

| 维度 | V2 分层架构 | V3 六边形架构 |
|------|-------------|---------------|
| **模型** | 贫血模型 | 充血模型 |
| **依赖方向** | 上层依赖下层 | 外层依赖内层（DIP） |
| **业务逻辑** | 分散在Service | 集中在领域实体 |
| **接口定义** | Repository接口在数据层 | 端口接口在应用层 |
| **可测试性** | 需要Mock | 领域层无需Mock |
| **复杂度** | 中 | 高 |

## 何时使用六边形架构

✅ **适合场景**：
- 复杂业务逻辑
- 长期维护的核心系统
- 需要多种适配器（HTTP, gRPC, CLI）
- 团队熟悉DDD

❌ **不适合场景**：
- 简单CRUD应用
- 快速原型
- 团队不熟悉DDD
- 短期项目

## 演进到 V4 的动机

虽然六边形架构很优雅，但你可能会遇到新问题：

### 问题 1: 读写性能需求不同

```go
// 写操作：需要完整的领域对象
func (s *Service) CompleteTodo(cmd Command) error {
    todo, _ := s.repo.FindByID(id)  // 加载完整对象
    todo.Complete()                  // 修改
    s.repo.Update(todo)              // 保存
}

// 读操作：只需要展示数据
func (s *Service) GetTodoList() []TodoDTO {
    todos, _ := s.repo.FindAll()     // 也加载了完整对象
    return toDTO(todos)               // 但只是展示
}
```

**问题**：读操作不需要领域逻辑，却有领域对象的开销

### 问题 2: 复杂查询

```go
// 需要跨多个聚合根查询
func (s *Service) GetUserTodoStatistics(userID int) Statistics {
    // 需要复杂的数据库查询
    // 但领域模型不适合这种查询
}
```

**这时候，我们需要 CQRS（命令查询职责分离）！**

## 练习任务

### 必做任务
1. ✅ 实现充血的 Todo 领域模型
2. ✅ 实现值对象（Title, Priority）
3. ✅ 实现应用服务和端口
4. ✅ 实现 HTTP 和 Memory 适配器
5. ✅ 为领域层编写单元测试

### 进阶任务
1. 🔧 添加 TodoList 聚合根（管理多个Todo）
2. 🔧 实现 SQLite 适配器
3. 🔧 添加 CLI 适配器
4. 🔧 实现领域事件（TodoCompleted事件）

### 思考题
1. 💭 贫血模型和充血模型的本质区别是什么？
2. 💭 为什么端口接口要由内部定义？
3. 💭 如何处理跨聚合根的事务？

---

**完成 V3 后，继续学习 [V4: CQRS架构](./v4-cqrs.md)**

# V2: 分层架构 (Layered Architecture)

## 架构概述

分层架构是最经典的架构模式之一，将应用程序分为不同的层次，每层只能依赖其下层。这种架构强调**关注点分离**和**依赖方向**的控制。

## 从 V1 到 V2 的演进

### V1 的问题回顾

```go
// V1: 所有逻辑混在一起
func createTodo(c *gin.Context) {
    var todo Todo
    c.ShouldBindJSON(&todo)           // HTTP层
    todo.ID = nextID                  // 业务逻辑
    nextID++
    mu.Lock()                         // 数据访问
    todos = append(todos, todo)
    mu.Unlock()
    c.JSON(200, todo)                 // HTTP层
}
```

问题：
- ❌ HTTP框架与业务逻辑耦合，无法测试
- ❌ 数据访问逻辑分散在各处
- ❌ 无法替换存储方式（内存→数据库）

### V2 的解决方案

将应用分为三层：
```
┌─────────────────────┐
│   Presentation      │  ← HTTP Handlers (Gin)
├─────────────────────┤
│   Business Logic    │  ← Service Layer
├─────────────────────┤
│   Data Access       │  ← Repository Layer
└─────────────────────┘
```

## 设计目标

- ✅ 关注点分离（Separation of Concerns）
- ✅ 可测试性（每层可独立测试）
- ✅ 可替换性（数据库可替换）
- ✅ 代码复用
- ✅ 清晰的依赖方向

## 目录结构

```
v2-layered/
├── go.mod
├── go.sum
├── main.go                  # 程序入口，组装各层
├── README.md
│
├── model/                   # 数据模型（贯穿各层）
│   └── todo.go
│
├── handler/                 # 表示层（Presentation Layer）
│   └── todo_handler.go      # HTTP 请求处理
│
├── service/                 # 业务逻辑层（Business Logic Layer）
│   └── todo_service.go      # 业务规则和流程
│
├── repository/              # 数据访问层（Data Access Layer）
│   ├── todo_repository.go   # Repository接口定义
│   └── memory/
│       └── todo_memory.go   # 内存实现
│
└── tests/
    ├── handler_test.go
    ├── service_test.go
    └── repository_test.go
```

## 核心概念

### 1. Model Layer (数据模型)

定义领域对象，贯穿各层：

```go
// model/todo.go
package model

import "time"

type Todo struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// 请求和响应 DTO
type CreateTodoRequest struct {
    Title       string `json:"title" binding:"required"`
    Description string `json:"description"`
}

type UpdateTodoRequest struct {
    Title       *string `json:"title"`
    Description *string `json:"description"`
    Completed   *bool   `json:"completed"`
}
```

### 2. Repository Layer (数据访问层)

负责数据的 CRUD 操作，定义接口而非实现：

```go
// repository/todo_repository.go
package repository

import "your-module/model"

// TodoRepository 定义数据访问接口
type TodoRepository interface {
    Create(todo *model.Todo) error
    FindByID(id int) (*model.Todo, error)
    FindAll() ([]*model.Todo, error)
    Update(todo *model.Todo) error
    Delete(id int) error
}
```

**关键点**：
- 定义接口，不依赖具体实现
- 使用指针避免大量拷贝
- 返回错误而非 panic

### 3. Service Layer (业务逻辑层)

包含业务规则和流程编排：

```go
// service/todo_service.go
package service

import (
    "errors"
    "time"
    "your-module/model"
    "your-module/repository"
)

type TodoService struct {
    repo repository.TodoRepository
}

func NewTodoService(repo repository.TodoRepository) *TodoService {
    return &TodoService{repo: repo}
}

// 业务方法示例
func (s *TodoService) CreateTodo(req model.CreateTodoRequest) (*model.Todo, error) {
    // 业务验证
    if len(req.Title) > 100 {
        return nil, errors.New("title too long")
    }

    // 构建领域对象
    todo := &model.Todo{
        Title:       req.Title,
        Description: req.Description,
        Completed:   false,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    // 委托给 Repository
    if err := s.repo.Create(todo); err != nil {
        return nil, err
    }

    return todo, nil
}
```

**关键点**：
- 包含业务验证逻辑
- 不关心数据如何存储
- 依赖 Repository 接口，不是具体实现

### 4. Handler Layer (表示层)

处理 HTTP 请求和响应：

```go
// handler/todo_handler.go
package handler

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "your-module/model"
    "your-module/service"
)

type TodoHandler struct {
    service *service.TodoService
}

func NewTodoHandler(service *service.TodoService) *TodoHandler {
    return &TodoHandler{service: service}
}

func (h *TodoHandler) Create(c *gin.Context) {
    var req model.CreateTodoRequest

    // 解析请求
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 调用 Service
    todo, err := h.service.CreateTodo(req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 返回响应
    c.JSON(http.StatusCreated, todo)
}
```

**关键点**：
- 只负责 HTTP 相关逻辑
- 请求解析、响应格式化
- 错误码映射（业务错误 → HTTP状态码）

## 实现步骤

### Step 1: 定义数据模型

```bash
mkdir -p model
# 创建 model/todo.go
# 定义 Todo、CreateTodoRequest、UpdateTodoRequest
```

### Step 2: 实现 Repository 层

```bash
mkdir -p repository/memory
# 创建 repository/todo_repository.go (接口)
# 创建 repository/memory/todo_memory.go (内存实现)
```

**内存实现要点**：
```go
type MemoryTodoRepository struct {
    todos  map[int]*model.Todo
    nextID int
    mu     sync.RWMutex
}

func (r *MemoryTodoRepository) Create(todo *model.Todo) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    todo.ID = r.nextID
    r.nextID++
    r.todos[todo.ID] = todo

    return nil
}
```

### Step 3: 实现 Service 层

```bash
mkdir -p service
# 创建 service/todo_service.go
# 实现所有业务方法
```

**测试 Service 层**：
```go
// tests/service_test.go
func TestTodoService_CreateTodo(t *testing.T) {
    repo := memory.NewMemoryTodoRepository()
    service := service.NewTodoService(repo)

    req := model.CreateTodoRequest{
        Title: "Test Todo",
    }

    todo, err := service.CreateTodo(req)
    assert.NoError(t, err)
    assert.Equal(t, "Test Todo", todo.Title)
}
```

### Step 4: 实现 Handler 层

```bash
mkdir -p handler
# 创建 handler/todo_handler.go
# 实现所有 HTTP handlers
```

### Step 5: 组装应用（main.go）

```go
// main.go
package main

import (
    "github.com/gin-gonic/gin"
    "your-module/handler"
    "your-module/repository/memory"
    "your-module/service"
)

func main() {
    // 依赖注入：从下往上组装
    repo := memory.NewMemoryTodoRepository()
    svc := service.NewTodoService(repo)
    h := handler.NewTodoHandler(svc)

    // 路由配置
    r := gin.Default()

    todos := r.Group("/todos")
    {
        todos.POST("", h.Create)
        todos.GET("", h.GetAll)
        todos.GET("/:id", h.Get)
        todos.PUT("/:id", h.Update)
        todos.DELETE("/:id", h.Delete)
    }

    r.Run(":8080")
}
```

## 依赖方向

```
Handler → Service → Repository → Model
  ↓         ↓          ↓
 只依赖下层，不依赖上层
```

**关键原则**：
- Handler 依赖 Service
- Service 依赖 Repository 接口
- Repository 不依赖任何业务层

## 测试策略

### 1. Repository 层测试

```go
// 测试数据访问逻辑
func TestMemoryRepository_Create(t *testing.T) {
    repo := memory.NewMemoryTodoRepository()
    todo := &model.Todo{Title: "Test"}

    err := repo.Create(todo)
    assert.NoError(t, err)
    assert.NotZero(t, todo.ID)
}
```

### 2. Service 层测试

```go
// 使用 Mock Repository 测试业务逻辑
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Create(todo *model.Todo) error {
    args := m.Called(todo)
    return args.Error(0)
}

func TestService_CreateTodo_TitleTooLong(t *testing.T) {
    mockRepo := new(MockRepository)
    service := service.NewTodoService(mockRepo)

    req := model.CreateTodoRequest{
        Title: strings.Repeat("a", 101),
    }

    _, err := service.CreateTodo(req)
    assert.Error(t, err)
    mockRepo.AssertNotCalled(t, "Create")
}
```

### 3. Handler 层测试

```go
// 使用 httptest 测试 HTTP 逻辑
func TestHandler_Create(t *testing.T) {
    repo := memory.NewMemoryTodoRepository()
    svc := service.NewTodoService(repo)
    handler := handler.NewTodoHandler(svc)

    gin.SetMode(gin.TestMode)
    router := gin.New()
    router.POST("/todos", handler.Create)

    body := `{"title":"Test Todo"}`
    req := httptest.NewRequest("POST", "/todos", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
}
```

## 优点分析

| 优点 | 说明 | 示例 |
|------|------|------|
| **关注点分离** | 每层职责明确 | Handler只管HTTP，Service只管业务 |
| **可测试性** | 各层独立测试 | Service可用Mock Repository测试 |
| **可替换性** | 实现可替换 | 内存存储→数据库，无需改业务代码 |
| **代码复用** | Service可被多个Handler使用 | HTTP、gRPC共用Service |
| **易于理解** | 结构清晰 | 新人容易上手 |

## 缺点分析

| 缺点 | 说明 | 影响 |
|------|------|------|
| **层间耦合** | 上层依赖下层的具体结构 | Model变化影响所有层 |
| **性能开销** | 层层调用有开销 | 简单查询也要经过三层 |
| **过度设计** | 简单功能也要三层 | 增加代码量 |
| **领域模型贫血** | Model只有数据，没有行为 | 业务逻辑分散在Service |
| **数据传递繁琐** | 需要DTO转换 | Request → Model → Response |

## 常见问题

### Q1: Service 和 Repository 的边界在哪里？

**Repository**：
- 数据的 CRUD
- 简单查询（按ID、按条件）
- 不包含业务逻辑

**Service**：
- 业务验证（标题长度、状态转换规则）
- 业务流程（创建待办+发送通知）
- 跨Repository操作

### Q2: 何时需要 DTO 转换？

```go
// 需要转换的场景
type CreateTodoRequest struct {
    Title string `json:"title"`
}

type TodoResponse struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
    // 不暴露内部字段
}

// 不需要转换的场景
// 如果 API 响应和 Model 完全一致，可以直接返回
```

### Q3: 错误处理在哪一层？

```go
// Repository: 返回基础错误
func (r *Repo) FindByID(id int) (*Todo, error) {
    if todo, ok := r.todos[id]; ok {
        return todo, nil
    }
    return nil, errors.New("not found")
}

// Service: 业务错误
func (s *Service) GetTodo(id int) (*Todo, error) {
    todo, err := s.repo.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("todo %d not found", id)
    }
    return todo, nil
}

// Handler: HTTP状态码映射
func (h *Handler) Get(c *gin.Context) {
    todo, err := h.service.GetTodo(id)
    if err != nil {
        c.JSON(404, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, todo)
}
```

## 进阶功能

### 1. 添加持久化存储

创建 SQLite 实现：
```
repository/
├── todo_repository.go
├── memory/
│   └── todo_memory.go
└── sqlite/
    └── todo_sqlite.go    # 新增
```

```go
// repository/sqlite/todo_sqlite.go
type SQLiteTodoRepository struct {
    db *sql.DB
}

func (r *SQLiteTodoRepository) Create(todo *model.Todo) error {
    result, err := r.db.Exec(
        "INSERT INTO todos (title, description, completed) VALUES (?, ?, ?)",
        todo.Title, todo.Description, todo.Completed,
    )
    // ...
    todo.ID = int(lastID)
    return nil
}
```

**main.go 只需改一行**：
```go
// repo := memory.NewMemoryTodoRepository()  // 旧
repo := sqlite.NewSQLiteTodoRepository(db)   // 新
```

### 2. 添加用户系统

增加 User 相关的三层：
```
├── model/
│   ├── todo.go
│   └── user.go          # 新增
├── repository/
│   ├── todo_repository.go
│   └── user_repository.go  # 新增
├── service/
│   ├── todo_service.go
│   └── user_service.go     # 新增
└── handler/
    ├── todo_handler.go
    └── user_handler.go     # 新增
```

## 何时使用分层架构

✅ **适合场景**：
- 中小型单体应用
- 团队成员技术水平不一
- 需求相对稳定
- CRUD 密集型应用

❌ **不适合场景**：
- 复杂业务逻辑（推荐DDD）
- 高性能要求（层级开销）
- 频繁的横向扩展需求

## 演进到 V3 的动机

当你实现完 V2 后，思考以下问题：

### 问题 1: 业务逻辑分散

```go
// 业务规则散落在 Service 的各个方法中
func (s *Service) CompleteTodo(id int) error {
    todo, _ := s.repo.FindByID(id)
    if todo.Completed {
        return errors.New("already completed")  // 业务规则
    }
    todo.Completed = true
    return s.repo.Update(todo)
}
```

**问题**：业务规则没有集中在领域对象上

### 问题 2: 依赖技术细节

```go
// Service 依赖了 Repository 接口
// 但 Repository 接口设计受数据库影响
type TodoRepository interface {
    FindByID(id int) (*Todo, error)  // 假设有主键
}
```

**问题**：业务层依赖了技术实现的假设

### 问题 3: 难以应对复杂业务

假设需求变更：
- 待办事项有优先级
- 高优先级的待办不能直接完成，需要审批
- 完成后发送通知

**问题**：这些业务规则应该放在哪里？

**这时候，我们需要六边形架构（领域驱动设计）！**

## 练习任务

### 必做任务
1. ✅ 按照分层架构实现所有功能
2. ✅ 为每一层编写单元测试
3. ✅ 实现 SQLite Repository（替换内存实现）
4. ✅ 添加输入验证和错误处理

### 进阶任务
1. 🔧 添加用户系统（User Model + 三层）
2. 🔧 实现待办事项归属于用户
3. 🔧 添加分页功能（Repository层）
4. 🔧 添加日志记录（使用中间件）

### 思考题
1. 💭 如果要添加 gRPC API，需要改哪些代码？
2. 💭 如果 Todo 有复杂的状态转换规则，应该放在哪一层？
3. 💭 如何避免"贫血模型"（Model只有数据没有行为）？

## 对比 V1

| 维度 | V1 单体 | V2 分层 |
|------|---------|---------|
| 代码行数 | ~100行 | ~300行 |
| 可测试性 | ❌ 无法测试 | ✅ 各层独立测试 |
| 可维护性 | ❌ 修改影响大 | ✅ 修改局限在一层 |
| 学习成本 | 低 | 中 |
| 适用规模 | 原型 | 中小型应用 |

---

**完成 V2 后，继续学习 [V3: 六边形架构](./v3-hexagonal.md)**

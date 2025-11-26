# V2: 分层架构实现

## 开始之前

请先阅读架构设计文档：[V2 分层架构](../docs/v2-layered.md)

## 实现目标

将 V1 的单体代码重构为分层架构，按照职责分为三层。

## 目录结构

创建以下目录和文件：

```
v2-layered/
├── go.mod
├── main.go
├── model/
│   └── todo.go
├── repository/
│   ├── todo_repository.go
│   └── memory/
│       └── todo_memory.go
├── service/
│   └── todo_service.go
├── handler/
│   └── todo_handler.go
└── tests/
    ├── repository_test.go
    ├── service_test.go
    └── handler_test.go
```

## 实现步骤

### Step 1: Model 层

```bash
mkdir -p model
# 创建 model/todo.go
# 定义 Todo, CreateTodoRequest, UpdateTodoRequest
```

### Step 2: Repository 层

```bash
mkdir -p repository/memory
# 创建 repository/todo_repository.go（接口）
# 创建 repository/memory/todo_memory.go（实现）
```

### Step 3: Service 层

```bash
mkdir -p service
# 创建 service/todo_service.go
# 实现业务逻辑
```

### Step 4: Handler 层

```bash
mkdir -p handler
# 创建 handler/todo_handler.go
# 实现 HTTP 处理
```

### Step 5: 组装（main.go）

依赖注入，从下往上组装：

```go
repo := memory.NewMemoryTodoRepository()
service := service.NewTodoService(repo)
handler := handler.NewTodoHandler(service)
```

## 测试

### 单元测试

```bash
# 测试 Repository
go test ./repository/memory -v

# 测试 Service（使用 Mock Repository）
go test ./service -v

# 测试 Handler
go test ./handler -v
```

### 集成测试

```bash
go run main.go
# 使用 curl 测试所有端点
```

## 进阶任务

### 1. 添加 SQLite 实现

```bash
mkdir -p repository/sqlite
# 实现 repository/sqlite/todo_sqlite.go
# 只需修改 main.go 即可切换存储
```

### 2. 添加输入验证

在 Service 层添加：
- 标题不能为空
- 标题长度限制
- Description 长度限制

### 3. 添加日志

使用 `log/slog` 记录：
- 每个请求的日志
- 错误日志
- 性能日志

## 对比 V1

| 维度 | V1 | V2 |
|------|----|----|
| 文件数 | 1 | 8+ |
| 可测试性 | 差 | 好 |
| 可维护性 | 差 | 好 |

## 完成后

思考以下问题：

1. Service 和 Repository 的边界在哪里？
2. 如何避免"贫血模型"？
3. 为什么 Repository 要定义接口？

完成后继续学习 [V3: 六边形架构](../docs/v3-hexagonal.md)

# V3: 六边形架构实现

## 开始之前

请先阅读架构设计文档：[V3 六边形架构](../docs/v3-hexagonal.md)

## 实现目标

实现六边形架构（端口与适配器），核心业务逻辑与技术实现完全隔离。

## 核心概念

- **Domain（领域）**：业务逻辑核心，充血模型
- **Application（应用）**：编排领域对象
- **Ports（端口）**：接口定义
- **Adapters（适配器）**：技术实现

## 目录结构

```
v3-hexagonal/
├── go.mod
├── main.go
├── domain/                 # 核心领域
│   ├── todo.go            # 充血模型
│   ├── value_objects.go   # 值对象
│   └── errors.go
├── application/            # 应用层
│   ├── ports/
│   │   ├── input/         # Driving Ports
│   │   │   └── todo_service.go
│   │   └── output/        # Driven Ports
│   │       └── todo_repository.go
│   └── services/
│       └── todo_service_impl.go
└── adapters/               # 适配器
    ├── input/
    │   └── http/
    │       └── todo_handler.go
    └── output/
        └── persistence/
            └── memory/
                └── todo_memory.go
```

## 实现步骤

### Step 1: 领域层（最重要）

```bash
mkdir -p domain
```

实现：
1. **充血模型**：`Todo` 实体包含业务方法
   - `Complete()` - 完成待办
   - `IncreasePriority()` - 提升优先级
   - `ChangeTo()` - 修改标题

2. **值对象**：`Title`, `TodoID`, `Priority`
   - 带验证逻辑
   - 不可变

3. **领域错误**：
   - `ErrAlreadyCompleted`
   - `ErrHighPriorityNeedsApproval`

### Step 2: 应用层（端口定义）

```bash
mkdir -p application/ports/{input,output}
mkdir -p application/services
```

实现：
1. **Input Port**：`TodoService` 接口
   - 定义命令对象（Command）
   - 定义查询对象（Query）

2. **Output Port**：`TodoRepository` 接口
   - 使用领域对象
   - 不依赖技术细节

3. **Application Service**：实现 Input Port
   - 编排领域对象
   - 调用 Output Port

### Step 3: 适配器层

```bash
mkdir -p adapters/input/http
mkdir -p adapters/output/persistence/memory
```

实现：
1. **HTTP Adapter**：实现 HTTP 处理
2. **Memory Adapter**：实现 Repository 接口

## 关键实现

### 充血模型示例

```go
type Todo struct {
    id     TodoID
    title  Title
    status TodoStatus
}

// 业务方法
func (t *Todo) Complete() error {
    if t.status == StatusCompleted {
        return ErrAlreadyCompleted
    }
    t.status = StatusCompleted
    return nil
}
```

### 值对象示例

```go
type Title struct {
    value string
}

func NewTitle(value string) (Title, error) {
    if value == "" {
        return Title{}, errors.New("title cannot be empty")
    }
    if len(value) > 100 {
        return Title{}, errors.New("title too long")
    }
    return Title{value: value}, nil
}
```

## 测试

### 领域层测试（最重要）

```go
func TestTodo_Complete(t *testing.T) {
    title, _ := domain.NewTitle("Test")
    todo, _ := domain.NewTodo(title, "")

    err := todo.Complete()

    assert.NoError(t, err)
    assert.Equal(t, domain.StatusCompleted, todo.Status())
}
```

**关键**：无需任何 Mock，直接测试业务逻辑！

## 对比 V2

| 维度 | V2 分层 | V3 六边形 |
|------|---------|-----------|
| 模型 | 贫血 | 充血 |
| 依赖方向 | 上→下 | 外→内 |
| 业务逻辑 | Service中 | Domain中 |
| 可测试性 | 需要Mock | 领域层无需Mock |

## 完成后

思考以下问题：

1. 贫血模型和充血模型的区别？
2. 为什么端口要由内部定义？
3. 依赖倒置原则如何体现？

完成后继续学习 [V4: CQRS架构](../docs/v4-cqrs.md)

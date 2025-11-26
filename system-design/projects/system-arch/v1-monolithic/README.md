# V1: 单体架构实现

## 开始之前

请先阅读架构设计文档：[V1 单体架构](../docs/v1-monolithic.md)

## 实现目标

实现一个简单的待办事项管理系统，所有代码都在 `main.go` 中。

### 功能需求

- [x] 创建待办事项（POST /todos）
- [x] 获取所有待办（GET /todos）
- [x] 获取单个待办（GET /todos/:id）
- [x] 更新待办事项（PUT /todos/:id）
- [x] 删除待办事项（DELETE /todos/:id）

## 开始实现

### 1. 初始化项目

```bash
cd v1-monolithic
go mod init github.com/yourusername/todo-v1
go get -u github.com/gin-gonic/gin
```

### 2. 创建 main.go

参考文档中的代码结构：
- 定义 `Todo` 结构体
- 创建全局变量存储数据
- 实现 CRUD handler 函数
- 设置路由和启动服务器

### 3. 运行和测试

```bash
# 运行程序
go run main.go

# 测试API（另开一个终端）
# 创建待办
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"学习架构设计","description":"完成V1实现"}'

# 获取列表
curl http://localhost:8080/todos

# 获取单个
curl http://localhost:8080/todos/1

# 更新
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"completed":true}'

# 删除
curl -X DELETE http://localhost:8080/todos/1
```

## 实现提示

### 数据结构

```go
type Todo struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### 并发安全

记得使用 `sync.RWMutex` 保护全局变量：

```go
var (
    todos  []Todo
    nextID = 1
    mu     sync.RWMutex
)
```

### 边界情况处理

- ID 不存在时返回 404
- 标题为空时返回 400
- JSON 解析失败时返回 400

## 完成后

思考以下问题：

1. 如果要添加用户系统，需要改哪些代码？
2. 如何为这些函数编写单元测试？
3. 重启程序后数据丢失，如何解决？

完成后继续学习 [V2: 分层架构](../docs/v2-layered.md)

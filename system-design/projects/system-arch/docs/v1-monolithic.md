# V1: 单体架构 (Monolithic Architecture)

## 架构概述

单体架构是最简单直接的架构模式，所有代码都在一个文件或少数几个文件中，没有明确的分层和模块划分。这是快速原型开发和概念验证的理想选择。

## 设计目标

- ✅ 快速实现功能
- ✅ 代码简单直接
- ✅ 容易理解和调试
- ❌ 不考虑扩展性
- ❌ 不考虑测试性

## 目录结构

```
v1-monolithic/
├── go.mod
├── go.sum
├── main.go          # 所有代码都在这里
└── README.md
```

## 核心概念

### 数据结构

```go
// Todo 表示一个待办事项
type Todo struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### 功能列表

1. **创建待办事项** - `POST /todos`
2. **获取所有待办** - `GET /todos`
3. **获取单个待办** - `GET /todos/:id`
4. **更新待办事项** - `PUT /todos/:id`
5. **删除待办事项** - `DELETE /todos/:id`

## 实现要点

### 1. 数据存储

使用内存存储（简单的 slice）：
```go
var (
    todos  []Todo
    nextID = 1
    mu     sync.RWMutex // 保护并发访问
)
```

### 2. HTTP 服务器

使用标准库 `net/http` 或 `gin` 框架：
- 定义路由
- 直接在 handler 中处理业务逻辑
- 直接操作全局变量

### 3. 请求处理流程

```
HTTP 请求 → Handler 函数 → 直接操作数据 → 返回 JSON
```

## 代码实现指南

### main.go 结构

```go
package main

import (
    // 导入必要的包
)

// 1. 定义数据结构
type Todo struct { ... }

// 2. 全局变量（数据存储）
var todos []Todo
var nextID int
var mu sync.RWMutex

// 3. Handler 函数
func createTodo(c *gin.Context) {
    // 解析请求
    // 创建 Todo
    // 添加到 slice
    // 返回响应
}

func getTodos(c *gin.Context) { ... }
func getTodo(c *gin.Context) { ... }
func updateTodo(c *gin.Context) { ... }
func deleteTodo(c *gin.Context) { ... }

// 4. 主函数
func main() {
    // 创建路由
    // 注册 handlers
    // 启动服务器
}
```

### 示例 API 调用

```bash
# 创建待办事项
curl -X POST http://localhost:8080/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"学习Go","description":"完成待办系统"}'

# 获取所有待办
curl http://localhost:8080/todos

# 获取单个待办
curl http://localhost:8080/todos/1

# 更新待办
curl -X PUT http://localhost:8080/todos/1 \
  -H "Content-Type: application/json" \
  -d '{"completed":true}'

# 删除待办
curl -X DELETE http://localhost:8080/todos/1
```

## 实现步骤

### Step 1: 初始化项目
```bash
mkdir v1-monolithic
cd v1-monolithic
go mod init todo-v1
go get -u github.com/gin-gonic/gin
```

### Step 2: 实现基本结构
1. 定义 `Todo` 结构体
2. 创建全局变量存储数据
3. 实现 `main` 函数和路由

### Step 3: 实现 CRUD 操作
1. `createTodo` - 创建
2. `getTodos` - 查询列表
3. `getTodo` - 查询单个
4. `updateTodo` - 更新
5. `deleteTodo` - 删除

### Step 4: 测试
手动测试所有 API 端点

## 优点分析

| 优点 | 说明 |
|------|------|
| **简单直接** | 所有代码在一个文件，容易理解 |
| **开发速度快** | 不需要考虑架构设计，直接实现功能 |
| **易于调试** | 调用栈简单，容易跟踪 |
| **部署简单** | 单一可执行文件，直接运行 |
| **适合原型** | 快速验证想法 |

## 缺点分析

| 缺点 | 说明 | 影响 |
|------|------|------|
| **代码混乱** | 所有逻辑混在一起 | 难以维护 |
| **难以测试** | 逻辑与框架耦合 | 无法单元测试 |
| **无法重用** | 代码紧密耦合 | 重复代码 |
| **并发问题** | 全局状态需要锁保护 | 性能瓶颈 |
| **数据丢失** | 内存存储，重启丢失 | 不适合生产 |
| **难以扩展** | 添加功能导致代码更混乱 | 技术债务 |

## 遇到的问题

随着功能增加，你会发现：

### 问题 1: 代码重复
```go
// 每个 handler 都要重复错误处理
func createTodo(c *gin.Context) {
    var todo Todo
    if err := c.ShouldBindJSON(&todo); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})  // 重复
        return
    }
    // ...
}
```

### 问题 2: 业务逻辑与框架耦合
```go
// 业务逻辑直接依赖 gin.Context，无法单独测试
func updateTodo(c *gin.Context) {
    id := c.Param("id")  // 依赖 gin
    // 业务逻辑...
}
```

### 问题 3: 全局状态管理混乱
```go
// 多处修改全局变量，难以追踪
var todos []Todo  // 谁都可以修改

func createTodo(...) { todos = append(todos, ...) }
func deleteTodo(...) { todos = remove(todos, ...) }
```

## 何时使用单体架构

✅ **适合场景**：
- 快速原型验证
- 学习新技术
- 简单的工具脚本
- 短期项目（< 1周）

❌ **不适合场景**：
- 需要长期维护的项目
- 需要团队协作
- 功能会持续增加
- 需要单元测试

## 演进到 V2 的动机

当你实现完 V1 后，尝试添加以下功能：

1. **添加用户系统** - 每个用户有自己的待办列表
2. **持久化存储** - 使用数据库而不是内存
3. **添加测试** - 为业务逻辑编写单元测试

你会发现：
- ❌ 所有逻辑混在一起，不知道从哪里开始修改
- ❌ 无法测试业务逻辑（耦合了 HTTP 框架）
- ❌ 添加数据库需要改动大量代码

**这时候，我们需要分层架构！**

## 练习任务

### 必做任务
1. ✅ 实现所有 CRUD 功能
2. ✅ 使用 Postman/curl 测试所有 API
3. ✅ 处理边界情况（ID不存在、空标题等）

### 进阶任务
1. 🔧 添加过滤功能（只显示已完成/未完成）
2. 🔧 添加排序功能（按创建时间、标题）
3. 🔧 添加输入验证（标题不能为空、长度限制）

### 思考题
1. 💭 如果要添加"用户登录"功能，需要改哪些代码？
2. 💭 如果要把数据存到文件/数据库，需要改哪些代码？
3. 💭 如何为 `createTodo` 函数编写单元测试？

## 参考资源

- [Gin 框架文档](https://gin-gonic.com/docs/)
- [Go HTTP 标准库](https://pkg.go.dev/net/http)
- [RESTful API 设计](https://restfulapi.net/)

---

**完成 V1 后，继续学习 [V2: 分层架构](./v2-layered.md)**

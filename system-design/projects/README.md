# 系统设计实践项目

这个目录用于存放学习过程中的所有代码实践项目。

## 目录结构

每周的项目都存放在对应的 `week{N}` 目录下:

```
projects/
├── week1/          # Week 1: 系统设计基础概念
├── week2/          # Week 2: 数据存储基础
├── week3/          # Week 3: 网络与通信基础
├── week4/          # Week 4: 缓存系统设计
├── week5/          # Week 5: 数据库进阶
├── week6/          # Week 6: 消息队列与异步处理
├── week7/          # Week 7: 微服务架构
├── week8/          # Week 8: 分布式系统理论
├── week9/          # Week 9: 监控、日志与追踪
├── week10/         # Week 10: 安全与性能优化
├── week11/         # Week 11: 经典系统设计案例
└── week12/         # Week 12: 综合项目实战
```

## 项目规范

### Go 项目结构
```
project-name/
├── go.mod
├── go.sum
├── main.go
├── README.md
├── pkg/            # 可复用的包
├── internal/       # 内部代码
├── cmd/            # 命令行工具
└── tests/          # 测试文件
```

### Python 项目结构
```
project-name/
├── requirements.txt
├── README.md
├── main.py
├── src/            # 源代码
├── tests/          # 测试文件
└── config/         # 配置文件
```

## 代码规范

### Go 代码规范
- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 编写单元测试 (使用 `testing` 包)
- 添加必要的注释和文档

### Python 代码规范
- 遵循 PEP 8 代码规范
- 使用 `black` 格式化代码
- 编写单元测试 (使用 `pytest`)
- 添加类型注解 (Python 3.6+)

## 如何使用

1. **开始新项目**:
   ```bash
   cd projects/week{N}
   mkdir project-name
   cd project-name
   ```

2. **Go 项目初始化**:
   ```bash
   go mod init github.com/yourusername/project-name
   ```

3. **Python 项目初始化**:
   ```bash
   python -m venv venv
   source venv/bin/activate  # Windows: venv\Scripts\activate
   pip install -r requirements.txt
   ```

4. **运行测试**:
   - Go: `go test ./...`
   - Python: `pytest`

## 学习建议

1. **每个项目都要有 README**: 说明项目目的、如何运行、学到了什么
2. **编写测试**: 养成测试驱动开发的习惯
3. **代码注释**: 特别是复杂逻辑,一定要写清楚
4. **版本控制**: 使用 Git 管理代码
5. **持续改进**: 完成后可以回来优化代码

## 推荐工具

### Go 开发工具
- IDE: VSCode + Go 扩展, GoLand
- 依赖管理: go modules
- 测试: testify, gomock
- Linter: golangci-lint

### Python 开发工具
- IDE: VSCode + Python 扩展, PyCharm
- 依赖管理: pip, poetry
- 测试: pytest, unittest
- Linter: pylint, flake8

## 参考资源

- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Python Best Practices](https://docs.python-guide.org/)
- [Real Python Tutorials](https://realpython.com/)

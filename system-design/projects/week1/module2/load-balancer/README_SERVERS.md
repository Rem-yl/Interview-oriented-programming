# 多服务器测试环境

## 快速开始

```bash
# 1. 启动 5 个后端服务器（端口 8081-8085）
./start_server.sh

# 2. 查看服务器状态
./status_server.sh

# 3. 测试服务器
curl http://localhost:8081/
curl http://localhost:8081/health

# 4. 停止所有服务器
./stop_server.sh
```

## 文件说明

| 文件 | 用途 |
|------|------|
| `cmd/server/main.go` | 服务器启动程序（接受 -port 参数） |
| `start_server.sh` | 启动多个服务器实例 |
| `stop_server.sh` | 停止所有服务器 |
| `status_server.sh` | 查看服务器运行状态 |
| `TESTING_GUIDE.md` | 详细的测试指南和设计思路 |

## 服务器端点

每个服务器提供两个端点：

1. **`GET /`** - 返回服务器端口信息
   ```json
   {"data":"Hello, you are run on port: 8081"}
   ```

2. **`GET /health`** - 健康检查端点
   ```json
   {"data":"healthy"}
   ```

## 目录结构（自动创建）

```
load-balancer/
├── bin/              # 编译后的可执行文件
│   └── server
├── logs/             # 服务器日志
│   ├── server-8081.log
│   ├── server-8082.log
│   └── ...
└── pids/             # 进程 PID 文件
    ├── server-8081.pid
    ├── server-8082.pid
    └── ...
```

## 自定义端口

编辑 `start_server.sh` 的第 10 行：

```bash
# 修改端口列表
PORTS=(8081 8082 8083 8084 8085 8086 8087)
```

## 查看日志

```bash
# 查看单个服务器日志
tail -f logs/server-8081.log

# 查看所有服务器日志
tail -f logs/server-*.log
```

## 下一步

现在你可以开始实现负载均衡器了！参考 `DEVELOPMENT_GUIDE.md` 进行开发。

测试负载均衡器时：
1. 启动多个后端服务器：`./start_server.sh`
2. 启动你的负载均衡器
3. 发送请求，观察负载分布
4. 查看各服务器日志验证

**提示**：查看 `TESTING_GUIDE.md` 了解更多测试场景和高级用法。

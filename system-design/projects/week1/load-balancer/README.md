# 负载均衡器实现

## 项目概述

使用 Go 实现一个简单的负载均衡器,支持多种负载均衡算法和健康检查功能。

## 学习目标

- 理解负载均衡的工作原理
- 掌握常见的负载均衡算法
- 学习健康检查机制
- 实践 Go 语言的并发编程

## 功能特性

- [ ] Round Robin (轮询) 算法
- [ ] Weighted Round Robin (加权轮询) 算法
- [ ] Least Connections (最少连接) 算法
- [ ] 后端服务器健康检查
- [ ] 故障服务器自动剔除
- [ ] 服务器恢复后自动加入

## 项目结构

```
load-balancer/
├── main.go              # 主程序入口
├── go.mod              # Go 模块配置
├── go.sum
├── README.md
├── pkg/
│   ├── server.go       # 后端服务器定义
│   ├── pool.go         # 服务器池管理
│   └── algorithms/
│       ├── round_robin.go
│       ├── weighted_round_robin.go
│       └── least_connections.go
├── internal/
│   └── healthcheck/
│       └── checker.go  # 健康检查
└── tests/
    └── load_balancer_test.go
```

## 技术栈

- Go 1.23+
- net/http (HTTP 服务)
- sync (并发控制)
- time (定时任务)

## 快速开始

### 1. 初始化项目

```bash
cd projects/week1/load-balancer
go mod init github.com/yourusername/load-balancer
```

### 2. 创建后端测试服务器

```bash
# 终端1: 启动服务器1
go run test-server/main.go -port 8081

# 终端2: 启动服务器2
go run test-server/main.go -port 8082

# 终端3: 启动服务器3
go run test-server/main.go -port 8083
```

### 3. 运行负载均衡器

```bash
go run main.go
```

### 4. 测试

```bash
# 发送多个请求,观察负载分配
for i in {1..10}; do
  curl http://localhost:8080
done
```

## 实现要点

### 1. Round Robin 算法

```go
type RoundRobin struct {
    servers []*Server
    current int
    mu      sync.Mutex
}

func (rr *RoundRobin) NextServer() *Server {
    rr.mu.Lock()
    defer rr.mu.Unlock()

    server := rr.servers[rr.current]
    rr.current = (rr.current + 1) % len(rr.servers)

    return server
}
```

### 2. 健康检查

```go
func (p *ServerPool) HealthCheck() {
    ticker := time.NewTicker(30 * time.Second)

    for range ticker.C {
        for _, server := range p.servers {
            if !server.IsHealthy() {
                server.SetAlive(false)
            } else {
                server.SetAlive(true)
            }
        }
    }
}
```

### 3. 请求转发

```go
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    server := lb.algorithm.NextServer()

    if server == nil {
        http.Error(w, "No available servers", http.StatusServiceUnavailable)
        return
    }

    server.ReverseProxy.ServeHTTP(w, r)
}
```

## 测试用例

```go
func TestRoundRobin(t *testing.T) {
    servers := []*Server{
        NewServer("http://localhost:8081"),
        NewServer("http://localhost:8082"),
        NewServer("http://localhost:8083"),
    }

    rr := NewRoundRobin(servers)

    // 测试轮询顺序
    assert.Equal(t, servers[0], rr.NextServer())
    assert.Equal(t, servers[1], rr.NextServer())
    assert.Equal(t, servers[2], rr.NextServer())
    assert.Equal(t, servers[0], rr.NextServer()) // 循环
}
```

## 扩展功能

完成基础功能后,可以尝试:

- [ ] 实现 IP Hash 算法
- [ ] 添加请求日志记录
- [ ] 支持 HTTPS
- [ ] 实现会话保持 (Session Persistence)
- [ ] 添加 Prometheus 监控指标
- [ ] 支持动态添加/删除后端服务器

## 参考资料

- [Nginx 负载均衡算法](https://nginx.org/en/docs/http/load_balancing.html)
- [Go net/http/httputil](https://pkg.go.dev/net/http/httputil)
- [HAProxy Documentation](http://www.haproxy.org/)

## 学习笔记

记录学习过程中的要点:

### 收获
- ...

### 遇到的问题
- ...

### 改进想法
- ...

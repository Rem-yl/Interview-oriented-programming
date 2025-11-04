# 快速开始

```bash
cd system-design/projects/week1/module2/load-balancer
```


1. 启动服务器
```bash
bash scripts/start_server.sh
```

2. 启动`load_balance`服务
```bash
go run cmd/lb/main.go
```

3. 使用客户端访问
```bash
go run cmd/client/main.go --port 8187
```

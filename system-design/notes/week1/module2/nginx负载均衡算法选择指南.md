# NGINX 负载均衡算法选择指南

> 参考: NGINX官方文档 - Choosing a Load Balancing Method

## 目录

1. [负载均衡算法概述](#负载均衡算法概述)
2. [Round Robin (轮询)](#round-robin-轮询)
3. [Least Connections (最少连接)](#least-connections-最少连接)
4. [IP Hash (IP哈希)](#ip-hash-ip哈希)
5. [Generic Hash (通用哈希)](#generic-hash-通用哈希)
6. [Least Time (最短时间)](#least-time-最短时间)
7. [Random (随机)](#random-随机)
8. [算法选择决策树](#算法选择决策树)

---

## 负载均衡算法概述

NGINX和NGINX Plus支持多种负载均衡算法，每种算法适用于不同的场景。选择合适的算法需要考虑：

- **应用特性**: 无状态 vs 有状态
- **流量模式**: 均匀 vs 不均匀
- **后端差异**: 性能一致 vs 性能差异大
- **会话要求**: 需要会话保持 vs 不需要

### 算法分类

**静态算法** (不考虑服务器当前状态):
- Round Robin (轮询)
- Weighted Round Robin (加权轮询)
- IP Hash (IP哈希)
- Generic Hash (通用哈希)
- Random (随机)

**动态算法** (根据服务器当前状态选择):
- Least Connections (最少连接)
- Weighted Least Connections (加权最少连接)
- Least Time (最短时间) - NGINX Plus专有

---

## Round Robin (轮询)

### 工作原理

按顺序将请求依次分配给每个服务器。这是**NGINX的默认算法**。

```nginx
upstream backend {
    server backend1.example.com;
    server backend2.example.com;
    server backend3.example.com;
}
```

### 加权轮询 (Weighted Round Robin)

为不同性能的服务器分配不同权重:

```nginx
upstream backend {
    server backend1.example.com weight=5;  # 处理5/7的请求
    server backend2.example.com weight=1;  # 处理1/7的请求
    server backend3.example.com weight=1;  # 处理1/7的请求
}
```

### 优点

✅ **简单高效**: 算法实现简单，性能开销小
✅ **公平分配**: 每个服务器获得相同的请求数(无权重时)
✅ **适用广泛**: 适合无状态应用

### 缺点

❌ **不考虑负载**: 不管服务器当前负载如何
❌ **长连接问题**: 如果有些连接持续时间长，会导致不均衡
❌ **无会话保持**: 同一用户的请求可能被分配到不同服务器

### 适用场景

- ✅ 无状态应用 (RESTful API)
- ✅ 短连接请求
- ✅ 后端服务器性能一致
- ✅ 请求处理时间相近

### 不适用场景

- ❌ 需要会话保持的应用
- ❌ 长连接应用 (WebSocket)
- ❌ 请求处理时间差异大

---

## Least Connections (最少连接)

### 工作原理

将请求分配给**当前活动连接数最少**的服务器。这是一种**动态算法**。

```nginx
upstream backend {
    least_conn;  # 启用最少连接算法

    server backend1.example.com;
    server backend2.example.com;
    server backend3.example.com;
}
```

### 加权最少连接 (Weighted Least Connections)

结合权重和连接数:

```nginx
upstream backend {
    least_conn;

    server backend1.example.com weight=3;
    server backend2.example.com weight=1;
    server backend3.example.com weight=1;
}
```

**选择逻辑**: 选择 `活动连接数 / 权重` 最小的服务器

### 优点

✅ **负载均衡**: 考虑服务器当前负载
✅ **适应性强**: 自动适应请求处理时间的差异
✅ **处理长连接**: 能较好地处理长连接场景

### 缺点

❌ **复杂度高**: 需要维护连接计数
❌ **性能开销**: 比轮询算法稍慢
❌ **可能不稳定**: 快速变化的负载下可能不如轮询稳定

### 适用场景

- ✅ 请求处理时间差异大
- ✅ 长连接应用 (HTTP/1.1 keep-alive, HTTP/2)
- ✅ 后端服务器性能不一致
- ✅ 数据库连接池

### 实际案例

**场景**: 视频流媒体服务
- 某些用户看完整视频(长连接)
- 某些用户只看几秒就关闭(短连接)
- 使用 Least Connections 可以避免部分服务器因长连接过多而过载

---

## IP Hash (IP哈希)

### 工作原理

根据客户端IP地址的哈希值选择服务器，**同一IP总是被分配到同一台服务器**。

```nginx
upstream backend {
    ip_hash;  # 启用IP哈希

    server backend1.example.com;
    server backend2.example.com;
    server backend3.example.com;
}
```

### 哈希计算

```
hash_value = hash(client_ip)
server_index = hash_value % server_count
```

**注意**: 使用客户端IP的**前3个字节**(IPv4)进行哈希

### 优点

✅ **会话保持**: 同一用户总是访问同一服务器
✅ **简单稳定**: 实现简单，分配稳定
✅ **无需Session共享**: 不需要集中式Session存储

### 缺点

❌ **分布不均**: 如果用户IP分布不均，会导致负载不均
❌ **扩容问题**: 添加/删除服务器会导致大量重新分配
❌ **代理问题**: NAT/代理后多个用户可能共享同一IP

### 适用场景

- ✅ 需要会话保持的应用
- ✅ 购物车系统
- ✅ 用户登录状态管理
- ✅ 服务器数量相对固定

### 不适用场景

- ❌ 频繁扩容/缩容
- ❌ 用户通过代理访问(如企业网络)
- ❌ 对负载均衡要求很高

### 改进方案: 一致性哈希

见后续章节的 Generic Hash

---

## Generic Hash (通用哈希)

### 工作原理

根据**自定义的键**进行哈希，比IP Hash更灵活。

```nginx
upstream backend {
    hash $request_uri consistent;  # 根据URI哈希，使用一致性哈希

    server backend1.example.com;
    server backend2.example.com;
    server backend3.example.com;
}
```

### 常用哈希键

#### 1. 基于URI (缓存友好)

```nginx
hash $request_uri consistent;
```

**适用**: CDN、缓存服务器

#### 2. 基于Cookie (用户会话)

```nginx
hash $cookie_jsessionid consistent;
```

**适用**: 需要会话保持的Java应用

#### 3. 基于请求参数

```nginx
hash $arg_user_id consistent;
```

**适用**: 基于用户ID的分片

#### 4. 组合键

```nginx
map $request_uri $route_key {
    default "$request_uri";
}
hash $route_key consistent;
```

### 一致性哈希 (Consistent Hashing)

添加 `consistent` 参数启用一致性哈希:

```nginx
hash $request_uri consistent;
```

**优点**:
- ✅ 添加/删除服务器时，只有 1/N 的数据需要重新分配
- ✅ 更好的扩展性

**实现原理**: 使用虚拟节点构建哈希环

### 优点

✅ **灵活性高**: 可以根据任何变量哈希
✅ **缓存友好**: 同一资源总是路由到同一服务器
✅ **可扩展**: 一致性哈希支持动态扩缩容

### 缺点

❌ **配置复杂**: 需要选择合适的哈希键
❌ **分布问题**: 哈希键分布不均会导致负载不均

### 适用场景

#### 场景1: 缓存服务器

```nginx
upstream cache_backend {
    hash $request_uri consistent;

    server cache1.example.com;
    server cache2.example.com;
    server cache3.example.com;
}
```

同一URL总是路由到同一台缓存服务器，提高缓存命中率

#### 场景2: 有状态服务分片

```nginx
upstream user_service {
    hash $arg_user_id consistent;

    server user1.example.com;
    server user2.example.com;
}
```

根据用户ID分片，每台服务器只处理部分用户

---

## Least Time (最短时间)

> **注意**: 这是 NGINX Plus 的商业特性

### 工作原理

选择**响应时间最短且活动连接数最少**的服务器。

```nginx
upstream backend {
    least_time header;  # 基于响应头时间

    server backend1.example.com;
    server backend2.example.com;
    server backend3.example.com;
}
```

### 模式

#### 1. `least_time header`

基于接收到响应头的时间

#### 2. `least_time last_byte`

基于接收到完整响应的时间

### 计算公式

```
score = (active_connections + 1) / weight
选择 score 最小且 average_response_time 最短的服务器
```

### 优点

✅ **性能优化**: 自动路由到响应最快的服务器
✅ **智能感知**: 能感知服务器性能差异
✅ **实时调整**: 根据实时性能动态调整

### 缺点

❌ **商业版专有**: 需要NGINX Plus
❌ **复杂度高**: 需要持续监控响应时间
❌ **可能不稳定**: 在快速变化的网络环境下可能频繁切换

### 适用场景

- ✅ 后端服务器性能差异大
- ✅ 对延迟敏感的应用
- ✅ 混合云环境(不同地理位置的服务器)
- ✅ API网关

---

## Random (随机)

### 工作原理

**随机**选择一台服务器处理请求。

```nginx
upstream backend {
    random;  # 启用随机算法

    server backend1.example.com;
    server backend2.example.com;
    server backend3.example.com;
}
```

### 加权随机

```nginx
upstream backend {
    random two least_conn;  # 随机选2台，再用least_conn选择

    server backend1.example.com weight=3;
    server backend2.example.com weight=1;
    server backend3.example.com weight=1;
}
```

### Two Random Choices

```nginx
random two;           # 随机选2台，选连接数少的
random two least_time;  # 随机选2台，选响应时间短的
```

**原理**: Power of Two Choices 算法
- 随机选择2台服务器
- 从这2台中选择负载更小的
- 性能接近全局最优，但开销更小

### 优点

✅ **简单**: 实现非常简单
✅ **无状态**: 不需要维护任何状态
✅ **分布均匀**: 大量请求下分布趋于均匀

### 缺点

❌ **短期不均**: 短时间内可能分布不均
❌ **不够智能**: 不考虑服务器负载

### 适用场景

- ✅ 大规模集群 (服务器数量多)
- ✅ 请求量大
- ✅ 与其他算法结合使用 (two random choices)

---

## 算法选择决策树

```
开始
  |
  ├─ 需要会话保持?
  │   ├─ 是 → 用户识别方式?
  │   │   ├─ IP地址 → IP Hash
  │   │   ├─ Cookie → Generic Hash (cookie)
  │   │   └─ 用户ID → Generic Hash (user_id)
  │   │
  │   └─ 否 → 继续
  │
  ├─ 是否有缓存需求?
  │   ├─ 是 → Generic Hash ($request_uri) + consistent
  │   └─ 否 → 继续
  │
  ├─ 请求处理时间差异大?
  │   ├─ 是 → 连接类型?
  │   │   ├─ 长连接 → Least Connections
  │   │   └─ 短连接 → Least Time (Plus) 或 Random Two
  │   │
  │   └─ 否 → 继续
  │
  ├─ 后端性能一致?
  │   ├─ 是 → Round Robin
  │   └─ 否 → Weighted Round Robin 或 Least Connections
  │
  └─ 大规模集群 (100+服务器)?
      ├─ 是 → Random Two + Least Connections
      └─ 否 → Round Robin (默认)
```

---

## 实际场景示例

### 场景1: 电商网站

**需求**:
- 需要购物车会话保持
- 流量大，峰值明显
- 部分页面可缓存

**方案**:

```nginx
# 静态资源 - 缓存友好
upstream static_backend {
    hash $request_uri consistent;
    server static1.example.com;
    server static2.example.com;
    server static3.example.com;
}

# 动态请求 - 会话保持
upstream app_backend {
    ip_hash;
    server app1.example.com weight=3;
    server app2.example.com weight=2;
    server app3.example.com weight=2;
}

# API请求 - 负载均衡
upstream api_backend {
    least_conn;
    server api1.example.com;
    server api2.example.com;
    server api3.example.com;
}
```

---

### 场景2: 视频流媒体

**需求**:
- 长连接 (流式传输)
- 请求处理时间差异大 (视频长度不同)
- 需要优化带宽利用

**方案**:

```nginx
upstream video_backend {
    least_conn;  # 避免单台服务器长连接过多

    server video1.example.com max_conns=50;
    server video2.example.com max_conns=50;
    server video3.example.com max_conns=50;
}
```

---

### 场景3: 微服务API网关

**需求**:
- 多个微服务
- 不同服务性能差异大
- 需要快速响应

**方案**:

```nginx
# 用户服务 - 访问频繁，需要缓存
upstream user_service {
    hash $arg_user_id consistent;
    server user1.example.com;
    server user2.example.com;
}

# 订单服务 - 请求时间差异大
upstream order_service {
    least_conn;
    server order1.example.com weight=2;
    server order2.example.com weight=1;
}

# 搜索服务 - 计算密集，需要负载均衡
upstream search_service {
    random two least_conn;
    server search1.example.com;
    server search2.example.com;
    server search3.example.com;
}
```

---

### 场景4: WebSocket应用

**需求**:
- 持久连接
- 需要会话保持

**方案**:

```nginx
upstream websocket_backend {
    # 方案1: IP Hash (简单)
    ip_hash;

    # 方案2: 基于Cookie (更准确)
    # hash $cookie_session_id consistent;

    server ws1.example.com max_conns=1000;
    server ws2.example.com max_conns=1000;
    server ws3.example.com max_conns=1000;
}

server {
    location /ws {
        proxy_pass http://websocket_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

---

## 高级配置技巧

### 1. 健康检查配合

```nginx
upstream backend {
    least_conn;

    server backend1.example.com max_fails=3 fail_timeout=30s;
    server backend2.example.com max_fails=3 fail_timeout=30s;
    server backend3.example.com backup;  # 备份服务器
}
```

### 2. 慢启动 (NGINX Plus)

```nginx
upstream backend {
    least_conn;

    server backend1.example.com slow_start=30s;  # 新服务器30秒内逐渐增加流量
    server backend2.example.com;
}
```

### 3. 连接限制

```nginx
upstream backend {
    least_conn;

    server backend1.example.com max_conns=100;  # 限制最大连接数
    server backend2.example.com max_conns=100;
}
```

### 4. 排空服务器 (NGINX Plus)

```nginx
upstream backend {
    server backend1.example.com;
    server backend2.example.com;
    server backend3.example.com drain;  # 不接受新连接，等待现有连接结束
}
```

---

## 性能对比

| 算法 | 复杂度 | 分布均匀性 | 会话保持 | 动态适应 | 适用规模 |
|------|--------|------------|----------|----------|----------|
| Round Robin | O(1) | ⭐⭐⭐⭐⭐ | ❌ | ❌ | 小-大 |
| Weighted RR | O(N) | ⭐⭐⭐⭐ | ❌ | ❌ | 小-中 |
| Least Conn | O(N) | ⭐⭐⭐⭐ | ❌ | ✅ | 小-中 |
| IP Hash | O(1) | ⭐⭐⭐ | ✅ | ❌ | 小-中 |
| Generic Hash | O(1) | ⭐⭐⭐⭐ | ✅ | ❌ | 小-大 |
| Least Time | O(N) | ⭐⭐⭐⭐⭐ | ❌ | ✅✅ | 小-中 |
| Random | O(1) | ⭐⭐⭐⭐ | ❌ | ❌ | 大 |
| Random Two | O(1) | ⭐⭐⭐⭐⭐ | ❌ | ✅ | 大 |

---

## 总结与最佳实践

### 通用建议

1. **从简单开始**: 如果不确定，先用默认的 Round Robin
2. **监控后调整**: 根据实际监控数据选择合适的算法
3. **混合使用**: 不同类型的请求可以使用不同的算法
4. **测试验证**: 在生产环境前充分测试

### 快速选择指南

- **默认选择**: Round Robin
- **长连接**: Least Connections
- **会话保持**: IP Hash 或 Generic Hash
- **缓存优化**: Generic Hash (URI)
- **性能优化**: Least Time (Plus) 或 Random Two
- **大规模**: Random Two + Least Connections

### 常见误区

❌ **误区1**: "Least Connections 总是比 Round Robin 好"
- 实际: 对于短连接、处理时间一致的场景，Round Robin 更高效

❌ **误区2**: "IP Hash 能完美保持会话"
- 实际: NAT/代理会导致多个用户共享IP

❌ **误区3**: "加权重就能精确控制流量比例"
- 实际: 只能在长期统计上接近目标比例

❌ **误区4**: "算法越复杂越好"
- 实际: 简单算法往往更稳定、性能更好

---

## 参考资料

1. [NGINX官方文档 - HTTP Load Balancing](http://nginx.org/en/docs/http/load_balancing.html)
2. [NGINX Plus Admin Guide](https://docs.nginx.com/nginx/admin-guide/load-balancer/)
3. [The Power of Two Random Choices](https://brooker.co.za/blog/2012/01/17/two-random.html)
4. [Consistent Hashing](https://en.wikipedia.org/wiki/Consistent_hashing)

---

**下一步学习**:
1. 实现 Least Connections 算法
2. 对比不同算法在实际场景中的表现
3. 学习一致性哈希的实现细节

💡 **记住**: 没有"最好"的算法，只有"最合适"的算法!

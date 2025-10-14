# 系统设计学习路线

> 从零开始的系统设计之旅 - 面向 Python & Go 开发者

## 学习目标

通过系统化学习,掌握大规模分布式系统设计的核心概念和实践技能,能够设计、实现和优化真实世界的后端系统。

## 学习周期规划

- **总时长**: 12 周 (约 3 个月)
- **每日学习**: 2-3 小时
- **学习模式**: 理论 + 实践 + 每日打卡

---

## 📚 学习路线图

### Phase 1: 基础篇 (第 1-3 周)

#### Week 1: 系统设计基础概念
**目标**: 建立系统设计的整体认知框架

- **Day 1-2**: 系统设计面试概述
  - 什么是系统设计
  - 系统设计的常见问题类型
  - 如何进行系统设计面试
  - 资源推荐: [Grokking the System Design Interview](https://www.educative.io/courses/grokking-the-system-design-interview)

- **Day 3-4**: 可扩展性基础 (Scalability)
  - 垂直扩展 vs 水平扩展
  - 无状态 vs 有状态服务
  - 负载均衡基础
  - **实践**: 用 Go 实现简单的负载均衡器

- **Day 5-6**: 性能指标与估算
  - QPS, TPS, Latency, Throughput
  - 容量估算 (Capacity Estimation)
  - 回退计算 (Back-of-envelope Calculation)
  - **实践**: 估算 Twitter 的存储需求

- **Day 7**: 周总结与复习
  - 整理笔记
  - 完成基础概念测试题
  - 准备下周学习

#### Week 2: 数据存储基础
**目标**: 理解不同存储系统的特点和应用场景

- **Day 1-2**: 关系型数据库 (SQL)
  - ACID 特性
  - 索引原理 (B-Tree, Hash)
  - 数据库分片 (Sharding)
  - **实践**: Python + PostgreSQL 实现分片策略

- **Day 3-4**: NoSQL 数据库
  - Key-Value Store (Redis)
  - Document Store (MongoDB)
  - Column-Family Store (Cassandra)
  - Graph Database (Neo4j)
  - **实践**: Go + Redis 实现缓存层

- **Day 5-6**: CAP 定理与一致性
  - CAP 理论详解
  - 最终一致性 vs 强一致性
  - BASE 理论
  - **实践**: 实现最终一致性示例

- **Day 7**: 周总结
  - 对比不同数据库的选型
  - 完成数据存储决策树

#### Week 3: 网络与通信基础
**目标**: 掌握分布式系统的通信机制

- **Day 1-2**: HTTP/HTTPS 深入
  - RESTful API 设计原则
  - HTTP/2 vs HTTP/3
  - WebSocket 实时通信
  - **实践**: Go 实现 RESTful API 服务

- **Day 3-4**: RPC 与消息队列
  - gRPC 原理与实践
  - Protocol Buffers
  - 消息队列基础 (RabbitMQ, Kafka)
  - **实践**: Python gRPC 服务示例

- **Day 5-6**: API 网关与服务网格
  - API Gateway 设计模式
  - 限流、熔断、降级
  - Service Mesh 简介
  - **实践**: 实现简单的 API 限流器

- **Day 7**: Phase 1 总结
  - 完成基础知识测试
  - 整理学习笔记
  - 准备进入中级阶段

---

### Phase 2: 进阶篇 (第 4-7 周)

#### Week 4: 缓存系统设计
**目标**: 深入理解缓存策略与实现

- **Day 1-2**: 缓存基础
  - 缓存淘汰策略 (LRU, LFU, FIFO)
  - 缓存穿透、击穿、雪崩
  - 多级缓存架构
  - **实践**: Go 实现 LRU Cache

- **Day 3-4**: 分布式缓存
  - Redis 集群架构
  - 一致性哈希 (Consistent Hashing)
  - 缓存预热与更新策略
  - **实践**: Python 实现一致性哈希环

- **Day 5-6**: CDN 与边缘缓存
  - CDN 工作原理
  - 静态资源优化
  - 缓存失效策略
  - **实践**: 设计图片缓存系统

- **Day 7**: 缓存系统设计实战
  - 设计题: 实现一个缓存系统

#### Week 5: 数据库进阶
**目标**: 掌握数据库扩展与优化技巧

- **Day 1-2**: 数据库分片 (Sharding)
  - 水平分片 vs 垂直分片
  - 分片键选择
  - 跨分片查询
  - **实践**: Go 实现数据库分片中间件

- **Day 3-4**: 读写分离与主从复制
  - Master-Slave 架构
  - 复制延迟处理
  - 故障转移 (Failover)
  - **实践**: Python 实现读写分离代理

- **Day 5-6**: 数据库优化
  - 索引优化
  - 查询优化
  - 连接池管理
  - **实践**: SQL 性能分析与优化

- **Day 7**: 数据库设计实战
  - 设计题: 设计 Instagram 的数据库架构

#### Week 6: 消息队列与异步处理
**目标**: 理解异步系统设计模式

- **Day 1-2**: 消息队列深入
  - Kafka 架构与原理
  - RabbitMQ vs Kafka
  - 消息可靠性保证
  - **实践**: Go + Kafka 实现消息生产消费

- **Day 3-4**: 事件驱动架构
  - Event Sourcing
  - CQRS 模式
  - Saga 分布式事务
  - **实践**: Python 实现事件驱动系统

- **Day 5-6**: 任务队列与异步任务
  - Celery 原理 (Python)
  - 延迟任务与定时任务
  - 任务重试与幂等性
  - **实践**: Celery + Redis 异步任务系统

- **Day 7**: 消息系统设计实战
  - 设计题: 设计通知系统

#### Week 7: 微服务架构
**目标**: 掌握微服务设计原则

- **Day 1-2**: 微服务基础
  - 单体 vs 微服务
  - 服务拆分原则
  - 服务间通信
  - **实践**: Go 实现微服务框架基础

- **Day 3-4**: 服务发现与配置管理
  - Consul, Etcd, Zookeeper
  - 服务注册与发现
  - 配置中心设计
  - **实践**: Python + Consul 服务注册

- **Day 5-6**: 容错与弹性设计
  - 熔断器模式 (Circuit Breaker)
  - 重试与超时
  - 降级与限流
  - **实践**: Go 实现熔断器

- **Day 7**: Phase 2 总结
  - 完成进阶测试
  - 微服务架构实战设计

---

### Phase 3: 高级篇 (第 8-10 周)

#### Week 8: 分布式系统理论
**目标**: 理解分布式系统的核心理论

- **Day 1-2**: 分布式一致性
  - Paxos 算法
  - Raft 算法
  - 2PC/3PC 协议
  - **实践**: Go 实现简化版 Raft

- **Day 3-4**: 分布式锁与协调
  - 基于 Redis 的分布式锁
  - 基于 Zookeeper 的分布式锁
  - Redlock 算法
  - **实践**: Python 实现分布式锁

- **Day 5-6**: 分布式事务
  - TCC 模式
  - Saga 模式
  - 本地消息表
  - **实践**: Go 实现 Saga 模式

- **Day 7**: 分布式理论总结
  - 完成分布式理论测试

#### Week 9: 监控、日志与追踪
**目标**: 掌握系统可观测性设计

- **Day 1-2**: 日志系统
  - 集中式日志架构 (ELK)
  - 日志采集与聚合
  - 日志分析与告警
  - **实践**: Go 结构化日志实现

- **Day 3-4**: 监控系统
  - Prometheus + Grafana
  - 指标采集与可视化
  - 告警规则设计
  - **实践**: Python 应用监控接入

- **Day 5-6**: 分布式追踪
  - OpenTelemetry
  - Jaeger 链路追踪
  - 追踪数据分析
  - **实践**: Go 微服务链路追踪

- **Day 7**: 可观测性设计实战
  - 设计完整的监控系统

#### Week 10: 安全与性能优化
**目标**: 提升系统安全性和性能

- **Day 1-2**: 系统安全设计
  - 认证与授权 (OAuth2, JWT)
  - API 安全最佳实践
  - SQL 注入与 XSS 防护
  - **实践**: Go 实现 JWT 认证

- **Day 3-4**: 性能优化策略
  - 数据库查询优化
  - 代码层面优化
  - 并发优化
  - **实践**: Python 性能分析工具使用

- **Day 5-6**: 高可用架构
  - 多活架构
  - 容灾设计
  - 灰度发布
  - **实践**: 设计高可用方案

- **Day 7**: Phase 3 总结
  - 完成高级知识测试

---

### Phase 4: 实战篇 (第 11-12 周)

#### Week 11: 经典系统设计案例
**目标**: 通过经典案例巩固知识

- **Day 1**: 设计 URL 短链接服务
  - 需求分析
  - 架构设计
  - 实现要点

- **Day 2**: 设计 Twitter/微博
  - Feed 流设计
  - 时间线算法
  - 扩展性考虑

- **Day 3**: 设计 Instagram
  - 图片存储
  - Feed 生成
  - 关注关系处理

- **Day 4**: 设计分布式文件系统
  - GFS/HDFS 原理
  - 数据分布与复制
  - 故障恢复

- **Day 5**: 设计网约车系统
  - 地理位置索引
  - 实时匹配算法
  - 高并发处理

- **Day 6**: 设计秒杀系统
  - 高并发设计
  - 库存扣减
  - 防刷策略

- **Day 7**: 案例总结与复盘

#### Week 12: 综合项目实战
**目标**: 完成一个完整的系统设计项目

- **Day 1-2**: 项目选择与需求分析
  - 选择感兴趣的系统
  - 完成需求文档
  - 技术选型

- **Day 3-5**: 系统实现
  - 核心功能开发
  - Go/Python 混合开发
  - 单元测试

- **Day 6**: 压力测试与优化
  - 性能测试
  - 瓶颈分析
  - 优化改进

- **Day 7**: 项目总结与展示
  - 完成项目文档
  - 架构图绘制
  - 学习总结

---

## 🛠️ 技术栈推荐

### Python 生态
- **Web 框架**: FastAPI, Django, Flask
- **异步**: asyncio, aiohttp
- **消息队列**: Celery, RQ
- **数据库**: SQLAlchemy, psycopg2, pymongo, redis-py
- **测试**: pytest, unittest
- **监控**: prometheus_client

### Go 生态
- **Web 框架**: Gin, Echo, Fiber
- **RPC**: gRPC, go-micro
- **数据库**: GORM, sqlx, go-redis
- **消息队列**: Sarama (Kafka), AMQP (RabbitMQ)
- **测试**: testify, gomock
- **监控**: prometheus/client_golang

---

## 📖 学习资源

### 书籍
1. **《Designing Data-Intensive Applications》** - Martin Kleppmann (必读)
2. **《System Design Interview》** - Alex Xu
3. **《微服务架构设计模式》** - Chris Richardson
4. **《高性能 MySQL》** - Baron Schwartz

### 在线课程
1. [Grokking the System Design Interview](https://www.educative.io/courses/grokking-the-system-design-interview)
2. [System Design Primer](https://github.com/donnemartin/system-design-primer)
3. [ByteByteGo](https://bytebytego.com/)

### 博客与社区
- [High Scalability](http://highscalability.com/)
- [Martin Fowler's Blog](https://martinfowler.com/)
- [Engineering Blogs](https://github.com/kilimchoi/engineering-blogs)

### YouTube 频道
- [Gaurav Sen](https://www.youtube.com/c/GauravSensei)
- [Tech Dummies](https://www.youtube.com/c/TechDummiesNarendraL)
- [System Design Interview](https://www.youtube.com/c/SystemDesignInterview)

---

## 📝 学习方法建议

### 1. 理论 + 实践结合
- 每学习一个概念,必须动手实践
- 代码实现加深理解
- 建立 GitHub 仓库记录学习成果

### 2. 画图思考
- 学会画系统架构图
- 使用工具: draw.io, Excalidraw, PlantUML
- 将抽象概念可视化

### 3. 总结输出
- 每周写学习总结
- 整理笔记到 Markdown
- 可以考虑写博客分享

### 4. 面试导向
- 每个主题结束后做相关面试题
- 模拟系统设计面试
- 练习口头表达设计思路

### 5. 社区交流
- 加入技术社区讨论
- 参与开源项目
- Code Review 学习他人代码

---

## 🎯 每日学习流程建议

1. **理论学习** (40 分钟)
   - 阅读文档/教程
   - 观看视频课程
   - 记录笔记

2. **代码实践** (60-90 分钟)
   - 完成每日代码练习
   - 实现示例项目
   - 调试与优化

3. **复习总结** (20-30 分钟)
   - 整理当天学习内容
   - 更新学习笔记
   - 完成打卡任务

4. **拓展阅读** (可选)
   - 阅读相关博客文章
   - 查看开源项目实现
   - 技术社区讨论

---

## 📊 学习进度追踪

详见 [每日打卡清单](./DAILY_CHECKLIST.md)

---

## 🎓 结业目标

完成本学习路线后,你将能够:

1. ✅ 独立设计中大型分布式系统
2. ✅ 评估不同技术方案的优劣
3. ✅ 识别系统瓶颈并提出优化方案
4. ✅ 通过系统设计面试
5. ✅ 用 Go/Python 实现系统设计方案
6. ✅ 理解主流互联网架构模式

---

## 📞 反馈与更新

- 学习过程中遇到问题,及时记录
- 根据实际情况调整学习节奏
- 定期回顾和更新学习路线

**祝你学习愉快,早日成为系统设计专家!** 🚀

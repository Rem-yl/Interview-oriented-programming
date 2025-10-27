# 系统设计深度学习 - 学习进度追踪

> 12周系统化课程 - 模块化学习,深度优先

## 📋 如何使用本清单

本清单采用**模块化追踪**,而非严格的每日任务:

1. **按模块学习**: 每周分为多个模块,每个模块专注一个核心主题
2. **深度优先**: 完成模块内所有内容(理论+源码+实践)后再进入下一模块
3. **灵活安排**: 根据理解深度调整学习节奏,不必严格按天数
4. **记录学习**: 在 `notes/` 目录记录深入思考和源码阅读笔记
5. **代码实践**: 每个模块的项目必须完成,代码质量高于速度

### 标记说明

- `[ ]` 未开始
- `[~]` 进行中
- `[x]` 已完成
- `[!]` 需要加强理解

---

## 第一阶段: 基础原理与核心组件 (Week 1-3)

### Week 1: 可扩展性原理与负载均衡

**本周目标**: 理解系统如何从单机走向分布式

#### 模块 1: 可扩展性基础 (建议 2-3 天)

**核心学习内容**:

**1. 垂直扩展 vs 水平扩展**

- [X] 📖 阅读《Designing Data-Intensive Applications》Chapter 1 → 📝 [笔记](notes/resources/books/DDIA/Ch01_scalability.md)
- [X] 📺 观看视频: [Horizontal vs Vertical Scaling](https://www.youtube.com/watch?v=xpDnVSmNFX0) (15分钟)
- [X] 📄 阅读文章: [Scalability for Dummies](https://www.lecloud.net/post/7295452622/scalability-for-dummies-part-1-clones) - 4篇系列文章
- [X] ✍️ **任务**: 列表对比垂直扩展与水平扩展 → 📝 [对比表格](notes/resources/books/DDIA/Ch01_scalability.md)

**2. CAP 定理初探** -> 📝 [笔记](notes/resources/books/DDIA/ch09_note_and_question.md)

- [X] 📖 阅读《DDIA》Chapter 9: Consistency and Consensus → 📝 [笔记](notes/resources/books/DDIA/Ch09_一致性与共识.md)
- [X] 📄 阅读论文: [CAP Twelve Years Later](https://www.infoq.com/articles/cap-twelve-years-later-how-the-rules-have-changed/) → 📝 [笔记](notes/resources/papers/CAP-Twelve-Years-Later.md)
- [X] 📺 观看视频: [What is CAP Theorem?](https://www.youtube.com/watch?v=k-Yaq8AHlFA) (8分钟)
- [X] ✍️ **任务**: CAP 权衡案例分析 → 📝 [案例](notes/week1/module1-scalability.md#cap-定理案例分析)
  - 核心理解: 分区不可避免，节点无法确定其他节点状态（网络故障 vs 崩溃）
  - ATM 案例: 保证可用性会导致不一致；保证一致性要拒绝服务

**3. 无状态服务设计** -> 📝 [笔记](notes/week1/module1/module1-3-stateless-service.md)

- [X] 📄 阅读: [The Twelve-Factor App](https://12factor.net/) - 特别关注 VI. Processes 章节
- [X] 📄 阅读: [Stateless vs Stateful Services](https://medium.com/@maniakhitoccori/stateless-vs-stateful-architecture-63194d749c08)
- [X] 📺 观看: [Stateless Architecture](https://www.youtube.com/watch?v=20tpk8A_xa0&t=14s) (10分钟)
- [X] ✍️ **任务**: 设计一个无状态的用户认证系统架构图

**4. 会话状态管理** -> 📝 [笔记](notes/week1/module1/module1-4-session-management.md)

- [X] 📄 阅读: [Session Management Strategies](https://stackoverflow.blog/2021/10/06/best-practices-for-authentication-and-authorization-for-rest-apis/)
- [X] 📄 阅读: [Sticky Sessions vs Session Replication](https://www.nginx.com/blog/nginx-plus-sticky-sessions/)
- [X] 📖 阅读 Redis 官方文档: [Session Store Pattern](https://redis.io/docs/manual/patterns/distributed-locks/)
- [X] ✍️ **任务**: 对比 Session Affinity、集中式 Session(Redis)、JWT Token 三种方案,列出场景选择决策树

**深入研究**:

- [X] 📚 阅读 Nginx 架构设计: [Inside NGINX: How We Designed for Performance &amp; Scale](https://www.nginx.com/blog/inside-nginx-how-we-designed-for-performance-scale/) -> 📝 [中文文档](./notes/week1/module1/Inside-NGINX-Performance-Architecture.md)
- [X] 📚 阅读 HAProxy 文档: [HAProxy Configuration Manual](http://cbonte.github.io/haproxy-dconv/2.6/configuration.html) - 关注 balance 算法部分
- [X] 🔬 **实验**: 使用go写一个简单的事件循环 -> 💻 [代码目录](./projects/week1/module1/nginx-event-loop/)

---

#### 模块 2: 负载均衡算法与实现 (建议 3-4 天)

**核心学习内容**:

**1. Round Robin 与加权轮询**

- [X] 📄 阅读: [Load Balancing Algorithms](https://kemptechnologies.com/load-balancer/load-balancing-algorithms-techniques/) - Kemp 官方文档 -> 📝 [笔记](notes/week1/module2/load_blance_note.md)
- [X] 📄 阅读: [Weighted Round-Robin Scheduling](https://kb.linuxvirtualserver.org/wiki/Weighted_Round-Robin_Scheduling) - LVS Wiki
- [X] 💻 **代码阅读**: Nginx 源码中的加权轮询实现 [ngx_http_upstream_round_robin.c](https://github.com/nginx/nginx/blob/master/src/http/ngx_http_upstream_round_robin.c)
- [X] ✍️ **任务**: 手写 Round Robin 算法(Go),处理权重为 [5, 1, 1] 的情况,验证分配比例

**2. 最少连接与最快响应**

- [X] 📄 阅读: [Least Connections Load Balancing](https://www.nginx.com/blog/choosing-nginx-plus-load-balancing-techniques/)
- [X] 📄 阅读: [Dynamic Load Balancing Algorithms](https://www.haproxy.com/blog/haproxy-load-balancing-algorithms/)
- [X] ✍️ **任务**: 对比静态算法(RR)与动态算法(Least Conn)的适用场景,绘制决策树

**3. 一致性哈希算法**

- [ ] 📖 阅读《DDIA》Chapter 6: Partitioning (第6章 - 分区部分)
- [ ] 📄 阅读经典论文: [Consistent Hashing and Random Trees](https://www.akamai.com/us/en/multimedia/documents/technical-publication/consistent-hashing-and-random-trees-distributed-caching-protocols-for-relieving-hot-spots-on-the-world-wide-web-technical-publication.pdf) (Karger et al.)
- [ ] 📺 观看视频: [Consistent Hashing Explained](https://www.youtube.com/watch?v=zaRkONvyGr8) (12分钟)
- [ ] 📄 阅读: [Consistent Hashing in Practice](https://www.toptal.com/big-data/consistent-hashing)
- [ ] ✍️ **任务**:
  - 手写一致性哈希实现,计算 100 个节点,每个节点 150 个虚拟节点的哈希环
  - 模拟节点上下线,计算数据迁移比例(理论值应接近 1/N)
  - 绘制哈希环可视化图

**4. 健康检查机制**

- [ ] 📄 阅读: [Health Checks in Load Balancing](https://docs.nginx.com/nginx/admin-guide/load-balancer/http-health-check/)
- [ ] 📄 阅读 HAProxy 文档: [Health Check](https://www.haproxy.com/documentation/hapee/latest/load-balancing/health-checking/active-health-checks/)
- [ ] 📄 阅读: [Passive vs Active Health Checks](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/health_checking)
- [ ] ✍️ **任务**: 设计主动健康检查(HTTP /health)和被动健康检查(错误率统计)的组合策略

**5. 故障检测与自动转移**

- [ ] 📄 阅读: [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html) - Martin Fowler
- [ ] 📄 阅读: [Failure Detection in Distributed Systems](https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/tr-2007-153.pdf)
- [ ] ✍️ **任务**: 设计故障检测算法,定义"不健康"的判定标准(连续失败次数、成功率阈值)

**实战项目**: Go 实现生产级负载均衡器

**开发步骤**:

- [ ] **Step 1**: 创建项目结构
  ```bash
  mkdir -p projects/week1/load-balancer/{cmd,internal/{lb,health,config},pkg/algorithms}
  ```
- [ ] **Step 2**: 实现负载均衡算法
  - [ ] RoundRobin 算法(支持权重)
  - [ ] LeastConnection 算法
  - [ ] ConsistentHash 算法(使用 MurmurHash3)
  - [ ] 算法工厂模式,支持运行时切换
- [ ] **Step 3**: 实现健康检查
  - [ ] 主动检查:定时 HTTP 健康探测(可配置间隔、超时、路径)
  - [ ] 被动检查:记录后端错误率,超过阈值标记为不健康
  - [ ] 状态管理:Healthy、Unhealthy、Draining 三种状态
- [ ] **Step 4**: 实现熔断保护
  - [ ] 错误率统计(滑动窗口)
  - [ ] 熔断状态机(Closed → Open → Half-Open)
  - [ ] 快速失败,避免雪崩
- [ ] **Step 5**: 性能优化
  - [ ] 使用 sync.Pool 复用对象
  - [ ] 使用 sync/atomic 减少锁竞争
  - [ ] HTTP/2 连接复用
- [ ] **Step 6**: 监控与可视化
  - [ ] 暴露 Prometheus 指标(请求数、延迟、后端健康状态)
  - [ ] 实现 /metrics 端点
- [ ] **Step 7**: 测试
  - [ ] 单元测试:各个算法的正确性
  - [ ] 集成测试:启动 mock 后端,验证负载均衡效果
  - [ ] 性能测试:使用 wrk 或 hey 压测,测量 QPS、P99 延迟

**参考代码**:

- [traefik/traefik](https://github.com/traefik/traefik) - 生产级负载均衡器
- [hashicorp/consul](https://github.com/hashicorp/consul/tree/main/agent/consul) - 健康检查实现

**项目路径**: `projects/week1/load-balancer/`

**代码质量检查**:

- [ ] 所有公共函数有文档注释(符合 Go doc 规范)
- [ ] 测试覆盖率 > 80% (运行 `go test -cover ./...`)
- [ ] 通过性能基准测试 (QPS > 10K, P99 < 10ms)
- [ ] 代码通过 golangci-lint 检查 (`golangci-lint run`)
- [ ] README 包含架构图、使用示例、性能数据

**笔记路径**: `notes/week1/module2-load-balancing.md`

**学习时长**: _____ 小时
**性能指标**: QPS _____, P99延迟 _____ ms
**踩过的坑**: _____________________

---

#### 模块 3: 性能评估与容量规划 (建议 2 天)

**核心学习内容**:

**1. QPS vs TPS**

- [ ] 📄 阅读: [Understanding QPS, TPS, and Throughput](https://aws.amazon.com/builders-library/using-load-testing-to-ensure-that-your-service-can-scale/)
- [ ] 📄 阅读: [Performance Metrics Explained](https://www.brendangregg.com/usemethod.html) - Brendan Gregg
- [ ] ✍️ **任务**:
  - 计算一个电商系统的 QPS:假设 100 万 DAU,每个用户平均每天 20 个请求,峰值流量是平均值的 3 倍
  - 公式:峰值 QPS = (DAU × 请求数 / 86400) × 峰值因子

**2. Latency 分析(P50、P95、P99)**

- [ ] 📖 阅读《Site Reliability Engineering》Chapter 4: Service Level Objectives
- [ ] 📄 阅读: [Percentile Latency - What it means](https://www.elastic.co/blog/averages-can-dangerous-use-percentile)
- [ ] 📺 观看视频: [Understanding Latency Percentiles](https://www.youtube.com/watch?v=lJ8ydIuPFeU) (15分钟)
- [ ] ✍️ **任务**:
  - 给定延迟数据:[1ms, 2ms, 3ms, ..., 100ms],手工计算 P50、P95、P99
  - 解释为什么 P99 比平均值更重要

**3. Throughput 与带宽估算**

- [ ] 📄 阅读: [Back-of-the-envelope Calculations](https://github.com/donnemartin/system-design-primer#back-of-the-envelope-calculations)
- [ ] 📄 阅读: [Numbers Every Programmer Should Know](https://gist.github.com/jboner/2841832)
- [ ] ✍️ **任务**:
  - 背诵常用数字:L1 cache (0.5ns), 内存访问(100ns), SSD 读(16μs), 网络往返(0.5ms)
  - 计算:1 Gbps 带宽每秒能传输多少 1MB 的图片?

**4. Little's Law**

- [ ] 📖 阅读《DDIA》Appendix: Little's Law
- [ ] 📄 阅读: [Little&#39;s Law in Practice](https://brooker.co.za/blog/2018/06/20/littles-law.html)
- [ ] 📄 阅读: [Applying Little&#39;s Law to System Design](https://www.speedshop.co/2015/10/05/rack-miniprofile.html)
- [ ] ✍️ **任务**:
  - 公式:并发用户数 = QPS × 平均响应时间
  - 练习:如果 QPS=1000,平均响应时间=100ms,系统需要支持多少并发连接?

**5. 容量规划方法论**

- [ ] 📖 阅读《System Design Interview》Chapter 1: Scale From Zero to Millions of Users
- [ ] 📄 阅读: [Capacity Planning Best Practices](https://aws.amazon.com/builders-library/reliability-and-constant-work/)
- [ ] 📄 阅读: [Instagram&#39;s Infrastructure](https://instagram-engineering.com/what-powers-instagram-hundreds-of-instances-dozens-of-technologies-adf2e22da2ad)
- [ ] ✍️ **任务**: 学习自底向上估算法(存储 → 带宽 → 服务器数)

**实践项目**: 为真实系统建立完整容量模型

**选择系统**: Twitter / Instagram / TikTok / 微博(任选一个)

**Step 1: 需求分析**

- [ ] 确定功能范围:发推文、关注、点赞、评论、搜索
- [ ] 估算用户规模:
  - DAU(日活):假设 2 亿
  - MAU(月活):假设 5 亿
  - 注册用户:假设 10 亿

**Step 2: 流量估算**

- [ ] 估算 QPS
  - 读操作:用户刷 timeline,平均每人每天 100 次请求
  - 写操作:平均每人每天发 2 条推文
  - 计算平均 QPS 和峰值 QPS(3倍峰值因子)
- [ ] 示例计算:
  ```
  读 QPS = (2亿 × 100) / 86400 ≈ 231,000 QPS
  峰值读 QPS ≈ 693,000 QPS
  写 QPS = (2亿 × 2) / 86400 ≈ 4,630 QPS
  ```

**Step 3: 存储容量估算**

- [ ] 推文存储
  - 单条推文:文本(140字 × 2bytes) + 元数据(用户ID、时间戳) ≈ 500 bytes
  - 每天新增推文:2亿用户 × 2条 = 4亿条
  - 每天存储:4亿 × 500 bytes ≈ 200 GB/天
  - 5年存储:200 GB × 365 × 5 ≈ 365 TB
- [ ] 图片/视频存储
  - 假设 20% 的推文包含图片(平均 200KB)
  - 每天:4亿 × 20% × 200KB ≈ 16 TB/天
  - 5年:16 TB × 365 × 5 ≈ 29 PB
- [ ] 总存储需求:约 30 PB(考虑副本,实际需要 90 PB)

**Step 4: 带宽估算**

- [ ] 入带宽(写)
  - 推文:200 GB / 86400s ≈ 2.3 MB/s
  - 图片:16 TB / 86400s ≈ 185 MB/s
  - 总写带宽:约 200 MB/s ≈ 1.6 Gbps
- [ ] 出带宽(读)
  - 假设读写比 100:1
  - 总读带宽:1.6 Gbps × 100 ≈ 160 Gbps

**Step 5: 服务器数量估算**

- [ ] 应用服务器
  - 假设单台服务器处理 1000 QPS
  - 需要服务器:693,000 / 1000 ≈ 700 台(考虑冗余,实际 1000 台)
- [ ] 数据库服务器
  - 假设单台 MySQL 处理 5000 QPS(读)
  - 需要从库:693,000 / 5000 ≈ 140 台读从库
  - 主库:按写 QPS,约 10 台(分片)
- [ ] 缓存服务器(Redis)
  - 假设缓存热数据 20%,需要约 6 PB × 20% = 1.2 PB 内存
  - 单台 Redis 256 GB 内存,需要约 5000 台

**Step 6: 成本估算**

- [ ] 服务器成本(AWS EC2 按需价格)
- [ ] 存储成本(S3 对象存储)
- [ ] 带宽成本
- [ ] 总 TCO(Total Cost of Ownership)

**参考资料**:

- [Twitter&#39;s Infrastructure](https://blog.twitter.com/engineering/en_us/topics/infrastructure)
- [Instagram&#39;s Scale](https://instagram-engineering.com/)

**笔记路径**: `notes/week1/module3-capacity-planning.md`

**输出成果**:

- 完整的容量规划 Excel 表格
- 架构图(包含服务器数量、存储、带宽标注)
- 成本预算表

**学习时长**: _____ 小时
**案例系统**: _____________________
**容量结论**: _____________________

---

#### Week 1 总结

- [ ] 完成周总结: `notes/week1/weekly-summary.md`
- [ ] 复习本周所有模块笔记
- [ ] 负载均衡器项目代码审查与优化
- [ ] 整理本周学习的架构图和流程图
- [ ] 规划下周学习重点

**本周总时长**: _____ 小时
**完成模块**: ___/3
**项目质量**: ⭐⭐⭐⭐⭐ (自评)
**最大突破**: _____________________
**需要加强**: _____________________

---

### Week 2: 存储系统原理

**本周目标**: 深入理解不同存储系统的设计权衡

#### 模块 1: 关系型数据库内核 (建议 2-3 天)

**核心学习内容**:

**1. ACID 特性的底层实现**

**1.1 原子性(Atomicity): Undo Log 机制**

- [ ] 📖 阅读《MySQL技术内幕:InnoDB存储引擎》第6章 - 锁
- [ ] 📖 阅读《DDIA》Chapter 7: Transactions
- [ ] 📄 阅读: [MySQL InnoDB Undo Log详解](https://dev.mysql.com/doc/refman/8.0/en/innodb-undo-logs.html)
- [ ] 💻 **源码阅读**: PostgreSQL Undo Log 实现 [src/backend/access/transam/xlog.c](https://github.com/postgres/postgres/blob/master/src/backend/access/transam/xlog.c)
- [ ] ✍️ **任务**: 画出事务回滚时 Undo Log 的工作流程图

**1.2 一致性(Consistency): 约束检查**

- [ ] 📖 阅读数据库原理教材:完整性约束部分
- [ ] ✍️ **任务**: 列举 5 种数据库约束(主键、外键、唯一、检查、非空),说明如何保证一致性

**1.3 隔离性(Isolation): MVCC 多版本并发控制**

- [ ] 📖 阅读《DDIA》Chapter 7: Weak Isolation Levels
- [ ] 📄 阅读: [PostgreSQL MVCC 详解](https://www.postgresql.org/docs/current/mvcc.html)
- [ ] 📄 阅读: [MySQL InnoDB MVCC 实现原理](https://dev.mysql.com/doc/refman/8.0/en/innodb-multi-versioning.html)
- [ ] 📺 观看视频: [How MVCC Works](https://www.youtube.com/watch?v=jlcKKQbXWNI) (20分钟)
- [ ] 💻 **源码阅读**: InnoDB MVCC 实现
  - 文件: `storage/innobase/read/read0read.cc` (ReadView 实现)
  - 重点理解: trx_id、roll_pointer、read view 机制
- [ ] ✍️ **任务**:
  - 画出 MVCC 的版本链示意图
  - 解释为什么 MVCC 能避免读写冲突
  - 对比 MVCC 与锁机制的性能差异

**1.4 持久性(Durability): WAL(Write-Ahead Logging)**

- [ ] 📖 阅读《DDIA》Chapter 3: Storage and Retrieval - Log-Structured Storage
- [ ] 📄 阅读经典论文: [ARIES: A Transaction Recovery Method](https://cs.stanford.edu/people/chrismre/cs345/rl/aries.pdf) (IBM Research)
- [ ] 📄 阅读: [PostgreSQL WAL Internals](https://www.postgresql.org/docs/current/wal-intro.html)
- [ ] 💻 **源码阅读**: PostgreSQL WAL 实现
  - 文件: `src/backend/access/transam/xlog.c`
  - 关注函数: `XLogInsert()`, `XLogFlush()`
- [ ] ✍️ **任务**:
  - 画出 WAL 写入流程:内存 → WAL Buffer → WAL File → 数据页
  - 解释为什么 WAL 能保证持久性(即使断电)
  - 理解 checkpo int机制如何减少恢复时间

**2. 索引结构深度剖析**

**2.1 B-Tree vs B+Tree**

- [ ] 📖 阅读《算法导论》第18章: B树
- [ ] 📄 阅读: [B-Tree vs B+Tree in Databases](https://www.geeksforgeeks.org/difference-between-b-tree-and-b-tree/)
- [ ] 📄 阅读: [Why MySQL Uses B+Tree](https://medium.com/@mena.meseha/what-is-the-difference-between-b-tree-and-b-tree-in-database-5662cda8f8b1)
- [ ] 💻 **源码阅读**: InnoDB B+Tree 实现
  - 文件: `storage/innobase/include/page0page.h`
  - 重点: B+Tree 节点结构、页分裂逻辑
- [ ] ✍️ **任务**:
  - 对比表格: B-Tree vs B+Tree(叶子节点、内部节点、范围查询性能)
  - 手绘一棵 B+Tree(阶数=3,插入数据 [1,3,5,7,9,11,13,15])
  - 计算: 高度为 3 的 B+Tree,阶数为 200,能存储多少数据?

**2.2 Hash 索引**

- [ ] 📄 阅读: [Hash Index in MySQL](https://dev.mysql.com/doc/refman/8.0/en/index-btree-hash.html)
- [ ] ✍️ **任务**: 列举 Hash 索引的优缺点,说明为什么不支持范围查询

**2.3 聚簇索引 vs 非聚簇索引**

- [ ] 📖 阅读《高性能MySQL》第5章: 索引
- [ ] 📄 阅读: [Clustered vs Non-Clustered Index](https://www.sqlshack.com/what-is-the-difference-between-clustered-and-non-clustered-indexes-in-sql-server/)
- [ ] ✍️ **任务**:
  - 画出 InnoDB 聚簇索引结构(主键索引、二级索引)
  - 解释为什么 InnoDB 二级索引需要回表查询

**3. 查询优化器**

- [ ] 📖 阅读《数据库系统概念》第13章: 查询优化
- [ ] 📄 阅读: [PostgreSQL Query Planner](https://www.postgresql.org/docs/current/planner-optimizer.html)
- [ ] 📄 阅读: [MySQL EXPLAIN Output](https://dev.mysql.com/doc/refman/8.0/en/explain-output.html)
- [ ] ✍️ **任务**:
  - 使用 `EXPLAIN` 分析一个复杂查询
  - 理解查询计划中的 type、possible_keys、key、rows 字段

**4. 事务隔离级别**

- [ ] 📖 阅读《DDIA》Chapter 7: Isolation Levels
- [ ] 📄 阅读: [Transaction Isolation Levels](https://www.postgresql.org/docs/current/transaction-iso.html)
- [ ] 📺 观看视频: [Database Isolation Levels](https://www.youtube.com/watch?v=4EajrPgJAk0) (15分钟)
- [ ] ✍️ **任务**:
  - 对比表格: Read Uncommitted、Read Committed、Repeatable Read、Serializable
  - 每个级别可能出现的问题: 脏读、不可重复读、幻读
  - 在 MySQL 中实验各个隔离级别(使用两个并发事务)

**深入研究**:

**源码阅读任务**:

- [ ] 💻 阅读 PostgreSQL MVCC 核心代码(2000行)
  - `src/backend/access/heap/heapam.c` - Heap Access Method
  - `src/backend/utils/time/tqual.c` - Visibility Rules
  - 理解: 可见性判断逻辑(HeapTupleSatisfiesMVCC)
- [ ] 💻 阅读 InnoDB WAL 实现(1000行)
  - `storage/innobase/log/log0log.cc`
  - 理解: 日志写入、刷盘、checkpoint 机制

**实验任务**:

- [ ] 🔬 实验 MVCC 避免读写冲突
  - 开启两个 MySQL 客户端
  - 事务1: BEGIN; SELECT * FROM users WHERE id=1;
  - 事务2: BEGIN; UPDATE users SET name='new' WHERE id=1; COMMIT;
  - 事务1: SELECT * FROM users WHERE id=1; (观察结果,理解快照读)
- [ ] 🔬 实验 WAL 持久性
  - 写入数据但不刷新到数据文件
  - 模拟崩溃恢复,观察 WAL 如何恢复数据

**实战项目**: Python 实现简化版 B+Tree 索引

**项目要求**:

- [ ] **Step 1**: 实现 B+Tree 节点类
  ```python
  class BPlusTreeNode:
      def __init__(self, order):
          self.order = order  # 阶数
          self.keys = []
          self.children = []
          self.is_leaf = True
  ```
- [ ] **Step 2**: 实现插入操作(包含节点分裂)
- [ ] **Step 3**: 实现查找操作(单点查询、范围查询)
- [ ] **Step 4**: 实现删除操作(包含节点合并)
- [ ] **Step 5**: 可视化 B+Tree 结构
  - 使用 graphviz 或 matplotlib
  - 显示每个节点的 keys 和指针
- [ ] **Step 6**: 性能对比实验
  - 插入 100万 条数据,对比 B+Tree vs 线性查找
  - 测量查询时间、内存占用
- [ ] **Step 7**: 单元测试
  - 测试插入、删除、查询的正确性
  - 边界条件测试(空树、单节点、满节点)

**参考代码**:

- [SQLite B-Tree Implementation](https://github.com/sqlite/sqlite/blob/master/src/btree.c)
- [Python B+Tree Example](https://github.com/NicolasKeita/BPlusTree_Python)

**项目路径**: `projects/week2/btree-index/`

**笔记路径**: `notes/week2/module1-rdbms-internals.md`

**学习时长**: _____ 小时
**源码阅读**: PostgreSQL _____ 行, InnoDB _____ 行
**关键理解**: _____________________

---

#### 模块 2: NoSQL 存储系统 (建议 3-4 天)

**核心学习内容**:

- [ ] LSM-Tree 存储引擎深度剖析
  - [ ] MemTable、SSTable 结构设计
  - [ ] Compaction 策略(Size-Tiered、Leveled)
  - [ ] 布隆过滤器优化读性能
  - [ ] LSM-Tree vs B-Tree: 写优化与读优化权衡
- [ ] Redis 内部实现
  - [ ] SDS(Simple Dynamic String)设计
  - [ ] 跳表实现 Sorted Set
  - [ ] 压缩列表节省内存
  - [ ] 持久化机制: RDB vs AOF
- [ ] MongoDB WiredTiger 存储引擎
- [ ] Cassandra 架构与一致性哈希

**深入研究**:

- [ ] 阅读 RocksDB 源码(Compaction 部分)
- [ ] 阅读 Redis 跳表实现源码
- [ ] 理解为什么 LSM-Tree 适合写多读少场景

**实战项目**: Go 实现 LSM-Tree 存储引擎原型

- [ ] 实现 MemTable(使用跳表或红黑树)
- [ ] 实现 WAL 持久化
- [ ] 实现 SSTable 生成与读取
- [ ] 实现简单的 Compaction(Major Compaction)
- [ ] 实现布隆过滤器优化
- [ ] 性能测试:写吞吐量、读延迟

**项目路径**: `projects/week2/lsm-tree/`

**笔记路径**: `notes/week2/module2-nosql-systems.md`

**学习时长**: _____ 小时
**写吞吐**: _____ ops/s
**读延迟**: P99 _____ ms
**核心优化**: _____________________

---

#### 模块 3: 分布式存储与一致性 (建议 2 天)

**核心学习内容**:

- [ ] 数据分片(Sharding)策略
  - [ ] 范围分片 vs 哈希分片
  - [ ] 分片键选择原则
  - [ ] 数据迁移与再平衡
- [ ] 数据复制(Replication)
  - [ ] 主从复制原理
  - [ ] 多主复制的冲突解决
  - [ ] 无主复制与 Quorum 机制
- [ ] CAP 定理深入理解
  - [ ] 理论证明与现实意义
  - [ ] CP vs AP 的选择场景
- [ ] 一致性模型谱系
  - [ ] 强一致性、顺序一致性、因果一致性
  - [ ] 最终一致性与冲突解决
  - [ ] 向量时钟原理

**实践**: 实现支持最终一致性的分布式 KV 存储

- [ ] 基于一致性哈希的数据分片
- [ ] 实现 W + R > N 的 Quorum 读写
- [ ] 实现向量时钟检测冲突
- [ ] 简单的冲突解决策略(Last Write Wins)

**项目路径**: `projects/week2/distributed-kv/`

**笔记路径**: `notes/week2/module3-distributed-storage.md`

**学习时长**: _____ 小时
**CAP选择**: _____(CP/AP)
**核心权衡**: _____________________

---

#### Week 2 总结

- [ ] 完成周总结: `notes/week2/weekly-summary.md`
- [ ] 对比 B-Tree、LSM-Tree、Hash Index 的适用场景
- [ ] 绘制存储系统技术选型决策树
- [ ] 复习 ACID、CAP、BASE 理论
- [ ] 整理源码阅读笔记

**本周总时长**: _____ 小时
**完成模块**: ___/3
**源码阅读**: _____ 行
**实现项目**: B-Tree[x/~/_], LSM-Tree[x/~/_], KV[x/~/_]
**理论突破**: _____________________

---

### Week 3: 网络通信与 API 设计

**本周目标**: 掌握分布式系统的通信机制

#### 模块 1: HTTP 协议深入 (建议 2 天)

**核心学习内容**:

- [ ] HTTP/1.1 vs HTTP/2 vs HTTP/3
  - [ ] 多路复用原理
  - [ ] 头部压缩(HPACK、QPACK)
  - [ ] QUIC 协议与 UDP 传输
- [ ] RESTful API 设计最佳实践
  - [ ] 资源建模与 URL 设计
  - [ ] HTTP 方法语义(GET、POST、PUT、PATCH、DELETE)
  - [ ] 状态码的正确使用
  - [ ] 幂等性设计
- [ ] API 版本管理策略(URL、Header、Media Type)

**实战项目**: Go 实现符合 REST 规范的 API 服务

- [ ] 实现完整的 CRUD 操作
- [ ] 设计 RESTful URL 结构
- [ ] 实现请求验证中间件
- [ ] 实现日志中间件
- [ ] 实现认证中间件(基础版)
- [ ] 支持分页、过滤、排序
- [ ] API 文档(OpenAPI/Swagger)

**项目路径**: `projects/week3/restful-api/`

**笔记路径**: `notes/week3/module1-http-rest.md`

**学习时长**: _____ 小时
**API 端点数**: _____
**设计亮点**: _____________________

---

#### 模块 2: RPC 框架原理 (建议 2-3 天)

**核心学习内容**:

- [ ] RPC 工作原理深入
  - [ ] 序列化与反序列化(JSON、Protobuf、Thrift)
  - [ ] 网络传输层设计
  - [ ] 服务发现与注册
- [ ] gRPC 深度剖析
  - [ ] Protocol Buffers 语法与最佳实践
  - [ ] 四种通信模式(Unary、Server Streaming、Client Streaming、Bidirectional)
  - [ ] 拦截器与中间件
  - [ ] 负载均衡与重试

**深入研究**:

- [ ] 阅读 gRPC-Go 源码(核心流程)
- [ ] 理解 Protobuf 编码原理
- [ ] 对比 JSON vs Protobuf 性能

**实战项目**: 实现简化版 RPC 框架

- [ ] 自定义协议设计(魔数、版本、消息ID、消息体)
- [ ] 实现编解码器(支持 JSON 和自定义格式)
- [ ] 实现服务注册与发现(基于内存或 etcd)
- [ ] 实现客户端代理(动态生成)
- [ ] 实现超时和重试机制
- [ ] 性能对比:自定义 RPC vs HTTP/REST

**项目路径**: `projects/week3/simple-rpc/`

**笔记路径**: `notes/week3/module2-rpc-framework.md`

**学习时长**: _____ 小时
**协议设计**: _____________________
**性能对比**: 自定义RPC _____ vs HTTP _____

---

#### 模块 3: 消息队列基础 (建议 3 天)

**核心学习内容**:

- [ ] 消息队列核心概念
  - [ ] 点对点 vs 发布订阅
  - [ ] 消息持久化机制
  - [ ] 消息确认与重试
- [ ] Kafka 架构深度剖析
  - [ ] Topic、Partition、Segment 设计
  - [ ] Producer、Broker、Consumer 工作流程
  - [ ] 副本机制与 ISR(In-Sync Replicas)
  - [ ] 高性能设计:顺序写、零拷贝、批量压缩
- [ ] RabbitMQ vs Kafka 对比

**深入研究**:

- [ ] 阅读 Kafka 官方文档(设计部分)
- [ ] 理解 Kafka 如何实现高吞吐量
- [ ] 研究 Kafka 的 exactly-once 语义

**实战项目**: Go + Kafka 实现可靠消息系统

- [ ] 实现消息生产者(带确认机制)
- [ ] 实现消费者组(自动负载均衡)
- [ ] 实现消息去重(使用幂等性key)
- [ ] 实现消息追踪(记录生产、消费时间)
- [ ] 处理消费失败(重试 + 死信队列)
- [ ] 性能测试:吞吐量、延迟

**项目路径**: `projects/week3/kafka-messaging/`

**笔记路径**: `notes/week3/module3-message-queue.md`

**学习时长**: _____ 小时
**吞吐量**: _____ msg/s
**端到端延迟**: P99 _____ ms
**可靠性保证**: _____________________

---

#### Week 3 总结 & Phase 1 总结

- [ ] 完成 Week 3 总结: `notes/week3/weekly-summary.md`
- [ ] 完成 Phase 1 总结: `notes/phase1-summary.md`
- [ ] 对比 HTTP、RPC、消息队列的适用场景
- [ ] 复习前三周所有核心概念
- [ ] 整理 Phase 1 所有架构图
- [ ] 准备进入 Phase 2

**本周总时长**: _____ 小时
**Phase 1 总时长**: _____ 小时
**完成模块**: ___/9
**完成项目**: _____ 个
**最大收获**: _____________________
**Phase 2 目标**: _____________________

---

## 第二阶段: 高级架构模式 (Week 4-7)

### Week 4: 缓存系统设计

**本周目标**: 深入理解缓存原理与最佳实践

#### 模块 1: 缓存理论与算法 (建议 2 天)

**核心学习内容**:

- [ ] 缓存淘汰算法深度剖析
  - [ ] LRU 实现:双向链表 + HashMap,O(1)复杂度证明
  - [ ] LFU 实现与优化
  - [ ] ARC(Adaptive Replacement Cache)
  - [ ] 时间与空间复杂度分析
- [ ] 缓存问题与解决方案
  - [ ] 缓存穿透:布隆过滤器、空值缓存
  - [ ] 缓存击穿:互斥锁、永不过期
  - [ ] 缓存雪崩:随机过期、熔断限流
  - [ ] 缓存污染问题

**实战项目**: Go 实现高性能 LRU/LFU Cache

- [ ] LRU Cache: 双向链表 + HashMap
  - [ ] O(1) Get/Put 操作
  - [ ] 支持 TTL(过期时间)
  - [ ] 线程安全(sync.RWMutex)
  - [ ] 统计信息(命中率、淘汰次数)
- [ ] LFU Cache: 频率计数优化
- [ ] 性能基准测试与对比
- [ ] 并发压测(多 goroutine 读写)

**项目路径**: `projects/week4/lru-cache/`

**笔记路径**: `notes/week4/module1-cache-algorithms.md`

**学习时长**: _____ 小时
**命中率**: _____%, 淘汰次数: _____
**并发QPS**: _____ ops/s

---

#### 模块 2: 分布式缓存架构 (建议 3 天)

**核心学习内容**:

- [ ] Redis 集群架构
  - [ ] 主从复制原理
  - [ ] Sentinel 高可用方案
  - [ ] Cluster 分片方案
- [ ] 一致性哈希深度实现
  - [ ] 虚拟节点数量优化
  - [ ] 节点增删的数据迁移
- [ ] 缓存更新策略
  - [ ] Cache-Aside Pattern
  - [ ] Read-Through / Write-Through
  - [ ] Write-Behind Caching
- [ ] 多级缓存架构设计
  - [ ] 本地缓存 + 分布式缓存
  - [ ] 缓存预热策略

**实战项目**: Python 实现一致性哈希环

- [ ] 实现哈希环数据结构
- [ ] 实现虚拟节点(每个物理节点映射多个虚拟节点)
- [ ] 支持节点动态增删
- [ ] 计算数据分布均衡度
- [ ] 可视化哈希环与数据分布

**项目路径**: `projects/week4/consistent-hashing/`

**笔记路径**: `notes/week4/module2-distributed-cache.md`

**学习时长**: _____ 小时
**虚拟节点数**: _____
**数据均衡度**: 标准差 _____

---

#### 模块 3: CDN 与边缘计算 (建议 2 天)

**核心学习内容**:

- [ ] CDN 工作原理
  - [ ] DNS 解析与调度
  - [ ] 回源与缓存策略
  - [ ] 缓存失效与更新(Purge、Refresh)
- [ ] 静态资源优化
  - [ ] 图片优化(格式选择、压缩、WebP)
  - [ ] 视频流媒体分发(HLS、DASH)
- [ ] 边缘计算概念

**案例分析**: 设计全球化图片分发系统

- [ ] 需求分析(用户分布、访问模式)
- [ ] CDN 节点部署策略
- [ ] 缓存策略设计(TTL、淘汰策略)
- [ ] 回源优化(合并回源、预加载)
- [ ] 成本估算

**笔记路径**: `notes/week4/module3-cdn-design.md`

**学习时长**: _____ 小时
**案例系统**: _____________________

---

#### Week 4 总结

- [ ] 完成周总结: `notes/week4/weekly-summary.md`
- [ ] 对比不同缓存淘汰算法的性能
- [ ] 总结缓存穿透/击穿/雪崩的解决方案
- [ ] 复习一致性哈希原理
- [ ] 缓存系统设计最佳实践总结

**本周总时长**: _____ 小时
**完成模块**: ___/3
**LRU Cache性能**: _____
**理论突破**: _____________________

---

### Week 5: 数据库进阶与分布式数据

**本周目标**: 掌握大规模数据管理技术

#### 模块 1: 数据库分片实战 (建议 2 天)

**核心学习内容**:

- [ ] 分片策略深入
  - [ ] 水平分片 vs 垂直分片
  - [ ] 分片键选择的考量因素
  - [ ] 跨分片事务处理
  - [ ] 全局唯一 ID 生成(Snowflake、UUID)
- [ ] 分片带来的挑战
  - [ ] 跨分片查询优化
  - [ ] 分片再平衡
  - [ ] 数据迁移策略

**实战项目**: Go 实现 Sharding 中间件

- [ ] SQL 解析(简单的 SELECT、INSERT)
- [ ] 分片路由(根据分片键路由到正确的分片)
- [ ] 跨分片查询结果聚合
- [ ] 支持多种分片策略(范围、哈希、取模)

**项目路径**: `projects/week5/sharding-middleware/`

**笔记路径**: `notes/week5/module1-database-sharding.md`

---

#### 模块 2: 读写分离与高可用 (建议 2 天)

**核心学习内容**:

- [ ] 主从复制深入
  - [ ] 异步复制 vs 同步复制 vs 半同步复制
  - [ ] 复制延迟问题与解决方案
  - [ ] 主从一致性保证
- [ ] 故障转移(Failover)
  - [ ] 主从切换流程
  - [ ] 脑裂问题与 Fencing
  - [ ] Raft/Paxos 在主备选举中的应用

**实战项目**: Python 实现数据库代理

- [ ] 读写分离路由(写走主库,读走从库)
- [ ] 支持主从切换
- [ ] 健康检查
- [ ] 连接池管理

**项目路径**: `projects/week5/rw-split-proxy/`

**笔记路径**: `notes/week5/module2-replication-ha.md`

---

#### 模块 3: 分布式事务 (建议 3 天)

**核心学习内容**:

- [ ] 分布式事务理论
  - [ ] 两阶段提交(2PC)原理与问题
  - [ ] 三阶段提交(3PC)改进
  - [ ] 最终一致性方案
- [ ] Saga 模式深入
  - [ ] 编排式 Saga vs 编舞式 Saga
  - [ ] 补偿事务设计
- [ ] TCC(Try-Confirm-Cancel)模式
- [ ] 本地消息表模式

**实战项目**: Go 实现 Saga 协调器

- [ ] 状态机管理(Saga 状态转换)
- [ ] 编排式 Saga 实现
- [ ] 补偿逻辑执行
- [ ] 异常处理与回滚
- [ ] 持久化 Saga 状态

**项目路径**: `projects/week5/saga-coordinator/`

**笔记路径**: `notes/week5/module3-distributed-transaction.md`

---

#### Week 5 总结

- [ ] 完成周总结: `notes/week5/weekly-summary.md`
- [ ] 对比不同分布式事务方案
- [ ] 总结数据库扩展的关键技术
- [ ] Instagram 数据库架构案例分析

**本周总时长**: _____ 小时

---

### Week 6: 事件驱动与异步架构

**本周目标**: 掌握事件驱动架构设计

#### 模块 1: 事件驱动架构 (建议 2 天)

**核心学习内容**:

- [ ] Event Sourcing 模式
  - [ ] 事件存储设计
  - [ ] 事件重放与状态重建
  - [ ] 快照机制
- [ ] CQRS(命令查询职责分离)
  - [ ] 读写模型分离
  - [ ] 最终一致性处理
  - [ ] 与 Event Sourcing 结合

**案例分析**: 电商订单系统的事件驱动设计

**笔记路径**: `notes/week6/module1-event-driven.md`

---

#### 模块 2: 消息队列深入应用 (建议 3 天)

**核心学习内容**:

- [ ] Kafka 高级特性
  - [ ] 事务消息
  - [ ] 精确一次语义(Exactly Once)
  - [ ] Consumer Offset 管理
  - [ ] Rebalance 机制
- [ ] 消息可靠性保证
  - [ ] 生产者确认机制
  - [ ] 消费者 ACK 机制
  - [ ] 死信队列设计

**实战项目**: 实现可靠的事件总线

- [ ] At-Least-Once 语义保证
- [ ] 消息去重
- [ ] 消息追踪

**项目路径**: `projects/week6/event-bus/`

**笔记路径**: `notes/week6/module2-message-queue-advanced.md`

---

#### 模块 3: 异步任务处理 (建议 2 天)

**核心学习内容**:

- [ ] 任务队列设计
  - [ ] Celery 架构分析
  - [ ] 延迟任务实现
  - [ ] 定时任务调度
- [ ] 任务可靠性
  - [ ] 重试策略(指数退避)
  - [ ] 幂等性设计
  - [ ] 任务超时处理

**实战项目**: Python + Redis 实现任务队列

**项目路径**: `projects/week6/task-queue/`

**笔记路径**: `notes/week6/module3-async-tasks.md`

---

#### Week 6 总结

- [ ] 完成周总结: `notes/week6/weekly-summary.md`
- [ ] 通知系统设计实战

**本周总时长**: _____ 小时

---

### Week 7: 微服务架构

**本周目标**: 掌握微服务设计原则

#### 模块 1: 微服务基础 (建议 2 天)

**核心学习内容**:

- [ ] 微服务架构演进
  - [ ] 单体应用的问题
  - [ ] 微服务的优势与挑战
  - [ ] 服务拆分原则(DDD、业务边界)
- [ ] 服务间通信
  - [ ] 同步通信(HTTP/gRPC)
  - [ ] 异步通信(消息队列)
  - [ ] 服务网格(Service Mesh)简介

**案例分析**: 单体应用拆分实践

**笔记路径**: `notes/week7/module1-microservices-intro.md`

---

#### 模块 2: 服务治理 (建议 3 天)

**核心学习内容**:

- [ ] 服务注册与发现
  - [ ] Consul、Etcd、Zookeeper 对比
  - [ ] 客户端发现 vs 服务端发现
  - [ ] 健康检查机制
- [ ] 配置管理
  - [ ] 集中式配置中心
  - [ ] 配置热更新
  - [ ] 配置版本管理
- [ ] 服务路由与负载均衡

**实战项目**: Go + Consul 实现服务治理

- [ ] 服务注册与发现
- [ ] 配置动态更新
- [ ] 健康检查

**项目路径**: `projects/week7/service-governance/`

**笔记路径**: `notes/week7/module2-service-governance.md`

---

#### 模块 3: 容错与弹性设计 (建议 2 天)

**核心学习内容**:

- [ ] 熔断器模式(Circuit Breaker)
  - [ ] 状态机设计(Closed、Open、Half-Open)
  - [ ] 错误率阈值计算
  - [ ] 恢复策略
- [ ] 限流(Rate Limiting)
  - [ ] 令牌桶算法
  - [ ] 漏桶算法
  - [ ] 滑动窗口算法
- [ ] 降级与超时

**实战项目**: 实现 Hystrix 风格的熔断器

**项目路径**: `projects/week7/circuit-breaker/`

**笔记路径**: `notes/week7/module3-resilience.md`

---

#### Week 7 总结 & Phase 2 总结

- [ ] 完成 Week 7 总结: `notes/week7/weekly-summary.md`
- [ ] 完成 Phase 2 总结: `notes/phase2-summary.md`
- [ ] 电商系统微服务架构设计实战

**本周总时长**: _____ 小时
**Phase 2 总时长**: _____ 小时

---

## 第三阶段: 分布式系统核心 (Week 8-10)

### Week 8: 分布式一致性算法

**本周目标**: 理解分布式一致性的理论与实现

#### 模块 1: 一致性算法理论 (建议 2 天)

**核心学习内容**:

- [ ] Paxos 算法
  - [ ] Basic Paxos 原理
  - [ ] Multi-Paxos 优化
  - [ ] 活锁问题与解决
- [ ] Raft 算法深入
  - [ ] Leader 选举机制
  - [ ] 日志复制流程
  - [ ] 安全性证明
  - [ ] 集群成员变更

**深入研究**:

- [ ] 阅读 Raft 论文
- [ ] 阅读 etcd Raft 实现源码

**笔记路径**: `notes/week8/module1-consensus-theory.md`

---

#### 模块 2: Raft 算法实现 (建议 3-5 天)

**前置准备**:

**必读论文**:

- [ ] 📄 阅读 Raft 论文: [In Search of an Understandable Consensus Algorithm](https://raft.github.io/raft.pdf) (Diego Ongaro, 2014)
  - 重点阅读 Section 5: The Raft consensus algorithm
  - Section 6: Cluster membership changes
  - Section 7: Log compaction
- [ ] 📺 观看 Raft 作者讲解: [Raft Lecture (Raft user study)](https://www.youtube.com/watch?v=YbZ3zDzDnrw) (1小时)
- [ ] 🎮 交互式学习: [Raft Visualization](https://raft.github.io/) - 动画演示 Leader 选举和日志复制

**源码阅读**:

- [ ] 💻 阅读 etcd Raft 实现(核心 3000 行)
  - [etcd-io/raft](https://github.com/etcd-io/raft)
  - 重点文件:
    - `raft.go` - Raft 状态机核心
    - `log.go` - 日志管理
    - `node.go` - 节点接口
  - 理解: 状态转换、消息传递机制

**核心项目**: Go 从零实现 Raft 算法

**项目架构**:

```
projects/week8/raft/
├── raft/
│   ├── raft.go          # Raft 核心状态机
│   ├── log.go           # 日志管理
│   ├── rpc.go           # RPC 消息定义
│   ├── persister.go     # 持久化
│   └── snapshot.go      # 快照
├── kvraft/              # 基于 Raft 的 KV 存储
├── test/                # 测试用例
└── README.md
```

**实现步骤**:

**Phase 1: Leader 选举 (Day 1)**

- [ ] **Step 1.1**: 定义 Raft 状态
  ```go
  type Raft struct {
      mu        sync.Mutex
      peers     []*labrpc.ClientEnd
      persister *Persister
      me        int

      // Persistent state
      currentTerm int
      votedFor    int
      log         []LogEntry

      // Volatile state
      commitIndex int
      lastApplied int
      state       NodeState  // Follower/Candidate/Leader

      // Leader state
      nextIndex   []int
      matchIndex  []int

      // Election
      electionTimer  *time.Timer
      heartbeatTimer *time.Timer
  }
  ```
- [ ] **Step 1.2**: 实现 RequestVote RPC
  - 处理投票请求
  - 检查 term、日志新旧程度
  - 投票规则:每个 term 只能投一票,且日志不比自己旧
- [ ] **Step 1.3**: 实现选举超时与发起选举
  - 随机选举超时(150-300ms)
  - 状态转换: Follower → Candidate
  - 向所有节点发送 RequestVote
  - 获得多数票后成为 Leader
- [ ] **Step 1.4**: 测试 Leader 选举
  - 启动 3 个节点,验证选出 1 个 Leader
  - Leader 宕机,验证重新选举
  - 网络分区,验证 split-brain 处理

**Phase 2: 日志复制 (Day 2)**

- [ ] **Step 2.1**: 定义日志条目
  ```go
  type LogEntry struct {
      Term    int
      Command interface{}
  }
  ```
- [ ] **Step 2.2**: 实现 AppendEntries RPC
  - 心跳机制(无日志条目)
  - 日志一致性检查(prevLogIndex, prevLogTerm)
  - 日志冲突解决(删除冲突日志)
  - 更新 commitIndex
- [ ] **Step 2.3**: Leader 日志复制逻辑
  - 客户端命令 → Leader 日志
  - 并行发送 AppendEntries 到 Followers
  - 收到多数派 ACK 后提交(commit)
  - 应用到状态机(apply)
- [ ] **Step 2.4**: Follower 日志复制逻辑
  - 接收 AppendEntries
  - 一致性检查,回退 nextIndex
  - 追加日志,更新 commitIndex
- [ ] **Step 2.5**: 测试日志复制
  - 正常情况:所有节点日志一致
  - Follower 宕机重启:追赶日志
  - Leader 宕机:新 Leader 覆盖未提交日志

**Phase 3: 持久化 (Day 3)**

- [ ] **Step 3.1**: 实现持久化接口
  ```go
  func (rf *Raft) persist() {
      w := new(bytes.Buffer)
      e := labgob.NewEncoder(w)
      e.Encode(rf.currentTerm)
      e.Encode(rf.votedFor)
      e.Encode(rf.log)
      data := w.Bytes()
      rf.persister.SaveRaftState(data)
  }
  ```
- [ ] **Step 3.2**: 关键点持久化
  - 每次更新 currentTerm、votedFor、log 后调用 persist()
- [ ] **Step 3.3**: 崩溃恢复
  - 启动时从持久化状态恢复
  - 重新初始化 volatile state
- [ ] **Step 3.4**: 测试持久化
  - 节点崩溃重启,验证状态恢复
  - 崩溃前未提交的日志丢失,已提交的保留

**Phase 4: 日志压缩(快照) (Day 3)**

- [ ] **Step 4.1**: 实现 Snapshot RPC
  - 发送快照给落后太多的 Follower
  - InstallSnapshot RPC
- [ ] **Step 4.2**: 创建快照
  - 应用层触发快照(日志过大时)
  - 保存状态机快照 + lastIncludedIndex/Term
  - 丢弃旧日志
- [ ] **Step 4.3**: 测试快照
  - 日志过长触发快照
  - Follower 通过快照恢复

**Phase 5: 高级优化 (Day 4-5)**

- [ ] **优化1: Batch 批量复制**
  - 一次 AppendEntries 包含多个日志条目
  - 减少 RPC 次数
- [ ] **优化2: Pipeline 流水线**
  - 不等待 ACK,继续发送下一批日志
  - 类似 TCP 滑动窗口
- [ ] **优化3: 并行日志应用**
  - 日志提交后,异步应用到状态机
  - 避免阻塞日志复制
- [ ] **优化4: PreVote**
  - 避免网络分区节点频繁发起选举扰乱集群
  - 发起选举前先 PreVote,确认能获得多数票

**测试与验证**:

**单元测试**:

- [ ] TestInitialElection: 初始选举
- [ ] TestReElection: Leader 宕机重新选举
- [ ] TestBasicAgree: 基本日志复制
- [ ] TestFailAgree: Follower 宕机后的日志复制
- [ ] TestFailNoAgree: 多数派宕机无法提交
- [ ] TestConcurrentStarts: 并发客户端请求
- [ ] TestRejoin: 节点宕机重新加入
- [ ] TestBackup: 日志冲突解决
- [ ] TestPersist: 持久化与崩溃恢复
- [ ] TestSnapshot: 快照机制

**混沌测试**(Chaos Testing):

- [ ] 随机杀死节点
- [ ] 随机网络分区
- [ ] 随机延迟/丢包
- [ ] 验证线性一致性(Linearizability)

**性能测试**:

- [ ] 吞吐量: 每秒提交多少日志条目
- [ ] 延迟: 从客户端请求到提交的时间(P50、P99)
- [ ] 选举时间: Leader 宕机到新 Leader 选出的时间

**参考实现**:

- [MIT 6.824 Lab 2: Raft](https://pdos.csail.mit.edu/6.824/labs/lab-raft.html) - 经典实验
- [etcd/raft](https://github.com/etcd-io/raft) - 生产级实现
- [hashicorp/raft](https://github.com/hashicorp/raft) - 另一个生产级实现

**项目路径**: `projects/week8/raft/`

**代码质量要求**:

- [ ] 所有 RPC 有详细注释
- [ ] 关键状态转换有日志输出
- [ ] 测试覆盖率 > 90%
- [ ] 通过 MIT 6.824 的所有测试用例
- [ ] 通过 race detector (`go test -race`)

**笔记路径**: `notes/week8/module2-raft-implementation.md`

**学习时长**: _____ 小时
**代码行数**: _____ 行
**通过测试**: ___/10
**最大挑战**: _____________________

---

#### 模块 3: 分布式锁与协调 (建议 2 天)

**核心学习内容**:

- [ ] 分布式锁实现方案
  - [ ] 基于 Redis 的分布式锁(SETNX、Lua)
  - [ ] 基于 Zookeeper 的分布式锁
  - [ ] Redlock 算法分析与争议
- [ ] 分布式协调服务
  - [ ] Zookeeper 原理
  - [ ] etcd 架构
  - [ ] Leader 选举、配置管理应用

**实战**: 实现生产级分布式锁

**项目路径**: `projects/week8/distributed-lock/`

**笔记路径**: `notes/week8/module3-distributed-lock.md`

---

#### Week 8 总结

- [ ] 完成周总结: `notes/week8/weekly-summary.md`
- [ ] Raft 实现代码审查与优化

**本周总时长**: _____ 小时

---

### Week 9: 可观测性系统

**本周目标**: 构建完整的监控、日志、追踪体系

#### 模块 1: 日志系统 (建议 2 天)

**核心学习内容**:

- [ ] 集中式日志架构
  - [ ] ELK Stack 架构
  - [ ] Filebeat、Logstash 工作原理
  - [ ] Elasticsearch 存储与查询
- [ ] 日志采集与聚合
  - [ ] 日志格式标准化
  - [ ] 结构化日志设计
- [ ] 日志分析与告警

**实战项目**: Go 实现结构化日志库

**项目路径**: `projects/week9/structured-logging/`

**笔记路径**: `notes/week9/module1-logging.md`

---

#### 模块 2: 监控系统 (建议 3 天)

**核心学习内容**:

- [ ] Prometheus 架构深入
  - [ ] 时序数据模型
  - [ ] 指标类型(Counter、Gauge、Histogram、Summary)
  - [ ] PromQL 查询语言
  - [ ] 服务发现机制
- [ ] Grafana 可视化
  - [ ] Dashboard 设计
  - [ ] 告警规则配置
- [ ] 监控最佳实践
  - [ ] RED(Rate、Errors、Duration)方法
  - [ ] USE(Utilization、Saturation、Errors)方法
  - [ ] 金字塔监控模型

**实战项目**: 为微服务接入完整监控

- [ ] 暴露业务指标
- [ ] 自定义 Dashboard

**项目路径**: `projects/week9/monitoring/`

**笔记路径**: `notes/week9/module2-monitoring.md`

---

#### 模块 3: 分布式追踪 (建议 2 天)

**核心学习内容**:

- [ ] OpenTelemetry 标准
  - [ ] Trace、Span 概念
  - [ ] Context 传播
  - [ ] Sampling 策略
- [ ] Jaeger 架构
  - [ ] Agent、Collector、Query 组件
  - [ ] 存储后端选择
- [ ] 追踪数据分析
  - [ ] 调用链分析
  - [ ] 性能瓶颈定位

**实战项目**: Go 微服务链路追踪

- [ ] 集成 OpenTelemetry
- [ ] 跨服务追踪

**项目路径**: `projects/week9/tracing/`

**笔记路径**: `notes/week9/module3-tracing.md`

---

#### Week 9 总结

- [ ] 完成周总结: `notes/week9/weekly-summary.md`
- [ ] 设计完整的监控系统

**本周总时长**: _____ 小时

---

### Week 10: 安全与性能优化

**本周目标**: 系统安全设计与性能调优

#### 模块 1: 认证与授权 (建议 2 天)

**核心学习内容**:

- [ ] OAuth 2.0 协议深入
  - [ ] 四种授权模式
  - [ ] Token 刷新机制
  - [ ] PKCE 扩展
- [ ] JWT(JSON Web Token)
  - [ ] 结构与签名验证
  - [ ] Stateless 认证
  - [ ] Token 撤销策略
- [ ] RBAC 权限模型设计

**实战项目**: 实现完整的认证授权系统

**项目路径**: `projects/week10/auth-system/`

**笔记路径**: `notes/week10/module1-authentication.md`

---

#### 模块 2: API 安全 (建议 2 天)

**核心学习内容**:

- [ ] 常见安全威胁与防护
  - [ ] SQL 注入防护
  - [ ] XSS 攻击防护
  - [ ] CSRF 攻击防护
  - [ ] 重放攻击防护
- [ ] API 安全最佳实践
  - [ ] HTTPS 与证书管理
  - [ ] API 签名验证
  - [ ] 请求加密
- [ ] DDoS 防护策略

**实践**: 安全审计与漏洞扫描

**笔记路径**: `notes/week10/module2-api-security.md`

---

#### 模块 3: 性能优化 (建议 3 天)

**核心学习内容**:

- [ ] 数据库性能优化
  - [ ] 慢查询分析与优化
  - [ ] 索引设计优化
  - [ ] 查询重写技巧
  - [ ] 连接池调优
- [ ] 应用层优化
  - [ ] 代码性能分析(pprof、cProfile)
  - [ ] 并发优化技巧
  - [ ] 内存优化
  - [ ] GC 调优
- [ ] 架构层优化
  - [ ] 异步化改造
  - [ ] 批处理优化
  - [ ] 缓存优化

**实战**: 性能基准测试与调优

- [ ] 压测工具使用(wrk、JMeter)
- [ ] 性能瓶颈定位
- [ ] 优化效果验证

**项目路径**: `projects/week10/performance-tuning/`

**笔记路径**: `notes/week10/module3-performance.md`

---

#### Week 10 总结 & Phase 3 总结

- [ ] 完成 Week 10 总结: `notes/week10/weekly-summary.md`
- [ ] 完成 Phase 3 总结: `notes/phase3-summary.md`
- [ ] 高可用电商系统综合设计

**本周总时长**: _____ 小时
**Phase 3 总时长**: _____ 小时

---

## 第四阶段: 综合实战 (Week 11-12)

### Week 11: 经典系统设计深入分析

**本周目标**: 通过经典案例巩固所学知识

**每天深入研究一个经典系统**:

#### Day 1: URL 短链接服务

- [ ] 需求分析与功能设计
- [ ] 短链生成算法(Base62、Hash)
- [ ] 数据库设计与分片策略
- [ ] 缓存方案
- [ ] 高可用架构
- [ ] 完整代码实现与部署

**笔记路径**: `notes/week11/day1-url-shortener.md`

---

#### Day 2: Twitter 设计

- [ ] Feed 流架构(推拉结合)
- [ ] Timeline 生成算法
- [ ] 关注关系存储(图数据库)
- [ ] 热点数据处理(大V 问题)
- [ ] 数据分片与复制
- [ ] 分析 Twitter 真实架构演进

**笔记路径**: `notes/week11/day2-twitter.md`

---

#### Day 3: Instagram 设计

- [ ] 图片存储架构(对象存储 + CDN)
- [ ] Feed 排序算法
- [ ] 社交图谱设计
- [ ] 实时消息推送
- [ ] Facebook Instagram 技术分享解读

**笔记路径**: `notes/week11/day3-instagram.md`

---

#### Day 4: 分布式文件系统

- [ ] GFS/HDFS 架构分析
- [ ] Master-Slave 设计
- [ ] 数据分块与副本
- [ ] 故障恢复机制
- [ ] 阅读 Google File System 论文

**笔记路径**: `notes/week11/day4-distributed-fs.md`

---

#### Day 5: 网约车系统

- [ ] 地理位置索引(GeoHash、QuadTree、S2)
- [ ] 实时匹配算法
- [ ] 动态定价系统
- [ ] 高并发架构
- [ ] Uber 技术架构分析

**笔记路径**: `notes/week11/day5-ride-sharing.md`

---

#### Day 6: 秒杀系统

- [ ] 流量削峰方案
- [ ] 库存扣减(Redis + Lua)
- [ ] 分布式锁应用
- [ ] 防刷策略
- [ ] 降级与限流
- [ ] 完整秒杀系统实现

**笔记路径**: `notes/week11/day6-flash-sale.md`

---

#### Day 7: 总结与答辩

- [ ] 整理所有设计文档
- [ ] 准备技术分享
- [ ] 系统设计模板总结
- [ ] 完成 Week 11 总结: `notes/week11/weekly-summary.md`

**本周总时长**: _____ 小时
**设计案例**: ___/6
**代码实现**: _____ 个

---

### Week 12: 毕业项目

**本周目标**: 完成一个真实的分布式系统项目

#### Day 1-2: 项目规划

**项目选择**(选择一个):

- [ ] 实时协作文档系统(类 Google Docs)
- [ ] 分布式任务调度系统
- [ ] 实时聊天系统(支持群聊)
- [ ] 微博客平台
- [ ] 分布式缓存系统
- [ ] 其他: _____________________

**任务**:

- [ ] 需求分析与技术选型
- [ ] 架构设计文档
- [ ] 数据库设计
- [ ] API 设计

**文档路径**: `projects/week12/final-project/design.md`

---

#### Day 3-5: 核心功能开发

- [ ] 后端服务开发(Go 或 Python)
- [ ] 数据库集成
- [ ] 缓存层实现
- [ ] 消息队列集成
- [ ] 服务间通信
- [ ] 单元测试与集成测试

**项目路径**: `projects/week12/final-project/`

**代码质量检查**:

- [ ] 测试覆盖率 > 80%
- [ ] 所有公共 API 有文档
- [ ] 通过静态代码分析
- [ ] 性能基准测试通过

---

#### Day 6: 可观测性与优化

- [ ] 集成监控(Prometheus)
- [ ] 日志系统接入
- [ ] 链路追踪
- [ ] 性能测试
- [ ] 瓶颈分析与优化

**性能指标**:

- QPS: _____
- P99 延迟: _____ ms
- 可用性: _____%

---

#### Day 7: 项目总结与展示

- [ ] 完善项目文档
- [ ] 绘制完整架构图
- [ ] 部署到云平台
- [ ] 准备技术分享 PPT
- [ ] 代码开源到 GitHub

**最终总结**:

- [ ] 完成最终总结: `notes/final-summary.md`
- [ ] 12 周学习回顾
- [ ] 技能提升总结
- [ ] 未来学习计划

**项目 GitHub**: _____________________

---

## 🎉 课程完成总结

### 学习统计

**总学习时长**: _____ 小时

**完成进度**:

- Phase 1 (Week 1-3): ___/9 模块
- Phase 2 (Week 4-7): ___/12 模块
- Phase 3 (Week 8-10): ___/9 模块
- Phase 4 (Week 11-12): ___/7 案例 + 1 项目

**项目成果**:

- Go 项目: _____ 个
- Python 项目: _____ 个
- 源码阅读: _____ 行
- 笔记总量: _____ 字

**知识掌握度** (1-5 星自评):

- 可扩展性原理: ⭐⭐⭐⭐⭐
- 存储系统: ⭐⭐⭐⭐☆
- 网络通信: ⭐⭐⭐⭐☆
- 缓存系统: ⭐⭐⭐⭐⭐
- 数据库进阶: ⭐⭐⭐⭐☆
- 事件驱动: ⭐⭐⭐☆☆
- 微服务: ⭐⭐⭐⭐☆
- 分布式一致性: ⭐⭐⭐⭐⭐
- 可观测性: ⭐⭐⭐⭐☆
- 性能优化: ⭐⭐⭐⭐☆

### 能力提升

**编程能力**:

- Go: ⭐⭐⭐⭐⭐
- Python: ⭐⭐⭐⭐☆
- 并发编程: ⭐⭐⭐⭐☆
- 性能调优: ⭐⭐⭐⭐☆

**系统设计能力**:

- 需求分析: ⭐⭐⭐⭐☆
- 架构设计: ⭐⭐⭐⭐⭐
- 技术选型: ⭐⭐⭐⭐☆
- 权衡分析: ⭐⭐⭐⭐⭐

**工程能力**:

- 代码质量: ⭐⭐⭐⭐☆
- 测试能力: ⭐⭐⭐⭐☆
- 性能分析: ⭐⭐⭐⭐☆
- 源码阅读: ⭐⭐⭐⭐☆

### 最大收获

1. ---
2. ---
3. ---
4. ---
5. ---

### 未来计划

- [ ] 继续深入: _____________________
- [ ] 新的学习方向: _____________________
- [ ] 实践应用: _____________________
- [ ] 开源贡献: _____________________

---

## 💡 学习建议

### 深度学习的关键

1. **不要追求速度**: 理解一个概念的原理比完成进度更重要
2. **源码是最好的老师**: 每周至少阅读一个开源项目的核心代码
3. **动手实践**: 不要只看不写,每个概念都要亲手实现
4. **记录思考**: 笔记不是复制,而是你的理解和思考
5. **提问与讨论**: 遇到不懂的要深入研究,与他人讨论

### 模块化学习的优势

- ✅ 按主题深入,而非按天数赶进度
- ✅ 灵活调整节奏,适应个人理解深度
- ✅ 项目质量优于完成速度
- ✅ 鼓励源码阅读和原理探究
- ✅ 培养系统思维和权衡能力

### 如何判断模块完成

一个模块真正完成的标准:

- [ ] 能用自己的话解释核心原理
- [ ] 理解为什么这样设计(trade-offs)
- [ ] 代码实现了核心功能并通过测试
- [ ] 阅读了相关的开源项目源码
- [ ] 能够举一反三,应用到新场景

---

**坚持深度学习,成为真正的系统架构师!** 🚀

> "Depth over breadth. Understanding over memorization." - 深度优于广度,理解优于记忆

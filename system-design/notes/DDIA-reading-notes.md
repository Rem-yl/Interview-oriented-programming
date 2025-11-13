# 《数据密集型应用系统设计》(DDIA) 阅读笔记

> Designing Data-Intensive Applications by Martin Kleppmann
>
> 这是系统设计领域的圣经级著作,深入浅出地讲解了分布式系统的核心概念

---

## 📚 全书结构概览

### Part I: 数据系统基础 (Foundations of Data Systems)

- **Chapter 1**: 可靠、可扩展、可维护的应用系统
- **Chapter 2**: 数据模型与查询语言
- **Chapter 3**: 存储与检索
- **Chapter 4**: 编码与演化

### Part II: 分布式数据 (Distributed Data)

- **Chapter 5**: 复制
- **Chapter 6**: 分区
- **Chapter 7**: 事务
- **Chapter 8**: 分布式系统的麻烦
- **Chapter 9**: 一致性与共识

### Part III: 派生数据 (Derived Data)

- **Chapter 10**: 批处理
- **Chapter 11**: 流处理
- **Chapter 12**: 数据系统的未来

---

## 第一部分: 数据系统基础

### Chapter 1: 可靠、可扩展、可维护的应用系统

**阅读状态**: ✅ 已完成 | **日期**: 2025-10-16

#### 核心概念

**三大质量属性**:

1. **可靠性 (Reliability)**
   - 故障类型: 硬件故障、软件错误、人为错误
   - 容错设计原则
   - 关键词: Fault vs Failure, Fault-Tolerant

2. **可扩展性 (Scalability)**
   - 负载描述: QPS, 读写比, Twitter 扇出案例
   - 性能度量: 响应时间, 吞吐量, 百分位数 (P50/P95/P99)
   - 扩展方式: 垂直扩展 vs 水平扩展
   - 关键洞察: 没有通用架构,需根据负载特征设计

3. **可维护性 (Maintainability)**
   - 可操作性 (Operability): 让运维轻松
   - 简单性 (Simplicity): 通过抽象减少复杂度
   - 可演化性 (Evolvability): 让系统易于改变

#### 经典案例分析

**Twitter 时间线设计**:
- 方案1 (Pull): 读时查询合并 → 读压力大
- 方案2 (Push): 写时扇出 → 大V写入成本高
- 混合方案: 普通用户扇出 + 大V特殊处理

**关键数字**:
- 发推文: 4.6k req/s (平均), 12k req/s (峰值)
- 主页时间线: 300k req/s
- 瓶颈: **扇出导致的写放大**

#### 关键洞察

1. **P99 延迟比平均值重要**
   - Amazon: 100ms 延迟 = 1% 销售额损失
   - 最慢请求往往来自最有价值客户

2. **设计权衡无处不在**
   - Twitter: 优化读 vs 优化写
   - 没有银弹,根据负载特征选择

3. **复杂性是可维护性的大敌**
   - 意外复杂性 (实现) vs 本质复杂性 (问题)
   - 通过抽象减少复杂度

#### 延伸阅读

**必读文章**:
- [Scalability for Dummies - Part 1: Clones](https://www.lecloud.net/post/7295452622/scalability-for-dummies-part-1-clones)
- [Scalability for Dummies - Part 2: Database](https://www.lecloud.net/post/7994751381/scalability-for-dummies-part-2-database)
- [Scalability for Dummies - Part 3: Cache](https://www.lecloud.net/post/9246290032/scalability-for-dummies-part-3-cache)
- [Scalability for Dummies - Part 4: Asynchronism](https://www.lecloud.net/post/9699762917/scalability-for-dummies-part-4-asynchronism)

**视频资源**:
- [Horizontal vs Vertical Scaling](https://www.youtube.com/watch?v=xpDnVSmNFX0) (15分钟)
- [Understanding Latency Percentiles](https://www.youtube.com/watch?v=lJ8ydIuPFeU) (15分钟)

**技术博客**:
- [Twitter's Timeline Architecture](https://www.infoq.com/presentations/Twitter-Timeline-Scalability/)
- [Amazon: How 1 Second Could Cost $1.6B](https://www.fastcompany.com/1825005/how-one-second-could-cost-amazon-16-billion-sales)
- [Brendan Gregg's USE Method](https://www.brendangregg.com/usemethod.html)

**相关书籍**:
- 《Site Reliability Engineering》 - Google SRE 实践
- 《The Art of Scalability》 - AKF 扩展立方体

#### 实践练习

1. **容量规划练习**:
   - 假设: 100万 DAU, 每人 20 请求/天, 峰值因子 3x
   - 计算: 峰值 QPS = (1,000,000 × 20 / 86,400) × 3 ≈ 694 QPS

2. **Little's Law 应用**:
   - 公式: 并发用户数 = QPS × 平均响应时间
   - 示例: QPS=1000, RT=100ms → 需要 100 并发

3. **百分位数计算**:
   - 给定延迟数据 [1ms, 2ms, ..., 100ms]
   - 手工计算 P50, P95, P99

#### 我的思考

**问题**:
- ❓ 如果 Twitter 完全采用查询时合并会怎样?
- ❓ 为什么云时代更需要软件容错?
- ❓ 如何在微服务架构中保持简单性?

**答案要点**:
- 查询时合并会导致主页时间线查询成为瓶颈(需要 join 大量数据)
- 云环境虚拟机随时可能消失,机器数量增加,故障概率上升
- 通过领域驱动设计(DDD)、清晰的服务边界、标准化接口减少复杂度

---

### Chapter 2: 数据模型与查询语言

**阅读状态**: ⏳ 待学习

#### 预习大纲

**核心主题**:
1. 关系模型 vs 文档模型 vs 图模型
2. SQL vs NoSQL 的权衡
3. 声明式查询 vs 命令式查询
4. MapReduce 查询

**关键概念**:
- Schema-on-read vs Schema-on-write
- 数据局部性 (Data Locality)
- 多对一、多对多关系
- 图查询语言 (Cypher, SPARQL)

#### 延伸资料 (待整理)

**必读文章**:
- [MongoDB Schema Design Best Practices](https://www.mongodb.com/developer/products/mongodb/mongodb-schema-design-best-practices/)
- [Neo4j Graph Database Use Cases](https://neo4j.com/use-cases/)

**视频**:
- [SQL vs NoSQL - What's the Difference?](https://www.youtube.com/watch?v=Q_9cX9aKn1Y)

---

### Chapter 3: 存储与检索

**阅读状态**: ✅ 已完成 | **日期**: 2025-11-13

> 深入探讨数据库存储引擎的内部原理,理解不同数据结构的权衡

#### 核心概念

**本章核心问题**: 数据库如何存储数据?如何高效检索数据?

---

### 1️⃣ 存储引擎的两大流派

#### 流派对比

**日志结构存储引擎 (Log-Structured Storage)**:
- 代表: LSM-Tree (Cassandra, HBase, RocksDB, LevelDB)
- 核心思想: 追加写入,后台压缩
- 优势: **写优化**,顺序 I/O
- 劣势: 读需要查找多个数据结构

**面向页的存储引擎 (Page-Oriented Storage)**:
- 代表: B-Tree (MySQL InnoDB, PostgreSQL)
- 核心思想: 原地更新,固定大小页
- 优势: **读优化**,成熟稳定
- 劣势: 写放大,需要 WAL

---

### 2️⃣ 最简单的数据库: 追加日志

#### 极简实现

**Bash 版数据库**:
```bash
#!/bin/bash

# 写入 (追加到文件)
db_set() {
  echo "$1,$2" >> database
}

# 读取 (倒序查找最新值)
db_get() {
  grep "^$1," database | tail -n 1 | cut -d',' -f2
}
```

**使用示例**:
```bash
$ db_set 123 '{"name":"London","attractions":["Big Ben"]}'
$ db_set 456 '{"name":"San Francisco","attractions":["Golden Gate"]}'
$ db_get 123
{"name":"London","attractions":["Big Ben"]}
```

**性能特点**:
- **写入**: O(1) - 追加到文件末尾
- **读取**: O(n) - 扫描整个文件

**问题**: 随着数据增长,读取越来越慢!

---

### 3️⃣ 索引: 加速查询的代价

#### 索引的本质

**定义**: 额外的数据结构,用于快速定位数据

**权衡**:
- ✅ 加速读取
- ❌ 拖慢写入 (需要同时更新索引)
- ❌ 占用额外空间

**关键洞察**: 索引不是免费的,需要根据查询模式选择

#### 哈希索引 (Hash Index)

**最简单的索引结构**:

**实现原理**:
```
内存哈希表:
key → 文件偏移量

文件内容:
0: 123,{"name":"London"}
28: 456,{"name":"SF"}
50: 123,{"name":"New London"}  ← 更新

哈希表更新:
123 → 50  (指向最新值)
456 → 28
```

**性能**:
- **写入**: O(1) - 追加 + 更新哈希表
- **读取**: O(1) - 哈希查找 + 一次磁盘 I/O

**真实案例: Bitcask (Riak 的默认存储引擎)**
- 适合场景: 值频繁更新,键数量不大
- 示例: 视频播放次数、用户会话数据
- 限制: 所有键必须放入内存

#### 压缩 (Compaction) 与合并

**问题**: 追加日志会无限增长,浪费空间

**解决方案: 段压缩 (Segment Compaction)**

**工作原理**:
```
原始段:
123,{"name":"London"}
456,{"name":"SF"}
123,{"name":"New London"}  ← 覆盖旧值
456,{"name":"San Francisco"}  ← 覆盖旧值

压缩后:
123,{"name":"New London"}
456,{"name":"San Francisco"}
```

**压缩策略**:
1. **日志分段 (Segmentation)**
   - 达到阈值(如 1MB)后,关闭当前段,创建新段
   - 后台线程合并旧段

2. **合并过程**:
   ```
   段1: a=1, b=2, c=3
   段2: a=4, b=5
   段3: a=6

   合并后: a=6, b=5, c=3
   ```

3. **读取逻辑**:
   - 从新段到旧段依次查找
   - 找到第一个匹配即停止

**实现细节** (Bitcask):
- **删除**: 写入特殊"墓碑"标记,压缩时忽略
- **崩溃恢复**: 启动时重建哈希表 (或从快照恢复)
- **并发控制**: 只有一个写线程,多个读线程
- **部分写入**: 校验和检测损坏记录

**限制**:
- ❌ 哈希表必须放入内存
- ❌ 不支持范围查询 (如 key 从 kitty00000 到 kitty99999)

---

### 4️⃣ SSTables 与 LSM-Tree

#### SSTable (Sorted String Table)

**核心改进**: 键在段内**有序存储**

**SSTable vs 无序日志**:

**无序日志**:
```
123,London
456,SF
789,Tokyo
```

**SSTable (键有序)**:
```
123,London
456,SF
789,Tokyo
```

**有序带来的优势**:

**1. 合并更高效 (归并排序)**
```
段1: a=1, c=3, e=5
段2: b=2, c=4, d=6

合并 (类似归并排序):
a=1, b=2, c=4, d=6, e=5
```
- 线性扫描,不需要全部加载到内存
- 时间复杂度: O(n)

**2. 索引可以稀疏**
```
SSTable (每 4KB 一个索引点):
0KB:    "apple"
4KB:    "banana"
8KB:    "cherry"
12KB:   "grape"

查找 "coconut":
1. 索引查找: "banana" < "coconut" < "cherry"
2. 读取 4KB-8KB 块,二分查找
```
- 不需要为每个键维护索引
- 节省内存!

**3. 压缩块以节省 I/O**
- 稀疏索引指向压缩块
- 减少磁盘读取和网络传输

#### 构建与维护 SSTable

**问题**: 如何保证写入数据有序?

**解决方案: 内存排序 + 定期刷盘**

**LSM-Tree 架构**:

```
写入流程:
1. 写入 MemTable (内存红黑树/AVL树)
   MemTable: {c:3, a:1, b:2} → 有序: a:1, b:2, c:3

2. MemTable 达到阈值(如 4MB)
   刷盘为 SSTable (磁盘)

3. SSTable 不可变,只追加

磁盘结构:
SSTable-1 (newest): a:10, b:20, c:30
SSTable-2:          a:5,  d:15
SSTable-3 (oldest): a:1,  e:25

后台压缩:
合并多个 SSTable → 新 SSTable
删除旧文件
```

**查询流程**:
```
查找 key "a":
1. 先查 MemTable → 找到 a:10
2. 如果没找到,查 SSTable-1 (最新) → 找到就返回
3. 再查 SSTable-2, SSTable-3...
```

**性能优化: Bloom Filter**

**问题**: 查询不存在的键需要扫描所有 SSTable

**Bloom Filter 原理**:
```
位数组 + 哈希函数

插入 "apple":
hash1("apple") = 3 → 设置 bit[3] = 1
hash2("apple") = 7 → 设置 bit[7] = 1

查询 "banana":
hash1("banana") = 3 → bit[3] = 1
hash2("banana") = 5 → bit[5] = 0
结果: 一定不存在! (快速返回)

查询 "cherry":
hash1("cherry") = 3 → bit[3] = 1
hash2("cherry") = 7 → bit[7] = 1
结果: 可能存在 (需要查文件确认)
```

**特点**:
- 空间效率高(每个键几个 bit)
- 无假阴性(存在的键一定返回"可能存在")
- 有假阳性(不存在的键可能返回"可能存在",但概率可控)

#### 压缩策略

**Size-Tiered Compaction** (Cassandra, HBase):
```
Level 0: [1MB] [1MB] [1MB] [1MB]
         ↓ 合并
Level 1: [4MB] [4MB]
         ↓ 合并
Level 2: [16MB]
```
- 相似大小的段合并
- 写放大较小,但空间放大较大

**Leveled Compaction** (LevelDB, RocksDB):
```
Level 0: 4个 SSTable (可重叠)
Level 1: 10MB (不重叠,10 个 SSTable,每个 1MB)
Level 2: 100MB (不重叠,100 个 SSTable,每个 1MB)
...
```
- 每层容量限制
- 层间合并时,只处理重叠键范围
- 读放大小,但写放大较大

**对比**:
| 压缩策略 | 写放大 | 读放大 | 空间放大 |
|---------|-------|-------|---------|
| Size-Tiered | 低 | 高 | 高 |
| Leveled | 高 | 低 | 低 |

#### LSM-Tree 的优势与劣势

**优势**:
- ✅ **写吞吐量高**: 顺序写入,批量合并
- ✅ **压缩效率高**: 数据有序,易于压缩
- ✅ **适合写密集型**: 日志、监控数据

**劣势**:
- ❌ **读可能慢**: 需要查找多个 SSTable
- ❌ **压缩开销**: 后台压缩消耗 CPU 和 I/O
- ❌ **写放大**: 数据可能被多次重写

**真实案例**:
- **LevelDB**: Google 开发,Chrome 浏览器存储
- **RocksDB**: Facebook 基于 LevelDB 优化
- **Cassandra**: 分布式 NoSQL 数据库
- **HBase**: Hadoop 生态系统

---

### 5️⃣ B-Tree: 最流行的索引结构

#### B-Tree 的设计

**核心思想**: 将数据库分解成固定大小的**页 (page)**,通常 4KB

**B-Tree 结构**:
```
                [Root Page]
               /     |     \
         [Page1]  [Page2]  [Page3]
         /  |  \   /  \     /  \
      [Leaf][Leaf][Leaf][Leaf][Leaf]...

示例 (每页最多 3 个键):
              [10 | 20]
            /    |     \
      [3|5|7] [12|15] [25|30|35]

查找 12:
1. 根页: 10 < 12 < 20 → 走中间分支
2. 叶页: 找到 12
磁盘 I/O: 2 次
```

**关键特性**:

**1. 平衡树**
- 所有叶子节点深度相同
- 深度 = log_b(n), b 是分支因子(branching factor)

**示例**:
```
分支因子 b = 100
4KB 页,容纳 100 个键
100 万键: log_100(1,000,000) ≈ 3 层
10 亿键: log_100(1,000,000,000) ≈ 4-5 层
```

**2. 原地更新**
- 找到叶页,直接覆盖
- 不像 LSM-Tree 追加写入

**3. 范围查询高效**
```
查询 key 从 10 到 30:
定位到 10 的叶页
顺序扫描相邻叶页
```

#### B-Tree 的插入与分裂

**正常插入**:
```
插入前: [3 | 7 | 12]
插入 5:  [3 | 5 | 7 | 12]
```

**页满时分裂**:
```
插入前 (页满,最多 3 个键):
Parent:    [10]
            /  \
Child:  [3|5|7|9]  [12|15]

插入 6:
1. 叶页满,分裂成两页
2. 中间键 (7) 提升到父页

插入后:
Parent:    [7 | 10]
          /   |   \
      [3|5|6] [9] [12|15]
```

**父页也满时**:
```
递归分裂,可能导致根页分裂 → 树高度增加
```

**关键洞察**: B-Tree 深度增加缓慢 (log 增长)

#### B-Tree 的可靠性保障

**问题 1: 崩溃恢复**

**场景**:
```
正在分裂页:
1. 写入新页
2. 更新父页指针
3. 释放旧页

如果在步骤 2 崩溃? → 数据库损坏!
```

**解决方案: Write-Ahead Log (WAL)**
```
操作顺序:
1. 写 WAL: "即将插入 key=5, 分裂页 X"
2. 执行操作: 修改 B-Tree 页
3. WAL 标记完成

崩溃恢复:
扫描 WAL → 重放未完成操作
```

**WAL 的作用**:
- 重做日志 (Redo Log) - InnoDB
- 崩溃后恢复一致性状态

**问题 2: 并发控制**

**场景**: 多个线程同时修改 B-Tree

**解决方案: Latches (轻量级锁)**
```
读操作: 共享锁 (允许多个读)
写操作: 排他锁 (独占访问)

优化: Latch Crabbing
从根到叶一路加锁
确认子节点安全后,释放父节点锁
```

#### B-Tree 优化技巧

**1. 写时复制 (Copy-on-Write)**
```
修改页时:
1. 复制一份新页
2. 修改新页
3. 更新父页指针

优势:
- 无需 WAL
- 天然支持快照隔离

示例: LMDB, BoltDB
```

**2. 页内压缩**
```
不存储完整键,只存储差异

原始:
"handlebars", "handlebra", "handlebar"

压缩:
"handlebars", "-s+a", "-a+r"

节省空间,每页容纳更多键
```

**3. 兄弟页指针**
```
叶页间维护链表:
[Page1] → [Page2] → [Page3]

范围查询不需要回到父页
```

**4. B+ Tree 变种**
```
B-Tree: 内部节点和叶节点都存数据
B+ Tree: 只有叶节点存数据,内部节点只存键

优势:
- 内部节点更小,可以缓存更多
- 范围查询更快 (叶节点链表)

MySQL InnoDB 使用 B+ Tree
```

---

### 6️⃣ LSM-Tree vs B-Tree 对比

#### 性能对比

| 维度 | LSM-Tree | B-Tree |
|-----|----------|--------|
| **写性能** | ✅ 高 (顺序写) | ❌ 低 (随机写 + WAL) |
| **读性能** | ❌ 慢 (多个 SSTable) | ✅ 快 (直接定位) |
| **范围查询** | ❌ 较慢 | ✅ 快 (顺序扫描) |
| **空间放大** | 中等 (取决于压缩策略) | 低 (页内碎片) |
| **写放大** | 高 (多次压缩) | 中等 (WAL + 页更新) |

#### 写放大 (Write Amplification)

**定义**: 写入 1 字节数据,实际磁盘写入多少字节?

**B-Tree 的写放大**:
```
写入 1 个键值对 (100 字节):
1. 写 WAL: 100 字节
2. 写 B-Tree 页: 4KB (整页写入)
3. 可能分裂: 再写 4KB

写放大 = 8KB / 100 字节 = 80x
```

**LSM-Tree 的写放大**:
```
写入 1 个键值对 (100 字节):
1. 写 MemTable: 100 字节
2. 刷盘 SSTable: 假设 4MB
3. 压缩 5 次 (Leveled): 每次重写

写放大 = 5 倍以上
```

**但 LSM-Tree 是顺序写, B-Tree 是随机写**:
- HDD: 顺序写 100MB/s vs 随机写 1MB/s (100x 差异!)
- SSD: 顺序写仍快于随机写

#### 真实场景选择

**选择 LSM-Tree (写密集)**:
- 日志收集系统 (Elasticsearch, Cassandra)
- 时序数据库 (InfluxDB)
- 高写入吞吐量需求

**选择 B-Tree (读密集)**:
- 事务型数据库 (MySQL, PostgreSQL)
- 范围查询频繁
- 写入量可控

**混合场景**:
- TiDB: LSM-Tree 存储 + B-Tree 索引
- MyRocks: MySQL + RocksDB

---

### 7️⃣ 其他索引结构

#### 聚簇索引 vs 非聚簇索引

**聚簇索引 (Clustered Index)**:
```
B-Tree 叶节点直接存储完整行

索引结构 (主键 = user_id):
        [Root]
         /  \
    [Leaf1] [Leaf2]
      ↓       ↓
   Row data Row data
   (id=1,   (id=5,
    name=   name=
    "Alice") "Bob")

优势: 查询主键不需要额外 I/O
劣势: 二级索引需要存储主键 (占空间)

示例: MySQL InnoDB (主键索引)
```

**非聚簇索引 (Secondary Index)**:
```
B-Tree 叶节点存储指针 (行号或主键)

索引结构 (name):
       [Root]
        /  \
   [Leaf1] [Leaf2]
     ↓       ↓
   "Alice" → RowID=1 → Heap File Row 1
   "Bob"   → RowID=5 → Heap File Row 5

需要回表查询!

示例: MySQL InnoDB (非主键索引指向主键)
```

**覆盖索引 (Covering Index)**:
```
索引包含查询所需所有列

CREATE INDEX idx_name_age ON users(name, age);

SELECT name, age FROM users WHERE name = 'Alice';
↑ 无需回表,直接从索引返回!
```

#### 多列索引 (Multi-Column Index)

**1. 组合索引 (Concatenated Index)**
```
CREATE INDEX idx_lastname_firstname ON users(lastname, firstname);

数据组织:
"Smith, John"
"Smith, Alice"
"Wilson, Bob"

支持查询:
✅ WHERE lastname = 'Smith'
✅ WHERE lastname = 'Smith' AND firstname = 'John'
❌ WHERE firstname = 'John' (无法使用索引!)
```

**2. 多维索引 (Multi-Dimensional Index)**

**R-Tree (空间索引)**:
```
用于地理位置查询

数据: 餐厅位置 (经度, 纬度)

查询: 查找附近 1km 内的餐厅
SELECT * FROM restaurants
WHERE ST_Distance(location, '(37.7749, -122.4194)') < 1000;

R-Tree 将二维空间分割成矩形
高效裁剪搜索空间

PostGIS, MongoDB 地理索引使用 R-Tree
```

**Space-Filling Curve**:
```
将二维空间映射到一维

Z-order curve (Morton code):
(x, y) → 一维值

示例:
(0,0) → 00
(1,0) → 01
(0,1) → 10
(1,1) → 11

可以用 B-Tree 索引一维值!
```

#### 全文搜索索引

**倒排索引 (Inverted Index)**:
```
文档:
Doc1: "the quick brown fox"
Doc2: "the fox jumped"

倒排索引:
"brown" → [Doc1]
"fox"   → [Doc1, Doc2]
"jumped" → [Doc2]
"quick" → [Doc1]
"the"   → [Doc1, Doc2]

查询 "fox jumped":
"fox" → [Doc1, Doc2]
"jumped" → [Doc2]
交集 → [Doc2]
```

**优化: 位置信息**:
```
"fox" → [(Doc1, pos=3), (Doc2, pos=1)]
"jumped" → [(Doc2, pos=2)]

短语查询 "fox jumped" (相邻词):
Doc2 中 "fox" 在位置 1, "jumped" 在位置 2 → 匹配!
```

**示例**:
- Elasticsearch, Solr: 基于 Lucene
- MongoDB: 全文索引

#### 内存数据库索引

**问题**: 内存数据库为什么还需要索引?

**答案**: 避免顺序扫描,提升查询速度

**特殊优化**:
- **无需持久化索引** (内存足够快)
- **使用更复杂数据结构** (CPU 缓存友好)

**T-Tree** (内存 B-Tree 变种):
```
节点间指针替代页引用
更紧凑的内存布局
```

**Judy Array** (压缩前缀树):
```
极致压缩的 Trie
用于稀疏键
```

---

### 8️⃣ 事务处理与分析 (OLTP vs OLAP)

#### 两种工作负载

**OLTP (Online Transaction Processing)**:
```
特点:
- 小批量读写 (SELECT, INSERT, UPDATE 单行)
- 低延迟 (毫秒级)
- 随机访问
- 用户请求驱动

示例:
- 电商下单
- 银行转账
- 用户登录
```

**OLAP (Online Analytical Processing)**:
```
特点:
- 大批量读取 (聚合查询,扫描数百万行)
- 高吞吐量
- 顺序扫描
- 分析师查询驱动

示例:
- 销售报表 (过去 30 天总销售额)
- 用户行为分析
- 数据挖掘
```

**对比表**:

| 维度 | OLTP | OLAP |
|-----|------|------|
| **读模式** | 少量行,根据键查询 | 大量行聚合 |
| **写模式** | 随机写入,低延迟 | 批量导入 (ETL) |
| **用户** | 终端用户 (Web 应用) | 分析师 (BI 工具) |
| **数据量** | GB ~ TB | TB ~ PB |
| **索引** | B-Tree, LSM-Tree | 列存储, 位图索引 |

#### 数据仓库 (Data Warehouse)

**问题**: 在 OLTP 数据库上跑分析查询会怎样?

**后果**:
- 扫描大量数据,锁表
- 拖慢交易处理
- 影响用户体验

**解决方案: 分离存储**
```
OLTP 数据库 → ETL → 数据仓库 (OLAP)

ETL (Extract-Transform-Load):
1. 从生产数据库提取数据
2. 转换为分析友好格式 (星型/雪花模式)
3. 加载到数据仓库
```

**星型模式 (Star Schema)**:
```
事实表 (Fact Table):
sales (sale_id, product_id, store_id, date_id, amount)

维度表 (Dimension Tables):
products (product_id, name, category)
stores (store_id, name, city)
dates (date_id, year, month, day)

查询示例:
SELECT p.category, SUM(s.amount)
FROM sales s
JOIN products p ON s.product_id = p.product_id
WHERE d.year = 2024
GROUP BY p.category
```

**雪花模式 (Snowflake Schema)**:
```
维度表进一步规范化

products (product_id, name, category_id)
  ↓
categories (category_id, name, department_id)
  ↓
departments (department_id, name)
```

---

### 9️⃣ 列式存储 (Column-Oriented Storage)

#### 为什么需要列存储?

**分析查询特点**:
```sql
SELECT SUM(sales.amount)
FROM sales
WHERE sales.date >= '2024-01-01';
```
- 只需要 `amount` 和 `date` 两列
- 但行存储要读取整行!

**行存储 vs 列存储**:

**行存储 (Row-Oriented)**:
```
磁盘布局:
Row1: id=1, date=2024-01-01, product_id=10, amount=100
Row2: id=2, date=2024-01-02, product_id=20, amount=200
Row3: id=3, date=2024-01-03, product_id=10, amount=150

读取 100 万行:
需要读取所有列 → 浪费 I/O!
```

**列存储 (Column-Oriented)**:
```
磁盘布局:
id:         [1, 2, 3, ...]
date:       [2024-01-01, 2024-01-02, 2024-01-03, ...]
product_id: [10, 20, 10, ...]
amount:     [100, 200, 150, ...]

读取 100 万行的 amount:
只读取 amount 列 → 节省 I/O!
```

**性能提升**:
```
示例: 100 列, 只查询 3 列
行存储: 读取 100 列
列存储: 读取 3 列
I/O 减少 97%!
```

#### 列压缩

**列存储天然适合压缩**:

**原因**: 同一列数据类型相同,重复值多

**Bitmap Encoding** (位图编码):
```
原始数据 (product_id 列):
[10, 20, 10, 10, 30, 20, 10]

编码:
product_id=10: [1, 0, 1, 1, 0, 0, 1]
product_id=20: [0, 1, 0, 0, 0, 1, 0]
product_id=30: [0, 0, 0, 0, 1, 0, 0]

查询 product_id IN (10, 20):
bitmap_10 OR bitmap_20
= [1, 1, 1, 1, 0, 1, 1]

CPU 可以直接对位图做位运算!
```

**Run-Length Encoding** (游程编码):
```
原始位图:
[1, 1, 1, 1, 0, 0, 0, 1, 1]

压缩:
1×4, 0×3, 1×2

大量连续重复值时极致压缩!
```

**压缩比**:
- 典型数据仓库: 10:1 ~ 100:1 压缩比
- 更少 I/O,更快查询!

#### 列存储的写入

**问题**: 列存储如何高效写入?

**挑战**:
- 插入一行需要更新所有列文件
- 原地更新会破坏压缩

**解决方案: LSM-Tree 思想**
```
1. 新数据写入内存存储 (行存储)
2. 积累足够数据后,批量转换为列存储
3. 查询时合并内存和磁盘数据

类似 Vertica, Parquet 的设计
```

**Parquet 文件格式**:
```
文件结构:
Row Group 1 (数百万行):
  Column 1 Chunk (压缩)
  Column 2 Chunk (压缩)
  ...
Row Group 2:
  ...

优势:
- 大块顺序 I/O
- 极致压缩
- 支持嵌套数据 (JSON)
```

---

### 🔟 聚合: 物化视图与数据立方

#### 物化视图 (Materialized View)

**普通视图 (Virtual View)**:
```sql
CREATE VIEW sales_summary AS
SELECT product_id, SUM(amount) as total
FROM sales
GROUP BY product_id;

-- 查询时实时计算
SELECT * FROM sales_summary;
```

**物化视图 (Materialized View)**:
```sql
CREATE MATERIALIZED VIEW sales_summary AS
SELECT product_id, SUM(amount) as total
FROM sales
GROUP BY product_id;

-- 预先计算并存储结果
-- 查询直接读取,极快!
```

**权衡**:
- ✅ 查询快 (预计算)
- ❌ 占用空间
- ❌ 写入慢 (需要更新物化视图)

**适用场景**: 频繁查询的聚合结果

#### 数据立方 (Data Cube / OLAP Cube)

**多维聚合**:
```
维度: date, product, store
度量: SUM(sales)

预计算所有组合:
- 按 date 聚合
- 按 product 聚合
- 按 date + product 聚合
- 按 date + product + store 聚合
- ...

2^N 种组合!
```

**立方体可视化**:
```
        Store
         ↑
         |
         +----→ Product
        /
       /
      ↓
    Date

每个单元格存储预聚合值
查询瞬间返回!
```

**劣势**:
- 维度多时,组合爆炸 (10 维 = 1024 种组合)
- 占用海量空间
- 灵活性差 (只能查询预定义维度)

**现代趋势**: 放弃数据立方,使用列存储 + 压缩

---

### 关键洞察

1. **没有万能的存储引擎**
   - LSM-Tree: 写优化,适合日志和时序数据
   - B-Tree: 读优化,适合事务型数据库

2. **索引是查询性能的关键**
   - 但索引不是免费的 (写放大,空间占用)
   - 需要根据查询模式选择

3. **OLTP 和 OLAP 需要不同存储**
   - OLTP: 行存储 + B-Tree
   - OLAP: 列存储 + 压缩

4. **顺序 I/O 是性能关键**
   - LSM-Tree 通过顺序写提升性能
   - 列存储通过批量读提升性能

5. **压缩可以大幅提升性能**
   - 减少 I/O
   - 更好利用 CPU 缓存

---

### 延伸阅读

#### 经典论文

**LSM-Tree**:
- [The Log-Structured Merge-Tree (LSM-Tree)](https://www.cs.umb.edu/~poneil/lsmtree.pdf) - O'Neil et al., 1996

**B-Tree**:
- [The Ubiquitous B-Tree](https://dl.acm.org/doi/10.1145/356770.356776) - Comer, 1979
- [Organization and Maintenance of Large Ordered Indices](https://infolab.usc.edu/csci585/Spring2010/den_ar/indexing.pdf) - Bayer & McCreight, 1970 (原始 B-Tree 论文)

**列存储**:
- [C-Store: A Column-oriented DBMS](http://db.csail.mit.edu/projects/cstore/vldb.pdf) - Stonebraker et al., 2005
- [Dremel: Interactive Analysis of Web-Scale Datasets](https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/36632.pdf) - Google, 2010

#### 技术文章

**LSM-Tree 深度解析**:
- [LSM Trees: The Go-To Data Structure for Databases and Storage Systems](https://tikv.org/deep-dive/key-value-engine/b-tree-vs-lsm/)
- [RocksDB: A Persistent Key-Value Store](https://github.com/facebook/rocksdb/wiki)
- [LevelDB Implementation Notes](https://github.com/google/leveldb/blob/main/doc/impl.md)

**B-Tree 实现**:
- [SQLite B-Tree Module](https://www.sqlite.org/btreemodule.html)
- [PostgreSQL B-Tree Index](https://www.postgresql.org/docs/current/btree-implementation.html)
- [InnoDB B+Tree Index](https://dev.mysql.com/doc/refman/8.0/en/innodb-physical-structure.html)

**列存储**:
- [Parquet: Columnar Storage for Hadoop](https://parquet.apache.org/docs/)
- [Apache Arrow: In-Memory Columnar Format](https://arrow.apache.org/)

#### 视频资源

- [B-Trees and B+ Trees. How they are useful in Databases](https://www.youtube.com/watch?v=aZjYr87r1b8) (15 分钟)
- [LSM Trees Explained](https://www.youtube.com/watch?v=I6jB0nM9SKU) (20 分钟)
- [Column vs Row-Oriented Databases](https://www.youtube.com/watch?v=Vw1fCeD06YI) (12 分钟)

#### 开源项目源码

**LSM-Tree 实现**:
- [RocksDB](https://github.com/facebook/rocksdb) - Facebook, C++
- [LevelDB](https://github.com/google/leveldb) - Google, C++
- [BadgerDB](https://github.com/dgraph-io/badger) - Go 实现

**B-Tree 实现**:
- [BoltDB](https://github.com/boltdb/bolt) - Go, B+Tree, Copy-on-Write
- [LMDB](https://github.com/LMDB/lmdb) - C, B+Tree, Memory-Mapped

**列存储**:
- [Apache Parquet](https://github.com/apache/parquet-format)
- [ClickHouse](https://github.com/ClickHouse/ClickHouse) - 列存储 OLAP 数据库

#### 相关书籍

- 《Database Internals》 - Alex Petrov (深入数据库内部实现)
- 《The Art of Database Design》 - 数据库设计原理

---

### 实践练习

#### 练习 1: 实现简单 LSM-Tree

**任务**: 用 Go/Python 实现一个简化版 LSM-Tree

**要求**:
1. 内存 MemTable (使用跳表或红黑树)
2. SSTable 文件格式 (键有序)
3. 合并压缩逻辑
4. Bloom Filter 优化

**参考**: [Week 2 项目 - LSM-Tree 存储引擎](../../projects/week2/)

#### 练习 2: B-Tree 插入模拟

**任务**: 手动模拟 B-Tree 插入过程

**场景**:
- B-Tree 阶数 = 3 (每个节点最多 2 个键)
- 依次插入: 10, 20, 5, 6, 12, 30, 7, 17

**步骤**:
1. 画出每次插入后的树结构
2. 标注分裂时刻
3. 计算树高度

#### 练习 3: 列存储性能对比

**任务**: 对比行存储和列存储的查询性能

**实验**:
1. 创建 100 万行数据 (10 列)
2. 行存储: CSV 文件
3. 列存储: Parquet 文件
4. 查询: `SELECT AVG(column_3) WHERE column_1 > 50000`
5. 对比 I/O 和时间

**工具**: Python + Pandas + PyArrow

#### 练习 4: 索引选择题

**场景 1**: 查询 `SELECT * FROM users WHERE email = 'alice@example.com'`
- 应该建什么索引?
- **答案**: B-Tree 索引 on `email`

**场景 2**: 查询 `SELECT * FROM events WHERE timestamp BETWEEN '2024-01-01' AND '2024-01-31'`
- 应该建什么索引?
- **答案**: B-Tree 索引 on `timestamp` (支持范围查询)

**场景 3**: 查询 `SELECT * FROM restaurants WHERE ST_Distance(location, POINT(37.7, -122.4)) < 1000`
- 应该建什么索引?
- **答案**: R-Tree 空间索引 on `location`

**场景 4**: 时序日志数据,频繁写入,偶尔查询
- 应该用什么存储引擎?
- **答案**: LSM-Tree (InfluxDB, Elasticsearch)

---

### 我的思考

**问题**:
- ❓ 为什么 RocksDB 比 LevelDB 更快?
- ❓ SSD 是否让 LSM-Tree 失去优势?
- ❓ 为什么 PostgreSQL 不使用 LSM-Tree?

**答案要点**:

**RocksDB vs LevelDB**:
- RocksDB 优化: 多线程压缩, Column Family, 更灵活的压缩策略
- 适合生产环境,LevelDB 更适合学习

**SSD 的影响**:
- SSD 随机读写比 HDD 快,但顺序写仍优于随机写
- LSM-Tree 减少写放大,延长 SSD 寿命 (减少擦除次数)
- LSM-Tree 优势仍在,但差距缩小

**PostgreSQL 选择 B-Tree**:
- 事务型数据库,需要原地更新 (UPDATE 操作)
- LSM-Tree 的删除和更新需要墓碑标记,压缩才能回收
- B-Tree 更符合 MVCC 实现 (通过版本链)

**在实际项目中的应用**:
- Week 2 项目会实现 B-Tree 索引
- Week 4 项目会实现 LSM-Tree 存储引擎
- Week 5 项目会实现列存储格式

---

---

### Chapter 4: 编码与演化

**阅读状态**: ⏳ 待学习

#### 预习大纲

**核心主题**:
1. 编码格式对比
   - JSON, XML, Protocol Buffers, Avro, Thrift
2. 模式演化
   - 向前兼容、向后兼容
3. 数据流模式
   - 数据库中的数据流
   - 服务间的数据流 (REST, RPC)
   - 异步消息传递

**关键概念**:
- Schema Evolution
- Backward/Forward Compatibility
- RPC vs REST

#### 延伸资料 (待整理)

**技术对比**:
- [Protocol Buffers vs JSON: Size and Performance](https://auth0.com/blog/beating-json-performance-with-protobuf/)
- [Avro vs Protobuf vs Thrift](https://martin.kleppmann.com/2012/12/05/schema-evolution-in-avro-protocol-buffers-thrift.html)

---

## 第二部分: 分布式数据

### Chapter 5: 复制

**阅读状态**: ⏳ 待学习

#### 预习大纲

**核心主题**:
1. 主从复制 (Leader-Follower)
   - 同步复制 vs 异步复制
   - 复制延迟问题
2. 多主复制 (Multi-Leader)
   - 冲突检测与解决
   - 拓扑结构
3. 无主复制 (Leaderless)
   - Quorum 一致性
   - 反熵与读修复

**关键问题**:
- 复制延迟导致的一致性问题
- 写入冲突的解决策略
- 读己之写、单调读、一致性前缀读

#### 延伸资料 (待整理)

**论文**:
- [Dynamo: Amazon's Highly Available Key-value Store](https://www.allthingsdistributed.com/files/amazon-dynamo-sosp2007.pdf)

**文章**:
- [MySQL Replication Explained](https://dev.mysql.com/doc/refman/8.0/en/replication.html)
- [Cassandra Replication Strategy](https://cassandra.apache.org/doc/latest/cassandra/architecture/dynamo.html)

---

### Chapter 6: 分区

**阅读状态**: ⏳ 待学习

#### 预习大纲

**核心主题**:
1. 分区方法
   - 键范围分区
   - 哈希分区
2. 分区与二级索引
   - 基于文档的分区
   - 基于词条的分区
3. 分区再平衡
   - 固定数量分区
   - 动态分区
   - 按节点比例分区

**关键概念**:
- Consistent Hashing
- Hot Spot (热点)
- Skewed Workload (倾斜负载)

#### 延伸资料 (待整理)

**论文**:
- [Consistent Hashing and Random Trees](https://www.akamai.com/content/dam/site/en/documents/research-paper/consistent-hashing-and-random-trees-distributed-caching-protocols-for-relieving-hot-spots-on-the-world-wide-web-technical-publication.pdf) - Karger et al.

**文章**:
- [Consistent Hashing in Practice](https://www.toptal.com/big-data/consistent-hashing)
- [Cassandra Partitioning](https://cassandra.apache.org/doc/latest/cassandra/architecture/dynamo.html#dataset-partitioning-consistent-hashing)

---

### Chapter 7: 事务

**阅读状态**: ⏳ 待学习

#### 预习大纲

**核心主题**:
1. ACID 的含义
   - 原子性、一致性、隔离性、持久性
2. 弱隔离级别
   - 读已提交
   - 快照隔离 (MVCC)
   - 可重复读
3. 可串行化
   - 真正的串行执行
   - 两阶段锁 (2PL)
   - 可串行化快照隔离 (SSI)

**关键问题**:
- 脏读、脏写、幻读
- 写倾斜、丢失更新
- MVCC 的实现原理

#### 延伸资料 (待整理)

**论文**:
- [A Critique of ANSI SQL Isolation Levels](https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/tr-95-51.pdf) - Berenson et al.

**文章**:
- [PostgreSQL MVCC Explained](https://www.postgresql.org/docs/current/mvcc.html)
- [MySQL InnoDB MVCC](https://dev.mysql.com/doc/refman/8.0/en/innodb-multi-versioning.html)

---

### Chapter 8: 分布式系统的麻烦

**阅读状态**: ✅ 已完成 | **日期**: 2025-10-16

> 这一章揭示了分布式系统的残酷现实:网络不可靠、时钟不可信、节点随时可能失效

#### 核心概念

**本章主题**: 分布式系统中那些不可避免的、违背直觉的问题

---

### 1️⃣ 故障与部分失效

#### 单机系统 vs 分布式系统

**单机系统的特点**:
- 要么全部工作,要么完全失效(确定性)
- 硬件设计为"全有或全无"
- 软件 bug 是确定性的(相同输入总是相同输出)

**分布式系统的特点**:
- **部分失效 (Partial Failure)** 是常态
- 某些节点正常,某些节点失效
- 系统部分可用,部分不可用
- **不确定性 (Nondeterminism)** 是核心挑战

#### 部分失效的例子

```
场景: 分布式数据库,3 个节点

正常情况:
节点A: ✅ 正常
节点B: ✅ 正常
节点C: ✅ 正常

部分失效:
节点A: ✅ 正常
节点B: ❌ 网络不可达 (到底是宕机还是网络问题?)
节点C: ✅ 正常

问题: 无法区分节点B是宕机还是网络故障!
```

**关键洞察**: 在分布式系统中,你无法确定远程节点的状态

---

### 2️⃣ 不可靠的网络

#### 网络故障的现实

**网络并不像我们想象的那么可靠**

**常见网络问题**:
1. **请求丢失** - 数据包在传输中丢失
2. **请求排队** - 网络拥塞,请求被延迟
3. **远程节点失效** - 目标节点崩溃
4. **远程节点临时停止响应** - GC 暂停、CPU 占满
5. **响应丢失** - 响应包在返回途中丢失
6. **响应延迟** - 响应被严重延迟

**核心问题**: 发送请求后没有收到响应,你无法区分这些情况!

#### 真实世界的网络故障

**案例: Aphyr 的网络分区实验**
- 测试了多个分布式系统
- 发现大多数系统在网络分区时表现不佳
- 数据丢失、一致性违背、脑裂问题

**统计数据** (来自真实生产环境):
- 中型数据中心: 12 次网络故障/月
- 网络延迟 P99.9: 可能比中位数高 10 倍以上
- 交换机故障: 导致大规模连接中断

**数据中心网络 vs 互联网**:
- 数据中心内: 相对可靠,但仍有故障
- 跨数据中心: 故障率更高
- 移动网络: 极不可靠

#### 超时与无限等待

**问题**: 请求发出后,应该等多久?

**权衡**:
- **超时太短**: 误判节点失效,导致不必要的故障转移
- **超时太长**: 真正故障时恢复时间长,影响用户体验

**没有完美的超时时间**:
- 网络延迟是变化的(P50 vs P99 差异巨大)
- 节点处理时间也是变化的(GC 暂停)

**实践方法**:
1. **动态调整超时**: 根据响应时间分布自动调整
2. **Phi Accrual Failure Detector**: 不是二元判断,而是给出故障概率
3. **多次探测**: 避免单次超时就判定失效

#### 网络拥塞与排队

**数据包在哪里排队?**

1. **网络交换机队列**
   - 多个节点同时发送到同一目标
   - 交换机端口带宽饱和,数据包排队

2. **操作系统队列**
   - CPU 被占满,网络数据包在 OS 层排队
   - 虚拟机争夺 CPU,导致额外延迟

3. **TCP 流控制**
   - 接收方处理不过来,发送方被限速

4. **应用层队列**
   - 消息队列、线程池满了
   - 请求在应用内部排队

**关键洞察**: 可变延迟的主要来源是**排队**

**解决方案**:
- **实验性测量**: 持续监控网络延迟分布
- **过载保护**: 限流、熔断、降级
- **流量整形**: 避免突发流量

---

### 3️⃣ 不可靠的时钟

#### 时钟的两种类型

**1. 墙上时钟 (Time-of-Day Clock)**

**特点**:
- 返回当前日期和时间(如 2025-10-16 14:30:00)
- 对应 UNIX 时间戳
- 可以通过 NTP 同步

**问题**:
- **可以回拨** - NTP 发现时钟快了,会向后调整
- **跳跃性变化** - 闰秒、时区调整
- **不适合测量持续时间**

**用途**: 记录事件发生的时间点

**2. 单调时钟 (Monotonic Clock)**

**特点**:
- 保证单调递增
- 适合测量时间间隔(如 5 秒后超时)
- 不对应现实世界时间
- **不能跨节点比较**

**用途**: 超时、性能测量、速率限制

#### 时钟同步的现实

**NTP 同步并不完美**:

**延迟问题**:
```
客户端发送请求: T1 = 10:00:00.000
服务器接收请求: T2 = 10:00:00.050 (服务器时间)
服务器发送响应: T3 = 10:00:00.051 (服务器时间)
客户端接收响应: T4 = 10:00:00.100

网络延迟: (T4 - T1) - (T3 - T2) = 100ms - 1ms = 99ms
时钟偏移: ((T2 - T1) + (T3 - T4)) / 2

问题: 假设往返延迟对称(实际上不一定!)
```

**精度限制**:
- 公网 NTP: ±35 毫秒(Google 报告)
- 本地 NTP: ±1 毫秒
- GPS 接收器: ±1 微秒
- 原子钟: 极高精度,但昂贵

**时钟漂移**:
- 石英钟每天漂移: ±17 秒
- 需要持续同步
- 虚拟机的时钟更不可靠

#### 依赖时钟的危险

**案例 1: Last Write Wins (LWW)**

```
场景: 分布式数据库,用时间戳解决写冲突

节点 A 的时钟: 10:00:00.000 (实际慢了 1 秒)
节点 B 的时钟: 10:00:01.000 (正确)

客户端 1 → 节点 A: 写入 x=1, timestamp=10:00:00.000
客户端 2 → 节点 B: 写入 x=2, timestamp=10:00:01.000

结果: x=2 胜出 (因为时间戳更新)
实际: 客户端 1 可能是后写入的,但因为时钟慢,被覆盖!

数据丢失! ❌
```

**案例 2: 分布式锁**

```
场景: 使用租约(lease)实现分布式锁

客户端 1 获得锁,租约到 10:00:10
客户端 1 发生 GC 暂停 15 秒
客户端 1 以为锁还有效(本地时钟出错)
客户端 2 已经获得锁
两个客户端同时操作数据!

安全性违背! ❌
```

**案例 3: 有序事件排序**

```
场景: 用时间戳给事件排序

服务器 A: 事件 E1, timestamp=10:00:00.100
服务器 B: 事件 E2, timestamp=10:00:00.050 (时钟慢)

如果 E2 实际发生在 E1 之后,但时间戳更小
排序错误!
```

#### 时钟的正确用法

**✅ 可以依赖时钟的场景**:
- 日志时间戳(不需要绝对精确)
- 缓存过期时间(容忍一定误差)
- 监控指标时间

**❌ 不能依赖时钟的场景**:
- 分布式锁的正确性
- 事件的因果关系
- 跨节点的严格排序

**正确的替代方案**:
- **逻辑时钟** (Lamport Timestamp, Vector Clock)
- **自增 ID** (Snowflake)
- **共识算法** (Raft, Paxos)

---

### 4️⃣ 进程暂停

#### GC 导致的暂停

**问题**: 应用程序可能在任意时刻暂停

**案例: Java GC 暂停**
```
线程 A: 持有分布式锁
GC 暂停 15 秒 (Stop-the-World)
锁的租约过期
线程 B: 获得锁
线程 A: 恢复执行,以为自己还持有锁
两个线程同时操作数据! ❌
```

**其他导致暂停的原因**:
- 虚拟机被挂起(VM live migration)
- 操作系统上下文切换
- 磁盘 I/O 阻塞(page fault)
- SIGSTOP 信号
- 笔记本合盖(suspend)

**关键洞察**: 单线程代码在分布式环境中不再有时间保证

#### 响应时间保证

**硬实时系统 (Hard Real-Time)**:
- 航空电子、汽车安全系统
- 必须在截止时间内完成,否则灾难性后果
- 需要特殊硬件和 RTOS(实时操作系统)
- 不使用 GC、动态内存分配

**软实时系统 (Soft Real-Time)**:
- 视频会议、在线游戏
- 尽量保证低延迟,但偶尔违背可接受

**普通服务器系统**:
- 没有响应时间保证
- 追求高吞吐量 > 低延迟保证
- GC、虚拟化、分时多任务

**在分布式系统中的影响**:
- 租约、超时、故障检测都可能因暂停而误判
- 需要设计能容忍暂停的算法

---

### 5️⃣ 知识、真相与谎言

#### 真相由多数派定义

**问题**: 节点无法确定自己的状态

**案例: 脑裂 (Split Brain)**
```
集群: 节点 A(Leader), 节点 B, 节点 C

网络分区:
- 节点 A 无法连接 B、C
- 节点 B、C 可以互相连接

节点 A 的视角:
- 我是 Leader
- B 和 C 失效了

节点 B、C 的视角:
- Leader A 失效了
- 选举新 Leader

结果: 两个 Leader! ❌
```

**解决方案: Quorum (法定人数)**
```
集群 3 个节点,Quorum = 2

节点 A 被隔离:
- 无法联系到多数派(2 个节点)
- 主动放弃 Leader 身份

节点 B、C:
- 可以形成多数派
- 选举新 Leader

确保只有一个 Leader ✅
```

#### Fencing Token (隔离令牌)

**问题**: 旧 Leader 可能不知道自己已被替换

**解决方案: 单调递增的 Token**

```
1. 锁服务(如 Zookeeper)为每次锁授予分配递增的 Token
   客户端 1 获得锁, Token=33
   客户端 2 获得锁, Token=34

2. 客户端携带 Token 访问资源
   客户端 1: 写入请求, Token=33
   客户端 2: 写入请求, Token=34

3. 存储服务拒绝旧 Token 的请求
   收到 Token=33 的请求 → 拒绝(已经看到 Token=34)
   收到 Token=34 的请求 → 接受

即使客户端 1 GC 暂停后恢复,也无法破坏安全性 ✅
```

**关键点**:
- Token 必须单调递增
- 资源必须检查并拒绝旧 Token
- 类似于 Raft 的 term number

#### 拜占庭故障

**非拜占庭故障 (Crash-Stop Fault)**:
- 节点崩溃、网络故障、进程暂停
- 节点"诚实",不会说谎

**拜占庭故障 (Byzantine Fault)**:
- 节点可能发送虚假信息
- 恶意行为或软件 bug

**拜占庭将军问题**:
```
场景: 多个将军围攻城市,需要协调一致行动

挑战:
- 将军间通过信使传递消息
- 某些将军是叛徒,会发送虚假消息
- 如何让忠诚将军达成一致?

解决方案:
- 需要至少 3f+1 个将军才能容忍 f 个叛徒
- 拜占庭容错算法(PBFT、区块链共识)
```

**实际应用**:
- **区块链**: 节点可能是恶意的
- **航空航天**: 辐射可能导致位翻转
- **多组织系统**: 不同公司运行节点,互不信任

**大多数系统假设**:
- 节点是诚实的,但可能故障
- 使用更简单的算法(Raft、Paxos)

---

### 6️⃣ 系统模型与现实

#### 系统模型的三种类型

**1. 同步模型 (Synchronous Model)**
- 网络延迟有上界
- 进程暂停有上界
- 时钟误差有界

**特点**: 理想化,现实中很难实现

**2. 部分同步模型 (Partially Synchronous Model)**
- 大部分时间表现良好
- 偶尔超出界限

**特点**: 更接近现实(Raft、Paxos 假设这种模型)

**3. 异步模型 (Asynchronous Model)**
- 没有任何时间假设
- 没有时钟,没有超时

**特点**: 极端保守,但有理论价值

#### FLP 不可能性定理

**FLP Theorem** (Fischer, Lynch, Paterson, 1985):

> 在异步系统中,即使只有一个节点可能失效,也不存在确定性的共识算法

**含义**:
- 在最恶劣的情况下,共识算法可能永远无法终止
- 但实践中,通过超时、随机化可以解决(Raft、Paxos)

---

### 关键洞察

1. **分布式系统充满不确定性**
   - 网络不可靠
   - 时钟不可信
   - 进程会暂停

2. **无法区分慢和死**
   - 超时不能确定节点是否真的失效
   - 只能做出概率性判断

3. **真相由多数派定义**
   - 单个节点无法确定自己的状态
   - 需要 Quorum 机制

4. **时钟只能用于近似**
   - 不能依赖时钟做精确判断
   - 使用逻辑时钟、共识算法

5. **算法必须容忍部分失效**
   - 设计时假设最坏情况
   - 使用 Fencing Token 等机制保证安全性

---

### 延伸阅读

**经典论文**:
- [Time, Clocks, and the Ordering of Events in a Distributed System](https://lamport.azurewebsites.net/pubs/time-clocks.pdf) - Lamport, 1978 (逻辑时钟的开创性工作)
- [Impossibility of Distributed Consensus with One Faulty Process](https://groups.csail.mit.edu/tds/papers/Lynch/jacm85.pdf) - FLP, 1985
- [Unreliable Failure Detectors for Reliable Distributed Systems](https://www.cs.utexas.edu/~lorenzo/corsi/cs380d/papers/p225-chandra.pdf) - Chandra & Toueg, 1996

**技术文章**:
- [The Network is Reliable](https://aphyr.com/posts/288-the-network-is-reliable) - Aphyr (反讽标题,揭示网络的不可靠性)
- [There is No Now](https://queue.acm.org/detail.cfm?id=2745385) - 时钟在分布式系统中的问题
- [Jepsen: Testing the Partition Tolerance of Distributed Systems](https://aphyr.com/tags/jepsen) - Kyle Kingsbury 的分布式系统测试

**视频资源**:
- [Distributed Systems Lecture - Lamport Clocks](https://www.youtube.com/watch?v=x-D8iFU1d-o) - Martin Kleppmann
- [The Trouble with Timestamps](https://www.youtube.com/watch?v=BhosKsE8up8)

**相关书籍**:
- 《Time, Clocks, and the Ordering of Events》 - Lamport 的经典论文
- 《Distributed Systems: Principles and Paradigms》 - Tanenbaum

---

### 实践练习

**练习 1: 网络超时设计**

场景: 设计一个分布式 RPC 框架的超时机制

思考:
- 如何设置合理的超时时间?
- 如何区分慢响应和节点失效?
- 如何避免误判导致的雪崩?

参考方案:
- 动态超时(基于 P99 延迟)
- 重试机制(指数退避)
- 熔断器模式

**练习 2: 分布式锁安全性**

场景: 实现一个安全的分布式锁

要求:
- 容忍 GC 暂停
- 防止脑裂
- 使用 Fencing Token

实现步骤:
1. 使用 Zookeeper/etcd 管理锁
2. 每次授予锁时分配递增的 Token
3. 资源检查 Token,拒绝旧 Token 的请求

**练习 3: 时钟依赖分析**

分析以下场景是否安全:
1. 用本地时间戳作为日志文件名 → 安全 ✅
2. 用时间戳作为分布式事务 ID → 不安全 ❌
3. 用时间戳判断缓存是否过期 → 基本安全 ✅
4. 用时间戳排序分布式事件 → 不安全 ❌

**练习 4: Quorum 计算**

给定集群规模 N,计算:
- 写 Quorum: W
- 读 Quorum: R
- 容错数: F

条件: W + R > N, F = floor((N-1)/2)

| N | W | R | F |
|---|---|---|---|
| 3 | 2 | 2 | 1 |
| 5 | 3 | 3 | 2 |
| 7 | 4 | 4 | 3 |

---

### 我的思考

**问题**:
- ❓ 为什么 Google Spanner 敢使用时间戳(TrueTime)?
- ❓ 如果时钟不可信,如何实现全局有序的事件日志?
- ❓ 拜占庭容错算法的性能为什么这么差?

**答案要点**:
- **TrueTime**: Google 使用原子钟 + GPS,误差极小(±7ms),且 API 返回时间区间而非单一时间点
- **全局有序**: 使用逻辑时钟(Lamport Timestamp)或共识算法(Raft),而非物理时间
- **拜占庭容错**: 需要 3f+1 个节点容忍 f 个恶意节点,消息复杂度 O(n²),性能差但安全性高

**与 CAP 定理的联系**:
- 网络分区(P)是本章的核心
- 当分区发生时,必须在一致性(C)和可用性(A)之间选择
- 时钟不可靠导致无法用时间戳实现一致性

**在系统设计中的应用**:
- 设计分布式锁时,必须使用 Fencing Token
- 设计分布式事务时,不能依赖时钟排序
- 设计故障检测时,使用自适应超时和 Quorum

---

### Chapter 9: 一致性与共识

**阅读状态**: ⏳ 待学习

#### 预习大纲

**核心主题**:
1. 一致性保证
   - 线性一致性
   - 因果一致性
   - 最终一致性
2. 顺序保证
   - 全序广播
   - 时间戳排序
3. 分布式事务与共识
   - 两阶段提交 (2PC)
   - 容错共识算法 (Paxos, Raft)

**关键概念**:
- Linearizability
- Total Order Broadcast
- Consensus

#### 延伸资料 (待整理)

**论文**:
- [Paxos Made Simple](https://lamport.azurewebsites.net/pubs/paxos-simple.pdf) - Lamport, 2001
- [In Search of an Understandable Consensus Algorithm (Raft)](https://raft.github.io/raft.pdf) - Ongaro & Ousterhout, 2014

**文章**:
- [Raft Visualization](https://raft.github.io/)
- [Paxos vs Raft](https://www.quora.com/What-is-the-difference-between-Paxos-and-Raft-consensus-algorithms)

---

## 第三部分: 派生数据

### Chapter 10: 批处理

**阅读状态**: ⏳ 待学习

#### 预习大纲

**核心主题**:
1. MapReduce 与分布式文件系统
2. MapReduce 工作流
3. 超越 MapReduce
   - Spark, Flink

**关键概念**:
- HDFS, Map, Reduce, Shuffle
- Dataflow Engines

#### 延伸资料 (待整理)

**论文**:
- [MapReduce: Simplified Data Processing](https://static.googleusercontent.com/media/research.google.com/en//archive/mapreduce-osdi04.pdf) - Google, 2004
- [The Google File System](https://static.googleusercontent.com/media/research.google.com/en//archive/gfs-sosp2003.pdf) - Google, 2003

---

### Chapter 11: 流处理

**阅读状态**: ⏳ 待学习

#### 预习大纲

**核心主题**:
1. 消息系统
   - 消息代理 vs 日志
2. 数据库与流
   - 变更数据捕获 (CDC)
   - 事件溯源
3. 流处理框架
   - Kafka Streams, Apache Flink

**关键概念**:
- Event Sourcing
- CQRS
- Exactly-Once Semantics

#### 延伸资料 (待整理)

**论文**:
- [Kafka: a Distributed Messaging System](https://www.microsoft.com/en-us/research/wp-content/uploads/2017/09/Kafka.pdf)

**文章**:
- [Event Sourcing Pattern](https://martinfowler.com/eaaDev/EventSourcing.html) - Martin Fowler

---

### Chapter 12: 数据系统的未来

**阅读状态**: ⏳ 待学习

#### 预习大纲

**核心主题**:
1. 数据集成
2. 解耦批处理与流处理
3. 正确性保证
4. 数据隐私与伦理

---

## 📖 阅读计划

### Week 1-3: 基础部分
- [x] Chapter 1: 可靠性、可扩展性、可维护性 ✅ 2025-10-16
- [ ] Chapter 2: 数据模型与查询语言 (计划: Week 2)
- [ ] Chapter 3: 存储与检索 (计划: Week 2)
- [ ] Chapter 4: 编码与演化 (计划: Week 3)

### Week 4-6: 分布式数据
- [ ] Chapter 5: 复制 (计划: Week 2-3)
- [ ] Chapter 6: 分区 (计划: Week 2)
- [ ] Chapter 7: 事务 (计划: Week 2)
- [ ] Chapter 8: 分布式系统的麻烦 (计划: Week 8)
- [ ] Chapter 9: 一致性与共识 (计划: Week 8)

### Week 7-8: 派生数据
- [ ] Chapter 10: 批处理 (可选)
- [ ] Chapter 11: 流处理 (计划: Week 6)
- [ ] Chapter 12: 数据系统的未来 (可选)

---

## 💡 学习方法

### 阅读策略

1. **首次阅读**: 快速浏览,理解章节结构
2. **深入阅读**: 仔细阅读,做标注和笔记
3. **实践验证**: 通过代码或实验验证概念
4. **总结输出**: 用自己的话总结核心要点

### 笔记模板

每章完成后,记录:
- ✅ **核心概念**: 3-5 个关键概念
- ✅ **经典案例**: 书中的真实系统案例
- ✅ **关键洞察**: 最重要的 takeaway
- ✅ **延伸阅读**: 论文、文章、视频链接
- ✅ **实践练习**: 动手验证的实验或代码
- ✅ **我的思考**: 问题和思考

### 与课程结合

**DDIA 章节对应课程模块**:

| DDIA 章节 | 课程模块 |
|----------|---------|
| Chapter 1 | Week 1 - 可扩展性基础 |
| Chapter 3 | Week 2 - 存储系统原理 |
| Chapter 5 | Week 2 - 分布式存储与一致性 |
| Chapter 6 | Week 2 - 数据分片, Week 5 - 数据库分片 |
| Chapter 7 | Week 2 - ACID 实现, Week 5 - 分布式事务 |
| Chapter 9 | Week 8 - 一致性算法 |
| Chapter 11 | Week 3 - 消息队列, Week 6 - 事件驱动 |

---

## 🔗 全书相关资源

### 官方资源

- [DDIA 官网](https://dataintensive.net/)
- [作者 Martin Kleppmann 博客](https://martin.kleppmann.com/)
- [勘误表](https://github.com/ept/ddia-references)

### 配套课程

- [Martin Kleppmann's Distributed Systems Lecture](https://www.youtube.com/playlist?list=PLeKd45zvjcDFUEv_ohr_HdUFe97RItdiB)
- [MIT 6.824: Distributed Systems](https://pdos.csail.mit.edu/6.824/)

### 中文资源

- 中文翻译版: 《数据密集型应用系统设计》
- [ddia-cn (中文版在线阅读)](https://github.com/Vonng/ddia)

### 讨论社区

- [r/ddia on Reddit](https://www.reddit.com/r/ddia/)
- [DDIA Book Club](https://github.com/ept/ddia-references/issues)

---

## 📊 阅读进度统计

**总进度**: 1/12 章节 (8.3%)

**Part I**: 1/4 章节完成
**Part II**: 0/5 章节完成
**Part III**: 0/3 章节完成

**总学习时长**: _____ 小时
**笔记总字数**: _____ 字

---

## 💭 整体思考

### 为什么读 DDIA?

1. **系统设计的理论基础**: 理解分布式系统的核心概念
2. **工程实践的指导**: 真实案例和权衡分析
3. **面试准备**: 系统设计面试的必备知识
4. **技术选型**: 为项目选择合适的技术栈

### 最大收获 (持续更新)

1. **Twitter 扇出案例**: 理解了读写权衡的本质
2. **P99 延迟的重要性**: 性能指标应该关注长尾
3. **没有银弹**: 系统设计取决于负载特征

### 待深入的主题

- [ ] CAP 定理与 PACELC 的关系
- [ ] 分布式事务的各种实现方案对比
- [ ] 一致性模型的谱系
- [ ] 流处理与批处理的统一

---

**更新日期**: 2025-10-16
**下次更新**: 开始 Chapter 2 时更新

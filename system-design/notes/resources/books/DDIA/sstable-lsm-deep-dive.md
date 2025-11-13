# SSTable 与 LSM-Tree 深度解析

> 从哈希索引到 SSTable，理解现代数据库存储引擎的进化之路

## 目录

- [为什么需要 SSTable？](#为什么需要-sstable)
- [SSTable 核心概念](#sstable-核心概念)
- [LSM-Tree 完整架构](#lsm-tree-完整架构)
- [写入路径详解](#写入路径详解)
- [读取路径详解](#读取路径详解)
- [压缩策略深度对比](#压缩策略深度对比)
- [性能优化技术](#性能优化技术)
- [真实系统案例](#真实系统案例)

---

## 为什么需要 SSTable？

### 哈希索引的根本缺陷

回顾我们之前的哈希索引数据库（Bitcask 模型）：

```
内存哈希表：
key123 → offset=1024  (segment-0.log)
key456 → offset=2048  (segment-0.log)
key789 → offset=3072  (segment-1.log)

磁盘段文件（无序追加）：
segment-0.log: key123,value1 | key456,value2 | ...
segment-1.log: key789,value3 | ...
```

**问题 1：内存限制**

```
假设每个键平均 20 字节 + 文件偏移量 16 字节 = 36 字节/键
1 亿个键 × 36 字节 = 3.6 GB 内存
10 亿个键 = 36 GB 内存  ← 无法接受！
```

**问题 2：无范围查询**

```sql
-- ❌ 哈希索引无法高效支持
SELECT * FROM users WHERE age BETWEEN 20 AND 30;
SELECT * FROM logs WHERE timestamp >= '2024-01-01' AND timestamp < '2024-02-01';

-- 只能支持点查询
SELECT * FROM users WHERE id = 123;  ✅
```

**问题 3：压缩效率低**

```
合并无序段时：
段1: [c=3, a=1, b=2, d=4]
段2: [a=5, e=6, b=7]

需要：
1. 将所有键值对读入内存
2. 去重（保留最新值）
3. 写入新段

时间复杂度：O(n)，空间复杂度：O(n)
无法流式处理大文件！
```

### SSTable 的核心改进

**SSTable = Sorted String Table（有序字符串表）**

**唯一的改变**：键在段内**有序存储**

```
无序段（哈希索引）：
[key123=v1, key789=v2, key456=v3, key999=v4]

SSTable（有序段）：
[key123=v1, key456=v3, key789=v2, key999=v4]
       ↑ 按键排序
```

**这个简单的改变带来了革命性的优势！**

---

## SSTable 核心概念

### 优势 1：稀疏索引（解决内存问题）

#### 传统哈希索引

```
必须为每个键维护索引：

内存索引（100万键）：
key000001 → offset=0
key000002 → offset=100
key000003 → offset=200
...
key999999 → offset=99999900

内存占用：100万 × 36字节 = 36MB
```

#### SSTable 稀疏索引

```
只需为每隔N个键（或每N KB）维护一个索引点：

SSTable 文件（有序）：
0KB:    key000001=v1
        key000002=v2
        ...
4KB:    key000100=v100  ← 索引点
        key000101=v101
        ...
8KB:    key000200=v200  ← 索引点
        ...

内存稀疏索引：
0KB  → key000001
4KB  → key000100
8KB  → key000200
12KB → key000300
...

内存占用：(1MB文件 / 4KB) × 36字节 = 9KB  ← 减少 99% 内存！
```

#### 查找过程

```
查询 key000150：

1. 在内存稀疏索引中二分查找：
   key000100 < key000150 < key000200
   ↓
   确定在 4KB-8KB 块中

2. 读取 4KB-8KB 的数据块到内存
   [key000100...key000150...key000199]

3. 在数据块内二分查找 key000150
   ↓
   返回 value

磁盘 I/O：1 次（读取 4KB 块）
时间复杂度：O(log n)
```

**关键洞察**：

- 哈希索引：O(1) 查找，但需要全部键在内存
- SSTable：O(log n) 查找，但只需少量索引点
- 对于海量数据，**能存下** > **速度快一点点**

---

### 优势 2：高效合并（归并排序）

#### 无序段合并（哈希索引）

```
段1（无序）: [c=3, a=1, b=2]
段2（无序）: [a=5, b=6, d=7]

合并过程：
1. 读取所有键值对到内存：[c=3, a=1, b=2, a=5, b=6, d=7]
2. 去重（HashMap）: {a:5, b:6, c:3, d:7}
3. 写入新段

问题：
- 必须全部加载到内存（内存限制！）
- 无法流式处理
```

#### SSTable 归并（类似归并排序）

```
段1（有序）: [a=1, b=2, c=3]
段2（有序）: [a=5, b=6, d=7]

归并过程（双指针）：

指针1 → a=1     指针2 → a=5
         ↓               ↓
比较：a == a，保留新段的值 → 输出 a=5

指针1 → b=2     指针2 → b=6
         ↓               ↓
比较：b == b，保留新段的值 → 输出 b=6

指针1 → c=3     指针2 → d=7
         ↓               ↓
比较：c < d → 输出 c=3

指针1 → 结束    指针2 → d=7
                         ↓
输出剩余 → 输出 d=7

结果：[a=5, b=6, c=3, d=7]
```

**伪代码**：

```python
def merge_sstables(sstable1, sstable2):
    """
    归并两个 SSTable，类似归并排序
    关键：流式处理，无需全部加载到内存
    """
    iter1 = sstable1.iterator()  # 顺序读取
    iter2 = sstable2.iterator()

    kv1 = iter1.next()
    kv2 = iter2.next()

    output = new_sstable()

    while kv1 is not None and kv2 is not None:
        if kv1.key < kv2.key:
            output.write(kv1)
            kv1 = iter1.next()
        elif kv1.key > kv2.key:
            output.write(kv2)
            kv2 = iter2.next()
        else:  # kv1.key == kv2.key
            # 保留新段的值（sstable2 更新）
            output.write(kv2)
            kv1 = iter1.next()
            kv2 = iter2.next()

    # 写入剩余元素
    while kv1 is not None:
        output.write(kv1)
        kv1 = iter1.next()

    while kv2 is not None:
        output.write(kv2)
        kv2 = iter2.next()

    return output
```

**优势**：

- ✅ **流式处理**：不需要全部加载到内存
- ✅ **时间复杂度**：O(n)，线性扫描
- ✅ **空间复杂度**：O(1)，只需读写缓冲区
- ✅ **可扩展**：能处理 GB 级别的段文件

---

### 优势 3：范围查询支持

```
SSTable（有序）：
[age=18, age=20, age=25, age=30, age=35, age=40]

查询：age BETWEEN 20 AND 35

1. 二分查找定位到 age=20
2. 顺序扫描：age=20, age=25, age=30, age=35
3. 遇到 age=40 > 35，停止

时间复杂度：O(log n + k)，k 为结果数量
```

**对比哈希索引**：

```
哈希索引（无序）：
必须扫描所有键，检查每个键是否在范围内
时间复杂度：O(n)  ← 无法接受！
```

---

### 优势 4：块压缩

```
SSTable 结构：
┌─────────────┬─────────────┬─────────────┐
│  Block 0    │  Block 1    │  Block 2    │
│  (4KB)      │  (4KB)      │  (4KB)      │
│  未压缩     │  未压缩     │  未压缩     │
└─────────────┴─────────────┴─────────────┘

优化后（块压缩）：
┌─────────────┬─────────────┬─────────────┐
│  Block 0    │  Block 1    │  Block 2    │
│  (4KB)      │  (4KB)      │  (4KB)      │
│  ↓ Snappy   │  ↓ Snappy   │  ↓ Snappy   │
│  2KB        │  1.8KB      │  2.2KB      │
└─────────────┴─────────────┴─────────────┘

内存稀疏索引：
0KB  → Block 0 (compressed 2KB)
2KB  → Block 1 (compressed 1.8KB)
4KB  → Block 2 (compressed 2.2KB)
```

**读取流程**：

```
查询 key150（在 Block 1）：
1. 稀疏索引定位 → Block 1
2. 读取压缩块（1.8KB）
3. 解压到内存（4KB）
4. 二分查找 key150

好处：
- 减少磁盘 I/O（读 1.8KB 而不是 4KB）
- 减少网络传输（分布式系统）
- 缓存更多数据（压缩后占用更少内存）
```

**常用压缩算法**：

| 算法   | 压缩比 | 压缩速度 | 解压速度 | 适用场景         |
| ------ | ------ | -------- | -------- | ---------------- |
| Snappy | 2-3x   | 极快     | 极快     | RocksDB, LevelDB |
| LZ4    | 2-3x   | 极快     | 极快     | 通用             |
| Zstd   | 3-5x   | 快       | 快       | 高压缩比需求     |
| Gzip   | 5-10x  | 慢       | 中等     | 存档场景         |

**RocksDB/LevelDB 选择 Snappy 的原因**：

- 解压速度 > 500 MB/s（远快于磁盘 I/O）
- 压缩开销可忽略不计
- 压缩比足够好（2-3x）

---

## LSM-Tree 完整架构

### 问题：如何构建有序的 SSTable？

**矛盾**：

- 写入需要追加（顺序 I/O，高性能）
- SSTable 需要有序（不能边追加边排序）

**解决方案：两阶段写入**

```
写入 → MemTable（内存，自动排序）→ 刷盘 → SSTable（磁盘，有序不可变）
```

---

### LSM-Tree 分层架构

```
┌─────────────────────────────────────────────────────────┐
│                    Write Path                           │
├─────────────────────────────────────────────────────────┤
│  1. Write to WAL (Write-Ahead Log)                      │
│     ↓                                                    │
│  2. Write to MemTable (in-memory sorted structure)      │
│     ├─ Red-Black Tree / AVL Tree / Skip List            │
│     └─ Auto-sorted on insert                            │
├─────────────────────────────────────────────────────────┤
│                    Flush to Disk                        │
├─────────────────────────────────────────────────────────┤
│  3. MemTable Full (e.g., 4MB) → Freeze                  │
│     ↓                                                    │
│  4. Background Thread Flush to SSTable (Level 0)        │
│     ├─ Write sorted KV pairs to disk                    │
│     └─ Immutable, never modified                        │
├─────────────────────────────────────────────────────────┤
│                    Read Path                            │
├─────────────────────────────────────────────────────────┤
│  5. Query Key:                                          │
│     MemTable → Level 0 → Level 1 → ... → Level N        │
│     ├─ Return first match (newest)                      │
│     └─ Bloom Filter optimization                        │
├─────────────────────────────────────────────────────────┤
│                    Compaction                           │
├─────────────────────────────────────────────────────────┤
│  6. Background Compaction Thread:                       │
│     ├─ Merge SSTables (remove old versions)             │
│     ├─ Remove tombstones (deleted keys)                 │
│     └─ Keep data sorted                                 │
└─────────────────────────────────────────────────────────┘
```

---

### 核心组件详解

#### 1. MemTable（内存表）

**作用**：内存中的可变、有序数据结构

**常用实现**：

```
跳表（Skip List）- LevelDB, RocksDB 默认
├─ 插入：O(log n)
├─ 查找：O(log n)
├─ 范围查询：O(log n + k)
└─ 实现简单，无需旋转操作

红黑树（Red-Black Tree）
├─ 插入：O(log n)
├─ 查找：O(log n)
└─ 需要旋转操作，实现复杂

AVL 树
├─ 查找更快（更严格平衡）
└─ 插入慢（需要更多旋转）
```

**跳表示例**：

```
Level 3:  1 ────────────────────────────> 25
           ↓                               ↓
Level 2:  1 ────────> 10 ────────────────> 25
           ↓           ↓                   ↓
Level 1:  1 ──> 5 ──> 10 ──> 15 ──> 20 ──> 25
           ↓     ↓     ↓      ↓      ↓      ↓
Level 0:  1->3->5->7->10->12->15->18->20->23->25

插入 key=12:
1. 从顶层开始：1 < 12 < 25
2. 下降到 Level 2：1 < 12，走到 10；10 < 12 < 25
3. 下降到 Level 1：10 < 12 < 15
4. 下降到 Level 0：插入 12

查找 key=15:
Level 3: 1 → 25（跳过中间）
Level 2: 1 → 10 → 25
Level 1: 10 → 15（找到！）

跳跃查找，O(log n) 时间复杂度
```

**为什么 LevelDB 选择跳表？**

1. **实现简单**：无需复杂的旋转操作
2. **并发友好**：更容易实现无锁版本
3. **内存局部性**：节点可独立分配
4. **性能足够**：O(log n)，与红黑树相当

---

#### 2. WAL（Write-Ahead Log）

**作用**：崩溃恢复保障

```
写入流程（原子性保证）：

1. 写入 WAL（顺序追加）
   wal.log: "PUT key=123 value=London timestamp=1234567890"
   ↓
   fsync() 确保持久化到磁盘
   ↓
2. 写入 MemTable
   memtable.put("123", "London")
   ↓
3. 返回成功

崩溃恢复：
1. 读取 WAL 文件
2. 重放所有操作到 MemTable
3. 恢复到崩溃前状态
```

**WAL 格式示例**：

```
WAL Record Format:
┌──────────┬───────────┬────────┬──────────┬─────────┬───────┐
│ CRC      │ Seq Number│ Type   │ Key Size │ Val Size│ Key   │
│ (4 byte) │ (8 byte)  │(1 byte)│ (4 byte) │ (4 byte)│ (var) │
└──────────┴───────────┴────────┴──────────┴─────────┴───────┘
└────────────────────────────────────────────────────────────┘
                          Value (var)

Type:
- 0x01: PUT
- 0x02: DELETE (tombstone)
```

**性能权衡**：

```
写入 WAL 后立即 fsync():
✅ 崩溃安全（不丢数据）
❌ 写入延迟高（~10ms/次）

写入 WAL 但延迟 fsync():
✅ 写入延迟低（~0.1ms/次）
❌ 可能丢失最近 1 秒的数据

RocksDB 配置项：
options.sync = true   // 每次写入都 fsync（安全）
options.sync = false  // 后台定期 fsync（快速）
```

---

#### 3. Immutable MemTable

**作用**：冻结的只读 MemTable，等待刷盘

```
写入流程：

MemTable (可写) [3MB]
    ↓ 写入数据
MemTable (可写) [4MB] ← 达到阈值！
    ↓
Immutable MemTable (只读) [4MB] ← 冻结
    +
MemTable (新建，可写) [0MB]
    ↓ 后台线程
刷盘为 SSTable

好处：
1. 写入不阻塞（新 MemTable 继续接受写入）
2. 刷盘在后台进行
3. 读取可以查询 Immutable MemTable
```

---

#### 4. SSTable 文件结构

**完整 SSTable 文件格式**：

```
SSTable File Layout:
┌─────────────────────────────────────┐
│         Data Block 0                │  ← 实际键值对（压缩）
│  [key1=val1, key2=val2, ...]        │
├─────────────────────────────────────┤
│         Data Block 1                │
│  [key100=val100, ...]               │
├─────────────────────────────────────┤
│         ...                         │
├─────────────────────────────────────┤
│         Data Block N                │
├─────────────────────────────────────┤
│         Meta Block                  │  ← Bloom Filter, 统计信息
│  - Bloom Filter                     │
│  - Stats (min/max key, count)       │
├─────────────────────────────────────┤
│         Index Block                 │  ← 稀疏索引
│  Block 0 → offset=0, size=4096      │
│  Block 1 → offset=4096, size=3890   │
│  ...                                │
├─────────────────────────────────────┤
│         Footer                      │  ← 指向索引和元数据的指针
│  - Meta Block Handle                │
│  - Index Block Handle               │
│  - Magic Number (验证文件完整性)    │
└─────────────────────────────────────┘
```

**读取流程**：

```
打开 SSTable 文件：
1. 读取 Footer（固定位置，文件末尾 48 字节）
2. 根据 Footer 中的指针读取 Index Block
3. 加载 Index Block 到内存（稀疏索引）
4. 可选：加载 Bloom Filter 到内存

查询 key=150：
1. 检查 Bloom Filter → 可能存在
2. 在 Index Block 中二分查找 → 定位到 Block 1
3. 读取 Block 1 到内存（可能需要解压）
4. 在 Block 1 内二分查找 → 找到 key=150
5. 返回 value
```

---

## 写入路径详解

### 完整写入流程

```
PUT key=user123 value={"name":"Alice","age":30}

┌─────────────────────────────────────────────────────────┐
│ Step 1: Write to WAL                                    │
├─────────────────────────────────────────────────────────┤
│  wal.append("PUT user123 {name:Alice,age:30} ts=...")   │
│  wal.fsync()  ← 可选，取决于配置                         │
│  Time: ~0.1ms (no fsync) or ~10ms (with fsync)          │
└─────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────┐
│ Step 2: Write to MemTable                               │
├─────────────────────────────────────────────────────────┤
│  memtable.put("user123", "{name:Alice,age:30}")         │
│  Time: ~0.01ms (in-memory operation)                    │
└─────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────┐
│ Step 3: Check MemTable Size                             │
├─────────────────────────────────────────────────────────┤
│  if memtable.size >= threshold (e.g., 4MB):             │
│    freeze_memtable()                                    │
│    trigger_background_flush()                           │
└─────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────┐
│ Step 4: Return Success to Client                        │
└─────────────────────────────────────────────────────────┘

总延迟：~0.1ms (无 fsync) 或 ~10ms (有 fsync)
```

### 后台刷盘流程

```
Background Flush Thread:

1. 从 Immutable MemTable 中迭代读取（已排序）
   iterator.next() → (key1, val1)
   iterator.next() → (key2, val2)
   ...

2. 写入新的 SSTable 文件
   sstable_writer.add(key1, val1)
   sstable_writer.add(key2, val2)
   ...

3. 每写入 4KB，创建一个 Data Block
   - 压缩 Block（Snappy）
   - 记录索引项

4. 写入 Meta Block（Bloom Filter）

5. 写入 Index Block

6. 写入 Footer

7. fsync() 确保持久化

8. 更新元数据：
   - 添加 SSTable 到 Level 0
   - 删除对应的 WAL 文件
   - 释放 Immutable MemTable

9. 可能触发压缩（如果 Level 0 文件过多）
```

### 写放大分析

```
用户写入 100 字节：

1. WAL: 100 字节
2. MemTable: 100 字节（内存，不算磁盘写）
3. 刷盘 SSTable: 假设 MemTable 4MB，写入 4MB
4. 后续压缩（Leveled Compaction）:
   - Level 0 → Level 1: 4MB 写入
   - Level 1 → Level 2: 可能 10 倍（40MB）
   - Level 2 → Level 3: 可能 10 倍（400MB）

写放大 = 实际写入磁盘的数据 / 用户写入的数据
      ≈ 10-100 倍（取决于压缩策略）

但关键：顺序写，不是随机写！
HDD: 顺序写 100MB/s vs 随机写 1MB/s (100x 差异)
```

---

## 读取路径详解

### 完整读取流程

```
GET key=user123

┌─────────────────────────────────────────────────────────┐
│ Step 1: Check MemTable                                  │
├─────────────────────────────────────────────────────────┤
│  value = memtable.get("user123")                        │
│  if value != NULL:                                      │
│    return value  ← 最新数据在这里！                      │
│  Time: ~0.01ms                                          │
└─────────────────────────────────────────────────────────┘
         ↓ (not found)
┌─────────────────────────────────────────────────────────┐
│ Step 2: Check Immutable MemTable                        │
├─────────────────────────────────────────────────────────┤
│  value = immutable_memtable.get("user123")              │
│  if value != NULL:                                      │
│    return value                                         │
│  Time: ~0.01ms                                          │
└─────────────────────────────────────────────────────────┘
         ↓ (not found)
┌─────────────────────────────────────────────────────────┐
│ Step 3: Check Level 0 SSTables (newest to oldest)       │
├─────────────────────────────────────────────────────────┤
│  for sstable in level0_sstables.reverse():              │
│    # Bloom Filter 快速判断                              │
│    if !sstable.bloom_filter.may_contain("user123"):     │
│      continue  ← 跳过，一定不存在                        │
│                                                         │
│    # 在索引中查找                                        │
│    block_handle = sstable.index.find("user123")         │
│    if block_handle == NULL:                             │
│      continue                                           │
│                                                         │
│    # 读取数据块                                          │
│    block = read_block(block_handle)  ← 磁盘 I/O         │
│    decompress(block)                                    │
│    value = block.binary_search("user123")               │
│    if value != NULL:                                    │
│      return value                                       │
│                                                         │
│  Time: 取决于 Level 0 文件数量，最坏 ~N × 1ms            │
└─────────────────────────────────────────────────────────┘
         ↓ (not found)
┌─────────────────────────────────────────────────────────┐
│ Step 4: Check Level 1, 2, 3... (逐层查找)                │
├─────────────────────────────────────────────────────────┤
│  对于 Leveled Compaction:                               │
│    每层最多查找 1 个 SSTable（键范围不重叠）             │
│                                                         │
│  Time: 最坏 ~层数 × 1ms                                 │
└─────────────────────────────────────────────────────────┘
         ↓ (not found)
┌─────────────────────────────────────────────────────────┐
│ Step 5: Return Not Found                                │
└─────────────────────────────────────────────────────────┘

总延迟：
- 最好情况（在 MemTable）: ~0.01ms
- 中等情况（在 Level 0）: ~1-5ms
- 最坏情况（在 Level 3）: ~4-10ms
```

### Bloom Filter 优化

**问题**：查询不存在的键需要扫描所有 SSTable

```
查询 key=nonexistent（不存在）:

无 Bloom Filter:
1. 检查 MemTable ❌
2. 检查 Level 0 (4 个 SSTable) → 4 次磁盘 I/O ❌
3. 检查 Level 1 (10 个 SSTable) → 1 次磁盘 I/O ❌
4. 检查 Level 2 (100 个 SSTable) → 1 次磁盘 I/O ❌
5. 返回 Not Found

总计：6 次磁盘 I/O，~6ms
```

**Bloom Filter 原理**：

```
位数组 + 多个哈希函数

创建 Bloom Filter（为 SSTable 的所有键）:
keys = ["user123", "user456", "user789"]

bit_array = [0] * 1000  # 1000 位

for key in keys:
    h1 = hash1(key) % 1000  # 假设 = 123
    h2 = hash2(key) % 1000  # 假设 = 456
    h3 = hash3(key) % 1000  # 假设 = 789

    bit_array[h1] = 1
    bit_array[h2] = 1
    bit_array[h3] = 1

查询 key=nonexistent:
h1 = hash1("nonexistent") % 1000 = 234
h2 = hash2("nonexistent") % 1000 = 567
h3 = hash3("nonexistent") % 1000 = 890

if bit_array[234] == 0:  ← 有一个为 0
    return "一定不存在"  ← 无需读磁盘！

if bit_array[234] == 1 and bit_array[567] == 1 and bit_array[890] == 1:
    return "可能存在"  ← 需要读磁盘确认（假阳性）
```

**假阳性率**：

```
位数组大小 m = 10,000 位
键数量 n = 1,000
哈希函数数量 k = 3

假阳性率 ≈ (1 - e^(-kn/m))^k
        ≈ (1 - e^(-3×1000/10000))^3
        ≈ 0.008
        ≈ 0.8%

意味着：
- 99.2% 的不存在的键可以立即判断（无磁盘 I/O）
- 0.8% 的不存在的键会误判为"可能存在"（需要读磁盘）
```

**内存占用**：

```
推荐配置：每个键 10 位
1 百万键 × 10 位 = 10 Mb = 1.25 MB

极小的内存代价，巨大的性能提升！
```

---

## 压缩策略深度对比

### Size-Tiered Compaction Strategy (STCS)

**策略**：当有 N 个相似大小的 SSTable 时，合并它们

```
Time 0:
Level 0: [1MB] [1MB] [1MB] [1MB]
         ↓ 触发压缩（4 个 1MB 文件）
Time 1:
Level 0: []
Level 1: [4MB]

Time 2:
Level 0: [1MB] [1MB] [1MB] [1MB]
Level 1: [4MB] [4MB] [4MB] [4MB]
         ↓ Level 1 触发压缩（4 个 4MB 文件）
Time 3:
Level 0: [1MB] [1MB] [1MB] [1MB]
Level 1: []
Level 2: [16MB]
```

**合并过程**：

```
输入：4 个 1MB 文件
文件1: [a=1, b=2, c=3, ...]
文件2: [a=5, d=6, e=7, ...]
文件3: [b=8, f=9, ...]
文件4: [a=10, g=11, ...]

输出：1 个 4MB 文件（去重）
[a=10, b=8, c=3, d=6, e=7, f=9, g=11, ...]
 ↑最新值
```

**优点**：

```
✅ 写放大低
   每个 SSTable 平均合并次数 = log₄(总大小/SSTable大小)
   例如：1GB 数据，1MB SSTable → log₄(1000) ≈ 5 次合并

✅ 写吞吐量高
   后台压缩不频繁
```

**缺点**：

```
❌ 读放大高
   查询可能需要扫描每层的多个文件
   Level 0: 4 个文件
   Level 1: 4 个文件
   Level 2: 4 个文件
   最坏情况：12 次磁盘 I/O

❌ 空间放大高
   旧版本数据保留时间长
   压缩前：4 个 4MB 文件 = 16MB
   压缩中：16MB（旧）+ 16MB（新）= 32MB
   空间放大 2 倍
```

**适用场景**：

- 写密集型工作负载
- 时序数据（旧数据很少查询）
- 日志系统

**使用 STCS 的系统**：

- Apache Cassandra (默认)
- Apache HBase
- ScyllaDB

---

### Leveled Compaction Strategy (LCS)

**策略**：每层有容量限制，层内 SSTable 键范围不重叠

```
Level 0: (可重叠，直接从 MemTable 刷盘)
[SSTable-1: a-z, 2MB]
[SSTable-2: c-m, 2MB]
[SSTable-3: f-t, 2MB]
[SSTable-4: k-z, 2MB]

Level 1: (不重叠，容量 10MB)
[SSTable-1: a-c, 1MB]
[SSTable-2: d-f, 1MB]
[SSTable-3: g-j, 1MB]
[SSTable-4: k-m, 1MB]
[SSTable-5: n-p, 1MB]
[SSTable-6: q-s, 1MB]
[SSTable-7: t-v, 1MB]
[SSTable-8: w-z, 1MB]

Level 2: (不重叠，容量 100MB)
[SSTable-1: a-b, 10MB]
[SSTable-2: c-d, 10MB]
...
[SSTable-10: y-z, 10MB]
```

**合并触发条件**：

```
Level 0 达到 4 个文件（阈值）:
1. 选择 Level 0 的所有文件
   [SSTable-1: a-z]
   [SSTable-2: c-m]
   [SSTable-3: f-t]
   [SSTable-4: k-z]

2. 找出键范围：min=a, max=z

3. 选择 Level 1 中重叠的文件
   Level 1: 所有文件都重叠（a-z 覆盖所有）
   [SSTable-1: a-c]
   [SSTable-2: d-f]
   ...
   [SSTable-8: w-z]

4. 归并排序，生成新的 Level 1 文件
   输出：[a-c, 1MB] [d-f, 1MB] ... [w-z, 1MB]
```

**关键特性：层内不重叠**

```
查询 key=e:

Level 0: 可能在任何文件（需要检查 Bloom Filter）
  ├─ SSTable-1: Bloom Filter 检查 ❌
  ├─ SSTable-2: Bloom Filter 检查 ✅ → 读取
  ├─ SSTable-3: Bloom Filter 检查 ❌
  └─ SSTable-4: Bloom Filter 检查 ❌

Level 1: 键范围不重叠，只可能在一个文件
  ├─ SSTable-1: a-c ❌
  ├─ SSTable-2: d-f ✅ ← 只需读这一个！
  └─ ...

Level 2: 同理，只读一个文件

总磁盘 I/O：最多 Level 0 文件数 + 层数
            = 4 + 3 = 7 次（最坏情况）
```

**优点**：

```
✅ 读放大低
   每层最多读 1 个 SSTable（除了 Level 0）

✅ 空间放大低
   旧版本数据快速被压缩清理

✅ 范围查询友好
   层内有序且不重叠，可快速定位
```

**缺点**：

```
❌ 写放大高
   Level 0 → Level 1: 合并所有重叠文件
   Level 1 → Level 2: 可能需要合并 10 倍数据

   示例：
   写入 1MB 到 Level 0
   → Level 0 → Level 1: 合并 10MB
   → Level 1 → Level 2: 合并 100MB
   → Level 2 → Level 3: 合并 1000MB

   写放大 = (1 + 10 + 100 + 1000) / 1 ≈ 1000x

❌ 写吞吐量相对低
   后台压缩频繁，占用 I/O 带宽
```

**适用场景**：

- 读密集型工作负载
- 需要低延迟点查询
- 需要高效范围查询

**使用 LCS 的系统**：

- LevelDB
- RocksDB (默认)
- Cassandra (可选)

---

### Universal Compaction (折中方案)

**策略**：STCS 和 LCS 的混合

```
Level 0-N: 类似 STCS，相似大小合并
Level N+1: 类似 LCS，不重叠

动态调整：
- 写多时倾向 STCS
- 读多时倾向 LCS
```

**RocksDB Universal Compaction 示例**：

```
文件按时间排列（新到旧）：
File 1: 1MB  (newest)
File 2: 1MB
File 3: 1MB
File 4: 1MB
File 5: 10MB
File 6: 100MB (oldest)

触发条件 1：相似大小文件数 >= 阈值
File 1-4: 都是 1MB → 合并为 4MB

触发条件 2：空间放大比例 > 阈值
File 5 (10MB) + File 6 (100MB) = 110MB
如果 File 1-4 总共 20MB，空间放大 = 130MB / 100MB = 1.3
超过阈值（如 1.2）→ 全部合并
```

**优点**：

- 自适应工作负载
- 配置灵活

**缺点**：

- 参数复杂，难以调优
- 行为不如 STCS/LCS 可预测

---

### 三种策略对比表

| 维度         | Size-Tiered     | Leveled         | Universal      |
| ------------ | --------------- | --------------- | -------------- |
| **写放大**   | ✅ 低 (2-4x)    | ❌ 高 (10-100x) | 中 (5-20x)     |
| **读放大**   | ❌ 高           | ✅ 低           | 中             |
| **空间放大** | ❌ 高 (2x)      | ✅ 低 (1.1x)    | 中 (1.3-1.5x)  |
| **写吞吐**   | ✅ 高           | ❌ 低           | 中             |
| **读延迟**   | ❌ 高且不稳定   | ✅ 低且稳定     | 中             |
| **配置难度** | ✅ 简单         | 中              | ❌ 复杂        |
| **适用场景** | 写多，时序数据  | 读多，事务型    | 混合负载       |
| **代表系统** | Cassandra/HBase | LevelDB/RocksDB | RocksDB (可选) |

---

## 性能优化技术

### 1. Block Cache（块缓存）

```
问题：频繁读取相同数据导致重复磁盘 I/O

解决：LRU 缓存热点数据块

┌─────────────────────────────────────┐
│       Block Cache (LRU)             │
│  [Block 1: key100-key199, 4KB]      │
│  [Block 5: key500-key599, 4KB]      │
│  [Block 10: key1000-key1099, 4KB]   │
│  ...                                │
│  Total: 512MB (configurable)        │
└─────────────────────────────────────┘

读取流程（带缓存）：
1. 检查 Block Cache
   if hit: 返回数据（~0.01ms）
   if miss: 读取磁盘 → 加入缓存（~1ms）

缓存命中率：
- 热点数据：> 95%
- 冷数据：< 10%
```

**RocksDB 配置**：

```cpp
rocksdb::BlockBasedTableOptions table_options;
table_options.block_cache = rocksdb::NewLRUCache(512 * 1024 * 1024);  // 512MB
options.table_factory.reset(NewBlockBasedTableFactory(table_options));
```

---

### 2. Prefix Bloom Filter（前缀布隆过滤器）

```
问题：范围查询时 Bloom Filter 无效

传统 Bloom Filter:
查询 key=user:123
  ✅ 可判断 user:123 是否存在

查询 prefix=user:*（范围查询）
  ❌ 无法判断

Prefix Bloom Filter:
存储前缀的 Bloom Filter
prefix("user:123") = "user:"

查询 prefix=user:*
  ✅ 可快速判断该 SSTable 是否包含 user: 前缀的键
```

**应用场景**：

```
用户数据分片：
user:123:profile
user:123:posts
user:456:profile
user:456:posts

查询某个用户的所有数据：
SELECT * FROM data WHERE key LIKE 'user:123:%'

Prefix Bloom Filter 可快速跳过不包含 user:123: 的 SSTable
```

---

### 3. Partitioned Index/Filter（分区索引）

```
问题：大 SSTable 的索引和 Bloom Filter 占用内存大

传统方式：
1GB SSTable
├─ Index: 10MB（全部加载到内存）
├─ Bloom Filter: 5MB（全部加载到内存）

Partitioned 方式：
1GB SSTable
├─ Top-level Index: 100KB（常驻内存）
│   ├─ Partition 0: 0-100MB → Index offset
│   ├─ Partition 1: 100-200MB → Index offset
│   └─ ...
├─ Partition Indexes: 10MB（按需加载）
└─ Partition Bloom Filters: 5MB（按需加载）

查询流程：
1. 查 Top-level Index（内存）→ 定位到 Partition 3
2. 加载 Partition 3 的 Bloom Filter（磁盘 → 内存）
3. 加载 Partition 3 的 Index（磁盘 → 内存）
4. 查询数据块

内存节省：
- 传统：15MB（全部常驻）
- Partitioned：100KB（常驻）+ 按需加载

对于 100 个 1GB SSTable：
- 传统：1.5GB 内存
- Partitioned：10MB 内存 + 按需加载
```

---

### 4. Direct I/O（绕过操作系统缓存）

```
问题：操作系统页缓存和数据库缓存双重缓存，浪费内存

传统方式：
写入数据 → OS Page Cache → 磁盘
读取数据 → 磁盘 → OS Page Cache → 应用

内存浪费：
OS Page Cache: 1GB
Block Cache: 1GB
总计：2GB（存储相同数据！）

Direct I/O:
写入数据 → 磁盘（绕过 Page Cache）
读取数据 → 磁盘 → Block Cache

内存使用：
Block Cache: 1GB
总计：1GB

适用场景：
- 数据库自己管理缓存
- 大文件顺序 I/O
```

---

### 5. Parallel Compaction（并行压缩）

```
问题：单线程压缩成为瓶颈

单线程压缩：
Thread 1: Compact Level 0 → Level 1  (20 MB/s)

并行压缩：
Thread 1: Compact L0[0-25%] → L1     (20 MB/s)
Thread 2: Compact L0[25-50%] → L1    (20 MB/s)
Thread 3: Compact L0[50-75%] → L1    (20 MB/s)
Thread 4: Compact L0[75-100%] → L1   (20 MB/s)

总吞吐：80 MB/s

RocksDB 配置：
options.max_background_compactions = 4;
```

---

## 真实系统案例

### LevelDB（Google）

**设计目标**：

- 单机嵌入式 KV 存储
- 简单可靠
- 高性能读写

**关键设计**：

```
组件：
├─ MemTable: Skip List
├─ Compaction: Leveled
├─ Compression: Snappy
└─ WAL: 简单追加日志

优化：
├─ Block Cache (LRU)
├─ Bloom Filter（10 bits/key）
└─ Snappy 压缩

性能（SSD）：
├─ 写入：~40,000 ops/s
├─ 随机读：~100,000 ops/s
└─ 顺序读：~300,000 ops/s
```

**应用**：

- Chrome 浏览器 IndexedDB
- Bitcoin Core（区块链存储）
- 嵌入式数据库

---

### RocksDB（Facebook）

**基于 LevelDB，增强功能**：

```
新特性：
├─ Column Families（列族）
│   └─ 不同数据集使用不同配置
├─ 并行压缩
│   └─ 多线程后台压缩
├─ 多种压缩策略
│   ├─ Leveled (默认)
│   ├─ Universal
│   └─ FIFO
├─ Partitioned Index/Filter
├─ Direct I/O
└─ 统计信息和监控

性能（SSD）：
├─ 写入：~100,000 ops/s
├─ 随机读：~200,000 ops/s
└─ 顺序读：~500,000 ops/s
```

**应用**：

- MySQL MyRocks 存储引擎
- TiKV（TiDB 分布式 KV 存储）
- Apache Flink State Backend
- Kafka Streams State Store

---

### Cassandra（Apache）

**LSM-Tree 分布式实现**：

```
存储引擎：
├─ MemTable: ConcurrentSkipListMap
├─ Compaction: Size-Tiered (默认) / Leveled (可选)
└─ Commit Log (WAL)

分布式特性：
├─ 一致性哈希分片
├─ Gossip 协议
├─ Tunable Consistency
└─ Multi-DC 复制

写入路径：
1. 写入 Commit Log（所有副本）
2. 写入 MemTable
3. 返回成功（可配置一致性级别）

读取路径：
1. 查询 MemTable
2. 查询 Row Cache
3. 查询 SSTables（可能需要合并多个版本）
4. Read Repair（修复不一致）
```

**性能调优**：

```
写密集型（Size-Tiered）：
compaction_strategy = SizeTieredCompactionStrategy
write_throughput = 50,000 writes/s/node

读密集型（Leveled）：
compaction_strategy = LeveledCompactionStrategy
read_latency = p99 < 10ms
```

---

### TiKV（PingCAP）

**Rust 实现的分布式 KV 存储**：

```
存储引擎：RocksDB

分布式特性：
├─ Raft 一致性协议
├─ Multi-Raft（每个 Region 一个 Raft Group）
├─ Region 自动分裂/合并
└─ 分布式事务（Percolator）

优化：
├─ RocksDB Column Families
│   ├─ Default CF: 用户数据
│   ├─ Write CF: MVCC 写记录
│   └─ Lock CF: 事务锁
├─ Titan（大 Value 分离存储）
└─ Pipelined Raft

性能：
├─ 写入：~20,000 TPS（3 副本）
└─ 读取：~50,000 QPS
```

---

## 深度思考问题

### 1. 为什么 B-Tree 不用 WAL + MemTable 模式？

**回答**：

B-Tree 是原地更新，不需要排序缓冲区：

```
LSM-Tree:
写入 → MemTable（排序）→ 刷盘（有序 SSTable）

B-Tree:
写入 → WAL → 直接更新 B-Tree 页（原地修改）

B-Tree 使用 WAL 的原因：
- 崩溃恢复（Redo Log）
- 不是为了排序
```

---

### 2. 为什么不直接在磁盘上维护有序结构？

**回答**：

磁盘随机写极慢：

```
HDD 性能：
顺序写：100 MB/s
随机写：1 MB/s（100x 差异！）

维护磁盘有序结构（如磁盘 B-Tree）：
插入一个键 → 可能触发页分裂 → 多次随机写

LSM-Tree 方法：
所有写入都是顺序的（MemTable → SSTable 刷盘）
后台压缩也是顺序读写（归并排序）
```

---

### 3. 如果所有层都不重叠会怎样？

**回答**：

这就是 Leveled Compaction！

```
优势：
- 读放大低（每层最多查 1 个文件）
- 空间放大低

代价：
- 写放大高（需要频繁重写数据以保持不重叠）
```

如果 Level 0 也不重叠：

```
问题：
MemTable 刷盘时需要与 Level 0 合并
→ 刷盘变慢
→ MemTable 堆积
→ 写入阻塞

解决方案（实际系统）：
Level 0 允许重叠，Level 1+ 不重叠
平衡写入延迟和读取性能
```

---

### 4. 为什么 SSD 时代还需要 LSM-Tree？

**回答**：

SSD 随机写仍慢于顺序写：

```
SSD 性能：
顺序写：500 MB/s
随机写：100 MB/s（5x 差异）

更重要：
SSD 写放大问题
- 小写入 → SSD 内部读-改-写整个 Page（4KB）
- LSM-Tree 批量写入 → 减少 SSD 内部写放大
- 延长 SSD 寿命
```

此外：

- 压缩节省存储空间
- 批量写入提高吞吐量
- 适合写密集型场景

---

## 学习建议

### 理论基础

1. **理解权衡**：
   - 没有完美的数据结构
   - LSM-Tree：写优化，读代价
   - B-Tree：读优化，写代价

2. **掌握核心概念**：
   - SSTable 的有序性是一切优化的基础
   - 归并排序是合并的核心
   - Bloom Filter 是读优化的关键

3. **理解分层思想**：
   - 为什么需要多层？
   - 每层的作用是什么？
   - 如何平衡层数和文件数量？

### 实践项目

1. **实现简单 LSM-Tree**（推荐）：
   - MemTable（跳表或红黑树）
   - SSTable 文件格式
   - 简单的 Leveled Compaction
   - Bloom Filter

2. **阅读源码**：
   - LevelDB（C++，简洁）
   - BadgerDB（Go，现代）
   - 重点：`db_impl.cc`, `version_set.cc`, `table_builder.cc`

3. **性能测试**：
   - 对比不同压缩策略
   - 测试 Bloom Filter 效果
   - 分析写放大/读放大/空间放大

### 延伸阅读

**论文**：

- [The Log-Structured Merge-Tree (LSM-Tree)](https://www.cs.umb.edu/~poneil/lsmtree.pdf)
- [Bigtable: A Distributed Storage System](https://static.googleusercontent.com/media/research.google.com/en//archive/bigtable-osdi06.pdf)

**博客**：

- [LSM-based Storage Techniques: A Survey](https://arxiv.org/abs/1812.07527)
- [RocksDB Wiki](https://github.com/facebook/rocksdb/wiki)
- [Designing Data-Intensive Applications](https://dataintensive.net/) - Chapter 3

**视频**：

- [CMU Database Systems - LSM Trees](https://www.youtube.com/watch?v=I6jB0nM9SKU)
- [The Log-Structured Merge-Tree](https://www.youtube.com/watch?v=b6SI8VbcT4w)

---

## 总结

### 核心要点

1. **SSTable = 有序存储**
   - 稀疏索引解决内存问题
   - 归并排序高效合并
   - 支持范围查询
   - 块压缩节省空间

2. **LSM-Tree = 写优化**
   - MemTable 缓冲写入
   - 后台压缩异步处理
   - 顺序 I/O，高吞吐
   - 适合写密集型场景

3. **权衡无处不在**
   - 写放大 vs 读放大 vs 空间放大
   - Leveled vs Size-Tiered
   - 内存 vs 磁盘
   - 吞吐量 vs 延迟

4. **优化是关键**
   - Bloom Filter（减少无效查询）
   - Block Cache（缓存热点数据）
   - Parallel Compaction（提高压缩速度）
   - 压缩算法（减少 I/O）

### 最重要的洞察

**LSM-Tree 的成功秘诀**：

> 将随机写转换为顺序写，将写入成本转移到后台压缩，用写放大换取高吞吐量。

这个思想适用于所有需要高性能写入的系统！

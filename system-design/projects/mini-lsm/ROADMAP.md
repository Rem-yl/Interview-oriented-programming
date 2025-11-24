# Mini-LSM Go 实现路线图

> 参考项目：https://github.com/skyzh/mini-lsm
>
> 目标：使用 Go 语言实现一个教学版的 LSM-Tree 存储引擎

## 项目概述

LSM-Tree (Log-Structured Merge-Tree) 是一种针对写密集型工作负载优化的数据结构，被广泛应用于现代数据库系统中（如 LevelDB、RocksDB、Cassandra 等）。

### 核心特性
- **写优化**：所有写操作先进入内存，批量刷盘，顺序写入磁盘
- **读优化**：通过 Bloom Filter、索引等减少磁盘访问
- **空间回收**：通过 Compaction 合并和删除过时数据

---

## 实现计划（10 周）

### Week 1: Block - 基础存储单元

**目标**：实现 LSM-Tree 中最基础的数据块结构

#### 核心概念
- Block 是 SSTable 的基本存储单元
- 每个 Block 存储多个有序的 key-value 对
- 包含数据区和偏移量索引区

#### 实现任务
1. **Block 结构设计**
   ```go
   type Block struct {
       data    []byte  // 存储实际数据
       offsets []uint16 // key-value 对的起始偏移量
   }
   ```

2. **BlockBuilder**
   - `Add(key, value []byte)` - 添加 kv 对
   - `Build() *Block` - 构建最终的 Block
   - 实现编码格式：`[entry1][entry2]...[offset1][offset2]...[num_entries]`

3. **BlockIterator**
   - `SeekToFirst()` - 定位到第一个元素
   - `SeekToKey(key []byte)` - 二分查找定位
   - `Next()` - 移动到下一个元素
   - `Key() []byte` 和 `Value() []byte` - 读取当前元素

#### 测试用例
- 添加单个/多个 kv 对
- 编码/解码正确性
- Iterator 遍历和查找

#### 参考资料
- LevelDB Block Format: https://github.com/google/leveldb/blob/main/doc/table_format.md

---

### Week 2: SSTable - 持久化有序表

**目标**：实现可持久化到磁盘的 Sorted String Table

#### 核心概念
- SSTable 由多个 Block 组成
- 包含数据 Block + 索引 Block + Meta Block + Footer
- 不可变结构，一旦写入就不再修改

#### 实现任务
1. **SSTable 结构**
   ```go
   type SSTable struct {
       file      *os.File
       blocks    []BlockMeta  // Block 元数据
       blockIdx  *Block       // 索引 Block
       firstKey  []byte
       lastKey   []byte
   }
   ```

2. **SSTableBuilder**
   - `Add(key, value []byte)` - 添加有序的 kv 对
   - `Build(path string) error` - 写入磁盘
   - 文件格式：`[Block1][Block2]...[BlockIndex][BlockMeta][Footer]`

3. **SSTableIterator**
   - 在多个 Block 之间迭代
   - 实现 `Seek`, `Next`, `Key`, `Value`

4. **SSTable Reader**
   - `Open(path string) (*SSTable, error)` - 从磁盘加载
   - `Get(key []byte) ([]byte, bool)` - 点查询

#### 测试用例
- 构建包含多个 Block 的 SSTable
- 持久化和重新加载
- 范围查询和点查询

---

### Week 3: MemTable - 内存表

**目标**：实现高效的内存可变数据结构

#### 核心概念
- MemTable 接收所有写操作
- 使用 SkipList 或 TreeMap 保持有序
- 达到阈值后刷入磁盘成为 SSTable

#### 实现任务
1. **MemTable 结构**
   ```go
   type MemTable struct {
       data *skiplist.SkipList  // 或使用 sync.Map + slice
       size int64                // 当前大小
   }
   ```

2. **基本操作**
   - `Put(key, value []byte)` - 写入数据
   - `Get(key []byte) ([]byte, bool)` - 读取数据
   - `Delete(key []byte)` - 删除（写入 tombstone）
   - `Flush(path string) error` - 刷盘为 SSTable

3. **MemTableIterator**
   - 按序遍历内存中的数据

#### 实现选择
- **选项 1**：使用第三方 skiplist 库
- **选项 2**：使用 Go 标准库 + 自己实现排序

#### 测试用例
- 并发写入安全性
- 有序性验证
- Flush 到 SSTable

---

### Week 4: Merge Iterator - 多路归并

**目标**：实现合并多个有序数据源的迭代器

#### 核心概念
- LSM 读取需要查询 MemTable + 多个 SSTable
- 使用优先队列（最小堆）合并多个 Iterator
- 处理重复 key（取最新版本）

#### 实现任务
1. **MergeIterator**
   ```go
   type MergeIterator struct {
       iters []Iterator    // 子迭代器
       heap  *IteratorHeap // 最小堆
   }
   ```

2. **核心逻辑**
   - 使用 `container/heap` 实现优先队列
   - 每次返回所有 Iterator 中最小的 key
   - 处理相同 key 的多个版本

3. **Two-Merge Iterator**（优化）
   - 专门用于合并两个 Iterator
   - 避免堆的开销

#### 测试用例
- 合并 2 个有序序列
- 合并 N 个有序序列
- 处理重复 key

---

### Week 5: LSM Storage Engine - 引擎框架

**目标**：整合所有组件，实现完整的读写路径

#### 核心概念
- 写入路径：Write → MemTable → (满了) → SSTable (L0)
- 读取路径：MemTable → L0 SSTables → L1 SSTables → ...
- 使用 Manifest 管理元数据

#### 实现任务
1. **LSMStorage 结构**
   ```go
   type LSMStorage struct {
       memtable    *MemTable
       imemtables  []*MemTable  // 正在刷盘的 MemTable
       l0SSTables  []*SSTable
       levels      [][]*SSTable
       mu          sync.RWMutex
   }
   ```

2. **写操作**
   - `Put(key, value []byte) error`
   - `Delete(key []byte) error`
   - MemTable 满了触发 Flush

3. **读操作**
   - `Get(key []byte) ([]byte, error)`
   - 按顺序查询：memtable → imemtables → L0 → L1 → ...
   - `Scan(start, end []byte) Iterator`

4. **后台任务**
   - Flush MemTable 到 L0
   - 定期 Compaction

#### 测试用例
- 基本的 CRUD 操作
- 重启后数据恢复
- 并发读写

---

### Week 6: Write-Ahead Log (WAL)

**目标**：实现 WAL 保证数据持久性

#### 核心概念
- 写操作先追加到 WAL 文件
- MemTable 崩溃可从 WAL 恢复
- WAL 可以在 MemTable 刷盘后删除

#### 实现任务
1. **WAL 结构**
   ```go
   type WAL struct {
       file *os.File
       mu   sync.Mutex
   }
   ```

2. **WAL 格式**
   - 每条记录：`[checksum][key_len][value_len][key][value]`
   - 支持 Put 和 Delete 两种操作

3. **集成到引擎**
   - `Put/Delete` 先写 WAL，再写 MemTable
   - 启动时从 WAL 恢复
   - MemTable 刷盘后删除对应 WAL

#### 测试用例
- 写入 WAL 并恢复
- 崩溃恢复场景模拟
- Checksum 验证

---

### Week 7: Bloom Filter

**目标**：优化不存在 key 的查询性能

#### 核心概念
- Bloom Filter 是概率性数据结构
- 可以快速判断 key 一定不存在
- 每个 SSTable 配备一个 Bloom Filter

#### 实现任务
1. **Bloom Filter 实现**
   ```go
   type BloomFilter struct {
       bits   []byte
       k      int  // hash 函数个数
   }
   ```

2. **基本操作**
   - `Add(key []byte)` - 添加 key
   - `MayContain(key []byte) bool` - 查询
   - 使用 murmur3 等 hash 函数

3. **集成到 SSTable**
   - SSTableBuilder 构建时生成 Bloom Filter
   - 查询时先检查 Bloom Filter

#### 测试用例
- False positive rate 测试
- 性能对比（有/无 Bloom Filter）

---

### Week 8: Compaction - 压缩策略

**目标**：实现 Simple Leveled Compaction

#### 核心概念
- L0 层文件重叠，达到阈值后合并到 L1
- L1+ 层文件不重叠，每层大小呈指数增长
- Compaction 删除过时数据，回收空间

#### 实现任务
1. **Leveled Compaction**
   - L0 → L1：选择所有重叠的 L0 文件
   - L1 → L2：选择大小超限的文件
   - 多路归并，输出新的 SSTable

2. **触发时机**
   - L0 文件数 > 阈值（如 4）
   - Ln 层总大小 > `10^n * base_size`

3. **Manifest 管理**
   - 记录每次 Compaction 的变更
   - 保证元数据持久化

#### 测试用例
- 触发 L0 → L1 Compaction
- 验证文件数和数据正确性
- 空间回收验证

---

### Week 9-10: 优化与完善

#### 性能优化
1. **并发优化**
   - 读写分离锁
   - Immutable MemTable 允许并发读

2. **缓存**
   - Block Cache (LRU)
   - SSTable 文件句柄缓存

3. **压缩**
   - Block 级别的 Snappy 压缩

#### 完善功能
1. **Snapshot**
   - 支持一致性快照读取

2. **Range Delete**
   - 高效删除范围数据

3. **统计信息**
   - 读写放大统计
   - Compaction 统计

#### 测试与基准
1. **集成测试**
   - 大规模数据写入
   - 崩溃恢复测试
   - 并发读写压力测试

2. **性能基准**
   - `go test -bench=.`
   - 与 LevelDB/RocksDB 对比

---

## 项目结构

```
mini-lsm/
├── pkg/
│   ├── block/           # Week 1
│   │   ├── block.go
│   │   ├── builder.go
│   │   └── iterator.go
│   ├── sstable/         # Week 2
│   │   ├── sstable.go
│   │   ├── builder.go
│   │   └── iterator.go
│   ├── memtable/        # Week 3
│   │   ├── memtable.go
│   │   └── iterator.go
│   ├── iterators/       # Week 4
│   │   ├── merge.go
│   │   └── two_merge.go
│   ├── lsm/             # Week 5
│   │   ├── storage.go
│   │   ├── write.go
│   │   └── read.go
│   ├── wal/             # Week 6
│   │   └── wal.go
│   ├── bloom/           # Week 7
│   │   └── bloom.go
│   └── compact/         # Week 8
│       ├── leveled.go
│       └── manifest.go
├── tests/
│   └── integration_test.go
├── cmd/
│   └── mini-lsm/
│       └── main.go      # CLI 工具
├── go.mod
└── README.md
```

---

## 学习资源

### 必读论文
1. **The Log-Structured Merge-Tree (LSM-Tree)** - Patrick O'Neil et al.
   - http://www.cs.umb.edu/~poneil/lsmtree.pdf

2. **Bigtable: A Distributed Storage System**
   - https://static.googleusercontent.com/media/research.google.com/en//archive/bigtable-osdi06.pdf

### 参考实现
1. **LevelDB** (C++) - Google
   - https://github.com/google/leveldb
   - 简单清晰，适合学习

2. **RocksDB** (C++) - Meta
   - https://github.com/facebook/rocksdb
   - 生产级，功能丰富

3. **BadgerDB** (Go) - DGRAPH
   - https://github.com/dgraph-io/badger
   - Go 语言实现，可参考

### 书籍
- **Database Internals** - Alex Petrov
  - 第 7 章详细介绍 LSM-Tree

---

## 开发建议

### 1. 测试驱动开发 (TDD)
- 每个模块先写测试用例
- 使用 `testing` 包和 `testify` 断言库
- 目标：测试覆盖率 > 80%

### 2. 性能基准
```go
func BenchmarkBlockIterator(b *testing.B) {
    // 测试代码
}
```

### 3. 代码质量
- 使用 `golangci-lint` 静态检查
- 代码注释解释"为什么"而不是"是什么"
- 关键算法配图说明

### 4. 渐进式实现
- 先实现最简单的版本（能跑通）
- 再添加优化（性能、并发安全）
- 最后完善边界情况

### 5. 文档化
- 每个 Week 结束写总结文档
- 记录设计决策和权衡
- 画架构图和数据流图

---

## 里程碑

- **Week 1-2**: 完成基础组件（Block + SSTable）
- **Week 3-4**: 实现内存和迭代器（MemTable + Iterator）
- **Week 5**: 🎯 **第一个可用版本**（基本读写）
- **Week 6-7**: 增强可靠性和性能（WAL + Bloom Filter）
- **Week 8**: 实现空间回收（Compaction）
- **Week 9-10**: 🎯 **生产级版本**（优化 + 完善测试）

---

## 下一步

1. ✅ 初始化 Go Module: `go mod init github.com/yourname/mini-lsm`
2. ✅ 创建项目目录结构
3. 📖 阅读 LSM-Tree 论文了解原理
4. 💻 开始 Week 1: 实现 Block 组件

Good luck! 🚀

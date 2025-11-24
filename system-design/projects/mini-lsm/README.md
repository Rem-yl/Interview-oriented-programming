# Mini-LSM - Go Implementation

> A tutorial implementation of LSM-Tree storage engine in Go
>
> 参考项目: https://github.com/skyzh/mini-lsm (Rust)

## 项目简介

这是一个用 Go 语言实现的简化版 LSM-Tree（Log-Structured Merge-Tree）存储引擎，主要用于学习和理解现代数据库存储引擎的核心原理。

### 什么是 LSM-Tree？

LSM-Tree 是一种针对**写密集型**工作负载优化的数据结构，被广泛应用于：
- **LevelDB** / **RocksDB** - 嵌入式 KV 存储
- **Cassandra** / **HBase** - 分布式数据库
- **ClickHouse** - 分析型数据库

### 核心思想

1. **写优化**: 所有写操作先写入内存（MemTable），达到阈值后批量刷盘
2. **顺序写**: 磁盘写入都是顺序的，充分利用磁盘带宽
3. **分层存储**: 数据分多层存储，定期合并（Compaction）
4. **读优化**: 使用 Bloom Filter、缓存等减少磁盘 I/O

## 功能特性

- [x] 基础 Block 编码/解码
- [ ] SSTable 持久化存储
- [ ] MemTable 内存表
- [ ] 多路归并迭代器
- [ ] LSM Storage Engine
- [ ] Write-Ahead Log (WAL)
- [ ] Bloom Filter
- [ ] Leveled Compaction

## 快速开始

### 安装

```bash
git clone https://github.com/rem/mini-lsm.git
cd mini-lsm
go mod tidy
```

### 使用示例

```go
package main

import (
    "github.com/rem/mini-lsm/pkg/lsm"
)

func main() {
    // 打开存储引擎
    storage, err := lsm.Open("/tmp/mini-lsm")
    if err != nil {
        panic(err)
    }
    defer storage.Close()

    // 写入数据
    storage.Put([]byte("key1"), []byte("value1"))
    storage.Put([]byte("key2"), []byte("value2"))

    // 读取数据
    value, err := storage.Get([]byte("key1"))
    if err != nil {
        panic(err)
    }
    println(string(value)) // "value1"

    // 删除数据
    storage.Delete([]byte("key1"))

    // 范围扫描
    iter := storage.Scan([]byte("key1"), []byte("key3"))
    for iter.Valid() {
        println(string(iter.Key()), string(iter.Value()))
        iter.Next()
    }
}
```

## 项目结构

```
mini-lsm/
├── pkg/
│   ├── block/           # Block 存储单元
│   ├── sstable/         # SSTable 持久化
│   ├── memtable/        # 内存表
│   ├── iterators/       # 迭代器
│   ├── lsm/             # LSM 引擎核心
│   ├── wal/             # Write-Ahead Log
│   ├── bloom/           # Bloom Filter
│   └── compact/         # Compaction 策略
├── tests/               # 集成测试
├── cmd/mini-lsm/        # CLI 工具
├── ROADMAP.md           # 详细实现路线图
└── README.md
```

## 开发进度

查看 [ROADMAP.md](./ROADMAP.md) 了解详细的 10 周实现计划。

### 当前进度: Week 1

- [ ] Block 编码/解码
- [ ] BlockBuilder
- [ ] BlockIterator

## 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./pkg/block -v

# 运行性能基准
go test ./pkg/block -bench=. -benchmem
```

## 学习资源

### 论文
- [The Log-Structured Merge-Tree (LSM-Tree)](http://www.cs.umb.edu/~poneil/lsmtree.pdf)
- [Bigtable: A Distributed Storage System](https://research.google/pubs/pub27898/)

### 开源项目
- [LevelDB](https://github.com/google/leveldb) - Google 的 LSM 实现（C++）
- [RocksDB](https://github.com/facebook/rocksdb) - Meta 优化版（C++）
- [BadgerDB](https://github.com/dgraph-io/badger) - Go 实现
- [mini-lsm](https://github.com/skyzh/mini-lsm) - Rust 教学实现

### 书籍
- Database Internals (Chapter 7: Log-Structured Storage)
- Designing Data-Intensive Applications

## 性能目标

- **写吞吐**: > 100K writes/sec
- **读延迟**: < 1ms (缓存命中)
- **空间放大**: < 3x
- **写放大**: < 10x

## 贡献

欢迎提 Issue 和 PR！

## License

MIT License

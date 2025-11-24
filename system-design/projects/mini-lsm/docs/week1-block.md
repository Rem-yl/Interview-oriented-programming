# Week 1: Block - 基础存储单元

## 概述

Block 是 LSM-Tree 中最基础的数据存储单元，类似于数据库中的"页"（Page）概念。每个 Block 存储多个有序的 key-value 对，是 SSTable 的组成部分。

## 学习目标

- 理解 Block 的编码格式
- 实现 BlockBuilder 和 Block
- 实现 BlockIterator 支持遍历和二分查找
- 掌握变长编码（varint）的使用

## Block 编码格式

### 设计考虑

1. **存储效率**: 紧凑的编码格式，减少空间占用
2. **查询效率**: 支持快速的二分查找定位
3. **简单性**: 编解码逻辑简单，易于实现和调试

### 格式定义

```
+------------------+------------------+-----+------------------+
|      Entry 1     |      Entry 2     | ... |      Entry N     |
+------------------+------------------+-----+------------------+
|   Offset 1 (2B)  |   Offset 2 (2B)  | ... |   Offset N (2B)  |
+------------------+------------------+-----+------------------+
|         Number of entries (2B)                               |
+------------------------------------------------------------------+
```

每个 Entry 的格式：

```
+--------------------+--------------------+--------------------+--------------------+
| key_overlap_len(2B)| key_rest_len (2B)  |  value_len (2B)    |  key_rest | value  |
+--------------------+--------------------+--------------------+--------------------+
```

### 前缀压缩（Key Overlap）

相邻的 key 通常有公共前缀，可以只存储不同的部分：

```
Entry 1: "apple"     → overlap=0, rest="apple"
Entry 2: "application" → overlap=4, rest="ication"  (共享 "appl")
Entry 3: "apply"     → overlap=4, rest="y"         (共享 "appl")
```

### 示例

假设要存储以下 KV 对：

```
"key1" => "value1"
"key2" => "value2"
```

编码后的数据：

```
数据区:
  Entry 1: [0, 0] [4, 0] [6, 0] "key1" "value1"  (overlap=0, key_len=4, val_len=6)
  Entry 2: [3, 0] [1, 0] [6, 0] "2" "value2"     (overlap=3, key_len=1, val_len=6)

索引区:
  [0, 0]     (Entry 1 的 offset = 0)
  [18, 0]    (Entry 2 的 offset = 18)

元数据:
  [2, 0]     (共 2 个 entries)
```

## 实现步骤

### Step 1: 定义数据结构

```go
// pkg/block/block.go

package block

const (
    // Block 最大大小（4KB）
    BlockSize = 4096

    // 每个 offset 占用 2 字节
    SizeOfU16 = 2
)

// Block 是不可变的数据块
type Block struct {
    data    []byte   // 存储 entries
    offsets []uint16 // 每个 entry 的起始位置
}

// Entry 表示一个 KV 对
type Entry struct {
    Key   []byte
    Value []byte
}
```

### Step 2: 实现 BlockBuilder

```go
// pkg/block/builder.go

package block

import "encoding/binary"

type BlockBuilder struct {
    data       []byte   // 数据缓冲区
    offsets    []uint16 // offset 数组
    blockSize  int      // 目标 block 大小
    firstKey   []byte   // 第一个 key（用于压缩）
}

func NewBlockBuilder(blockSize int) *BlockBuilder {
    return &BlockBuilder{
        data:      make([]byte, 0, blockSize),
        offsets:   make([]uint16, 0),
        blockSize: blockSize,
    }
}

// Add 添加一个 KV 对
// 注意：key 必须按顺序添加
func (b *BlockBuilder) Add(key, value []byte) bool {
    // TODO: 实现
    // 1. 计算 key overlap
    // 2. 编码 entry
    // 3. 检查 block 是否已满
    // 4. 追加到 data 和 offsets
    return true
}

// EstimatedSize 返回当前 block 的估计大小
func (b *BlockBuilder) EstimatedSize() int {
    // data + offsets + num_entries
    return len(b.data) + len(b.offsets)*SizeOfU16 + SizeOfU16
}

// IsEmpty 检查是否为空
func (b *BlockBuilder) IsEmpty() bool {
    return len(b.offsets) == 0
}

// Build 构建最终的 Block
func (b *BlockBuilder) Build() *Block {
    // TODO: 实现
    // 1. 将 offsets 和 num_entries 追加到 data
    // 2. 返回 Block
    return nil
}
```

### Step 3: 实现 Block 解码

```go
// pkg/block/block.go

// Decode 从字节数组解码 Block
func Decode(data []byte) (*Block, error) {
    // TODO: 实现
    // 1. 读取最后 2 字节得到 num_entries
    // 2. 读取 offsets 数组
    // 3. 返回 Block
    return nil, nil
}

// Encode 编码 Block 到字节数组
func (b *Block) Encode() []byte {
    // TODO: 实现
    return nil
}
```

### Step 4: 实现 BlockIterator

```go
// pkg/block/iterator.go

package block

type BlockIterator struct {
    block      *Block
    idx        int      // 当前 entry 索引
    key        []byte   // 当前 key
    value      []byte   // 当前 value
    firstKey   []byte   // 第一个完整 key（用于解压缩）
}

func NewBlockIterator(block *Block) *BlockIterator {
    iter := &BlockIterator{
        block: block,
        idx:   0,
    }
    iter.SeekToFirst()
    return iter
}

// SeekToFirst 移动到第一个 entry
func (it *BlockIterator) SeekToFirst() {
    // TODO: 实现
    it.idx = 0
    it.decodeEntry()
}

// SeekToKey 二分查找定位到 >= key 的第一个 entry
func (it *BlockIterator) SeekToKey(key []byte) {
    // TODO: 实现
    // 1. 使用二分查找定位 offset
    // 2. 解码 entry
}

// Next 移动到下一个 entry
func (it *BlockIterator) Next() {
    it.idx++
    if it.idx < len(it.block.offsets) {
        it.decodeEntry()
    }
}

// Valid 检查迭代器是否有效
func (it *BlockIterator) Valid() bool {
    return it.idx < len(it.block.offsets)
}

// Key 返回当前 key
func (it *BlockIterator) Key() []byte {
    return it.key
}

// Value 返回当前 value
func (it *BlockIterator) Value() []byte {
    return it.value
}

// decodeEntry 解码当前 entry
func (it *BlockIterator) decodeEntry() {
    // TODO: 实现
    // 1. 根据 offset 读取 entry
    // 2. 解析 key_overlap, key_rest, value
    // 3. 重建完整 key
}
```

## 测试用例

### 基础测试

```go
// pkg/block/block_test.go

package block

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestBlockBuildAndDecode(t *testing.T) {
    builder := NewBlockBuilder(BlockSize)

    // 添加数据
    builder.Add([]byte("key1"), []byte("value1"))
    builder.Add([]byte("key2"), []byte("value2"))
    builder.Add([]byte("key3"), []byte("value3"))

    // 构建 block
    block := builder.Build()
    assert.NotNil(t, block)

    // 编码和解码
    data := block.Encode()
    decoded, err := Decode(data)
    assert.NoError(t, err)

    // 验证解码后的数据
    iter := NewBlockIterator(decoded)
    assert.True(t, iter.Valid())
    assert.Equal(t, []byte("key1"), iter.Key())
    assert.Equal(t, []byte("value1"), iter.Value())
}

func TestBlockIterator(t *testing.T) {
    builder := NewBlockBuilder(BlockSize)
    builder.Add([]byte("apple"), []byte("fruit"))
    builder.Add([]byte("banana"), []byte("fruit"))
    builder.Add([]byte("cherry"), []byte("fruit"))

    block := builder.Build()
    iter := NewBlockIterator(block)

    // 测试顺序遍历
    keys := []string{}
    for iter.Valid() {
        keys = append(keys, string(iter.Key()))
        iter.Next()
    }
    assert.Equal(t, []string{"apple", "banana", "cherry"}, keys)
}

func TestBlockIteratorSeek(t *testing.T) {
    builder := NewBlockBuilder(BlockSize)
    builder.Add([]byte("a"), []byte("1"))
    builder.Add([]byte("c"), []byte("3"))
    builder.Add([]byte("e"), []byte("5"))

    block := builder.Build()
    iter := NewBlockIterator(block)

    // SeekToKey("b") 应该定位到 "c"
    iter.SeekToKey([]byte("b"))
    assert.Equal(t, []byte("c"), iter.Key())

    // SeekToKey("d") 应该定位到 "e"
    iter.SeekToKey([]byte("d"))
    assert.Equal(t, []byte("e"), iter.Key())
}
```

### 性能基准测试

```go
func BenchmarkBlockBuild(b *testing.B) {
    for i := 0; i < b.N; i++ {
        builder := NewBlockBuilder(BlockSize)
        for j := 0; j < 100; j++ {
            key := []byte(fmt.Sprintf("key%04d", j))
            value := []byte("value")
            builder.Add(key, value)
        }
        builder.Build()
    }
}

func BenchmarkBlockIterator(b *testing.B) {
    builder := NewBlockBuilder(BlockSize)
    for j := 0; j < 100; j++ {
        key := []byte(fmt.Sprintf("key%04d", j))
        value := []byte("value")
        builder.Add(key, value)
    }
    block := builder.Build()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        iter := NewBlockIterator(block)
        for iter.Valid() {
            _ = iter.Key()
            _ = iter.Value()
            iter.Next()
        }
    }
}
```

## 实现提示

### 1. 计算 Key Overlap

```go
func keyOverlap(prev, curr []byte) int {
    minLen := len(prev)
    if len(curr) < minLen {
        minLen = len(curr)
    }

    for i := 0; i < minLen; i++ {
        if prev[i] != curr[i] {
            return i
        }
    }
    return minLen
}
```

### 2. 编码 uint16

```go
func encodeU16(buf []byte, val uint16) {
    binary.LittleEndian.PutUint16(buf, val)
}

func decodeU16(buf []byte) uint16 {
    return binary.LittleEndian.Uint16(buf)
}
```

### 3. 二分查找

```go
func (it *BlockIterator) binarySearch(key []byte) int {
    left, right := 0, len(it.block.offsets)

    for left < right {
        mid := (left + right) / 2
        // 解码 mid 位置的 key
        midKey := it.decodeKeyAt(mid)

        if bytes.Compare(midKey, key) < 0 {
            left = mid + 1
        } else {
            right = mid
        }
    }
    return left
}
```

## 常见问题

### Q1: 为什么使用 uint16 存储 offset？

A: uint16 可以表示 0-65535，对于 4KB 的 Block 足够使用。如果需要更大的 Block，可以使用 uint32。

### Q2: Key 必须有序吗？

A: 是的！Block 内的 key 必须严格有序，这样才能：
- 使用前缀压缩
- 使用二分查找
- 支持范围查询

### Q3: 如何选择 Block 大小？

A: 常见选择：
- LevelDB: 4KB
- RocksDB: 4KB-64KB（可配置）
- 权衡：更大的 Block → 更好的压缩率，但读放大更严重

## 验收标准

完成 Week 1 后，你应该能够：

- ✅ 实现 Block 编码/解码，通过所有测试用例
- ✅ BlockIterator 支持顺序遍历
- ✅ BlockIterator 支持二分查找（SeekToKey）
- ✅ 理解前缀压缩的原理和实现
- ✅ 性能基准测试显示合理的性能

## 下一步

完成 Week 1 后，继续 [Week 2: SSTable](./week2-sstable.md)，将多个 Block 组织成持久化的 SSTable 文件。

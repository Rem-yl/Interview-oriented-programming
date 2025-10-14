# LRU Cache 实现

## 项目概述

使用 Go 实现一个高效的 LRU (Least Recently Used) 缓存,支持快速的读写操作和自动淘汰机制。

## 学习目标

- 理解 LRU 缓存淘汰算法
- 掌握双向链表 + HashMap 的数据结构组合
- 学习 Go 的并发安全实现
- 实践时间复杂度优化

## 功能特性

- [ ] O(1) 时间复杂度的 Get 操作
- [ ] O(1) 时间复杂度的 Put 操作
- [ ] 容量限制和自动淘汰
- [ ] 线程安全 (goroutine-safe)
- [ ] 支持过期时间 (TTL)
- [ ] 统计信息 (命中率、淘汰次数等)

## 项目结构

```
lru-cache/
├── main.go              # 示例程序
├── go.mod
├── go.sum
├── README.md
├── pkg/
│   ├── lru.go          # LRU 核心实现
│   ├── node.go         # 双向链表节点
│   └── stats.go        # 统计信息
└── tests/
    ├── lru_test.go     # 单元测试
    └── benchmark_test.go  # 性能测试
```

## 数据结构设计

### 节点定义

```go
type Node struct {
    key   string
    value interface{}
    prev  *Node
    next  *Node
    expireAt time.Time
}
```

### LRU Cache 结构

```go
type LRUCache struct {
    capacity int
    cache    map[string]*Node
    head     *Node  // 虚拟头节点
    tail     *Node  // 虚拟尾节点
    mu       sync.RWMutex
    stats    *Stats
}
```

## 核心算法

### 1. Get 操作

```go
func (lru *LRUCache) Get(key string) (interface{}, bool) {
    lru.mu.Lock()
    defer lru.mu.Unlock()

    node, exists := lru.cache[key]
    if !exists {
        lru.stats.RecordMiss()
        return nil, false
    }

    // 检查是否过期
    if !node.expireAt.IsZero() && time.Now().After(node.expireAt) {
        lru.removeNode(node)
        delete(lru.cache, key)
        lru.stats.RecordMiss()
        return nil, false
    }

    // 移动到链表头部 (最近使用)
    lru.moveToHead(node)
    lru.stats.RecordHit()
    return node.value, true
}
```

### 2. Put 操作

```go
func (lru *LRUCache) Put(key string, value interface{}, ttl time.Duration) {
    lru.mu.Lock()
    defer lru.mu.Unlock()

    // 如果 key 已存在,更新并移到头部
    if node, exists := lru.cache[key]; exists {
        node.value = value
        if ttl > 0 {
            node.expireAt = time.Now().Add(ttl)
        }
        lru.moveToHead(node)
        return
    }

    // 创建新节点
    newNode := &Node{
        key:   key,
        value: value,
    }
    if ttl > 0 {
        newNode.expireAt = time.Now().Add(ttl)
    }

    lru.cache[key] = newNode
    lru.addToHead(newNode)

    // 检查容量,必要时淘汰最久未使用的
    if len(lru.cache) > lru.capacity {
        tail := lru.removeTail()
        delete(lru.cache, tail.key)
        lru.stats.RecordEviction()
    }
}
```

### 3. 链表操作

```go
// 移动节点到头部
func (lru *LRUCache) moveToHead(node *Node) {
    lru.removeNode(node)
    lru.addToHead(node)
}

// 添加节点到头部
func (lru *LRUCache) addToHead(node *Node) {
    node.prev = lru.head
    node.next = lru.head.next
    lru.head.next.prev = node
    lru.head.next = node
}

// 移除节点
func (lru *LRUCache) removeNode(node *Node) {
    node.prev.next = node.next
    node.next.prev = node.prev
}

// 移除尾部节点
func (lru *LRUCache) removeTail() *Node {
    node := lru.tail.prev
    lru.removeNode(node)
    return node
}
```

## 使用示例

```go
package main

import (
    "fmt"
    "time"
    "your-module/pkg"
)

func main() {
    // 创建容量为 3 的 LRU 缓存
    cache := pkg.NewLRUCache(3)

    // 添加数据
    cache.Put("user:1", "Alice", 0)
    cache.Put("user:2", "Bob", 0)
    cache.Put("user:3", "Charlie", 0)

    // 读取数据 (会将 user:1 移到最前)
    if val, ok := cache.Get("user:1"); ok {
        fmt.Println("Found:", val)
    }

    // 添加新数据,会淘汰 user:2 (最久未使用)
    cache.Put("user:4", "David", 0)

    // user:2 已被淘汰
    if _, ok := cache.Get("user:2"); !ok {
        fmt.Println("user:2 was evicted")
    }

    // 带 TTL 的缓存
    cache.Put("session:123", "token-xyz", 5*time.Second)

    // 打印统计信息
    fmt.Println(cache.Stats())
}
```

## 测试用例

```go
func TestLRUCache(t *testing.T) {
    cache := NewLRUCache(2)

    cache.Put("1", "one", 0)
    cache.Put("2", "two", 0)

    // 测试正常获取
    val, ok := cache.Get("1")
    assert.True(t, ok)
    assert.Equal(t, "one", val)

    // 添加第3个元素,应该淘汰 "2"
    cache.Put("3", "three", 0)

    _, ok = cache.Get("2")
    assert.False(t, ok)

    // "1" 和 "3" 应该还在
    _, ok = cache.Get("1")
    assert.True(t, ok)
    _, ok = cache.Get("3")
    assert.True(t, ok)
}

func TestLRUCacheTTL(t *testing.T) {
    cache := NewLRUCache(10)

    cache.Put("temp", "value", 100*time.Millisecond)

    // 立即获取应该成功
    _, ok := cache.Get("temp")
    assert.True(t, ok)

    // 等待过期
    time.Sleep(150 * time.Millisecond)

    // 过期后应该获取失败
    _, ok = cache.Get("temp")
    assert.False(t, ok)
}
```

## 性能测试

```go
func BenchmarkLRUCachePut(b *testing.B) {
    cache := NewLRUCache(1000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Put(fmt.Sprintf("key-%d", i), i, 0)
    }
}

func BenchmarkLRUCacheGet(b *testing.B) {
    cache := NewLRUCache(1000)

    // 预填充数据
    for i := 0; i < 1000; i++ {
        cache.Put(fmt.Sprintf("key-%d", i), i, 0)
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Get(fmt.Sprintf("key-%d", i%1000))
    }
}
```

## 扩展功能

- [ ] 实现 LFU (Least Frequently Used) 缓存
- [ ] 添加批量操作 (GetMany, PutMany)
- [ ] 实现缓存持久化
- [ ] 支持缓存预热
- [ ] 添加回调函数 (淘汰时触发)
- [ ] 实现分段锁,提高并发性能

## 对比其他实现

| 实现方式 | 时间复杂度 | 空间复杂度 | 并发安全 |
|---------|-----------|-----------|---------|
| HashMap + 双向链表 | O(1) | O(n) | ✓ |
| 有序 Map (TreeMap) | O(log n) | O(n) | ✓ |
| 数组模拟 | O(n) | O(n) | ✓ |

## 参考资料

- [LeetCode 146. LRU Cache](https://leetcode.com/problems/lru-cache/)
- [Groupcache](https://github.com/golang/groupcache) - Google 的分布式缓存库
- [Ristretto](https://github.com/dgraph-io/ristretto) - 高性能缓存库

## 学习总结

### 关键要点
- ...

### 性能优化技巧
- ...

### 实际应用场景
- ...

# GCD 权重轮换算法详解

> 版本2：加权轮询负载均衡的 GCD 优化实现

---

## 📝 算法概述

**GCD权重轮换法**是加权轮询算法的一种优化实现，通过使用**最大公约数（GCD）**作为递减步长，避免了扩展列表法的内存浪费。

**核心思想**：
- 维护一个动态递减的**权重阈值** `currentWeight`
- 只选择 `weight >= currentWeight` 的服务器
- 通过轮换降低阈值，实现按权重比例分配

---

## 🔑 核心变量

```go
type WeightedRoundRobinBalancer struct {
    servers       []*WeightedServer  // 服务器列表
    currentIndex  int                // 当前检查的服务器索引
    currentWeight int                // 当前权重阈值（动态变化）
    maxWeight     int                // 最大权重值
    gcdWeight     int                // 权重的最大公约数（递减步长）
}
```

| 变量 | 作用 | 示例值 |
|------|------|--------|
| `currentWeight` | 权重阈值，从 maxWeight 开始递减 | 5 → 4 → 3 → 2 → 1 → 5 ... |
| `currentIndex` | 当前检查的服务器索引 | 0 → 1 → 2 → 0 → 1 ... |
| `maxWeight` | 所有服务器中的最大权重 | 5 (来自权重 [5,1,1]) |
| `gcdWeight` | GCD，用作递减步长 | 1 (gcd(5,1,1) = 1) |

---

## 🎯 为什么需要 GCD？

### 问题场景

**权重 [10, 5, 5]**：
- 如果每次递减 1：需要遍历 10 次才完成一个周期
- 如果每次递减 5（GCD）：只需遍历 2 次！

### GCD 的作用

GCD 是**最优递减步长**，避免无效遍历。

```
权重 [10, 5, 5]，GCD = 5

使用步长 1（慢）：
  currentWeight: 10 → 9 → 8 → 7 → 6 → 5 → 4 → 3 → 2 → 1
  有效值：      10          5                   （大部分无效）

使用步长 5（快）：
  currentWeight: 10 → 5
  有效值：      10    5                          （全部有效）
```

**结论**：GCD 让算法跳过无效的中间值，直达关键阈值。

---

## 📊 算法执行流程图

```
┌─────────────────────────────────────────────────────────────┐
│                    收到新请求                                │
└─────────────────────────────────────────────────────────────┘
                            ↓
                ┌──────────────────────────┐
                │ currentIndex++           │
                │ (移动到下一个服务器索引)  │
                └──────────────────────────┘
                            ↓
                ┌─────────────────────────┐
                │ currentIndex %= n       │
                │ (循环回绕到 0)          │
                └─────────────────────────┘
                            ↓
                    currentIndex == 0？
                    ┌───────┴──────┐
                   是│             │否
                    ↓              ↓
        ┌──────────────────────┐  跳过
        │ currentWeight        │
        │    -= gcdWeight      │
        └──────────────────────┘
                    ↓
            currentWeight <= 0？
            ┌───────┴──────┐
           是│             │否
            ↓              ↓
┌──────────────────────┐  跳过
│ currentWeight =      │
│    maxWeight         │
└──────────────────────┘
            │
            └──────────┬────────────┘
                       ↓
        ┌──────────────────────────────────┐
        │ 检查：                            │
        │ servers[currentIndex].Weight     │
        │    >= currentWeight？            │
        └──────────────────────────────────┘
                ┌───────┴──────┐
               是│             │否
                ↓              ↓
        ┌──────────────┐  ┌──────────┐
        │ 返回该服务器  │  │ 继续循环  │
        │ (找到了！)   │  │ (不满足)  │
        └──────────────┘  └──────────┘
```

---

## 🔍 详细执行过程

### 初始配置

```
服务器列表:
  Server-1: weight = 5
  Server-2: weight = 1
  Server-3: weight = 1

计算结果:
  maxWeight = max(5, 1, 1) = 5
  gcdWeight = gcd(5, 1, 1) = 1

初始状态:
  currentIndex  = 0
  currentWeight = 0
```

### 请求 #1 的完整过程

```
┌───────────────────────────────────────────────────────────────┐
│ 请求 #1 到达                                                   │
└───────────────────────────────────────────────────────────────┘

【循环第1次】
  ├─ currentIndex++ → 1
  ├─ 回到 0？否
  ├─ 检查: Server-2 (weight=1) >= currentWeight(0)？
  │         1 >= 0？否（为什么？因为初始值特殊）
  └─ 继续...

【循环第2次】
  ├─ currentIndex++ → 2
  ├─ 回到 0？否
  ├─ 检查: Server-3 (weight=1) >= currentWeight(0)？
  │         1 >= 0？否
  └─ 继续...

【循环第3次】
  ├─ currentIndex++ → 0
  ├─ 回到 0？**是**
  │   ├─ currentWeight -= gcdWeight
  │   │     0 - 1 = -1
  │   └─ currentWeight <= 0？是
  │       └─ currentWeight = maxWeight = 5  ← 重置！
  │
  ├─ 检查: Server-1 (weight=5) >= currentWeight(5)？
  │         5 >= 5？**是** ✓
  └─ **返回 Server-1**

状态更新:
  currentIndex  = 0
  currentWeight = 5
```

### 请求 #2-#7 的追踪表

| 请求 | 开始状态 | 循环过程 | currentWeight 变化 | 选中服务器 | 结束状态 |
|------|---------|---------|-------------------|-----------|---------|
| **#1** | index=0, cw=0 | 3次循环 | 0 → -1 → 5 | Server-1 | index=0, cw=5 |
| **#2** | index=0, cw=5 | 3次循环 | 5 → 4 | Server-1 | index=0, cw=4 |
| **#3** | index=0, cw=4 | 3次循环 | 4 → 3 | Server-1 | index=0, cw=3 |
| **#4** | index=0, cw=3 | 3次循环 | 3 → 2 | Server-1 | index=0, cw=2 |
| **#5** | index=0, cw=2 | 3次循环 | 2 → 1 | Server-1 | index=0, cw=1 |
| **#6** | index=0, cw=1 | 1次循环 | 不变 | Server-2 | index=1, cw=1 |
| **#7** | index=1, cw=1 | 1次循环 | 不变 | Server-3 | index=2, cw=1 |

### currentWeight 的周期性变化

```
时间轴（每个 | 代表一次选择）

currentWeight:
  5  |  4  |  3  |  2  |  1  |  1  |  1  |  5  |  4  | ...
  ↓     ↓     ↓     ↓     ↓     ↓     ↓     ↓     ↓
  A     A     A     A     A     B     C     A     A  ...

周期: └────────── 一个完整周期（7次选择）──────────┘
                  A被选5次，B和C各1次
```

---

## 💡 关键理解点

### 1️⃣ 为什么能按权重分配？

**魔法在于阈值过滤**：

```
currentWeight = 5 时：
  ✓ Server-1 (weight=5): 5 >= 5 满足
  ✗ Server-2 (weight=1): 1 >= 5 不满足
  ✗ Server-3 (weight=1): 1 >= 5 不满足
  → 只有 Server-1 可以被选中

currentWeight = 4 时：
  ✓ Server-1 (weight=5): 5 >= 4 满足
  ✗ Server-2 (weight=1): 1 >= 4 不满足
  ✗ Server-3 (weight=1): 1 >= 4 不满足
  → 只有 Server-1 可以被选中

currentWeight = 1 时：
  ✓ Server-1 (weight=5): 5 >= 1 满足
  ✓ Server-2 (weight=1): 1 >= 1 满足
  ✓ Server-3 (weight=1): 1 >= 1 满足
  → 所有服务器都可以被选中
```

**结论**：
- 在一个周期内（currentWeight 从 5 降到 1）
- Server-1 在所有 5 个阈值下都满足条件 → 被选 5 次
- Server-2 和 Server-3 只在阈值为 1 时满足 → 各被选 1 次
- **比例**：5:1:1 ✓

### 2️⃣ 为什么需要循环？

```go
for {
    currentIndex = (currentIndex + 1) % len(servers)

    if currentIndex == 0 {
        currentWeight -= gcdWeight
        if currentWeight <= 0 {
            currentWeight = maxWeight
        }
    }

    if servers[currentIndex].Weight >= currentWeight {
        return servers[currentIndex]  // 找到了！
    }
    // 继续循环...
}
```

**原因**：不是每个服务器都满足 `weight >= currentWeight`

**示例**：
- currentWeight = 5，遍历 Server-1 时：5 >= 5 ✓ 立即返回
- currentWeight = 5，遍历 Server-2 时：1 >= 5 ✗ 继续
- currentWeight = 5，遍历 Server-3 时：1 >= 5 ✗ 继续
- 回到 Server-1，再检查...

### 3️⃣ 为什么从 currentIndex++ 开始？

**保证公平性**：

```
如果从 currentIndex 开始检查（错误）:
  → 每次都从同一个位置开始
  → Server-1 总是第一个被检查
  → 偏向性！

如果先 currentIndex++（正确）:
  → 每次从不同位置开始
  → 轮流给每个服务器机会
  → 公平！
```

---

## 📈 性能分析

### 时间复杂度

**最坏情况**：O(n)
- 遍历所有 n 个服务器才找到满足条件的
- 例如：currentWeight = 1，前面所有服务器 weight = 0

**平均情况**：O(1) 均摊
- 高权重服务器通常很快被选中
- 例如：currentWeight = 5，Server-1 第一次就满足

**每个周期**：O(总权重)
- 权重 5:1:1，周期内选择 7 次
- 每次平均循环 1-2 次
- 总复杂度 ≈ O(7) = O(总权重)

### 空间复杂度

**O(1)**：只使用了几个变量
- currentIndex
- currentWeight
- maxWeight
- gcdWeight

**对比版本1（扩展列表法）**：
- 版本1：O(总权重)，权重 1000:1:1 需要 1002 个元素
- 版本2：O(1)，只需 4 个变量

---

## ⚖️ 优缺点对比

### ✅ 优点

1. **节省内存**
   - 不需要扩展列表
   - 权重 1000:1 时，版本1需要 1001 个元素，版本2只需 4 个变量

2. **支持大权重**
   - 可以处理 1000:1 这样的极端权重比例
   - 版本1会因内存爆炸而不可行

3. **实现相对简洁**
   - 代码量适中（约 30 行）
   - 核心逻辑清晰

### ❌ 缺点

1. **不够平滑**
   - 输出：`A A A A A B C A A A A A B C ...`
   - 连续 5 个 A，突发流量！

2. **算法复杂度较高**
   - 需要理解 GCD 概念
   - 需要理解权重阈值递减逻辑

3. **有无效遍历**
   - currentWeight = 5 时，仍然需要检查 Server-2 和 Server-3（尽管它们不满足）
   - 性能开销

---

## 🆚 与其他版本对比

### 版本1：扩展列表法

```go
// 扩展列表
expandedList = [A, A, A, A, A, B, C]
current = (current + 1) % len(expandedList)
```

**优点**：
- ✅ 简单易懂
- ✅ O(1) 严格时间复杂度

**缺点**：
- ❌ 内存浪费：O(总权重)
- ❌ 不平滑：连续 5 个 A

### 版本2：GCD权重轮换（当前版本）

```go
// 权重阈值递减
currentWeight: 5 → 4 → 3 → 2 → 1 → 5 ...
选择: weight >= currentWeight
```

**优点**：
- ✅ 内存高效：O(1)
- ✅ 支持大权重

**缺点**：
- ❌ 不平滑：连续 5 个 A
- ❌ 算法复杂

### 版本3：NGINX平滑加权轮询

```go
// 动态调整 currentWeight
每次: currentWeight += weight
选择最大 currentWeight 的服务器
被选中: currentWeight -= 总权重
```

**优点**：
- ✅ 平滑：`A A B A C A A`
- ✅ 代码最简洁（20 行）
- ✅ O(1) 内存

**缺点**：
- ⚠️ 需要理解动态权重概念

---

## 🎓 学习建议

### 理解路径

```
1. 先理解版本1（扩展列表）
   ↓ 理解"权重 = 副本数"

2. 发现问题：内存浪费
   ↓ 思考：能否不扩展列表？

3. 学习版本2（GCD轮换）
   ↓ 理解"阈值过滤"思想

4. 发现问题：不平滑
   ↓ 思考：能否分散选择？

5. 学习版本3（NGINX平滑）
   ↓ 理解"动态调整"思想
```

### 手动模拟

**建议**：用纸笔模拟请求 #1-#7 的完整过程

1. 画出状态表：
   ```
   | 请求 | currentIndex | currentWeight | 选择 |
   |------|-------------|---------------|------|
   | #1   | 0 → 0       | 0 → 5         | A    |
   | #2   | 0 → 0       | 5 → 4         | A    |
   ...
   ```

2. 观察规律：
   - currentWeight 如何变化？
   - 为什么 A 连续出现？
   - 何时选择 B 和 C？

### 代码实现

1. **先实现最简版本**：
   ```go
   // 固定权重 [5, 1, 1]，不考虑 GCD
   maxWeight = 5
   gcdWeight = 1  // 硬编码
   ```

2. **添加 GCD 计算**：
   ```go
   func gcd(a, b int) int {
       for b != 0 {
           a, b = b, a%b
       }
       return a
   }
   ```

3. **测试不同权重**：
   - [5, 1, 1]
   - [10, 5, 5]
   - [3, 2, 1]

---

## 🧪 测试用例

```go
func TestGCDWeightedRR(t *testing.T) {
    servers := []*WeightedServer{
        {Name: "A", Weight: 5},
        {Name: "B", Weight: 1},
        {Name: "C", Weight: 1},
    }

    balancer := NewWeightedRR(servers)

    // 测试一个完整周期（7次）
    expected := []string{"A", "A", "A", "A", "A", "B", "C"}

    for i, want := range expected {
        got := balancer.NextServer().Name
        if got != want {
            t.Errorf("request #%d: got %s, want %s", i+1, got, want)
        }
    }

    // 验证周期性
    for i := 0; i < 7; i++ {
        got := balancer.NextServer().Name
        want := expected[i]
        if got != want {
            t.Errorf("second cycle request #%d: got %s, want %s", i+1, got, want)
        }
    }
}
```

---

## 📚 总结

### 核心思想

**阈值过滤 + 周期递减 = 按权重分配**

### 关键公式

```
每个周期的选择次数 = 满足条件的阈值数量

Server-1 (weight=5):
  满足 currentWeight = [5, 4, 3, 2, 1]
  → 选择 5 次

Server-2 (weight=1):
  满足 currentWeight = [1]
  → 选择 1 次
```

### 为什么重要？

1. **理解演进过程**：从简单到复杂的优化思路
2. **算法设计思想**：用阈值代替扩展，用循环代替存储
3. **权衡取舍**：性能 vs 简单性 vs 平滑性

### 下一步

学习版本3（NGINX平滑加权轮询），理解如何解决"不平滑"问题！

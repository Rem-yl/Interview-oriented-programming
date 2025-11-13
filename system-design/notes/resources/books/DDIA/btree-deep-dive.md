# B-Tree 深度解析

> 从二叉搜索树到 B-Tree，理解数据库索引的基石

## 目录

- [为什么需要 B-Tree？](#为什么需要-b-tree)
- [B-Tree 核心概念](#b-tree-核心概念)
- [查找操作详解](#查找操作详解)
- [插入与分裂](#插入与分裂)
- [删除与合并](#删除与合并)
- [B-Tree 可靠性保障](#b-tree-可靠性保障)
- [B-Tree 优化技术](#b-tree-优化技术)
- [B+Tree 变种](#btree-变种)
- [LSM-Tree vs B-Tree 深度对比](#lsm-tree-vs-b-tree-深度对比)
- [真实系统案例](#真实系统案例)

---

## 为什么需要 B-Tree？

### 从二叉搜索树说起

**二叉搜索树（BST）的问题**：

```
理想情况（平衡）：
       50
      /  \
    30    70
   /  \  /  \
  20  40 60  80

查找深度：log₂(n)
n=100万，深度 ≈ 20

但磁盘 I/O 是昂贵的！
每层需要一次磁盘 I/O
20 次 I/O × 10ms = 200ms  ← 太慢！
```

```
退化情况（不平衡）：
1
 \
  2
   \
    3
     \
      4
       \
        ...

查找深度：O(n)  ← 完全不可接受！
```

### 为什么不用红黑树/AVL树？

**问题 1：深度太大**

```
红黑树/AVL树（二叉树）：
节点数 n=100万
深度 = log₂(1,000,000) ≈ 20

每次查找：20 次磁盘 I/O
延迟：20 × 10ms = 200ms
```

**问题 2：磁盘 I/O 特性不匹配**

```
磁盘读取特性：
- 最小读取单位：扇区（512 字节）或页（4KB）
- 读 4 字节和读 4KB 时间相同
- 顺序读远快于随机读

二叉树节点：
struct Node {
    int key;      // 4 字节
    void* left;   // 8 字节
    void* right;  // 8 字节
}  // 总共 20 字节

浪费：读取 4KB，只用 20 字节！
利用率 = 20/4096 ≈ 0.5%  ← 极低！
```

### B-Tree 的设计哲学

**核心思想**：

1. **增加分支因子**（每个节点多个键）
2. **与磁盘页对齐**（节点大小 = 4KB）
3. **保持平衡**（所有叶子深度相同）

```
B-Tree 节点（4KB 页）：
┌────────────────────────────────────────────┐
│ Key1 | Key2 | ... | Key100                 │  ← 可以存 100 个键！
│ Ptr1 | Ptr2 | ... | Ptr101                 │  ← 101 个子指针
└────────────────────────────────────────────┘

分支因子 b ≈ 100

查找深度：
n=100万，深度 = log₁₀₀(1,000,000) ≈ 3
仅需 3 次磁盘 I/O！

n=10亿，深度 = log₁₀₀(1,000,000,000) ≈ 4-5
仅需 4-5 次磁盘 I/O！
```

**与 LSM-Tree 的根本差异**：

```
LSM-Tree 哲学：
- 写入优化（追加写，顺序 I/O）
- 读取需要查找多个 SSTable
- 后台压缩清理旧数据

B-Tree 哲学：
- 读取优化（原地更新，直接定位）
- 写入需要随机 I/O
- 使用 WAL 保证可靠性
```

---

## B-Tree 核心概念

### 定义与性质

**B-Tree of order m**（m 阶 B-Tree）：

```
性质：
1. 每个节点最多 m 个子节点
2. 每个非叶节点（除根）至少 ⌈m/2⌉ 个子节点
3. 根节点至少 2 个子节点（除非是叶节点）
4. 所有叶节点在同一层（完美平衡）
5. 有 k 个子节点的节点包含 k-1 个键

常见配置：
- m = 100-200（4KB 页）
- m = 500-1000（16KB 页）
```

### 节点结构

**内部节点（Internal Node）**：

```
┌─────────────────────────────────────────────────────────┐
│  n (key count) | isLeaf (false)                         │
├─────────────────────────────────────────────────────────┤
│  Key[0] | Key[1] | Key[2] | ... | Key[n-1]              │
├─────────────────────────────────────────────────────────┤
│  Ptr[0] | Ptr[1] | Ptr[2] | ... | Ptr[n]                │
└─────────────────────────────────────────────────────────┘

键的排序：Key[0] < Key[1] < Key[2] < ... < Key[n-1]

子树范围：
Ptr[0] → 键 < Key[0]
Ptr[1] → Key[0] ≤ 键 < Key[1]
Ptr[2] → Key[1] ≤ 键 < Key[2]
...
Ptr[n] → 键 ≥ Key[n-1]
```

**叶节点（Leaf Node）**：

```
┌─────────────────────────────────────────────────────────┐
│  n (key count) | isLeaf (true)                          │
├─────────────────────────────────────────────────────────┤
│  Key[0] | Key[1] | Key[2] | ... | Key[n-1]              │
├─────────────────────────────────────────────────────────┤
│  Val[0] | Val[1] | Val[2] | ... | Val[n-1]              │
├─────────────────────────────────────────────────────────┤
│  Next Leaf Pointer (optional, for B+Tree)               │
└─────────────────────────────────────────────────────────┘
```

### 示例 B-Tree

**5 阶 B-Tree（m=5，每个节点最多 4 个键）**：

```
                      [50]
                    /      \
              [20|30]      [70|80]
             /   |   \    /   |   \
        [10] [25] [35] [60] [75] [90]
```

**详细标注**：

```
根节点 [50]:
- 1 个键：50
- 2 个子指针：left (<50), right (≥50)

内部节点 [20|30]:
- 2 个键：20, 30
- 3 个子指针：
  Ptr[0] → 键 < 20 → [10]
  Ptr[1] → 20 ≤ 键 < 30 → [25]
  Ptr[2] → 键 ≥ 30 → [35]

叶节点 [10]:
- 1 个键值对：(10, value)
```

### 深度与容量分析

**深度公式**：

```
深度 h = ⌈log_⌈m/2⌉(n)⌉

示例：
m = 100（每个节点最多 99 个键）
最小分支 = ⌈100/2⌉ = 50

n = 100万条记录
h = log₅₀(1,000,000) ≈ 3.5 ≈ 4 层

n = 10亿条记录
h = log₅₀(1,000,000,000) ≈ 5.3 ≈ 6 层
```

**容量分析**：

```
m = 100 的 B-Tree：

层数 1（根）：
最少：1 个节点，50 个键
最多：1 个节点，99 个键

层数 2：
最少：50 个子节点，50×50 = 2,500 个键
最多：100 个子节点，100×99 = 9,900 个键

层数 3：
最少：50² = 2,500 个子节点，125,000 个键
最多：100² = 10,000 个子节点，990,000 个键

层数 4：
最少：50³ = 125,000 个子节点，6,250,000 个键
最多：100³ = 1,000,000 个子节点，99,000,000 个键

结论：
3 层可存储：12.5万 ~ 100万 条记录
4 层可存储：625万 ~ 1亿 条记录
5 层可存储：3.1亿 ~ 100亿 条记录
```

**关键洞察**：

> B-Tree 深度增长极慢！即使数据量从百万级增长到亿级，深度只增加 1-2 层。

---

## 查找操作详解

### 查找算法

**伪代码**：

```python
def search(node, key):
    """
    在 B-Tree 中查找键
    返回：(value, True) 如果找到，否则 (None, False)
    """
    # 在当前节点中二分查找
    i = binary_search(node.keys, key)

    if i < node.n and node.keys[i] == key:
        # 找到了
        if node.is_leaf:
            return (node.values[i], True)
        else:
            # 内部节点也可能存值（B-Tree），或继续向下（B+Tree）
            return (node.values[i], True)

    if node.is_leaf:
        # 叶节点未找到
        return (None, False)

    # 递归到子节点
    # 读取子节点（磁盘 I/O）
    child = read_page(node.children[i])
    return search(child, key)


def binary_search(keys, target):
    """
    二分查找，返回第一个 >= target 的索引
    """
    left, right = 0, len(keys)
    while left < right:
        mid = (left + right) // 2
        if keys[mid] < target:
            left = mid + 1
        else:
            right = mid
    return left
```

### 查找示例

**查找 key=35**：

```
                      [50]                    ← 步骤 1：读取根节点
                    /      \                     35 < 50，走左子树
              [20|30]      [70|80]            ← 步骤 2：读取左子节点
             /   |   \    /   |   \              30 < 35 < 50，走中间
        [10] [25] [35] [60] [75] [90]         ← 步骤 3：读取叶节点
                   ↑                             找到 35！
                  找到

磁盘 I/O 次数：3 次
总延迟：3 × 10ms = 30ms
```

**查找 key=45（不存在）**：

```
                      [50]                    ← 步骤 1：35 < 50，左
              [20|30]      [70|80]            ← 步骤 2：30 < 45 < 50，中间
             /   |   \
        [10] [25] [35] [60]                   ← 步骤 3：35 < 45，但已是叶节点
                   ↑                             未找到
                  这里

磁盘 I/O 次数：3 次
结果：未找到
```

### 性能分析

**时间复杂度**：

```
磁盘 I/O 次数：O(log_m n)
- m = 分支因子（通常 100-1000）
- n = 记录数

每个节点内二分查找：O(log m)
- m 通常很小（<1000），可视为 O(1)

总时间：O(log_m n) 次磁盘 I/O
```

**与 LSM-Tree 对比**：

```
B-Tree 查找：
- 最好：O(log_m n) 次 I/O（直接定位）
- 最坏：O(log_m n) 次 I/O（稳定）
- 示例：3-5 次 I/O（百万到亿级数据）

LSM-Tree 查找：
- 最好：O(1)（在 MemTable）
- 最坏：O(层数 × 每层 SSTable 数)
- 示例：
  - Leveled Compaction：~层数次 I/O（4-6 次）
  - Size-Tiered Compaction：~10-20 次 I/O
  - 需要 Bloom Filter 优化
```

---

## 插入与分裂

### 正常插入（节点未满）

**示例：插入 key=32**

```
初始状态：
                      [50]
                    /      \
              [20|30]      [70|80]
             /   |   \    /   |   \
        [10] [25] [35] [60] [75] [90]

步骤 1：查找插入位置
- 32 < 50 → 左子树
- 30 < 32 < 50 → 中间子树
- 到达叶节点 [35]

步骤 2：插入到叶节点
[35] → [32|35]

结果：
                      [50]
                    /      \
              [20|30]      [70|80]
             /   |   \    /   |   \
        [10] [25] [32|35] [60] [75] [90]
                  ↑ 插入成功
```

### 叶节点分裂

**示例：插入 key=33（叶节点已满）**

```
假设 m=5（每个节点最多 4 个键）

初始状态（叶节点已满）：
              [20|30|40]
             /    |    \    \
        [10] [25] [32|35|38] [50]
                   ↑ 已有 3 个键，m=5 时已满

插入 33：
步骤 1：叶节点满，需要分裂
临时节点：[32|33|35|38]（4 个键）

步骤 2：分裂
中间位置 = ⌈4/2⌉ = 2
左节点：[32|33]（前 2 个）
右节点：[35|38]（后 2 个）
提升键：35（提升到父节点）

步骤 3：更新父节点
              [20|30|35|40]  ← 插入 35
             /    |    |   \    \
        [10] [25] [32|33] [35|38] [50]
                   ↑新      ↑新
```

### 内部节点分裂

**示例：父节点也满，需要递归分裂**

```
初始状态（m=5，父节点已满）：
              [20|30|35|40]  ← 4 个键，已满
             /    |   |   \    \
        [10] [25] [32] [36] [50]

插入 37：
步骤 1：37 在 [36] 叶节点，插入后 [36|37]，未满，OK

但假设我们再插入 38：
[36|37|38]，再插入 39：
[36|37|38|39]，再插入 34：

叶节点 [36|37|38|39] 需要分裂：
左：[36|37]
右：[38|39]
提升：38

父节点插入 38：
[20|30|35|38|40]  ← 5 个键，超过 m-1=4！

步骤 2：父节点分裂
临时：[20|30|35|38|40]
中间：35（索引 2）
左父：[20|30]
右父：[38|40]
提升到祖父：35

如果没有祖父（是根）：
创建新根 [35]
       /     \
  [20|30]   [38|40]

树高度增加 1！
```

### 完整分裂流程图

```
插入过程（自顶向下查找，自底向上分裂）：

1. 从根开始，向下查找插入位置
   ↓
2. 到达叶节点，尝试插入
   ↓
3. 叶节点满？
   NO → 直接插入，结束
   YES ↓
4. 分裂叶节点
   - 创建右兄弟节点
   - 移动一半键到右兄弟
   - 提升中间键到父节点
   ↓
5. 父节点满？
   NO → 插入提升的键，结束
   YES ↓
6. 递归分裂父节点
   ↓
7. 直到根节点
   - 如果根分裂，创建新根
   - 树高度 +1
```

### 分裂的代码实现

```python
def insert(tree, key, value):
    """
    插入键值对，可能触发分裂
    """
    root = tree.root

    # 情况 1：根节点满，需要预先分裂
    if is_full(root):
        new_root = create_node(is_leaf=False)
        new_root.children[0] = root
        split_child(new_root, 0)  # 分裂旧根
        tree.root = new_root
        # 树高度 +1

    # 插入到非满的树中
    insert_non_full(tree.root, key, value)


def insert_non_full(node, key, value):
    """
    插入到保证不满的节点（可能递归）
    """
    i = node.n - 1  # 最后一个键的索引

    if node.is_leaf:
        # 叶节点：直接插入
        # 找到插入位置（保持有序）
        while i >= 0 and key < node.keys[i]:
            node.keys[i + 1] = node.keys[i]
            node.values[i + 1] = node.values[i]
            i -= 1

        node.keys[i + 1] = key
        node.values[i + 1] = value
        node.n += 1

        # 写回磁盘
        write_page(node)
    else:
        # 内部节点：找到子节点
        while i >= 0 and key < node.keys[i]:
            i -= 1
        i += 1  # 子节点索引

        # 读取子节点
        child = read_page(node.children[i])

        # 如果子节点满，预先分裂
        if is_full(child):
            split_child(node, i)
            # 分裂后，node.keys[i] 是新的分割键
            if key > node.keys[i]:
                i += 1  # 插入到右子节点
            child = read_page(node.children[i])

        # 递归插入
        insert_non_full(child, key, value)


def split_child(parent, index):
    """
    分裂 parent 的第 index 个子节点
    """
    full_child = read_page(parent.children[index])
    new_child = create_node(is_leaf=full_child.is_leaf)

    # 假设 m=5，t=⌈m/2⌉=3（最小度数）
    t = TREE_ORDER // 2

    # 移动后半部分键到新节点
    new_child.keys = full_child.keys[t:]
    new_child.values = full_child.values[t:]
    new_child.n = len(new_child.keys)

    if not full_child.is_leaf:
        # 移动子指针
        new_child.children = full_child.children[t:]

    # 中间键提升到父节点
    median_key = full_child.keys[t - 1]

    # 截断原节点
    full_child.keys = full_child.keys[:t - 1]
    full_child.values = full_child.values[:t - 1]
    full_child.n = t - 1

    # 在父节点中插入中间键和新子节点
    parent.keys.insert(index, median_key)
    parent.children.insert(index + 1, new_child.page_id)
    parent.n += 1

    # 写回磁盘
    write_page(full_child)
    write_page(new_child)
    write_page(parent)


def is_full(node):
    """节点是否已满"""
    return node.n >= TREE_ORDER - 1
```

### 插入性能分析

**时间复杂度**：

```
查找路径：O(log_m n) 次磁盘 I/O（向下查找）

分裂次数：
- 最坏情况：从叶到根所有节点都满，全部分裂
- 分裂次数 = 树高度 = O(log_m n)
- 每次分裂：O(m) 次内存操作 + 3 次磁盘写入
  - 写旧节点
  - 写新节点
  - 写父节点

总磁盘 I/O：
- 读：O(log_m n)
- 写：最坏 O(log_m n)，平均 O(1)

实际优化：
- 预分配缓冲区，减少分裂
- 延迟分裂（lazy split）
```

---

## 删除与合并

### 删除操作的三种情况

**情况 1：从叶节点删除（简单）**

```
删除 key=25：

初始：
              [20|30]
             /   |   \
        [10] [25] [35]

结果：
              [20|30]
             /   |   \
        [10] []   [35]  ← [25] 变空？

问题：违反 B-Tree 性质（最少 ⌈m/2⌉-1 个键）
```

**情况 2：从内部节点删除**

```
删除 key=30：

初始：
              [20|30|40]
             /    |    \    \
        [10] [25] [35] [50]

30 在内部节点！

解决方案：
方案 A：找左子树最大值（前驱）替换
  - 左子树 [25] 的最大值 = 25
  - 用 25 替换 30
  - 递归删除 25

方案 B：找右子树最小值（后继）替换
  - 右子树 [35] 的最小值 = 35
  - 用 35 替换 30
  - 递归删除 35

结果（使用方案 B）：
              [20|35|40]
             /    |    \    \
        [10] [25] []   [50]
```

**情况 3：节点下溢（underflow）需要合并**

```
删除导致节点键数 < ⌈m/2⌉-1

m=5，最少 ⌈5/2⌉-1 = 2-1 = 1 个键

删除后 [25] 变 []，违反性质！

解决方案 A：从兄弟借键（rotation）
条件：兄弟节点有富余键

左兄弟 [10] 只有 1 个键，无富余
右兄弟 [35] 只有 1 个键，无富余

解决方案 B：与兄弟合并（merge）
合并 [] 和右兄弟 [35]，加上父节点的分隔键 30

合并：[] + 30 + [35] = [30|35]

父节点删除 30：
[20|30|40] → [20|40]

结果：
              [20|40]
             /    |    \
        [10] [30|35] [50]
```

### 借键（Rotation）示例

**向左兄弟借键**：

```
初始状态（m=5）：
              [20|30]
             /   |   \
      [10|15] [25] [35|40]
                ↑ 将要删除 25，删除后为空

左兄弟 [10|15] 有 2 个键，富余！

步骤：
1. 父节点的分隔键 20 下降到当前节点
2. 左兄弟最大键 15 上升到父节点

结果：
              [15|30]
             /   |   \
        [10] [20] [35|40]
              ↑ 25 被 20 替换
```

**详细过程**：

```
向右旋转（left sibling → parent → current）：

    Parent: [20|30]
           /   |   \
  Left: [10|15]  Current: [25]  Right: [35|40]

删除 25：
Current 变空，从 Left 借

1. Left.max (15) 上升到 Parent
2. Parent 的分隔键 (20) 下降到 Current

    Parent: [15|30]
           /   |   \
  Left: [10]  Current: [20]  Right: [35|40]
```

### 合并（Merge）示例

**合并兄弟节点**：

```
初始状态（m=5）：
              [20|30|40]
             /    |    \    \
        [10] [25] [35] [50]

删除 10，删除 25：

[10] 删除后为空
[25] 删除后为空

两个节点都下溢！

步骤 1：合并 [] 和 [25]
[] + 分隔键20 + [25] = [20|25]

              [30|40]
             /    |    \
      [20|25] [35] [50]

步骤 2：再删除 25
[20|25] → [20]（仍满足最少 1 个键）

步骤 3：继续删除 20
[20] → []（下溢）

合并 [] 和 [35]：
[] + 30 + [35] = [30|35]

              [40]
             /    \
      [30|35]  [50]

如果继续删除，根节点可能只剩 1 个键
如果根只有 1 个子节点 → 删除根，树高度 -1
```

### 删除算法

```python
def delete(tree, key):
    """
    从 B-Tree 删除键
    """
    root = tree.root
    delete_from_node(root, key)

    # 如果根节点为空且有子节点，降低树高度
    if root.n == 0 and not root.is_leaf:
        tree.root = read_page(root.children[0])
        # 树高度 -1


def delete_from_node(node, key):
    """
    从节点删除键（可能递归）
    """
    i = binary_search(node.keys, key)

    if i < node.n and node.keys[i] == key:
        # 找到了要删除的键
        if node.is_leaf:
            # 情况 1：叶节点，直接删除
            delete_from_leaf(node, i)
        else:
            # 情况 2：内部节点，用前驱或后继替换
            delete_from_internal(node, i)
    else:
        # 键不在当前节点
        if node.is_leaf:
            # 键不存在
            return

        # 递归到子节点
        is_in_last_child = (i == node.n)
        child = read_page(node.children[i])

        # 确保子节点有足够的键
        if child.n < TREE_ORDER // 2:
            fill_child(node, i)
            # fill 后可能改变子节点位置
            child = read_page(node.children[i])

        delete_from_node(child, key)


def delete_from_leaf(node, index):
    """从叶节点删除"""
    node.keys.pop(index)
    node.values.pop(index)
    node.n -= 1
    write_page(node)


def delete_from_internal(node, index):
    """从内部节点删除"""
    key = node.keys[index]

    left_child = read_page(node.children[index])
    right_child = read_page(node.children[index + 1])

    if left_child.n >= TREE_ORDER // 2:
        # 左子树有富余，用前驱替换
        predecessor = get_predecessor(left_child)
        node.keys[index] = predecessor.key
        node.values[index] = predecessor.value
        write_page(node)
        delete_from_node(left_child, predecessor.key)
    elif right_child.n >= TREE_ORDER // 2:
        # 右子树有富余，用后继替换
        successor = get_successor(right_child)
        node.keys[index] = successor.key
        node.values[index] = successor.value
        write_page(node)
        delete_from_node(right_child, successor.key)
    else:
        # 两边都不够，合并
        merge_children(node, index)
        delete_from_node(left_child, key)


def fill_child(parent, index):
    """
    确保子节点有足够的键（通过借或合并）
    """
    child = read_page(parent.children[index])

    # 尝试从左兄弟借
    if index > 0:
        left_sibling = read_page(parent.children[index - 1])
        if left_sibling.n >= TREE_ORDER // 2:
            borrow_from_left(parent, index)
            return

    # 尝试从右兄弟借
    if index < parent.n:
        right_sibling = read_page(parent.children[index + 1])
        if right_sibling.n >= TREE_ORDER // 2:
            borrow_from_right(parent, index)
            return

    # 无法借，只能合并
    if index > 0:
        merge_children(parent, index - 1)
    else:
        merge_children(parent, index)


def merge_children(parent, index):
    """
    合并 parent.children[index] 和 parent.children[index+1]
    """
    left_child = read_page(parent.children[index])
    right_child = read_page(parent.children[index + 1])

    # 左节点 + 父节点分隔键 + 右节点
    left_child.keys.append(parent.keys[index])
    left_child.keys.extend(right_child.keys)

    left_child.values.append(parent.values[index])
    left_child.values.extend(right_child.values)

    if not left_child.is_leaf:
        left_child.children.extend(right_child.children)

    left_child.n = len(left_child.keys)

    # 从父节点删除分隔键和右子节点
    parent.keys.pop(index)
    parent.values.pop(index)
    parent.children.pop(index + 1)
    parent.n -= 1

    # 写回
    write_page(left_child)
    delete_page(right_child)  # 删除右节点
    write_page(parent)
```

---

## B-Tree 可靠性保障

### 问题：崩溃一致性

**场景 1：插入时崩溃**

```
插入 key=33，导致节点分裂：

操作序列：
1. 写新右节点 [35|38]
2. 写旧左节点 [32|33]
3. 写父节点 [20|30|35|40]  ← 如果在这里崩溃？

结果：
- 新右节点已写入磁盘
- 父节点未更新
- 父节点仍然指向旧的未分裂节点
- 数据库损坏！
```

**场景 2：更新时崩溃**

```
更新 key=50 的值：

操作序列：
1. 读取叶节点页 [50:old_value]
2. 修改为 [50:new_value]
3. 写回磁盘  ← 如果写了一半崩溃？

结果：
- 页面部分写入（torn page）
- 数据损坏
```

### 解决方案：Write-Ahead Log (WAL)

**WAL 原理**：

```
核心规则：
1. 修改数据前，先写日志
2. 日志持久化后，才修改数据页
3. 事务提交前，日志必须落盘

操作顺序：
┌─────────────────────────────────────┐
│ 1. Begin Transaction                │
│    写 WAL: "BEGIN TXN 123"           │
├─────────────────────────────────────┤
│ 2. Modify Pages                     │
│    写 WAL: "UPDATE Page 10, key=50"  │
│    (数据页还在内存，未写磁盘)         │
├─────────────────────────────────────┤
│ 3. Commit                           │
│    写 WAL: "COMMIT TXN 123"          │
│    fsync(WAL)  ← 确保 WAL 落盘      │
│    返回成功给客户端                  │
├─────────────────────────────────────┤
│ 4. Checkpoint (后台异步)             │
│    将脏页写回磁盘                    │
│    写 WAL: "CHECKPOINT"              │
└─────────────────────────────────────┘
```

**WAL 日志格式**：

```
WAL Record:
┌────────────────────────────────────────────┐
│ LSN (Log Sequence Number)                  │  8 bytes
│ Transaction ID                             │  8 bytes
│ Record Type (INSERT/UPDATE/DELETE/COMMIT)  │  1 byte
│ Page ID                                    │  8 bytes
│ Offset                                     │  4 bytes
│ Before Image (old value)                   │  variable
│ After Image (new value)                    │  variable
│ CRC Checksum                               │  4 bytes
└────────────────────────────────────────────┘
```

**崩溃恢复流程**：

```python
def crash_recovery():
    """
    从 WAL 恢复数据库
    """
    wal = open_wal_file()
    last_checkpoint = find_last_checkpoint(wal)

    # Phase 1: Analysis
    # 扫描 WAL，找出未完成的事务和脏页
    active_txns = set()
    dirty_pages = set()

    for record in wal.scan_from(last_checkpoint):
        if record.type == BEGIN:
            active_txns.add(record.txn_id)
        elif record.type == COMMIT:
            active_txns.discard(record.txn_id)
        elif record.type in (INSERT, UPDATE, DELETE):
            dirty_pages.add(record.page_id)

    # Phase 2: Redo
    # 重放所有已提交事务的操作
    for record in wal.scan_from(last_checkpoint):
        if record.type in (INSERT, UPDATE, DELETE):
            if record.txn_id not in active_txns:  # 已提交
                # 重做操作
                page = read_page(record.page_id)
                apply_change(page, record.after_image)
                write_page(page)

    # Phase 3: Undo
    # 回滚未完成的事务
    for txn_id in active_txns:
        rollback_transaction(wal, txn_id)

    # 恢复完成
    truncate_wal(last_checkpoint)
```

### ARIES 恢复算法

**ARIES = Algorithms for Recovery and Isolation Exploiting Semantics**

**核心思想**：

1. **WAL（Write-Ahead Logging）**：先写日志
2. **Redo**：重做所有历史操作（已提交和未提交）
3. **Undo**：回滚未提交事务

**为什么 Redo 未提交的事务？**

```
简化恢复逻辑！

传统方法（只 Redo 已提交）：
- 需要判断每条日志对应的事务是否提交
- 复杂，容易出错

ARIES 方法：
- Redo 阶段：无脑重放所有日志（幂等操作）
- Undo 阶段：回滚未提交事务

优势：
- 逻辑简单
- Redo 可以并行
- 支持模糊检查点（fuzzy checkpoint）
```

**示例**：

```
WAL 日志：
LSN 100: BEGIN TXN 1
LSN 101: UPDATE Page 5, key=10, old=100, new=200  (TXN 1)
LSN 102: BEGIN TXN 2
LSN 103: UPDATE Page 6, key=20, old=300, new=400  (TXN 2)
LSN 104: COMMIT TXN 1
LSN 105: UPDATE Page 7, key=30, old=500, new=600  (TXN 2)
LSN 106: [CRASH]  ← 在这里崩溃

恢复：
Redo 阶段（重放所有）：
- LSN 101: Page 5, key=10 = 200  ✓
- LSN 103: Page 6, key=20 = 400  ✓
- LSN 105: Page 7, key=30 = 600  ✓

Undo 阶段（回滚 TXN 2）：
- LSN 105: Page 7, key=30 = 500  (恢复 old value)
- LSN 103: Page 6, key=20 = 300  (恢复 old value)

最终状态：
- TXN 1 的修改保留（已提交）
- TXN 2 的修改回滚（未提交）
```

---

## 并发控制

### 问题：多线程并发修改

```
线程 1：插入 key=50
线程 2：插入 key=60

都需要修改同一个叶节点 [40|70]

无锁情况：
线程 1 读取 [40|70]
线程 2 读取 [40|70]
线程 1 写入 [40|50|70]
线程 2 写入 [40|60|70]  ← 覆盖了线程 1 的修改！

结果：key=50 丢失！
```

### 解决方案 1：页锁（Latch）

**Latch vs Lock**：

```
Latch（闩锁）：
- 保护内存数据结构（如 B-Tree 节点）
- 持有时间短（微秒级）
- 不支持死锁检测
- 类似 mutex / rwlock

Lock（锁）：
- 保护事务语义（如行锁、表锁）
- 持有时间长（事务持续时间）
- 支持死锁检测
- 用于隔离级别
```

**读写 Latch**：

```cpp
class BTreePage {
    RWLatch latch;  // 读写锁

    // 读操作
    Value get(Key key) {
        latch.read_lock();
        auto result = search_in_page(key);
        latch.read_unlock();
        return result;
    }

    // 写操作
    void insert(Key key, Value value) {
        latch.write_lock();
        insert_into_page(key, value);
        latch.write_unlock();
    }
};
```

### Latch Crabbing（锁蟹行）

**问题**：简单的锁整个路径会阻塞所有并发操作

**优化：逐级加锁，安全后释放父节点**

```
安全节点定义：
- 插入后不会分裂
- 删除后不会下溢

Latch Crabbing 规则：
1. 从根开始，加锁父节点
2. 加锁子节点
3. 如果子节点安全，释放所有祖先锁
4. 重复直到叶节点
```

**插入示例**：

```
插入 key=55（m=5，最多 4 个键）

树结构：
                 [50]           ← Root
               /      \
         [20|30]      [70|80]   ← 内部节点
        /   |   \    /   |   \
    [10] [25] [35] [60] [75] [90]  ← 叶节点

步骤 1：加锁根节点（写锁）
Root [50] - LOCKED

步骤 2：判断根节点安全？
- 只有 1 个键，插入不会分裂 → 安全
- 加锁右子节点 [70|80]

步骤 3：释放根锁（因为安全）
Root [50] - UNLOCKED
[70|80] - LOCKED

步骤 4：判断 [70|80] 安全？
- 有 2 个键，还能插 1 个不分裂（< 4） → 安全
- 加锁右子节点 [60]

步骤 5：释放 [70|80]（因为安全）
[70|80] - UNLOCKED
[60] - LOCKED

步骤 6：插入到 [60]
[60] → [55|60]
[60] - UNLOCKED

并发性：
- 其他线程可以在释放祖先锁后立即访问
- 只有最后的叶节点被短暂锁定
```

**分裂示例**：

```
插入 key=77（导致分裂）

叶节点 [75] 已有 3 个键：[75|76|78]（假设 m=5，最多 4 键）

步骤 1：加锁根 [50]
Root [50] - LOCKED

步骤 2：根节点安全？
- 只有 1 个键 → 安全
- 释放根锁？NO！因为子节点可能不安全

步骤 3：继续向下，保持祖先锁
Root [50] - LOCKED
[70|80] - LOCKED

步骤 4：[70|80] 有 2 个键，可能分裂时需要插入
- 不安全！保持锁

步骤 5：加锁 [75|76|78]
Root [50] - LOCKED
[70|80] - LOCKED
[75|76|78] - LOCKED

步骤 6：叶节点不安全（满），分裂
[75|76|77|78] → [75|76] + [77|78]
提升 77 到父节点 [70|77|80]

步骤 7：父节点仍安全（3 个键 < 4）
- 释放所有锁

并发性：
- 路径上的所有节点都被锁定
- 分裂完成后才释放
- 其他线程等待
```

**优化：乐观 Latch Crabbing**

```
假设大部分操作不会导致分裂/合并

乐观策略：
1. 向下使用读锁
2. 到达叶节点，检查是否安全
3. 如果安全：升级为写锁，完成操作
4. 如果不安全：释放所有锁，重新以悲观模式（写锁）执行

适用场景：
- 节点通常不满（插入）
- 节点通常有富余键（删除）
- 分裂/合并概率低
```

---

## B-Tree 优化技术

### 1. 写时复制（Copy-on-Write）

**传统 B-Tree 问题**：

```
需要 WAL 确保崩溃一致性
每次修改：
1. 写 WAL（1 次磁盘 I/O）
2. 写数据页（1 次磁盘 I/O）
总计：2 次磁盘 I/O
```

**COW B-Tree 方法**：

```
修改节点时：
1. 复制节点到新位置
2. 在新节点修改
3. 原子更新父节点指针（指向新节点）
4. 旧节点延迟删除（垃圾回收）

优势：
✅ 无需 WAL（旧版本仍保留，崩溃时回滚到旧版本）
✅ 天然支持 MVCC（多版本并发控制）
✅ 天然支持快照（保留旧版本根节点）

劣势：
❌ 写放大（每次修改都复制整个路径）
❌ 需要垃圾回收
```

**示例：LMDB, BoltDB**

```
初始树（根在 Page 100）：
Page 100: [50]
         /    \
Page 10: [20]  Page 20: [70]

插入 key=25：
1. 复制 Page 10 → Page 11
   Page 11: [20|25]

2. 复制 Page 100 → Page 101
   Page 101: [50]
            /    \
   Page 11: [20|25]  Page 20: [70]

3. 更新根指针：root = Page 101

旧版本：
root = Page 100（仍保留，可用于快照）

垃圾回收：
后台标记 Page 10 和 Page 100 可删除
```

**写放大分析**：

```
修改 1 个叶节点：
需要复制：
- 叶节点（4KB）
- 所有祖先节点（树高度 × 4KB）

树高度 = 4，写放大 = 4 × 4KB = 16KB
用户修改 100 字节 → 实际写入 16KB
写放大 = 160x

但优势：
- 所有写入都是顺序的（分配新页）
- 无需 WAL
- 支持 MVCC
```

---

### 2. 前缀压缩（Prefix Compression）

**问题**：B-Tree 节点内键浪费空间

```
叶节点存储：
[
  "user:123:profile",
  "user:123:settings",
  "user:123:posts",
  "user:456:profile"
]

公共前缀 "user:" 重复存储！
```

**压缩方法**：

```
方法 1：前缀消除
存储：
prefix = "user:"
keys = [
  "123:profile",
  "123:settings",
  "123:posts",
  "456:profile"
]

方法 2：增量编码
存储：
"user:123:profile"     (完整)
"-profile+settings"    (删除 profile，添加 settings)
"-settings+posts"      (删除 settings，添加 posts)
"-123:posts+456:profile"

方法 3：字典编码（内部节点）
内部节点只需存储分隔键的最小区分前缀

原始：
[10, 20, 30, 40, 50]

压缩（假设值的前缀）：
["handlebar", "handlebag", "handsome", ...]

存储：
["bar", "bag", "so", ...]  (仅存区分所需的后缀)
```

**InnoDB 前缀压缩**：

```
页内记录格式：
┌──────────────────────────────────┐
│ Record 1: "application"          │  20 bytes
├──────────────────────────────────┤
│ Record 2: (8, "tion")            │  8 bytes 公共前缀 + 4 bytes
├──────────────────────────────────┤
│ Record 3: (11, "s")              │  11 bytes 公共前缀 + 1 byte
└──────────────────────────────────┘

"application" → 完整存储
"applic" + "tion" → (8, "tion")
"application" + "s" → (11, "s")

压缩效果：
原始：20 + 11 + 12 = 43 bytes
压缩：20 + 12 + 12 = 44 bytes（这个例子压缩效果不明显）

但对于长前缀：
"user:123456789:profile" × 1000
压缩可节省 50% 空间
```

---

### 3. 批量加载（Bulk Loading）

**问题**：逐个插入构建 B-Tree 效率低

```
插入 100 万条有序记录：

传统方法：
for each record:
    insert(record)  # 每次可能触发分裂

问题：
- 频繁分裂（节点半满）
- 大量随机 I/O
- 最终节点只有 ~50% 填充率
```

**批量加载优化**：

```
假设数据已排序：
1. 直接构建满的叶节点
2. 自底向上构建索引层

伪代码：
def bulk_load(sorted_records):
    # Level 0: 构建叶节点
    leaf_nodes = []
    current_leaf = new_leaf_node()

    for record in sorted_records:
        if current_leaf.is_full():
            leaf_nodes.append(current_leaf)
            current_leaf = new_leaf_node()
        current_leaf.add(record)

    leaf_nodes.append(current_leaf)

    # Level 1+: 自底向上构建索引
    level = leaf_nodes
    while len(level) > 1:
        next_level = []
        current_internal = new_internal_node()

        for child in level:
            if current_internal.is_full():
                next_level.append(current_internal)
                current_internal = new_internal_node()

            current_internal.add_child(child, child.min_key)

        next_level.append(current_internal)
        level = next_level

    return level[0]  # Root

优势：
✅ 顺序写入（所有页按顺序分配）
✅ 节点填充率 ~100%（更少的页）
✅ 无分裂开销
✅ 性能提升 10-100 倍

适用场景：
- 初始化数据库
- 重建索引
- ETL 批量导入
```

**MySQL InnoDB 批量加载**：

```sql
-- 禁用键检查，加速批量插入
SET unique_checks=0;
SET foreign_key_checks=0;

-- 批量插入（已排序）
LOAD DATA INFILE 'data.csv' INTO TABLE my_table;

-- 重新启用检查
SET unique_checks=1;
SET foreign_key_checks=1;

内部实现：
1. 关闭唯一性检查
2. 直接构建 B-Tree（bulk load）
3. 最后验证约束
```

---

### 4. 缓冲池（Buffer Pool）

**问题**：频繁磁盘 I/O 慢

**解决：内存缓存热点页**

```
Buffer Pool（LRU 缓存）：
┌────────────────────────────────────┐
│  Page 10: [热点数据]                │
│  Page 25: [索引根节点]              │
│  Page 100: [常访问叶节点]           │
│  ...                               │
│  Total: 1GB (configurable)         │
└────────────────────────────────────┘

读取流程（带缓存）：
1. 检查 Buffer Pool
   if hit: 返回（~0.001ms）
   if miss: 读磁盘 → 加入 Buffer Pool（~10ms）

2. LRU 淘汰
   新页加入 → 淘汰最少使用页
   脏页（modified） → 写回磁盘再淘汰
```

**InnoDB Buffer Pool**：

```
配置：
innodb_buffer_pool_size = 8G  # 通常设为内存的 50-80%

结构：
┌─────────────────────────────────────┐
│  Free List (空闲页)                 │
├─────────────────────────────────────┤
│  LRU List (访问顺序)                │
│    ├─ Young (热数据，5/8)            │
│    └─ Old (冷数据，3/8)              │
├─────────────────────────────────────┤
│  Flush List (脏页，待写回)           │
└─────────────────────────────────────┘

优化：
- 预读（Read-Ahead）：预测性读取相邻页
- 分区（Partitioning）：减少锁竞争
- 压缩（Compression）：缓存更多页
```

**命中率监控**：

```sql
SHOW STATUS LIKE 'Innodb_buffer_pool%';

关键指标：
Innodb_buffer_pool_read_requests:  100,000,000  # 逻辑读
Innodb_buffer_pool_reads:          1,000,000    # 物理读

命中率 = (逻辑读 - 物理读) / 逻辑读
      = (100M - 1M) / 100M
      = 99%  ← 良好

目标：> 95% 命中率
```

---

### 5. 兄弟节点指针

**问题**：范围查询需要回到父节点

```
传统 B-Tree 范围查询（key 20-50）:

             [30]
           /      \
       [10|20]    [40|50]

查询 20-50：
1. 定位到 [10|20]，返回 20
2. 回到父节点 [30]
3. 下到 [40|50]，返回 40, 50

需要多次上下移动！
```

**优化：叶节点链表**

```
B+Tree 叶节点链表：

       [10|20] → [30|40] → [50|60]
           ↑         ↑         ↑
         next      next      next

范围查询（key 20-50）：
1. 定位到 [10|20]，返回 20
2. 跟随 next 指针 → [30|40]，返回 30, 40
3. 跟随 next 指针 → [50|60]，返回 50，停止

无需回到父节点！
顺序扫描，性能极高
```

---

## B+Tree 变种

### B-Tree vs B+Tree

**B-Tree（原始）**：

```
内部节点和叶节点都存数据

             [30:val30]
           /            \
   [10:val10|20:val20]  [40:val40|50:val50]

优点：
- 找到键即可返回（可能在内部节点）

缺点：
- 内部节点存值，导致分支因子降低
- 范围查询需要中序遍历（回到父节点）
```

**B+Tree（常用）**：

```
只有叶节点存数据，内部节点只存键

             [30|40]         ← 只存键（索引）
           /    |    \
         /      |      \
    [10|20]  [30|35]  [40|50]  ← 叶节点存数据
      ↓        ↓        ↓
    values   values   values

叶节点链表：
[10|20] → [30|35] → [40|50]

优点：
✅ 内部节点更小，分支因子更大
   - B-Tree 内部节点：键 + 值 + 指针
   - B+Tree 内部节点：键 + 指针
   - 分支因子可提升 2-3 倍

✅ 范围查询高效
   - 叶节点链表，顺序扫描

✅ 所有查询都到叶节点
   - 查询性能稳定

缺点：
❌ 所有查询都到叶节点（即使键在上层）
   - 但现代系统缓存根节点，影响不大
```

### MySQL InnoDB 的 B+Tree

**聚簇索引（Clustered Index）**：

```
主键索引，叶节点存完整行数据

表结构：
CREATE TABLE users (
    id INT PRIMARY KEY,
    name VARCHAR(50),
    email VARCHAR(100)
);

B+Tree 结构：
             [100|200]
           /     |     \
         /       |       \
    [50|75]  [125|150]  [250|300]
      ↓         ↓          ↓
  (50,Alice,a@x)  (125,Bob,b@x)  (250,Charlie,c@x)
  (75,David,d@x)  (150,Eve,e@x)  (300,Frank,f@x)

叶节点 = 实际数据行

查询 id=125:
1. 根节点: 100 < 125 < 200 → 中间
2. 内部节点: 125 → 左
3. 叶节点: 找到完整行 (125, Bob, b@x)

优点：
- 主键查询极快（一次 I/O 拿到所有列）
- 范围查询高效（顺序扫描叶节点）
```

**二级索引（Secondary Index）**：

```
非主键索引，叶节点存主键值

索引：
CREATE INDEX idx_email ON users(email);

B+Tree 结构：
             [m@x|s@x]
           /     |     \
    [a@x|d@x]  [n@x|p@x]  [u@x|z@x]
        ↓          ↓          ↓
    a@x→50     n@x→250    u@x→75
    d@x→75     p@x→100    z@x→300

叶节点存：email → 主键 id

查询 email='n@x':
1. 在二级索引中查找 'n@x' → 得到 id=250
2. 回到主键索引（聚簇索引）查找 id=250
3. 返回完整行

两次查询：
- 1 次二级索引查询
- 1 次主键索引查询（回表）

优化：覆盖索引
CREATE INDEX idx_email_name ON users(email, name);

查询：SELECT name FROM users WHERE email='n@x';
→ 无需回表！（索引已包含 name）
```

---

### PostgreSQL 的 B-Tree

**索引结构**：

```
所有索引都是二级索引（non-clustered）

表数据：Heap File（无序）
┌────────────────────────┐
│ TID (1,0): (50,Alice)  │
│ TID (1,1): (125,Bob)   │
│ TID (2,0): (75,David)  │
│ ...                    │
└────────────────────────┘

B-Tree 索引（主键）：
             [75|125]
           /    |     \
      [50]    [100]   [150|200]
       ↓        ↓        ↓
     TID(1,0) TID(2,1) TID(3,0)

叶节点存：TID（Tuple ID，指向 Heap）

查询 id=50:
1. B-Tree 查找 50 → TID (1,0)
2. 读取 Heap 页 1，偏移 0 → 数据

所有查询都需要回表！

优点：
- 更新非索引列无需更新索引
- 支持多个索引（都指向 Heap）

缺点：
- 所有查询都需要额外 I/O（回表）
```

**HOT（Heap-Only Tuple）优化**：

```
问题：更新行导致索引更新

UPDATE users SET name='Alice2' WHERE id=50;

传统：
1. 在 Heap 中创建新版本（MVCC）
2. 更新所有索引指向新 TID

HOT 优化：
1. 如果新版本在同一页，且未更新索引列
2. 旧版本链接到新版本（链表）
3. 索引不更新！

读取：
索引 → TID (1,0) → 旧版本 → 跟随链接 → 新版本

减少索引更新开销
```

---

## LSM-Tree vs B-Tree 深度对比

### 写入路径对比

**LSM-Tree 写入**：

```
PUT key=123, value=abc

1. 写 WAL（顺序追加）
   wal.append("PUT 123 abc")

2. 写 MemTable（内存，O(log n)）
   memtable.put(123, abc)

3. 返回成功
   总延迟：~0.1ms

后台异步：
4. MemTable 满 → 刷盘为 SSTable（顺序写）
5. 后台压缩（顺序读写）

写路径：全部顺序 I/O！
```

**B-Tree 写入**：

```
PUT key=123, value=abc

1. 写 WAL（顺序追加）
   wal.append("UPDATE Page 10 key=123")

2. 查找叶节点（随机读）
   读 Page 1 (root)
   读 Page 5 (internal)
   读 Page 10 (leaf)

3. 修改叶节点（随机写）
   Page 10: [100|120] → [100|123|120]
   写 Page 10

4. 可能分裂（更多随机 I/O）

5. 返回成功
   总延迟：~10-30ms

写路径：随机 I/O（慢）
```

**写性能对比**：

```
HDD：
- LSM-Tree：顺序写 100 MB/s → ~100,000 writes/s
- B-Tree：随机写 1 MB/s → ~100 writes/s
差异：1000 倍！

SSD：
- LSM-Tree：顺序写 500 MB/s → ~500,000 writes/s
- B-Tree：随机写 100 MB/s → ~10,000 writes/s
差异：50 倍

批量写入（B-Tree 优化）：
- 缓冲 1000 个写入
- 排序后批量写入
- 性能接近 LSM-Tree
```

---

### 读取路径对比

**B-Tree 读取**：

```
GET key=123

1. 查找（缓存未命中）
   读 Page 1 (root) - 缓存
   读 Page 5 (internal) - 缓存
   读 Page 10 (leaf) - 磁盘 I/O

2. 返回 value
   总延迟：~10ms（1 次磁盘 I/O）

缓存命中：
- 根节点：100% 命中（常驻内存）
- 内部节点：95% 命中
- 叶节点：80% 命中（热点数据）
- 平均：0.2 次磁盘 I/O
- 总延迟：~2ms
```

**LSM-Tree 读取**：

```
GET key=123

1. 查 MemTable（内存）
   if found: return（~0.01ms）

2. 查 Level 0（4 个 SSTable）
   for each SSTable:
       check Bloom Filter → 可能存在
       读 Index Block（可能缓存）
       读 Data Block（磁盘 I/O）

3. 查 Level 1（10 个 SSTable，但不重叠）
   定位 1 个 SSTable
   读取（磁盘 I/O）

4. 查 Level 2, 3...

总延迟：
- 最好（MemTable）：0.01ms
- 中等（Level 0-1）：10-20ms
- 最坏（Level 3）：50ms

Bloom Filter 优化：
- 90% 的不存在键可快速判断
- 实际 I/O 次数：~3-5 次
- 总延迟：~30ms
```

**读性能对比**：

```
点查询（随机读）：
- B-Tree：~1-2 次 I/O（稳定）
- LSM-Tree：~3-10 次 I/O（不稳定）
B-Tree 胜出

范围查询：
- B-Tree：定位起点 + 顺序扫描叶节点
- LSM-Tree：需要归并多层数据

小范围（<100 行）：
- B-Tree：~1 次 I/O
- LSM-Tree：~10 次 I/O（每层）
B-Tree 胜出

大范围（>10,000 行）：
- B-Tree：顺序扫描（~100 MB/s）
- LSM-Tree：顺序扫描 + 归并（~50 MB/s）
B-Tree 仍胜出
```

---

### 空间放大对比

**B-Tree 空间放大**：

```
页内碎片：
- 分裂后节点 ~50% 填充
- 删除后产生空洞
- 平均填充率：~69%（理论）

空间放大 = 1 / 0.69 ≈ 1.45 倍

优化：
- 定期 VACUUM（PostgreSQL）
- OPTIMIZE TABLE（MySQL）
- 提高填充率到 ~90%
```

**LSM-Tree 空间放大**：

```
Size-Tiered Compaction：
- 合并前：多层数据（旧版本）
- 空间放大：2-5 倍
- 取决于压缩频率

Leveled Compaction：
- 层间快速合并
- 空间放大：1.1-1.3 倍
- 接近 B-Tree

压缩算法：
- Snappy 压缩：2-3 倍
- B-Tree 通常不压缩（随机访问需要）

总体：
- LSM-Tree (Leveled + Snappy)：0.4-0.5 倍原始大小
- B-Tree（无压缩）：1.45 倍
```

---

### 写放大对比

**B-Tree 写放大**：

```
插入 1 条记录（100 字节）：

1. WAL：100 字节
2. 更新叶节点页：4KB（整页写入）
3. 可能分裂：再写 4KB（新页）
4. 可能更新父节点：4KB

总写入：100 + 4KB + 4KB + 4KB ≈ 12KB
写放大：120 倍

优化（Group Commit）：
- 批量提交 1000 条记录
- 共享 WAL 写入
- 写放大：~10 倍
```

**LSM-Tree 写放大**：

```
插入 1 条记录（100 字节）：

1. WAL：100 字节
2. MemTable：100 字节（内存）

刷盘（4MB MemTable）：
3. 写 SSTable：4MB

压缩（Leveled，层比 10:1）：
4. L0 → L1：写 40MB
5. L1 → L2：写 400MB
6. L2 → L3：写 4000MB

总写入：4 + 40 + 400 + 4000 ≈ 4444 MB
写放大：4444 MB / 100 字节 ≈ 44,440 倍！

但：
- 所有写入都是顺序的
- 分摊到大量记录
- 实际：10-50 倍
```

**对比**：

```
写放大：
- B-Tree：10-100 倍（随机 I/O）
- LSM-Tree：10-50 倍（顺序 I/O）

关键差异：
- B-Tree：小写放大，但随机写慢
- LSM-Tree：大写放大，但顺序写快

HDD 上：
- LSM-Tree 仍快（顺序写）

SSD 上：
- 写放大影响寿命
- B-Tree 更友好
```

---

### 决策树

```
选择存储引擎：

工作负载分析：
├─ 写多读少（> 80% 写）
│  ├─ 时序数据（只追加）
│  │  └─ LSM-Tree (Size-Tiered)
│  ├─ 日志系统
│  │  └─ LSM-Tree (Size-Tiered)
│  └─ 高写入吞吐需求
│     └─ LSM-Tree (Leveled)
│
├─ 读多写少（> 80% 读）
│  ├─ 事务型应用（OLTP）
│  │  └─ B-Tree
│  ├─ 低延迟需求（< 10ms）
│  │  └─ B-Tree
│  └─ 范围查询频繁
│     └─ B+Tree
│
└─ 混合负载（读写均衡）
   ├─ 点查询为主
   │  └─ B-Tree（配合缓存）
   ├─ 范围查询为主
   │  └─ B+Tree
   └─ 需要 MVCC/快照
      └─ LSM-Tree 或 COW B-Tree
```

---

## 真实系统案例

### MySQL InnoDB

**存储引擎**：B+Tree

**架构**：

```
Buffer Pool (缓存)
    ↓
Adaptive Hash Index (自动哈希索引)
    ↓
B+Tree Index
    ├─ Clustered Index (主键)
    └─ Secondary Index (非主键)
    ↓
Redo Log (WAL)
Undo Log (MVCC)
Doublewrite Buffer (防止页损坏)
```

**关键特性**：

```
1. 聚簇索引
   - 主键即数据
   - 叶节点存完整行

2. 自适应哈希索引
   - 热点数据自动建哈希索引
   - 加速点查询

3. Change Buffer
   - 缓冲非唯一二级索引的修改
   - 减少随机 I/O

4. Doublewrite Buffer
   - 写入前先写到 doublewrite
   - 防止页面部分写入（torn page）

5. MVCC
   - Undo Log 保存历史版本
   - 支持 Read Committed, Repeatable Read
```

**配置优化**：

```sql
-- Buffer Pool（内存的 70-80%）
innodb_buffer_pool_size = 8G

-- Redo Log（写性能 vs 恢复时间）
innodb_log_file_size = 1G
innodb_log_buffer_size = 16M

-- I/O 优化
innodb_flush_log_at_trx_commit = 2  # 0=快速,2=折中,1=安全
innodb_flush_method = O_DIRECT      # 绕过 OS 缓存

-- 并发
innodb_thread_concurrency = 0       # 0=无限制
```

---

### PostgreSQL

**存储引擎**：B-Tree（索引）+ Heap File（数据）

**架构**：

```
Shared Buffers (缓存)
    ↓
B-Tree Index (所有索引都是二级索引)
    ↓
Heap File (数据表，无序)
    ↓
WAL (Write-Ahead Log)
MVCC (多版本存储在 Heap)
```

**关键特性**：

```
1. 非聚簇索引
   - 所有索引都指向 Heap（TID）
   - 更新灵活（索引不需要更新）

2. MVCC (Multi-Version Concurrency Control)
   - 每行有多个版本（xmin, xmax）
   - 无锁读

3. TOAST (The Oversized-Attribute Storage Technique)
   - 大字段单独存储
   - 减少表扫描开销

4. HOT (Heap-Only Tuple)
   - 更新不触发索引更新
   - 提升更新性能

5. VACUUM
   - 清理旧版本
   - 回收空间
```

**索引类型**：

```sql
-- B-Tree（默认，适合大部分场景）
CREATE INDEX idx_id ON users(id);

-- Hash（等值查询，不支持范围）
CREATE INDEX idx_hash ON users USING hash(email);

-- GiST（几何、全文搜索）
CREATE INDEX idx_location ON places USING gist(location);

-- GIN（数组、JSON、全文）
CREATE INDEX idx_tags ON posts USING gin(tags);

-- BRIN（块范围索引，时序数据）
CREATE INDEX idx_timestamp ON logs USING brin(timestamp);
```

---

### SQLite

**存储引擎**：B-Tree（单文件）

**架构**：

```
Single File Database
    ↓
B-Tree (通用，支持表和索引)
    ├─ Table B-Tree (rowid 为键)
    └─ Index B-Tree (索引列为键)
    ↓
Pager (页管理)
    ↓
Journal (回滚日志)
WAL (可选，Write-Ahead Log)
```

**关键特性**：

```
1. 嵌入式
   - 无服务器进程
   - 直接读写文件
   - 适合移动应用、浏览器

2. ACID
   - Journal Mode（默认，回滚日志）
   - WAL Mode（更好的并发）

3. 轻量级
   - 库大小 ~500KB
   - 内存占用小

4. 限制
   - 单写者（同一时间只有 1 个写事务）
   - 适合读多写少场景
```

**性能优化**：

```sql
-- 启用 WAL（提高并发）
PRAGMA journal_mode=WAL;

-- 增加缓存
PRAGMA cache_size=10000;  -- 页数

-- 同步模式（性能 vs 安全）
PRAGMA synchronous=NORMAL;  -- FULL=安全,NORMAL=折中,OFF=快速

-- 自动清理
PRAGMA auto_vacuum=INCREMENTAL;
```

---

### LMDB (Lightning Memory-Mapped Database)

**存储引擎**：COW B+Tree（Copy-on-Write）

**架构**：

```
Memory-Mapped File (mmap)
    ↓
COW B+Tree
    ├─ 修改时复制
    └─ 无需 WAL
    ↓
MVCC (多版本)
    ├─ 读不阻塞写
    └─ 写不阻塞读
```

**关键特性**：

```
1. 零拷贝读取
   - 直接 mmap 文件
   - 无需序列化/反序列化

2. 无 WAL
   - COW 保证原子性
   - 崩溃安全

3. 极快的读性能
   - 内存访问速度
   - 无系统调用开销

4. MVCC
   - 快照隔离
   - 读不阻塞写

5. 限制
   - 单写者
   - 数据库大小需预分配
   - mmap 在 32 位系统上受限
```

**性能**：

```
LMDB vs LevelDB（随机读）:
LMDB: ~400,000 reads/s
LevelDB: ~100,000 reads/s

LMDB 快 4 倍！（零拷贝优势）

写入：
LMDB: ~40,000 writes/s（单写者限制）
LevelDB: ~100,000 writes/s（顺序写）
```

---

## 深度思考问题

### 1. 为什么 B-Tree 分支因子要这么大？

**回答**：

```
关键：磁盘 I/O 是瓶颈！

二叉树：
分支因子 = 2
深度 = log₂(n)
n=100 万，深度 = 20
20 次磁盘 I/O × 10ms = 200ms

B-Tree（m=100）：
分支因子 = 100
深度 = log₁₀₀(n)
n=100 万，深度 = 3
3 次磁盘 I/O × 10ms = 30ms

优化 6 倍以上！

关键权衡：
- 节点内二分查找：log₂(100) ≈ 7 次比较（内存，极快）
- 减少 1 次磁盘 I/O：节省 10ms
→ 值得！

最优分支因子：
使节点大小 ≈ 磁盘页大小（4KB-16KB）
最大化利用每次 I/O
```

---

### 2. 为什么 MySQL 使用 B+Tree 而不是 B-Tree？

**回答**：

```
B+Tree 优势：

1. 内部节点更小
   B-Tree 内部节点：键 + 值 + 指针
   B+Tree 内部节点：键 + 指针

   同样 4KB 页：
   B-Tree：50 个键
   B+Tree：100 个键

   深度减少 → 减少 I/O

2. 范围查询高效
   B+Tree 叶节点链表 → 顺序扫描
   B-Tree 需要中序遍历 → 回到父节点

   范围查询性能差异：10 倍以上

3. 查询性能稳定
   B+Tree：所有查询都到叶节点（深度一致）
   B-Tree：可能在任何层找到（深度不一致）

MySQL 的聚簇索引：
- 叶节点存完整行 → 必须用 B+Tree
- 内部节点只存主键 → 分支因子大
```

---

### 3. COW B-Tree 为什么不流行？

**回答**：

```
COW B-Tree 优势：
✅ 无需 WAL
✅ 天然 MVCC
✅ 崩溃安全

劣势：
❌ 写放大极大
   修改 1 个叶节点 → 复制整个路径
   树高 4 → 写放大 16KB（4 个页）

❌ 需要垃圾回收
   旧版本页需要异步清理
   增加复杂性

❌ 写性能不如 LSM-Tree
   顺序写但写放大大

适用场景：
- 读多写少
- 需要快照功能
- 嵌入式系统（LMDB）

不适用：
- 写密集型（不如 LSM-Tree）
- 事务型数据库（WAL + 传统 B-Tree 更成熟）
```

---

### 4. 为什么不用红黑树/AVL 树做数据库索引？

**回答**：

```
红黑树/AVL 树设计目标：内存数据结构
B-Tree 设计目标：磁盘数据结构

关键差异：

1. 深度
   红黑树：log₂(n) ≈ 20（100 万条）
   B-Tree：log₁₀₀(n) ≈ 3

   磁盘 I/O 次数：20 vs 3（差 6 倍）

2. 缓存局部性
   红黑树：节点分散，缓存不友好
   B-Tree：节点连续（4KB 页），缓存友好

3. 旋转操作
   红黑树：插入/删除需要旋转（修改多个节点）
   B-Tree：分裂/合并（局部操作）

4. 磁盘利用率
   红黑树节点：~20 字节，读 4KB 浪费
   B-Tree 节点：4KB，充分利用

结论：
内存索引：红黑树/跳表（Go map, Java TreeMap）
磁盘索引：B-Tree（MySQL, PostgreSQL）
```

---

## 学习建议

### 理论基础

1. **理解 I/O 模型**：
   - 磁盘 vs SSD 性能差异
   - 随机 I/O vs 顺序 I/O
   - 页缓存机制

2. **掌握核心概念**：
   - 分支因子与深度的关系
   - 分裂/合并算法
   - WAL 崩溃恢复

3. **对比思维**：
   - B-Tree vs LSM-Tree
   - B-Tree vs B+Tree
   - 不同数据库的设计权衡

### 实践项目

1. **实现简单 B+Tree**（推荐）：
   - 内存版本（无持久化）
   - 支持插入、查找、删除
   - 支持分裂和合并
   - 范围查询

2. **阅读源码**：
   - SQLite B-Tree 实现（C，简洁）
   - PostgreSQL B-Tree（C，完整）
   - LMDB（COW B-Tree，C）

3. **性能测试**：
   - 对比不同分支因子的性能
   - 测试分裂/合并开销
   - 分析 Buffer Pool 命中率

### 延伸阅读

**论文**：

- [The Ubiquitous B-Tree](https://dl.acm.org/doi/10.1145/356770.356776) - Comer, 1979
- [Organization and Maintenance of Large Ordered Indices](https://infolab.usc.edu/csci585/Spring2010/den_ar/indexing.pdf) - Bayer & McCreight, 1970（B-Tree 原始论文）
- [ARIES: A Transaction Recovery Method](https://cs.stanford.edu/people/chrismre/cs345/rl/aries.pdf) - IBM, 1992

**书籍**：

- Database Internals - Alex Petrov（第 4-5 章）
- Database System Concepts - Silberschatz（第 11 章）

**博客**：

- [SQLite B-Tree Module](https://www.sqlite.org/btreemodule.html)
- [PostgreSQL B-Tree Implementation](https://www.postgresql.org/docs/current/btree-implementation.html)
- [InnoDB Index Structure](https://dev.mysql.com/doc/refman/8.0/en/innodb-physical-structure.html)

---

## 总结

### 核心要点

1. **B-Tree 设计哲学**
   - 为磁盘优化（分支因子大）
   - 与磁盘页对齐（节点 = 4KB）
   - 保持平衡（所有叶子深度相同）

2. **关键操作**
   - 查找：O(log_m n) 磁盘 I/O
   - 插入：可能分裂（写放大）
   - 删除：可能合并（复杂）

3. **可靠性保障**
   - WAL（崩溃恢复）
   - Latch（并发控制）
   - ARIES（恢复算法）

4. **优化技术**
   - COW（无需 WAL）
   - 前缀压缩（节省空间）
   - 批量加载（快速构建）
   - Buffer Pool（减少 I/O）

5. **B+Tree 变种**
   - 只有叶节点存数据
   - 分支因子更大
   - 范围查询高效
   - MySQL InnoDB 聚簇索引

6. **与 LSM-Tree 对比**
   - B-Tree：读优化，写慢（随机 I/O）
   - LSM-Tree：写优化，读慢（多层查找）
   - 选择取决于工作负载

### 最重要的洞察

**B-Tree 的成功秘诀**：

> 通过增大分支因子，将深度从 log₂(n) 降低到 log₁₀₀(n)，使得即使亿级数据也只需 4-5 次磁盘 I/O。这是针对磁盘特性的完美优化。

**设计权衡**：

> 没有完美的数据结构。B-Tree 优化读取，代价是写入慢；LSM-Tree 优化写入，代价是读取慢。理解工作负载，选择合适的工具。

祝学习愉快！🚀

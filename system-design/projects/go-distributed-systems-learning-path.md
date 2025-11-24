# Go语言分布式系统学习路径

> **目标**: 从基础数据结构到实际分布式系统的渐进式学习路径
>
> **适用人群**: 已阅读完DDIA,希望通过Go语言项目实战巩固分布式系统知识
>
> **预计时长**: 6-8个月 (兼职学习)

---

## 📊 阶段1: 基础数据结构 (对应DDIA第3章 - 存储与检索)

### 项目1: 实现LSM-Tree存储引擎

**推荐项目**: [goLevelDB](https://github.com/syndtr/goleveldb) 源码学习 + 自己实现简化版

**核心实现目标**:
- MemTable (跳表或红黑树)
- WAL (Write-Ahead Log)
- SSTable写入和读取
- Compaction策略

**对应DDIA概念**:
- 第3章: LSM-Tree原理
- 第7章: WAL在事务中的作用

**实现步骤**:
```
1. 实现简单的MemTable (使用skiplist)
2. 实现WAL持久化
3. 实现SSTable格式化存储
4. 实现简单的Major Compaction
```

**参考资源**:
- [mini-lsm](https://github.com/skyzh/mini-lsm) (Rust实现,但思路清晰)
- [rosedb](https://github.com/roseduan/rosedb) (Go语言的BitCask实现)

**关键代码框架**:
```go
type LSMTree struct {
    memTable    *MemTable
    immutables  []*MemTable
    sstables    []*SSTable
    wal         *WAL
    mu          sync.RWMutex
}

func (lsm *LSMTree) Put(key, value []byte) error {
    // 1. 写WAL
    lsm.wal.Append(key, value)

    // 2. 写MemTable
    lsm.memTable.Put(key, value)

    // 3. 检查是否需要flush
    if lsm.memTable.Size() > threshold {
        lsm.flushMemTable()
    }

    return nil
}

func (lsm *LSMTree) Get(key []byte) ([]byte, error) {
    // 1. 查询MemTable
    if val, ok := lsm.memTable.Get(key); ok {
        return val, nil
    }

    // 2. 查询Immutable MemTables
    for _, imm := range lsm.immutables {
        if val, ok := imm.Get(key); ok {
            return val, nil
        }
    }

    // 3. 查询SSTables (从新到旧)
    for i := len(lsm.sstables) - 1; i >= 0; i-- {
        if val, ok := lsm.sstables[i].Get(key); ok {
            return val, nil
        }
    }

    return nil, ErrNotFound
}
```

**预计时间**: 2周

---

### 项目2: 实现B+树索引

**推荐**: 从零实现B+树存储引擎

**核心实现**:
- B+树节点分裂和合并
- 页面管理
- 缓冲池 (Buffer Pool)
- 简单的事务支持

**对应DDIA概念**:
- 第3章: B-Tree索引结构
- 第7章: 写时复制 (Copy-on-Write)

**GitHub参考**:
- [go-btree](https://github.com/colincrawford/go-btree)
- [google/btree](https://github.com/google/btree)

**关键数据结构**:
```go
type BPlusTree struct {
    root  *Node
    order int  // 每个节点的最大子节点数
}

type Node struct {
    isLeaf   bool
    keys     [][]byte
    children []*Node    // 内部节点使用
    values   [][]byte   // 叶子节点使用
    next     *Node      // 叶子节点链表
}

func (tree *BPlusTree) Insert(key, value []byte) {
    // 找到插入位置
    leaf := tree.findLeaf(key)

    // 插入到叶子节点
    leaf.insert(key, value)

    // 检查是否需要分裂
    if leaf.needSplit() {
        tree.split(leaf)
    }
}

func (tree *BPlusTree) RangeScan(startKey, endKey []byte) [][]byte {
    // B+树的优势: 叶子节点链表支持高效范围扫描
    leaf := tree.findLeaf(startKey)
    results := [][]byte{}

    for leaf != nil {
        for i, key := range leaf.keys {
            if bytes.Compare(key, endKey) > 0 {
                return results
            }
            if bytes.Compare(key, startKey) >= 0 {
                results = append(results, leaf.values[i])
            }
        }
        leaf = leaf.next
    }

    return results
}
```

**预计时间**: 2周

---

## 🔄 阶段2: 单机事务 (对应DDIA第7章)

### 项目3: 实现MVCC存储引擎

**推荐**: 基于前面的LSM或B+树,添加MVCC

**核心实现**:
- 版本链管理
- 快照隔离 (Snapshot Isolation)
- GC机制清理旧版本

**实现要点**:
```go
type MVCCKey struct {
    Key       []byte
    Timestamp uint64  // 作为版本号,递减存储(新版本在前)
}

type MVCCValue struct {
    Value   []byte
    TxnID   uint64
    Deleted bool
}

type MVCCStore struct {
    store     *LSMTree  // 底层存储
    txnIDGen  *atomic.Uint64
}

// 写入新版本
func (s *MVCCStore) Put(key, value []byte, txnID uint64) error {
    mvccKey := MVCCKey{
        Key:       key,
        Timestamp: s.nextTimestamp(),
    }

    mvccVal := MVCCValue{
        Value:   value,
        TxnID:   txnID,
        Deleted: false,
    }

    return s.store.Put(encode(mvccKey), encode(mvccVal))
}

// 快照读: 读取小于等于snapshotTS的最新版本
func (s *MVCCStore) Get(key []byte, snapshotTS uint64) ([]byte, error) {
    // 扫描所有版本,找到第一个 <= snapshotTS 的版本
    iter := s.store.Scan(MVCCKey{Key: key, Timestamp: snapshotTS})

    for iter.Valid() {
        mvccKey := decode(iter.Key())
        mvccVal := decode(iter.Value())

        // 检查是否是同一个key
        if !bytes.Equal(mvccKey.Key, key) {
            break
        }

        // 检查版本是否可见
        if mvccKey.Timestamp <= snapshotTS {
            if mvccVal.Deleted {
                return nil, ErrNotFound
            }
            return mvccVal.Value, nil
        }

        iter.Next()
    }

    return nil, ErrNotFound
}

// 垃圾回收: 清理旧版本
func (s *MVCCStore) GC(gcTimestamp uint64) {
    // 对每个key,保留最新版本,删除所有早于gcTimestamp的旧版本
    // 实现略
}
```

**对应DDIA概念**:
- 第7章: 快照隔离
- 第7章: MVCC实现原理
- 第7章: 防止Lost Updates

**参考**:
- [CockroachDB的MVCC实现](https://github.com/cockroachdb/cockroach/tree/master/pkg/storage)
- [TiKV的MVCC](https://tikv.org/deep-dive/key-value-engine/mvcc/)

**预计时间**: 2周

---

### 项目4: 实现两阶段锁 (2PL)

**目标**: 在单机存储引擎上实现Serializable隔离级别

**核心实现**:
- 行级锁管理器
- 死锁检测 (Waits-For Graph)
- 可串行化快照隔离 (SSI)

**实现框架**:
```go
type LockManager struct {
    locks     map[string]*Lock  // key -> lock
    waitGraph *WaitGraph       // 死锁检测
    mu        sync.Mutex
}

type Lock struct {
    holders   map[uint64]LockMode  // txnID -> mode
    waiters   []LockRequest
}

type LockMode int
const (
    SharedLock    LockMode = iota
    ExclusiveLock
)

type LockRequest struct {
    TxnID uint64
    Mode  LockMode
    Done  chan bool
}

func (lm *LockManager) AcquireLock(txnID uint64, key string, mode LockMode) error {
    lm.mu.Lock()
    defer lm.mu.Unlock()

    lock := lm.getOrCreateLock(key)

    // 检查是否可以立即获取
    if lock.canAcquire(txnID, mode) {
        lock.grant(txnID, mode)
        return nil
    }

    // 需要等待,检查死锁
    if lm.wouldDeadlock(txnID, lock.holders) {
        return ErrDeadlock
    }

    // 加入等待队列
    req := LockRequest{
        TxnID: txnID,
        Mode:  mode,
        Done:  make(chan bool),
    }
    lock.waiters = append(lock.waiters, req)
    lm.waitGraph.AddEdge(txnID, lock.holders)

    lm.mu.Unlock()
    <-req.Done  // 等待锁
    lm.mu.Lock()

    return nil
}

func (lm *LockManager) ReleaseLocks(txnID uint64) {
    lm.mu.Lock()
    defer lm.mu.Unlock()

    // 释放所有锁,唤醒等待者
    for key, lock := range lm.locks {
        if _, held := lock.holders[txnID]; held {
            delete(lock.holders, txnID)
            lm.wakeupWaiters(key)
        }
    }

    lm.waitGraph.RemoveNode(txnID)
}

// 死锁检测: DFS检测环
func (wg *WaitGraph) HasCycle(startTxn uint64) bool {
    visited := make(map[uint64]bool)
    recStack := make(map[uint64]bool)

    return wg.dfs(startTxn, visited, recStack)
}
```

**对应DDIA概念**:
- 第7章: 两阶段锁
- 第7章: 死锁检测
- 第7章: Serializable Snapshot Isolation (SSI)

**预计时间**: 2周

---

## 🌐 阶段3: 分布式基础 (对应DDIA第5-6章)

### 项目5: 实现数据分区和复制

**推荐项目**: [Consistent Hash Ring](https://github.com/stathat/consistent)

**核心实现**:
- 一致性哈希环
- 主从复制 (Leader-Follower)
- 异步复制和同步复制

**一致性哈希实现**:
```go
type ConsistentHash struct {
    ring         map[uint32]string  // hash -> node
    sortedHashes []uint32
    vnodes       int  // 虚拟节点数
}

func (ch *ConsistentHash) AddNode(node string) {
    for i := 0; i < ch.vnodes; i++ {
        hash := ch.hash(fmt.Sprintf("%s:%d", node, i))
        ch.ring[hash] = node
        ch.sortedHashes = append(ch.sortedHashes, hash)
    }
    sort.Slice(ch.sortedHashes, func(i, j int) bool {
        return ch.sortedHashes[i] < ch.sortedHashes[j]
    })
}

func (ch *ConsistentHash) GetNode(key string) string {
    hash := ch.hash(key)

    // 二分查找第一个 >= hash 的节点
    idx := sort.Search(len(ch.sortedHashes), func(i int) bool {
        return ch.sortedHashes[i] >= hash
    })

    if idx == len(ch.sortedHashes) {
        idx = 0  // 环形,回到第一个
    }

    return ch.ring[ch.sortedHashes[idx]]
}
```

**主从复制实现**:
```go
type Replicator struct {
    leader    *Node
    followers []*Node
    replQueue chan *WriteOp
    mode      ReplicationMode
}

type ReplicationMode int
const (
    AsyncReplication  ReplicationMode = iota  // 异步复制
    SyncReplication                           // 同步复制
    SemiSyncReplication                       // 半同步复制
)

func (r *Replicator) Replicate(op *WriteOp) error {
    switch r.mode {
    case AsyncReplication:
        return r.replicateAsync(op)
    case SyncReplication:
        return r.replicateSync(op)
    case SemiSyncReplication:
        return r.replicateSemiSync(op)
    }
    return nil
}

func (r *Replicator) replicateAsync(op *WriteOp) error {
    // 发送到所有follower,不等待确认
    for _, f := range r.followers {
        go func(follower *Node) {
            follower.Apply(op)
        }(f)
    }
    return nil
}

func (r *Replicator) replicateSync(op *WriteOp) error {
    // 等待所有follower确认
    acks := make(chan error, len(r.followers))

    for _, f := range r.followers {
        go func(follower *Node) {
            acks <- follower.Apply(op)
        }(f)
    }

    // 等待所有确认
    for i := 0; i < len(r.followers); i++ {
        if err := <-acks; err != nil {
            return err
        }
    }

    return nil
}

func (r *Replicator) replicateSemiSync(op *WriteOp) error {
    // 等待至少一个follower确认即可
    acks := make(chan error, len(r.followers))

    for _, f := range r.followers {
        go func(follower *Node) {
            acks <- follower.Apply(op)
        }(f)
    }

    // 至少等待一个成功
    return <-acks
}
```

**对应DDIA概念**:
- 第5章: 主从复制
- 第5章: 同步复制 vs 异步复制
- 第6章: 分区策略
- 第6章: 一致性哈希

**预计时间**: 2周

---

### 项目6: 实现Quorum读写

**目标**: 基于项目5,实现N/W/R可配置的仲裁机制

**核心实现**:
- 配置N个副本
- 写入需要W个确认
- 读取需要R个响应
- 版本向量检测冲突

**实现框架**:
```go
type QuorumStore struct {
    nodes      []*Node
    N          int  // 副本数
    W          int  // 写入quorum
    R          int  // 读取quorum
    hashRing   *ConsistentHash
}

func NewQuorumStore(nodes []*Node, N, W, R int) *QuorumStore {
    // 验证: W + R > N (保证读能看到最新写)
    if W+R <= N {
        panic("W + R must be > N for consistency")
    }

    return &QuorumStore{
        nodes:    nodes,
        N:        N,
        W:        W,
        R:        R,
        hashRing: NewConsistentHash(),
    }
}

func (qs *QuorumStore) Put(key, value []byte) error {
    // 1. 确定副本节点
    replicas := qs.getReplicaNodes(key, qs.N)

    // 2. 生成版本向量
    version := qs.generateVersion()

    // 3. 并发写入,等待W个确认
    acks := make(chan error, len(replicas))

    for _, node := range replicas {
        go func(n *Node) {
            acks <- n.Put(key, value, version)
        }(node)
    }

    // 等待W个成功
    successCount := 0
    for i := 0; i < len(replicas); i++ {
        if err := <-acks; err == nil {
            successCount++
            if successCount >= qs.W {
                return nil  // 达到quorum
            }
        }
    }

    return ErrQuorumNotMet
}

func (qs *QuorumStore) Get(key []byte) ([]byte, error) {
    // 1. 确定副本节点
    replicas := qs.getReplicaNodes(key, qs.N)

    // 2. 并发读取,等待R个响应
    type response struct {
        value   []byte
        version *VersionVector
        err     error
    }

    responses := make(chan response, len(replicas))

    for _, node := range replicas {
        go func(n *Node) {
            val, ver, err := n.Get(key)
            responses <- response{value: val, version: ver, err: err}
        }(node)
    }

    // 收集R个响应
    collected := []response{}
    for i := 0; i < len(replicas); i++ {
        resp := <-responses
        if resp.err == nil {
            collected = append(collected, resp)
            if len(collected) >= qs.R {
                break
            }
        }
    }

    if len(collected) < qs.R {
        return nil, ErrQuorumNotMet
    }

    // 3. 版本向量比较,找最新值
    latest := collected[0]
    conflicts := []response{}

    for _, resp := range collected[1:] {
        cmp := latest.version.Compare(resp.version)
        switch cmp {
        case VectorAfter:
            // latest更新,跳过
        case VectorBefore:
            // resp更新,更新latest
            latest = resp
            conflicts = []response{}
        case VectorConcurrent:
            // 并发冲突,记录
            conflicts = append(conflicts, resp)
        }
    }

    // 4. 处理冲突
    if len(conflicts) > 0 {
        // 返回所有冲突版本,由应用层解决
        return qs.resolveConflicts(latest, conflicts)
    }

    // 5. 读修复: 将最新值写回落后的副本
    go qs.readRepair(key, latest, collected)

    return latest.value, nil
}

// 版本向量
type VersionVector map[string]uint64  // nodeID -> counter

func (v1 VersionVector) Compare(v2 VersionVector) VectorComparison {
    v1Greater := false
    v2Greater := false

    // 检查所有节点
    allNodes := make(map[string]bool)
    for node := range v1 {
        allNodes[node] = true
    }
    for node := range v2 {
        allNodes[node] = true
    }

    for node := range allNodes {
        c1 := v1[node]
        c2 := v2[node]

        if c1 > c2 {
            v1Greater = true
        } else if c1 < c2 {
            v2Greater = true
        }
    }

    if v1Greater && !v2Greater {
        return VectorAfter  // v1 > v2
    } else if !v1Greater && v2Greater {
        return VectorBefore  // v1 < v2
    } else if v1Greater && v2Greater {
        return VectorConcurrent  // 冲突
    } else {
        return VectorEqual
    }
}
```

**对应DDIA概念**:
- 第5章: 无主复制
- 第5章: Quorum读写
- 第8章: 版本向量
- 第5章: 读修复 (Read Repair)

**参考**:
- Riak的Quorum实现
- Cassandra的一致性级别

**预计时间**: 2周

---

## 🎯 阶段4: 分布式协调 (对应DDIA第8-9章)

### 项目7: 实现Raft共识算法 ⭐ 核心项目

**推荐**: [MIT 6.824](https://pdos.csail.mit.edu/6.824/) Lab 2

**核心实现**:
- Leader选举
- 日志复制
- 安全性保证 (Election Safety, Log Matching)
- 日志压缩 (Snapshot)

**实现步骤**:
```
Lab 2A: Leader Election (1-2周)
Lab 2B: Log Replication (2-3周)
Lab 2C: Persistence (1周)
Lab 2D: Log Compaction (1-2周)
```

**关键数据结构**:
```go
type Raft struct {
    mu        sync.Mutex
    peers     []*RaftClient
    persister *Persister
    me        int  // 自己的索引

    // 持久化状态
    currentTerm int
    votedFor    int
    log         []LogEntry

    // 易失状态
    commitIndex int
    lastApplied int

    // Leader易失状态
    nextIndex   []int  // 每个follower的下一条日志索引
    matchIndex  []int  // 每个follower已复制的最高日志索引

    // 角色
    state       NodeState  // Follower/Candidate/Leader

    // 选举定时器
    electionTimer  *time.Timer
    heartbeatTimer *time.Timer

    // 应用通道
    applyCh chan ApplyMsg
}

type LogEntry struct {
    Term    int
    Command interface{}
}

type NodeState int
const (
    Follower NodeState = iota
    Candidate
    Leader
)
```

**Leader选举实现**:
```go
func (rf *Raft) startElection() {
    rf.mu.Lock()
    rf.currentTerm++
    rf.state = Candidate
    rf.votedFor = rf.me
    currentTerm := rf.currentTerm
    lastLogIndex := len(rf.log) - 1
    lastLogTerm := 0
    if lastLogIndex >= 0 {
        lastLogTerm = rf.log[lastLogIndex].Term
    }
    rf.mu.Unlock()

    // 向所有peer请求投票
    votes := 1  // 投给自己
    finished := 1

    for i := range rf.peers {
        if i == rf.me {
            continue
        }

        go func(peer int) {
            args := RequestVoteArgs{
                Term:         currentTerm,
                CandidateId:  rf.me,
                LastLogIndex: lastLogIndex,
                LastLogTerm:  lastLogTerm,
            }

            reply := RequestVoteReply{}
            ok := rf.sendRequestVote(peer, &args, &reply)

            rf.mu.Lock()
            defer rf.mu.Unlock()

            if !ok {
                return
            }

            // 检查term
            if reply.Term > rf.currentTerm {
                rf.becomeFollower(reply.Term)
                return
            }

            // 检查是否仍是同一term的candidate
            if rf.state != Candidate || rf.currentTerm != currentTerm {
                return
            }

            finished++
            if reply.VoteGranted {
                votes++
            }

            // 获得多数票
            if votes > len(rf.peers)/2 {
                rf.becomeLeader()
            }
        }(i)
    }
}

func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
    rf.mu.Lock()
    defer rf.mu.Unlock()

    reply.Term = rf.currentTerm
    reply.VoteGranted = false

    // 1. Reply false if term < currentTerm
    if args.Term < rf.currentTerm {
        return nil
    }

    // 2. If RPC term > currentTerm, convert to follower
    if args.Term > rf.currentTerm {
        rf.becomeFollower(args.Term)
    }

    // 3. 检查是否已投票
    if rf.votedFor != -1 && rf.votedFor != args.CandidateId {
        return nil
    }

    // 4. 检查日志是否至少一样新
    lastLogIndex := len(rf.log) - 1
    lastLogTerm := 0
    if lastLogIndex >= 0 {
        lastLogTerm = rf.log[lastLogIndex].Term
    }

    logUpToDate := args.LastLogTerm > lastLogTerm ||
        (args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex)

    if !logUpToDate {
        return nil
    }

    // 5. 投票
    rf.votedFor = args.CandidateId
    reply.VoteGranted = true
    rf.resetElectionTimer()

    return nil
}
```

**日志复制实现**:
```go
func (rf *Raft) sendHeartbeats() {
    for i := range rf.peers {
        if i == rf.me {
            continue
        }

        go func(peer int) {
            rf.mu.Lock()

            if rf.state != Leader {
                rf.mu.Unlock()
                return
            }

            prevLogIndex := rf.nextIndex[peer] - 1
            prevLogTerm := 0
            if prevLogIndex >= 0 && prevLogIndex < len(rf.log) {
                prevLogTerm = rf.log[prevLogIndex].Term
            }

            entries := rf.log[rf.nextIndex[peer]:]

            args := AppendEntriesArgs{
                Term:         rf.currentTerm,
                LeaderId:     rf.me,
                PrevLogIndex: prevLogIndex,
                PrevLogTerm:  prevLogTerm,
                Entries:      entries,
                LeaderCommit: rf.commitIndex,
            }

            rf.mu.Unlock()

            reply := AppendEntriesReply{}
            ok := rf.sendAppendEntries(peer, &args, &reply)

            if !ok {
                return
            }

            rf.mu.Lock()
            defer rf.mu.Unlock()

            if reply.Term > rf.currentTerm {
                rf.becomeFollower(reply.Term)
                return
            }

            if rf.state != Leader || rf.currentTerm != args.Term {
                return
            }

            if reply.Success {
                // 更新nextIndex和matchIndex
                rf.nextIndex[peer] = prevLogIndex + len(entries) + 1
                rf.matchIndex[peer] = rf.nextIndex[peer] - 1

                // 尝试提交
                rf.tryCommit()
            } else {
                // 日志不匹配,回退nextIndex
                rf.nextIndex[peer]--
            }
        }(i)
    }
}

func (rf *Raft) tryCommit() {
    // 找到多数派已复制的最大索引
    for n := len(rf.log) - 1; n > rf.commitIndex; n-- {
        // 只能提交当前term的日志
        if rf.log[n].Term != rf.currentTerm {
            continue
        }

        // 计算有多少节点复制了log[n]
        count := 1  // leader自己
        for i := range rf.peers {
            if i != rf.me && rf.matchIndex[i] >= n {
                count++
            }
        }

        // 多数派确认
        if count > len(rf.peers)/2 {
            rf.commitIndex = n
            go rf.applyLogs()
            break
        }
    }
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
    rf.mu.Lock()
    defer rf.mu.Unlock()

    reply.Term = rf.currentTerm
    reply.Success = false

    // 1. term检查
    if args.Term < rf.currentTerm {
        return nil
    }

    if args.Term > rf.currentTerm {
        rf.becomeFollower(args.Term)
    }

    rf.resetElectionTimer()

    // 2. 检查prevLog是否匹配
    if args.PrevLogIndex >= 0 {
        if args.PrevLogIndex >= len(rf.log) {
            return nil  // 日志太短
        }
        if rf.log[args.PrevLogIndex].Term != args.PrevLogTerm {
            return nil  // term不匹配
        }
    }

    // 3. 追加新日志
    for i, entry := range args.Entries {
        index := args.PrevLogIndex + 1 + i

        if index < len(rf.log) {
            // 冲突检测
            if rf.log[index].Term != entry.Term {
                rf.log = rf.log[:index]  // 删除冲突及之后的日志
                rf.log = append(rf.log, entry)
            }
        } else {
            rf.log = append(rf.log, entry)
        }
    }

    // 4. 更新commitIndex
    if args.LeaderCommit > rf.commitIndex {
        rf.commitIndex = min(args.LeaderCommit, len(rf.log)-1)
        go rf.applyLogs()
    }

    reply.Success = true
    return nil
}
```

**对应DDIA概念**:
- 第9章: 共识算法原理
- 第8章: 网络分区处理
- 第8章: Fencing Token机制 (Raft的term就是fencing token)
- 第9章: Total Order Broadcast

**关键测试**:
```bash
# MIT 6.824提供的测试
go test -run 2A  # Leader Election
go test -run 2B  # Log Replication
go test -run 2C  # Persistence
go test -run 2D  # Log Compaction

# 压力测试
go test -run 2B -race -count 100
```

**参考资源**:
- [Raft论文](https://raft.github.io/raft.pdf)
- [Raft可视化](https://raft.github.io/)
- [etcd的Raft实现](https://github.com/etcd-io/etcd/tree/main/raft)
- [MIT 6.824课程](https://pdos.csail.mit.edu/6.824/)

**预计时间**: 6-8周 (这是最重要的项目,建议投入足够时间)

---

### 项目8: 基于Raft实现分布式KV存储 ⭐ 综合项目

**推荐**: MIT 6.824 Lab 3

**核心实现**:
- 在Raft之上构建KV服务
- 客户端请求去重 (幂等性)
- 线性一致性读写
- 快照和状态恢复

**架构设计**:
```go
type KVServer struct {
    mu      sync.Mutex
    me      int
    rf      *Raft
    applyCh chan ApplyMsg

    // KV存储
    db map[string]string

    // 去重: 记录每个客户端的最后一次请求
    lastApplied map[int64]*OpContext  // clientID -> last op

    // 等待通道: 等待Raft提交
    notifyCh map[int]chan OpResult  // log index -> notify chan

    maxraftstate int  // 快照阈值
}

type OpContext struct {
    SeqNum int64
    Result OpResult
}

type Op struct {
    Type     OpType  // Get/Put/Append
    Key      string
    Value    string
    ClientID int64
    SeqNum   int64
}

type OpResult struct {
    Err   Err
    Value string
}
```

**Put/Append实现**:
```go
func (kv *KVServer) Put(args *PutAppendArgs, reply *PutAppendReply) {
    // 1. 检查幂等性
    kv.mu.Lock()
    if lastOp, ok := kv.lastApplied[args.ClientID]; ok {
        if lastOp.SeqNum == args.SeqNum {
            // 重复请求,直接返回缓存结果
            reply.Err = lastOp.Result.Err
            kv.mu.Unlock()
            return
        }
    }
    kv.mu.Unlock()

    // 2. 构造Op,提交到Raft
    op := Op{
        Type:     OpPut,
        Key:      args.Key,
        Value:    args.Value,
        ClientID: args.ClientID,
        SeqNum:   args.SeqNum,
    }

    index, term, isLeader := kv.rf.Start(op)
    if !isLeader {
        reply.Err = ErrWrongLeader
        return
    }

    // 3. 等待Raft提交
    kv.mu.Lock()
    ch := make(chan OpResult, 1)
    kv.notifyCh[index] = ch
    kv.mu.Unlock()

    // 设置超时
    select {
    case result := <-ch:
        reply.Err = result.Err
    case <-time.After(1 * time.Second):
        reply.Err = ErrTimeout
    }

    // 4. 清理
    kv.mu.Lock()
    delete(kv.notifyCh, index)
    kv.mu.Unlock()
}
```

**应用Raft日志**:
```go
func (kv *KVServer) applyLoop() {
    for msg := range kv.applyCh {
        if msg.CommandValid {
            kv.applyCommand(msg)
        } else if msg.SnapshotValid {
            kv.applySnapshot(msg.Snapshot)
        }
    }
}

func (kv *KVServer) applyCommand(msg ApplyMsg) {
    kv.mu.Lock()
    defer kv.mu.Unlock()

    op := msg.Command.(Op)

    // 检查是否已应用过 (幂等性)
    var result OpResult
    if lastOp, ok := kv.lastApplied[op.ClientID]; ok && lastOp.SeqNum >= op.SeqNum {
        // 已应用过,返回缓存结果
        result = lastOp.Result
    } else {
        // 应用到状态机
        result = kv.applyOp(op)

        // 记录
        kv.lastApplied[op.ClientID] = &OpContext{
            SeqNum: op.SeqNum,
            Result: result,
        }
    }

    // 通知等待的RPC
    if ch, ok := kv.notifyCh[msg.CommandIndex]; ok {
        ch <- result
    }

    // 检查是否需要快照
    if kv.maxraftstate > 0 && kv.rf.RaftStateSize() > kv.maxraftstate {
        kv.takeSnapshot(msg.CommandIndex)
    }
}

func (kv *KVServer) applyOp(op Op) OpResult {
    result := OpResult{Err: OK}

    switch op.Type {
    case OpGet:
        if value, ok := kv.db[op.Key]; ok {
            result.Value = value
        } else {
            result.Err = ErrNoKey
        }
    case OpPut:
        kv.db[op.Key] = op.Value
    case OpAppend:
        kv.db[op.Key] += op.Value
    }

    return result
}
```

**线性一致性读优化**:
```go
// 方案1: Read through Raft (强一致,但慢)
func (kv *KVServer) GetStrong(args *GetArgs, reply *GetReply) {
    // 将Read也走Raft log
    op := Op{
        Type:     OpGet,
        Key:      args.Key,
        ClientID: args.ClientID,
        SeqNum:   args.SeqNum,
    }

    // 走和Put相同的流程
    // ...
}

// 方案2: Read Index (优化,不写日志)
func (kv *KVServer) GetOptimized(args *GetArgs, reply *GetReply) {
    // 1. 获取当前commitIndex
    readIndex := kv.rf.GetCommitIndex()

    // 2. 发送心跳确认仍是leader
    if !kv.rf.SendHeartbeat() {
        reply.Err = ErrWrongLeader
        return
    }

    // 3. 等待applyIndex >= readIndex
    for kv.getApplyIndex() < readIndex {
        time.Sleep(10 * time.Millisecond)
    }

    // 4. 直接读取状态机
    kv.mu.Lock()
    if value, ok := kv.db[args.Key]; ok {
        reply.Value = value
        reply.Err = OK
    } else {
        reply.Err = ErrNoKey
    }
    kv.mu.Unlock()
}
```

**快照实现**:
```go
func (kv *KVServer) takeSnapshot(index int) {
    w := new(bytes.Buffer)
    e := gob.NewEncoder(w)

    // 序列化状态机
    e.Encode(kv.db)
    e.Encode(kv.lastApplied)

    snapshot := w.Bytes()
    kv.rf.Snapshot(index, snapshot)
}

func (kv *KVServer) applySnapshot(snapshot []byte) {
    if snapshot == nil || len(snapshot) == 0 {
        return
    }

    r := bytes.NewBuffer(snapshot)
    d := gob.NewDecoder(r)

    kv.mu.Lock()
    defer kv.mu.Unlock()

    d.Decode(&kv.db)
    d.Decode(&kv.lastApplied)
}
```

**客户端实现**:
```go
type Clerk struct {
    servers  []*RaftClient
    clientID int64
    seqNum   int64
    leaderID int
}

func (ck *Clerk) Get(key string) string {
    args := GetArgs{
        Key:      key,
        ClientID: ck.clientID,
        SeqNum:   atomic.AddInt64(&ck.seqNum, 1),
    }

    for {
        reply := GetReply{}
        ok := ck.servers[ck.leaderID].Call("KVServer.Get", &args, &reply)

        if ok && reply.Err == OK {
            return reply.Value
        }

        if reply.Err == ErrNoKey {
            return ""
        }

        // 切换leader重试
        ck.leaderID = (ck.leaderID + 1) % len(ck.servers)
        time.Sleep(100 * time.Millisecond)
    }
}

func (ck *Clerk) Put(key, value string) {
    ck.PutAppend(key, value, OpPut)
}

func (ck *Clerk) Append(key, value string) {
    ck.PutAppend(key, value, OpAppend)
}

func (ck *Clerk) PutAppend(key, value string, op OpType) {
    args := PutAppendArgs{
        Key:      key,
        Value:    value,
        Op:       op,
        ClientID: ck.clientID,
        SeqNum:   atomic.AddInt64(&ck.seqNum, 1),
    }

    for {
        reply := PutAppendReply{}
        ok := ck.servers[ck.leaderID].Call("KVServer.PutAppend", &args, &reply)

        if ok && reply.Err == OK {
            return
        }

        // 切换leader重试
        ck.leaderID = (ck.leaderID + 1) % len(ck.servers)
        time.Sleep(100 * time.Millisecond)
    }
}
```

**对应DDIA概念**:
- 第9章: 分布式事务
- 第7章: 线性一致性
- 第8章: 幂等性设计
- 第9章: Total Order Broadcast
- 第3章: 快照 (Snapshot)

**测试重点**:
```bash
# 基本功能
go test -run 3A

# 快照
go test -run 3B

# 压力测试
go test -run 3A -race -count 10
go test -run 3B -race -count 10
```

**预计时间**: 4周

---

## 🚀 阶段5: 分布式事务 (对应DDIA第9章)

### 项目9: 实现两阶段提交 (2PC)

**目标**: 实现跨分片事务

**核心实现**:
- 事务协调器 (Transaction Coordinator)
- Prepare阶段
- Commit阶段
- 超时和恢复

**架构设计**:
```go
type TwoPhaseCommit struct {
    coordinator *Coordinator
    participants []*Participant
}

type Coordinator struct {
    txnLog map[string]*TxnRecord  // txnID -> record
    mu     sync.Mutex
}

type TxnRecord struct {
    TxnID        string
    Participants []string
    State        TxnState
    PrepareOK    map[string]bool  // participant -> prepared
}

type TxnState int
const (
    TxnInit TxnState = iota
    TxnPreparing
    TxnCommitted
    TxnAborted
)

type Participant struct {
    id     string
    db     map[string]string
    txnLog map[string]*ParticipantTxn
    mu     sync.Mutex
}

type ParticipantTxn struct {
    TxnID   string
    Writes  map[string]string  // 暂存的写入
    State   TxnState
}
```

**Prepare阶段**:
```go
func (c *Coordinator) ExecuteTransaction(txnID string, operations []Operation) error {
    // 1. 创建事务记录
    c.mu.Lock()
    txnRecord := &TxnRecord{
        TxnID:        txnID,
        Participants: getParticipants(operations),
        State:        TxnPreparing,
        PrepareOK:    make(map[string]bool),
    }
    c.txnLog[txnID] = txnRecord
    c.persist(txnRecord)  // 持久化
    c.mu.Unlock()

    // 2. Phase 1: Prepare
    prepareCh := make(chan PrepareResult, len(txnRecord.Participants))

    for _, participantID := range txnRecord.Participants {
        go func(pid string) {
            participant := c.getParticipant(pid)
            ops := getOperationsForParticipant(operations, pid)

            ok := participant.Prepare(txnID, ops)
            prepareCh <- PrepareResult{
                ParticipantID: pid,
                OK:            ok,
            }
        }(participantID)
    }

    // 收集Prepare结果
    allOK := true
    for i := 0; i < len(txnRecord.Participants); i++ {
        result := <-prepareCh
        txnRecord.PrepareOK[result.ParticipantID] = result.OK
        if !result.OK {
            allOK = false
        }
    }

    // 3. Phase 2: Commit or Abort
    c.mu.Lock()
    if allOK {
        txnRecord.State = TxnCommitted
        c.persist(txnRecord)
        c.mu.Unlock()

        // 并行提交
        for _, pid := range txnRecord.Participants {
            go func(participantID string) {
                participant := c.getParticipant(participantID)
                participant.Commit(txnID)
            }(pid)
        }

        return nil
    } else {
        txnRecord.State = TxnAborted
        c.persist(txnRecord)
        c.mu.Unlock()

        // 并行回滚
        for _, pid := range txnRecord.Participants {
            go func(participantID string) {
                participant := c.getParticipant(participantID)
                participant.Abort(txnID)
            }(pid)
        }

        return ErrTxnAborted
    }
}
```

**Participant实现**:
```go
func (p *Participant) Prepare(txnID string, operations []Operation) bool {
    p.mu.Lock()
    defer p.mu.Unlock()

    // 1. 验证操作可行性
    for _, op := range operations {
        if op.Type == OpUpdate {
            if _, exists := p.db[op.Key]; !exists {
                return false  // key不存在
            }
        }
        // 可以加更多验证: 约束检查等
    }

    // 2. 创建事务记录,暂存写入
    ptxn := &ParticipantTxn{
        TxnID:  txnID,
        Writes: make(map[string]string),
        State:  TxnPreparing,
    }

    for _, op := range operations {
        ptxn.Writes[op.Key] = op.Value
    }

    p.txnLog[txnID] = ptxn

    // 3. 持久化prepare状态 (关键!)
    p.persistTxn(ptxn)

    return true
}

func (p *Participant) Commit(txnID string) {
    p.mu.Lock()
    defer p.mu.Unlock()

    ptxn := p.txnLog[txnID]
    if ptxn == nil {
        return
    }

    // 1. 应用暂存的写入
    for key, value := range ptxn.Writes {
        p.db[key] = value
    }

    // 2. 更新状态
    ptxn.State = TxnCommitted
    p.persistTxn(ptxn)

    // 3. 清理事务记录
    delete(p.txnLog, txnID)
}

func (p *Participant) Abort(txnID string) {
    p.mu.Lock()
    defer p.mu.Unlock()

    ptxn := p.txnLog[txnID]
    if ptxn == nil {
        return
    }

    // 1. 丢弃暂存的写入
    ptxn.State = TxnAborted
    p.persistTxn(ptxn)

    // 2. 清理
    delete(p.txnLog, txnID)
}
```

**崩溃恢复**:
```go
func (c *Coordinator) Recover() {
    // 从日志恢复未完成的事务
    for txnID, txn := range c.txnLog {
        switch txn.State {
        case TxnPreparing:
            // Prepare阶段崩溃,回滚
            c.abortTransaction(txnID)

        case TxnCommitted:
            // Commit阶段崩溃,继续提交
            c.retryCommit(txnID)

        case TxnAborted:
            // Abort阶段崩溃,继续回滚
            c.retryAbort(txnID)
        }
    }
}

func (p *Participant) Recover() {
    for txnID, ptxn := range p.txnLog {
        if ptxn.State == TxnPreparing {
            // 询问coordinator决定
            decision := p.askCoordinator(txnID)

            if decision == TxnCommitted {
                p.Commit(txnID)
            } else {
                p.Abort(txnID)
            }
        }
    }
}
```

**对应DDIA概念**:
- 第9章: 两阶段提交
- 第9章: 分布式事务问题
- 第9章: Coordinator失败处理

**2PC的问题**:
```go
// 问题1: Coordinator单点故障
// 如果coordinator崩溃且日志丢失:
// → Participant永远等待,资源被锁住

// 解决: 使用Raft保证coordinator高可用
type RaftCoordinator struct {
    rf      *Raft
    applyCh chan ApplyMsg
    txnLog  map[string]*TxnRecord
}

// 问题2: 阻塞
// Prepare后,participant必须等待coordinator决定
// → 持有锁,阻塞其他事务

// 解决: 超时机制
func (p *Participant) PrepareWithTimeout(txnID string, ops []Operation) bool {
    if !p.Prepare(txnID, ops) {
        return false
    }

    // 设置超时,自动abort
    time.AfterFunc(30*time.Second, func() {
        p.mu.Lock()
        defer p.mu.Unlock()

        if ptxn := p.txnLog[txnID]; ptxn != nil && ptxn.State == TxnPreparing {
            p.Abort(txnID)
        }
    })

    return true
}
```

**预计时间**: 2周

---

### 项目10: 实现Percolator事务模型 (进阶)

**推荐**: 参考TiKV的实现

**核心概念**:
- 乐观事务
- 分布式死锁检测
- MVCC + 2PC结合

**Percolator事务流程**:
```go
type PercolatorTxn struct {
    startTS   uint64  // 事务开始时间戳
    commitTS  uint64  // 提交时间戳
    writes    map[string][]byte
    primary   string  // 主键
}

// Prewrite阶段: 写入所有key的lock
func (txn *PercolatorTxn) Prewrite() error {
    // 1. 选择primary key
    txn.primary = txn.selectPrimary()

    // 2. Prewrite所有key
    for key, value := range txn.writes {
        isPrimary := (key == txn.primary)

        err := txn.prewriteKey(key, value, isPrimary)
        if err != nil {
            // 冲突,回滚
            txn.cleanup()
            return err
        }
    }

    return nil
}

func (txn *PercolatorTxn) prewriteKey(key string, value []byte, isPrimary bool) error {
    // 1. 检查是否有lock (写写冲突)
    if lock := getLock(key); lock != nil {
        if lock.ts < txn.startTS {
            // 旧事务还在,等待或清理
            return ErrLockConflict
        }
    }

    // 2. 检查是否有新的write (写写冲突)
    if latestWrite := getLatestWrite(key); latestWrite != nil {
        if latestWrite.commitTS > txn.startTS {
            // 有更新的版本,冲突
            return ErrWriteConflict
        }
    }

    // 3. 写入lock和data
    lock := Lock{
        primary:   txn.primary,
        ts:        txn.startTS,
        isPrimary: isPrimary,
    }

    writeLock(key, lock)
    writeData(key, txn.startTS, value)

    return nil
}

// Commit阶段: 两阶段提交
func (txn *PercolatorTxn) Commit() error {
    // 1. 获取commitTS
    txn.commitTS = getTimestamp()

    // 2. Phase 1: 提交primary
    err := txn.commitPrimary()
    if err != nil {
        txn.cleanup()
        return err
    }

    // 3. Phase 2: 异步提交secondaries
    go txn.commitSecondaries()

    return nil
}

func (txn *PercolatorTxn) commitPrimary() error {
    key := txn.primary

    // 1. 检查lock是否还存在
    lock := getLock(key)
    if lock == nil || lock.ts != txn.startTS {
        return ErrLockNotFound
    }

    // 2. 写入write record (原子操作)
    write := Write{
        startTS:  txn.startTS,
        commitTS: txn.commitTS,
    }

    // 原子操作: 写write + 删lock
    return atomicCommit(key, write, lock)
}

func (txn *PercolatorTxn) commitSecondaries() {
    for key := range txn.writes {
        if key == txn.primary {
            continue
        }

        write := Write{
            startTS:  txn.startTS,
            commitTS: txn.commitTS,
        }

        atomicCommit(key, write, nil)
    }
}
```

**MVCC读取**:
```go
func Get(key string, readTS uint64) ([]byte, error) {
    // 1. 检查lock
    if lock := getLock(key); lock != nil {
        if lock.ts <= readTS {
            // lock的事务应该在readTS之前提交
            // 检查primary是否已提交
            if isCommitted := checkPrimaryCommitted(lock.primary); isCommitted {
                // 已提交,返回对应版本
                return getData(key, lock.ts)
            } else {
                // 未提交,忽略这个版本
            }
        }
    }

    // 2. 找到最新的committed write
    writes := getWrites(key)
    for _, write := range writes {
        if write.commitTS <= readTS {
            // 找到可见版本
            return getData(key, write.startTS)
        }
    }

    return nil, ErrNotFound
}
```

**数据布局** (基于LSM/RocksDB):
```
CF_DEFAULT:  // 数据
  key:startTS -> value

CF_LOCK:     // 锁
  key -> Lock{primary, ts, ...}

CF_WRITE:    // 提交记录
  key:commitTS -> Write{startTS, ...}
```

**对应DDIA概念**:
- 第7章: 乐观并发控制
- 第7章: MVCC
- 第9章: 分布式事务
- 第9章: 两阶段提交变种

**参考资源**:
- [Percolator论文](https://research.google/pubs/pub36726/)
- [TiKV源码](https://github.com/tikv/tikv)
- [TiKV事务模型文档](https://tikv.org/deep-dive/distributed-transaction/introduction/)

**预计时间**: 2-3周 (选做,难度较高)

---

## 📚 推荐学习路径时间表

### 第1-2个月: 基础数据结构
- **Week 1-2**: 项目1 - LSM-Tree
- **Week 3-4**: 项目2 - B+树
- **Week 5-6**: 项目3 - MVCC
- **Week 7-8**: 项目4 - 2PL

### 第3-4个月: 分布式基础
- **Week 9-10**: 项目5 - 分区和复制
- **Week 11-12**: 项目6 - Quorum
- **Week 13-16**: 阅读Raft论文,准备Lab 2

### 第5-6个月: Raft核心 ⭐
- **Week 17-18**: Lab 2A - Leader Election
- **Week 19-21**: Lab 2B - Log Replication
- **Week 22**: Lab 2C - Persistence
- **Week 23-24**: Lab 2D - Log Compaction

### 第7-8个月: 综合应用
- **Week 25-28**: 项目8 - 分布式KV
- **Week 29-30**: 项目9 - 2PC
- **Week 31-32**: 项目10 - Percolator (可选)

---

## 🛠️ 开发工具和测试

### 推荐工具

**并发测试**:
```bash
# Race detector
go test -race

# 压力测试
go test -run TestBasic -count 100

# 并行测试
go test -parallel 8
```

**性能分析**:
```bash
# CPU profiling
go test -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof
go tool pprof mem.prof

# Benchmark
go test -bench=. -benchmem
```

**压力测试工具**:
```bash
# go-stress-test
https://github.com/link1st/go-stress-testing

# vegeta (HTTP压测)
https://github.com/tsenart/vegeta
```

**分布式追踪**:
```bash
# OpenTelemetry
https://opentelemetry.io/

# Jaeger
https://www.jaegertracing.io/
```

### 测试策略

**单元测试**:
```go
func TestLSMTreeBasic(t *testing.T) {
    lsm := NewLSMTree()

    // 写入
    lsm.Put([]byte("key1"), []byte("value1"))

    // 读取
    val, err := lsm.Get([]byte("key1"))
    assert.NoError(t, err)
    assert.Equal(t, []byte("value1"), val)
}
```

**集成测试**:
```go
func TestRaftElection(t *testing.T) {
    // 创建3个节点
    nodes := make([]*Raft, 3)
    for i := 0; i < 3; i++ {
        nodes[i] = Make(...)
    }

    // 等待选出leader
    time.Sleep(2 * time.Second)

    // 验证只有一个leader
    leaders := 0
    for _, n := range nodes {
        if n.state == Leader {
            leaders++
        }
    }
    assert.Equal(t, 1, leaders)
}
```

**混沌测试**:
```go
type ChaosNetwork struct {
    nodes      []*Node
    partitions [][]int
    dropRate   float64
}

func (n *ChaosNetwork) Send(from, to int, msg Message) {
    // 随机延迟
    delay := rand.Intn(100)
    time.Sleep(time.Duration(delay) * time.Millisecond)

    // 随机丢包
    if rand.Float64() < n.dropRate {
        return
    }

    // 检查网络分区
    if n.isPartitioned(from, to) {
        return
    }

    n.nodes[to].Receive(msg)
}

func (n *ChaosNetwork) CreatePartition(group1, group2 []int) {
    n.partitions = [][]int{group1, group2}
}

func (n *ChaosNetwork) HealPartition() {
    n.partitions = nil
}

// 测试
func TestRaftUnderPartition(t *testing.T) {
    network := NewChaosNetwork(5)

    // 创建网络分区: {0,1} vs {2,3,4}
    network.CreatePartition([]int{0, 1}, []int{2, 3, 4})

    // 多数派应该能选出leader
    time.Sleep(3 * time.Second)

    // 验证
    majorityLeaders := 0
    for _, id := range []int{2, 3, 4} {
        if network.nodes[id].state == Leader {
            majorityLeaders++
        }
    }
    assert.Equal(t, 1, majorityLeaders)

    // 少数派不应该有leader
    for _, id := range []int{0, 1} {
        assert.NotEqual(t, Leader, network.nodes[id].state)
    }
}
```

**网络模拟**:
```go
// 模拟不同的网络条件
type NetworkSimulator struct {
    latency    time.Duration  // 延迟
    jitter     time.Duration  // 抖动
    packetLoss float64        // 丢包率
    bandwidth  int            // 带宽限制 (bytes/sec)
}

func (ns *NetworkSimulator) Send(data []byte) {
    // 带宽限制
    transmitTime := time.Duration(len(data)) * time.Second / time.Duration(ns.bandwidth)
    time.Sleep(transmitTime)

    // 延迟 + 抖动
    delay := ns.latency + time.Duration(rand.Int63n(int64(ns.jitter)))
    time.Sleep(delay)

    // 丢包
    if rand.Float64() < ns.packetLoss {
        return  // 丢弃
    }

    // 发送
    actualSend(data)
}
```

---

## 📖 配套学习资源

### 核心课程

**1. MIT 6.824 - Distributed Systems**
- 链接: https://pdos.csail.mit.edu/6.824/
- 内容: 最佳分布式系统课程
- Lab质量: ⭐⭐⭐⭐⭐
- 必做: Lab 2 (Raft), Lab 3 (KV), Lab 4 (Sharded KV)

**2. CMU 15-445 - Database Systems**
- 链接: https://15445.courses.cs.cmu.edu/
- 内容: 数据库内核实现
- Project: 实现B+树、Buffer Pool、事务管理

**3. PingCAP Talent Plan**
- 链接: https://github.com/pingcap/talent-plan
- 项目: TinyKV, TinySQL
- 特点: 生产级代码质量

### 必读论文

**存储引擎**:
- [LSM-Tree论文](https://www.cs.umb.edu/~poneil/lsmtree.pdf)
- [RocksDB设计](https://github.com/facebook/rocksdb/wiki)

**共识算法**:
- [Raft](https://raft.github.io/raft.pdf) ⭐ 必读
- [Paxos Made Simple](https://lamport.azurewebsites.net/pubs/paxos-simple.pdf)
- [ZAB (ZooKeeper)](https://marcoserafini.github.io/papers/zab.pdf)

**分布式事务**:
- [Percolator](https://research.google/pubs/pub36726/)
- [Spanner](https://research.google/pubs/pub39966/)
- [Calvin](http://cs.yale.edu/homes/thomson/publications/calvin-sigmod12.pdf)

**MVCC和隔离级别**:
- [A Critique of ANSI SQL Isolation Levels](https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/tr-95-51.pdf)

### 开源项目源码

**存储引擎**:
- [LevelDB](https://github.com/google/leveldb) - LSM-Tree参考实现
- [RocksDB](https://github.com/facebook/rocksdb) - 生产级LSM
- [BadgerDB](https://github.com/dgraph-io/badger) - Go语言LSM

**Raft实现**:
- [etcd/raft](https://github.com/etcd-io/etcd/tree/main/raft) - 生产级Raft
- [hashicorp/raft](https://github.com/hashicorp/raft) - 易读的Raft

**分布式KV**:
- [TiKV](https://github.com/tikv/tikv) - 分布式事务KV
- [CockroachDB](https://github.com/cockroachdb/cockroach) - 分布式SQL

### 书籍

**DDIA (你已读完 ✓)**
- 最佳分布式系统入门书

**Database Internals**
- 作者: Alex Petrov
- 内容: 数据库内核实现细节
- 适合: 深入理解存储引擎

**Transaction Processing**
- 作者: Jim Gray
- 内容: 事务处理经典

---

## 🎯 学习建议

### 1. 循序渐进
不要跳过基础项目直接做Raft,扎实的基础会让后续学习事半功倍。

### 2. 重点在Raft
项目7 (Raft)是整个学习路径的核心,建议投入6-8周时间深入理解。Raft是理解分布式共识的最佳切入点。

### 3. 测试驱动
每个项目都要写充分的测试:
- 单元测试: 覆盖核心逻辑
- 集成测试: 测试节点间交互
- 混沌测试: 模拟故障场景

### 4. 阅读源码
在实现后阅读生产级实现的源码:
- etcd的Raft
- TiKV的MVCC和事务
- RocksDB的LSM-Tree

对比自己的实现,学习工程实践。

### 5. 写博客总结
每完成一个项目,写一篇技术博客:
- 巩固理解
- 锻炼表达能力
- 建立个人品牌

### 6. 加入社区
- 参与etcd/TiKV等项目的讨论
- 阅读设计文档和RFC
- 贡献代码

### 7. 性能优化
基本功能实现后,尝试优化:
- 使用pprof分析性能瓶颈
- 减少锁竞争
- 批处理优化
- 并行化

### 8. 故障演练
故意引入故障,测试系统行为:
- 杀掉Leader
- 网络分区
- 磁盘故障
- 时钟偏移

---

## 💡 常见问题

### Q1: 我应该从哪个项目开始?
A: 从项目1 (LSM-Tree)开始。即使你更感兴趣分布式,也建议先打好存储引擎基础。

### Q2: Raft太难了,可以跳过吗?
A: 不建议。Raft是理解分布式系统的关键。建议:
1. 先看Raft论文
2. 看Raft可视化动画
3. 参考etcd源码
4. 一点一点实现

### Q3: 需要完成所有项目吗?
A: 核心项目 (1-8) 强烈建议完成。项目9-10可根据兴趣选做。

### Q4: Go语言不熟悉怎么办?
A: 建议先学习Go基础:
- [Go Tour](https://tour.golang.org/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- 1-2周即可掌握基础

### Q5: 如何验证实现正确性?
A: 三层测试:
1. 单元测试: 测试单个组件
2. 集成测试: 测试多节点交互
3. 混沌测试: 随机故障注入

MIT 6.824提供了完整的测试套件。

### Q6: 性能达不到生产级怎么办?
A: 学习阶段重点是正确性,不是性能。理解原理后,可以:
1. 阅读生产级实现优化技巧
2. 使用profiling工具分析瓶颈
3. 逐步优化

---

## 🚀 进阶方向

完成核心项目后,可以选择以下方向深入:

### 方向1: 分布式SQL
- 项目: 实现分布式SQL引擎
- 参考: CockroachDB, TiDB
- 技术: SQL解析, 查询优化, 分布式执行

### 方向2: 分布式流处理
- 项目: 实现流处理引擎
- 参考: Apache Flink, Apache Kafka Streams
- 技术: Watermark, 状态管理, Exactly-once语义

### 方向3: 分布式缓存
- 项目: 实现分布式缓存系统
- 参考: Redis Cluster, Memcached
- 技术: 一致性哈希, 缓存淘汰, 复制

### 方向4: 云原生存储
- 项目: 实现对象存储或块存储
- 参考: MinIO, Ceph
- 技术: Erasure Coding, 元数据管理, 多租户

---

## 📈 学习成果检验

完成学习路径后,你应该能够:

**理论理解**:
- ✅ 解释LSM-Tree和B+Tree的权衡
- ✅ 理解MVCC的实现原理
- ✅ 解释CAP定理和实际应用
- ✅ 理解Raft共识算法
- ✅ 解释两阶段提交的问题

**实践能力**:
- ✅ 从零实现一个存储引擎
- ✅ 实现Raft并通过所有测试
- ✅ 构建分布式KV存储
- ✅ 实现分布式事务

**工程能力**:
- ✅ 编写可靠的并发代码
- ✅ 设计测试用例发现bug
- ✅ 使用profiling工具优化性能
- ✅ 阅读和理解生产级代码

**面试准备**:
- ✅ 能够设计一个分布式系统
- ✅ 回答常见的分布式面试题
- ✅ 有实际项目经验可以讨论

---

祝学习顺利! 🎉

如有问题,欢迎查阅DDIA对应章节或查看项目的参考资源。记住:分布式系统很难,但一步一步来,你一定能掌握!

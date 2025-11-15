package main

import (
	"fmt"
	"hash/crc32"
	"sort"
	"strconv"
)

type ConsistentHash struct {
	circle       map[uint32]string // 哈希环
	sortedHashes []uint32          // 排序的哈希值
	virtualNodes int               // 虚拟节点数量
}

func NewConsistentHash(virtualNodes int) *ConsistentHash {
	return &ConsistentHash{
		circle:       make(map[uint32]string),
		sortedHashes: []uint32{},
		virtualNodes: virtualNodes,
	}
}

// 添加节点
func (c *ConsistentHash) AddNode(node string) {
	for i := 0; i < c.virtualNodes; i++ {
		virtualKey := node + "#" + strconv.Itoa(i)
		hash := crc32.ChecksumIEEE([]byte(virtualKey))
		c.circle[hash] = node
		c.sortedHashes = append(c.sortedHashes, hash)
	}
	sort.Slice(c.sortedHashes, func(i, j int) bool {
		return c.sortedHashes[i] < c.sortedHashes[j]
	})
}

// 移除节点
func (c *ConsistentHash) RemoveNode(node string) {
	for i := 0; i < c.virtualNodes; i++ {
		virtualKey := node + "#" + strconv.Itoa(i)
		hash := crc32.ChecksumIEEE([]byte(virtualKey))
		delete(c.circle, hash)
	}

	// 重建排序列表
	c.sortedHashes = make([]uint32, 0, len(c.circle))
	for h := range c.circle {
		c.sortedHashes = append(c.sortedHashes, h)
	}
	sort.Slice(c.sortedHashes, func(i, j int) bool {
		return c.sortedHashes[i] < c.sortedHashes[j]
	})
}

// 获取数据应该存储的节点
func (c *ConsistentHash) GetNode(key string) string {
	if len(c.circle) == 0 {
		return ""
	}

	hash := crc32.ChecksumIEEE([]byte(key))

	// 二分查找第一个大于等于hash的节点
	idx := sort.Search(len(c.sortedHashes), func(i int) bool {
		return c.sortedHashes[i] >= hash
	})

	// 如果没找到，返回第一个节点（环形）
	if idx == len(c.sortedHashes) {
		idx = 0
	}

	return c.circle[c.sortedHashes[idx]]
}

func main() {
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("一致性哈希演示")
	fmt.Println("=" + string(make([]byte, 70)) + "=")

	ch := NewConsistentHash(150) // 每个物理节点150个虚拟节点

	// 添加服务器节点
	nodes := []string{"server1", "server2", "server3"}
	for _, node := range nodes {
		ch.AddNode(node)
		fmt.Printf("添加节点: %s\n", node)
	}

	// 测试数据分布
	fmt.Println("\n初始数据分布:")
	testKeys := []string{
		"user:1001", "user:1002", "user:1003", "user:1004", "user:1005",
		"user:1006", "user:1007", "user:1008", "user:1009", "user:1010",
	}

	distribution := make(map[string][]string)
	for _, key := range testKeys {
		node := ch.GetNode(key)
		distribution[node] = append(distribution[node], key)
		fmt.Printf("%s -> %s\n", key, node)
	}

	fmt.Println("\n节点负载统计:")
	for _, node := range nodes {
		keys := distribution[node]
		fmt.Printf("%s: %d keys %v\n", node, len(keys), keys)
	}

	// 演示节点增加的场景
	fmt.Println("\n" + "=" + string(make([]byte, 70)) + "=")
	fmt.Println("场景1: 添加新节点 server4")
	fmt.Println("=" + string(make([]byte, 70)) + "=")

	ch.AddNode("server4")
	newDistribution := make(map[string][]string)
	migrations := 0

	for _, key := range testKeys {
		newNode := ch.GetNode(key)
		oldNode := distribution[ch.GetNode(key)]
		newDistribution[newNode] = append(newDistribution[newNode], key)

		if len(oldNode) > 0 && oldNode[0] != key {
			// 检查是否发生了迁移
			found := false
			for _, k := range oldNode {
				if k == key {
					found = true
					break
				}
			}
			if found {
				fmt.Printf("%s: %s -> %s (迁移)\n", key, oldNode[0], newNode)
				migrations++
			}
		}
	}

	fmt.Println("\n新的节点负载统计:")
	allNodes := append(nodes, "server4")
	for _, node := range allNodes {
		keys := newDistribution[node]
		fmt.Printf("%s: %d keys\n", node, len(keys))
	}
	fmt.Printf("\n迁移的键数量: %d / %d (%.1f%%)\n", migrations, len(testKeys), float64(migrations)/float64(len(testKeys))*100)

	// 演示节点移除的场景
	fmt.Println("\n" + "=" + string(make([]byte, 70)) + "=")
	fmt.Println("场景2: 移除节点 server2")
	fmt.Println("=" + string(make([]byte, 70)) + "=")

	ch.RemoveNode("server2")
	finalDistribution := make(map[string][]string)

	for _, key := range testKeys {
		node := ch.GetNode(key)
		finalDistribution[node] = append(finalDistribution[node], key)
		fmt.Printf("%s -> %s\n", key, node)
	}

	fmt.Println("\n最终节点负载统计:")
	remainingNodes := []string{"server1", "server3", "server4"}
	for _, node := range remainingNodes {
		keys := finalDistribution[node]
		fmt.Printf("%s: %d keys\n", node, len(keys))
	}

	// 一致性哈希的优势说明
	fmt.Println("\n" + "=" + string(make([]byte, 70)) + "=")
	fmt.Println("一致性哈希的优势:")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("1. 节点增减时，只需迁移部分数据（约 1/N）")
	fmt.Println("2. 虚拟节点技术保证负载均衡")
	fmt.Println("3. 去中心化，无需全局协调")
	fmt.Println("4. 广泛应用于分布式缓存、负载均衡、分布式存储")
	fmt.Println("\n应用场景:")
	fmt.Println("- Memcached、Redis 集群")
	fmt.Println("- CDN 内容分发")
	fmt.Println("- 分布式数据库分片")
	fmt.Println("- 负载均衡器")
}

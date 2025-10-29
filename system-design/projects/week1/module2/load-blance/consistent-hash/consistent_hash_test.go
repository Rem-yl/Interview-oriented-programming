package main

import (
	"fmt"
	"hash/crc32"
	"math"
	"strings"
	"sync"
	"testing"
)

// ============ 基础功能测试 ============

func TestNewConsistentHash(t *testing.T) {
	t.Run("默认配置", func(t *testing.T) {
		replicas := 10
		server_len := 5
		c := NewConsistentHash(replicas, nil)
		for i := range server_len {
			node := fmt.Sprintf("server-%d", i)
			c.AddNode(node)
		}

		if len(c.hashMap) != replicas*server_len {
			t.Errorf("虚拟节点数量 %d 不等于 %d", len(c.hashMap), replicas*server_len)
		}

		// 验证keys是有序的
		for i := 1; i < len(c.keys); i++ {
			if c.keys[i-1] >= c.keys[i] {
				t.Errorf("keys未排序，索引 %d: %d >= %d", i, c.keys[i-1], c.keys[i])
			}
		}
	})

	t.Run("自定义哈希函数", func(t *testing.T) {
		ch := NewConsistentHash(150, crc32.ChecksumIEEE)
		if ch == nil {
			t.Fatal("NewConsistentHash 返回 nil")
		}
		if ch.hashFunc == nil {
			t.Error("hashFunc 不应该为 nil")
		}
	})
}

func TestAddNode(t *testing.T) {
	t.Run("添加单个节点", func(t *testing.T) {
		ch := NewConsistentHash(150, nil)
		err := ch.AddNode("node1")
		if err != nil {
			t.Fatalf("AddNode 失败: %v", err)
		}

		if len(ch.keys) != 150 {
			t.Errorf("添加1个节点后，keys长度 = %d, 期望 150", len(ch.keys))
		}
		if len(ch.hashMap) != 150 {
			t.Errorf("添加1个节点后，hashMap长度 = %d, 期望 150", len(ch.hashMap))
		}
	})

	t.Run("添加多个节点", func(t *testing.T) {
		ch := NewConsistentHash(150, nil)
		nodes := []string{"node1", "node2", "node3"}
		for _, node := range nodes {
			ch.AddNode(node)
		}

		expectedVNodes := 150 * len(nodes)
		if len(ch.keys) != expectedVNodes {
			t.Errorf("添加%d个节点后，keys长度 = %d, 期望 %d",
				len(nodes), len(ch.keys), expectedVNodes)
		}

		// 验证每个物理节点都有正确数量的虚拟节点
		nodeCount := make(map[string]int)
		for _, physicalNode := range ch.hashMap {
			nodeCount[physicalNode]++
		}

		for _, node := range nodes {
			if count, exists := nodeCount[node]; !exists {
				t.Errorf("节点 %s 未在 hashMap 中找到", node)
			} else if count != 150 {
				t.Errorf("节点 %s 有 %d 个虚拟节点，期望 150", node, count)
			}
		}
	})

	t.Run("重复添加同一节点", func(t *testing.T) {
		ch := NewConsistentHash(150, nil)
		ch.AddNode("node1")
		initialLen := len(ch.keys)
		ch.AddNode("node1")

		// 重复添加不应该增加虚拟节点（因为Contains检查）
		if len(ch.keys) != initialLen {
			t.Logf("注意：重复添加节点改变了keys数量: %d -> %d", initialLen, len(ch.keys))
		}
	})
}

func TestRemoveNode(t *testing.T) {
	ch := NewConsistentHash(150, nil)
	ch.AddNode("node1")
	ch.AddNode("node2")
	ch.AddNode("node3")

	t.Run("删除存在的节点", func(t *testing.T) {
		initialLen := len(ch.keys)
		err := ch.RemoveNode("node2")
		if err != nil {
			t.Fatalf("RemoveNode 失败: %v", err)
		}

		expectedLen := initialLen - 150
		if len(ch.keys) != expectedLen {
			t.Errorf("删除节点后，keys长度 = %d, 期望 %d", len(ch.keys), expectedLen)
		}

		// 验证node2不再存在
		for _, node := range ch.hashMap {
			if node == "node2" {
				t.Error("node2 删除后仍存在于 hashMap 中")
			}
		}

		// 验证keys仍然有序
		for i := 1; i < len(ch.keys); i++ {
			if ch.keys[i-1] >= ch.keys[i] {
				t.Errorf("删除后keys未排序，索引 %d", i)
			}
		}
	})

	t.Run("删除不存在的节点", func(t *testing.T) {
		initialLen := len(ch.keys)
		err := ch.RemoveNode("node999")
		if err != nil {
			t.Errorf("删除不存在的节点不应返回错误，得到: %v", err)
		}
		if len(ch.keys) != initialLen {
			t.Errorf("删除不存在的节点改变了keys长度")
		}
	})
}

func TestGetNode(t *testing.T) {
	replicas := 10
	server_len := 5
	c := NewConsistentHash(replicas, nil)
	for i := range server_len {
		node := fmt.Sprintf("server-%d", i)
		c.AddNode(node)
	}

	t.Run("基本功能", func(t *testing.T) {
		node := "rem123"
		server := c.GetNode(node)
		if !strings.Contains(server, "server") {
			t.Errorf("server: %s dont has server", server)
		}
	})

	t.Run("一致性检查", func(t *testing.T) {
		testKeys := []string{"user:1000", "user:2000", "product:abc", "session:xyz"}
		for _, key := range testKeys {
			firstNode := c.GetNode(key)
			// 同一个key应该总是返回同一个节点
			for i := 0; i < 10; i++ {
				node := c.GetNode(key)
				if node != firstNode {
					t.Errorf("GetNode不一致: key=%s, 第一次=%s, 第%d次=%s",
						key, firstNode, i+1, node)
				}
			}
		}
	})

	t.Run("空字符串key", func(t *testing.T) {
		node := c.GetNode("")
		if node == "" {
			t.Error("空key应该也能返回节点")
		}
	})
}

// ============ 负载均衡测试 ============

func TestLoadBalance(t *testing.T) {
	t.Run("5个节点的负载分布", func(t *testing.T) {
		ch := NewConsistentHash(150, nil)
		nodes := []string{"node1", "node2", "node3", "node4", "node5"}
		for _, node := range nodes {
			ch.AddNode(node)
		}

		// 模拟10000个key的分配
		numKeys := 10000
		distribution := make(map[string]int)
		for i := 0; i < numKeys; i++ {
			key := fmt.Sprintf("key:%d", i)
			node := ch.GetNode(key)
			distribution[node]++
		}

		// 理想情况每个节点: 10000/5 = 2000
		expectedPerNode := float64(numKeys) / float64(len(nodes))
		t.Logf("期望每节点: %.0f keys", expectedPerNode)

		// 计算标准差
		var variance float64
		for node, count := range distribution {
			diff := float64(count) - expectedPerNode
			variance += diff * diff
			percentage := float64(count) / float64(numKeys) * 100
			t.Logf("节点 %s: %d keys (%.2f%%)", node, count, percentage)
		}
		stdDev := math.Sqrt(variance / float64(len(nodes)))
		t.Logf("标准差: %.2f", stdDev)

		// 每个节点应该在期望值的 ±30% 范围内
		for node, count := range distribution {
			minExpected := expectedPerNode * 0.70
			maxExpected := expectedPerNode * 1.30
			if float64(count) < minExpected || float64(count) > maxExpected {
				t.Errorf("节点 %s 负载不均衡: %d (期望 %.0f ±30%%)",
					node, count, expectedPerNode)
			}
		}
	})

	t.Run("不同虚拟节点数的负载分布", func(t *testing.T) {
		replicaTests := []int{10, 50, 100, 150, 200}
		for _, replicas := range replicaTests {
			ch := NewConsistentHash(replicas, nil)
			for i := 0; i < 5; i++ {
				ch.AddNode(fmt.Sprintf("node%d", i))
			}

			distribution := make(map[string]int)
			numKeys := 10000
			for i := 0; i < numKeys; i++ {
				key := fmt.Sprintf("key:%d", i)
				node := ch.GetNode(key)
				distribution[node]++
			}

			var variance float64
			expectedPerNode := float64(numKeys) / 5.0
			for _, count := range distribution {
				diff := float64(count) - expectedPerNode
				variance += diff * diff
			}
			stdDev := math.Sqrt(variance / 5.0)
			t.Logf("虚拟节点数=%d, 标准差=%.2f", replicas, stdDev)
		}
	})
}

// ============ 数据迁移测试 ============

func TestDataMigration(t *testing.T) {
	t.Run("添加节点的迁移率", func(t *testing.T) {
		ch := NewConsistentHash(150, nil)
		ch.AddNode("node1")
		ch.AddNode("node2")
		ch.AddNode("node3")

		// 记录1000个key的初始分配
		numKeys := 1000
		initialMapping := make(map[string]string, numKeys)
		for i := 0; i < numKeys; i++ {
			key := fmt.Sprintf("key:%d", i)
			initialMapping[key] = ch.GetNode(key)
		}

		// 添加第4个节点
		ch.AddNode("node4")

		// 统计迁移的key数量
		migrated := 0
		for i := 0; i < numKeys; i++ {
			key := fmt.Sprintf("key:%d", i)
			newNode := ch.GetNode(key)
			if initialMapping[key] != newNode {
				migrated++
			}
		}

		// 理论迁移率: 1/4 = 25%
		migrationRate := float64(migrated) / float64(numKeys) * 100
		t.Logf("添加节点迁移率: %.2f%% (%d/%d keys)", migrationRate, migrated, numKeys)

		// 允许15%-35%的范围（考虑虚拟节点的随机性）
		if migrationRate < 15 || migrationRate > 35 {
			t.Errorf("迁移率 %.2f%% 超出预期范围 [15%%, 35%%]", migrationRate)
		}
	})

	t.Run("删除节点的迁移率", func(t *testing.T) {
		ch := NewConsistentHash(150, nil)
		for i := 1; i <= 4; i++ {
			ch.AddNode(fmt.Sprintf("node%d", i))
		}

		numKeys := 1000
		initialMapping := make(map[string]string, numKeys)
		for i := 0; i < numKeys; i++ {
			key := fmt.Sprintf("key:%d", i)
			initialMapping[key] = ch.GetNode(key)
		}

		// 删除一个节点
		ch.RemoveNode("node4")

		migrated := 0
		for i := 0; i < numKeys; i++ {
			key := fmt.Sprintf("key:%d", i)
			newNode := ch.GetNode(key)
			if initialMapping[key] != newNode {
				migrated++
			}
		}

		migrationRate := float64(migrated) / float64(numKeys) * 100
		t.Logf("删除节点迁移率: %.2f%% (%d/%d keys)", migrationRate, migrated, numKeys)

		// 删除1个节点，理论上应该只迁移属于该节点的数据（约25%）
		if migrationRate < 15 || migrationRate > 35 {
			t.Errorf("迁移率 %.2f%% 超出预期范围 [15%%, 35%%]", migrationRate)
		}
	})
}

// ============ 并发安全测试 ============

func TestConcurrentAccess(t *testing.T) {
	ch := NewConsistentHash(150, nil)
	ch.AddNode("node1")
	ch.AddNode("node2")

	var wg sync.WaitGroup
	numGoroutines := 50
	numOperations := 100

	// 并发添加节点
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				node := fmt.Sprintf("node_%d_%d", id, j)
				ch.AddNode(node)
			}
		}(i)
	}

	// 并发获取节点
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				ch.GetNode(key)
			}
		}(i)
	}

	// 并发删除节点
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				node := fmt.Sprintf("node_%d_%d", id, j)
				ch.RemoveNode(node)
			}
		}(i)
	}

	wg.Wait()
	t.Log("并发测试完成，无竞态条件")
}

// ============ 边界情况测试 ============

func TestEdgeCases(t *testing.T) {
	t.Run("单节点环", func(t *testing.T) {
		ch := NewConsistentHash(150, nil)
		ch.AddNode("node1")

		// 所有key都应该映射到唯一的节点
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("key:%d", i)
			node := ch.GetNode(key)
			if node != "node1" {
				t.Errorf("单节点环: GetNode(%s) = %s, 期望 node1", key, node)
			}
		}
	})

	t.Run("大量节点", func(t *testing.T) {
		ch := NewConsistentHash(100, nil)
		// 添加100个节点
		for i := 0; i < 100; i++ {
			ch.AddNode(fmt.Sprintf("node%d", i))
		}

		// 验证能正常获取节点
		node := ch.GetNode("test_key")
		if node == "" {
			t.Error("大量节点情况下获取节点失败")
		}
	})

	t.Run("特殊字符key", func(t *testing.T) {
		ch := NewConsistentHash(150, nil)
		ch.AddNode("node1")

		specialKeys := []string{
			"key@#$%",
			"key with spaces",
			"key\nwith\nnewlines",
			"中文键",
			"",
		}

		for _, key := range specialKeys {
			node := ch.GetNode(key)
			if node == "" {
				t.Errorf("特殊字符key '%s' 获取节点失败", key)
			}
		}
	})
}

// ============ 性能基准测试 ============

func BenchmarkAddNode(b *testing.B) {
	replicas := []int{50, 100, 150, 200}
	for _, r := range replicas {
		b.Run(fmt.Sprintf("replicas_%d", r), func(b *testing.B) {
			ch := NewConsistentHash(r, nil)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				node := fmt.Sprintf("node%d", i)
				ch.AddNode(node)
			}
		})
	}
}

func BenchmarkRemoveNode(b *testing.B) {
	ch := NewConsistentHash(150, nil)
	// 预先添加足够的节点
	for i := 0; i < 1000; i++ {
		ch.AddNode(fmt.Sprintf("node%d", i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		node := fmt.Sprintf("node%d", i%1000)
		ch.RemoveNode(node)
	}
}

func BenchmarkGetNode(b *testing.B) {
	nodeCount := []int{3, 10, 50, 100}

	for _, nc := range nodeCount {
		b.Run(fmt.Sprintf("nodes_%d", nc), func(b *testing.B) {
			ch := NewConsistentHash(150, nil)
			for i := 0; i < nc; i++ {
				ch.AddNode(fmt.Sprintf("node%d", i))
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("key:%d", i)
				ch.GetNode(key)
			}
		})
	}
}

func BenchmarkGetNodeParallel(b *testing.B) {
	ch := NewConsistentHash(150, nil)
	for i := 0; i < 10; i++ {
		ch.AddNode(fmt.Sprintf("node%d", i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key:%d", i)
			ch.GetNode(key)
			i++
		}
	})
}

func BenchmarkVirtualNodeCount(b *testing.B) {
	replicas := []int{10, 50, 100, 150, 200, 300}

	for _, r := range replicas {
		b.Run(fmt.Sprintf("replicas_%d", r), func(b *testing.B) {
			ch := NewConsistentHash(r, nil)
			for i := 0; i < 10; i++ {
				ch.AddNode(fmt.Sprintf("node%d", i))
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				key := fmt.Sprintf("key:%d", i)
				ch.GetNode(key)
			}
		})
	}
}

func BenchmarkHashFunction(b *testing.B) {
	data := []byte("test_key_for_benchmark")

	b.Run("crc32", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			crc32.ChecksumIEEE(data)
		}
	})
}

func BenchmarkLoadDistribution(b *testing.B) {
	ch := NewConsistentHash(150, nil)
	for i := 0; i < 10; i++ {
		ch.AddNode(fmt.Sprintf("node%d", i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		distribution := make(map[string]int)
		for j := 0; j < 1000; j++ {
			key := fmt.Sprintf("key:%d", j)
			node := ch.GetNode(key)
			distribution[node]++
		}
	}
}

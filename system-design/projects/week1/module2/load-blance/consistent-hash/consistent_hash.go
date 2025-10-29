package main

import (
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
)

type ConsistentHash struct {
	hashFunc func(data []byte) uint32
	replicas int // 每个物理节点的虚拟节点数
	keys     []uint32
	hashMap  map[uint32]string // 哈希值与节点关键字对应
	mutex    sync.Mutex
}

func Contains[T comparable](s []T, value T) bool {
	for _, v := range s {
		if v == value {
			return true
		}
	}

	return false
}

func (c *ConsistentHash) AddNode(node string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i := range c.replicas {
		key := node + fmt.Sprintf("#vnode%d", i)
		hashValue := c.hashFunc([]byte(key))
		if Contains(c.keys, hashValue) {
			continue
		}
		c.keys = append(c.keys, hashValue)
		c.hashMap[hashValue] = node
	}

	sort.Slice(c.keys, func(i, j int) bool {
		return c.keys[i] < c.keys[j]
	})

	return nil
}

func (c *ConsistentHash) deleteHashValue(hashValue uint32) {
	for i := 0; i < len(c.keys); i++ {
		if c.keys[i] == hashValue {
			c.keys = append(c.keys[:i], c.keys[i+1:]...)
			break
		}
	}
}

func (c *ConsistentHash) RemoveNode(node string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i := range c.replicas {
		key := node + fmt.Sprintf("#vnode%d", i)
		hashValue := c.hashFunc([]byte(key))
		c.deleteHashValue(hashValue)
		delete(c.hashMap, hashValue)
	}

	sort.Slice(c.keys, func(i, j int) bool {
		return c.keys[i] < c.keys[j]
	})

	return nil
}

func (c *ConsistentHash) GetNode(key string) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	hashValue := c.hashFunc([]byte(key))

	idx := sort.Search(len(c.keys), func(i int) bool {
		return c.keys[i] >= hashValue
	})

	if idx == len(c.keys) {
		idx = 0
	}

	return c.hashMap[c.keys[idx]]
}

func (c *ConsistentHash) GetNodes() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	nodeSet := make(map[string]bool) // 使用 set 去重
	for _, node := range c.hashMap {
		nodeSet[node] = true // 直接使用物理节点名
	}

	nodes := make([]string, 0, len(nodeSet))
	for node := range nodeSet {
		nodes = append(nodes, node)
	}
	return nodes
}

func NewConsistentHash(replicas int, hashFunc func(data []byte) uint32) *ConsistentHash {
	if hashFunc == nil {
		hashFunc = crc32.ChecksumIEEE
	}

	return &ConsistentHash{
		hashFunc: hashFunc,
		replicas: replicas,
		keys:     make([]uint32, 0),
		hashMap:  make(map[uint32]string, 0),
	}
}

func main() {
	fmt.Println("hello")
}

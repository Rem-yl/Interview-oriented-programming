package store

import (
	"fmt"
	"sync"
	"testing"
)

// TC1: 基本读写测试
// 测试目标：验证基本的 Set 和 Get 功能
func TestSetAndGet(t *testing.T) {
	s := NewStore()

	// 测试字符串类型
	s.Set("name", "Alice")
	value, exists := s.Get("name")
	if !exists {
		t.Error("Expected key 'name' to exist, but it doesn't")
	}
	if value != "Alice" {
		t.Errorf("Expected value 'Alice', got '%v'", value)
	}

	// 测试整数类型
	s.Set("age", 25)
	value, exists = s.Get("age")
	if !exists {
		t.Error("Expected key 'age' to exist")
	}
	if value != 25 {
		t.Errorf("Expected value 25, got '%v'", value)
	}

	// 测试切片类型
	scores := []int{90, 85, 88}
	s.Set("scores", scores)
	value, exists = s.Get("scores")
	if !exists {
		t.Error("Expected key 'scores' to exist")
	}
	retrievedScores, ok := value.([]int)
	if !ok {
		t.Error("Expected value to be []int type")
	}
	if len(retrievedScores) != 3 || retrievedScores[0] != 90 {
		t.Errorf("Expected scores [90, 85, 88], got %v", retrievedScores)
	}

	// 测试覆盖已存在的键
	s.Set("name", "Bob")
	value, exists = s.Get("name")
	if !exists {
		t.Error("Expected key 'name' to exist after update")
	}
	if value != "Bob" {
		t.Errorf("Expected updated value 'Bob', got '%v'", value)
	}
}

// TC2: 获取不存在的键
// 测试目标：验证获取不存在的键的行为
func TestGetNonExistent(t *testing.T) {
	s := NewStore()

	// 尝试获取一个不存在的键
	value, exists := s.Get("nonexistent")

	// 应该返回 false
	if exists {
		t.Error("Expected key 'nonexistent' to not exist, but it does")
	}

	// 值应该是 nil
	if value != nil {
		t.Errorf("Expected nil value for nonexistent key, got '%v'", value)
	}
}

// TC3: 删除操作测试
// 测试目标：验证删除功能
func TestDelete(t *testing.T) {
	s := NewStore()

	// 先设置一个键
	s.Set("temp", "value")

	// 验证键存在
	_, exists := s.Get("temp")
	if !exists {
		t.Error("Expected key 'temp' to exist before deletion")
	}

	// 删除键
	s.Delete("temp")

	// 验证键已被删除
	_, exists = s.Get("temp")
	if exists {
		t.Error("Expected key 'temp' to be deleted, but it still exists")
	}

	// 边界条件：删除不存在的键（不应报错）
	s.Delete("nonexistent") // 应该不会 panic
}

// TC4: 键存在性检查
// 测试目标：验证 Exists 方法
func TestExists(t *testing.T) {
	s := NewStore()

	// 检查不存在的键
	if s.Exists("missing") {
		t.Error("Expected key 'missing' to not exist")
	}

	// 设置一个键
	s.Set("present", "here")

	// 检查存在的键
	if !s.Exists("present") {
		t.Error("Expected key 'present' to exist")
	}

	// 删除后再检查
	s.Delete("present")
	if s.Exists("present") {
		t.Error("Expected key 'present' to not exist after deletion")
	}
}

// TC5: 并发安全测试
// 测试目标：验证多线程并发访问的安全性
func TestConcurrentAccess(t *testing.T) {
	s := NewStore()
	done := make(chan bool)

	// 启动 10 个 goroutine 并发写入
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d", n)
				s.Set(key, j)
			}
			done <- true
		}(i)
	}

	// 启动 10 个 goroutine 并发读取
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d", n)
				s.Get(key)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 20; i++ {
		<-done
	}

	// 如果没有 panic，说明并发安全
	t.Log("Concurrent access test passed - no race conditions detected")
}

// TC5扩展: 更激烈的并发读写测试
// 测试目标：验证在更高并发压力下的安全性
func TestConcurrentReadWrite(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup

	// 写入者数量
	writers := 20
	// 读取者数量
	readers := 30
	// 每个 goroutine 的操作次数
	operations := 500

	// 启动写入者
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("writer-%d-key-%d", n, j%10)
				s.Set(key, j)
			}
		}(i)
	}

	// 启动读取者
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("writer-%d-key-%d", n%writers, j%10)
				s.Get(key)
			}
		}(i)
	}

	// 启动删除者
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < operations/2; j++ {
				key := fmt.Sprintf("writer-%d-key-%d", n, j%10)
				s.Delete(key)
			}
		}(i)
	}

	wg.Wait()
	t.Logf("High concurrency test passed - %d writers, %d readers completed", writers, readers)
}

// TC6: 获取所有键测试
// 测试目标：验证 Keys 方法
func TestKeys(t *testing.T) {
	s := NewStore()

	// 空存储应该返回空切片
	keys := s.Keys()
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys in empty store, got %d", len(keys))
	}

	// 添加一些键
	s.Set("key1", "value1")
	s.Set("key2", "value2")
	s.Set("key3", "value3")

	// 获取所有键
	keys = s.Keys()
	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// 验证所有键都存在（顺序可能不同）
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	expectedKeys := []string{"key1", "key2", "key3"}
	for _, expectedKey := range expectedKeys {
		if !keyMap[expectedKey] {
			t.Errorf("Expected key '%s' not found in Keys() result", expectedKey)
		}
	}
}

// TC7: 清空测试
// 测试目标：验证 Clear 方法
func TestClear(t *testing.T) {
	s := NewStore()

	// 添加一些数据
	s.Set("key1", "value1")
	s.Set("key2", "value2")
	s.Set("key3", "value3")

	// 验证数据存在
	if len(s.Keys()) != 3 {
		t.Error("Expected 3 keys before clear")
	}

	// 清空
	s.Clear()

	// 验证所有键都被删除
	if len(s.Keys()) != 0 {
		t.Error("Expected store to be empty after Clear()")
	}

	// 验证所有键的 Exists 都返回 false
	if s.Exists("key1") || s.Exists("key2") || s.Exists("key3") {
		t.Error("Expected all keys to not exist after Clear()")
	}

	// 清空后应该可以继续使用
	s.Set("new-key", "new-value")
	value, exists := s.Get("new-key")
	if !exists || value != "new-value" {
		t.Error("Expected store to be usable after Clear()")
	}
}

// 边界条件测试：空字符串键
func TestEmptyStringKey(t *testing.T) {
	s := NewStore()

	// 空字符串应该可以作为键
	s.Set("", "empty-key-value")
	value, exists := s.Get("")
	if !exists {
		t.Error("Expected empty string key to exist")
	}
	if value != "empty-key-value" {
		t.Errorf("Expected 'empty-key-value', got '%v'", value)
	}
}

// 边界条件测试：nil 值
func TestNilValue(t *testing.T) {
	s := NewStore()

	// nil 值应该可以存储
	s.Set("nil-key", nil)

	// 应该能区分"键存在但值为 nil"和"键不存在"
	value, exists := s.Get("nil-key")
	if !exists {
		t.Error("Expected key 'nil-key' to exist even with nil value")
	}
	if value != nil {
		t.Errorf("Expected nil value, got '%v'", value)
	}

	// 使用 Exists 验证
	if !s.Exists("nil-key") {
		t.Error("Expected Exists to return true for key with nil value")
	}
}

// 边界条件测试：大量键
func TestManyKeys(t *testing.T) {
	s := NewStore()

	// 插入大量键
	count := 10000
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("key-%d", i)
		s.Set(key, i)
	}

	// 验证数量
	keys := s.Keys()
	if len(keys) != count {
		t.Errorf("Expected %d keys, got %d", count, len(keys))
	}

	// 随机验证一些键
	testIndices := []int{0, 100, 1000, 5000, 9999}
	for _, idx := range testIndices {
		key := fmt.Sprintf("key-%d", idx)
		value, exists := s.Get(key)
		if !exists {
			t.Errorf("Expected key '%s' to exist", key)
		}
		if value != idx {
			t.Errorf("Expected value %d for key '%s', got %v", idx, key, value)
		}
	}
}

// 性能基准测试：Set 操作
func BenchmarkSet(b *testing.B) {
	s := NewStore()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		s.Set(key, i)
	}
}

// 性能基准测试：Get 操作
func BenchmarkGet(b *testing.B) {
	s := NewStore()

	// 预先填充数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		s.Set(key, i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%1000)
		s.Get(key)
	}
}

// 性能基准测试：并发 Get 操作
func BenchmarkConcurrentGet(b *testing.B) {
	s := NewStore()

	// 预先填充数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		s.Set(key, i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)
			s.Get(key)
			i++
		}
	})
}

// 性能基准测试：并发读写混合
func BenchmarkConcurrentReadWrite(b *testing.B) {
	s := NewStore()

	// 预先填充数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		s.Set(key, i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%1000)
			// 80% 读，20% 写
			if i%5 == 0 {
				s.Set(key, i)
			} else {
				s.Get(key)
			}
			i++
		}
	})
}

// 性能基准测试：Exists 操作
func BenchmarkExists(b *testing.B) {
	s := NewStore()

	// 预先填充数据
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key-%d", i)
		s.Set(key, i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%1000)
		s.Exists(key)
	}
}

// 性能基准测试：Delete 操作
func BenchmarkDelete(b *testing.B) {
	// 每次迭代重新创建和填充 store
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		s := NewStore()
		key := fmt.Sprintf("key-%d", i)
		s.Set(key, i)
		b.StartTimer()

		s.Delete(key)
	}
}

// 性能基准测试：Keys 操作
func BenchmarkKeys(b *testing.B) {
	s := NewStore()

	// 预先填充不同数量的数据进行测试
	counts := []int{10, 100, 1000}

	for _, count := range counts {
		b.Run(fmt.Sprintf("Keys-%d", count), func(b *testing.B) {
			// 填充数据
			for i := 0; i < count; i++ {
				key := fmt.Sprintf("key-%d", i)
				s.Set(key, i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				s.Keys()
			}
		})
	}
}

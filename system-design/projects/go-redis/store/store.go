package store

import (
	"go-redis/logger"
	"sync"

	"github.com/sirupsen/logrus"
)

// Store 是一个线程安全的内存键值存储。
// 它使用读写锁（RWMutex）来保证并发访问的安全性。
// 支持任意类型的值（interface{}）。
type Store struct {
	mu   sync.RWMutex          // 读写锁
	data map[string]interface{} // 数据存储
}

// NewStore 创建一个新的 Store 实例
func NewStore() *Store {
	logger.Debug("创建新的 Store 实例")
	return &Store{
		data: make(map[string]interface{}),
	}
}

// Set 设置键值对
func (s *Store) Set(key string, value interface{}) {
	logger.WithFields(logrus.Fields{
		"operation": "SET",
		"key":       key,
	}).Debug("执行 Set 操作")

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value

	logger.WithField("key", key).Debug("Set 操作完成")
}

// Get 获取指定键的值
// 返回值和是否存在的布尔值
func (s *Store) Get(key string) (interface{}, bool) {
	logger.WithFields(logrus.Fields{
		"operation": "GET",
		"key":       key,
	}).Debug("执行 Get 操作")

	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]

	logger.WithFields(logrus.Fields{
		"key":    key,
		"exists": exists,
	}).Debug("Get 操作完成")

	return value, exists
}

// Delete 删除指定的键
func (s *Store) Delete(key string) {
	logger.WithFields(logrus.Fields{
		"operation": "DELETE",
		"key":       key,
	}).Debug("执行 Delete 操作")

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)

	logger.WithField("key", key).Debug("Delete 操作完成")
}

// Exists 检查键是否存在
func (s *Store) Exists(key string) bool {
	logger.WithFields(logrus.Fields{
		"operation": "EXISTS",
		"key":       key,
	}).Debug("执行 Exists 操作")

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.data[key]

	logger.WithFields(logrus.Fields{
		"key":    key,
		"exists": exists,
	}).Debug("Exists 操作完成")

	return exists
}

// Keys 返回所有键的切片
func (s *Store) Keys() []string {
	logger.WithField("operation", "KEYS").Debug("执行 Keys 操作")

	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}

	logger.WithFields(logrus.Fields{
		"count": len(keys),
	}).Debug("Keys 操作完成")

	return keys
}

// Clear 清空所有数据
func (s *Store) Clear() {
	logger.WithField("operation", "CLEAR").Debug("执行 Clear 操作")

	s.mu.Lock()
	defer s.mu.Unlock()

	oldCount := len(s.data)
	s.data = make(map[string]interface{})

	logger.WithFields(logrus.Fields{
		"cleared_count": oldCount,
	}).Info("Clear 操作完成，已清空所有数据")
}

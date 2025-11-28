package handler

import (
	"go-redis/protocol"
	"go-redis/store"
	"strings"
)

type KeysHandler struct {
	db *store.Store
}

func NewKeysHandler(db *store.Store) *KeysHandler {
	return &KeysHandler{
		db: db,
	}
}

// Handle 处理 KEYS 命令
// KEYS pattern - 查找所有匹配模式的键
func (h *KeysHandler) Handle(args []protocol.Value) *protocol.Value {
	// KEYS 需要恰好 1 个参数：pattern
	if len(args) != 1 {
		return protocol.Error("ERR wrong number of arguments for 'keys' command")
	}

	pattern := args[0].Str

	// 获取所有键
	allKeys := h.db.Keys()

	// 过滤匹配模式的键
	matchedKeys := make([]protocol.Value, 0)
	for _, key := range allKeys {
		if matchPattern(pattern, key) {
			matchedKeys = append(matchedKeys, protocol.Value{
				Type: protocol.BulkStringType,
				Str:  key,
			})
		}
	}

	// 返回数组类型
	return &protocol.Value{
		Type:  protocol.ArrayType,
		Array: matchedKeys,
	}
}

// matchPattern 实现简单的通配符匹配
// 支持的模式：
// - "*" 匹配所有字符串
// - "prefix*" 匹配以 prefix 开头的字符串
// - "*suffix" 匹配以 suffix 结尾的字符串
// - "exact" 精确匹配
func matchPattern(pattern, str string) bool {
	// "*" 匹配所有
	if pattern == "*" {
		return true
	}

	// "*suffix" - 匹配以 suffix 结尾的字符串
	if strings.HasPrefix(pattern, "*") {
		suffix := pattern[1:]
		return strings.HasSuffix(str, suffix)
	}

	// "prefix*" - 匹配以 prefix 开头的字符串
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(str, prefix)
	}

	// 精确匹配
	return pattern == str
}

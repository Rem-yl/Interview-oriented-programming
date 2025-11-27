package protocol

import (
	"io"
	"strings"
	"testing"

	"go-redis/logger"

	"github.com/sirupsen/logrus"
)

func init() {
	// 测试时禁用日志输出
	logger.SetOutput(io.Discard)
	logger.SetLevel(logrus.ErrorLevel)
}

// TestSerializeSimpleString 测试简单字符串序列化
func TestSerializeSimpleString(t *testing.T) {
	tests := []struct {
		name     string
		value    *Value
		expected string
	}{
		{
			name: "simple OK",
			value: &Value{
				Type: StringType,
				Str:  "OK",
			},
			expected: "+OK\r\n",
		},
		{
			name: "simple PONG",
			value: &Value{
				Type: StringType,
				Str:  "PONG",
			},
			expected: "+PONG\r\n",
		},
		{
			name: "empty string",
			value: &Value{
				Type: StringType,
				Str:  "",
			},
			expected: "+\r\n",
		},
		{
			name: "string with spaces",
			value: &Value{
				Type: StringType,
				Str:  "Hello World",
			},
			expected: "+Hello World\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Serialize(tt.value)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSerializeBulkString 测试批量字符串序列化
func TestSerializeBulkString(t *testing.T) {
	tests := []struct {
		name     string
		value    *Value
		expected string
	}{
		{
			name: "simple bulk string",
			value: &Value{
				Type: BulkStringType,
				Str:  "foobar",
			},
			expected: "$6\r\nfoobar\r\n",
		},
		{
			name: "bulk string with spaces",
			value: &Value{
				Type: BulkStringType,
				Str:  "hello world",
			},
			expected: "$11\r\nhello world\r\n",
		},
		{
			name: "bulk string with newline",
			value: &Value{
				Type: BulkStringType,
				Str:  "hello\r\nworld",
			},
			expected: "$12\r\nhello\r\nworld\r\n",
		},
		{
			name: "empty bulk string",
			value: &Value{
				Type: BulkStringType,
				Str:  "",
			},
			expected: "$0\r\n\r\n",
		},
		{
			name: "null bulk string",
			value: &Value{
				Type:   BulkStringType,
				IsNull: true,
			},
			expected: "$-1\r\n",
		},
		{
			name: "bulk string with UTF-8",
			value: &Value{
				Type: BulkStringType,
				Str:  "你好世界",
			},
			expected: "$12\r\n你好世界\r\n",
		},
		{
			name: "bulk string with special chars",
			value: &Value{
				Type: BulkStringType,
				Str:  "!@#$%^&*()_+",
			},
			expected: "$12\r\n!@#$%^&*()_+\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Serialize(tt.value)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSerializeError 测试错误序列化
func TestSerializeError(t *testing.T) {
	tests := []struct {
		name     string
		value    *Value
		expected string
	}{
		{
			name: "simple error",
			value: &Value{
				Type: ErrorType,
				Str:  "ERR unknown command",
			},
			expected: "-ERR unknown command\r\n",
		},
		{
			name: "WRONGTYPE error",
			value: &Value{
				Type: ErrorType,
				Str:  "WRONGTYPE Operation against a key holding the wrong kind of value",
			},
			expected: "-WRONGTYPE Operation against a key holding the wrong kind of value\r\n",
		},
		{
			name: "empty error",
			value: &Value{
				Type: ErrorType,
				Str:  "",
			},
			expected: "-\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Serialize(tt.value)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSerializeInteger 测试整数序列化
func TestSerializeInteger(t *testing.T) {
	tests := []struct {
		name     string
		value    *Value
		expected string
	}{
		{
			name: "zero",
			value: &Value{
				Type: IntType,
				Int:  0,
			},
			expected: ":0\r\n",
		},
		{
			name: "positive integer",
			value: &Value{
				Type: IntType,
				Int:  100,
			},
			expected: ":100\r\n",
		},
		{
			name: "negative integer",
			value: &Value{
				Type: IntType,
				Int:  -50,
			},
			expected: ":-50\r\n",
		},
		{
			name: "large integer",
			value: &Value{
				Type: IntType,
				Int:  9223372036854775807,
			},
			expected: ":9223372036854775807\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Serialize(tt.value)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSerializeArray 测试数组序列化
func TestSerializeArray(t *testing.T) {
	tests := []struct {
		name     string
		value    *Value
		expected string
	}{
		{
			name: "empty array",
			value: &Value{
				Type:  ArrayType,
				Array: []Value{},
			},
			expected: "*0\r\n",
		},
		{
			name: "null array",
			value: &Value{
				Type:   ArrayType,
				IsNull: true,
			},
			expected: "*-1\r\n",
		},
		{
			name: "array with two bulk strings",
			value: &Value{
				Type: ArrayType,
				Array: []Value{
					{Type: BulkStringType, Str: "foo"},
					{Type: BulkStringType, Str: "bar"},
				},
			},
			expected: "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
		},
		{
			name: "array with integers",
			value: &Value{
				Type: ArrayType,
				Array: []Value{
					{Type: IntType, Int: 1},
					{Type: IntType, Int: 2},
					{Type: IntType, Int: 3},
				},
			},
			expected: "*3\r\n:1\r\n:2\r\n:3\r\n",
		},
		{
			name: "mixed array",
			value: &Value{
				Type: ArrayType,
				Array: []Value{
					{Type: IntType, Int: 1},
					{Type: BulkStringType, Str: "foo"},
					{Type: StringType, Str: "OK"},
				},
			},
			expected: "*3\r\n:1\r\n$3\r\nfoo\r\n+OK\r\n",
		},
		{
			name: "nested array",
			value: &Value{
				Type: ArrayType,
				Array: []Value{
					{
						Type: ArrayType,
						Array: []Value{
							{Type: IntType, Int: 1},
							{Type: IntType, Int: 2},
						},
					},
					{
						Type: ArrayType,
						Array: []Value{
							{Type: BulkStringType, Str: "foo"},
						},
					},
				},
			},
			expected: "*2\r\n*2\r\n:1\r\n:2\r\n*1\r\n$3\r\nfoo\r\n",
		},
		{
			name: "array with null element",
			value: &Value{
				Type: ArrayType,
				Array: []Value{
					{Type: BulkStringType, Str: "foo"},
					{Type: BulkStringType, IsNull: true},
					{Type: BulkStringType, Str: "bar"},
				},
			},
			expected: "*3\r\n$3\r\nfoo\r\n$-1\r\n$3\r\nbar\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Serialize(tt.value)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSerializeRedisCommands 测试真实的 Redis 命令序列化
func TestSerializeRedisCommands(t *testing.T) {
	tests := []struct {
		name     string
		value    *Value
		expected string
	}{
		{
			name: "SET command",
			value: &Value{
				Type: ArrayType,
				Array: []Value{
					{Type: BulkStringType, Str: "SET"},
					{Type: BulkStringType, Str: "key"},
					{Type: BulkStringType, Str: "value"},
				},
			},
			expected: "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
		},
		{
			name: "GET command",
			value: &Value{
				Type: ArrayType,
				Array: []Value{
					{Type: BulkStringType, Str: "GET"},
					{Type: BulkStringType, Str: "key"},
				},
			},
			expected: "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n",
		},
		{
			name: "DEL command with multiple keys",
			value: &Value{
				Type: ArrayType,
				Array: []Value{
					{Type: BulkStringType, Str: "DEL"},
					{Type: BulkStringType, Str: "key1"},
					{Type: BulkStringType, Str: "key2"},
					{Type: BulkStringType, Str: "key3"},
				},
			},
			expected: "*4\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n$4\r\nkey3\r\n",
		},
		{
			name: "PING response",
			value: &Value{
				Type: StringType,
				Str:  "PONG",
			},
			expected: "+PONG\r\n",
		},
		{
			name: "SET OK response",
			value: &Value{
				Type: StringType,
				Str:  "OK",
			},
			expected: "+OK\r\n",
		},
		{
			name: "GET response (string value)",
			value: &Value{
				Type: BulkStringType,
				Str:  "value",
			},
			expected: "$5\r\nvalue\r\n",
		},
		{
			name: "GET response (null - key not exist)",
			value: &Value{
				Type:   BulkStringType,
				IsNull: true,
			},
			expected: "$-1\r\n",
		},
		{
			name: "DEL response (count)",
			value: &Value{
				Type: IntType,
				Int:  2,
			},
			expected: ":2\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Serialize(tt.value)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestRoundTrip 测试往返一致性：parse → serialize → parse
func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple string",
			input: "+OK\r\n",
		},
		{
			name:  "error",
			input: "-ERR unknown command\r\n",
		},
		{
			name:  "integer",
			input: ":100\r\n",
		},
		{
			name:  "bulk string",
			input: "$6\r\nfoobar\r\n",
		},
		{
			name:  "null bulk string",
			input: "$-1\r\n",
		},
		{
			name:  "empty bulk string",
			input: "$0\r\n\r\n",
		},
		{
			name:  "array",
			input: "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
		},
		{
			name:  "empty array",
			input: "*0\r\n",
		},
		{
			name:  "null array",
			input: "*-1\r\n",
		},
		{
			name:  "nested array",
			input: "*2\r\n*2\r\n:1\r\n:2\r\n*1\r\n$3\r\nfoo\r\n",
		},
		{
			name:  "SET command",
			input: "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 第一步：解析原始输入
			parser := NewParser(strings.NewReader(tt.input))
			value, err := parser.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			// 第二步：序列化
			serialized := Serialize(value)

			// 第三步：验证序列化结果与原始输入一致
			if serialized != tt.input {
				t.Errorf("roundtrip failed:\noriginal: %q\nserialized: %q", tt.input, serialized)
			}

			// 第四步：再次解析序列化的结果
			parser2 := NewParser(strings.NewReader(serialized))
			value2, err := parser2.Parse()
			if err != nil {
				t.Fatalf("second parse error: %v", err)
			}

			// 第五步：比较两次解析的结果
			if !valuesEqual(value, value2) {
				t.Errorf("values not equal after roundtrip")
			}
		})
	}
}

// valuesEqual 比较两个 Value 是否相等
func valuesEqual(v1, v2 *Value) bool {
	if v1.Type != v2.Type {
		return false
	}
	if v1.IsNull != v2.IsNull {
		return false
	}
	if v1.Str != v2.Str {
		return false
	}
	if v1.Int != v2.Int {
		return false
	}
	if len(v1.Array) != len(v2.Array) {
		return false
	}
	for i := range v1.Array {
		if !valuesEqual(&v1.Array[i], &v2.Array[i]) {
			return false
		}
	}
	return true
}

// Benchmark 性能测试

func BenchmarkSerializeSimpleString(b *testing.B) {
	value := &Value{
		Type: StringType,
		Str:  "OK",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Serialize(value)
	}
}

func BenchmarkSerializeBulkString(b *testing.B) {
	value := &Value{
		Type: BulkStringType,
		Str:  "hello world",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Serialize(value)
	}
}

func BenchmarkSerializeInteger(b *testing.B) {
	value := &Value{
		Type: IntType,
		Int:  100,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Serialize(value)
	}
}

func BenchmarkSerializeArray(b *testing.B) {
	value := &Value{
		Type: ArrayType,
		Array: []Value{
			{Type: BulkStringType, Str: "SET"},
			{Type: BulkStringType, Str: "key"},
			{Type: BulkStringType, Str: "value"},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Serialize(value)
	}
}

func BenchmarkSerializeNestedArray(b *testing.B) {
	value := &Value{
		Type: ArrayType,
		Array: []Value{
			{
				Type: ArrayType,
				Array: []Value{
					{Type: IntType, Int: 1},
					{Type: IntType, Int: 2},
				},
			},
			{
				Type: ArrayType,
				Array: []Value{
					{Type: BulkStringType, Str: "foo"},
					{Type: BulkStringType, Str: "bar"},
				},
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Serialize(value)
	}
}

package protocol

import (
	"io"
	"strings"
	"testing"

	"go-redis/logger"

	"github.com/sirupsen/logrus"
)

func init() {
	// 测试时禁用日志输出，避免干扰测试结果
	logger.SetOutput(io.Discard)
	logger.SetLevel(logrus.ErrorLevel)
}

// TestParseSimpleString 测试简单字符串解析
func TestParseSimpleString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "normal string",
			input:    "OK\r\n",
			expected: "OK",
			wantErr:  false,
		},
		{
			name:     "string with spaces",
			input:    "hello world\r\n",
			expected: "hello world",
			wantErr:  false,
		},
		{
			name:     "empty string",
			input:    "\r\n",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "long string",
			input:    "this is a very long simple string for testing\r\n",
			expected: "this is a very long simple string for testing",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			parser := NewParser(reader)
			value, err := parser.parseSimpleString()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if value.Type != StringType {
				t.Errorf("expected type StringType, got %v", value.Type)
			}

			if value.Str != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, value.Str)
			}

			if value.IsNull {
				t.Errorf("expected IsNull to be false")
			}
		})
	}
}

// TestParseError 测试错误类型解析
func TestParseError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple error",
			input:    "Error message\r\n",
			expected: "Error message",
			wantErr:  false,
		},
		{
			name:     "ERR prefix error",
			input:    "ERR unknown command 'foobar'\r\n",
			expected: "ERR unknown command 'foobar'",
			wantErr:  false,
		},
		{
			name:     "WRONGTYPE error",
			input:    "WRONGTYPE Operation against a key holding the wrong kind of value\r\n",
			expected: "WRONGTYPE Operation against a key holding the wrong kind of value",
			wantErr:  false,
		},
		{
			name:     "empty error",
			input:    "\r\n",
			expected: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			parser := NewParser(reader)
			value, err := parser.parseErr()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if value.Type != ErrorType {
				t.Errorf("expected type ErrorType, got %v", value.Type)
			}

			if value.Str != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, value.Str)
			}

			if value.IsNull {
				t.Errorf("expected IsNull to be false")
			}
		})
	}
}

// TestParseInteger 测试整数解析
func TestParseInteger(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		{
			name:     "positive integer",
			input:    "100\r\n",
			expected: 100,
			wantErr:  false,
		},
		{
			name:     "negative integer",
			input:    "-50\r\n",
			expected: -50,
			wantErr:  false,
		},
		{
			name:     "zero",
			input:    "0\r\n",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "large integer",
			input:    "9223372036854775807\r\n",
			expected: 9223372036854775807,
			wantErr:  false,
		},
		{
			name:     "invalid integer - multiple numbers",
			input:    "100 200\r\n",
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "invalid integer - text",
			input:    "abc\r\n",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			parser := NewParser(reader)
			value, err := parser.parseInt()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if value.Type != IntType {
				t.Errorf("expected type IntType, got %v", value.Type)
			}

			if value.Int != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, value.Int)
			}

			if value.IsNull {
				t.Errorf("expected IsNull to be false")
			}
		})
	}
}

// TestParseBulkString 测试批量字符串解析
func TestParseBulkString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isNull   bool
		wantErr  bool
	}{
		{
			name:     "simple bulk string",
			input:    "5\r\nhello\r\n",
			expected: "hello",
			isNull:   false,
			wantErr:  false,
		},
		{
			name:     "bulk string with spaces",
			input:    "11\r\nhello world\r\n",
			expected: "hello world",
			isNull:   false,
			wantErr:  false,
		},
		{
			name:     "bulk string with newlines",
			input:    "12\r\nhello\r\nworld\r\n",
			expected: "hello\r\nworld",
			isNull:   false,
			wantErr:  false,
		},
		{
			name:     "empty bulk string",
			input:    "0\r\n\r\n",
			expected: "",
			isNull:   false,
			wantErr:  false,
		},
		{
			name:     "null bulk string",
			input:    "-1\r\n",
			expected: "",
			isNull:   true,
			wantErr:  false,
		},
		{
			name:     "bulk string with special chars",
			input:    "12\r\n!@#$%^&*()_+\r\n",
			expected: "!@#$%^&*()_+",
			isNull:   false,
			wantErr:  false,
		},
		{
			name:     "bulk string with UTF-8",
			input:    "12\r\n你好世界\r\n",
			expected: "你好世界",
			isNull:   false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			parser := NewParser(reader)
			value, err := parser.parseBulkString()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.isNull {
				if value.Type != BulkStringType {
					t.Errorf("expected type NullType, got %v", value.Type)
				}
				if !value.IsNull {
					t.Errorf("expected IsNull to be true")
				}
			} else {
				if value.Type != BulkStringType {
					t.Errorf("expected type StringType, got %v", value.Type)
				}
				if value.Str != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, value.Str)
				}
				if value.IsNull {
					t.Errorf("expected IsNull to be false")
				}
			}
		})
	}
}

// TestParseArray 测试数组解析
func TestParseArray(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLen int
		isNull      bool
		wantErr     bool
		validate    func(*testing.T, *Value)
	}{
		{
			name:        "empty array",
			input:       "0\r\n",
			expectedLen: 0,
			isNull:      false,
			wantErr:     false,
			validate: func(t *testing.T, v *Value) {
				if len(v.Array) != 0 {
					t.Errorf("expected empty array, got length %d", len(v.Array))
				}
			},
		},
		{
			name:        "null array",
			input:       "-1\r\n",
			expectedLen: 0,
			isNull:      true,
			wantErr:     false,
			validate: func(t *testing.T, v *Value) {
				if !v.IsNull {
					t.Errorf("expected null array")
				}
			},
		},
		{
			name:        "array with two bulk strings",
			input:       "2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
			expectedLen: 2,
			isNull:      false,
			wantErr:     false,
			validate: func(t *testing.T, v *Value) {
				if len(v.Array) != 2 {
					t.Fatalf("expected array length 2, got %d", len(v.Array))
				}
				if v.Array[0].Str != "foo" {
					t.Errorf("expected first element 'foo', got %q", v.Array[0].Str)
				}
				if v.Array[1].Str != "bar" {
					t.Errorf("expected second element 'bar', got %q", v.Array[1].Str)
				}
			},
		},
		{
			name:        "array with integers",
			input:       "3\r\n:1\r\n:2\r\n:3\r\n",
			expectedLen: 3,
			isNull:      false,
			wantErr:     false,
			validate: func(t *testing.T, v *Value) {
				if len(v.Array) != 3 {
					t.Fatalf("expected array length 3, got %d", len(v.Array))
				}
				for i, val := range v.Array {
					if val.Int != int64(i+1) {
						t.Errorf("expected element %d to be %d, got %d", i, i+1, val.Int)
					}
				}
			},
		},
		{
			name:        "mixed array",
			input:       "3\r\n:1\r\n$3\r\nfoo\r\n+OK\r\n",
			expectedLen: 3,
			isNull:      false,
			wantErr:     false,
			validate: func(t *testing.T, v *Value) {
				if len(v.Array) != 3 {
					t.Fatalf("expected array length 3, got %d", len(v.Array))
				}
				if v.Array[0].Type != IntType || v.Array[0].Int != 1 {
					t.Errorf("expected first element to be int 1")
				}
				if v.Array[1].Type != BulkStringType || v.Array[1].Str != "foo" {
					t.Errorf("expected second element to be string 'foo'")
				}
				if v.Array[2].Type != StringType || v.Array[2].Str != "OK" {
					t.Errorf("expected third element to be string 'OK'")
				}
			},
		},
		{
			name:        "nested array",
			input:       "2\r\n*2\r\n:1\r\n:2\r\n*1\r\n$3\r\nfoo\r\n",
			expectedLen: 2,
			isNull:      false,
			wantErr:     false,
			validate: func(t *testing.T, v *Value) {
				if len(v.Array) != 2 {
					t.Fatalf("expected array length 2, got %d", len(v.Array))
				}
				// 第一个元素是数组 [1, 2]
				if v.Array[0].Type != ArrayType {
					t.Errorf("expected first element to be array")
				}
				if len(v.Array[0].Array) != 2 {
					t.Errorf("expected first nested array length 2")
				}
				// 第二个元素是数组 ["foo"]
				if v.Array[1].Type != ArrayType {
					t.Errorf("expected second element to be array")
				}
				if len(v.Array[1].Array) != 1 {
					t.Errorf("expected second nested array length 1")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			parser := NewParser(reader)
			value, err := parser.parseArray()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.isNull {
				if value.Type != ArrayType {
					t.Errorf("expected type NullType, got %v", value.Type)
				}
			} else {
				if value.Type != ArrayType {
					t.Errorf("expected type ArrayType, got %v", value.Type)
				}
			}

			if tt.validate != nil {
				tt.validate(t, value)
			}
		})
	}
}

// TestParse 测试完整的 Parse 函数
func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType ValueType
		wantErr  bool
		validate func(*testing.T, *Value)
	}{
		{
			name:     "simple string",
			input:    "+OK\r\n",
			wantType: StringType,
			wantErr:  false,
			validate: func(t *testing.T, v *Value) {
				if v.Str != "OK" {
					t.Errorf("expected 'OK', got %q", v.Str)
				}
			},
		},
		{
			name:     "error",
			input:    "-Error message\r\n",
			wantType: ErrorType,
			wantErr:  false,
			validate: func(t *testing.T, v *Value) {
				if v.Str != "Error message" {
					t.Errorf("expected 'Error message', got %q", v.Str)
				}
			},
		},
		{
			name:     "integer",
			input:    ":100\r\n",
			wantType: IntType,
			wantErr:  false,
			validate: func(t *testing.T, v *Value) {
				if v.Int != 100 {
					t.Errorf("expected 100, got %d", v.Int)
				}
			},
		},
		{
			name:     "bulk string",
			input:    "$5\r\nhello\r\n",
			wantType: BulkStringType,
			wantErr:  false,
			validate: func(t *testing.T, v *Value) {
				if v.Str != "hello" {
					t.Errorf("expected 'hello', got %q", v.Str)
				}
			},
		},
		{
			name:     "array",
			input:    "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
			wantType: ArrayType,
			wantErr:  false,
			validate: func(t *testing.T, v *Value) {
				if len(v.Array) != 2 {
					t.Errorf("expected array length 2, got %d", len(v.Array))
				}
			},
		},
		{
			name:     "unknown type",
			input:    "?invalid\r\n",
			wantType: "",
			wantErr:  true,
			validate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			parser := NewParser(reader)
			value, err := parser.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if value.Type != tt.wantType {
				t.Errorf("expected type %v, got %v", tt.wantType, value.Type)
			}

			if tt.validate != nil {
				tt.validate(t, value)
			}
		})
	}
}

// TestParseRealWorldCommands 测试真实的 Redis 命令
func TestParseRealWorldCommands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "PING command",
			input:    "*1\r\n$4\r\nPING\r\n",
			expected: []string{"PING"},
		},
		{
			name:     "SET command",
			input:    "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n",
			expected: []string{"SET", "key", "value"},
		},
		{
			name:     "GET command",
			input:    "*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n",
			expected: []string{"GET", "key"},
		},
		{
			name:     "DEL command with multiple keys",
			input:    "*4\r\n$3\r\nDEL\r\n$4\r\nkey1\r\n$4\r\nkey2\r\n$4\r\nkey3\r\n",
			expected: []string{"DEL", "key1", "key2", "key3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			parser := NewParser(reader)
			value, err := parser.Parse()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if value.Type != ArrayType {
				t.Fatalf("expected ArrayType, got %v", value.Type)
			}

			if len(value.Array) != len(tt.expected) {
				t.Fatalf("expected %d elements, got %d", len(tt.expected), len(value.Array))
			}

			for i, expected := range tt.expected {
				if value.Array[i].Str != expected {
					t.Errorf("element %d: expected %q, got %q", i, expected, value.Array[i].Str)
				}
			}

			t.Logf("Successfully parsed: %v", tt.expected)
		})
	}
}

// TestParseMultipleCommands 测试连续解析多个命令
func TestParseMultipleCommands(t *testing.T) {
	input := "+OK\r\n:100\r\n$5\r\nhello\r\n"
	reader := strings.NewReader(input)
	parser := NewParser(reader)

	// 第一个命令：+OK
	value1, err := parser.Parse()
	if err != nil {
		t.Fatalf("first parse error: %v", err)
	}
	if value1.Type != StringType || value1.Str != "OK" {
		t.Errorf("first command: expected StringType 'OK', got %v %q", value1.Type, value1.Str)
	}

	// 第二个命令：:100
	value2, err := parser.Parse()
	if err != nil {
		t.Fatalf("second parse error: %v", err)
	}
	if value2.Type != IntType || value2.Int != 100 {
		t.Errorf("second command: expected IntType 100, got %v %d", value2.Type, value2.Int)
	}

	// 第三个命令：$5\r\nhello\r\n
	value3, err := parser.Parse()
	if err != nil {
		t.Fatalf("third parse error: %v", err)
	}
	if value3.Type != BulkStringType || value3.Str != "hello" {
		t.Errorf("third command: expected StringType 'hello', got %v %q", value3.Type, value3.Str)
	}

	// 第四次应该返回 EOF
	_, err = parser.Parse()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}

	t.Log("Successfully parsed multiple consecutive commands")
}

// Benchmark 性能测试
func BenchmarkParseSimpleString(b *testing.B) {
	input := "+OK\r\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(input)
		parser := NewParser(reader)
		parser.Parse()
	}
}

func BenchmarkParseBulkString(b *testing.B) {
	input := "$11\r\nhello world\r\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(input)
		parser := NewParser(reader)
		parser.Parse()
	}
}

func BenchmarkParseArray(b *testing.B) {
	input := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(input)
		parser := NewParser(reader)
		parser.Parse()
	}
}

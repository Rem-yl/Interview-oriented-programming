package protocol

// SimpleString 创建简单字符串类型的 Value
// 用于返回简单的状态回复，如 "+OK\r\n"
func SimpleString(s string) *Value {
	return &Value{
		Type: StringType,
		Str:  s,
	}
}

// Error 创建错误类型的 Value
// 用于返回错误信息，如 "-ERR unknown command\r\n"
func Error(msg string) *Value {
	return &Value{
		Type: ErrorType,
		Str:  msg,
	}
}

// Integer 创建整数类型的 Value
// 用于返回整数结果，如 ":100\r\n"
func Integer(n int64) *Value {
	return &Value{
		Type: IntType,
		Int:  n,
	}
}

// BulkString 创建批量字符串类型的 Value
// 用于返回二进制安全的字符串，如 "$6\r\nfoobar\r\n"
func BulkString(s string) *Value {
	return &Value{
		Type: BulkStringType,
		Str:  s,
	}
}

// NullBulkString 创建 NULL 批量字符串
// 用于表示键不存在等情况，序列化为 "$-1\r\n"
func NullBulkString() *Value {
	return &Value{
		Type:   BulkStringType,
		IsNull: true,
	}
}

// Array 创建数组类型的 Value
// 用于返回多个值，如 "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
func Array(values []Value) *Value {
	return &Value{
		Type:  ArrayType,
		Array: values,
	}
}

// EmptyArray 创建空数组
// 序列化为 "*0\r\n"
func EmptyArray() *Value {
	return &Value{
		Type:  ArrayType,
		Array: []Value{},
	}
}

// NullArray 创建 NULL 数组
// 序列化为 "*-1\r\n"
func NullArray() *Value {
	return &Value{
		Type:   ArrayType,
		IsNull: true,
	}
}

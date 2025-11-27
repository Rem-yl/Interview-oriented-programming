package protocol

import "fmt"

func serializeSimpleString(v *Value) string {
	return fmt.Sprintf("+%s\r\n", v.Str)
}

func serializeBulkString(v *Value) string {
	if v.IsNull {
		return "$-1\r\n"
	}

	return fmt.Sprintf("$%d\r\n%s\r\n", len(v.Str), v.Str)
}

func serializeError(v *Value) string {
	return fmt.Sprintf("-%s\r\n", v.Str)
}

func serializeInt(v *Value) string {
	return fmt.Sprintf(":%d\r\n", v.Int)
}

func serializeArray(v *Value) string {
	if v.IsNull {
		return "*-1\r\n"
	}

	length := len(v.Array)
	res := fmt.Sprintf("*%d\r\n", length)

	for _, array := range v.Array {
		res += Serialize(&array)
	}

	return res
}

func Serialize(v *Value) string {
	switch v.Type {
	case StringType:
		return serializeSimpleString(v)
	case BulkStringType:
		return serializeBulkString(v)
	case ErrorType:
		return serializeError(v)
	case IntType:
		return serializeInt(v)
	case ArrayType:
		return serializeArray(v)
	default:
		return "-ERR unknown command\r\n"
	}
}

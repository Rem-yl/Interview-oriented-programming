package protocol

type ValueType string

const (
	StringType     ValueType = "string"
	BulkStringType ValueType = "bulkString"
	ErrorType      ValueType = "err"
	IntType        ValueType = "int"
	ArrayType      ValueType = "array"
)

type Value struct {
	Type   ValueType
	Str    string
	Int    int64
	Array  []Value
	IsNull bool
}

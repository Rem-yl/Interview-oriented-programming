package protocol

type ValueType string

const (
	StringType ValueType = "string"
	ErrorType  ValueType = "err"
	IntType    ValueType = "int"
	ArrayType  ValueType = "array"
	NullType   ValueType = "null"
)

type Value struct {
	Type   ValueType
	Str    string
	Int    int64
	Array  []Value
	IsNull bool
}

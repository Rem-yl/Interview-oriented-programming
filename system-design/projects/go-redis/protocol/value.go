package protocol

type Value struct {
	Type   string
	Str    string
	Int    int64
	Array  []Value
	IsNull bool
}

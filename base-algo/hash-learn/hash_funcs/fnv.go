package hashfuncs

import "encoding/binary"

type FNV struct{}

const (
	fnvOffsetBasis32 = 2166136261
	fnvPrime32       = 16777619
)

func NewFNV() *FNV {
	return &FNV{}
}

func (h *FNV) Sum(data []byte) ([]byte, error) {
	hash := uint32(fnvOffsetBasis32)
	for _, b := range data {
		hash ^= uint32(b)
		hash *= fnvPrime32
	}

	out := make([]byte, 4)
	binary.BigEndian.PutUint32(out, hash)

	return out, nil
}

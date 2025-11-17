package hashfuncs

import (
	"encoding/binary"
)

type DGB2 struct {
}

func NewDGB2() *DGB2 {
	return &DGB2{}
}

func (h *DGB2) Sum(data []byte) ([]byte, error) {
	hash := uint32(5381)

	for _, b := range data {
		hash = ((hash << 5) + hash) + uint32(b)
	}

	out := make([]byte, 4)
	binary.BigEndian.PutUint32(out, hash)
	return out, nil
}

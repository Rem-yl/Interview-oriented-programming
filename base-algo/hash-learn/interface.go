package hashlearn

type HashFunc interface {
	Sum(data []byte) ([]byte, error)
}

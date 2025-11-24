package backend

type BackEnd interface {
	GetURL() string
	GetName() string
	GetWeight() int
}

package backend

type SimpleBackEnd struct {
	URL  string
	Name string
}

func (b *SimpleBackEnd) GetURL() string {
	return b.URL
}

func (b *SimpleBackEnd) GetName() string {
	return b.Name
}

func NewSimpleBackEnd(URL, Name string) *SimpleBackEnd {
	return &SimpleBackEnd{
		URL:  URL,
		Name: Name,
	}
}

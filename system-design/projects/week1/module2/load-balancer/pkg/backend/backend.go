package backend

type SimpleBackEnd struct {
	URL  string `yaml:"url"`
	Name string `yaml:"name"`
}

func (b *SimpleBackEnd) GetURL() string {
	return b.URL
}

func (b *SimpleBackEnd) GetName() string {
	return b.Name
}

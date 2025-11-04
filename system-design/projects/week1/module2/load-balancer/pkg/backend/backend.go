package backend

type SimpleBackEnd struct {
	URL  string `yaml:"url" json:"url"`
	Name string `yaml:"name" json:"name"`
}

func (b *SimpleBackEnd) GetURL() string {
	return b.URL
}

func (b *SimpleBackEnd) GetName() string {
	return b.Name
}

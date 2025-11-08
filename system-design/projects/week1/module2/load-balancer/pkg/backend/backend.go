package backend

// ------------------- SimpleBackEnd ------------------- //
type SimpleBackEnd struct {
	URL    string `yaml:"url" json:"url"`
	Name   string `yaml:"name" json:"name"`
	Weight int    `yaml:"weight" json:"weight"`
}

func (b *SimpleBackEnd) GetURL() string {
	return b.URL
}

func (b *SimpleBackEnd) GetName() string {
	return b.Name
}

func (b *SimpleBackEnd) GetWeight() int {
	return b.Weight
}

// ------------------- SwrrBackEnd ------------------- //
type SwrrBackEnd struct {
	SimpleBackEnd
	CurWeight int
}

// ------------------- HashBackend ------------------- //
type HashBackend struct {
	SimpleBackEnd
	hashvalue string // TODO
}

package backend

import (
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

func NewSimpleBackEnd(URL, Name string, Weight int) *SimpleBackEnd {
	return &SimpleBackEnd{
		URL:    URL,
		Name:   Name,
		Weight: Weight,
	}
}

func NewSwrrBackEnd(URL, Name string, Weight int) *SwrrBackEnd {
	return &SwrrBackEnd{
		SimpleBackEnd: SimpleBackEnd{
			URL:    URL,
			Name:   Name,
			Weight: Weight,
		},
		CurWeight: 0,
	}
}

func NewSimpleBackEndFromYaml(path string) []*SimpleBackEnd {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	var backends []*SimpleBackEnd
	if err := yaml.Unmarshal(data, &backends); err != nil {
		log.Fatalf("failed to parse yaml: %v", err)
	}

	return backends
}

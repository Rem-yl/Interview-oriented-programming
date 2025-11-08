package backend

import (
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

func NewSimpleBackEnd(url, name string, weight int) *SimpleBackEnd {
	return &SimpleBackEnd{
		URL:    url,
		Name:   name,
		Weight: weight,
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

func NewSwrrBackEnd(url, name string, weight int) *SwrrBackEnd {
	simpleBackend := NewSimpleBackEnd(url, name, weight)

	return &SwrrBackEnd{
		SimpleBackEnd: *simpleBackend,
		CurWeight:     0,
	}
}

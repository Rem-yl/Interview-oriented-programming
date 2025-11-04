package backend

import (
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

func NewSimpleBackEnd(URL, Name string) *SimpleBackEnd {
	return &SimpleBackEnd{
		URL:  URL,
		Name: Name,
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

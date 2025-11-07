package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

func LoadSimpleBackendConfig(path string) (*SimpleBackEndConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg SimpleBackEndConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

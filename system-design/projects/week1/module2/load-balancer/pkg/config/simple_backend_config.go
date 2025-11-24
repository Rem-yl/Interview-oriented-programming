package config

import (
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

type SimpleBackEndConfig struct {
	LoadBalancer LoadBalancer `yaml:"load_balancer"`
	Servers      []*Server    `yaml:"server"`
}

type LoadBalancer struct {
	URL  string `yaml:"url"`
	Mode string `yaml:"mode"`
}

type Server struct {
	Name   string `yaml:"name"`
	URL    string `yaml:"url"`
	Weight int    `yaml:"weight"`
}

func LoadSimpleBackendConfig(path string) (*SimpleBackEndConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("failed to read config file: %v", err)
		return nil, err
	}

	var config SimpleBackEndConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Printf("failed to parse yaml: %v", err)
		return nil, err
	}

	return &config, nil
}

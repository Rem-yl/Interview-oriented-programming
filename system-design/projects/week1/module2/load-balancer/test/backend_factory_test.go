package test

import (
	"testing"

	"github.com/rem/load-balancer/pkg/backend"
)

var simpleBackEndConfigPath = "configs/simple_backend_config.yaml"

func TestNewSimpleBackEndFromYaml(t *testing.T) {
	backends := backend.NewSimpleBackEndFromYaml(simpleBackEndConfigPath)
	if len(backends) != 2 {
		t.Errorf("len(backends) = %d != 2 \n", len(backends))
	}

	backend1 := backends[0]
	if backend1.Name != "backend1" || backend1.URL != "http://127.0.0.1:8080" {
		t.Errorf("Want backend1 name: %s, url: %s. Get name: %s, url: %s \n", "backend1", "http://127.0.0.1:8080", backend1.Name, backend1.URL)
	}

	backend2 := backends[1]
	if backend2.Name != "backend2" || backend2.URL != "http://127.0.0.1:8081" {
		t.Errorf("Want backend2 name: %s, url: %s. Get name: %s, url: %s \n", "backend1", "http://127.0.0.1:8081", backend2.Name, backend2.URL)
	}
}

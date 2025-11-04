package config

type SimpleBackEndConfig struct {
	LoadBalancer LoadBalancer `yaml:"load_balancer"`
	Servers      []*Server    `yaml:"server"`
}

type LoadBalancer struct {
	URL string `yaml:"url"`
}

type Server struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

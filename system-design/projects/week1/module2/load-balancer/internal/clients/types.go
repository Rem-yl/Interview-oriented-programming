package clients

import "github.com/rem/load-balancer/pkg/backend"

type SimpleBackEndResponse struct {
	Data backend.SimpleBackEnd `json:"data"`
}

type HelloServerResponse struct {
	Data string `json:"data"`
}

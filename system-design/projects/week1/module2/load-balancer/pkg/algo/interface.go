package algo

import (
	"github.com/rem/load-balancer/pkg/backend"
)

type LoadBalanceAlgo interface {
	GetBackEnd() (*backend.BackEnd, error)
}

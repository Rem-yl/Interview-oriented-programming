package tester

import "context"

type Tester interface {
	Run(ctx context.Context) ([]RequestResult, error)
}

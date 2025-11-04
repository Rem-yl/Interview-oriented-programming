package errs

import "errors"

var (
	ErrNoServerList = errors.New("No server list!")
	ErrNoBackEnd    = errors.New("No backend founded!")
)

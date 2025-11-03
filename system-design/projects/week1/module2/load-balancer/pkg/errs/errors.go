package errs

import "errors"

var (
	ErrNoServerList     = errors.New("No server list!")
	ErrNoBackEnd        = errors.New("No backend founded!")
	ErrCreateGetRequest = errors.New("Create GET request failed!")
	ErrSendRequest      = errors.New("Send GET request failed")
)

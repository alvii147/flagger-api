package errutils

import "errors"

// General shared errors.
var (
	ErrInvalidToken        = errors.New("invalid token")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrAPIKeyAlreadyExists = errors.New("api key already exists")
	ErrAPIKeyNotFound      = errors.New("api key not found")
	ErrFlagAlreadyExists   = errors.New("flag already exists")
	ErrFlagNotFound        = errors.New("flag not found")
)

package errutils

import "errors"

// Database errors.
var (
	ErrDatabaseUniqueViolation = errors.New("unique key constraint violation.")
	ErrDatabaseNoRowsAffected  = errors.New("no rows affected")
	ErrDatabaseNoRowsReturned  = errors.New("no rows returned")
)

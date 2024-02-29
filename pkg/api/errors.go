package api

// Error codes.
const (
	ErrCodeInvalidRequest      = "invalid_request"
	ErrCodeResourceExists      = "resource_exists"
	ErrCodeResourceNotFound    = "resource_not_found"
	ErrCodeInvalidCredentials  = "invalid_credentials"
	ErrCodeMissingCredentials  = "missing_credentials"
	ErrCodeInternalServerError = "internal_server_error"
)

// Error details.
const (
	ErrDetailInvalidRequestData     = "Invalid or malformed request data."
	ErrDetailUserExists             = "User already exists"
	ErrDetailUserNotFound           = "User not found"
	ErrDetailInvalidEmailOrPassword = "Incorrect email or password."
	ErrDetailInvalidToken           = "Provided token is invalid"
	ErrDetailMissingCredentials     = "No credentials were provided"
	ErrDetailInternalServerError    = "Internal server error occurred."
	ErrDetailAPIKeyNotFound         = "API key not found"
	ErrDetailFlagNotFound           = "Flag not found"
)

// ErrorResponse represents the general error response body.
type ErrorResponse struct {
	Code               string              `json:"code"`
	Detail             string              `json:"detail"`
	ValidationFailures map[string][]string `json:"failures,omitempty"`
}

package api

import (
	"time"

	"github.com/alvii147/flagger-api/pkg/utils"
	"github.com/alvii147/flagger-api/pkg/validate"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5/pgtype"
)

// AuthJWTClaims represents claims in JWTs used for User authentication.
type AuthJWTClaims struct {
	Subject   string              `json:"sub"`
	TokenType string              `json:"token_type"`
	IssuedAt  utils.JSONTimeStamp `json:"iat"`
	ExpiresAt utils.JSONTimeStamp `json:"exp"`
	JWTID     string              `json:"jti"`
	jwt.StandardClaims
}

// ActivationJWTClaims represents claims in JWTs used for User activation.
type ActivationJWTClaims struct {
	Subject   string              `json:"sub"`
	TokenType string              `json:"token_type"`
	IssuedAt  utils.JSONTimeStamp `json:"iat"`
	ExpiresAt utils.JSONTimeStamp `json:"exp"`
	JWTID     string              `json:"jti"`
	jwt.StandardClaims
}

// CreateUserRequest represents the request body for create User requests.
type CreateUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (r *CreateUserRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringEmail("email", r.Email)
	v.ValidatorStringNotBlank("email", r.Email)
	v.ValidatorStringNotBlank("password", r.Password)
	v.ValidatorStringNotBlank("first_name", r.FirstName)
	v.ValidatorStringNotBlank("last_name", r.LastName)

	return v.Passed(), v.Failures()
}

// CreateUserResponse represents the response body for create User requests.
type CreateUserResponse struct {
	UUID      string    `json:"uuid"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

// ActivateUserRequest represents the request body for activate User requests.
type ActivateUserRequest struct {
	Token string `json:"token"`
}

func (r *ActivateUserRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidatorStringNotBlank("token", r.Token)

	return v.Passed(), v.Failures()
}

// GetUserMeResponse represents the request body for get current User requests.
type GetUserMeResponse struct {
	UUID      string    `json:"uuid"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTokenRequest represents the request body for create token requests.
type CreateTokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *CreateTokenRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringEmail("email", r.Email)
	v.ValidatorStringNotBlank("email", r.Email)
	v.ValidatorStringNotBlank("password", r.Password)

	return v.Passed(), v.Failures()
}

// CreateTokenResponse represents the response body for create token requests.
type CreateTokenResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

// RefreshTokenRequest represents the request body for refresh token requests.
type RefreshTokenRequest struct {
	Refresh string `json:"refresh"`
}

func (r *RefreshTokenRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidatorStringNotBlank("refresh", r.Refresh)

	return v.Passed(), v.Failures()
}

// RefreshTokenResponse represents the response body for refresh token requests.
type RefreshTokenResponse struct {
	Access string `json:"access"`
}

// CreateAPIKeyRequest represents the request body for API Key creation requests.
type CreateAPIKeyRequest struct {
	Name      string           `json:"name"`
	ExpiresAt pgtype.Timestamp `json:"expires_at"`
}

func (r *CreateAPIKeyRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidatorStringNotBlank("name", r.Name)

	return v.Passed(), v.Failures()
}

// CreateAPIKeyResponse represents the response body for API Key creation requests.
type CreateAPIKeyResponse struct {
	ID        int              `json:"id"`
	RawKey    string           `json:"raw_key"`
	UserUUID  string           `json:"user_uuid"`
	Name      string           `json:"name"`
	CreatedAt time.Time        `json:"created_at"`
	ExpiresAt pgtype.Timestamp `json:"expires_at"`
}

// GetAPIKeyResponse represents the response body for a single API Key in API Key retrieval requests.
type GetAPIKeyResponse struct {
	ID        int              `json:"id"`
	UserUUID  string           `json:"user_uuid"`
	Prefix    string           `json:"prefix"`
	Name      string           `json:"name"`
	CreatedAt time.Time        `json:"created_at"`
	ExpiresAt pgtype.Timestamp `json:"expires_at"`
}

// ListAPIKeysResponse represents the response body for API Key retrieval requests.
type ListAPIKeysResponse struct {
	Keys []*GetAPIKeyResponse `json:"keys"`
}

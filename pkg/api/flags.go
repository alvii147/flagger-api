package api

import (
	"time"

	"github.com/alvii147/flagger-api/pkg/validate"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateFlagRequest represents the request body for Flag creation requests.
type CreateFlagRequest struct {
	Name string `json:"name"`
}

// Validate validates fields in CreateFlagRequest.
func (r *CreateFlagRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringNotBlank("name", r.Name)

	return v.Passed(), v.Failures()
}

// CreateFlagResponse represents the response body for Flag creation requests.
type CreateFlagResponse struct {
	ID        int       `json:"id"`
	UserUUID  string    `json:"user_uuid"`
	Name      string    `json:"name"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetFlagByIDResponse represents the response body for a single Flag in Flag retrieval requests.
type GetFlagByIDResponse struct {
	ID        int       `json:"id"`
	UserUUID  string    `json:"user_uuid"`
	Name      string    `json:"name"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetFlagByNameResponse represents the response body for a single Flag in Flag retrieval requests.
type GetFlagByNameResponse struct {
	ID        *int             `json:"id"`
	UserUUID  *string          `json:"user_uuid"`
	Name      string           `json:"name"`
	IsEnabled bool             `json:"is_enabled"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at"`
	Valid     bool             `json:"valid"`
}

// ListFlagsResponse represents the response body for Flag retrieval requests.
type ListFlagsResponse struct {
	Flags []*GetFlagByIDResponse `json:"flags"`
}

// EnableFlagResponse represents the response body for a single Flag in Flag enabling requests.
type EnableFlagResponse struct {
	ID        int       `json:"id"`
	UserUUID  string    `json:"user_uuid"`
	Name      string    `json:"name"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DisableFlagResponse represents the response body for a single Flag in Flag disabling requests.
type DisableFlagResponse struct {
	ID        int       `json:"id"`
	UserUUID  string    `json:"user_uuid"`
	Name      string    `json:"name"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateFlagRequest represents the request body for Flag update requests.
type UpdateFlagRequest struct {
	Name      string `json:"name"`
	IsEnabled bool   `json:"is_enabled"`
}

// Validate validates fields in UpdateFlagRequest.
func (r *UpdateFlagRequest) Validate() (bool, map[string][]string) {
	v := validate.NewValidator()
	v.ValidateStringNotBlank("name", r.Name)
	v.ValidateStringSlug("name", r.Name)

	return v.Passed(), v.Failures()
}

// UpdateFlagResponse represents the response body for a single Flag in Flag update requests.
type UpdateFlagResponse struct {
	ID        int       `json:"id"`
	UserUUID  string    `json:"user_uuid"`
	Name      string    `json:"name"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

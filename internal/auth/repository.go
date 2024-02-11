package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/alvii147/flagger-api/pkg/errutils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is used to access and update auth data.
type Repository interface {
	CreateUser(dbConn *pgxpool.Conn, user *User) (*User, error)
	ActivateUserByUUID(dbConn *pgxpool.Conn, userUUID string) error
	GetUserByEmail(dbConn *pgxpool.Conn, email string) (*User, error)
	GetUserByUUID(dbConn *pgxpool.Conn, userUUID string) (*User, error)
	UpdateUser(dbConn *pgxpool.Conn, userUUID string, firstName *string, lastName *string) (*User, error)
	CreateAPIKey(dbConn *pgxpool.Conn, apiKey *APIKey) (*APIKey, error)
	ListAPIKeysByUserUUID(dbConn *pgxpool.Conn, userUUID string) ([]*APIKey, error)
	ListActiveAPIKeysByPrefix(dbConn *pgxpool.Conn, prefix string) ([]*APIKey, error)
	DeleteAPIKey(dbConn *pgxpool.Conn, apiKeyID int, userUUID string) error
}

// repository implements Repository.
type repository struct{}

// NewRepository returns a new Repository.
func NewRepository() Repository {
	return &repository{}
}

// CreateUser creates User from email, password, first and last names, and active and superuser states.
func (repo *repository) CreateUser(dbConn *pgxpool.Conn, user *User) (*User, error) {
	createdUser := &User{}

	q := `
INSERT INTO "User" (
	uuid,
	email,
	password,
	first_name,
	last_name,
	is_active,
	is_superuser
)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5,
	$6,
	$7
)
RETURNING
	uuid,
	email,
	password,
	first_name,
	last_name,
	is_active,
	is_superuser,
	created_at;
	`

	err := dbConn.QueryRow(
		context.Background(),
		q,
		user.UUID,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.IsActive,
		user.IsSuperUser,
	).Scan(
		&createdUser.UUID,
		&createdUser.Email,
		&createdUser.Password,
		&createdUser.FirstName,
		&createdUser.LastName,
		&createdUser.IsActive,
		&createdUser.IsSuperUser,
		&createdUser.CreatedAt,
	)

	var pgErr *pgconn.PgError
	ok := errors.As(err, &pgErr)

	if ok && pgErr != nil && pgErr.Code == "23505" {
		return nil, fmt.Errorf("CreateUser failed to dbConn.Scan, %w: %w", errutils.ErrDatabaseUniqueViolation, pgErr)
	}

	if err != nil {
		return nil, fmt.Errorf("CreateUser failed to dbConn.Scan: %w", err)
	}

	return createdUser, nil
}

// ActivateUserByUUID activates User.
// If no User is affected, error is returned.
func (repo *repository) ActivateUserByUUID(dbConn *pgxpool.Conn, userUUID string) error {
	q := `
UPDATE
	"User"
SET
	is_active = TRUE
WHERE
	uuid = $1
	AND is_active = FALSE;
	`

	ct, err := dbConn.Exec(context.Background(), q, userUUID)
	if err != nil {
		return fmt.Errorf("ActivateUserByUUID failed to dbConn.Exec: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("ActivateUserByUUID failed: %w", errutils.ErrDatabaseNoRowsAffected)
	}

	return nil
}

// GetUserByEmail fetches User by email.
// If no User found, error is returned.
func (repo *repository) GetUserByEmail(dbConn *pgxpool.Conn, email string) (*User, error) {
	user := &User{}

	q := `
SELECT
	uuid,
	email,
	password,
	first_name,
	last_name,
	is_active,
	is_superuser,
	created_at
FROM
	"User"
WHERE
	email = $1
	AND is_active = TRUE;
	`

	err := dbConn.QueryRow(context.Background(), q, email).Scan(
		&user.UUID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.IsSuperUser,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("GetUserByEmail failed: %w", errutils.ErrDatabaseNoRowsReturned)
	}

	if err != nil {
		return nil, fmt.Errorf("GetUserByEmail failed to dbConn.Scan: %w", err)
	}

	return user, nil
}

// GetUserByUUID fetches User by UUID.
// If no User found, error is returned.
func (repo *repository) GetUserByUUID(dbConn *pgxpool.Conn, userUUID string) (*User, error) {
	user := &User{}

	q := `
SELECT
	uuid,
	email,
	password,
	first_name,
	last_name,
	is_active,
	is_superuser,
	created_at
FROM
	"User"
WHERE
	uuid = $1
	AND is_active = TRUE;
	`

	err := dbConn.QueryRow(context.Background(), q, userUUID).Scan(
		&user.UUID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.IsSuperUser,
		&user.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("GetUserByUUID failed: %w", errutils.ErrDatabaseNoRowsReturned)
	}

	if err != nil {
		return nil, fmt.Errorf("GetUserByUUID failed to dbConn.Scan: %w", err)
	}

	return user, nil
}

// UpdateUser updates User first and last names
func (repo *repository) UpdateUser(dbConn *pgxpool.Conn, userUUID string, firstName *string, lastName *string) (*User, error) {
	if firstName == nil && lastName == nil {
		return nil, fmt.Errorf("UpdateUser failed, all attributes are nil: %w", errutils.ErrDatabaseNoRowsAffected)
	}

	updatedUser := &User{}

	q := `
UPDATE
	"User"
SET
	first_name = COALESCE($1, first_name),
	last_name = COALESCE($2, last_name)
WHERE
	uuid = $3
	AND is_active = TRUE
RETURNING
	uuid,
	email,
	password,
	first_name,
	last_name,
	is_active,
	is_superuser,
	created_at;
	`

	err := dbConn.QueryRow(
		context.Background(),
		q,
		firstName,
		lastName,
		userUUID,
	).Scan(
		&updatedUser.UUID,
		&updatedUser.Email,
		&updatedUser.Password,
		&updatedUser.FirstName,
		&updatedUser.LastName,
		&updatedUser.IsActive,
		&updatedUser.IsSuperUser,
		&updatedUser.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("UpdateUser failed: %w", errutils.ErrDatabaseNoRowsAffected)
	}

	if err != nil {
		return nil, fmt.Errorf("UpdateUser failed to dbConn.Scan: %w", err)
	}

	return updatedUser, nil
}

// CreateAPIKey creates API key from user UUID, prefix, hashed key, name, and expiry date.
func (repo *repository) CreateAPIKey(dbConn *pgxpool.Conn, apiKey *APIKey) (*APIKey, error) {
	createdAPIKey := &APIKey{}

	q := `
INSERT INTO APIKey (
	user_uuid,
	prefix,
	hashed_key,
	name,
	expires_at
)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5
)
RETURNING
	id,
	user_uuid,
	prefix,
	hashed_key,
	name,
	created_at,
	expires_at;
	`
	err := dbConn.QueryRow(
		context.Background(),
		q,
		apiKey.UserUUID,
		apiKey.Prefix,
		apiKey.HashedKey,
		apiKey.Name,
		apiKey.ExpiresAt,
	).Scan(
		&createdAPIKey.ID,
		&createdAPIKey.UserUUID,
		&createdAPIKey.Prefix,
		&createdAPIKey.HashedKey,
		&createdAPIKey.Name,
		&createdAPIKey.CreatedAt,
		&createdAPIKey.ExpiresAt,
	)

	var pgErr *pgconn.PgError
	ok := errors.As(err, &pgErr)

	if ok && pgErr != nil && pgErr.Code == "23505" {
		return nil, fmt.Errorf("CreateUser failed to dbConn.Scan, %w: %w", errutils.ErrDatabaseUniqueViolation, pgErr)
	}

	if err != nil {
		return nil, fmt.Errorf("CreateAPIKey failed to dbConn.Scan: %w", err)
	}

	return createdAPIKey, nil
}

// ListAPIKeysByUserUUID fetches API keys under a given User UUID.
func (repo *repository) ListAPIKeysByUserUUID(dbConn *pgxpool.Conn, userUUID string) ([]*APIKey, error) {
	apiKeys := make([]*APIKey, 0)

	q := `
SELECT
	k.id,
	k.user_uuid,
	k.prefix,
	k.hashed_key,
	k.name,
	k.created_at,
	k.expires_at
FROM
	APIKey k
INNER JOIN
	"User" u
ON
	k.user_uuid = u.uuid
WHERE
	k.user_uuid = $1
	AND u.is_active = TRUE;
	`

	rows, err := dbConn.Query(context.Background(), q, userUUID)
	if err != nil {
		return nil, fmt.Errorf("ListAPIKeysByUserUUID failed to dbConn.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		apiKey := &APIKey{}
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserUUID,
			&apiKey.Prefix,
			&apiKey.HashedKey,
			&apiKey.Name,
			&apiKey.CreatedAt,
			&apiKey.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ListAPIKeysByUserUUID failed to rows.Scan: %w", err)
		}

		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// ListActiveAPIKeysByPrefix fetches API keys with a given prefix.
func (repo *repository) ListActiveAPIKeysByPrefix(dbConn *pgxpool.Conn, prefix string) ([]*APIKey, error) {
	apiKeys := make([]*APIKey, 0)

	q := `
SELECT
	k.id,
	k.user_uuid,
	k.prefix,
	k.hashed_key,
	k.name,
	k.created_at,
	k.expires_at
FROM
	APIKey k
INNER JOIN
	"User" u
ON
	k.user_uuid = u.uuid
WHERE
	k.prefix = $1
	AND (k.expires_at IS NULL OR k.expires_at > CURRENT_TIMESTAMP)
	AND u.is_active = TRUE;
	`

	rows, err := dbConn.Query(context.Background(), q, prefix)
	if err != nil {
		return nil, fmt.Errorf("ListActiveAPIKeysByPrefix failed to dbConn.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		apiKey := &APIKey{}
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserUUID,
			&apiKey.Prefix,
			&apiKey.HashedKey,
			&apiKey.Name,
			&apiKey.CreatedAt,
			&apiKey.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ListActiveAPIKeysByPrefix failed to rows.Scan: %w", err)
		}

		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// DeleteAPIKey deletes API key by ID.
// If no API key found, error is returned.
func (repo *repository) DeleteAPIKey(dbConn *pgxpool.Conn, apiKeyID int, userUUID string) error {
	q := `
DELETE FROM
	APIKey k
USING
	"User" u
WHERE
	k.id = $1
	AND k.user_uuid = $2
	AND k.user_uuid = u.uuid
	AND u.is_active = TRUE;
	`

	ct, err := dbConn.Exec(context.Background(), q, apiKeyID, userUUID)

	if err != nil {
		return fmt.Errorf("DeleteAPIKey failed to dbConn.Exec: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("DeleteAPIKey failed: %w", errutils.ErrDatabaseNoRowsAffected)
	}

	return nil
}

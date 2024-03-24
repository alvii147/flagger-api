package flags

import (
	"context"
	"errors"
	"fmt"

	"github.com/alvii147/flagger-api/pkg/errutils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is used to access and update Flags data.
type Repository interface {
	CreateFlag(dbConn *pgxpool.Conn, flag *Flag) (*Flag, error)
	GetFlagByID(dbConn *pgxpool.Conn, flagID int, userUUID string) (*Flag, error)
	GetFlagByName(dbConn *pgxpool.Conn, flagName string, userUUID string) (*Flag, error)
	ListFlagsByUserUUID(dbConn *pgxpool.Conn, userUUID string) ([]*Flag, error)
	UpdateFlag(dbConn *pgxpool.Conn, flag *Flag) (*Flag, error)
}

// repository implements Repository.
type repository struct{}

// NewRepository returns a new repository.
func NewRepository() *repository {
	return &repository{}
}

// CreateFlag creates new Flag given User UUID and Flag name.
func (repo *repository) CreateFlag(dbConn *pgxpool.Conn, flag *Flag) (*Flag, error) {
	createdFlag := &Flag{}

	q := `
INSERT INTO Flag (
	user_uuid,
	name
)
VALUES (
	$1,
	$2
)
RETURNING
	id,
	user_uuid,
	name,
	is_enabled,
	created_at,
	updated_at;
	`

	err := dbConn.QueryRow(
		context.Background(),
		q,
		flag.UserUUID,
		flag.Name,
	).Scan(
		&createdFlag.ID,
		&createdFlag.UserUUID,
		&createdFlag.Name,
		&createdFlag.IsEnabled,
		&createdFlag.CreatedAt,
		&createdFlag.UpdatedAt,
	)

	var pgErr *pgconn.PgError
	ok := errors.As(err, &pgErr)

	if ok && pgErr != nil && pgErr.Code == "23505" {
		return nil, fmt.Errorf("CreateFlag failed to dbConn.Scan, %w: %w", errutils.ErrDatabaseUniqueViolation, pgErr)
	}

	if err != nil {
		return nil, fmt.Errorf("CreateFlag failed to dbConn.Scan: %w", err)
	}

	return createdFlag, nil
}

// GetFlagByID fetches Flag by ID.
// If no Flag found, error is returned.
func (repo *repository) GetFlagByID(dbConn *pgxpool.Conn, flagID int, userUUID string) (*Flag, error) {
	flag := &Flag{}

	q := `
SELECT
	f.id,
	f.user_uuid,
	f.name,
	f.is_enabled,
	f.created_at,
	f.updated_at
FROM
	Flag f
INNER JOIN
	"User" u
ON
	f.user_uuid = u.uuid
WHERE
	f.id = $1
	AND f.user_uuid = $2
	AND u.is_active = TRUE;
	`

	err := dbConn.QueryRow(context.Background(), q, flagID, userUUID).Scan(
		&flag.ID,
		&flag.UserUUID,
		&flag.Name,
		&flag.IsEnabled,
		&flag.CreatedAt,
		&flag.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("GetFlagByID failed: %w", errutils.ErrDatabaseNoRowsReturned)
	}

	if err != nil {
		return nil, fmt.Errorf("GetFlagByID failed to dbConn.Scan: %w", err)
	}

	return flag, nil
}

// GetFlagByName fetches Flag by name.
// If no Flag found, error is returned.
func (repo *repository) GetFlagByName(dbConn *pgxpool.Conn, name string, userUUID string) (*Flag, error) {
	flag := &Flag{}

	q := `
SELECT
	f.id,
	f.user_uuid,
	f.name,
	f.is_enabled,
	f.created_at,
	f.updated_at
FROM
	Flag f
INNER JOIN
	"User" u
ON
	f.user_uuid = u.uuid
WHERE
	f.name = $1
	AND f.user_uuid = $2
	AND u.is_active = TRUE;
	`

	err := dbConn.QueryRow(context.Background(), q, name, userUUID).Scan(
		&flag.ID,
		&flag.UserUUID,
		&flag.Name,
		&flag.IsEnabled,
		&flag.CreatedAt,
		&flag.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("GetFlagByName failed: %w", errutils.ErrDatabaseNoRowsReturned)
	}

	if err != nil {
		return nil, fmt.Errorf("GetFlagByName failed to dbConn.Scan: %w", err)
	}

	return flag, nil
}

// ListFlagsByUserUUID fetches Flags under a given User UUID.
func (repo *repository) ListFlagsByUserUUID(dbConn *pgxpool.Conn, userUUID string) ([]*Flag, error) {
	flags := make([]*Flag, 0)

	q := `
SELECT
	f.id,
	f.user_uuid,
	f.name,
	f.is_enabled,
	f.created_at,
	f.updated_at
FROM
	Flag f
INNER JOIN
	"User" u
ON
	f.user_uuid = u.uuid
WHERE
	f.user_uuid = $1
	AND u.is_active = TRUE;
	`

	rows, err := dbConn.Query(context.Background(), q, userUUID)
	if err != nil {
		return nil, fmt.Errorf("ListFlagsByUserUUID failed to dbConn.Query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		flag := &Flag{}
		err := rows.Scan(
			&flag.ID,
			&flag.UserUUID,
			&flag.Name,
			&flag.IsEnabled,
			&flag.CreatedAt,
			&flag.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ListFlagsByUserUUID failed to rows.Scan: %w", err)
		}

		flags = append(flags, flag)
	}

	return flags, nil
}

// UpdateFlag updates a Flag's enabled state.
// If no Flag is affected, error is returned.
func (repo *repository) UpdateFlag(dbConn *pgxpool.Conn, flag *Flag) (*Flag, error) {
	updatedFlag := &Flag{}

	q := `
UPDATE
	Flag f
SET
	is_enabled = $1
FROM
	"User" u
WHERE
	f.id = $2
	AND f.user_uuid = $3
	AND f.user_uuid = u.uuid
	AND u.is_active = TRUE
RETURNING
	f.id,
	f.user_uuid,
	f.name,
	f.is_enabled,
	f.created_at,
	f.updated_at;
	`

	err := dbConn.QueryRow(
		context.Background(),
		q,
		flag.IsEnabled,
		flag.ID,
		flag.UserUUID,
	).Scan(
		&updatedFlag.ID,
		&updatedFlag.UserUUID,
		&updatedFlag.Name,
		&updatedFlag.IsEnabled,
		&updatedFlag.CreatedAt,
		&updatedFlag.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("UpdateFlag failed: %w", errutils.ErrDatabaseNoRowsAffected)
	}

	if err != nil {
		return nil, fmt.Errorf("UpdateFlag failed to dbConn.Scan: %w", err)
	}

	return updatedFlag, nil
}

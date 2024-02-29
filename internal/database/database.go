package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateConnString constructs PostgreSQL connection string from Config.
func CreateConnString(
	hostname string,
	port int,
	username string,
	password string,
	databaseName string,
) string {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		username,
		password,
		hostname,
		port,
		databaseName,
	)

	return connString
}

// CreatePool creates and returns a new database connection pool.
func CreatePool(
	hostname string,
	port int,
	username string,
	password string,
	databaseName string,
) (*pgxpool.Pool, error) {
	connString := CreateConnString(
		hostname,
		port,
		username,
		password,
		databaseName,
	)

	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("CreatePool failed to pgxpool.New to %s: %w", connString, err)
	}

	return dbPool, nil
}

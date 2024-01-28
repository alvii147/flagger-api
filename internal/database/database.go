package database

import (
	"context"
	"fmt"

	"github.com/alvii147/flagger-api/internal/env"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BuildConnString constructs PostgreSQL connection string from Config.
func BuildConnString() string {
	config := env.GetConfig()
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		config.PostgresUsername,
		config.PostgresPassword,
		config.PostgresHostname,
		config.PostgresPort,
		config.PostgresDatabaseName,
	)

	return connString
}

// CreatePool creates and returns a new database connection pool.
func CreatePool() (*pgxpool.Pool, error) {
	connString := BuildConnString()

	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("CreatePool failed to pgxpool.New to %s: %w", connString, err)
	}

	return dbPool, nil
}

package testkitinternal

import (
	"context"
	"testing"

	"github.com/alvii147/flagger-api/internal/database"
	"github.com/alvii147/flagger-api/internal/env"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// RequireCreateDatabasePool creates and returns a new database connection pool.
// It also asserts no error is returned and declares clean up function to close the pool.
func RequireCreateDatabasePool(t *testing.T) *pgxpool.Pool {
	config := env.NewConfig()
	dbPool, err := database.CreatePool(
		config.PostgresHostname,
		config.PostgresPort,
		config.PostgresUsername,
		config.PostgresPassword,
		config.PostgresDatabaseName,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		dbPool.Close()
	})

	return dbPool
}

// RequireCreateDatabaseConn creates and returns a new database connection from a given connection pool.
// It also asserts no error is returned and declares clean up function to close the connection.
func RequireCreateDatabaseConn(t *testing.T, dbPool *pgxpool.Pool, ctx context.Context) *pgxpool.Conn {
	dbConn, err := dbPool.Acquire(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		dbConn.Release()
	})

	return dbConn
}

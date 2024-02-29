package database_test

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/alvii147/flagger-api/internal/database"
	"github.com/alvii147/flagger-api/internal/env"
	"github.com/stretchr/testify/require"
)

func TestCreateConnString(t *testing.T) {
	t.Parallel()

	hostname := "localhost"
	port := 5432
	username := "user"
	password := "pass"
	databaseName := "db"
	connString := database.CreateConnString(
		hostname,
		port,
		username,
		password,
		databaseName,
	)

	re := regexp.MustCompile(`^(\S+):\/\/(\S+):(\S+)@(\S+):(\d+)/(\S+)$`)
	match := re.FindStringSubmatch(connString)

	require.Len(t, match, 7)

	require.Equal(t, "postgres", match[1])
	require.Equal(t, username, match[2])
	require.Equal(t, password, match[3])
	require.Equal(t, hostname, match[4])
	require.Equal(t, strconv.Itoa(port), match[5])
	require.Equal(t, databaseName, match[6])
}

func TestCreatePoolSuccess(t *testing.T) {
	t.Parallel()

	config := env.NewConfig()
	dbPool, err := database.CreatePool(
		config.PostgresHostname,
		config.PostgresPort,
		config.PostgresUsername,
		config.PostgresPassword,
		config.PostgresDatabaseName,
	)
	require.NoError(t, err)
	require.NotNil(t, dbPool)
}

func TestCreatePoolBadConnString(t *testing.T) {
	t.Parallel()

	_, err := database.CreatePool("", 0, "", "", "")
	require.Error(t, err)
}

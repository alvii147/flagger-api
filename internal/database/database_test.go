package database_test

import (
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/alvii147/flagger-api/internal/database"
	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	defer testkitinternal.TeardownTests()
	testkitinternal.SetupTests()
	code := m.Run()
	os.Exit(code)
}

func TestBuildConnString(t *testing.T) {
	t.Parallel()

	config := env.GetConfig()
	connString := database.BuildConnString()

	re := regexp.MustCompile(`^(\S+):\/\/(\S+):(\S+)@(\S+):(\d+)/(\S+)$`)
	match := re.FindStringSubmatch(connString)

	require.Len(t, match, 7)

	protocol, username, password, hostname, port, dbname := match[1], match[2], match[3], match[4], match[5], match[6]

	require.Equal(t, "postgres", protocol)
	require.Equal(t, config.PostgresUsername, username)
	require.Equal(t, config.PostgresPassword, password)
	require.Equal(t, config.PostgresHostname, hostname)
	require.Equal(t, strconv.Itoa(config.PostgresPort), port)
	require.Equal(t, config.PostgresDatabaseName, dbname)
}

func TestCreatePoolSuccess(t *testing.T) {
	t.Parallel()

	dbPool, err := database.CreatePool()
	require.NoError(t, err)
	require.NotNil(t, dbPool)
}

func TestCreatePoolBadConnString(t *testing.T) {
	defaultConfig := env.GetConfig()
	defer env.SetConfig(defaultConfig)

	config := env.NewConfig()
	config.PostgresUsername = ""
	config.PostgresPassword = ""
	config.PostgresHostname = ""
	config.PostgresPort = 0
	config.PostgresDatabaseName = ""
	env.SetConfig(config)

	_, err := database.CreatePool()
	require.Error(t, err)
}

package env_test

import (
	"testing"

	"github.com/alvii147/flagger-api/internal/env"
	"github.com/stretchr/testify/require"
)

func TestConfigString(t *testing.T) {
	t.Setenv("FLAGGERAPI_HOSTNAME", "127.0.0.1")

	config, err := env.NewConfig()
	require.NoError(t, err)

	require.Equal(t, "127.0.0.1", config.Hostname)
}

func TestConfigInt(t *testing.T) {
	t.Setenv("FLAGGERAPI_PORT", "3000")

	config, err := env.NewConfig()
	require.NoError(t, err)

	require.Equal(t, 3000, config.Port)
}

func TestConfigInvalidIntError(t *testing.T) {
	t.Setenv("FLAGGERAPI_PORT", "B33F")

	_, err := env.NewConfig()
	require.Error(t, err)
}

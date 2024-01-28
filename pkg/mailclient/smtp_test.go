package mailclient_test

import (
	"testing"

	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestNewSMTPMailClient(t *testing.T) {
	t.Parallel()

	hostname := testkit.MustGenerateRandomString(12, true, true, true)
	port := 587
	username := testkit.GenerateFakeEmail()
	password := testkit.GenerateFakePassword()
	templatesDir := "."

	_, err := mailclient.NewSMTPMailClient(hostname, port, username, password, templatesDir)
	require.NoError(t, err)
}

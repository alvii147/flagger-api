package mailclient_test

import (
	"testing"

	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/alvii147/flagger-api/pkg/testkit"
)

func TestNewSMTPMailClient(t *testing.T) {
	t.Parallel()

	hostname := testkit.MustGenerateRandomString(12, true, true, true)
	port := 587
	username := testkit.GenerateFakeEmail()
	password := testkit.GenerateFakePassword()

	mailclient.NewSMTPClient(hostname, port, username, password)
}

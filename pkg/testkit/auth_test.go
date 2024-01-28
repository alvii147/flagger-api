package testkit_test

import (
	"net/mail"
	"testing"

	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestMustGenerateRandomStringSuccess(t *testing.T) {
	t.Parallel()

	randomString := testkit.MustGenerateRandomString(8, true, true, true)
	require.Len(t, randomString, 8)
}

func TestMustGenerateRandomStringPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r)
	}()

	testkit.MustGenerateRandomString(8, false, false, false)
}

func TestGenerateFakeEmail(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		email := testkit.GenerateFakeEmail()
		_, err := mail.ParseAddress(email)
		require.NoError(t, err)
	}
}

func TestGenerateFakePassword(t *testing.T) {
	t.Parallel()

	for i := 0; i < 10; i++ {
		password := testkit.GenerateFakePassword()
		require.Greater(t, len(password), 0)
	}
}

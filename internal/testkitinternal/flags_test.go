package testkitinternal_test

import (
	"testing"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/stretchr/testify/require"
)

func TestMustCreateUserFlagSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	name := "deadbeef"
	flag := testkitinternal.MustCreateUserFlag(t, user.UUID, name)

	require.Equal(t, name, flag.Name)
}

func TestMustCreateUserFlagWrongUserUUID(t *testing.T) {
	t.Parallel()

	testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	defer func() {
		r := recover()
		require.NotNil(t, r)
	}()

	testkitinternal.MustCreateUserFlag(t, "dead-beef-dead-beef", "deadbeef")
}

package flags_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/flags"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/alvii147/flagger-api/pkg/errutils"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestServiceCreateFlagSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := flags.NewRepository()
	svc := flags.NewService(dbPool, repo)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user.UUID)
	name := "my-flag"
	now := time.Now().UTC()
	flag, err := svc.CreateFlag(ctx, name)
	require.NoError(t, err)

	require.Equal(t, user.UUID, flag.UserUUID)
	require.Equal(t, name, flag.Name)
	require.False(t, flag.IsEnabled)
	testkit.RequireTimeAlmostEqual(t, now, flag.CreatedAt)
	testkit.RequireTimeAlmostEqual(t, now, flag.UpdatedAt)
}

func TestServiceCreateFlagAlreadyExists(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	name := "my-flag"
	testkitinternal.MustCreateUserFlag(t, user.UUID, name)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := flags.NewRepository()
	svc := flags.NewService(dbPool, repo)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user.UUID)
	_, err := svc.CreateFlag(ctx, name)
	require.ErrorIs(t, err, errutils.ErrFlagAlreadyExists)
}

func TestServiceGetFlagByIDSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	flag := testkitinternal.MustCreateUserFlag(t, user.UUID, "my-flag")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := flags.NewRepository()
	svc := flags.NewService(dbPool, repo)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user.UUID)
	fetchedFlag, err := svc.GetFlagByID(ctx, flag.ID)
	require.NoError(t, err)

	require.Equal(t, flag.ID, fetchedFlag.ID)
	require.Equal(t, user.UUID, fetchedFlag.UserUUID)
	require.Equal(t, flag.Name, fetchedFlag.Name)
	require.Equal(t, flag.IsEnabled, fetchedFlag.IsEnabled)
	require.Equal(t, flag.CreatedAt, fetchedFlag.CreatedAt)
	require.Equal(t, flag.UpdatedAt, fetchedFlag.UpdatedAt)
}

func TestServiceGetFlagByIDError(t *testing.T) {
	t.Parallel()

	activeUser1, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	activeUser2, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	activeUser1Flag := testkitinternal.MustCreateUserFlag(t, activeUser1.UUID, "my-flag")
	inactiveUserFlag := testkitinternal.MustCreateUserFlag(t, inactiveUser.UUID, "my-flag")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := flags.NewRepository()
	svc := flags.NewService(dbPool, repo)

	testcases := []struct {
		name    string
		flagID  int
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "Flag not found",
			flagID:  42,
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, activeUser1.UUID),
			wantErr: errutils.ErrFlagNotFound,
		},
		{
			name:    "Flag not found for user",
			flagID:  activeUser1Flag.ID,
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, activeUser2.UUID),
			wantErr: errutils.ErrFlagNotFound,
		},
		{
			name:    "Inactive user",
			flagID:  inactiveUserFlag.ID,
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, inactiveUser.UUID),
			wantErr: errutils.ErrFlagNotFound,
		},
		{
			name:    "No user UUID in context",
			flagID:  activeUser1Flag.ID,
			ctx:     context.Background(),
			wantErr: nil,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			_, err := svc.GetFlagByID(testcase.ctx, testcase.flagID)
			require.Error(t, err)
			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceGetFlagByNameSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	name := "my-flag"
	flag := testkitinternal.MustCreateUserFlag(t, user.UUID, name)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := flags.NewRepository()
	svc := flags.NewService(dbPool, repo)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user.UUID)
	fetchedFlag, err := svc.GetFlagByName(ctx, name)
	require.NoError(t, err)

	require.Equal(t, flag.ID, fetchedFlag.ID)
	require.Equal(t, user.UUID, fetchedFlag.UserUUID)
	require.Equal(t, flag.Name, fetchedFlag.Name)
	require.Equal(t, flag.IsEnabled, fetchedFlag.IsEnabled)
	require.Equal(t, flag.CreatedAt, fetchedFlag.CreatedAt)
	require.Equal(t, flag.UpdatedAt, fetchedFlag.UpdatedAt)
}

func TestServiceGetFlagByNameError(t *testing.T) {
	t.Parallel()

	activeUser1, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	activeUser2, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	activeUser1Flag := testkitinternal.MustCreateUserFlag(t, activeUser1.UUID, "active-user-1-flag")
	inactiveUserFlag := testkitinternal.MustCreateUserFlag(t, inactiveUser.UUID, "inactive-user-flag")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := flags.NewRepository()
	svc := flags.NewService(dbPool, repo)

	testcases := []struct {
		name     string
		flagName string
		ctx      context.Context
		wantErr  error
	}{
		{
			name:     "Flag not found",
			flagName: "unknown-flag",
			ctx:      context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, activeUser1.UUID),
			wantErr:  errutils.ErrFlagNotFound,
		},
		{
			name:     "Flag not found for user",
			flagName: activeUser1Flag.Name,
			ctx:      context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, activeUser2.UUID),
			wantErr:  errutils.ErrFlagNotFound,
		},
		{
			name:     "Inactive user",
			flagName: inactiveUserFlag.Name,
			ctx:      context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, inactiveUser.UUID),
			wantErr:  errutils.ErrFlagNotFound,
		},
		{
			name:     "No user UUID in context",
			flagName: activeUser1Flag.Name,
			ctx:      context.Background(),
			wantErr:  nil,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			_, err := svc.GetFlagByName(testcase.ctx, testcase.flagName)
			require.Error(t, err)
			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceListFlags(t *testing.T) {
	t.Parallel()

	user1, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	user2, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	user1Flag1 := testkitinternal.MustCreateUserFlag(t, user1.UUID, "user-1-flag-1")
	user1Flag2 := testkitinternal.MustCreateUserFlag(t, user1.UUID, "user-1-flag-2")
	user2Flag := testkitinternal.MustCreateUserFlag(t, user2.UUID, "user-2-flag")

	user1Flags := []*flags.Flag{user1Flag1, user1Flag2}
	user2Flags := []*flags.Flag{user2Flag}

	sort.Slice(user1Flags, func(i, j int) bool {
		return user1Flags[i].Name < user1Flags[j].Name
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := flags.NewRepository()
	svc := flags.NewService(dbPool, repo)

	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user1.UUID)
	fetchedUser1Flags, err := svc.ListFlags(ctx)
	require.NoError(t, err)

	sort.Slice(fetchedUser1Flags, func(i, j int) bool {
		return fetchedUser1Flags[i].Name < fetchedUser1Flags[j].Name
	})

	ctx = context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user2.UUID)
	fetchedUser2Flags, err := svc.ListFlags(ctx)
	require.NoError(t, err)

	require.Len(t, fetchedUser1Flags, 2)
	require.Len(t, fetchedUser2Flags, 1)

	for i, fetchedFlag := range fetchedUser1Flags {
		wantFlag := user1Flags[i]
		require.Equal(t, wantFlag.ID, fetchedFlag.ID)
		require.Equal(t, wantFlag.UserUUID, fetchedFlag.UserUUID)
		require.Equal(t, wantFlag.Name, fetchedFlag.Name)
		require.Equal(t, wantFlag.IsEnabled, fetchedFlag.IsEnabled)
		require.Equal(t, wantFlag.CreatedAt, fetchedFlag.CreatedAt)
		require.Equal(t, wantFlag.UpdatedAt, fetchedFlag.UpdatedAt)
	}

	for i, fetchedFlag := range fetchedUser2Flags {
		wantFlag := user2Flags[i]
		require.Equal(t, wantFlag.ID, fetchedFlag.ID)
		require.Equal(t, wantFlag.UserUUID, fetchedFlag.UserUUID)
		require.Equal(t, wantFlag.Name, fetchedFlag.Name)
		require.Equal(t, wantFlag.IsEnabled, fetchedFlag.IsEnabled)
		require.Equal(t, wantFlag.CreatedAt, fetchedFlag.CreatedAt)
		require.Equal(t, wantFlag.UpdatedAt, fetchedFlag.UpdatedAt)
	}
}

func TestServiceUpdateFlagSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	flagName := "my-flag"
	flag := testkitinternal.MustCreateUserFlag(t, user.UUID, flagName)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := flags.NewRepository()
	svc := flags.NewService(dbPool, repo)

	updatedIsEnabled := true

	updatedAt := time.Now().UTC()
	ctx := context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, user.UUID)
	updatedFlag, err := svc.UpdateFlag(ctx, flag.ID, updatedIsEnabled)
	require.NoError(t, err)

	require.Equal(t, flag.ID, updatedFlag.ID)
	require.Equal(t, user.UUID, updatedFlag.UserUUID)
	require.Equal(t, flagName, updatedFlag.Name)
	require.Equal(t, updatedIsEnabled, updatedFlag.IsEnabled)
	require.Equal(t, flag.CreatedAt, updatedFlag.CreatedAt)
	testkit.RequireTimeAlmostEqual(t, updatedAt, updatedFlag.UpdatedAt)

}

func TestServiceUpdateFlagError(t *testing.T) {
	t.Parallel()

	activeUser1, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	activeUser2, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	activeUser1Flag := testkitinternal.MustCreateUserFlag(t, activeUser1.UUID, "my-flag")
	inactiveUserFlag := testkitinternal.MustCreateUserFlag(t, inactiveUser.UUID, "my-flag")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := flags.NewRepository()
	svc := flags.NewService(dbPool, repo)

	testcases := []struct {
		name    string
		flagID  int
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "Flag not found",
			flagID:  42,
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, activeUser1.UUID),
			wantErr: errutils.ErrFlagNotFound,
		},
		{
			name:    "Flag not found for user",
			flagID:  activeUser1Flag.ID,
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, activeUser2.UUID),
			wantErr: errutils.ErrFlagNotFound,
		},
		{
			name:    "Inactive user",
			flagID:  inactiveUserFlag.ID,
			ctx:     context.WithValue(context.Background(), auth.AuthContextKeyUserUUID, inactiveUser.UUID),
			wantErr: errutils.ErrFlagNotFound,
		},
		{
			name:    "No user UUID in context",
			flagID:  activeUser1Flag.ID,
			ctx:     context.Background(),
			wantErr: nil,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			_, err := svc.UpdateFlag(testcase.ctx, testcase.flagID, true)
			require.Error(t, err)
			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

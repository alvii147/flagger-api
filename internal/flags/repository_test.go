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

func TestRepositoryCreateFlagSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := flags.NewRepository()

	flag := &flags.Flag{
		UserUUID: user.UUID,
		Name:     "my-flag",
	}

	now := time.Now().UTC()
	createdFlag, err := repo.CreateFlag(dbConn, flag)
	require.NoError(t, err)

	require.Equal(t, flag.UserUUID, createdFlag.UserUUID)
	require.Equal(t, flag.Name, createdFlag.Name)
	require.False(t, createdFlag.IsEnabled)
	testkit.RequireTimeAlmostEqual(t, now, createdFlag.CreatedAt)
	testkit.RequireTimeAlmostEqual(t, now, createdFlag.UpdatedAt)
}

func TestRepositoryCreateFlagDuplicateName(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	testkitinternal.MustCreateUserFlag(t, user.UUID, "my-flag")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := flags.NewRepository()

	flag := &flags.Flag{
		UserUUID: user.UUID,
		Name:     "my-flag",
	}

	_, err := repo.CreateFlag(dbConn, flag)
	require.ErrorIs(t, err, errutils.ErrDatabaseUniqueViolation)
}

func TestRepositoryGetFlagByIDSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	flag := testkitinternal.MustCreateUserFlag(t, user.UUID, "my-flag")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := flags.NewRepository()

	fetchedFlag, err := repo.GetFlagByID(dbConn, flag.ID, user.UUID)
	require.NoError(t, err)

	require.Equal(t, flag.ID, fetchedFlag.ID)
	require.Equal(t, flag.UserUUID, fetchedFlag.UserUUID)
	require.Equal(t, flag.Name, fetchedFlag.Name)
	require.False(t, fetchedFlag.IsEnabled)
	testkit.RequireTimeAlmostEqual(t, flag.CreatedAt, fetchedFlag.CreatedAt)
	testkit.RequireTimeAlmostEqual(t, flag.UpdatedAt, fetchedFlag.UpdatedAt)
}

func TestRepositoryGetFlagByIDError(t *testing.T) {
	t.Parallel()

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	inactiveUserFlag := testkitinternal.MustCreateUserFlag(t, inactiveUser.UUID, "my-flag")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := flags.NewRepository()

	testcases := []struct {
		name     string
		flagID   int
		userUUID string
	}{
		{
			name:     "Active user with no flags",
			flagID:   42,
			userUUID: activeUser.UUID,
		},
		{
			name:     "Inactive user with valid flag",
			flagID:   inactiveUserFlag.ID,
			userUUID: inactiveUser.UUID,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
			_, err := repo.GetFlagByID(dbConn, testcase.flagID, testcase.userUUID)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)
		})
	}
}

func TestRepositoryGetFlagByNameSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	flag := testkitinternal.MustCreateUserFlag(t, user.UUID, "my-flag")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := flags.NewRepository()

	fetchedFlag, err := repo.GetFlagByName(dbConn, "my-flag", user.UUID)
	require.NoError(t, err)

	require.Equal(t, flag.ID, fetchedFlag.ID)
	require.Equal(t, flag.UserUUID, fetchedFlag.UserUUID)
	require.Equal(t, flag.Name, fetchedFlag.Name)
	require.False(t, fetchedFlag.IsEnabled)
	testkit.RequireTimeAlmostEqual(t, flag.CreatedAt, fetchedFlag.CreatedAt)
	testkit.RequireTimeAlmostEqual(t, flag.UpdatedAt, fetchedFlag.UpdatedAt)
}

func TestRepositoryGetFlagByNameError(t *testing.T) {
	t.Parallel()

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	inactiveUserFlag := testkitinternal.MustCreateUserFlag(t, inactiveUser.UUID, "my-flag")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := flags.NewRepository()

	testcases := []struct {
		name     string
		flagName string
		userUUID string
	}{
		{
			name:     "Active user with no flags",
			flagName: "unknown-flag",
			userUUID: activeUser.UUID,
		},
		{
			name:     "Inactive user with valid flag",
			flagName: inactiveUserFlag.Name,
			userUUID: inactiveUser.UUID,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
			_, err := repo.GetFlagByName(dbConn, testcase.flagName, testcase.userUUID)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)
		})
	}
}

func TestRepositoryListFlagsByUserUUIDSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	flag1 := testkitinternal.MustCreateUserFlag(t, user.UUID, "flag-1")
	flag2 := testkitinternal.MustCreateUserFlag(t, user.UUID, "flag-2")

	wantFlags := []*flags.Flag{flag1, flag2}
	sort.Slice(wantFlags, func(i, j int) bool {
		return wantFlags[i].Name < wantFlags[j].Name
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := flags.NewRepository()

	userFlags, err := repo.ListFlagsByUserUUID(dbConn, user.UUID)
	require.NoError(t, err)
	require.Len(t, userFlags, 2)

	sort.Slice(userFlags, func(i, j int) bool {
		return userFlags[i].Name < userFlags[j].Name
	})

	for i, flag := range userFlags {
		wantFlag := wantFlags[i]
		require.Equal(t, wantFlag.ID, flag.ID)
		require.Equal(t, wantFlag.UserUUID, flag.UserUUID)
		require.Equal(t, wantFlag.Name, flag.Name)
		require.Equal(t, wantFlag.IsEnabled, flag.IsEnabled)
		testkit.RequireTimeAlmostEqual(t, wantFlag.CreatedAt, flag.CreatedAt)
		testkit.RequireTimeAlmostEqual(t, wantFlag.UpdatedAt, flag.UpdatedAt)
	}
}

func TestRepositoryListFlagsByUserUUIDInactiveUser(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	testkitinternal.MustCreateUserFlag(t, user.UUID, "my-flag")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := flags.NewRepository()

	userFlags, err := repo.ListFlagsByUserUUID(dbConn, user.UUID)
	require.NoError(t, err)
	require.Empty(t, userFlags)
}

func TestRepositoryUpdateFlagByIDSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	flagName := "my-flag"
	createdAt := time.Now().UTC()
	flag := testkitinternal.MustCreateUserFlag(t, user.UUID, flagName)

	time.Sleep(4 * time.Second)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := flags.NewRepository()

	flag = &flags.Flag{
		ID:        flag.ID,
		UserUUID:  user.UUID,
		IsEnabled: true,
	}

	updatedAt := time.Now().UTC()
	updatedFlag, err := repo.UpdateFlagByID(dbConn, flag)
	require.NoError(t, err)

	require.Equal(t, flag.ID, updatedFlag.ID)
	require.Equal(t, flag.UserUUID, updatedFlag.UserUUID)
	require.Equal(t, flagName, updatedFlag.Name)
	require.True(t, updatedFlag.IsEnabled)
	testkit.RequireTimeAlmostEqual(t, createdAt, updatedFlag.CreatedAt)
	testkit.RequireTimeAlmostEqual(t, updatedAt, updatedFlag.UpdatedAt)
}

func TestRepositoryUpdateFlagByIDError(t *testing.T) {
	t.Parallel()

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	inactiveUserFlag := testkitinternal.MustCreateUserFlag(t, inactiveUser.UUID, "my-flag")

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := flags.NewRepository()

	testcases := []struct {
		name     string
		flagID   int
		userUUID string
	}{
		{
			name:     "Active user with no flag",
			flagID:   42,
			userUUID: activeUser.UUID,
		},
		{
			name:     "Inactive user with valid flag",
			flagID:   inactiveUserFlag.ID,
			userUUID: inactiveUser.UUID,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			flag := &flags.Flag{
				ID:        testcase.flagID,
				UserUUID:  testcase.userUUID,
				IsEnabled: true,
			}

			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
			_, err := repo.UpdateFlagByID(dbConn, flag)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsAffected)
		})
	}
}

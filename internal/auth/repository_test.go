package auth_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/alvii147/flagger-api/pkg/errutils"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestRepositoryCreateUserSuccess(t *testing.T) {
	t.Parallel()

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := auth.NewRepository()

	user := &auth.User{
		UUID:        uuid.NewString(),
		Email:       testkit.GenerateFakeEmail(),
		Password:    testkitinternal.MustHashPassword(testkit.GenerateFakePassword()),
		FirstName:   testkit.MustGenerateRandomString(8, true, true, false),
		LastName:    testkit.MustGenerateRandomString(8, true, true, false),
		IsActive:    false,
		IsSuperUser: false,
	}

	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())

	createdUser, err := repo.CreateUser(dbConn, user)
	require.NoError(t, err)

	require.Equal(t, user.Email, createdUser.Email)
	require.Equal(t, user.Password, createdUser.Password)
	require.Equal(t, user.FirstName, createdUser.FirstName)
	require.Equal(t, user.LastName, createdUser.LastName)
	require.Equal(t, user.IsActive, createdUser.IsActive)
	require.Equal(t, user.IsSuperUser, createdUser.IsSuperUser)
}

func TestRepositoryCreateUserDuplicateEmail(t *testing.T) {
	t.Parallel()

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := auth.NewRepository()

	email := testkit.GenerateFakeEmail()
	user1 := &auth.User{
		UUID:        uuid.NewString(),
		Email:       email,
		Password:    testkitinternal.MustHashPassword(testkit.GenerateFakePassword()),
		FirstName:   testkit.MustGenerateRandomString(8, true, true, false),
		LastName:    testkit.MustGenerateRandomString(8, true, true, false),
		IsActive:    false,
		IsSuperUser: false,
	}

	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())

	_, err := repo.CreateUser(dbConn, user1)
	require.NoError(t, err)

	user2 := &auth.User{
		UUID:        uuid.NewString(),
		Email:       email,
		Password:    testkitinternal.MustHashPassword(testkit.GenerateFakePassword()),
		FirstName:   testkit.MustGenerateRandomString(8, true, true, false),
		LastName:    testkit.MustGenerateRandomString(8, true, true, false),
		IsActive:    false,
		IsSuperUser: false,
	}

	_, err = repo.CreateUser(dbConn, user2)
	require.ErrorIs(t, err, errutils.ErrDatabaseUniqueViolation)
}

func TestRepositoryActivateUserByUUIDSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := auth.NewRepository()

	_, err := repo.GetUserByEmail(dbConn, user.Email)
	require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)

	err = repo.ActivateUserByUUID(dbConn, user.UUID)
	require.NoError(t, err)

	fetchedUser, err := repo.GetUserByEmail(dbConn, user.Email)
	require.NoError(t, err)

	require.True(t, fetchedUser.IsActive)
}

func TestRepositoryActivateUserByUUIDError(t *testing.T) {
	t.Parallel()

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := auth.NewRepository()

	testcases := []struct {
		name     string
		userUUID string
	}{
		{
			name:     "No user under given UUID",
			userUUID: uuid.NewString(),
		},
		{
			name:     "No inactive user under given UUID",
			userUUID: activeUser.UUID,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
			err := repo.ActivateUserByUUID(dbConn, uuid.New().String())
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsAffected)
		})
	}
}

func TestRepositoryGetUserByEmailSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := auth.NewRepository()

	fetchedUser, err := repo.GetUserByEmail(dbConn, user.Email)
	require.NoError(t, err)

	require.Equal(t, user.UUID, fetchedUser.UUID)
	require.Equal(t, user.Email, fetchedUser.Email)
	require.Equal(t, user.Password, fetchedUser.Password)
	require.Equal(t, user.FirstName, fetchedUser.FirstName)
	require.Equal(t, user.LastName, fetchedUser.LastName)
	require.Equal(t, user.IsActive, fetchedUser.IsActive)
	require.Equal(t, user.IsSuperUser, fetchedUser.IsSuperUser)

	testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), fetchedUser.CreatedAt)
}

func TestRepositoryGetUserByEmailError(t *testing.T) {
	t.Parallel()

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := auth.NewRepository()

	testcases := []struct {
		name  string
		email string
	}{
		{
			name:  "No user",
			email: testkit.GenerateFakeEmail(),
		},
		{
			name:  "No active user",
			email: inactiveUser.Email,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
			_, err := repo.GetUserByEmail(dbConn, testcase.email)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)
		})
	}
}

func TestRepositoryGetUserByUUIDSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := auth.NewRepository()

	fetchedUser, err := repo.GetUserByUUID(dbConn, user.UUID)
	require.NoError(t, err)

	require.Equal(t, user.UUID, fetchedUser.UUID)
	require.Equal(t, user.Email, fetchedUser.Email)
	require.Equal(t, user.Password, fetchedUser.Password)
	require.Equal(t, user.FirstName, fetchedUser.FirstName)
	require.Equal(t, user.LastName, fetchedUser.LastName)
	require.Equal(t, user.IsActive, fetchedUser.IsActive)
	require.Equal(t, user.IsSuperUser, fetchedUser.IsSuperUser)
}

func TestRepositoryGetUserByUUIDError(t *testing.T) {
	t.Parallel()

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := auth.NewRepository()

	testcases := []struct {
		name     string
		userUUID string
	}{
		{
			name:     "No user under UUID",
			userUUID: uuid.NewString(),
		},
		{
			name:     "No active user",
			userUUID: inactiveUser.UUID,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
			_, err := repo.GetUserByUUID(dbConn, testcase.userUUID)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsReturned)
		})
	}
}

func TestRepositoryUpdateUserSuccess(t *testing.T) {
	t.Parallel()

	startingFirstName := "Firstname"
	startingLastName := "Lastname"
	updatedFirstName := "Updatedfirstname"
	updatedLastName := "Updatedlastname"

	testcases := []struct {
		name          string
		firstName     *string
		lastName      *string
		wantFirstName string
		wantLastName  string
	}{
		{
			name:          "Update both first and last names",
			firstName:     &updatedFirstName,
			lastName:      &updatedLastName,
			wantFirstName: updatedFirstName,
			wantLastName:  updatedLastName,
		},
		{
			name:          "Update first name",
			firstName:     &updatedFirstName,
			lastName:      nil,
			wantFirstName: updatedFirstName,
			wantLastName:  startingLastName,
		},
		{
			name:          "Update last name",
			firstName:     nil,
			lastName:      &updatedLastName,
			wantFirstName: startingFirstName,
			wantLastName:  updatedLastName,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
				u.FirstName = startingFirstName
				u.LastName = startingLastName
				u.IsActive = true
			})

			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
			repo := auth.NewRepository()

			updatedUser, err := repo.UpdateUser(dbConn, user.UUID, testcase.firstName, testcase.lastName)
			require.NoError(t, err)

			require.Equal(t, user.UUID, updatedUser.UUID)
			require.Equal(t, user.Email, updatedUser.Email)
			require.Equal(t, user.Password, updatedUser.Password)
			require.Equal(t, testcase.wantFirstName, updatedUser.FirstName)
			require.Equal(t, testcase.wantLastName, updatedUser.LastName)
			require.Equal(t, user.IsActive, updatedUser.IsActive)
			require.Equal(t, user.IsSuperUser, updatedUser.IsSuperUser)
			require.Equal(t, user.CreatedAt, updatedUser.CreatedAt)
		})
	}
}

func TestRepositoryUpdateUserError(t *testing.T) {
	t.Parallel()

	updatedFirstName := "Updatedfirstname"
	updatedLastName := "Updatedlastname"

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	testcases := []struct {
		name      string
		userUUID  string
		firstName *string
		lastName  *string
	}{
		{
			name:      "Update neither first nor last name",
			userUUID:  activeUser.UUID,
			firstName: nil,
			lastName:  nil,
		},
		{
			name:      "Update non-existent user",
			userUUID:  uuid.NewString(),
			firstName: &updatedFirstName,
			lastName:  &updatedLastName,
		},
		{
			name:      "Update inactive user",
			userUUID:  inactiveUser.UUID,
			firstName: &updatedFirstName,
			lastName:  &updatedLastName,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			dbPool := testkitinternal.RequireCreateDatabasePool(t)
			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
			repo := auth.NewRepository()

			_, err := repo.UpdateUser(dbConn, testcase.userUUID, testcase.firstName, testcase.lastName)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsAffected)
		})
	}
}

func TestRepositoryCreateAPIKeySuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := auth.NewRepository()

	apiKey := &auth.APIKey{
		UserUUID:  user.UUID,
		Prefix:    testkit.MustGenerateRandomString(8, true, true, true),
		HashedKey: testkit.MustGenerateRandomString(16, true, true, true),
		Name:      "My API Key",
		ExpiresAt: pgtype.Timestamp{
			Valid: false,
		},
	}

	createdAPIKey, err := repo.CreateAPIKey(dbConn, apiKey)
	require.NoError(t, err)

	require.Equal(t, apiKey.Prefix, createdAPIKey.Prefix)
	require.Equal(t, apiKey.HashedKey, createdAPIKey.HashedKey)
	require.Equal(t, apiKey.Name, createdAPIKey.Name)
	require.False(t, apiKey.ExpiresAt.Valid)
}

func TestRepositoryCreateAPIKeyDuplicateName(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	testkitinternal.MustCreateUserAPIKey(t, user.UUID, func(k *auth.APIKey) {
		k.Name = "My API Key"
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := auth.NewRepository()

	apiKey := &auth.APIKey{
		UserUUID:  user.UUID,
		Prefix:    testkit.MustGenerateRandomString(8, true, true, true),
		HashedKey: testkit.MustGenerateRandomString(16, true, true, true),
		Name:      "My API Key",
		ExpiresAt: pgtype.Timestamp{
			Valid: false,
		},
	}

	_, err := repo.CreateAPIKey(dbConn, apiKey)
	require.ErrorIs(t, err, errutils.ErrDatabaseUniqueViolation)
}

func TestRepositoryListAPIKeysByUserUUIDSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	apiKey1, _ := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)
	apiKey2, _ := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	wantAPIKeys := []*auth.APIKey{apiKey1, apiKey2}
	sort.Slice(wantAPIKeys, func(i, j int) bool {
		return wantAPIKeys[i].Prefix < wantAPIKeys[j].Prefix
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := auth.NewRepository()

	apiKeys, err := repo.ListAPIKeysByUserUUID(dbConn, user.UUID)
	require.NoError(t, err)
	require.Len(t, apiKeys, 2)

	sort.Slice(apiKeys, func(i, j int) bool {
		return apiKeys[i].Prefix < apiKeys[j].Prefix
	})

	for i, apiKey := range apiKeys {
		wantKey := wantAPIKeys[i]
		require.Equal(t, wantKey.ID, apiKey.ID)
		require.Equal(t, wantKey.UserUUID, apiKey.UserUUID)
		require.Equal(t, wantKey.Prefix, apiKey.Prefix)
		require.Equal(t, wantKey.Name, apiKey.Name)
		require.Equal(t, wantKey.CreatedAt, apiKey.CreatedAt)
		require.Equal(t, wantKey.ExpiresAt, apiKey.ExpiresAt)
	}
}

func TestRepositoryListAPIKeysByUserUUIDInactiveUser(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := auth.NewRepository()

	apiKeys, err := repo.ListAPIKeysByUserUUID(dbConn, user.UUID)
	require.NoError(t, err)
	require.Empty(t, apiKeys)
}

func TestRepositoryListActiveAPIKeysByPrefixSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	apiKey, _ := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := auth.NewRepository()

	apiKeys, err := repo.ListActiveAPIKeysByPrefix(dbConn, apiKey.Prefix)
	require.NoError(t, err)
	require.Len(t, apiKeys, 1)

	var fetchedAPIKey *auth.APIKey
	for _, k := range apiKeys {
		if k.ID == apiKey.ID {
			fetchedAPIKey = k
			break
		}
	}

	require.NotNil(t, fetchedAPIKey)
	require.Equal(t, apiKey.ID, fetchedAPIKey.ID)
	require.Equal(t, apiKey.Prefix, fetchedAPIKey.Prefix)
	require.Equal(t, apiKey.HashedKey, fetchedAPIKey.HashedKey)
	require.Equal(t, apiKey.Name, fetchedAPIKey.Name)
	require.Equal(t, apiKey.ExpiresAt, fetchedAPIKey.ExpiresAt)
}

func TestRepositoryListActiveAPIKeysByPrefixEmpty(t *testing.T) {
	t.Parallel()

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	expiredAPIKey, _ := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, func(k *auth.APIKey) {
		k.ExpiresAt = pgtype.Timestamp{
			Time:  time.Now().UTC().AddDate(0, 0, -1),
			Valid: true,
		}
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	validAPIKey, _ := testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := auth.NewRepository()

	testcases := []struct {
		name   string
		apiKey *auth.APIKey
	}{
		{
			name:   "Active user with expired API key",
			apiKey: expiredAPIKey,
		},
		{
			name:   "Inactive user with valid API key",
			apiKey: validAPIKey,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
			apiKeys, err := repo.ListActiveAPIKeysByPrefix(dbConn, testcase.apiKey.Prefix)
			require.NoError(t, err)

			for _, k := range apiKeys {
				require.NotEqual(t, testcase.apiKey.ID, k.ID)
			}
		})
	}
}

func TestRepositoryDeleteAPIKeySuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	apiKey, _ := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := auth.NewRepository()

	err := repo.DeleteAPIKey(dbConn, apiKey.ID, user.UUID)
	require.NoError(t, err)

	apiKeys, err := repo.ListAPIKeysByUserUUID(dbConn, user.UUID)
	require.NoError(t, err)
	require.Empty(t, apiKeys)
}

func TestRepositoryDeleteAPIKeyError(t *testing.T) {
	t.Parallel()

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	inactiveUserAPIKey, _ := testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	repo := auth.NewRepository()

	testcases := []struct {
		name     string
		apiKeyID int
		userUUID string
	}{
		{
			name:     "Active user with no API keys",
			apiKeyID: 42,
			userUUID: activeUser.UUID,
		},
		{
			name:     "Inactive user with valid API key",
			apiKeyID: inactiveUserAPIKey.ID,
			userUUID: inactiveUser.UUID,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
			err := repo.DeleteAPIKey(dbConn, testcase.apiKeyID, testcase.userUUID)
			require.ErrorIs(t, err, errutils.ErrDatabaseNoRowsAffected)
		})
	}
}

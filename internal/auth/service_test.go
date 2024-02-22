package auth_test

import (
	"context"
	"regexp"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/internal/templatesmanager"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/errutils"
	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/alvii147/flagger-api/pkg/utils"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestServiceCreateUserSuccess(t *testing.T) {
	t.Parallel()

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	email := testkit.GenerateFakeEmail()
	password := testkit.GenerateFakePassword()
	firstName := testkit.MustGenerateRandomString(8, true, true, false)
	lastName := testkit.MustGenerateRandomString(8, true, true, false)

	mailCount := len(mailClient.MailLogs)

	var wg sync.WaitGroup
	user, err := svc.CreateUser(context.Background(), &wg, email, password, firstName, lastName)
	require.NoError(t, err)

	require.Equal(t, email, user.Email)
	require.Equal(t, firstName, user.FirstName)
	require.Equal(t, lastName, user.LastName)
	require.False(t, user.IsActive)
	require.False(t, user.IsSuperUser)

	testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), user.CreatedAt)

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	require.NoError(t, err)

	wg.Wait()

	require.Len(t, mailClient.MailLogs, mailCount+1)

	lastMail := mailClient.MailLogs[len(mailClient.MailLogs)-1]
	require.Equal(t, []string{user.Email}, lastMail.To)
	require.Equal(t, "Welcome to Flagger!", lastMail.Subject)
	testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), lastMail.SentAt)

	mailMessage := string(lastMail.Message)
	require.Contains(t, mailMessage, "Welcome to Flagger!")
	require.Contains(t, mailMessage, "Flagger - Activate Your Account")
}

func TestServiceCreateUserEmailExists(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	mailCount := len(mailClient.MailLogs)

	var wg sync.WaitGroup
	_, err := svc.CreateUser(
		context.Background(),
		&wg,
		user.Email,
		testkit.GenerateFakePassword(),
		testkit.MustGenerateRandomString(8, true, true, false),
		testkit.MustGenerateRandomString(8, true, true, false),
	)
	require.ErrorIs(t, err, errutils.ErrUserAlreadyExists)

	require.Len(t, mailClient.MailLogs, mailCount)
}

func TestServiceActivateUserSuccess(t *testing.T) {
	t.Parallel()

	config := env.GetConfig()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	now := time.Now().UTC()
	jti := uuid.NewString()
	token, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.ActivationJWTClaims{
			Subject:   user.UUID,
			TokenType: string(auth.JWTTypeActivation),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(config.SecretKey))
	require.NoError(t, err)

	err = svc.ActivateUser(context.Background(), token)
	require.NoError(t, err)

	activatedUser, err := repo.GetUserByEmail(dbConn, user.Email)
	require.NoError(t, err)

	require.Equal(t, user.Email, activatedUser.Email)
	require.Equal(t, user.Password, activatedUser.Password)
	require.Equal(t, user.FirstName, activatedUser.FirstName)
	require.Equal(t, user.LastName, activatedUser.LastName)
	require.True(t, activatedUser.IsActive)
	require.Equal(t, user.IsSuperUser, activatedUser.IsSuperUser)
}

func TestServiceActivateUserError(t *testing.T) {
	t.Parallel()

	config := env.GetConfig()

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	now := time.Now().UTC()
	invalidToken := "ed0730889507fdb8549acfcd31548ee5"
	expiredToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.ActivationJWTClaims{
			Subject:   uuid.NewString(),
			TokenType: string(auth.JWTTypeActivation),
			IssuedAt:  utils.JSONTimeStamp(now.Add(-2 * time.Hour)),
			ExpiresAt: utils.JSONTimeStamp(now.Add(-time.Hour)),
			JWTID:     uuid.NewString(),
		},
	).SignedString([]byte(config.SecretKey))
	require.NoError(t, err)
	validToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.ActivationJWTClaims{
			Subject:   uuid.NewString(),
			TokenType: string(auth.JWTTypeActivation),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Hour)),
			JWTID:     uuid.NewString(),
		},
	).SignedString([]byte(config.SecretKey))
	require.NoError(t, err)

	testcases := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "Invalid token",
			token:   invalidToken,
			wantErr: errutils.ErrInvalidToken,
		},
		{
			name:    "Expired token",
			token:   expiredToken,
			wantErr: errutils.ErrInvalidToken,
		},
		{
			name:    "Valid token with invalid user UUID",
			token:   validToken,
			wantErr: errutils.ErrUserNotFound,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			err = svc.ActivateUser(context.Background(), testcase.token)
			require.ErrorIs(t, err, testcase.wantErr)
		})
	}
}

func TestServiceGetCurrentUserSuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	ctx := context.WithValue(context.Background(), auth.UserUUIDContextKey, user.UUID)
	currentUser, err := svc.GetCurrentUser(ctx)
	require.NoError(t, err)

	require.Equal(t, user.Email, currentUser.Email)
	require.Equal(t, user.Password, currentUser.Password)
	require.Equal(t, user.FirstName, currentUser.FirstName)
	require.Equal(t, user.LastName, currentUser.LastName)
	require.Equal(t, user.IsActive, currentUser.IsActive)
	require.Equal(t, user.IsSuperUser, currentUser.IsSuperUser)
}

func TestServiceGetCurrentUserError(t *testing.T) {
	t.Parallel()

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	testcases := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "User not found",
			ctx:     context.WithValue(context.Background(), auth.UserUUIDContextKey, uuid.NewString()),
			wantErr: errutils.ErrUserNotFound,
		},
		{
			name:    "No user UUID in context",
			ctx:     context.Background(),
			wantErr: nil,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			_, err := svc.GetCurrentUser(testcase.ctx)
			require.Error(t, err)
			if testcase.wantErr != nil {
				require.ErrorIs(t, err, testcase.wantErr)
			}
		})
	}
}

func TestServiceCreateJWTSuccess(t *testing.T) {
	t.Parallel()

	config := env.GetConfig()

	user, password := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	accessToken, refreshToken, err := svc.CreateJWT(context.Background(), user.Email, password)
	require.NoError(t, err)

	accessClaims := &api.AuthJWTClaims{}
	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, accessClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.SecretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedAccessToken)
	require.True(t, parsedAccessToken.Valid)
	require.Equal(t, user.UUID, accessClaims.Subject)
	require.Equal(t, string(auth.JWTTypeAccess), accessClaims.TokenType)

	testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), time.Time(accessClaims.IssuedAt))
	testkit.RequireTimeAlmostEqual(t, time.Now().UTC().Add(time.Duration(config.AuthAccessLifetime)), time.Time(accessClaims.ExpiresAt))

	refreshClaims := &api.AuthJWTClaims{}
	parsedRefreshToken, err := jwt.ParseWithClaims(refreshToken, refreshClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.SecretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedRefreshToken)
	require.True(t, parsedRefreshToken.Valid)
	require.Equal(t, user.UUID, refreshClaims.Subject)
	require.Equal(t, string(auth.JWTTypeRefresh), refreshClaims.TokenType)

	testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), time.Time(refreshClaims.IssuedAt))
	testkit.RequireTimeAlmostEqual(t, time.Now().UTC().Add(time.Duration(config.AuthRefreshLifetime)), time.Time(refreshClaims.ExpiresAt))
}

func TestServiceCreateJWTError(t *testing.T) {
	t.Parallel()

	user, password := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	testcases := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{
			name:     "Incorrect email",
			email:    testkit.GenerateFakeEmail(),
			password: password,
			wantErr:  errutils.ErrInvalidCredentials,
		},
		{
			name:     "Incorrect password",
			email:    user.Email,
			password: testkit.GenerateFakePassword(),
			wantErr:  errutils.ErrInvalidCredentials,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := svc.CreateJWT(context.Background(), testcase.email, testcase.password)
			require.ErrorIs(t, err, testcase.wantErr)
		})
	}
}

func TestServiceRefreshJWTSuccess(t *testing.T) {
	t.Parallel()

	config := env.GetConfig()

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	userUUID := uuid.NewString()
	_, refreshToken := testkitinternal.MustCreateUserAuthJWTs(userUUID)

	accessToken, err := svc.RefreshJWT(context.Background(), refreshToken)
	require.NoError(t, err)

	claims := &api.AuthJWTClaims{}
	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.SecretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedAccessToken)
	require.True(t, parsedAccessToken.Valid)
	require.Equal(t, userUUID, claims.Subject)
	require.Equal(t, string(auth.JWTTypeAccess), claims.TokenType)

	testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), time.Time(claims.IssuedAt))
	testkit.RequireTimeAlmostEqual(t, time.Now().UTC().Add(time.Duration(config.AuthAccessLifetime)), time.Time(claims.ExpiresAt))
}

func TestServiceRefreshJWTError(t *testing.T) {
	t.Parallel()

	config := env.GetConfig()

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	userUUID := uuid.NewString()
	jti := uuid.NewString()
	now := time.Now().UTC()

	invalidToken := "ed0730889507fdb8549acfcd31548ee5"
	expiredToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(auth.JWTTypeRefresh),
			IssuedAt:  utils.JSONTimeStamp(now.Add(-2 * time.Hour)),
			ExpiresAt: utils.JSONTimeStamp(now.Add(-time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(config.SecretKey))
	require.NoError(t, err)

	testcases := []struct {
		name  string
		token string
	}{
		{
			name:  "Invalid refresh token",
			token: invalidToken,
		},
		{
			name:  "Expired refresh token",
			token: expiredToken,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			_, err := svc.RefreshJWT(context.Background(), testcase.token)
			require.ErrorIs(t, err, errutils.ErrInvalidToken)
		})
	}
}

func TestServiceCreateAPIKeySuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	ctx := context.WithValue(context.Background(), auth.UserUUIDContextKey, user.UUID)
	name := "My API Key"
	expiresAt := pgtype.Timestamp{
		Valid: false,
	}
	now := time.Now().UTC()
	apiKey, rawKey, err := svc.CreateAPIKey(ctx, name, expiresAt)
	require.NoError(t, err)

	require.NotNil(t, apiKey)
	require.Equal(t, apiKey.UserUUID, user.UUID)
	require.Equal(t, apiKey.Name, apiKey.Name)
	require.False(t, apiKey.ExpiresAt.Valid)
	testkit.RequireTimeAlmostEqual(t, now, apiKey.CreatedAt)

	err = bcrypt.CompareHashAndPassword([]byte(apiKey.HashedKey), []byte(rawKey))
	require.NoError(t, err)

	r := regexp.MustCompile(`^(\S+)\.(\S+)$`)
	matches := r.FindStringSubmatch(rawKey)

	require.Len(t, matches, 3)
	require.Equal(t, apiKey.Prefix, matches[1])
}

func TestServiceCreateAPIKeyAlreadyExists(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	testkitinternal.MustCreateUserAPIKey(t, user.UUID, func(k *auth.APIKey) {
		k.Name = "My API Key"
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	ctx := context.WithValue(context.Background(), auth.UserUUIDContextKey, user.UUID)
	name := "My API Key"
	expiresAt := pgtype.Timestamp{
		Valid: false,
	}
	_, _, err := svc.CreateAPIKey(ctx, name, expiresAt)
	require.ErrorIs(t, err, errutils.ErrAPIKeyAlreadyExists)
}

func TestServiceListAPIKeys(t *testing.T) {
	t.Parallel()

	user1, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	user2, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	user1Key1, _ := testkitinternal.MustCreateUserAPIKey(t, user1.UUID, func(k *auth.APIKey) {
		k.Name = "User1APIKey1"
		k.ExpiresAt = pgtype.Timestamp{
			Valid: false,
		}
	})

	user1Key2, _ := testkitinternal.MustCreateUserAPIKey(t, user1.UUID, func(k *auth.APIKey) {
		k.Name = "User1APIKey2"
		k.ExpiresAt = pgtype.Timestamp{
			Time:  time.Now().UTC().AddDate(0, 0, 1),
			Valid: true,
		}
	})

	user2Key, _ := testkitinternal.MustCreateUserAPIKey(t, user2.UUID, func(k *auth.APIKey) {
		k.Name = "User2APIKey"
		k.ExpiresAt = pgtype.Timestamp{
			Time:  time.Now().UTC().AddDate(0, 1, 0),
			Valid: true,
		}
	})

	user1Keys := []*auth.APIKey{user1Key1, user1Key2}
	user2Keys := []*auth.APIKey{user2Key}

	sort.Slice(user1Keys, func(i, j int) bool {
		return user1Keys[i].Prefix < user1Keys[j].Prefix
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	ctx := context.WithValue(context.Background(), auth.UserUUIDContextKey, user1.UUID)
	fetchedUser1Keys, err := svc.ListAPIKeys(ctx)
	require.NoError(t, err)

	sort.Slice(fetchedUser1Keys, func(i, j int) bool {
		return fetchedUser1Keys[i].Prefix < fetchedUser1Keys[j].Prefix
	})

	ctx = context.WithValue(context.Background(), auth.UserUUIDContextKey, user2.UUID)
	fetchedUser2Keys, err := svc.ListAPIKeys(ctx)
	require.NoError(t, err)

	require.Len(t, fetchedUser1Keys, 2)
	require.Len(t, fetchedUser2Keys, 1)

	for i, fetchedKey := range fetchedUser1Keys {
		wantKey := user1Keys[i]
		require.Equal(t, wantKey.ID, fetchedKey.ID)
		require.Equal(t, wantKey.UserUUID, fetchedKey.UserUUID)
		require.Equal(t, wantKey.Prefix, fetchedKey.Prefix)
		require.Equal(t, wantKey.Name, fetchedKey.Name)
		testkit.RequireTimeAlmostEqual(t, wantKey.CreatedAt, fetchedKey.CreatedAt)
		testkit.RequirePGTimestampAlmostEqual(t, wantKey.ExpiresAt, fetchedKey.ExpiresAt)
	}

	for i, fetchedKey := range fetchedUser2Keys {
		wantKey := user2Keys[i]
		require.Equal(t, wantKey.ID, fetchedKey.ID)
		require.Equal(t, wantKey.UserUUID, fetchedKey.UserUUID)
		require.Equal(t, wantKey.Prefix, fetchedKey.Prefix)
		require.Equal(t, wantKey.Name, fetchedKey.Name)
		testkit.RequireTimeAlmostEqual(t, wantKey.CreatedAt, fetchedKey.CreatedAt)
		testkit.RequirePGTimestampAlmostEqual(t, wantKey.ExpiresAt, fetchedKey.ExpiresAt)
	}
}

func TestServiceFindAPIKeySuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	apiKey, rawKey := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	foundAPIKey, err := svc.FindAPIKey(context.Background(), rawKey)
	require.NoError(t, err)

	require.Equal(t, apiKey.ID, foundAPIKey.ID)
	require.Equal(t, apiKey.UserUUID, foundAPIKey.UserUUID)
	require.Equal(t, apiKey.Prefix, foundAPIKey.Prefix)
	require.Equal(t, apiKey.Name, foundAPIKey.Name)
	testkit.RequireTimeAlmostEqual(t, apiKey.CreatedAt, foundAPIKey.CreatedAt)
	testkit.RequirePGTimestampAlmostEqual(t, apiKey.ExpiresAt, foundAPIKey.ExpiresAt)
}

func TestServiceFindAPIKeyNotFound(t *testing.T) {
	t.Parallel()

	_, rawKey, _, err := auth.CreateAPIKey()
	require.NoError(t, err)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	_, err = svc.FindAPIKey(context.Background(), rawKey)
	require.ErrorIs(t, err, errutils.ErrAPIKeyNotFound)
}

func TestServiceDeleteAPIKeySuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	apiKey, _ := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	dbConn := testkitinternal.RequireCreateDatabaseConn(t, dbPool, context.Background())
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	ctx := context.WithValue(context.Background(), auth.UserUUIDContextKey, user.UUID)
	err := svc.DeleteAPIKey(ctx, apiKey.ID)
	require.NoError(t, err)

	apiKeys, err := repo.ListAPIKeysByUserUUID(dbConn, user.UUID)
	require.NoError(t, err)
	require.Empty(t, apiKeys)
}

func TestServiceDeleteAPIKeyNotFound(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	mailClient := mailclient.NewInMemMailClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(dbPool, mailClient, tmplManager, repo)

	ctx := context.WithValue(context.Background(), auth.UserUUIDContextKey, user.UUID)
	err := svc.DeleteAPIKey(ctx, 42)
	require.ErrorIs(t, err, errutils.ErrAPIKeyNotFound)
}

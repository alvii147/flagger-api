package testkitinternal_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestMustHashPasswordSuccess(t *testing.T) {
	t.Parallel()

	password := "C0rr3ctH0rs3B4tt3rySt4p13"
	hashedPassword := testkitinternal.MustHashPassword(password)

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	require.NoError(t, err)
}

func TestMustHashPasswordTooLong(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r)
	}()

	longPassword := strings.Repeat("C0rr3ctH0rs3B4tt3rySt4p13", 3)
	testkitinternal.MustHashPassword(longPassword)
}

func TestMustCreateUserSuccess(t *testing.T) {
	t.Parallel()

	firstName := "dead"
	lastName := "beef"
	user, password := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.FirstName = firstName
		u.LastName = lastName
	})

	require.Equal(t, firstName, user.FirstName)
	require.Equal(t, lastName, user.LastName)

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	require.NoError(t, err)
}

func TestMustCreateUserDuplicateEmail(t *testing.T) {
	t.Parallel()

	email := testkit.GenerateFakeEmail()
	testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.Email = email
	})

	defer func() {
		r := recover()
		require.NotNil(t, r)
	}()

	testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.Email = email
	})
}

func TestMustCreateUserAuthJWTs(t *testing.T) {
	t.Parallel()

	config, err := env.NewConfig()
	require.NoError(t, err)

	userUUID := uuid.NewString()
	accessToken, refreshToken := testkitinternal.MustCreateUserAuthJWTs(userUUID)

	accessClaims := &api.AuthJWTClaims{}
	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, accessClaims, func(t *jwt.Token) (any, error) {
		return []byte(config.SecretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedAccessToken)
	require.True(t, parsedAccessToken.Valid)
	require.Equal(t, userUUID, accessClaims.Subject)
	require.Equal(t, string(auth.JWTTypeAccess), accessClaims.TokenType)

	testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), time.Time(accessClaims.IssuedAt))
	testkit.RequireTimeAlmostEqual(t, time.Now().UTC().Add(time.Duration(config.AuthAccessLifetime*int64(time.Minute))), time.Time(accessClaims.ExpiresAt))

	refreshClaims := &api.AuthJWTClaims{}
	parsedRefreshToken, err := jwt.ParseWithClaims(refreshToken, refreshClaims, func(t *jwt.Token) (any, error) {
		return []byte(config.SecretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedRefreshToken)
	require.True(t, parsedRefreshToken.Valid)
	require.Equal(t, userUUID, refreshClaims.Subject)
	require.Equal(t, string(auth.JWTTypeRefresh), refreshClaims.TokenType)

	testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), time.Time(refreshClaims.IssuedAt))
	testkit.RequireTimeAlmostEqual(t, time.Now().UTC().Add(time.Duration(config.AuthRefreshLifetime*int64(time.Minute))), time.Time(refreshClaims.ExpiresAt))
}

func TestMustCreateUserAPIKeySuccess(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	name := "MyAPIKey"
	apiKey, rawKey := testkitinternal.MustCreateUserAPIKey(t, user.UUID, func(k *auth.APIKey) {
		k.Name = name
	})

	require.Equal(t, name, apiKey.Name)

	err := bcrypt.CompareHashAndPassword([]byte(apiKey.HashedKey), []byte(rawKey))
	require.NoError(t, err)
}

func TestMustCreateUserAPIKeyWrongUserUUID(t *testing.T) {
	t.Parallel()

	testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	defer func() {
		r := recover()
		require.NotNil(t, r)
	}()

	testkitinternal.MustCreateUserAPIKey(t, "dead-beef-dead-beef", nil)
}

package testkitinternal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/alvii147/flagger-api/pkg/utils"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// MustHashPassword hashes a given password and panics on error
func MustHashPassword(password string) string {
	config := env.GetConfig()

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), config.HashingCost)
	if err != nil {
		panic(fmt.Sprintf("MustHashPassword failed to bcrypt.GenerateFromPassword: %v", err))
	}

	hashedPassword := string(hashedPasswordBytes)

	return hashedPassword
}

// MustCreateUser creates and returns a new user and panics on error.
func MustCreateUser(t *testing.T, modifier func(u *auth.User)) (*auth.User, string) {
	dbPool := RequireCreateDatabasePool(t)
	dbConn := RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := auth.NewRepository()

	userUUID := uuid.NewString()
	password := testkit.GenerateFakePassword()
	user := &auth.User{
		UUID:        userUUID,
		Email:       testkit.GenerateFakeEmail(),
		Password:    MustHashPassword(password),
		FirstName:   testkit.MustGenerateRandomString(8, true, true, false),
		LastName:    testkit.MustGenerateRandomString(8, true, true, false),
		IsActive:    true,
		IsSuperUser: false,
	}

	if modifier != nil {
		modifier(user)
	}

	user, err := repo.CreateUser(dbConn, user)
	if err != nil {
		panic(fmt.Sprintf("MustCreateUser failed to repo.CreateUser: %v", err))
	}

	return user, password
}

// MustCreateUserAuthJWTs creates and returns access and refresh JWTs for User and panics on error.
func MustCreateUserAuthJWTs(userUUID string) (string, string) {
	config := env.GetConfig()

	now := time.Now().UTC()
	accessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(auth.JWTTypeAccess),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Duration(config.AuthAccessLifetime * int64(time.Minute)))),
			JWTID:     uuid.NewString(),
		},
	).SignedString([]byte(config.SecretKey))
	if err != nil {
		panic(fmt.Sprintf("MustCreateUserAPIKey failed to jwt.Token.SignedString: %v", err))
	}

	refreshToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(auth.JWTTypeRefresh),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Duration(config.AuthRefreshLifetime * int64(time.Minute)))),
			JWTID:     uuid.NewString(),
		},
	).SignedString([]byte(config.SecretKey))
	if err != nil {
		panic(fmt.Sprintf("MustCreateUserAPIKey failed to jwt.Token.SignedString: %v", err))
	}

	return accessToken, refreshToken
}

// MustCreateUserAPIKey creates and returns a new API key for User and panics on error.
func MustCreateUserAPIKey(t *testing.T, userUUID string, modifier func(k *auth.APIKey)) (*auth.APIKey, string) {
	dbPool := RequireCreateDatabasePool(t)
	dbConn := RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := auth.NewRepository()

	name := testkit.MustGenerateRandomString(12, true, true, true)
	prefix := testkit.MustGenerateRandomString(8, true, true, true)
	secret := testkit.MustGenerateRandomString(32, true, true, true)
	rawKey := fmt.Sprintf("%s.%s", prefix, secret)
	hashedKey := MustHashPassword(rawKey)

	apiKey := &auth.APIKey{
		UserUUID:  userUUID,
		Prefix:    prefix,
		HashedKey: hashedKey,
		Name:      name,
		ExpiresAt: pgtype.Timestamp{
			Valid: false,
		},
	}

	if modifier != nil {
		modifier(apiKey)
	}

	apiKey, err := repo.CreateAPIKey(dbConn, apiKey)
	if err != nil {
		panic(fmt.Sprintf("MustCreateUserAPIKey failed to repo.CreateAPIKey: %v", err))
	}

	return apiKey, rawKey
}

package testkitinternal

import (
	"context"
	"fmt"
	"testing"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/pkg/testkit"
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

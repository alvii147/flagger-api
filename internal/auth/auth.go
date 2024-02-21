package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/alvii147/flagger-api/pkg/utils"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// User represents database table of users.
type User struct {
	UUID        string    `db:"uuid"`
	Email       string    `db:"email"`
	Password    string    `db:"password"`
	FirstName   string    `db:"first_name"`
	LastName    string    `db:"last_name"`
	IsActive    bool      `db:"is_active"`
	IsSuperUser bool      `db:"is_superuser"`
	CreatedAt   time.Time `db:"created_at"`
}

// APIKey represents database table of API keys.
type APIKey struct {
	ID        int              `db:"id"`
	UserUUID  string           `db:"user_uuid"`
	Prefix    string           `db:"prefix"`
	HashedKey string           `db:"hashed_key"`
	Name      string           `db:"name"`
	CreatedAt time.Time        `db:"created_at"`
	ExpiresAt pgtype.Timestamp `db:"expires_at"`
}

// ActivationEmailTemplateData represents data for User activation email templates.
type ActivationEmailTemplateData struct {
	RecipientEmail string
	ActivationURL  string
}

// JWTType is a string representing type of JWT.
// Allowed strings are "access", "refresh", and "activation".
type JWTType string

const (
	JWTTypeAccess     JWTType = "access"
	JWTTypeRefresh    JWTType = "refresh"
	JWTTypeActivation JWTType = "activation"
)

// hashPassword hashes given password.
func hashPassword(password string) (string, error) {
	config := env.GetConfig()

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), config.HashingCost)
	if err != nil {
		return "", fmt.Errorf("hashPassword failed to bcrypt.GenerateFromPassword: %w", err)
	}
	hashedPassword := string(hashedPasswordBytes)

	return hashedPassword, nil
}

// createAuthJWT creates JWTs for User authentication of given type.
// Returns error when token type is not access or refresh.
func createAuthJWT(userUUID string, tokenType JWTType) (string, error) {
	config := env.GetConfig()

	var lifetime time.Duration
	switch tokenType {
	case JWTTypeAccess:
		lifetime = time.Duration(config.AuthAccessLifetime)
	case JWTTypeRefresh:
		lifetime = time.Duration(config.AuthRefreshLifetime)
	default:
		return "", fmt.Errorf("createAuthJWT received invalid JWT type %s, expected JWT type %s or %s", tokenType, JWTTypeAccess, JWTTypeRefresh)
	}

	now := time.Now().UTC()
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(tokenType),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(lifetime)),
			JWTID:     uuid.NewString(),
		},
	)
	signedToken, err := token.SignedString([]byte(config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("createAuthJWTWithType failed to token.SignedString for user.UUID %s of token type %s: %w", userUUID, tokenType, err)
	}

	return signedToken, nil
}

// validateAuthJWT validates JWT for User authentication,
// checks that the JWT is not expired,
// and returns parsed JWT claims.
func validateAuthJWT(token string, tokenType JWTType) (*api.AuthJWTClaims, bool) {
	config := env.GetConfig()

	claims := &api.AuthJWTClaims{}
	ok := true

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.SecretKey), nil
	})

	if err != nil {
		ok = false
	}

	if parsedToken == nil || !parsedToken.Valid {
		ok = false
	}

	if subtle.ConstantTimeCompare([]byte(claims.TokenType), []byte(tokenType)) == 0 {
		ok = false
	}

	if time.Now().UTC().After(time.Time(claims.ExpiresAt)) {
		ok = false
	}

	if !ok {
		claims = nil
	}

	return claims, ok
}

// createActivationJWTWithType creates JWT for User activation.
func createActivationJWT(userUUID string) (string, error) {
	config := env.GetConfig()

	now := time.Now().UTC()
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.ActivationJWTClaims{
			Subject:   userUUID,
			TokenType: string(JWTTypeActivation),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Duration(config.ActivationLifetime))),
			JWTID:     uuid.NewString(),
		},
	)
	signedToken, err := token.SignedString([]byte(config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("createActivationJWT failed to token.SignedString for user.UUID %s of token type %s: %w", userUUID, JWTTypeActivation, err)
	}

	return signedToken, nil
}

// validateActivationJWT validates JWT for User activation using secret key,
// checks that the JWT is not expired,
// and returns parsed JWT claims.
func validateActivationJWT(token string) (*api.ActivationJWTClaims, bool) {
	config := env.GetConfig()

	claims := &api.ActivationJWTClaims{}
	ok := true

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.SecretKey), nil
	})

	if err != nil {
		ok = false
	}

	if parsedToken == nil || !parsedToken.Valid {
		ok = false
	}

	if subtle.ConstantTimeCompare([]byte(claims.TokenType), []byte(JWTTypeActivation)) == 0 {
		ok = false
	}

	if time.Now().UTC().After(time.Time(claims.ExpiresAt)) {
		ok = false
	}

	if !ok {
		return nil, false
	}

	return claims, true
}

// sendActivationMail sends activation email to User.
func sendActivationMail(user *User, mailClient mailclient.MailClient) error {
	config := env.GetConfig()

	activationToken, err := createActivationJWT(user.UUID)
	if err != nil {
		return fmt.Errorf("sendActivationMail failed to createActivationJWT: %w", err)
	}

	activationURL := fmt.Sprintf(config.FrontendBaseURL+config.FrontendActivationRoute, activationToken)
	templateData := ActivationEmailTemplateData{
		RecipientEmail: user.Email,
		ActivationURL:  activationURL,
	}

	err = mailClient.Send([]string{user.Email}, "Welcome to Flagger!", "activation.txt", "activation.html", templateData)
	if err != nil {
		return fmt.Errorf("sendActivationMail failed to mailClient.SendMail for email %s: %w", user.Email, err)
	}

	return nil
}

// createAPIKey creates prefix, secret, and hashed key for API key.
func createAPIKey() (string, string, string, error) {
	prefix, err := utils.GenerateRandomString(8, true, true, true)
	if err != nil {
		return "", "", "", fmt.Errorf("createAPIKey failed to utils.GenerateRandomString: %w", err)
	}

	secretBytes, err := utils.GenerateRandomBytes(32)
	if err != nil {
		return "", "", "", fmt.Errorf("createAPIKey failed to utils.GenerateRandomBytes: %w", err)
	}

	secret := base64.StdEncoding.EncodeToString(secretBytes)

	rawKey := fmt.Sprintf("%s.%s", prefix, secret)

	hashedKey, err := hashPassword(rawKey)
	if err != nil {
		return "", "", "", fmt.Errorf("createAPIKey failed to hashPassword: %w", err)
	}

	return prefix, rawKey, hashedKey, nil
}

// parseAPIKey parses API key and returns prefix and secret if successful.
func parseAPIKey(key string) (string, string, bool) {
	return strings.Cut(key, ".")
}

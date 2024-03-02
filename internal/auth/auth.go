package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/alvii147/flagger-api/internal/templatesmanager"
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

// JWTType is a string representing type of JWT.
// Allowed strings are "access", "refresh", and "activation".
type JWTType string

const (
	JWTTypeAccess     JWTType = "access"
	JWTTypeRefresh    JWTType = "refresh"
	JWTTypeActivation JWTType = "activation"
)

// AuthContextKey is a string representing context keys.
type AuthContextKey string

// AuthContextKeyUserUUID is the key in context where User UUID is stored after authentication.
const AuthContextKeyUserUUID AuthContextKey = "userUUID"

// hashPassword hashes given password using a given hashing cost.
func hashPassword(password string, hashingCost int) (string, error) {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), hashingCost)
	if err != nil {
		return "", fmt.Errorf("hashPassword failed to bcrypt.GenerateFromPassword: %w", err)
	}
	hashedPassword := string(hashedPasswordBytes)

	return hashedPassword, nil
}

// createAuthJWT creates JWTs for User authentication of given type.
// Returns error when token type is not access or refresh.
func createAuthJWT(
	userUUID string,
	tokenType JWTType,
	secretKey string,
	accessLifetime time.Duration,
	refreshLifetime time.Duration,
) (string, error) {
	var lifetime time.Duration
	switch tokenType {
	case JWTTypeAccess:
		lifetime = accessLifetime
	case JWTTypeRefresh:
		lifetime = refreshLifetime
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
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("createAuthJWTWithType failed to token.SignedString for user.UUID %s of token type %s: %w", userUUID, tokenType, err)
	}

	return signedToken, nil
}

// validateAuthJWT validates JWT for User authentication,
// checks that the JWT is not expired,
// and returns parsed JWT claims.
func validateAuthJWT(token string, tokenType JWTType, secretKey string) (*api.AuthJWTClaims, bool) {
	claims := &api.AuthJWTClaims{}
	ok := true

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(secretKey), nil
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
		return nil, false
	}

	return claims, true
}

// createActivationJWTWithType creates JWT for User activation.
func createActivationJWT(userUUID string, secretKey string, lifetime time.Duration) (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.ActivationJWTClaims{
			Subject:   userUUID,
			TokenType: string(JWTTypeActivation),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(lifetime)),
			JWTID:     uuid.NewString(),
		},
	)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("createActivationJWT failed to token.SignedString for user.UUID %s of token type %s: %w", userUUID, JWTTypeActivation, err)
	}

	return signedToken, nil
}

// validateActivationJWT validates JWT for User activation using secret key,
// checks that the JWT is not expired,
// and returns parsed JWT claims.
func validateActivationJWT(token string, secretKey string) (*api.ActivationJWTClaims, bool) {
	claims := &api.ActivationJWTClaims{}
	ok := true

	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(secretKey), nil
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
func sendActivationMail(
	user *User,
	mailClient mailclient.Client,
	templatesManager templatesmanager.Manager,
	frontendBaseURL string,
	frontendActivationRoute string,
	secretKey string,
	jwtLifetime time.Duration,
) error {
	activationToken, err := createActivationJWT(user.UUID, secretKey, jwtLifetime)
	if err != nil {
		return fmt.Errorf("sendActivationMail failed to createActivationJWT: %w", err)
	}

	activationURL := fmt.Sprintf(frontendBaseURL+frontendActivationRoute, activationToken)
	tmplData := templatesmanager.ActivationEmailTemplateData{
		RecipientEmail: user.Email,
		ActivationURL:  activationURL,
	}

	textTmpl, htmlTmpl, err := templatesManager.Load("activation")
	if err != nil {
		return fmt.Errorf("sendActivationMail failed to templates.LoadTemplate %s: %w", "activation", err)
	}

	err = mailClient.Send([]string{user.Email}, "Welcome to Flagger!", textTmpl, htmlTmpl, tmplData)
	if err != nil {
		return fmt.Errorf("sendActivationMail failed to mailClient.SendMail for email %s: %w", user.Email, err)
	}

	return nil
}

// createAPIKey creates prefix, secret, and hashed key for API key.
func createAPIKey(hashingCost int) (string, string, string, error) {
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

	hashedKey, err := hashPassword(rawKey, hashingCost)
	if err != nil {
		return "", "", "", fmt.Errorf("createAPIKey failed to hashPassword: %w", err)
	}

	return prefix, rawKey, hashedKey, nil
}

// parseAPIKey parses API key and returns prefix and secret if successful.
func parseAPIKey(key string) (string, string, bool) {
	return strings.Cut(key, ".")
}

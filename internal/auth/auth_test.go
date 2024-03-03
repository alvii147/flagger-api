package auth_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/templatesmanager"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/alvii147/flagger-api/pkg/utils"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name        string
		hashingCost int
		wantErr     bool
	}{
		{
			name:        "Valid hashing cost",
			hashingCost: 14,
			wantErr:     false,
		},
		{
			name:        "Invalid hashing cost",
			hashingCost: 32,
			wantErr:     true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			password := testkit.GenerateFakePassword()
			hashedPassword, err := auth.HashPassword(password, testcase.hashingCost)

			if testcase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
				require.NoError(t, err)

				hashedPassword, err = auth.HashPassword(testkit.GenerateFakePassword(), 14)
				require.NoError(t, err)

				err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
				require.Error(t, err)
			}
		})
	}
}

func TestCreateAuthJWTSuccess(t *testing.T) {
	t.Parallel()

	accessLifetime := time.Minute
	refreshLifetime := time.Hour

	testcases := []struct {
		name         string
		tokenType    auth.JWTType
		wantLifetime time.Duration
	}{
		{
			name:         "Access token",
			tokenType:    auth.JWTTypeAccess,
			wantLifetime: time.Minute,
		},
		{
			name:         "Refresh token",
			tokenType:    auth.JWTTypeRefresh,
			wantLifetime: time.Hour,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			userUUID := uuid.NewString()
			token, err := auth.CreateAuthJWT(
				userUUID,
				testcase.tokenType,
				"deadbeef",
				accessLifetime,
				refreshLifetime,
			)
			require.NoError(t, err)

			claims := &api.AuthJWTClaims{}
			parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
				return []byte("deadbeef"), nil
			})
			require.NoError(t, err)

			require.NotNil(t, parsedToken)
			require.True(t, parsedToken.Valid)
			require.Equal(t, userUUID, claims.Subject)
			require.Equal(t, string(testcase.tokenType), claims.TokenType)

			testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), time.Time(claims.IssuedAt))
			testkit.RequireTimeAlmostEqual(t, time.Now().UTC().Add(testcase.wantLifetime), time.Time(claims.ExpiresAt))
		})
	}
}

func TestCreateAuthJWTInvalidType(t *testing.T) {
	t.Parallel()

	_, err := auth.CreateAuthJWT(
		uuid.NewString(),
		auth.JWTType("invalidtype"),
		"deadbeef",
		time.Minute,
		time.Hour,
	)
	require.Error(t, err)
}

func TestValidateAuthJWT(t *testing.T) {
	t.Parallel()

	userUUID := uuid.NewString()
	jti := uuid.NewString()
	now := time.Now().UTC()
	oneDayAgo := now.Add(-24 * time.Hour)
	validSecretKey := "deadbeef"

	validAccessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(auth.JWTTypeAccess),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	validRefreshToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(auth.JWTTypeRefresh),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	tokenOfInvalidType, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string("invalidtype"),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	expiredToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(auth.JWTTypeAccess),
			IssuedAt:  utils.JSONTimeStamp(oneDayAgo),
			ExpiresAt: utils.JSONTimeStamp(oneDayAgo.Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	tokenWithInvalidClaim, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&struct {
			InvalidClaim string `json:"invalid_claim"`
			jwt.StandardClaims
		}{},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	testcases := []struct {
		name      string
		token     string
		tokenType auth.JWTType
		secretKey string
		wantOk    bool
	}{
		{
			name:      "Valid access token",
			token:     validAccessToken,
			tokenType: auth.JWTTypeAccess,
			secretKey: validSecretKey,
			wantOk:    true,
		},
		{
			name:      "Valid refresh token",
			token:     validRefreshToken,
			tokenType: auth.JWTTypeRefresh,
			secretKey: validSecretKey,
			wantOk:    true,
		},
		{
			name:      "Invalid secret key",
			token:     validAccessToken,
			tokenType: auth.JWTTypeAccess,
			secretKey: "invalidsecretkey",
			wantOk:    false,
		},
		{
			name:      "Token of invalid type",
			token:     tokenOfInvalidType,
			tokenType: auth.JWTTypeAccess,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Invalid token",
			token:     "ed0730889507fdb8549acfcd31548ee5",
			tokenType: auth.JWTTypeAccess,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Expired token",
			token:     expiredToken,
			tokenType: auth.JWTTypeAccess,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Token with invalid claim",
			token:     tokenWithInvalidClaim,
			tokenType: auth.JWTTypeAccess,
			secretKey: validSecretKey,
			wantOk:    false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			claims, ok := auth.ValidateAuthJWT(testcase.token, testcase.tokenType, testcase.secretKey)
			require.Equal(t, testcase.wantOk, ok)

			if testcase.wantOk {
				require.Equal(t, userUUID, claims.Subject)
				require.Equal(t, string(testcase.tokenType), claims.TokenType)
			}
		})
	}
}

func TestCreateActivationJWTSuccess(t *testing.T) {
	t.Parallel()

	userUUID := uuid.NewString()
	secretKey := "deadbeef"
	lifetime := time.Hour
	token, err := auth.CreateActivationJWT(userUUID, secretKey, lifetime)
	require.NoError(t, err)

	claims := &api.ActivationJWTClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedToken)
	require.True(t, parsedToken.Valid)
	require.Equal(t, userUUID, claims.Subject)
	require.Equal(t, string(auth.JWTTypeActivation), claims.TokenType)

	testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), time.Time(claims.IssuedAt))
	testkit.RequireTimeAlmostEqual(t, time.Now().UTC().Add(lifetime), time.Time(claims.ExpiresAt))
}

func TestValidateActivationJWT(t *testing.T) {
	t.Parallel()

	userUUID := uuid.NewString()
	jti := uuid.NewString()
	now := time.Now().UTC()
	oneDayAgo := now.Add(-24 * time.Hour)
	validSecretKey := "deadbeef"

	validToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.ActivationJWTClaims{
			Subject:   userUUID,
			TokenType: string(auth.JWTTypeActivation),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	tokenOfInvalidType, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.ActivationJWTClaims{
			Subject:   userUUID,
			TokenType: string("invalidtype"),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	expiredToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.ActivationJWTClaims{
			Subject:   userUUID,
			TokenType: string(auth.JWTTypeActivation),
			IssuedAt:  utils.JSONTimeStamp(oneDayAgo),
			ExpiresAt: utils.JSONTimeStamp(oneDayAgo.Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	tokenWithInvalidClaim, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&struct {
			InvalidClaim string `json:"invalid_claim"`
			jwt.StandardClaims
		}{},
	).SignedString([]byte(validSecretKey))
	require.NoError(t, err)

	testcases := []struct {
		name      string
		token     string
		secretKey string
		wantOk    bool
	}{
		{
			name:      "Valid token of correct type",
			token:     validToken,
			secretKey: validSecretKey,
			wantOk:    true,
		},
		{
			name:      "Token of incorrect type",
			token:     tokenOfInvalidType,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Invalid token",
			token:     "ed0730889507fdb8549acfcd31548ee5",
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Expired token",
			token:     expiredToken,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Token with invalid claim",
			token:     tokenWithInvalidClaim,
			secretKey: validSecretKey,
			wantOk:    false,
		},
		{
			name:      "Incorrect secret key",
			token:     validToken,
			secretKey: "incorrectsecretkey",
			wantOk:    false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			claims, ok := auth.ValidateActivationJWT(testcase.token, testcase.secretKey)
			require.Equal(t, testcase.wantOk, ok)

			if testcase.wantOk {
				require.Equal(t, userUUID, claims.Subject)
				require.Equal(t, string(auth.JWTTypeActivation), claims.TokenType)
			}
		})
	}
}

func TestSendActivationMail(t *testing.T) {
	t.Parallel()

	user := &auth.User{
		UUID:        uuid.NewString(),
		Email:       testkit.GenerateFakeEmail(),
		Password:    testkitinternal.MustHashPassword(testkit.GenerateFakePassword()),
		FirstName:   testkit.MustGenerateRandomString(8, true, true, false),
		LastName:    testkit.MustGenerateRandomString(8, true, true, false),
		IsActive:    true,
		IsSuperUser: false,
	}

	mailClient := mailclient.NewInMemClient("support@flagger.com")
	mailCount := len(mailClient.Logs)
	tmplManager := templatesmanager.NewManager()
	frontendBaseURL := "http://localhost:3000"
	frontendActivationRoute := "/signup/activate/%s"
	secretKey := "deadbeef"
	lifetime := time.Hour
	err := auth.SendActivationMail(
		user,
		mailClient,
		tmplManager,
		frontendBaseURL,
		frontendActivationRoute,
		secretKey,
		lifetime,
	)
	require.NoError(t, err)
	require.Len(t, mailClient.Logs, mailCount+1)

	lastMail := mailClient.Logs[len(mailClient.Logs)-1]
	require.Equal(t, []string{user.Email}, lastMail.To)
	require.Equal(t, "Welcome to Flagger!", lastMail.Subject)
	testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), lastMail.SentAt)

	mailMessage := string(lastMail.Message)
	require.Contains(t, mailMessage, "Welcome to Flagger!")
	require.Contains(t, mailMessage, "Flagger - Activate Your Account")

	pattern := fmt.Sprintf(frontendBaseURL+frontendActivationRoute, `(\S+)`)
	r, err := regexp.Compile(pattern)
	require.NoError(t, err)

	matches := r.FindStringSubmatch(mailMessage)
	require.Len(t, matches, 2)

	activationToken := matches[1]
	claims := &api.ActivationJWTClaims{}
	parsedToken, err := jwt.ParseWithClaims(activationToken, claims, func(t *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})
	require.NoError(t, err)

	require.NotNil(t, parsedToken)
	require.True(t, parsedToken.Valid)
	require.Equal(t, user.UUID, claims.Subject)
	require.Equal(t, string(auth.JWTTypeActivation), claims.TokenType)

	testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), time.Time(claims.IssuedAt))
	testkit.RequireTimeAlmostEqual(t, time.Now().UTC().Add(lifetime), time.Time(claims.ExpiresAt))
}

func TestCreateAPIKey(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name        string
		hashingCost int
		wantErr     bool
	}{
		{
			name:        "Valid hashing cost",
			hashingCost: 14,
			wantErr:     false,
		},
		{
			name:        "Invalid hashing cost",
			hashingCost: 32,
			wantErr:     true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			prefix, rawKey, hashedKey, err := auth.CreateAPIKey(testcase.hashingCost)
			if testcase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				err = bcrypt.CompareHashAndPassword([]byte(hashedKey), []byte(rawKey))
				require.NoError(t, err)

				r := regexp.MustCompile(`^(\S+)\.(\S+)$`)
				matches := r.FindStringSubmatch(rawKey)

				require.Len(t, matches, 3)
				require.Equal(t, prefix, matches[1])
			}
		})
	}
}

func TestParseAPIKey(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		key        string
		wantPrefix string
		wantSecret string
		wantOk     bool
	}{
		{
			name:       "Valid API key",
			key:        "TqxlYSSQ.Yj2j1jyAMC5407Nctsl51K7E8sOIPqYXn28SqT5Gnfg=",
			wantPrefix: "TqxlYSSQ",
			wantSecret: "Yj2j1jyAMC5407Nctsl51K7E8sOIPqYXn28SqT5Gnfg=",
			wantOk:     true,
		},
		{
			name:       "Invalid API key",
			key:        "DeAdBeEf",
			wantPrefix: "",
			wantSecret: "",
			wantOk:     false,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			prefix, secret, ok := auth.ParseAPIKey(testcase.key)
			require.Equal(t, testcase.wantOk, ok)

			if testcase.wantOk {
				require.Equal(t, testcase.wantPrefix, prefix)
				require.Equal(t, testcase.wantSecret, secret)
			}
		})
	}
}

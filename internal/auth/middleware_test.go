package auth_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/internal/templatesmanager"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/alvii147/flagger-api/pkg/utils"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJWTAuthMiddleware(t *testing.T) {
	t.Parallel()

	userUUID := uuid.NewString()
	jti := uuid.NewString()
	now := time.Now().UTC()
	oneDayAgo := now.Add(-24 * time.Hour)
	secretKey := "deadbeef"

	validAccessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.AuthJWTClaims{
			Subject:   userUUID,
			TokenType: string(auth.JWTTypeAccess),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(secretKey))
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
	).SignedString([]byte(secretKey))
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
	).SignedString([]byte(secretKey))
	require.NoError(t, err)

	tokenWithInvalidClaim, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&struct {
			InvalidClaim string `json:"invalid_claim"`
			jwt.StandardClaims
		}{},
	).SignedString([]byte(secretKey))
	require.NoError(t, err)

	validResponse := map[string]any{
		"email":      testkit.GenerateFakeEmail(),
		"first_name": testkit.MustGenerateRandomString(8, true, true, false),
		"last_name":  testkit.MustGenerateRandomString(8, true, true, false),
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
	validStatusCode := 200

	testcases := []struct {
		name           string
		wantNextCall   bool
		wantErr        bool
		wantErrCode    string
		wantStatusCode int
		setAuthHeader  bool
		authHeader     string
	}{
		{
			name:           "Authentication with valid JWT is successful",
			wantNextCall:   true,
			wantErr:        false,
			wantErrCode:    "",
			wantStatusCode: http.StatusOK,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("Bearer %s", validAccessToken),
		},
		{
			name:           "Authentication with no authorization header is unauthorized",
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  false,
			authHeader:     "",
		},
		{
			name:           "Authentication with invalid authentication type is unauthorized",
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("Invalidauthtype %s", validAccessToken),
		},
		{
			name:           "Authentication with invalid JWT is unauthorized",
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     "Bearer ed0730889507fdb8549acfcd31548ee5",
		},
		{
			name:           "Authentication with expired JWT is unauthorized",
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("Bearer %s", expiredToken),
		},
		{
			name:           "Authentication with valid JWT of invalid type is unauthorized",
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("Bearer %s", tokenOfInvalidType),
		},
		{
			name:           "Authentication with JWT with invalid claim is unauthorized",
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("Bearer %s", tokenWithInvalidClaim),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			nextCallCount := 0
			var next httputils.HandlerFunc = func(w *httputils.ResponseWriter, r *http.Request) {
				require.Equal(t, userUUID, r.Context().Value(auth.AuthContextKeyUserUUID))
				w.WriteJSON(validResponse, validStatusCode)
				nextCallCount++
			}

			rec := httptest.NewRecorder()
			w := &httputils.ResponseWriter{
				ResponseWriter: rec,
				StatusCode:     -1,
			}
			r := httptest.NewRequest(http.MethodGet, "/auth/users/me", http.NoBody)

			if testcase.setAuthHeader {
				r.Header.Set("Authorization", testcase.authHeader)
			}

			auth.JWTAuthMiddleware(next, secretKey)(w, r)

			result := rec.Result()
			t.Cleanup(func() {
				err := result.Body.Close()
				require.NoError(t, err)
			})

			responseBodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			var responseBody map[string]any
			err = json.Unmarshal(responseBodyBytes, &responseBody)
			require.NoError(t, err)

			require.Equal(t, testcase.wantStatusCode, result.StatusCode)

			wantNextCallCount := 0
			if testcase.wantNextCall {
				wantNextCallCount = 1
			}

			require.Equal(t, wantNextCallCount, nextCallCount)

			if testcase.wantErr {
				errCode, ok := responseBody["code"]
				require.True(t, ok)
				require.Equal(t, testcase.wantErrCode, errCode)
			} else {
				require.Equal(t, validResponse, responseBody)
			}
		})
	}
}

func TestAPIKeyAuthMiddleware(t *testing.T) {
	t.Parallel()

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	config, err := env.NewConfig()
	require.NoError(t, err)

	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	_, _, logger := testkit.CreateTestLogger()
	mailClient := mailclient.NewInMemClient("support@flagger.com")
	tmplManager := templatesmanager.NewManager()
	repo := auth.NewRepository()
	svc := auth.NewService(config, dbPool, logger, mailClient, tmplManager, repo)

	_, validAPIKey := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	validResponse := map[string]any{
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
	validStatusCode := 200

	testcases := []struct {
		name           string
		wantNextCall   bool
		wantErr        bool
		wantErrCode    string
		wantStatusCode int
		setAuthHeader  bool
		authHeader     string
	}{
		{
			name:           "Authentication with valid API key is successful",
			wantNextCall:   true,
			wantErr:        false,
			wantErrCode:    "",
			wantStatusCode: http.StatusOK,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("X-API-Key %s", validAPIKey),
		},
		{
			name:           "Authentication with no authorization header is unauthorized",
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  false,
			authHeader:     "",
		},
		{
			name:           "Authentication with invalid authentication type is unauthorized",
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     fmt.Sprintf("Invalidauthtype %s", validAPIKey),
		},
		{
			name:           "Authentication with invalid API key is unauthorized",
			wantNextCall:   false,
			wantErr:        true,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantStatusCode: http.StatusUnauthorized,
			setAuthHeader:  true,
			authHeader:     "X-API-Key DQGDG0Al.xoentiX0xPztDX6ybl6SNfveoCAT/M9Y6oXy96uMCGg=",
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			nextCallCount := 0
			var next httputils.HandlerFunc = func(w *httputils.ResponseWriter, r *http.Request) {
				require.Equal(t, user.UUID, r.Context().Value(auth.AuthContextKeyUserUUID))
				w.WriteJSON(validResponse, validStatusCode)
				nextCallCount++
			}

			rec := httptest.NewRecorder()
			w := &httputils.ResponseWriter{
				ResponseWriter: rec,
				StatusCode:     -1,
			}
			r := httptest.NewRequest(http.MethodGet, "/auth/users/me", http.NoBody)

			if testcase.setAuthHeader {
				r.Header.Set("Authorization", testcase.authHeader)
			}

			auth.APIKeyAuthMiddleware(next, svc)(w, r)

			result := rec.Result()
			t.Cleanup(func() {
				err := result.Body.Close()
				require.NoError(t, err)
			})

			responseBodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			var responseBody map[string]any
			err = json.Unmarshal(responseBodyBytes, &responseBody)
			require.NoError(t, err)

			require.Equal(t, testcase.wantStatusCode, result.StatusCode)

			wantNextCallCount := 0
			if testcase.wantNextCall {
				wantNextCallCount = 1
			}

			require.Equal(t, wantNextCallCount, nextCallCount)

			if testcase.wantErr {
				errCode, ok := responseBody["code"]
				require.True(t, ok)
				require.Equal(t, testcase.wantErrCode, errCode)
			} else {
				require.Equal(t, validResponse, responseBody)
			}
		})
	}
}

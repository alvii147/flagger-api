package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/internal/server"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/alvii147/flagger-api/pkg/utils"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestGetAPIKeyIDParam(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name         string
		pathValues   map[string]string
		wantAPIKeyID int
		wantErr      bool
	}{
		{
			name: "Valid API key ID",
			pathValues: map[string]string{
				"id": "42",
			},
			wantAPIKeyID: 42,
			wantErr:      false,
		},
		{
			name: "No API key ID",
			pathValues: map[string]string{
				"dead": "beef",
			},
			wantAPIKeyID: 0,
			wantErr:      true,
		},
		{
			name: "Invalid API key ID",
			pathValues: map[string]string{
				"id": "deadbeef",
			},
			wantAPIKeyID: 0,
			wantErr:      true,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req := &http.Request{}
			for name, value := range testcase.pathValues {
				req.SetPathValue(name, value)
			}

			apiKeyID, err := server.GetAPIKeyIDParam(req)
			if testcase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testcase.wantAPIKeyID, apiKeyID)
			}
		})
	}
}

func TestHandleCreateUser(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	existingUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	email := testkit.GenerateFakeEmail()
	password := testkit.GenerateFakePassword()
	firstName := testkit.MustGenerateRandomString(8, true, true, false)
	lastName := testkit.MustGenerateRandomString(8, true, true, false)

	testcases := []struct {
		name           string
		requestBody    string
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		{
			name: "Valid request",
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, email, password, firstName, lastName),
			wantStatusCode: http.StatusCreated,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Existing email",
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, existingUser.Email, password, firstName, lastName),
			wantStatusCode: http.StatusConflict,
			wantErrCode:    api.ErrCodeResourceExists,
			wantErrDetail:  api.ErrDetailUserExists,
		},
		{
			name: "Invalid email",
			requestBody: fmt.Sprintf(`
				{
					"email": "1nv4l1d3m41l",
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, password, firstName, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Missing email",
			requestBody: fmt.Sprintf(`
				{
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, password, firstName, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Empty email",
			requestBody: fmt.Sprintf(`
				{
					"email": "",
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, password, firstName, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Missing password",
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, email, firstName, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Empty password",
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, email, firstName, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Missing first name",
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s",
					"last_name": "%s"
				}
			`, email, password, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Empty first name",
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s",
					"first_name": "",
					"last_name": "%s"
				}
			`, email, password, lastName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Missing last name",
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s",
					"first_name": "%s",
				}
			`, email, password, firstName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Empty last name",
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s",
					"first_name": "%s",
					"last_name": ""
				}
			`, email, password, firstName),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/auth/users",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

			userCreatedAt := time.Now().UTC()
			res, err := httpClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := res.Body.Close()
				require.NoError(t, err)
			})

			require.Equal(t, testcase.wantStatusCode, res.StatusCode)

			if httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				var createUserResp api.CreateUserResponse
				err = json.NewDecoder(res.Body).Decode(&createUserResp)
				require.NoError(t, err)

				require.Equal(t, email, createUserResp.Email)
				require.Equal(t, firstName, createUserResp.FirstName)
				require.Equal(t, lastName, createUserResp.LastName)
				testkit.RequireTimeAlmostEqual(t, userCreatedAt, createUserResp.CreatedAt)
			} else {
				var errResp api.ErrorResponse
				err = json.NewDecoder(res.Body).Decode(&errResp)
				require.NoError(t, err)

				require.Equal(t, testcase.wantErrCode, errResp.Code)
				require.Equal(t, testcase.wantErrDetail, errResp.Detail)
			}
		})
	}
}

func TestHandleActivateUser(t *testing.T) {
	t.Parallel()

	config, err := env.NewConfig()
	require.NoError(t, err)

	httpClient := httputils.NewHTTPClient(nil)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	now := time.Now().UTC()
	activeUserToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.ActivationJWTClaims{
			Subject:   activeUser.UUID,
			TokenType: string(auth.JWTTypeActivation),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Duration(config.ActivationLifetime * int64(time.Minute)))),
			JWTID:     uuid.NewString(),
		},
	).SignedString([]byte(config.SecretKey))
	require.NoError(t, err)

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})

	inactiveUserToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.ActivationJWTClaims{
			Subject:   inactiveUser.UUID,
			TokenType: string(auth.JWTTypeActivation),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Duration(config.ActivationLifetime * int64(time.Minute)))),
			JWTID:     uuid.NewString(),
		},
	).SignedString([]byte(config.SecretKey))
	require.NoError(t, err)

	testcases := []struct {
		name           string
		requestBody    string
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		{
			name: "Inactive user",
			requestBody: fmt.Sprintf(`
				{
					"token": "%s"
				}
			`, inactiveUserToken),
			wantStatusCode: http.StatusOK,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Already active user",
			requestBody: fmt.Sprintf(`
				{
					"token": "%s"
				}
			`, activeUserToken),
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailUserNotFound,
		},
		{
			name: "Invalid token",
			requestBody: `
				{
					"token": "1nV4LiDT0k3n"
				}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidToken,
		},
		{
			name: "Missing token",
			requestBody: `
				{}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/auth/users/activate",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

			res, err := httpClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := res.Body.Close()
				require.NoError(t, err)
			})

			require.Equal(t, testcase.wantStatusCode, res.StatusCode)

			if !httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				var errResp api.ErrorResponse
				err = json.NewDecoder(res.Body).Decode(&errResp)
				require.NoError(t, err)

				require.Equal(t, testcase.wantErrCode, errResp.Code)
				require.Equal(t, testcase.wantErrDetail, errResp.Detail)
			}
		})
	}
}

func TestHandleGetUserMe(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	activeUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(activeUser.UUID)
	_, activeUserRawAPIKey := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, nil)

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})
	inactiveUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(inactiveUser.UUID)
	_, inactiveUserRawAPIKey := testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)

	testcases := []struct {
		name           string
		path           string
		headers        map[string]string
		user           *auth.User
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		{
			name: "Get active user using JWT",
			path: "/auth/users/me",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			user:           activeUser,
			wantStatusCode: http.StatusOK,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Get active user using API key",
			path: "/api/auth/users/me",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("X-API-Key %s", activeUserRawAPIKey),
			},
			user:           activeUser,
			wantStatusCode: http.StatusOK,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Get inactive user using JWT",
			path: "/auth/users/me",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			user:           inactiveUser,
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailUserNotFound,
		},
		{
			name: "Get inactive user using API key",
			path: "/api/auth/users/me",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("X-API-Key %s", inactiveUserRawAPIKey),
			},
			user:           inactiveUser,
			wantStatusCode: http.StatusUnauthorized,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantErrDetail:  api.ErrDetailInvalidToken,
		},
		{
			name:           "Get user without authentication",
			path:           "/auth/users/me",
			headers:        map[string]string{},
			user:           nil,
			wantStatusCode: http.StatusUnauthorized,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantErrDetail:  api.ErrDetailMissingCredentials,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, TestServerURL+testcase.path, http.NoBody)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

			res, err := httpClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := res.Body.Close()
				require.NoError(t, err)
			})

			require.Equal(t, testcase.wantStatusCode, res.StatusCode)

			if httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				var getUserMeResp api.GetUserMeResponse
				err = json.NewDecoder(res.Body).Decode(&getUserMeResp)
				require.NoError(t, err)

				require.Equal(t, testcase.user.UUID, getUserMeResp.UUID)
				require.Equal(t, testcase.user.Email, getUserMeResp.Email)
				require.Equal(t, testcase.user.FirstName, getUserMeResp.FirstName)
				require.Equal(t, testcase.user.LastName, getUserMeResp.LastName)
				testkit.RequireTimeAlmostEqual(t, testcase.user.CreatedAt, getUserMeResp.CreatedAt)
			} else {
				var errResp api.ErrorResponse
				err = json.NewDecoder(res.Body).Decode(&errResp)
				require.NoError(t, err)

				require.Equal(t, testcase.wantErrCode, errResp.Code)
				require.Equal(t, testcase.wantErrDetail, errResp.Detail)
			}
		})
	}
}

func TestHandleCreateJWT(t *testing.T) {
	t.Parallel()

	config, err := env.NewConfig()
	require.NoError(t, err)

	httpClient := httputils.NewHTTPClient(nil)

	user, password := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})

	testcases := []struct {
		name           string
		requestBody    string
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		{
			name: "Valid request",
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "%s"
				}
			`, user.Email, password),
			wantStatusCode: http.StatusCreated,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Incorrect password",
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": "1nc0rr3CTP455w0Rd"
				}
			`, user.Email),
			wantStatusCode: http.StatusUnauthorized,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantErrDetail:  api.ErrDetailInvalidEmailOrPassword,
		},
		{
			name: "Invalid email",
			requestBody: fmt.Sprintf(`
				{
					"email": "1nv4l1d3M4iL",
					"password": "%s"
				}
			`, password),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Missing email",
			requestBody: fmt.Sprintf(`
				{
					"password": "%s"
				}
			`, password),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Empty email",
			requestBody: fmt.Sprintf(`
				{
					"email": "",
					"password": "%s"
				}
			`, password),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Missing password",
			requestBody: fmt.Sprintf(`
				{
					"email": "%s"
				}
			`, user.Email),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Empty password",
			requestBody: fmt.Sprintf(`
				{
					"email": "%s",
					"password": ""
				}
			`, user.Email),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/auth/tokens",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

			res, err := httpClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := res.Body.Close()
				require.NoError(t, err)
			})

			require.Equal(t, testcase.wantStatusCode, res.StatusCode)

			if httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				var createTokenResp api.CreateTokenResponse
				err = json.NewDecoder(res.Body).Decode(&createTokenResp)
				require.NoError(t, err)

				accessClaims := &api.AuthJWTClaims{}
				parsedAccessToken, err := jwt.ParseWithClaims(createTokenResp.Access, accessClaims, func(t *jwt.Token) (any, error) {
					return []byte(config.SecretKey), nil
				})
				require.NoError(t, err)

				require.NotNil(t, parsedAccessToken)
				require.True(t, parsedAccessToken.Valid)
				require.Equal(t, user.UUID, accessClaims.Subject)
				require.Equal(t, string(auth.JWTTypeAccess), accessClaims.TokenType)

				testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), time.Time(accessClaims.IssuedAt))
				testkit.RequireTimeAlmostEqual(t, time.Now().UTC().Add(time.Duration(config.AuthAccessLifetime*int64(time.Minute))), time.Time(accessClaims.ExpiresAt))

				refreshClaims := &api.AuthJWTClaims{}
				parsedRefreshToken, err := jwt.ParseWithClaims(createTokenResp.Refresh, refreshClaims, func(t *jwt.Token) (any, error) {
					return []byte(config.SecretKey), nil
				})
				require.NoError(t, err)

				require.NotNil(t, parsedRefreshToken)
				require.True(t, parsedRefreshToken.Valid)
				require.Equal(t, user.UUID, refreshClaims.Subject)
				require.Equal(t, string(auth.JWTTypeRefresh), refreshClaims.TokenType)

				testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), time.Time(refreshClaims.IssuedAt))
				testkit.RequireTimeAlmostEqual(t, time.Now().UTC().Add(time.Duration(config.AuthRefreshLifetime*int64(time.Minute))), time.Time(refreshClaims.ExpiresAt))
			} else {
				var errResp api.ErrorResponse
				err = json.NewDecoder(res.Body).Decode(&errResp)
				require.NoError(t, err)

				require.Equal(t, testcase.wantErrCode, errResp.Code)
				require.Equal(t, testcase.wantErrDetail, errResp.Detail)
			}
		})
	}
}

func TestHandleRefreshJWT(t *testing.T) {
	t.Parallel()

	config, err := env.NewConfig()
	require.NoError(t, err)

	httpClient := httputils.NewHTTPClient(nil)

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	_, validRefreshToken := testkitinternal.MustCreateUserAuthJWTs(user.UUID)

	testcases := []struct {
		name           string
		requestBody    string
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		{
			name: "Valid request",
			requestBody: fmt.Sprintf(`
				{
					"refresh": "%s"
				}
			`, validRefreshToken),
			wantStatusCode: http.StatusCreated,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Invalid token",
			requestBody: `
				{
					"refresh": "iNv4liDT0k3N"
				}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Missing token",
			requestBody: `
				{}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/auth/tokens/refresh",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

			res, err := httpClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := res.Body.Close()
				require.NoError(t, err)
			})

			require.Equal(t, testcase.wantStatusCode, res.StatusCode)

			if httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				var refreshTokenResp api.RefreshTokenResponse
				err = json.NewDecoder(res.Body).Decode(&refreshTokenResp)
				require.NoError(t, err)

				claims := &api.AuthJWTClaims{}
				parsedToken, err := jwt.ParseWithClaims(refreshTokenResp.Access, claims, func(t *jwt.Token) (any, error) {
					return []byte(config.SecretKey), nil
				})
				require.NoError(t, err)

				require.NotNil(t, parsedToken)
				require.True(t, parsedToken.Valid)
				require.Equal(t, user.UUID, claims.Subject)
				require.Equal(t, string(auth.JWTTypeAccess), claims.TokenType)

				testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), time.Time(claims.IssuedAt))
				testkit.RequireTimeAlmostEqual(t, time.Now().UTC().Add(time.Duration(config.AuthAccessLifetime*int64(time.Minute))), time.Time(claims.ExpiresAt))
			} else {
				var errResp api.ErrorResponse
				err = json.NewDecoder(res.Body).Decode(&errResp)
				require.NoError(t, err)

				require.Equal(t, testcase.wantErrCode, errResp.Code)
				require.Equal(t, testcase.wantErrDetail, errResp.Detail)
			}
		})
	}
}

func TestHandleCreateAPIKey(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	userAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(user.UUID)

	expirationDate := time.Date(2038, 1, 19, 3, 14, 8, 0, time.UTC)
	expirationDateString := "2038-01-19T03:14:08Z"

	testcases := []struct {
		name               string
		headers            map[string]string
		requestBody        string
		wantStatusCode     int
		wantAPIKeyName     string
		wantExpirationDate bool
		wantErrCode        string
		wantErrDetail      string
	}{
		{
			name: "Valid request with no expiration date",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: `
				{
					"name": "My non-expiring API key"
				}
			`,
			wantStatusCode:     http.StatusCreated,
			wantAPIKeyName:     "My non-expiring API key",
			wantExpirationDate: false,
			wantErrCode:        "",
			wantErrDetail:      "",
		},
		{
			name: "Valid request with expiration date",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: fmt.Sprintf(`
				{
					"name": "My expiring API key",
					"expires_at": "%s"
				}
			`, expirationDateString),
			wantStatusCode:     http.StatusCreated,
			wantAPIKeyName:     "My expiring API key",
			wantExpirationDate: true,
			wantErrCode:        "",
			wantErrDetail:      "",
		},
		{
			name: "Name missing",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: fmt.Sprintf(`
				{
					"expires_at": "%s"
				}
			`, expirationDateString),
			wantStatusCode:     http.StatusBadRequest,
			wantAPIKeyName:     "My nameless API key",
			wantExpirationDate: true,
			wantErrCode:        api.ErrCodeInvalidRequest,
			wantErrDetail:      api.ErrDetailInvalidRequestData,
		},
		{
			name: "Invalid expiration date",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: `
				{
					"name": "My invalidly-expiring API key",
					"expires_at": "1nv4l1dd4t3"
				}
			`,
			wantStatusCode:     http.StatusBadRequest,
			wantAPIKeyName:     "My invalidly-expiring API key",
			wantExpirationDate: false,
			wantErrCode:        api.ErrCodeInvalidRequest,
			wantErrDetail:      api.ErrDetailInvalidRequestData,
		},
		{
			name:    "Unauthenticated request",
			headers: map[string]string{},
			requestBody: `
				{
					"name": "My unauthenticated API key"
				}
			`,
			wantStatusCode:     http.StatusUnauthorized,
			wantAPIKeyName:     "My unauthenticated API key",
			wantExpirationDate: false,
			wantErrCode:        api.ErrCodeMissingCredentials,
			wantErrDetail:      api.ErrDetailMissingCredentials,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/auth/api-keys",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

			apiKeyCreatedAt := time.Now().UTC()
			res, err := httpClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := res.Body.Close()
				require.NoError(t, err)
			})

			require.Equal(t, testcase.wantStatusCode, res.StatusCode)

			if httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				var createAPIKeyResp api.CreateAPIKeyResponse
				err = json.NewDecoder(res.Body).Decode(&createAPIKeyResp)
				require.NoError(t, err)

				require.Equal(t, user.UUID, createAPIKeyResp.UserUUID)
				require.Equal(t, testcase.wantAPIKeyName, createAPIKeyResp.Name)
				testkit.RequireTimeAlmostEqual(t, apiKeyCreatedAt, createAPIKeyResp.CreatedAt)
				testkit.RequirePGTimestampAlmostEqual(t, pgtype.Timestamp{
					Time:  expirationDate,
					Valid: testcase.wantExpirationDate,
				}, createAPIKeyResp.ExpiresAt)
			} else {
				var errResp api.ErrorResponse
				err = json.NewDecoder(res.Body).Decode(&errResp)
				require.NoError(t, err)

				require.Equal(t, testcase.wantErrCode, errResp.Code)
				require.Equal(t, testcase.wantErrDetail, errResp.Detail)
			}
		})
	}
}

func TestHandleListAPIKeys(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	activeUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(activeUser.UUID)
	activeUserAPIKey, activeUserRawAPIKey := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, nil)

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})
	inactiveUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(inactiveUser.UUID)
	testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)

	testcases := []struct {
		name                  string
		headers               map[string]string
		wantStatusCode        int
		wantAPIKeysInResponse bool
		wantErrCode           string
		wantErrDetail         string
	}{
		{
			name: "List API keys for active user",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			wantStatusCode:        http.StatusOK,
			wantAPIKeysInResponse: true,
			wantErrCode:           "",
			wantErrDetail:         "",
		},
		{
			name: "List API keys for inactive user",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			wantStatusCode:        http.StatusOK,
			wantAPIKeysInResponse: false,
			wantErrCode:           api.ErrCodeResourceNotFound,
			wantErrDetail:         api.ErrDetailUserNotFound,
		},
		{
			name:                  "List API keys without authentication",
			headers:               map[string]string{},
			wantStatusCode:        http.StatusUnauthorized,
			wantAPIKeysInResponse: false,
			wantErrCode:           api.ErrCodeMissingCredentials,
			wantErrDetail:         api.ErrDetailMissingCredentials,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodGet,
				TestServerURL+"/auth/api-keys",
				http.NoBody,
			)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

			res, err := httpClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := res.Body.Close()
				require.NoError(t, err)
			})

			require.Equal(t, testcase.wantStatusCode, res.StatusCode)

			if httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				var listAPIKeysResp api.ListAPIKeysResponse
				err = json.NewDecoder(res.Body).Decode(&listAPIKeysResp)
				require.NoError(t, err)

				if testcase.wantAPIKeysInResponse {
					require.Len(t, listAPIKeysResp.Keys, 1)
					require.Equal(t, activeUserAPIKey.ID, listAPIKeysResp.Keys[0].ID)
					require.Equal(t, activeUser.UUID, listAPIKeysResp.Keys[0].UserUUID)
					require.True(t, strings.HasPrefix(activeUserRawAPIKey, listAPIKeysResp.Keys[0].Prefix))
					require.Equal(t, activeUserAPIKey.Name, listAPIKeysResp.Keys[0].Name)
					testkit.RequireTimeAlmostEqual(t, activeUserAPIKey.CreatedAt, listAPIKeysResp.Keys[0].CreatedAt)
					testkit.RequirePGTimestampAlmostEqual(t, activeUserAPIKey.ExpiresAt, listAPIKeysResp.Keys[0].ExpiresAt)
				} else {
					require.Len(t, listAPIKeysResp.Keys, 0)
				}
			} else {
				var errResp api.ErrorResponse
				err = json.NewDecoder(res.Body).Decode(&errResp)
				require.NoError(t, err)

				require.Equal(t, testcase.wantErrCode, errResp.Code)
				require.Equal(t, testcase.wantErrDetail, errResp.Detail)
			}
		})
	}
}

func TestHandleDeleteAPIKey(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	activeUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(activeUser.UUID)
	activeUserAPIKey1, _ := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, func(k *auth.APIKey) {
		k.Name = "MyAPIKey1"
	})
	activeUserAPIKey2, _ := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, func(k *auth.APIKey) {
		k.Name = "MyAPIKey2"
	})

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})
	inactiveUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(inactiveUser.UUID)
	inactiveUserAPIKey, _ := testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)

	testcases := []struct {
		name           string
		path           string
		headers        map[string]string
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		{
			name: "Delete API key for active user",
			path: fmt.Sprintf("/auth/api-keys/%d", activeUserAPIKey1.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			wantStatusCode: http.StatusNoContent,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Delete API key for inactive user",
			path: fmt.Sprintf("/auth/api-keys/%d", inactiveUserAPIKey.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailAPIKeyNotFound,
		},
		{
			name: "Delete API key for another user",
			path: fmt.Sprintf("/auth/api-keys/%d", activeUserAPIKey2.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailAPIKeyNotFound,
		},
		{
			name: "Delete non-existent API key",
			path: fmt.Sprintf("/auth/api-keys/%d", 314159),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailAPIKeyNotFound,
		},
		{
			name:           "Delete API key without authentication",
			path:           fmt.Sprintf("/auth/api-keys/%d", activeUserAPIKey2.ID),
			headers:        map[string]string{},
			wantStatusCode: http.StatusUnauthorized,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantErrDetail:  api.ErrDetailMissingCredentials,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodDelete, TestServerURL+testcase.path, http.NoBody)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

			res, err := httpClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := res.Body.Close()
				require.NoError(t, err)
			})

			require.Equal(t, testcase.wantStatusCode, res.StatusCode)
			if !httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				var errResp api.ErrorResponse
				err = json.NewDecoder(res.Body).Decode(&errResp)
				require.NoError(t, err)

				require.Equal(t, testcase.wantErrCode, errResp.Code)
				require.Equal(t, testcase.wantErrDetail, errResp.Detail)
			}
		})
	}
}

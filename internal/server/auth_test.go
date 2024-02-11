package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/internal/server"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/alvii147/flagger-api/pkg/utils"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	defer testkitinternal.TeardownTests()
	testkitinternal.SetupTests()
	code := m.Run()
	os.Exit(code)
}

func TestGetAPIKeyIDParam(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name         string
		vars         map[string]string
		wantAPIKeyID int
		wantErr      bool
	}{
		{
			name: "Valid API key ID",
			vars: map[string]string{
				"id": "42",
			},
			wantAPIKeyID: 42,
			wantErr:      false,
		},
		{
			name: "No API key ID",
			vars: map[string]string{
				"dead": "beef",
			},
			wantAPIKeyID: 0,
			wantErr:      true,
		},
		{
			name: "Invalid API key ID",
			vars: map[string]string{
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

			apiKeyID, err := server.GetAPIKeyIDParam(testcase.vars)
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
	ctrl, err := server.NewController()
	require.NoError(t, err)

	router := ctrl.Route()
	srv := httptest.NewServer(router)

	t.Cleanup(func() {
		err = ctrl.Close()
		require.NoError(t, err)
	})

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

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
					"email": "%s",
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, "1nv4l1d3m41l", password, firstName, lastName),
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
					"email": "%s",
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, "", password, firstName, lastName),
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
					"password": "%s",
					"first_name": "%s",
					"last_name": "%s"
				}
			`, email, "", firstName, lastName),
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
					"first_name": "%s",
					"last_name": "%s"
				}
			`, email, password, "", lastName),
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
					"last_name": "%s"
				}
			`, email, password, firstName, ""),
			wantStatusCode: http.StatusBadRequest,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			mailCount := len(mailclient.GetInMemMailLogs())

			req, err := http.NewRequest(
				http.MethodPost,
				srv.URL+"/auth/users",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

			userCreatedAt := time.Now().UTC()
			res, err := httpClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, testcase.wantStatusCode, res.StatusCode)

			if httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				defer res.Body.Close()
				var createUserResp api.CreateUserResponse
				err = json.NewDecoder(res.Body).Decode(&createUserResp)
				require.NoError(t, err)

				require.Equal(t, email, createUserResp.Email)
				require.Equal(t, firstName, createUserResp.FirstName)
				require.Equal(t, lastName, createUserResp.LastName)
				testkit.RequireTimeAlmostEqual(t, userCreatedAt, createUserResp.CreatedAt)

				time.Sleep(5 * time.Second)

				mailLogs := mailclient.GetInMemMailLogs()
				require.Len(t, mailLogs, mailCount+1)

				lastMail := mailLogs[len(mailLogs)-1]
				require.Equal(t, []string{createUserResp.Email}, lastMail.To)
				require.Equal(t, "Welcome to Flagger!", lastMail.Subject)
				testkit.RequireTimeAlmostEqual(t, userCreatedAt, lastMail.SentAt)

				mailMessage := string(lastMail.Message)
				require.Contains(t, mailMessage, "Welcome to Flagger!")
				require.Contains(t, mailMessage, "Flagger - Activate Your Account")
			} else {
				mailLogs := mailclient.GetInMemMailLogs()
				require.Len(t, mailLogs, mailCount)

				defer res.Body.Close()
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

	config := env.GetConfig()

	ctrl, err := server.NewController()
	require.NoError(t, err)

	router := ctrl.Route()
	srv := httptest.NewServer(router)

	t.Cleanup(func() {
		err = ctrl.Close()
		require.NoError(t, err)
	})

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

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
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Duration(config.ActivationLifetime))),
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
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Duration(config.ActivationLifetime))),
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
				srv.URL+"/auth/users/activate",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

			res, err := httpClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, testcase.wantStatusCode, res.StatusCode)

			if !httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				defer res.Body.Close()
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

	ctrl, err := server.NewController()
	require.NoError(t, err)

	router := ctrl.Route()
	srv := httptest.NewServer(router)

	t.Cleanup(func() {
		err = ctrl.Close()
		require.NoError(t, err)
	})

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	activeUserAccessJWT, _ := testkitinternal.MustCreateUserAuthAccessJWTs(activeUser.UUID)
	_, activeUserAPIKey := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, nil)

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})
	inactiveUserAccessJWT, _ := testkitinternal.MustCreateUserAuthAccessJWTs(inactiveUser.UUID)
	_, inactiveUserAPIKey := testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)

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
				"Authorization": fmt.Sprintf("X-API-Key %s", activeUserAPIKey),
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
				"Authorization": fmt.Sprintf("X-API-Key %s", inactiveUserAPIKey),
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

			req, err := http.NewRequest(http.MethodGet, srv.URL+testcase.path, http.NoBody)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

			res, err := httpClient.Do(req)
			require.NoError(t, err)
			require.Equal(t, testcase.wantStatusCode, res.StatusCode)

			if httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				defer res.Body.Close()
				var getUserMeResp api.GetUserMeResponse
				err = json.NewDecoder(res.Body).Decode(&getUserMeResp)
				require.NoError(t, err)

				require.Equal(t, testcase.user.UUID, getUserMeResp.UUID)
				require.Equal(t, testcase.user.Email, getUserMeResp.Email)
				require.Equal(t, testcase.user.FirstName, getUserMeResp.FirstName)
				require.Equal(t, testcase.user.LastName, getUserMeResp.LastName)
				testkit.RequireTimeAlmostEqual(t, testcase.user.CreatedAt, getUserMeResp.CreatedAt)
			} else {
				defer res.Body.Close()
				var errResp api.ErrorResponse
				err = json.NewDecoder(res.Body).Decode(&errResp)
				require.NoError(t, err)

				require.Equal(t, testcase.wantErrCode, errResp.Code)
				require.Equal(t, testcase.wantErrDetail, errResp.Detail)
			}
		})
	}
}

func TestAuthFlow(t *testing.T) {
	config := env.GetConfig()

	ctrl, err := server.NewController()
	require.NoError(t, err)

	router := ctrl.Route()
	srv := httptest.NewServer(router)

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	email := testkit.GenerateFakeEmail()
	password := testkit.GenerateFakePassword()
	firstName := testkit.MustGenerateRandomString(8, true, true, false)
	lastName := testkit.MustGenerateRandomString(8, true, true, false)

	mailCount := len(mailclient.GetInMemMailLogs())

	reqBody := fmt.Sprintf(`
		{
			"email": "%s",
			"password": "%s",
			"first_name": "%s",
			"last_name": "%s"
		}
	`, email, password, firstName, lastName)
	req, err := http.NewRequest(
		http.MethodPost,
		srv.URL+"/auth/users",
		bytes.NewReader([]byte(reqBody)),
	)
	require.NoError(t, err)

	res, err := httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)

	defer res.Body.Close()
	var createUserResp api.CreateUserResponse
	err = json.NewDecoder(res.Body).Decode(&createUserResp)
	require.NoError(t, err)

	userCreatedAt := time.Now().UTC()
	require.Equal(t, email, createUserResp.Email)
	require.Equal(t, firstName, createUserResp.FirstName)
	require.Equal(t, lastName, createUserResp.LastName)
	testkit.RequireTimeAlmostEqual(t, userCreatedAt, createUserResp.CreatedAt)

	time.Sleep(5 * time.Second)

	mailLogs := mailclient.GetInMemMailLogs()
	require.Len(t, mailLogs, mailCount+1)

	lastMail := mailLogs[len(mailLogs)-1]
	require.Equal(t, []string{createUserResp.Email}, lastMail.To)
	require.Equal(t, "Welcome to Flagger!", lastMail.Subject)
	testkit.RequireTimeAlmostEqual(t, userCreatedAt, lastMail.SentAt)

	mailMessage := string(lastMail.Message)
	require.Contains(t, mailMessage, "Welcome to Flagger!")
	require.Contains(t, mailMessage, "Flagger - Activate Your Account")

	pattern := fmt.Sprintf(config.FrontendBaseURL+config.FrontendActivationRoute, `(\S+)`)
	r, err := regexp.Compile(pattern)
	require.NoError(t, err)

	matches := r.FindStringSubmatch(mailMessage)
	require.Len(t, matches, 2)

	activationToken := matches[1]
	reqBody = fmt.Sprintf(`
		{
			"token": "%s"
		}
	`, activationToken)
	req, err = http.NewRequest(
		http.MethodPost,
		srv.URL+"/auth/users/activate",
		bytes.NewReader([]byte(reqBody)),
	)
	require.NoError(t, err)

	res, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	defer res.Body.Close()

	reqBody = fmt.Sprintf(`
		{
			"email": "%s",
			"password": "%s"
		}
	`, email, password)
	req, err = http.NewRequest(
		http.MethodPost,
		srv.URL+"/auth/tokens",
		bytes.NewReader([]byte(reqBody)),
	)
	require.NoError(t, err)

	res, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)

	defer res.Body.Close()
	var createJWTResp api.CreateTokenResponse
	err = json.NewDecoder(res.Body).Decode(&createJWTResp)
	require.NoError(t, err)

	req, err = http.NewRequest(
		http.MethodGet,
		srv.URL+"/auth/users/me",
		nil,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", createJWTResp.Access))

	res, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	var getUserMeResp api.GetUserMeResponse
	err = json.NewDecoder(res.Body).Decode(&getUserMeResp)
	require.NoError(t, err)

	require.Equal(t, createUserResp.UUID, getUserMeResp.UUID)
	require.Equal(t, email, getUserMeResp.Email)
	require.Equal(t, firstName, getUserMeResp.FirstName)
	require.Equal(t, lastName, getUserMeResp.LastName)
	require.Equal(t, createUserResp.CreatedAt, getUserMeResp.CreatedAt)

	reqBody = fmt.Sprintf(`
		{
			"refresh": "%s"
		}
	`, createJWTResp.Refresh)
	req, err = http.NewRequest(
		http.MethodPost,
		srv.URL+"/auth/tokens/refresh",
		bytes.NewReader([]byte(reqBody)),
	)
	require.NoError(t, err)

	res, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)
}

func TestAPIKeyFlow(t *testing.T) {
	t.Parallel()

	config := env.GetConfig()

	ctrl, err := server.NewController()
	require.NoError(t, err)

	router := ctrl.Route()
	srv := httptest.NewServer(router)

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	jti := uuid.NewString()
	now := time.Now().UTC()
	accessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		&api.AuthJWTClaims{
			Subject:   user.UUID,
			TokenType: string(auth.JWTTypeAccess),
			IssuedAt:  utils.JSONTimeStamp(now),
			ExpiresAt: utils.JSONTimeStamp(now.Add(time.Hour)),
			JWTID:     jti,
		},
	).SignedString([]byte(config.SecretKey))
	require.NoError(t, err)

	apiKeyCreatedAt := time.Now().UTC()
	apiKeyExpiresAt := time.Now().UTC().AddDate(0, 1, 0)
	reqBody := fmt.Sprintf(`
		{
			"name": "%s",
			"expires_at": "%s"
		}
	`, "MyAPIKey", apiKeyExpiresAt.Format(time.RFC3339))
	req, err := http.NewRequest(
		http.MethodPost,
		srv.URL+"/auth/api-keys",
		bytes.NewReader([]byte(reqBody)),
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err := httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)

	defer res.Body.Close()
	var createAPIKeyResp api.CreateAPIKeyResponse
	err = json.NewDecoder(res.Body).Decode(&createAPIKeyResp)
	require.NoError(t, err)

	require.Equal(t, user.UUID, createAPIKeyResp.UserUUID)
	require.Equal(t, "MyAPIKey", createAPIKeyResp.Name)
	testkit.RequireTimeAlmostEqual(t, apiKeyCreatedAt, createAPIKeyResp.CreatedAt)
	testkit.RequirePGTimestampAlmostEqual(
		t,
		pgtype.Timestamp{
			Time:  apiKeyExpiresAt,
			Valid: true,
		},
		createAPIKeyResp.ExpiresAt,
	)

	req, err = http.NewRequest(
		http.MethodGet,
		srv.URL+"/auth/api-keys",
		nil,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	var listAPIKeysResp api.ListAPIKeysResponse
	err = json.NewDecoder(res.Body).Decode(&listAPIKeysResp)
	require.NoError(t, err)

	require.Len(t, listAPIKeysResp.Keys, 1)
	require.True(t, strings.HasPrefix(createAPIKeyResp.RawKey, listAPIKeysResp.Keys[0].Prefix))
	require.Equal(t, "MyAPIKey", listAPIKeysResp.Keys[0].Name)
	testkit.RequireTimeAlmostEqual(t, apiKeyCreatedAt, listAPIKeysResp.Keys[0].CreatedAt)
	testkit.RequirePGTimestampAlmostEqual(
		t,
		pgtype.Timestamp{
			Time:  apiKeyExpiresAt,
			Valid: true,
		},
		listAPIKeysResp.Keys[0].ExpiresAt,
	)

	req, err = http.NewRequest(
		http.MethodGet,
		srv.URL+"/api/auth/users/me",
		nil,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("X-API-Key %s", createAPIKeyResp.RawKey))

	res, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	var getUserMeResp api.GetUserMeResponse
	err = json.NewDecoder(res.Body).Decode(&getUserMeResp)
	require.NoError(t, err)

	require.Equal(t, user.Email, getUserMeResp.Email)
	require.Equal(t, user.FirstName, getUserMeResp.FirstName)
	require.Equal(t, user.LastName, getUserMeResp.LastName)
	require.Equal(t, user.CreatedAt, getUserMeResp.CreatedAt)

	req, err = http.NewRequest(
		http.MethodDelete,
		srv.URL+"/auth/api-keys/"+strconv.Itoa(createAPIKeyResp.ID),
		bytes.NewReader([]byte(reqBody)),
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
}

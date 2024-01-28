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

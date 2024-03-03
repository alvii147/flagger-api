package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/server"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestGetFlagIDParam(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		pathValues map[string]string
		wantFlagID int
		wantErr    bool
	}{
		{
			name: "Valid flag ID",
			pathValues: map[string]string{
				"id": "42",
			},
			wantFlagID: 42,
			wantErr:    false,
		},
		{
			name: "No flag ID",
			pathValues: map[string]string{
				"dead": "beef",
			},
			wantFlagID: 0,
			wantErr:    true,
		},
		{
			name: "Invalid flag ID",
			pathValues: map[string]string{
				"id": "deadbeef",
			},
			wantFlagID: 0,
			wantErr:    true,
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

			flagID, err := server.GetFlagIDParam(req)
			if testcase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testcase.wantFlagID, flagID)
			}
		})
	}
}

func TestGetFlagNameParam(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name         string
		pathValues   map[string]string
		wantFlagName string
		wantErr      bool
	}{
		{
			name: "Valid flag name",
			pathValues: map[string]string{
				"name": "my-flag",
			},
			wantFlagName: "my-flag",
			wantErr:      false,
		},
		{
			name: "No flag name",
			pathValues: map[string]string{
				"dead": "beef",
			},
			wantFlagName: "",
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

			name, err := server.GetFlagNameParam(req)
			if testcase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testcase.wantFlagName, name)
			}
		})
	}
}

func TestHandleCreateFlag(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	userAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(user.UUID)

	testcases := []struct {
		name           string
		headers        map[string]string
		requestBody    string
		wantStatusCode int
		wantFlagName   string
		wantErrCode    string
		wantErrDetail  string
	}{
		{
			name: "Valid request",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: `
				{
					"name": "my-flag"
				}
			`,
			wantStatusCode: http.StatusCreated,
			wantFlagName:   "my-flag",
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Name missing",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody:    `{}`,
			wantStatusCode: http.StatusBadRequest,
			wantFlagName:   "",
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name: "Non-slug name",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", userAccessJWT),
			},
			requestBody: `
				{
					"name": "MyFlag"
				}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantFlagName:   "",
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name:    "Unauthenticated request",
			headers: map[string]string{},
			requestBody: `
				{
					"name": "my-unauthenticated-flag"
				}
			`,
			wantStatusCode: http.StatusUnauthorized,
			wantFlagName:   "",
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantErrDetail:  api.ErrDetailMissingCredentials,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPost,
				TestServerURL+"/flags",
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

			flagCreatedAt := time.Now().UTC()
			res, err := httpClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := res.Body.Close()
				require.NoError(t, err)
			})

			require.Equal(t, testcase.wantStatusCode, res.StatusCode)

			if httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				var createFlagResp api.CreateFlagResponse
				err = json.NewDecoder(res.Body).Decode(&createFlagResp)
				require.NoError(t, err)

				require.Equal(t, user.UUID, createFlagResp.UserUUID)
				require.Equal(t, testcase.wantFlagName, createFlagResp.Name)
				require.False(t, createFlagResp.IsEnabled)
				testkit.RequireTimeAlmostEqual(t, flagCreatedAt, createFlagResp.CreatedAt)
				testkit.RequireTimeAlmostEqual(t, flagCreatedAt, createFlagResp.UpdatedAt)
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

func TestHandleGetFlagByID(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	activeUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(activeUser.UUID)
	activeUserFlag := testkitinternal.MustCreateUserFlag(t, activeUser.UUID, "active-user-flag")

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})
	inactiveUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(inactiveUser.UUID)
	inactiveUserFlag := testkitinternal.MustCreateUserFlag(t, inactiveUser.UUID, "inactive-user-flag")

	testcases := []struct {
		name           string
		path           string
		headers        map[string]string
		wantStatusCode int
		wantErrCode    string
		wantErrDetail  string
	}{
		{
			name: "Get flag for active user",
			path: fmt.Sprintf("/flags/%d", activeUserFlag.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			wantStatusCode: http.StatusOK,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Get flag for inactive user",
			path: fmt.Sprintf("/flags/%d", inactiveUserFlag.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailFlagNotFound,
		},
		{
			name: "Get flag for another user",
			path: fmt.Sprintf("/flags/%d", activeUserFlag.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailFlagNotFound,
		},
		{
			name: "Get non-existent flag",
			path: fmt.Sprintf("/flags/%d", 314159),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			wantStatusCode: http.StatusNotFound,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailFlagNotFound,
		},
		{
			name:           "Get flag without authentication",
			path:           fmt.Sprintf("/flags/%d", activeUserFlag.ID),
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
				var getFlagByIDResp api.GetFlagByIDResponse
				err = json.NewDecoder(res.Body).Decode(&getFlagByIDResp)
				require.NoError(t, err)

				require.Equal(t, activeUser.UUID, getFlagByIDResp.UserUUID)
				require.Equal(t, activeUserFlag.Name, getFlagByIDResp.Name)
				require.Equal(t, activeUserFlag.IsEnabled, getFlagByIDResp.IsEnabled)
				testkit.RequireTimeAlmostEqual(t, activeUserFlag.CreatedAt, getFlagByIDResp.CreatedAt)
				testkit.RequireTimeAlmostEqual(t, activeUserFlag.UpdatedAt, getFlagByIDResp.UpdatedAt)
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

func TestHandleGetFlagByName(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	_, activeUserRawAPIKey := testkitinternal.MustCreateUserAPIKey(t, activeUser.UUID, nil)
	activeUserFlag := testkitinternal.MustCreateUserFlag(t, activeUser.UUID, "active-user-flag")

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})
	_, inactiveUserRawAPIKey := testkitinternal.MustCreateUserAPIKey(t, inactiveUser.UUID, nil)
	inactiveUserFlag := testkitinternal.MustCreateUserFlag(t, inactiveUser.UUID, "inactive-user-flag")

	testcases := []struct {
		name           string
		path           string
		headers        map[string]string
		wantStatusCode int
		wantName       string
		wantValid      bool
		wantErrCode    string
		wantErrDetail  string
	}{
		{
			name: "Get flag for active user",
			path: fmt.Sprintf("/api/flags/%s", activeUserFlag.Name),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("X-API-Key %s", activeUserRawAPIKey),
			},
			wantStatusCode: http.StatusOK,
			wantName:       activeUserFlag.Name,
			wantValid:      true,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Get flag for inactive user",
			path: fmt.Sprintf("/api/flags/%s", inactiveUserFlag.Name),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("X-API-Key %s", inactiveUserRawAPIKey),
			},
			wantStatusCode: http.StatusUnauthorized,
			wantName:       "",
			wantValid:      false,
			wantErrCode:    api.ErrCodeInvalidCredentials,
			wantErrDetail:  api.ErrDetailInvalidToken,
		},
		{
			name: "Get flag for another user",
			path: fmt.Sprintf("/api/flags/%s", inactiveUserFlag.Name),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("X-API-Key %s", activeUserRawAPIKey),
			},
			wantStatusCode: http.StatusOK,
			wantName:       inactiveUserFlag.Name,
			wantValid:      false,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Get non-existent flag",
			path: fmt.Sprintf("/api/flags/%s", "non-existent-flag"),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("X-API-Key %s", activeUserRawAPIKey),
			},
			wantStatusCode: http.StatusOK,
			wantName:       "non-existent-flag",
			wantValid:      false,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name:           "Get flag without authentication",
			path:           fmt.Sprintf("/api/flags/%s", activeUserFlag.Name),
			headers:        map[string]string{},
			wantStatusCode: http.StatusUnauthorized,
			wantName:       "",
			wantValid:      false,
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
				var getFlagByNameResp api.GetFlagByNameResponse
				err = json.NewDecoder(res.Body).Decode(&getFlagByNameResp)
				require.NoError(t, err)

				require.Equal(t, testcase.wantName, getFlagByNameResp.Name)
				require.Equal(t, testcase.wantValid, getFlagByNameResp.Valid)

				if testcase.wantValid {
					require.NotNil(t, getFlagByNameResp.ID)
					require.NotNil(t, getFlagByNameResp.UserUUID)
					require.Equal(t, activeUser.UUID, *getFlagByNameResp.UserUUID)
					require.Equal(t, activeUserFlag.IsEnabled, getFlagByNameResp.IsEnabled)
					testkit.RequirePGTimestampAlmostEqual(t, pgtype.Timestamp{
						Time:  activeUserFlag.CreatedAt,
						Valid: true,
					}, getFlagByNameResp.CreatedAt)
					testkit.RequirePGTimestampAlmostEqual(t, pgtype.Timestamp{
						Time:  activeUserFlag.UpdatedAt,
						Valid: true,
					}, getFlagByNameResp.UpdatedAt)
				} else {
					require.Nil(t, getFlagByNameResp.ID)
					require.Nil(t, getFlagByNameResp.UserUUID)
					require.False(t, getFlagByNameResp.IsEnabled)
					require.False(t, getFlagByNameResp.CreatedAt.Valid)
					require.False(t, getFlagByNameResp.UpdatedAt.Valid)
					require.False(t, getFlagByNameResp.Valid)
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

func TestFlagFlow(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	user, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	accessToken, _ := testkitinternal.MustCreateUserAuthJWTs(user.UUID)

	_, rawKey := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	req, err := http.NewRequest(
		http.MethodGet,
		TestServerURL+"/api/flags/my-flag",
		http.NoBody,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("X-API-Key %s", rawKey))

	res, err := httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	var getFlagByNameResp api.GetFlagByNameResponse
	err = json.NewDecoder(res.Body).Decode(&getFlagByNameResp)
	require.NoError(t, err)

	require.Nil(t, getFlagByNameResp.ID)
	require.Nil(t, getFlagByNameResp.UserUUID)
	require.Equal(t, "my-flag", getFlagByNameResp.Name)
	require.False(t, getFlagByNameResp.IsEnabled)
	require.False(t, getFlagByNameResp.CreatedAt.Valid)
	require.False(t, getFlagByNameResp.UpdatedAt.Valid)

	flagName := "my-flag"
	reqBody := fmt.Sprintf(`
		{
			"name": "%s"
		}
	`, flagName)
	req, err = http.NewRequest(
		http.MethodPost,
		TestServerURL+"/flags",
		bytes.NewReader([]byte(reqBody)),
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	flagCreatedAt := time.Now().UTC()
	res, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, res.StatusCode)

	defer res.Body.Close()
	var createFlagResp api.CreateFlagResponse
	err = json.NewDecoder(res.Body).Decode(&createFlagResp)
	require.NoError(t, err)

	require.Equal(t, user.UUID, createFlagResp.UserUUID)
	require.Equal(t, flagName, createFlagResp.Name)
	require.False(t, createFlagResp.IsEnabled)
	testkit.RequireTimeAlmostEqual(t, flagCreatedAt, createFlagResp.CreatedAt)
	testkit.RequireTimeAlmostEqual(t, flagCreatedAt, createFlagResp.UpdatedAt)

	req, err = http.NewRequest(
		http.MethodGet,
		TestServerURL+"/flags",
		nil,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	var listFlagsResp api.ListFlagsResponse
	err = json.NewDecoder(res.Body).Decode(&listFlagsResp)
	require.NoError(t, err)

	require.Len(t, listFlagsResp.Flags, 1)
	require.Equal(t, user.UUID, listFlagsResp.Flags[0].UserUUID)
	require.Equal(t, flagName, listFlagsResp.Flags[0].Name)
	require.False(t, listFlagsResp.Flags[0].IsEnabled)
	testkit.RequireTimeAlmostEqual(t, flagCreatedAt, listFlagsResp.Flags[0].CreatedAt)
	testkit.RequireTimeAlmostEqual(t, flagCreatedAt, listFlagsResp.Flags[0].UpdatedAt)

	reqBody = `
		{
			"is_enabled": true
		}
	`
	req, err = http.NewRequest(
		http.MethodPut,
		TestServerURL+"/flags/"+strconv.Itoa(createFlagResp.ID),
		bytes.NewReader([]byte(reqBody)),
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	flagUpdatedAt := time.Now().UTC()
	res, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	var updateFlagsResp api.UpdateFlagResponse
	err = json.NewDecoder(res.Body).Decode(&updateFlagsResp)
	require.NoError(t, err)

	require.Equal(t, createFlagResp.ID, updateFlagsResp.ID)
	require.Equal(t, user.UUID, updateFlagsResp.UserUUID)
	require.Equal(t, flagName, updateFlagsResp.Name)
	require.True(t, updateFlagsResp.IsEnabled)
	testkit.RequireTimeAlmostEqual(t, createFlagResp.CreatedAt, updateFlagsResp.CreatedAt)
	testkit.RequireTimeAlmostEqual(t, flagUpdatedAt, updateFlagsResp.UpdatedAt)

	req, err = http.NewRequest(
		http.MethodGet,
		TestServerURL+"/flags/"+strconv.Itoa(updateFlagsResp.ID),
		http.NoBody,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	res, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	var getFlagByIDResp api.GetFlagByIDResponse
	err = json.NewDecoder(res.Body).Decode(&getFlagByIDResp)
	require.NoError(t, err)

	require.Equal(t, updateFlagsResp.ID, getFlagByIDResp.ID)
	require.Equal(t, user.UUID, getFlagByIDResp.UserUUID)
	require.Equal(t, updateFlagsResp.Name, getFlagByIDResp.Name)
	require.True(t, getFlagByIDResp.IsEnabled)
	testkit.RequireTimeAlmostEqual(t, updateFlagsResp.CreatedAt, getFlagByIDResp.CreatedAt)
	testkit.RequireTimeAlmostEqual(t, updateFlagsResp.UpdatedAt, getFlagByIDResp.UpdatedAt)

	req, err = http.NewRequest(
		http.MethodGet,
		TestServerURL+"/api/flags/"+getFlagByIDResp.Name,
		http.NoBody,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("X-API-Key %s", rawKey))

	res, err = httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	defer res.Body.Close()
	err = json.NewDecoder(res.Body).Decode(&getFlagByNameResp)
	require.NoError(t, err)

	require.NotNil(t, getFlagByNameResp.ID)
	require.Equal(t, getFlagByIDResp.ID, *getFlagByNameResp.ID)
	require.NotNil(t, getFlagByNameResp.UserUUID)
	require.Equal(t, user.UUID, *getFlagByNameResp.UserUUID)
	require.Equal(t, getFlagByIDResp.Name, getFlagByNameResp.Name)
	require.True(t, getFlagByNameResp.IsEnabled)
	testkit.RequirePGTimestampAlmostEqual(
		t,
		pgtype.Timestamp{
			Time:  getFlagByIDResp.CreatedAt,
			Valid: true,
		},
		getFlagByNameResp.CreatedAt,
	)
	testkit.RequirePGTimestampAlmostEqual(
		t,
		pgtype.Timestamp{
			Time:  getFlagByIDResp.UpdatedAt,
			Valid: true,
		},
		getFlagByNameResp.UpdatedAt,
	)
}

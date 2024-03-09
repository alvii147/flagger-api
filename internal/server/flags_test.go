package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/flags"
	"github.com/alvii147/flagger-api/internal/server"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/alvii147/flagger-api/pkg/testkit"
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
				require.Equal(t, activeUserFlag.CreatedAt, getFlagByIDResp.CreatedAt)
				require.Equal(t, activeUserFlag.UpdatedAt, getFlagByIDResp.UpdatedAt)
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
					require.True(t, getFlagByNameResp.CreatedAt.Valid)
					require.Equal(t, activeUserFlag.CreatedAt, getFlagByNameResp.CreatedAt.Time)
					require.True(t, getFlagByNameResp.UpdatedAt.Valid)
					require.Equal(t, activeUserFlag.UpdatedAt, getFlagByNameResp.UpdatedAt.Time)
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

func TestHandleListFlags(t *testing.T) {
	t.Parallel()

	httpClient := httputils.NewHTTPClient(nil)

	activeUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = true
	})
	activeUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(activeUser.UUID)
	activeUserFlag1 := testkitinternal.MustCreateUserFlag(t, activeUser.UUID, "active-user-flag-1")
	activeUserFlag2 := testkitinternal.MustCreateUserFlag(t, activeUser.UUID, "active-user-flag-2")

	inactiveUser, _ := testkitinternal.MustCreateUser(t, func(u *auth.User) {
		u.IsActive = false
	})
	inactiveUserAccessJWT, _ := testkitinternal.MustCreateUserAuthJWTs(inactiveUser.UUID)
	testkitinternal.MustCreateUserFlag(t, inactiveUser.UUID, "inactive-user-flag-1")

	testcases := []struct {
		name           string
		headers        map[string]string
		wantStatusCode int
		wantFlags      []*flags.Flag
		wantErrCode    string
		wantErrDetail  string
	}{
		{
			name: "List active user flags",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			wantStatusCode: http.StatusOK,
			wantFlags: []*flags.Flag{
				activeUserFlag1,
				activeUserFlag2,
			},
			wantErrCode:   "",
			wantErrDetail: "",
		},
		{
			name: "List inactive user flags returns no flags",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			wantStatusCode: http.StatusOK,
			wantFlags:      []*flags.Flag{},
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name:           "List flags without authentication",
			headers:        map[string]string{},
			wantStatusCode: http.StatusUnauthorized,
			wantFlags:      []*flags.Flag{},
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantErrDetail:  api.ErrDetailMissingCredentials,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, TestServerURL+"/flags", http.NoBody)
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
				var listFlagsResp api.ListFlagsResponse
				err = json.NewDecoder(res.Body).Decode(&listFlagsResp)
				require.NoError(t, err)
				require.Len(t, listFlagsResp.Flags, len(testcase.wantFlags))

				sort.Slice(listFlagsResp.Flags, func(i, j int) bool {
					return listFlagsResp.Flags[i].Name < listFlagsResp.Flags[j].Name
				})

				sort.Slice(testcase.wantFlags, func(i, j int) bool {
					return testcase.wantFlags[i].Name < testcase.wantFlags[j].Name
				})

				for i, listedFlag := range listFlagsResp.Flags {
					wantFlag := testcase.wantFlags[i]
					require.Equal(t, wantFlag.ID, listedFlag.ID)
					require.Equal(t, wantFlag.UserUUID, listedFlag.UserUUID)
					require.Equal(t, wantFlag.Name, listedFlag.Name)
					require.Equal(t, wantFlag.IsEnabled, listedFlag.IsEnabled)
					require.Equal(t, wantFlag.CreatedAt, listedFlag.CreatedAt)
					require.Equal(t, wantFlag.UpdatedAt, listedFlag.UpdatedAt)
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

func TestHandleUpdateFlag(t *testing.T) {
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
	inactiveUserFlag := testkitinternal.MustCreateUserFlag(t, inactiveUser.UUID, "inactive-user-flag-1")

	testcases := []struct {
		name           string
		path           string
		headers        map[string]string
		requestBody    string
		wantStatusCode int
		wantIsEnabled  bool
		wantErrCode    string
		wantErrDetail  string
	}{
		{
			name: "Update flag for active user",
			path: fmt.Sprintf("/flags/%d", activeUserFlag.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			requestBody: `
				{
					"is_enabled": true
				}
			`,
			wantStatusCode: http.StatusOK,
			wantIsEnabled:  true,
			wantErrCode:    "",
			wantErrDetail:  "",
		},
		{
			name: "Update flag for inactive user",
			path: fmt.Sprintf("/flags/%d", inactiveUserFlag.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", inactiveUserAccessJWT),
			},
			requestBody: `
				{
					"is_enabled": true
				}
			`,
			wantStatusCode: http.StatusNotFound,
			wantIsEnabled:  false,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailFlagNotFound,
		},
		{
			name: "Update flag for another user",
			path: fmt.Sprintf("/flags/%d", inactiveUserFlag.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			requestBody: `
				{
					"is_enabled": true
				}
			`,
			wantStatusCode: http.StatusNotFound,
			wantIsEnabled:  false,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailFlagNotFound,
		},
		{
			name: "Update non-existent flag",
			path: "/flags/314159",
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			requestBody: `
				{
					"is_enabled": true
				}
			`,
			wantStatusCode: http.StatusNotFound,
			wantIsEnabled:  false,
			wantErrCode:    api.ErrCodeResourceNotFound,
			wantErrDetail:  api.ErrDetailFlagNotFound,
		},
		{
			name: "Update flag with invalid is_enabled",
			path: fmt.Sprintf("/flags/%d", activeUserFlag.ID),
			headers: map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", activeUserAccessJWT),
			},
			requestBody: `
				{
					"is_enabled": "invalid"
				}
			`,
			wantStatusCode: http.StatusBadRequest,
			wantIsEnabled:  false,
			wantErrCode:    api.ErrCodeInvalidRequest,
			wantErrDetail:  api.ErrDetailInvalidRequestData,
		},
		{
			name:    "Update flag without authentication",
			path:    fmt.Sprintf("/flags/%d", activeUserFlag.ID),
			headers: map[string]string{},
			requestBody: `
				{
					"is_enabled": true
				}
			`,
			wantStatusCode: http.StatusUnauthorized,
			wantIsEnabled:  false,
			wantErrCode:    api.ErrCodeMissingCredentials,
			wantErrDetail:  api.ErrDetailMissingCredentials,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				http.MethodPut,
				TestServerURL+testcase.path,
				bytes.NewReader([]byte(testcase.requestBody)),
			)
			require.NoError(t, err)

			for key, value := range testcase.headers {
				req.Header.Add(key, value)
			}

			updatedAt := time.Now().UTC()
			res, err := httpClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := res.Body.Close()
				require.NoError(t, err)
			})

			require.Equal(t, testcase.wantStatusCode, res.StatusCode)
			if httputils.IsHTTPSuccess(testcase.wantStatusCode) {
				var updateFlagsResp api.UpdateFlagResponse
				err = json.NewDecoder(res.Body).Decode(&updateFlagsResp)
				require.NoError(t, err)

				require.Equal(t, activeUserFlag.ID, updateFlagsResp.ID)
				require.Equal(t, activeUser.UUID, updateFlagsResp.UserUUID)
				require.Equal(t, activeUserFlag.Name, updateFlagsResp.Name)
				require.Equal(t, testcase.wantIsEnabled, updateFlagsResp.IsEnabled)
				require.Equal(t, activeUserFlag.CreatedAt, updateFlagsResp.CreatedAt)
				testkit.RequireTimeAlmostEqual(t, updatedAt, updateFlagsResp.UpdatedAt)
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

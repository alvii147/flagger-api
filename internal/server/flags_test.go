package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/server"
	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/alvii147/flagger-api/pkg/api"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestGetFlagIDParam(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		vars       map[string]string
		wantFlagID int
		wantErr    bool
	}{
		{
			name: "Valid flag ID",
			vars: map[string]string{
				"id": "42",
			},
			wantFlagID: 42,
			wantErr:    false,
		},
		{
			name: "No flag ID",
			vars: map[string]string{
				"dead": "beef",
			},
			wantFlagID: 0,
			wantErr:    true,
		},
		{
			name: "Invalid flag ID",
			vars: map[string]string{
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

			flagID, err := server.GetFlagIDParam(testcase.vars)
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
		vars         map[string]string
		wantFlagName string
		wantErr      bool
	}{
		{
			name: "Valid flag name",
			vars: map[string]string{
				"name": "my-flag",
			},
			wantFlagName: "my-flag",
			wantErr:      false,
		},
		{
			name: "No flag name",
			vars: map[string]string{
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

			name, err := server.GetFlagNameParam(testcase.vars)
			if testcase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testcase.wantFlagName, name)
			}
		})
	}
}

func TestFlagFlow(t *testing.T) {
	t.Parallel()

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
	accessToken, _ := testkitinternal.MustCreateUserAuthJWTs(user.UUID)

	_, rawKey := testkitinternal.MustCreateUserAPIKey(t, user.UUID, nil)

	req, err := http.NewRequest(
		http.MethodGet,
		srv.URL+"/api/flags/my-flag",
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

	reqBody := `
		{
			"name": "my-flag"
		}
	`
	req, err = http.NewRequest(
		http.MethodPost,
		srv.URL+"/flags",
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
	require.Equal(t, "my-flag", createFlagResp.Name)
	require.False(t, createFlagResp.IsEnabled)
	testkit.RequireTimeAlmostEqual(t, flagCreatedAt, createFlagResp.CreatedAt)
	testkit.RequireTimeAlmostEqual(t, flagCreatedAt, createFlagResp.UpdatedAt)

	req, err = http.NewRequest(
		http.MethodGet,
		srv.URL+"/flags",
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
	require.Equal(t, "my-flag", listFlagsResp.Flags[0].Name)
	require.False(t, listFlagsResp.Flags[0].IsEnabled)
	testkit.RequireTimeAlmostEqual(t, flagCreatedAt, listFlagsResp.Flags[0].CreatedAt)
	testkit.RequireTimeAlmostEqual(t, flagCreatedAt, listFlagsResp.Flags[0].UpdatedAt)

	reqBody = `
		{
			"name": "my-updated-flag",
			"is_enabled": true
		}
	`
	req, err = http.NewRequest(
		http.MethodPut,
		srv.URL+"/flags/"+strconv.Itoa(createFlagResp.ID),
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
	require.Equal(t, "my-updated-flag", updateFlagsResp.Name)
	require.True(t, updateFlagsResp.IsEnabled)
	testkit.RequireTimeAlmostEqual(t, createFlagResp.CreatedAt, updateFlagsResp.CreatedAt)
	testkit.RequireTimeAlmostEqual(t, flagUpdatedAt, updateFlagsResp.UpdatedAt)

	req, err = http.NewRequest(
		http.MethodGet,
		srv.URL+"/flags/"+strconv.Itoa(updateFlagsResp.ID),
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
		srv.URL+"/api/flags/"+getFlagByIDResp.Name,
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

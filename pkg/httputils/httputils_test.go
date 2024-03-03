package httputils_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/stretchr/testify/require"
)

func TestGetAuthorizationHeader(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name      string
		header    http.Header
		authType  string
		wantToken string
		wantOk    bool
	}{
		{
			name: "Valid header with valid auth type",
			header: map[string][]string{
				"Authorization": {"Bearer 0xdeadbeef"},
			},
			authType:  "Bearer",
			wantToken: "0xdeadbeef",
			wantOk:    true,
		},
		{
			name:      "No header",
			header:    map[string][]string{},
			authType:  "Bearer",
			wantToken: "0xdeadbeef",
			wantOk:    false,
		},
		{
			name: "Invalid auth type",
			header: map[string][]string{
				"Authorization": {"Bearer 0xdeadbeef"},
			},
			authType:  "Basic",
			wantToken: "0xdeadbeef",
			wantOk:    false,
		},
		{
			name: "Valid header with spaces",
			header: map[string][]string{
				"Authorization": {"  Bearer   0xdeadbeef    "},
			},
			authType:  "Bearer",
			wantToken: "0xdeadbeef",
			wantOk:    true,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			token, ok := httputils.GetAuthorizationHeader(testcase.header, testcase.authType)
			require.Equal(t, testcase.wantOk, ok)
			if testcase.wantOk {
				require.Equal(t, testcase.wantToken, token)
			}
		})
	}
}

func TestIsHTTPSuccess(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name        string
		statusCode  int
		wantSuccess bool
	}{
		{
			name:        "200 OK is success",
			statusCode:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:        "201 Created is success",
			statusCode:  http.StatusCreated,
			wantSuccess: true,
		},
		{
			name:        "204 No content is success",
			statusCode:  http.StatusNoContent,
			wantSuccess: true,
		},
		{
			name:        "302 Found is not success",
			statusCode:  http.StatusFound,
			wantSuccess: false,
		},
		{
			name:        "400 Bad request is not success",
			statusCode:  http.StatusBadRequest,
			wantSuccess: false,
		},
		{
			name:        "401 Unauthorized is not success",
			statusCode:  http.StatusUnauthorized,
			wantSuccess: false,
		},
		{
			name:        "403 Forbidden is not success",
			statusCode:  http.StatusForbidden,
			wantSuccess: false,
		},
		{
			name:        "404 Not found is not success",
			statusCode:  http.StatusNotFound,
			wantSuccess: false,
		},
		{
			name:        "405 Method not allowed is not success",
			statusCode:  http.StatusMethodNotAllowed,
			wantSuccess: false,
		},
		{
			name:        "500 Internal server error is not success",
			statusCode:  http.StatusInternalServerError,
			wantSuccess: false,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, testcase.wantSuccess, httputils.IsHTTPSuccess(testcase.statusCode))
		})
	}
}

func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name        string
		modifier    func(c *http.Client)
		wantTimeout time.Duration
	}{
		{
			name:        "No modifier",
			modifier:    nil,
			wantTimeout: httputils.HTTPClientDefaultTimeout,
		},
		{
			name:        "Empty modifier",
			modifier:    func(c *http.Client) {},
			wantTimeout: httputils.HTTPClientDefaultTimeout,
		},
		{
			name: "Timeout modifier",
			modifier: func(c *http.Client) {
				c.Timeout = 5 * time.Second
			},
			wantTimeout: 5 * time.Second,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			httpClient := httputils.NewHTTPClient(testcase.modifier)
			require.Equal(t, testcase.wantTimeout, httpClient.Timeout)
		})
	}
}

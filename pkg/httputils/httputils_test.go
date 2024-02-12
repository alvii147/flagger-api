package httputils_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/stretchr/testify/require"
)

func TestResponseWriterHeader(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	w := httputils.NewResponseWriter(rec)

	w.Header().Set("Content-Type", "application/json")
	require.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestResponseWriterWrite(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	w := httputils.NewResponseWriter(rec)

	data := "DEADBEEF"

	w.Write([]byte(data))
	require.Equal(t, data, string(rec.Body.Bytes()))
}

func TestResponseWriterWriteHeader(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "Status code OK",
			statusCode: http.StatusOK,
		},
		{
			name:       "Status code created",
			statusCode: http.StatusCreated,
		},
		{
			name:       "Status code moved permanently",
			statusCode: http.StatusMovedPermanently,
		},
		{
			name:       "Status code found",
			statusCode: http.StatusFound,
		},
		{
			name:       "Status code not found",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Status code internal server error",
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			w := httputils.NewResponseWriter(rec)

			w.WriteHeader(testcase.statusCode)
			require.Equal(t, testcase.statusCode, w.StatusCode)
			require.Equal(t, testcase.statusCode, rec.Code)
		})
	}
}

func TestResponseWriterWriteJSON(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	w := httputils.NewResponseWriter(rec)

	data := map[string]interface{}{
		"number": float64(42),
		"string": "Hello",
		"null":   nil,
		"listOfNumbers": []interface{}{
			float64(3),
			float64(1),
			float64(4),
			float64(1),
			float64(6),
		},
	}

	w.WriteJSON(data, http.StatusOK)

	writtenData := make(map[string]interface{})
	json.NewDecoder(rec.Body).Decode(&writtenData)

	require.Equal(t, data["number"], writtenData["number"])
	require.Equal(t, data["string"], writtenData["string"])
	require.Equal(t, data["null"], writtenData["null"])
	require.Equal(t, data["listOfNumbers"], writtenData["listOfNumbers"])
}

func TestResponseWriterMiddleware(t *testing.T) {
	t.Parallel()

	nextCallCount := 0
	var next httputils.HandlerFunc = func(w *httputils.ResponseWriter, r *http.Request) {
		nextCallCount++
	}

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/request/url/path", http.NoBody)

	httputils.ResponseWriterMiddleware(next).ServeHTTP(rec, r)
	require.Equal(t, 1, nextCallCount)
}

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

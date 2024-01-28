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

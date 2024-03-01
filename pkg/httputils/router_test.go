package httputils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/stretchr/testify/require"
)

func TestRouterMethods(t *testing.T) {
	t.Parallel()

	getHandlerCallCount := 0
	getHandler := func(w *httputils.ResponseWriter, r *http.Request) {
		getHandlerCallCount++
		w.WriteHeader(http.StatusOK)
	}

	postHandlerCallCount := 0
	postHandler := func(w *httputils.ResponseWriter, r *http.Request) {
		postHandlerCallCount++
		w.WriteHeader(http.StatusOK)
	}

	putHandlerCallCount := 0
	putHandler := func(w *httputils.ResponseWriter, r *http.Request) {
		putHandlerCallCount++
		w.WriteHeader(http.StatusOK)
	}

	deleteHandlerCallCount := 0
	deleteHandler := func(w *httputils.ResponseWriter, r *http.Request) {
		deleteHandlerCallCount++
		w.WriteHeader(http.StatusOK)
	}

	router := httputils.NewRouter()
	router.GET("/path/get", getHandler)
	router.POST("/path/post", postHandler)
	router.PUT("/path/put", putHandler)
	router.DELETE("/path/delete", deleteHandler)

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	testcases := []struct {
		name           string
		method         string
		path           string
		wantStatusCode int
	}{
		{
			name:           "GET request",
			method:         http.MethodGet,
			path:           "/path/get",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "POST request",
			method:         http.MethodPost,
			path:           "/path/post",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "PUT request",
			method:         http.MethodPut,
			path:           "/path/put",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "DELETE request",
			method:         http.MethodDelete,
			path:           "/path/delete",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Unregistered path",
			method:         http.MethodGet,
			path:           "/path/deadbeef",
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "Unsupported method",
			method:         http.MethodPost,
			path:           "/path/get",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(
				testcase.method,
				srv.URL+testcase.path,
				http.NoBody,
			)
			require.NoError(t, err)

			res, err := httpClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() {
				err := res.Body.Close()
				require.NoError(t, err)
			})

			require.Equal(t, testcase.wantStatusCode, res.StatusCode)
		})
	}

	t.Cleanup(func() {
		require.Equal(t, getHandlerCallCount, 1)
		require.Equal(t, postHandlerCallCount, 1)
		require.Equal(t, putHandlerCallCount, 1)
		require.Equal(t, deleteHandlerCallCount, 1)
	})
}

func TestRouterMiddleware(t *testing.T) {
	t.Parallel()

	callStack := make([]string, 0)
	middleware1 := func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return httputils.HandlerFunc(func(w *httputils.ResponseWriter, r *http.Request) {
			callStack = append(callStack, "mw1")
			r.Header.Set("Middleware1", "deadbeef")
			next.ServeHTTP(w, r)
		})
	}
	middleware2 := func(next httputils.HandlerFunc) httputils.HandlerFunc {
		return httputils.HandlerFunc(func(w *httputils.ResponseWriter, r *http.Request) {
			callStack = append(callStack, "mw2")
			r.Header.Set("Middleware2", "deadbeef")
			next.ServeHTTP(w, r)
		})
	}
	handler := func(w *httputils.ResponseWriter, r *http.Request) {
		callStack = append(callStack, "h")
		require.Equal(t, "deadbeef", r.Header.Get("Middleware1"))
		require.Equal(t, "deadbeef", r.Header.Get("Middleware2"))
		w.WriteHeader(http.StatusOK)
	}

	router := httputils.NewRouter()
	router.GET("/path/get", handler, middleware1, middleware2)

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest(
		http.MethodGet,
		srv.URL+"/path/get",
		http.NoBody,
	)
	require.NoError(t, err)

	res, err := httpClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := res.Body.Close()
		require.NoError(t, err)
	})

	require.Equal(t, []string{"mw2", "mw1", "h"}, callStack)
	require.Equal(t, http.StatusOK, res.StatusCode)
}

func TestRouterPathParam(t *testing.T) {
	t.Parallel()

	router := httputils.NewRouter()
	router.GET(
		"/path/{param}",
		func(w *httputils.ResponseWriter, r *http.Request) {
			require.Equal(t, "deadbeef", r.PathValue("param"))
			w.WriteHeader(http.StatusOK)
		},
	)

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest(
		http.MethodGet,
		srv.URL+"/path/deadbeef",
		http.NoBody,
	)
	require.NoError(t, err)

	res, err := httpClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := res.Body.Close()
		require.NoError(t, err)
	})

	require.Equal(t, http.StatusOK, res.StatusCode)
}

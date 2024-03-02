package testkitinternal

import (
	"fmt"
	"net/http/httptest"

	"github.com/alvii147/flagger-api/internal/server"
)

// MustCreateTestServer creates a new Controller, sets up a new test HTTP server and panics on error.
func MustCreateTestServer() (server.Controller, *httptest.Server) {
	ctrl, err := server.NewController()
	if err != nil {
		panic(fmt.Sprintf("MustCreateTestServer failed to server.NewController: %v", err))
	}

	srv := httptest.NewServer(ctrl)

	return ctrl, srv
}

// MustCloseTestServer closes a Controller and an HTTP server.
func MustCloseTestServer(ctrl server.Controller, srv *httptest.Server) {
	ctrl.Close()
	srv.Close()
}

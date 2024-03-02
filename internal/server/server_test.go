package server_test

import (
	"os"
	"testing"

	"github.com/alvii147/flagger-api/internal/testkitinternal"
)

var TestServerURL = ""

func TestMain(m *testing.M) {
	ctrl, srv := testkitinternal.MustCreateTestServer()
	TestServerURL = srv.URL

	code := m.Run()

	testkitinternal.MustCloseTestServer(ctrl, srv)
	os.Exit(code)
}

package testkitinternal_test

import (
	"os"
	"testing"

	"github.com/alvii147/flagger-api/internal/testkitinternal"
)

func TestMain(m *testing.M) {
	defer testkitinternal.TeardownTests()
	testkitinternal.SetupTests()
	code := m.Run()
	testkitinternal.TeardownTests()
	os.Exit(code)
}

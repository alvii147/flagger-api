package testkitinternal

import (
	"os"

	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/pkg/logging"
)

// SetupTests prepares testing environment.
func SetupTests() {
	config := env.NewConfig()
	env.SetConfig(config)

	logger := logging.NewLogger(os.Stdout, os.Stderr)
	logging.SetLogger(logger)
}

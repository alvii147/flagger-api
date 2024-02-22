package testkitinternal

import (
	"github.com/alvii147/flagger-api/internal/env"
)

// SetupTests prepares testing environment.
func SetupTests() {
	config := env.NewConfig()
	env.SetConfig(config)
}

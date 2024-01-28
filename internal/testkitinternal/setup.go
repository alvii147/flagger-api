package testkitinternal

import (
	"fmt"
	"os"

	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/pkg/logging"
	"github.com/alvii147/flagger-api/pkg/mailclient"
)

// SetupTests prepares testing environment.
func SetupTests() {
	config := env.NewConfig()
	env.SetConfig(config)

	logger := logging.NewLogger(os.Stdout, os.Stderr)
	logging.SetLogger(logger)

	mailClient, err := mailclient.NewMailClient(
		config.MailClientType,
		config.SMTPHostname,
		config.SMTPPort,
		config.SMTPUsername,
		config.SMTPPassword,
		config.MailTemplatesDir,
	)
	if err != nil {
		panic(fmt.Sprintf("SetupTests failed to mail.NewMailClient: %v", err))
	}
	mailclient.SetMailClient(mailClient)
}

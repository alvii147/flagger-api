package mailclient

import (
	"fmt"
	"html/template"
	"io"
)

// consoleMailClient implements a MailClient that prints email contents to the console.
// This should typically be used in local development.
type consoleMailClient struct {
	username string
	writer   io.Writer
	tmpl     *template.Template
}

// NewConsoleMailClient returns a new consoleMailClient.
func NewConsoleMailClient(username string, writer io.Writer, templatesDir string) (*consoleMailClient, error) {
	tmplDirContents := templatesDir + "/*"
	tmpl, err := template.ParseGlob(tmplDirContents)
	if err != nil {
		return nil, fmt.Errorf("NewConsoleMailClient failed to template.ParseGlob %s: %w", tmplDirContents, err)
	}

	mailClient := &consoleMailClient{
		username: username,
		writer:   writer,
		tmpl:     tmpl,
	}

	return mailClient, nil
}

// Send prints email body to the console.
func (cmc *consoleMailClient) Send(to []string, subject string, textTemplate string, htmlTemplate string, templateData any) error {
	msg, err := BuildMail(cmc.username, to, subject, textTemplate, htmlTemplate, templateData, cmc.tmpl)
	if err != nil {
		return fmt.Errorf("Send failed to BuildMail: %w", err)
	}

	fmt.Fprint(cmc.writer, string(msg))

	return nil
}

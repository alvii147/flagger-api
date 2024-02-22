package mailclient

import (
	"fmt"
	htmltemplate "html/template"
	"io"
	texttemplate "text/template"
)

// consoleMailClient implements a MailClient that prints email contents to the console.
// This should typically be used in local development.
type consoleMailClient struct {
	username string
	writer   io.Writer
}

// NewConsoleMailClient returns a new consoleMailClient.
func NewConsoleMailClient(username string, writer io.Writer) *consoleMailClient {
	return &consoleMailClient{
		username: username,
		writer:   writer,
	}
}

// Send prints email body to the console.
func (cmc *consoleMailClient) Send(
	to []string,
	subject string,
	textTmpl *texttemplate.Template,
	htmlTmpl *htmltemplate.Template,
	tmplData interface{},
) error {
	msg, err := BuildMail(cmc.username, to, subject, textTmpl, htmlTmpl, tmplData)
	if err != nil {
		return fmt.Errorf("Send failed to BuildMail: %w", err)
	}

	fmt.Fprint(cmc.writer, string(msg))

	return nil
}

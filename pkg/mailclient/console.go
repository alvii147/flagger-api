package mailclient

import (
	"fmt"
	htmltemplate "html/template"
	"io"
	texttemplate "text/template"
)

// consoleClient implements a Client that prints email contents to the console.
// This should typically be used in local development.
type consoleClient struct {
	username string
	writer   io.Writer
}

// NewConsoleClient returns a new consoleClient.
func NewConsoleClient(username string, writer io.Writer) *consoleClient {
	return &consoleClient{
		username: username,
		writer:   writer,
	}
}

// Send prints email body to the console.
func (cmc *consoleClient) Send(
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

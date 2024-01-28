package mailclient

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

// mail client types.
const (
	MailClientTypeSMTP     = "smtp"
	MailClientTypeInMemory = "inmem"
	MailClientTypeConsole  = "console"
)

// MailClient is used to handle sending of emails.
type MailClient interface {
	SendMail(to []string, subject string, textTemplate string, htmlTemplate string, templateData any) error
}

var mc MailClient

// GetMailClient gets the currently set MailClient.
func GetMailClient() MailClient {
	return mc
}

// SetMailClient sets the current MailClient.
func SetMailClient(newClient MailClient) {
	mc = newClient
}

// NewMailClient returns a new MailClient.
func NewMailClient(clientType string, hostname string, port int, username string, password string, templatesDir string) (MailClient, error) {
	switch clientType {
	case MailClientTypeSMTP:
		return NewSMTPMailClient(hostname, port, username, password, templatesDir)
	case MailClientTypeInMemory:
		return NewInMemMailClient("support@flagger.com", templatesDir)
	case MailClientTypeConsole:
		return NewConsoleMailClient("support@flagger.com", os.Stdout, templatesDir)
	default:
		return nil, fmt.Errorf("NewMailClient failed, unknown mail client type %s", clientType)
	}
}

// BuildMail builds multi-line email body using MIME format.
func BuildMail(
	from string,
	to []string,
	subject string,
	textTemplate string,
	htmlTemplate string,
	templateData interface{},
	templateGlob *template.Template,
) ([]byte, error) {
	boundary := uuid.NewString()

	var mailBody bytes.Buffer
	mailBody.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\n", boundary))
	mailBody.WriteString("MIME-Version: 1.0\n")
	mailBody.WriteString(fmt.Sprintf("Subject: %s\n", subject))
	mailBody.WriteString(fmt.Sprintf("From: %s\n", from))
	mailBody.WriteString(fmt.Sprintf("From: %s\n", strings.Join(to, ", ")))
	mailBody.WriteString(fmt.Sprintf("Date: %s\n", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 -0700")))

	mailBody.WriteString("\n")

	mailBody.WriteString(fmt.Sprintf("--%s\n", boundary))
	mailBody.WriteString("Content-Type: text/plain; charset=\"utf-8\"\n")
	mailBody.WriteString("MIME-Version: 1.0\n")

	mailBody.WriteString("\n")

	err := templateGlob.ExecuteTemplate(&mailBody, textTemplate, templateData)
	if err != nil {
		return nil, fmt.Errorf("BuildMail failed to templateGlob.ExecuteTemplate %s: %w", textTemplate, err)
	}

	mailBody.WriteString("\n")

	mailBody.WriteString(fmt.Sprintf("--%s\n", boundary))
	mailBody.WriteString("Content-Type: text/html; charset=\"utf-8\"\n")
	mailBody.WriteString("MIME-Version: 1.0\n")

	mailBody.WriteString("\n")

	err = templateGlob.ExecuteTemplate(&mailBody, htmlTemplate, templateData)
	if err != nil {
		return nil, fmt.Errorf("BuildMail failed to templateGlob.ExecuteTemplate %s: %w", htmlTemplate, err)
	}

	mailBody.WriteString("\n")

	mailBody.WriteString(fmt.Sprintf("--%s--\n", boundary))

	msg := mailBody.Bytes()

	return msg, nil
}

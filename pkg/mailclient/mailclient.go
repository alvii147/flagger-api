package mailclient

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"
	"time"

	"github.com/google/uuid"
)

// mail client types.
const (
	ClientTypeSMTP     = "smtp"
	ClientTypeInMemory = "inmem"
	ClientTypeConsole  = "console"
)

// Client is used to handle sending of emails.
type Client interface {
	Send(
		to []string,
		subject string,
		textTmpl *texttemplate.Template,
		htmlTmpl *htmltemplate.Template,
		tmplData any,
	) error
}

// BuildMail builds multi-line email body using MIME format.
func BuildMail(
	from string,
	to []string,
	subject string,
	textTmpl *texttemplate.Template,
	htmlTmpl *htmltemplate.Template,
	tmplData any,
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

	err := textTmpl.Execute(&mailBody, tmplData)
	if err != nil {
		return nil, fmt.Errorf("BuildMail failed to textTmpl.Execute: %w", err)
	}

	mailBody.WriteString("\n")

	mailBody.WriteString(fmt.Sprintf("--%s\n", boundary))
	mailBody.WriteString("Content-Type: text/html; charset=\"utf-8\"\n")
	mailBody.WriteString("MIME-Version: 1.0\n")

	mailBody.WriteString("\n")

	err = htmlTmpl.Execute(&mailBody, tmplData)
	if err != nil {
		return nil, fmt.Errorf("BuildMail failed to htmlTmpl.Execute: %w", err)
	}

	mailBody.WriteString("\n")

	mailBody.WriteString(fmt.Sprintf("--%s--\n", boundary))

	msg := mailBody.Bytes()

	return msg, nil
}

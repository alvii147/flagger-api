package mailclient

import (
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
	"time"
)

// inMemMailLogEntry represents an in-memory entry of an email event.
type inMemMailLogEntry struct {
	From    string
	To      []string
	Subject string
	Message []byte
	SentAt  time.Time
}

// inMemMailClient implements a MailClient that saves email data in local memory.
// This should typically be used in unit tests.
type inMemMailClient struct {
	username string
	MailLogs []inMemMailLogEntry
}

// NewInMemMailClient returns a new inMemMailClient.
func NewInMemMailClient(username string) *inMemMailClient {
	return &inMemMailClient{
		username: username,
	}
}

// Send adds an email event to in-memory storage.
func (immc *inMemMailClient) Send(
	to []string,
	subject string,
	textTmpl *texttemplate.Template,
	htmlTmpl *htmltemplate.Template,
	tmplData interface{},
) error {
	msg, err := BuildMail(immc.username, to, subject, textTmpl, htmlTmpl, tmplData)
	if err != nil {
		return fmt.Errorf("Send failed to BuildMail: %w", err)
	}

	immc.MailLogs = append(
		immc.MailLogs,
		inMemMailLogEntry{
			From:    immc.username,
			To:      to,
			Subject: subject,
			Message: msg,
			SentAt:  time.Now().UTC(),
		},
	)

	return nil
}

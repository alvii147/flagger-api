package mailclient

import (
	"fmt"
	"html/template"
	"time"
)

// inMemMailClient implements a MailClient that saves email data in local memory.
// This should typically be used in unit tests.
type inMemMailClient struct {
	username string
	tmpl     *template.Template
}

// inMemMailLogEntry represents an in-memory entry of an email event.
type inMemMailLogEntry struct {
	From    string
	To      []string
	Subject string
	Message []byte
	SentAt  time.Time
}

var inMemMailLogs = []inMemMailLogEntry{}

// GetInMemMailLogs gets the current set of in-memory MailLogs.
func GetInMemMailLogs() []inMemMailLogEntry {
	return inMemMailLogs
}

// NewInMemMailClient returns a new inMemMailClient.
func NewInMemMailClient(username string, templatesDir string) (MailClient, error) {
	tmplDirContents := "../../" + templatesDir + "/*"
	tmpl, err := template.ParseGlob(tmplDirContents)
	if err != nil {
		return nil, fmt.Errorf("NewInMemMailClient failed to template.ParseGlob %s: %w", tmplDirContents, err)
	}

	mailClient := &inMemMailClient{
		username: username,
		tmpl:     tmpl,
	}

	return mailClient, nil
}

// SendMail adds an email event to in-memory storage.
func (immc *inMemMailClient) SendMail(to []string, subject string, textTemplate string, htmlTemplate string, templateData any) error {
	msg, err := BuildMail(immc.username, to, subject, textTemplate, htmlTemplate, templateData, immc.tmpl)
	if err != nil {
		return fmt.Errorf("SendMail failed to BuildMail: %w", err)
	}

	inMemMailLogs = append(
		inMemMailLogs,
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

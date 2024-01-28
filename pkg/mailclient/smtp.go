package mailclient

import (
	"fmt"
	"html/template"
	"net/smtp"
)

// smtpMailClient implements a MailClient that sends email through SMTP server.
// This should typically be used in production.
type smtpMailClient struct {
	hostname string
	addr     string
	username string
	password string
	tmpl     *template.Template
}

// NewSMTPMailClient returns a new smtpMailClient.
func NewSMTPMailClient(hostname string, port int, username string, password string, templatesDir string) (MailClient, error) {
	tmplDirContents := templatesDir + "/*"
	tmpl, err := template.ParseGlob(tmplDirContents)
	if err != nil {
		return nil, fmt.Errorf("NewSMTPMailClient failed to template.ParseGlob %s: %w", tmplDirContents, err)
	}

	return &smtpMailClient{
		hostname: hostname,
		addr:     fmt.Sprintf("%s:%d", hostname, port),
		username: username,
		password: password,
		tmpl:     tmpl,
	}, nil
}

// SendMail sends an email through SMTP server.
func (smc *smtpMailClient) SendMail(to []string, subject string, textTemplate string, htmlTemplate string, templateData interface{}) error {
	msg, err := BuildMail(smc.username, to, subject, textTemplate, htmlTemplate, templateData, smc.tmpl)
	if err != nil {
		return fmt.Errorf("SendMail failed to BuildMail: %w", err)
	}

	auth := smtp.PlainAuth("", smc.username, smc.password, smc.hostname)
	err = smtp.SendMail(smc.addr, auth, smc.username, to, msg)
	if err != nil {
		return fmt.Errorf("SendMail failed to smtp.SendMail: %w", err)
	}

	return nil
}

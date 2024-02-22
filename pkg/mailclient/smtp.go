package mailclient

import (
	"fmt"
	htmltemplate "html/template"
	"net/smtp"
	texttemplate "text/template"
)

// smtpClient implements a Client that sends email through SMTP server.
// This should typically be used in production.
type smtpClient struct {
	hostname string
	addr     string
	username string
	password string
}

// NewSMTPClient returns a new smtpClient.
func NewSMTPClient(hostname string, port int, username string, password string) *smtpClient {
	return &smtpClient{
		hostname: hostname,
		addr:     fmt.Sprintf("%s:%d", hostname, port),
		username: username,
		password: password,
	}
}

// Send sends an email through SMTP server.
func (smc *smtpClient) Send(
	to []string,
	subject string,
	textTmpl *texttemplate.Template,
	htmlTmpl *htmltemplate.Template,
	tmplData interface{},
) error {
	msg, err := BuildMail(smc.username, to, subject, textTmpl, htmlTmpl, tmplData)
	if err != nil {
		return fmt.Errorf("Send failed to BuildMail: %w", err)
	}

	auth := smtp.PlainAuth("", smc.username, smc.password, smc.hostname)
	err = smtp.SendMail(smc.addr, auth, smc.username, to, msg)
	if err != nil {
		return fmt.Errorf("Send failed to smtp.SendMail: %w", err)
	}

	return nil
}

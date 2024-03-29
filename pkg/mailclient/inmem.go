package mailclient

import (
	"fmt"
	htmltemplate "html/template"
	texttemplate "text/template"
	"time"
)

// logEntry represents an in-memory entry of an email event.
type logEntry struct {
	From    string
	To      []string
	Subject string
	Message []byte
	SentAt  time.Time
}

// inMemClient implements a Client that saves email data in local memory.
// This should typically be used in unit tests.
type inMemClient struct {
	username string
	sendErr  error
	Logs     []logEntry
}

// NewInMemClient returns a new inMemClient.
func NewInMemClient(username string) *inMemClient {
	return &inMemClient{
		username: username,
		sendErr:  nil,
		Logs:     make([]logEntry, 0),
	}
}

func (immc *inMemClient) SetSendError(err error) {
	immc.sendErr = err
}

// Send adds an email event to in-memory storage.
func (immc *inMemClient) Send(
	to []string,
	subject string,
	textTmpl *texttemplate.Template,
	htmlTmpl *htmltemplate.Template,
	tmplData any,
) error {
	if immc.sendErr != nil {
		return immc.sendErr
	}

	msg, err := BuildMail(immc.username, to, subject, textTmpl, htmlTmpl, tmplData)
	if err != nil {
		return fmt.Errorf("Send failed to BuildMail: %w", err)
	}

	immc.Logs = append(
		immc.Logs,
		logEntry{
			From:    immc.username,
			To:      to,
			Subject: subject,
			Message: msg,
			SentAt:  time.Now().UTC(),
		},
	)

	return nil
}

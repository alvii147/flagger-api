package mailclient_test

import (
	"errors"
	htmltemplate "html/template"
	"testing"
	texttemplate "text/template"
	"time"

	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestInMemMailClientSend(t *testing.T) {
	t.Parallel()

	username := testkit.GenerateFakeEmail()
	client := mailclient.NewInMemClient(username)

	mailCount := len(client.Logs)

	to := testkit.GenerateFakeEmail()
	subject := testkit.MustGenerateRandomString(12, true, true, true)
	textTmpl, err := texttemplate.New("textTmpl").Parse("Test Template Content: {{ .Value }}")
	require.NoError(t, err)
	htmlTmpl, err := htmltemplate.New("htmlTmpl").Parse("<div>Test Template Content: {{ .Value }}</div>")
	require.NoError(t, err)
	tmplData := map[string]int{
		"Value": 42,
	}

	err = client.Send([]string{to}, subject, textTmpl, htmlTmpl, tmplData)
	require.NoError(t, err)

	require.Len(t, client.Logs, mailCount+1)

	lastMail := client.Logs[len(client.Logs)-1]
	require.Equal(t, username, lastMail.From)
	require.Equal(t, []string{to}, lastMail.To)
	require.Equal(t, subject, lastMail.Subject)
	testkit.RequireTimeAlmostEqual(t, time.Now().UTC(), lastMail.SentAt)

	textMsg, htmlMsg := testkit.MustParseMailMessage(string(lastMail.Message))

	require.Regexp(t, `Content-Type:\s*text\/plain;\s*charset\s*=\s*"utf-8"`, textMsg)
	require.Regexp(t, `MIME-Version:\s*1.0`, textMsg)
	require.Contains(t, textMsg, "Test Template Content: 42")

	require.Regexp(t, `Content-Type:\s*text\/html;\s*charset\s*=\s*"utf-8"`, htmlMsg)
	require.Regexp(t, `MIME-Version:\s*1.0`, htmlMsg)
	require.Contains(t, htmlMsg, "Test Template Content: 42")
}

func TestInMemMailClientSetSendError(t *testing.T) {
	t.Parallel()

	username := testkit.GenerateFakeEmail()
	client := mailclient.NewInMemClient(username)
	client.SetSendError(errors.New("Send failed"))

	mailCount := len(client.Logs)

	to := testkit.GenerateFakeEmail()
	subject := testkit.MustGenerateRandomString(12, true, true, true)
	textTmpl, err := texttemplate.New("textTmpl").Parse("Test Template Content: {{ .Value }}")
	require.NoError(t, err)
	htmlTmpl, err := htmltemplate.New("htmlTmpl").Parse("<div>Test Template Content: {{ .Value }}</div>")
	require.NoError(t, err)
	tmplData := map[string]int{
		"Value": 42,
	}

	err = client.Send([]string{to}, subject, textTmpl, htmlTmpl, tmplData)
	require.Error(t, err)

	require.Len(t, client.Logs, mailCount)
}

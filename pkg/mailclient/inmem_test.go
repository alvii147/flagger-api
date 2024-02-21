package mailclient_test

import (
	"testing"
	"time"

	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestInMemMailClient(t *testing.T) {
	t.Parallel()

	username := testkit.GenerateFakeEmail()
	mailClient, err := mailclient.NewInMemMailClient(username, "pkg/mailclient")
	require.NoError(t, err)

	mailCount := len(mailClient.MailLogs)

	to := testkit.GenerateFakeEmail()
	subject := testkit.MustGenerateRandomString(12, true, true, true)
	textTemplate := "tmpl.txt"
	htmlTemplate := "tmpl.html"
	templateData := map[string]int{
		"Value": 42,
	}

	err = mailClient.Send([]string{to}, subject, textTemplate, htmlTemplate, templateData)
	require.NoError(t, err)

	require.Len(t, mailClient.MailLogs, mailCount+1)

	lastMail := mailClient.MailLogs[len(mailClient.MailLogs)-1]
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

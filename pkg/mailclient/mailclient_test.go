package mailclient_test

import (
	"html/template"
	"testing"

	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestBuildMail(t *testing.T) {
	t.Parallel()

	from := testkit.GenerateFakeEmail()
	to := testkit.GenerateFakeEmail()
	subject := testkit.MustGenerateRandomString(12, true, true, true)
	textTemplate := "tmpl.txt"
	htmlTemplate := "tmpl.html"
	templateData := map[string]int{
		"Value": 42,
	}
	templateGlob, err := template.ParseGlob("./*")
	require.NoError(t, err)

	msg, err := mailclient.BuildMail(
		from,
		[]string{to},
		subject,
		textTemplate,
		htmlTemplate,
		templateData,
		templateGlob,
	)
	require.NoError(t, err)

	textMsg, htmlMsg := testkit.MustParseMailMessage(string(msg))

	require.Regexp(t, `Content-Type:\s*text\/plain;\s*charset\s*=\s*"utf-8"`, textMsg)
	require.Regexp(t, `MIME-Version:\s*1.0`, textMsg)
	require.Contains(t, textMsg, "Test Template Content: 42")

	require.Regexp(t, `Content-Type:\s*text\/html;\s*charset\s*=\s*"utf-8"`, htmlMsg)
	require.Regexp(t, `MIME-Version:\s*1.0`, htmlMsg)
	require.Contains(t, htmlMsg, "Test Template Content: 42")
}

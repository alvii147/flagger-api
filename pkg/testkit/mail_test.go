package testkit_test

import (
	"testing"

	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestMustParseMailMessageSuccess(t *testing.T) {
	t.Parallel()

	msg := `Content-Type: multipart/alternative; boundary="deadbeef"
MIME-Version: 1.0
Subject: 1L9cBLMEBzSn
From: vfgtd7ujt535@ucgufkizok.bih
From: y32y4v6iyx5i@lyijjmvasg.tcn
Date: Thu, 25 Jan 2024 14:11:10 +0000

--deadbeef
Content-Type: text/plain; charset="utf-8"
MIME-Version: 1.0

Text Message
--deadbeef
Content-Type: text/html; charset="utf-8"
MIME-Version: 1.0

HTML Message
--deadbeef--
	`

	textMsg, htmlMsg := testkit.MustParseMailMessage(msg)
	require.Contains(t, textMsg, "Text Message")
	require.NotContains(t, textMsg, "HTML Message")
	require.Contains(t, htmlMsg, "HTML Message")
	require.NotContains(t, htmlMsg, "Text Message")
}

func TestMustParseMailMessageError(t *testing.T) {
	t.Parallel()

	invalidMsg := "1nv4l1d m3554g3"
	msgWithInvalidContentType := `Content-Type: invalid/type; boundary="deadbeef"
MIME-Version: 1.0
Subject: 1L9cBLMEBzSn
From: vfgtd7ujt535@ucgufkizok.bih
From: y32y4v6iyx5i@lyijjmvasg.tcn
Date: Thu, 25 Jan 2024 14:11:10 +0000

--deadbeef
Content-Type: text/plain; charset="utf-8"
MIME-Version: 1.0

Text Message
--deadbeef
Content-Type: text/html; charset="utf-8"
MIME-Version: 1.0

HTML Message
--deadbeef--
	`
	msgWithOneSection := `Content-Type: multipart/alternative; boundary="deadbeef"
MIME-Version: 1.0
Subject: 1L9cBLMEBzSn
From: vfgtd7ujt535@ucgufkizok.bih
From: y32y4v6iyx5i@lyijjmvasg.tcn
Date: Thu, 25 Jan 2024 14:11:10 +0000

--deadbeef
Content-Type: text/plain; charset="utf-8"
MIME-Version: 1.0

Text Message
--deadbeef
	`

	testcases := []struct {
		name string
		msg  string
	}{
		{
			name: "Invalid message",
			msg:  invalidMsg,
		},
		{
			name: "Message with invalid content type",
			msg:  msgWithInvalidContentType,
		},
		{
			name: "Message with one section",
			msg:  msgWithOneSection,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				require.NotNil(t, r)
			}()

			testkit.MustParseMailMessage(testcase.msg)
		})
	}
}

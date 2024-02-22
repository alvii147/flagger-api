package logging_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	t.Parallel()

	bufOut, bufErr, logger := testkit.CreateTestLogger()

	debugMessage := "Debug message"
	infoMessage := "Info message"
	warnMessage := "Warn message"
	errorMessage := "Error message"

	logger.LogDebug(debugMessage)
	logger.LogInfo(infoMessage)
	logger.LogWarn(warnMessage)
	logger.LogError(errorMessage)

	stdoutMessages := strings.Split(strings.TrimSpace(bufOut.String()), "\n")
	stderrMessages := strings.Split(strings.TrimSpace(bufErr.String()), "\n")

	require.Len(t, stdoutMessages, 3)
	require.Len(t, stderrMessages, 1)

	testcases := []struct {
		name            string
		capturedMessage string
		wantMessage     string
		wantLevel       string
	}{
		{
			name:            debugMessage,
			capturedMessage: stdoutMessages[0],
			wantMessage:     debugMessage,
			wantLevel:       "D",
		},
		{
			name:            infoMessage,
			capturedMessage: stdoutMessages[1],
			wantMessage:     infoMessage,
			wantLevel:       "I",
		},
		{
			name:            warnMessage,
			capturedMessage: stdoutMessages[2],
			wantMessage:     warnMessage,
			wantLevel:       "W",
		},
		{
			name:            errorMessage,
			capturedMessage: stderrMessages[0],
			wantMessage:     errorMessage,
			wantLevel:       "E",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			logLevel, logTime, logFile, logMsg := testkit.MustParseLogMessage(testcase.capturedMessage)
			require.Equal(t, testcase.wantLevel, logLevel)
			testkit.RequireTimeAlmostEqual(t, logTime, time.Now().UTC())
			require.Contains(t, logFile, "pkg/logging/logging_test.go")
			require.Equal(t, testcase.wantMessage, logMsg)
		})
		break
	}
}

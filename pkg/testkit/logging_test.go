package testkit_test

import (
	"testing"
	"time"

	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestMustParseLogMessageSuccess(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name      string
		rawMsg    string
		wantLevel string
		wantTime  time.Time
		wantFile  string
		wantMsg   string
	}{
		{
			name:      "Info message",
			rawMsg:    "[I] 2016/08/19 16:03:46 /file/path:30 0xDEADBEEF",
			wantLevel: "I",
			wantTime:  time.Date(2016, 8, 19, 16, 3, 46, 0, time.UTC),
			wantFile:  "/file/path",
			wantMsg:   "0xDEADBEEF",
		},
		{
			name:      "Warning message with irregular spacing",
			rawMsg:    "[W]  2016/08/19     16:03:46 /file/path:30             0x DEAD BEEF",
			wantLevel: "W",
			wantTime:  time.Date(2016, 8, 19, 16, 3, 46, 0, time.UTC),
			wantFile:  "/file/path",
			wantMsg:   "0x DEAD BEEF",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			logLevel, logTime, logFile, logMsg := testkit.MustParseLogMessage(testcase.rawMsg)
			require.Equal(t, testcase.wantLevel, logLevel)
			require.Equal(t, testcase.wantTime, logTime)
			require.Contains(t, logFile, testcase.wantFile)
			require.Equal(t, testcase.wantMsg, logMsg)
		})
	}
}

func TestMustParseLogMessageError(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name      string
		rawMsg    string
		wantLevel string
		wantTime  time.Time
		wantFile  string
		wantMsg   string
	}{
		{
			name:   "Invalid message",
			rawMsg: "1nv4l1d m3554g3",
		},
		{
			name:   "Invalid level",
			rawMsg: "[C] 2016/08/19 16:03:46 /file/path:30 0xDEADBEEF",
		},
		{
			name:   "Invalid time",
			rawMsg: "[I] 2016/31/42 28:67:82 /file/path:30 0xDEADBEEF",
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

			testkit.MustParseLogMessage(testcase.rawMsg)
		})
	}
}

func TestCreateTestLogger(t *testing.T) {
	t.Parallel()

	debugMessage := "Debug message"
	infoMessage := "Info message"
	warnMessage := "Warn message"
	errorMessage := "Error message"

	bufOut, bufErr, logger := testkit.CreateTestLogger()

	logger.LogDebug(debugMessage)
	logger.LogInfo(infoMessage)
	logger.LogWarn(warnMessage)
	logger.LogError(errorMessage)

	stdout := bufOut.String()
	stderr := bufErr.String()

	require.Contains(t, stdout, debugMessage)
	require.Contains(t, stdout, infoMessage)
	require.Contains(t, stdout, warnMessage)
	require.Contains(t, stderr, errorMessage)
}

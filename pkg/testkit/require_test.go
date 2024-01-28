package testkit_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

type MockTestingT struct {
	Failed    bool
	FailedNow bool
	Logs      []string
}

func NewMockTestingT() *MockTestingT {
	return &MockTestingT{
		Failed:    false,
		FailedNow: false,
		Logs:      make([]string, 0),
	}
}

func (t *MockTestingT) Logf(format string, args ...interface{}) {
	t.Logs = append(t.Logs, fmt.Sprintf(format, args...))
}

func (t *MockTestingT) Fail() {
	t.Failed = true
}

func (t *MockTestingT) FailNow() {
	t.FailedNow = true
}

func (t *MockTestingT) Errorf(format string, args ...interface{}) {
	t.Logf(format, args...)
	t.Fail()
}

func TestRequireTimeAlmostEqual(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name           string
		expectedTime   time.Time
		actualTime     time.Time
		wantFail       bool
		wantFailNow    bool
		wantLog        bool
		wantLogMessage string
	}{
		{
			name:           "Equal times",
			expectedTime:   time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
			actualTime:     time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
			wantFail:       false,
			wantFailNow:    false,
			wantLog:        false,
			wantLogMessage: "",
		},
		{
			name:           "Actual time is one second after expected time",
			expectedTime:   time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
			actualTime:     time.Date(2018, 9, 6, 6, 34, 9, 0, time.UTC),
			wantFail:       false,
			wantFailNow:    false,
			wantLog:        false,
			wantLogMessage: "",
		},
		{
			name:           "Actual time is one second before expected time",
			expectedTime:   time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
			actualTime:     time.Date(2018, 9, 6, 6, 34, 7, 0, time.UTC),
			wantFail:       false,
			wantFailNow:    false,
			wantLog:        false,
			wantLogMessage: "",
		},
		{
			name:           "Actual time is three seconds after expected time",
			expectedTime:   time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
			actualTime:     time.Date(2018, 9, 6, 6, 34, 11, 0, time.UTC),
			wantFail:       true,
			wantFailNow:    true,
			wantLog:        true,
			wantLogMessage: "actual time 2018-09-06T06:34:11Z occurs more than 2s after expected time 2018-09-06T06:34:08Z",
		},
		{
			name:           "Actual time is three seconds before expected time",
			expectedTime:   time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
			actualTime:     time.Date(2018, 9, 6, 6, 34, 5, 0, time.UTC),
			wantFail:       true,
			wantFailNow:    true,
			wantLog:        true,
			wantLogMessage: "actual time 2018-09-06T06:34:05Z occurs more than 2s before expected time 2018-09-06T06:34:08Z",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			mockT := NewMockTestingT()
			testkit.RequireTimeAlmostEqual(mockT, testcase.expectedTime, testcase.actualTime)

			require.Equal(t, testcase.wantFail, mockT.Failed)
			require.Equal(t, testcase.wantFailNow, mockT.FailedNow)

			errLogs := mockT.Logs
			if testcase.wantLog {
				require.GreaterOrEqual(t, len(errLogs), 1)

				foundLog := false
				for _, errLog := range errLogs {
					if strings.Contains(errLog, testcase.wantLogMessage) {
						foundLog = true
						break
					}
				}

				require.True(t, foundLog, fmt.Sprintf("failed to find \"%s\" in %v", testcase.wantLogMessage, errLogs))
			}
		})
	}
}

func TestRequirePGTimestampAlmostEqual(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name           string
		expectedTime   pgtype.Timestamp
		actualTime     pgtype.Timestamp
		wantFail       bool
		wantFailNow    bool
		wantLog        bool
		wantLogMessage string
	}{
		{
			name: "Equal times",
			expectedTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
				Valid: true,
			},
			actualTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
				Valid: true,
			},
			wantFail:       false,
			wantFailNow:    false,
			wantLog:        false,
			wantLogMessage: "",
		},
		{
			name: "Actual time is one second after expected time",
			expectedTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
				Valid: true,
			},
			actualTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 9, 0, time.UTC),
				Valid: true,
			},
			wantFail:       false,
			wantFailNow:    false,
			wantLog:        false,
			wantLogMessage: "",
		},
		{
			name: "Actual time is one second before expected time",
			expectedTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
				Valid: true,
			},
			actualTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 7, 0, time.UTC),
				Valid: true,
			},
			wantFail:       false,
			wantFailNow:    false,
			wantLog:        false,
			wantLogMessage: "",
		},
		{
			name: "Actual time is three seconds after expected time",
			expectedTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
				Valid: true,
			},
			actualTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 11, 0, time.UTC),
				Valid: true,
			},
			wantFail:       true,
			wantFailNow:    true,
			wantLog:        true,
			wantLogMessage: "actual time 2018-09-06T06:34:11Z occurs more than 2s after expected time 2018-09-06T06:34:08Z",
		},
		{
			name: "Actual time is three seconds before expected time",
			expectedTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
				Valid: true,
			},
			actualTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 5, 0, time.UTC),
				Valid: true,
			},
			wantFail:       true,
			wantFailNow:    true,
			wantLog:        true,
			wantLogMessage: "actual time 2018-09-06T06:34:05Z occurs more than 2s before expected time 2018-09-06T06:34:08Z",
		},
		{
			name: "Actual time is valid, expected time is invalid",
			expectedTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
				Valid: false,
			},
			actualTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
				Valid: true,
			},
			wantFail:       true,
			wantFailNow:    true,
			wantLog:        false,
			wantLogMessage: "",
		},
		{
			name: "Actual time is invalid, expected time is valid",
			expectedTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
				Valid: true,
			},
			actualTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
				Valid: false,
			},
			wantFail:       true,
			wantFailNow:    true,
			wantLog:        false,
			wantLogMessage: "",
		},
		{
			name: "Both times are invalid",
			expectedTime: pgtype.Timestamp{
				Time:  time.Date(1996, 9, 6, 6, 34, 8, 0, time.UTC),
				Valid: false,
			},
			actualTime: pgtype.Timestamp{
				Time:  time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC),
				Valid: false,
			},
			wantFail:       false,
			wantFailNow:    false,
			wantLog:        false,
			wantLogMessage: "",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			mockT := NewMockTestingT()
			testkit.RequirePGTimestampAlmostEqual(mockT, testcase.expectedTime, testcase.actualTime)

			require.Equal(t, testcase.wantFail, mockT.Failed)
			require.Equal(t, testcase.wantFailNow, mockT.FailedNow)

			errLogs := mockT.Logs
			if testcase.wantLog {
				require.GreaterOrEqual(t, len(errLogs), 1)

				foundLog := false
				for _, errLog := range errLogs {
					if strings.Contains(errLog, testcase.wantLogMessage) {
						foundLog = true
						break
					}
				}

				require.True(t, foundLog, fmt.Sprintf("failed to find \"%s\" in %v", testcase.wantLogMessage, errLogs))
			}
		})
	}
}

package testkit_test

import (
	"fmt"
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

	testTime := time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC)
	testcases := []struct {
		name         string
		expectedTime time.Time
		actualTime   time.Time
		wantFail     bool
		wantFailNow  bool
		wantLog      bool
	}{
		{
			name:         "Equal times",
			expectedTime: testTime,
			actualTime:   testTime,
			wantFail:     false,
			wantFailNow:  false,
			wantLog:      false,
		},
		{
			name:         "Actual time is one second after expected time",
			expectedTime: testTime,
			actualTime:   testTime.Add(time.Second),
			wantFail:     false,
			wantFailNow:  false,
			wantLog:      false,
		},
		{
			name:         "Actual time is one second before expected time",
			expectedTime: testTime,
			actualTime:   testTime.Add(-time.Second),
			wantFail:     false,
			wantFailNow:  false,
			wantLog:      false,
		},
		{
			name:         "Actual time is more than tolerance duration after expected time",
			expectedTime: testTime,
			actualTime:   testTime.Add(testkit.TimeEqualityTolerance + time.Second),
			wantFail:     true,
			wantFailNow:  true,
			wantLog:      true,
		},
		{
			name:         "Actual time is less than tolerance duration before expected time",
			expectedTime: testTime,
			actualTime:   testTime.Add(-testkit.TimeEqualityTolerance - time.Second),
			wantFail:     true,
			wantFailNow:  true,
			wantLog:      true,
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
		})
	}
}

func TestRequirePGTimestampAlmostEqual(t *testing.T) {
	t.Parallel()

	testTime := time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC)
	testcases := []struct {
		name         string
		expectedTime pgtype.Timestamp
		actualTime   pgtype.Timestamp
		wantFail     bool
		wantFailNow  bool
		wantLog      bool
	}{
		{
			name: "Equal times",
			expectedTime: pgtype.Timestamp{
				Time:  testTime,
				Valid: true,
			},
			actualTime: pgtype.Timestamp{
				Time:  testTime,
				Valid: true,
			},
			wantFail:    false,
			wantFailNow: false,
			wantLog:     false,
		},
		{
			name: "Actual time is one second after expected time",
			expectedTime: pgtype.Timestamp{
				Time:  testTime,
				Valid: true,
			},
			actualTime: pgtype.Timestamp{
				Time:  testTime.Add(time.Second),
				Valid: true,
			},
			wantFail:    false,
			wantFailNow: false,
			wantLog:     false,
		},
		{
			name: "Actual time is one second before expected time",
			expectedTime: pgtype.Timestamp{
				Time:  testTime,
				Valid: true,
			},
			actualTime: pgtype.Timestamp{
				Time:  testTime.Add(-time.Second),
				Valid: true,
			},
			wantFail:    false,
			wantFailNow: false,
			wantLog:     false,
		},
		{
			name: "Actual time is more than tolerance duration after expected time",
			expectedTime: pgtype.Timestamp{
				Time:  testTime,
				Valid: true,
			},
			actualTime: pgtype.Timestamp{
				Time:  testTime.Add(testkit.TimeEqualityTolerance + time.Second),
				Valid: true,
			},
			wantFail:    true,
			wantFailNow: true,
			wantLog:     true,
		},
		{
			name: "Actual time is less than tolerance duration before expected time",
			expectedTime: pgtype.Timestamp{
				Time:  testTime,
				Valid: true,
			},
			actualTime: pgtype.Timestamp{
				Time:  testTime.Add(-testkit.TimeEqualityTolerance - time.Second),
				Valid: true,
			},
			wantFail:    true,
			wantFailNow: true,
			wantLog:     true,
		},
		{
			name: "Actual time is valid, expected time is invalid",
			expectedTime: pgtype.Timestamp{
				Time:  testTime,
				Valid: false,
			},
			actualTime: pgtype.Timestamp{
				Time:  testTime,
				Valid: true,
			},
			wantFail:    true,
			wantFailNow: true,
			wantLog:     false,
		},
		{
			name: "Actual time is invalid, expected time is valid",
			expectedTime: pgtype.Timestamp{
				Time:  testTime,
				Valid: true,
			},
			actualTime: pgtype.Timestamp{
				Time:  testTime,
				Valid: false,
			},
			wantFail:    true,
			wantFailNow: true,
			wantLog:     false,
		},
		{
			name: "Both times are invalid",
			expectedTime: pgtype.Timestamp{
				Time:  testTime,
				Valid: false,
			},
			actualTime: pgtype.Timestamp{
				Time:  testTime.AddDate(1, 0, 0),
				Valid: false,
			},
			wantFail:    false,
			wantFailNow: false,
			wantLog:     false,
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
		})
	}
}

package testkit_test

import (
	"testing"
	"time"

	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestMockTestingTCleanup(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	cleanupFuncCalled := false
	cleanupFunc := func() {
		cleanupFuncCalled = true
	}

	mockT.Cleanup(cleanupFunc)
	require.False(t, mockT.HasFailed)
	require.False(t, cleanupFuncCalled)
	require.Len(t, mockT.Cleanups, 1)

	mockT.Cleanups[0]()
	require.True(t, cleanupFuncCalled)
}

func TestMockTestingTError(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	msg := "deadbeef"
	mockT.Error(msg)
	require.True(t, mockT.HasFailed)
	require.Len(t, mockT.Logs, 1)
	require.Equal(t, msg, mockT.Logs[0])
}

func TestMockTestingTErrorf(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	format := "d%ddb%df"
	args := []any{34, 33}
	msg := "d34db33f"

	mockT.Errorf(format, args...)
	require.True(t, mockT.HasFailed)
	require.Len(t, mockT.Logs, 1)
	require.Equal(t, msg, mockT.Logs[0])
}

func TestMockTestingTFail(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	mockT.Fail()
	require.True(t, mockT.HasFailed)
}

func TestMockTestingTFailNow(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	mockT.FailNow()
	require.True(t, mockT.HasFailed)
}

func TestMockTestingTFailed(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	mockT.HasFailed = true
	require.True(t, mockT.Failed())
}

func TestMockTestingTFatal(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	msg := "deadbeef"
	mockT.Fatal(msg)
	require.True(t, mockT.HasFailed)
	require.Len(t, mockT.Logs, 1)
	require.Equal(t, msg, mockT.Logs[0])
}

func TestMockTestingTFatalf(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	format := "d%ddb%df"
	args := []any{34, 33}
	msg := "d34db33f"

	mockT.Fatalf(format, args...)
	require.True(t, mockT.HasFailed)
	require.Len(t, mockT.Logs, 1)
	require.Equal(t, msg, mockT.Logs[0])
}

func TestMockTestingTLog(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	msg := "deadbeef"
	mockT.Log(msg)
	require.False(t, mockT.HasFailed)
	require.Len(t, mockT.Logs, 1)
	require.Equal(t, msg, mockT.Logs[0])
}

func TestMockTestingTLogf(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	format := "d%ddb%df"
	args := []any{34, 33}
	msg := "d34db33f"

	mockT.Logf(format, args...)
	require.False(t, mockT.HasFailed)
	require.Len(t, mockT.Logs, 1)
	require.Equal(t, msg, mockT.Logs[0])
}

func TestMockTestingTSkip(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	msg := "deadbeef"
	mockT.Skip(msg)
	require.True(t, mockT.HasSkipped)
	require.False(t, mockT.HasFailed)
	require.Len(t, mockT.Logs, 1)
	require.Equal(t, msg, mockT.Logs[0])
}

func TestMockTestingTSkipNow(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	mockT.SkipNow()
	require.True(t, mockT.HasSkipped)
	require.False(t, mockT.HasFailed)
}

func TestMockTestingTSkipf(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	format := "d%ddb%df"
	args := []any{34, 33}
	msg := "d34db33f"

	mockT.Skipf(format, args...)
	require.True(t, mockT.HasSkipped)
	require.False(t, mockT.HasFailed)
	require.Len(t, mockT.Logs, 1)
	require.Equal(t, msg, mockT.Logs[0])
}

func TestMockTestingTSkipped(t *testing.T) {
	mockT := testkit.NewMockTestingT()

	mockT.HasSkipped = true
	require.True(t, mockT.Skipped())
	require.False(t, mockT.HasFailed)
}

func TestRequireTimeAlmostEqual(t *testing.T) {
	t.Parallel()

	testTime := time.Date(2018, 9, 6, 6, 34, 8, 0, time.UTC)
	testcases := []struct {
		name         string
		expectedTime time.Time
		actualTime   time.Time
		wantFailed   bool
		wantLog      bool
	}{
		{
			name:         "Equal times",
			expectedTime: testTime,
			actualTime:   testTime,
			wantFailed:   false,
			wantLog:      false,
		},
		{
			name:         "Actual time is one second after expected time",
			expectedTime: testTime,
			actualTime:   testTime.Add(time.Second),
			wantFailed:   false,
			wantLog:      false,
		},
		{
			name:         "Actual time is one second before expected time",
			expectedTime: testTime,
			actualTime:   testTime.Add(-time.Second),
			wantFailed:   false,
			wantLog:      false,
		},
		{
			name:         "Actual time is more than tolerance duration after expected time",
			expectedTime: testTime,
			actualTime:   testTime.Add(testkit.TimeEqualityTolerance + time.Second),
			wantFailed:   true,
			wantLog:      true,
		},
		{
			name:         "Actual time is less than tolerance duration before expected time",
			expectedTime: testTime,
			actualTime:   testTime.Add(-testkit.TimeEqualityTolerance - time.Second),
			wantFailed:   true,
			wantLog:      true,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			mockT := testkit.NewMockTestingT()
			testkit.RequireTimeAlmostEqual(mockT, testcase.expectedTime, testcase.actualTime)

			require.Equal(t, testcase.wantFailed, mockT.Failed())
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
		wantFailed   bool
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
			wantFailed: false,
			wantLog:    false,
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
			wantFailed: false,
			wantLog:    false,
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
			wantFailed: false,
			wantLog:    false,
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
			wantFailed: true,
			wantLog:    true,
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
			wantFailed: true,
			wantLog:    true,
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
			wantFailed: true,
			wantLog:    false,
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
			wantFailed: true,
			wantLog:    false,
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
			wantFailed: false,
			wantLog:    false,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			mockT := testkit.NewMockTestingT()
			testkit.RequirePGTimestampAlmostEqual(mockT, testcase.expectedTime, testcase.actualTime)

			require.Equal(t, testcase.wantFailed, mockT.Failed())
		})
	}
}

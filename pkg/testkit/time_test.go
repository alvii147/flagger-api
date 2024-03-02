package testkit_test

import (
	"testing"
	"time"

	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

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

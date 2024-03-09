package testkit_test

import (
	"testing"
	"time"

	"github.com/alvii147/flagger-api/pkg/testkit"
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
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			mockT := testkit.NewMockTestingT()
			testkit.RequireTimeAlmostEqual(mockT, testcase.expectedTime, testcase.actualTime)

			require.Equal(t, testcase.wantFailed, mockT.Failed())
		})
	}
}

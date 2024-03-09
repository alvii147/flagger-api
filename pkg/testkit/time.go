package testkit

import (
	"fmt"
	"time"

	"github.com/stretchr/testify/require"
)

const TimeEqualityTolerance = 60 * time.Second

// RequireTimeAlmostEqual requires that expected and actual times
// are within at most a certain duration tolerance of each other.
func RequireTimeAlmostEqual(t TestingT, expected time.Time, actual time.Time) {
	expectedISO := expected.Format(time.RFC3339)
	actualISO := actual.Format(time.RFC3339)

	require.True(
		t,
		actual.After(expected.Add(-TimeEqualityTolerance)),
		fmt.Sprintf("actual time %s occurs more than %s before expected time %s", actualISO, TimeEqualityTolerance, expectedISO),
	)
	require.True(
		t,
		actual.Before(expected.Add(TimeEqualityTolerance)),
		fmt.Sprintf("actual time %s occurs more than %s after expected time %s", actualISO, TimeEqualityTolerance, expectedISO),
	)
}

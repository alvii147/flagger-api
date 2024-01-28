package testkit

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

const TimeEqualityTolerance = 2 * time.Second

// RequireTimeAlmostEqual requires that expected and actual times
// are within at most a certain duration tolerance of each other.
func RequireTimeAlmostEqual(t require.TestingT, expected time.Time, actual time.Time) {
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

// RequirePGTimestampAlmostEqual requires that expected and actual PostgreSQL timestamps
// are within at most 5 seconds of each other.
func RequirePGTimestampAlmostEqual(t require.TestingT, expected pgtype.Timestamp, actual pgtype.Timestamp) {
	require.Equal(t, expected.Valid, actual.Valid)

	if expected.Valid {
		RequireTimeAlmostEqual(t, expected.Time, actual.Time)
	}
}

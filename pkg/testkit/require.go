package testkit

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

const TimeEqualityTolerance = 60 * time.Second

// TestingT is an interface wrapper around *testing.T.
type TestingT interface {
	Cleanup(f func())
	Error(args ...any)
	Errorf(format string, args ...any)
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...any)
	Fatalf(format string, args ...any)
	Log(args ...any)
	Logf(format string, args ...any)
	Skip(args ...any)
	SkipNow()
	Skipf(format string, args ...any)
	Skipped() bool
}

// MockTestingT implements TestingT and is a mock implementation of testing.T.
type MockTestingT struct {
	Cleanups   []func()
	HasFailed  bool
	HasSkipped bool
	Logs       []string
}

// NewMockTestingT creates a new MockTestingT.
func NewMockTestingT() *MockTestingT {
	return &MockTestingT{
		Cleanups:   make([]func(), 0),
		HasFailed:  false,
		HasSkipped: false,
		Logs:       make([]string, 0),
	}
}

// Cleanup stores a cleanup function.
func (t *MockTestingT) Cleanup(f func()) {
	t.Cleanups = append(t.Cleanups, f)
}

// Error is equivalent to Log followed by Fail.
func (t *MockTestingT) Error(args ...any) {
	t.Log(args...)
	t.Fail()
}

// Errorf is equivalent to Logf followed by Fail.
func (t *MockTestingT) Errorf(format string, args ...any) {
	t.Logf(format, args...)
	t.Fail()
}

// Fail is equivalent to FailNow.
func (t *MockTestingT) Fail() {
	t.FailNow()
}

// FailNow records the test as failed.
func (t *MockTestingT) FailNow() {
	t.HasFailed = true
}

// Failed reports whether the test has failed.
func (t *MockTestingT) Failed() bool {
	return t.HasFailed
}

// Fatal is equivalent to Log followed by FailNow.
func (t *MockTestingT) Fatal(args ...any) {
	t.Log(args...)
	t.FailNow()
}

// Fatalf is equivalent to Logf followed by FailNow.
func (t *MockTestingT) Fatalf(format string, args ...any) {
	t.Logf(format, args...)
	t.FailNow()
}

// Log records a log entry.
func (t *MockTestingT) Log(args ...any) {
	t.Logs = append(t.Logs, fmt.Sprint(args...))
}

// Logf records a formatted log entry.
func (t *MockTestingT) Logf(format string, args ...any) {
	t.Logs = append(t.Logs, fmt.Sprintf(format, args...))
}

// Skip is equivalent to Log followed by SkipNow.
func (t *MockTestingT) Skip(args ...any) {
	t.Log(args...)
	t.SkipNow()
}

// SkipNow records the test to be skipped.
func (t *MockTestingT) SkipNow() {
	t.HasSkipped = true
}

// Skipf is equivalent to Logf followed by SkipNow.
func (t *MockTestingT) Skipf(format string, args ...any) {
	t.Logf(format, args...)
	t.SkipNow()
}

// Skipped reports whether the test was skipped.
func (t *MockTestingT) Skipped() bool {
	return t.HasSkipped
}

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

// RequirePGTimestampAlmostEqual requires that expected and actual PostgreSQL timestamps
// are within at most 5 seconds of each other.
func RequirePGTimestampAlmostEqual(t TestingT, expected pgtype.Timestamp, actual pgtype.Timestamp) {
	require.Equal(t, expected.Valid, actual.Valid)

	if expected.Valid {
		RequireTimeAlmostEqual(t, expected.Time, actual.Time)
	}
}

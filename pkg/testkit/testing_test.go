package testkit_test

import (
	"testing"

	"github.com/alvii147/flagger-api/pkg/testkit"
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

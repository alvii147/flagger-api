package testkitinternal_test

import (
	"context"
	"testing"

	"github.com/alvii147/flagger-api/internal/testkitinternal"
	"github.com/alvii147/flagger-api/pkg/testkit"
	"github.com/stretchr/testify/require"
)

func TestRequireCreateDatabasePool(t *testing.T) {
	t.Parallel()

	mockT := testkit.NewMockTestingT()
	testkitinternal.RequireCreateDatabasePool(mockT)

	require.False(t, mockT.HasFailed)
	require.Len(t, mockT.Cleanups, 1)

	cleanupFunc := mockT.Cleanups[0]
	cleanupFunc()
}

func TestRequireCreateDatabaseConn(t *testing.T) {
	t.Parallel()

	mockT := testkit.NewMockTestingT()
	dbPool := testkitinternal.RequireCreateDatabasePool(t)
	ctx := context.Background()

	stat := dbPool.Stat()
	require.EqualValues(t, 0, stat.AcquireCount())
	require.EqualValues(t, 0, stat.AcquiredConns())

	testkitinternal.RequireCreateDatabaseConn(mockT, dbPool, ctx)

	require.False(t, mockT.HasFailed)
	require.Len(t, mockT.Cleanups, 1)

	stat = dbPool.Stat()
	require.EqualValues(t, 1, stat.AcquireCount())
	require.EqualValues(t, 1, stat.AcquiredConns())

	cleanupFunc := mockT.Cleanups[0]
	cleanupFunc()

	stat = dbPool.Stat()
	require.EqualValues(t, 1, stat.AcquireCount())
	require.EqualValues(t, 0, stat.AcquiredConns())
}

package testkitinternal

import (
	"context"
	"fmt"

	"github.com/alvii147/flagger-api/internal/flags"
	"github.com/alvii147/flagger-api/pkg/testkit"
)

// MustCreateUserFlag creates and returns a new Flag for User and panics on error.
func MustCreateUserFlag(t testkit.TestingT, userUUID string, name string) *flags.Flag {
	dbPool := RequireCreateDatabasePool(t)
	dbConn := RequireCreateDatabaseConn(t, dbPool, context.Background())
	repo := flags.NewRepository()

	flag := &flags.Flag{
		UserUUID: userUUID,
		Name:     name,
	}

	flag, err := repo.CreateFlag(dbConn, flag)
	if err != nil {
		panic(fmt.Sprintf("MustCreateUserFlag failed to repo.CreateFlag: %v", err))
	}

	return flag
}

package flags

import (
	"context"
	"errors"
	"fmt"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/pkg/errutils"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service performs all Flag related business logic.
type Service interface {
	CreateFlag(ctx context.Context, name string) (*Flag, error)
	GetFlagByID(ctx context.Context, flagID int) (*Flag, error)
	GetFlagByName(ctx context.Context, name string) (*Flag, error)
	ListFlags(ctx context.Context) ([]*Flag, error)
	UpdateFlag(ctx context.Context, flagID int, name string, isEnabled bool) (*Flag, error)
}

// service implements Service.
type service struct {
	dbPool     *pgxpool.Pool
	repository Repository
}

// NewService returns a new Service.
func NewService(dbPool *pgxpool.Pool, repo Repository) Service {
	return &service{
		dbPool:     dbPool,
		repository: repo,
	}
}

// CreateFlag creates new Flag for User.
func (svc *service) CreateFlag(ctx context.Context, name string) (*Flag, error) {
	userUUID, ok := ctx.Value(auth.UserUUIDContextKey).(string)
	if !ok {
		return nil, errors.New("CreateFlag failed to ctx.Value user UUID from ctx")
	}

	flag := &Flag{
		UserUUID: userUUID,
		Name:     name,
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("CreateFlag failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	flag, err = svc.repository.CreateFlag(dbConn, flag)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseUniqueViolation):
			err = fmt.Errorf("CreateFlag failed to svc.repository.CreateFlag, %w: %w", errutils.ErrFlagAlreadyExists, err)
		default:
			err = fmt.Errorf("CreateFlag failed to svc.repository.CreateFlag: %w", err)
		}
		return nil, err
	}

	return flag, nil
}

// GetFlagByID retrieves Flag by ID for currently authenticated User.
func (svc *service) GetFlagByID(ctx context.Context, flagID int) (*Flag, error) {
	userUUID, ok := ctx.Value(auth.UserUUIDContextKey).(string)
	if !ok {
		return nil, errors.New("GetFlagByID failed to ctx.Value user UUID from ctx")
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetFlagByID failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	flag, err := svc.repository.GetFlagByID(dbConn, flagID, userUUID)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsReturned):
			err = fmt.Errorf("GetFlagByID failed to svc.repository.GetFlagByID, %w: %w", errutils.ErrFlagNotFound, err)
		default:
			err = fmt.Errorf("GetFlagByID failed to svc.repository.GetFlagByID: %w", err)
		}
		return nil, err
	}

	return flag, nil
}

// GetFlagByName retrieves Flag by name for currently authenticated User.
func (svc *service) GetFlagByName(ctx context.Context, name string) (*Flag, error) {
	userUUID, ok := ctx.Value(auth.UserUUIDContextKey).(string)
	if !ok {
		return nil, errors.New("GetFlagByName failed to ctx.Value user UUID from ctx")
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetFlagByName failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	flag, err := svc.repository.GetFlagByName(dbConn, name, userUUID)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsReturned):
			err = fmt.Errorf("GetFlagByName failed to svc.repository.GetFlagByID, %w: %w", errutils.ErrFlagNotFound, err)
		default:
			err = fmt.Errorf("GetFlagByName failed to svc.repository.GetFlagByName: %w", err)
		}
		return nil, err
	}

	return flag, nil
}

// ListFlags retrieves Flags for currently authenticated User.
func (svc *service) ListFlags(ctx context.Context) ([]*Flag, error) {
	userUUID, ok := ctx.Value(auth.UserUUIDContextKey).(string)
	if !ok {
		return nil, errors.New("ListFlags failed to ctx.Value user UUID from ctx")
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListFlags failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	flags, err := svc.repository.ListFlagsByUserUUID(dbConn, userUUID)
	if err != nil {
		return nil, fmt.Errorf("ListFlags failed to svc.repository.GetFlagsByUserUUID: %w", err)
	}

	return flags, nil
}

// UpdateFlag updates Flag by ID for currently authenticated User.
func (svc *service) UpdateFlag(ctx context.Context, flagID int, name string, isEnabled bool) (*Flag, error) {
	userUUID, ok := ctx.Value(auth.UserUUIDContextKey).(string)
	if !ok {
		return nil, errors.New("UpdateFlag failed to ctx.Value user UUID from ctx")
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("UpdateFlag failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	flag := &Flag{
		ID:        flagID,
		UserUUID:  userUUID,
		Name:      name,
		IsEnabled: isEnabled,
	}

	flag, err = svc.repository.UpdateFlagByID(dbConn, flag)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsAffected):
			err = fmt.Errorf("UpdateFlag failed to svc.repository.UpdateFlagByID, %w: %w", errutils.ErrFlagNotFound, err)
		default:
			err = fmt.Errorf("UpdateFlag failed to svc.repository.UpdateFlagByID: %w", err)
		}
		return nil, err
	}

	return flag, nil
}

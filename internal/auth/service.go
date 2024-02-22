package auth

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/alvii147/flagger-api/internal/templatesmanager"
	"github.com/alvii147/flagger-api/pkg/errutils"
	"github.com/alvii147/flagger-api/pkg/logging"
	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Service performs all auth-related business logic.
type Service interface {
	CreateUser(ctx context.Context, wg *sync.WaitGroup, email string, password string, firstName string, lastName string) (*User, error)
	ActivateUser(ctx context.Context, token string) error
	GetCurrentUser(ctx context.Context) (*User, error)
	CreateJWT(ctx context.Context, email string, password string) (string, string, error)
	RefreshJWT(ctx context.Context, token string) (string, error)
	CreateAPIKey(ctx context.Context, name string, expiresAt pgtype.Timestamp) (*APIKey, string, error)
	ListAPIKeys(ctx context.Context) ([]*APIKey, error)
	FindAPIKey(ctx context.Context, rawKey string) (*APIKey, error)
	DeleteAPIKey(ctx context.Context, apiKeyID int) error
}

// service implements Service.
type service struct {
	dbPool      *pgxpool.Pool
	mailClient  mailclient.MailClient
	tmplManager templatesmanager.Manager
	repository  Repository
}

// NewService returns a new Service.
func NewService(
	dbPool *pgxpool.Pool,
	mailClient mailclient.MailClient,
	tmplManager templatesmanager.Manager,
	repo Repository,
) Service {
	return &service{
		dbPool:      dbPool,
		mailClient:  mailClient,
		tmplManager: tmplManager,
		repository:  repo,
	}
}

// CreateUser creates a new User.
func (svc *service) CreateUser(
	ctx context.Context,
	wg *sync.WaitGroup,
	email string,
	password string,
	firstName string,
	lastName string,
) (*User, error) {
	logger := logging.GetLogger()

	hashedPasswordBytes, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("CreateUser failed to hashPassword: %w", err)
	}
	hashedPassword := string(hashedPasswordBytes)

	user := &User{
		UUID:        uuid.NewString(),
		Email:       email,
		Password:    hashedPassword,
		FirstName:   firstName,
		LastName:    lastName,
		IsActive:    false,
		IsSuperUser: false,
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("CreateUser failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	user, err = svc.repository.CreateUser(dbConn, user)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseUniqueViolation):
			err = fmt.Errorf("CreateUser failed to svc.repository.CreateUser, %w: %w", errutils.ErrUserAlreadyExists, err)
		default:
			err = fmt.Errorf("CreateUser failed to svc.repository.CreateUser: %w", err)
		}
		return nil, err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := sendActivationMail(user, svc.mailClient, svc.tmplManager)
		if err != nil {
			logger.LogError("CreateUser failed to sendActivationMail:", err)
		}
	}()

	return user, nil
}

// ActivateUser activates User from activation JWT.
func (svc *service) ActivateUser(ctx context.Context, token string) error {
	claims, ok := validateActivationJWT(token)
	if !ok {
		return fmt.Errorf("ActivateUser failed to validateActivationJWT %s: %w", token, errutils.ErrInvalidToken)
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("ActivateUser failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	err = svc.repository.ActivateUserByUUID(dbConn, claims.Subject)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsAffected):
			err = fmt.Errorf("ActivateUser failed to svc.repository.ActivateUserByUUID, %w: %w", errutils.ErrUserNotFound, err)
		default:
			err = fmt.Errorf("ActivateUser failed to svc.repository.ActivateUserByUUID: %w", err)
		}
		return err
	}

	return nil
}

// GetCurrentUser gets currently authenticated User.
func (svc *service) GetCurrentUser(ctx context.Context) (*User, error) {
	userUUID, ok := ctx.Value(AuthContextKeyUserUUID).(string)
	if !ok {
		return nil, errors.New("GetCurrentUser failed to ctx.Value user UUID from ctx")
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetCurrentUser failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	user, err := svc.repository.GetUserByUUID(dbConn, userUUID)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsReturned):
			err = fmt.Errorf("GetCurrentUser failed to svc.repository.GetUserByUUID, %w: %w", errutils.ErrUserNotFound, err)
		default:
			err = fmt.Errorf("GetCurrentUser failed to svc.repository.GetUserByUUID: %w", err)
		}
		return nil, err
	}

	return user, nil
}

// CreateJWT authenticates User and creates new access and refresh JWTs.
func (svc *service) CreateJWT(
	ctx context.Context,
	email string,
	password string,
) (string, string, error) {
	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return "", "", fmt.Errorf("CreateJWT failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	dummyUser := &User{}
	dummyUser.UUID = "93c58b3c-f087-4e97-805a-1e4676cdd5ec"
	dummyUser.Password = "$2a$14$078K5mTZHgbMDQ.K/U656OEb7v5HIyB9cBPLoXiQREAoXmCmgsywW"
	failAuth := false

	user, err := svc.repository.GetUserByEmail(dbConn, email)
	if err != nil && !errors.Is(err, errutils.ErrDatabaseNoRowsReturned) {
		return "", "", fmt.Errorf("CreateJWT failed to svc.repository.GetActiveUserByEmail: %w", err)
	}

	if err != nil || user == nil || !user.IsActive {
		failAuth = true
		user = dummyUser
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		failAuth = true
		user = dummyUser
	}

	accessToken, err := createAuthJWT(user.UUID, JWTTypeAccess)
	if err != nil {
		return "", "", fmt.Errorf("CreateJWT failed to createAuthJWT of type %s: %w", JWTTypeAccess, err)
	}

	refreshToken, err := createAuthJWT(user.UUID, JWTTypeRefresh)
	if err != nil {
		return "", "", fmt.Errorf("CreateJWT failed to createAuthJWT of type %s: %w", JWTTypeRefresh, err)
	}

	if failAuth {
		return "", "", fmt.Errorf("CreateJWT failed: %w", errutils.ErrInvalidCredentials)
	}

	return accessToken, refreshToken, nil
}

// RefreshJWT validates refresh token and creates new access token.
func (svc *service) RefreshJWT(ctx context.Context, token string) (string, error) {
	claims, ok := validateAuthJWT(token, JWTTypeRefresh)
	if !ok {
		return "", fmt.Errorf("RefreshJWT failed to validateAuthJWT %s: %w", token, errutils.ErrInvalidToken)
	}

	accessToken, err := createAuthJWT(claims.Subject, JWTTypeAccess)
	if err != nil {
		return "", fmt.Errorf("RefreshJWT failed to createAuthJWT of type %s: %w", JWTTypeAccess, err)
	}

	return accessToken, nil
}

// CreateAPIKey creates new API key for User.
func (svc *service) CreateAPIKey(ctx context.Context, name string, expiresAt pgtype.Timestamp) (*APIKey, string, error) {
	userUUID, ok := ctx.Value(AuthContextKeyUserUUID).(string)
	if !ok {
		return nil, "", errors.New("CreateAPIKey failed to ctx.Value user UUID from ctx")
	}

	prefix, rawKey, hashedKey, err := createAPIKey()
	if err != nil {
		return nil, "", fmt.Errorf("CreateAPIKey failed to generateAPIKey: %w", err)
	}

	apiKey := &APIKey{
		UserUUID:  userUUID,
		Prefix:    prefix,
		HashedKey: hashedKey,
		Name:      name,
		ExpiresAt: expiresAt,
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("CreateAPIKey failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	apiKey, err = svc.repository.CreateAPIKey(dbConn, apiKey)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseUniqueViolation):
			err = fmt.Errorf("CreateAPIKey failed to svc.repository.CreateAPIKey, %w: %w", errutils.ErrAPIKeyAlreadyExists, err)
		default:
			err = fmt.Errorf("CreateAPIKey failed to svc.repository.CreateAPIKey: %w", err)
		}
		return nil, "", err
	}

	return apiKey, rawKey, nil
}

// ListAPIKeys retrieves API keys for currently authenticated User.
func (svc *service) ListAPIKeys(ctx context.Context) ([]*APIKey, error) {
	userUUID, ok := ctx.Value(AuthContextKeyUserUUID).(string)
	if !ok {
		return nil, errors.New("ListAPIKeys failed to ctx.Value user UUID from ctx")
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListAPIKeys failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	apiKeys, err := svc.repository.ListAPIKeysByUserUUID(dbConn, userUUID)
	if err != nil {
		return nil, fmt.Errorf("ListAPIKeys failed to svc.repository.ListAPIKeysByUserUUID: %w", err)
	}

	return apiKeys, nil
}

// FindAPIKey parses and finds an API Key.
func (svc *service) FindAPIKey(ctx context.Context, rawKey string) (*APIKey, error) {
	prefix, _, ok := parseAPIKey(rawKey)
	if !ok {
		return nil, errors.New("FindAPIKey failed to parseAPIKey")
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("FindAPIKey failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	apiKeys, err := svc.repository.ListActiveAPIKeysByPrefix(dbConn, prefix)
	if err != nil {
		return nil, fmt.Errorf("FindAPIKey failed to svc.repository.ListActiveAPIKeysByPrefix: %w", err)
	}

	keyBytes := []byte(rawKey)
	for _, apiKey := range apiKeys {
		err = bcrypt.CompareHashAndPassword([]byte(apiKey.HashedKey), keyBytes)
		if err == nil {
			return apiKey, nil
		}
	}

	return nil, fmt.Errorf("FindAPIKey failed to find API key: %w", errutils.ErrAPIKeyNotFound)
}

// DeleteAPIKey deletes API key for currently authenticated User.
func (svc *service) DeleteAPIKey(ctx context.Context, apiKeyID int) error {
	userUUID, ok := ctx.Value(AuthContextKeyUserUUID).(string)
	if !ok {
		return errors.New("DeleteAPIKey failed to ctx.Value user UUID from ctx")
	}

	dbConn, err := svc.dbPool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("DeleteAPIKey failed to svc.dbPool.Acquire: %w", err)
	}
	defer dbConn.Release()

	err = svc.repository.DeleteAPIKey(dbConn, apiKeyID, userUUID)
	if err != nil {
		switch {
		case errors.Is(err, errutils.ErrDatabaseNoRowsAffected):
			err = fmt.Errorf("DeleteAPIKey failed to svc.repository.DeleteAPIKey, %w: %w", errutils.ErrAPIKeyNotFound, err)
		default:
			err = fmt.Errorf("DeleteAPIKey failed to svc.repository.DeleteAPIKey: %w", err)
		}
		return err
	}

	return nil
}

package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/database"
	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/internal/flags"
	"github.com/alvii147/flagger-api/pkg/logging"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Controller handles server API operations.
type Controller interface {
	Route() *mux.Router
	Serve(router *mux.Router) error
	Close() error
}

// controller implements Controller.
type controller struct {
	dbPool       *pgxpool.Pool
	authService  auth.Service
	flagsService flags.Service
}

// NewController sets up the server and returns a new Controller.
func NewController() (Controller, error) {
	dbPool, err := database.CreatePool()
	if err != nil {
		return nil, fmt.Errorf("NewController failed to database.GetPool: %w", err)
	}

	authRepository := auth.NewRepository()
	authService := auth.NewService(dbPool, authRepository)

	flagsRepository := flags.NewRepository()
	flagsService := flags.NewService(dbPool, flagsRepository)

	ctrl := &controller{
		dbPool:       dbPool,
		authService:  authService,
		flagsService: flagsService,
	}

	return ctrl, nil
}

// Serve runs the Controller server.
func (ctrl *controller) Serve(router *mux.Router) error {
	config := env.GetConfig()
	logger := logging.GetLogger()

	addr := fmt.Sprintf("%s:%d", config.Hostname, config.Port)
	logger.LogInfo("Server running on", addr)
	err := http.ListenAndServe(addr, router)
	if err != nil {
		return fmt.Errorf("Serve failed to http.ListenAndServe %s: %w", addr, err)
	}

	return nil
}

// Close closes the Controller and its connections.
func (ctrl *controller) Close() error {
	var err error
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctrl.dbPool.Close()
	}()

	wg.Wait()

	return err
}

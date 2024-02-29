package server

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/alvii147/flagger-api/internal/auth"
	"github.com/alvii147/flagger-api/internal/database"
	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/internal/flags"
	"github.com/alvii147/flagger-api/internal/templatesmanager"
	"github.com/alvii147/flagger-api/pkg/httputils"
	"github.com/alvii147/flagger-api/pkg/logging"
	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/jackc/pgx/v5/pgxpool"
)

// controller handles server API operations.
type controller struct {
	config       *env.Config
	router       httputils.Router
	dbPool       *pgxpool.Pool
	logger       logging.Logger
	mailClient   mailclient.Client
	tmplManager  templatesmanager.Manager
	authService  auth.Service
	flagsService flags.Service
}

// NewController sets up the server and returns a new controller.
func NewController() (*controller, error) {
	config, err := env.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("NewController failed to env.NewConfig: %w", err)
	}

	router := httputils.NewRouter()
	dbPool, err := database.CreatePool(
		config.PostgresHostname,
		config.PostgresPort,
		config.PostgresUsername,
		config.PostgresPassword,
		config.PostgresDatabaseName,
	)
	if err != nil {
		return nil, fmt.Errorf("NewController failed to database.GetPool: %w", err)
	}

	logger := logging.NewLogger(os.Stdout, os.Stderr)

	var mailClient mailclient.Client
	switch config.MailClientType {
	case mailclient.ClientTypeSMTP:
		mailClient = mailclient.NewSMTPClient(
			config.SMTPHostname,
			config.SMTPPort,
			config.SMTPUsername,
			config.SMTPPassword,
		)
	case mailclient.ClientTypeInMemory:
		mailClient = mailclient.NewInMemClient("support@flagger.com")
	case mailclient.ClientTypeConsole:
		mailClient = mailclient.NewConsoleClient("support@flagger.com", os.Stdout)
	default:
		return nil, fmt.Errorf("NewController failed, unknown mail client type %s", config.MailClientType)
	}

	tmplManager := templatesmanager.NewManager()

	authRepository := auth.NewRepository()
	authService := auth.NewService(
		config,
		dbPool,
		logger,
		mailClient,
		tmplManager,
		authRepository,
	)

	flagsRepository := flags.NewRepository()
	flagsService := flags.NewService(dbPool, flagsRepository)

	ctrl := &controller{
		config:       config,
		router:       router,
		dbPool:       dbPool,
		logger:       logger,
		mailClient:   mailClient,
		tmplManager:  tmplManager,
		authService:  authService,
		flagsService: flagsService,
	}

	return ctrl, nil
}

// Serve runs the Controller server.
func (ctrl *controller) Serve() error {
	addr := fmt.Sprintf("%s:%d", ctrl.config.Hostname, ctrl.config.Port)
	ctrl.logger.LogInfo("Server running on", addr)
	err := http.ListenAndServe(addr, ctrl.router)
	if err != nil {
		return fmt.Errorf("Serve failed to http.ListenAndServe %s: %w", addr, err)
	}

	return nil
}

// ServeHTTP takes in a given request and writes a response.
func (ctrl *controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctrl.router.ServeHTTP(w, r)
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

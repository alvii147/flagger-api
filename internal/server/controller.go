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
	"github.com/alvii147/flagger-api/pkg/logging"
	"github.com/alvii147/flagger-api/pkg/mailclient"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// controller handles server API operations.
type controller struct {
	dbPool       *pgxpool.Pool
	mailClient   mailclient.MailClient
	tmplManager  templatesmanager.Manager
	authService  auth.Service
	flagsService flags.Service
}

// NewController sets up the server and returns a new controller.
func NewController() (*controller, error) {
	config := env.GetConfig()

	dbPool, err := database.CreatePool()
	if err != nil {
		return nil, fmt.Errorf("NewController failed to database.GetPool: %w", err)
	}

	var mailClient mailclient.MailClient
	switch config.MailClientType {
	case mailclient.MailClientTypeSMTP:
		mailClient = mailclient.NewSMTPMailClient(
			config.SMTPHostname,
			config.SMTPPort,
			config.SMTPUsername,
			config.SMTPPassword,
		)
	case mailclient.MailClientTypeInMemory:
		mailClient = mailclient.NewInMemMailClient("support@flagger.com")
	case mailclient.MailClientTypeConsole:
		mailClient = mailclient.NewConsoleMailClient("support@flagger.com", os.Stdout)
	default:
		return nil, fmt.Errorf("NewController failed, unknown mail client type %s", config.MailClientType)
	}

	tmplManager := templatesmanager.NewManager()

	authRepository := auth.NewRepository()
	authService := auth.NewService(dbPool, mailClient, tmplManager, authRepository)

	flagsRepository := flags.NewRepository()
	flagsService := flags.NewService(dbPool, flagsRepository)

	ctrl := &controller{
		dbPool:       dbPool,
		mailClient:   mailClient,
		tmplManager:  tmplManager,
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

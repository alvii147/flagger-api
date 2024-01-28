package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/internal/server"
	"github.com/alvii147/flagger-api/pkg/logging"
	"github.com/alvii147/flagger-api/pkg/mailclient"
)

func main() {
	config := env.NewConfig()
	env.SetConfig(config)

	logger := logging.NewLogger(os.Stdout, os.Stderr)
	logging.SetLogger(logger)

	mailClient, err := mailclient.NewMailClient(
		config.MailClientType,
		config.SMTPHostname,
		config.SMTPPort,
		config.SMTPUsername,
		config.SMTPPassword,
		config.MailTemplatesDir,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to mail.NewMailClient: %v\n", err)
		return
	}
	mailclient.SetMailClient(mailClient)

	ctrl, err := server.NewController()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to controller.NewController: %v\n", err)
		return
	}
	defer ctrl.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		err := ctrl.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to ctrl.Close: %v\n", err)
			os.Exit(1)
		}

		os.Exit(0)
	}()

	router := ctrl.Route()

	err = ctrl.Serve(router)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to ctrl.Serve: %v\n", err)
		return
	}
}

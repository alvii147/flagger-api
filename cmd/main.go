package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alvii147/flagger-api/internal/env"
	"github.com/alvii147/flagger-api/internal/server"
)

func main() {
	config := env.NewConfig()
	env.SetConfig(config)

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

	mux := ctrl.Route()

	err = ctrl.Serve(mux)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to ctrl.Serve: %v\n", err)
		return
	}
}

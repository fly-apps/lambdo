package main

import (
	"context"
	"github.com/superfly/lambdo/cmd"
	"github.com/superfly/lambdo/internal/config"
	"github.com/superfly/lambdo/internal/logging"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	err := config.Configure()

	if err != nil {
		log.Printf("configuration error: %v", err)
		os.Exit(1)
	}

	err = logging.SetupLogging(config.GetConfig().Environment == "production")

	if err != nil {
		log.Printf("logging error: %v", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		logging.GetLogger().Debug("signal received", zap.String("signal", sig.String()))
		cancel()
		return
	}()

	logging.GetLogger().Info("starting lambdo")

	err = cmd.Execute(ctx)

	if err != nil {
		panic(err)
	}

	logging.GetLogger().Info("👋 exiting lambdo")
}

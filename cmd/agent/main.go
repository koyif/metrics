package main

import (
	"context"
	"github.com/koyif/metrics/pkg/logger"
	"log"
	"os/signal"
	"syscall"

	"github.com/koyif/metrics/internal/agent/app"
	"github.com/koyif/metrics/internal/agent/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	if err = logger.Initialize(); err != nil {
		log.Fatalf("error starting logger: %v", err)
	}

	a := app.New(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if err := a.Run(ctx); err != nil {
		logger.Log.Fatal("error running agent", logger.Error(err))
	}

	<-ctx.Done()
	logger.Log.Info("shutting down")
}

package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/koyif/metrics/internal/agent/app"
	"github.com/koyif/metrics/internal/agent/config"
	"github.com/koyif/metrics/pkg/logger"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func printBuildInfo() {
	version := "N/A"
	if buildVersion != "" {
		version = buildVersion
	}
	date := "N/A"
	if buildDate != "" {
		date = buildDate
	}
	commit := "N/A"
	if buildCommit != "" {
		commit = buildCommit
	}

	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)
}

func main() {
	printBuildInfo()

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

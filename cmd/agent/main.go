package main

import (
	"context"
	"github.com/koyif/metrics/internal/agent"
	"github.com/koyif/metrics/internal/agent/client"
	"github.com/koyif/metrics/internal/agent/config"
	"github.com/koyif/metrics/internal/agent/scraper"
	"github.com/koyif/metrics/internal/agent/storage"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.Load()

	if err := run(cfg); err != nil {
		panic(err)
	}
}

func run(cfg *config.Config) error {
	stg := storage.New()
	sc := scraper.New(stg)
	cl, err := client.New(cfg, &http.Client{})
	if err != nil {
		return err
	}

	a := agent.New(cfg, sc, cl)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.Start(ctx)

	<-stop
	slog.Info("shutting down")

	return nil
}

package app

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

type App struct {
	cfg *config.Config
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (app *App) Run() error {
	stg := storage.New()
	sc := scraper.New(stg)
	cl, err := client.New(app.cfg, &http.Client{})
	if err != nil {
		return err
	}

	a := agent.New(app.cfg, sc, cl)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.Start(ctx)

	<-stop
	slog.Info("shutting down")

	return nil
}

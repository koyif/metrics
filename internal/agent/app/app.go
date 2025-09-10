package app

import (
	"context"
	"github.com/koyif/metrics/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/koyif/metrics/internal/agent"
	"github.com/koyif/metrics/internal/agent/client"
	"github.com/koyif/metrics/internal/agent/config"
	"github.com/koyif/metrics/internal/agent/scraper"
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
	ctx, cancel := context.WithCancel(context.Background())
	sc := scraper.New(app.cfg)
	metricsCh := sc.Start(ctx)
	defer cancel()

	cl, err := client.New(app.cfg, &http.Client{})
	if err != nil {
		return err
	}

	a := agent.New(app.cfg, cl)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	a.Start(ctx, metricsCh)

	<-stop
	logger.Log.Info("shutting down")

	return nil
}

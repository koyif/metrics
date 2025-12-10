package app

import (
	"context"
	"net/http"
	"sync"

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

func (app *App) Run(ctx context.Context, wg *sync.WaitGroup) error {
	sc := scraper.New(app.cfg)
	metricsCh := sc.Start(ctx)

	cl, err := client.New(app.cfg, &http.Client{})
	if err != nil {
		return err
	}

	a := agent.New(app.cfg, cl)
	a.Start(ctx, wg, metricsCh)

	return nil
}

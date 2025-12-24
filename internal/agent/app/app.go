package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/koyif/metrics/internal/agent"
	"github.com/koyif/metrics/internal/agent/client"
	"github.com/koyif/metrics/internal/agent/config"
	"github.com/koyif/metrics/internal/agent/grpcclient"
	"github.com/koyif/metrics/internal/agent/scraper"
	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/pkg/logger"
)

type metricsClient interface {
	SendMetric(models.Metrics) error
	SendMetrics([]models.Metrics) error
}

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

	metricsClient, err := newMetricsClient(app.cfg)

	if err != nil {
		return fmt.Errorf("failed to create metrics client: %w", err)
	}

	a := agent.New(app.cfg, metricsClient)
	a.Start(ctx, wg, metricsCh)

	return nil
}

func newMetricsClient(cfg *config.Config) (metricsClient, error) {
	if cfg.UseGRPC {
		logger.Log.Info("using gRPC client")
		return grpcclient.New(cfg)
	} else {
		logger.Log.Info("using HTTP client")
		return client.New(cfg, &http.Client{Timeout: 10 * time.Second})
	}
}

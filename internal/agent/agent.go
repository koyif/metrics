package agent

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/koyif/metrics/internal/agent/config"
	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/pkg/logger"
)

type metricsClient interface {
	SendMetric(metric models.Metrics) error
	SendMetrics(metrics []models.Metrics) error
}

type Agent struct {
	cfg           *config.Config
	metricsClient metricsClient
}

func New(cfg *config.Config, cl metricsClient) *Agent {
	return &Agent{
		cfg:           cfg,
		metricsClient: cl,
	}
}

func (a *Agent) Start(ctx context.Context, ch <-chan []models.Metrics) {
	go func() {
		reportTicker := time.NewTicker(a.cfg.ReportInterval.Value())

		for {
			select {
			case <-ctx.Done():
				reportTicker.Stop()
				return
			case <-reportTicker.C:
				a.reportMetrics(<-ch)
			}
		}
	}()
}

func (a *Agent) reportMetrics(metrics []models.Metrics) {
	r := rand.Float64()

	metrics = append(metrics, models.Metrics{
		ID:    "RandomValue",
		MType: models.Gauge,
		Value: &r,
	},
	)

	err := a.metricsClient.SendMetrics(metrics)
	if err != nil {
		logger.Log.Error("error sending metrics", logger.Error(err))
		return
	}

	logger.Log.Info("sent metrics", logger.Int("count", len(metrics)))
}

package agent

import (
	"context"
	"github.com/koyif/metrics/pkg/logger"
	"math/rand/v2"
	"time"

	"github.com/koyif/metrics/internal/agent/config"
	"github.com/koyif/metrics/internal/models"
)

type scraper interface {
	Scrap()
	Count() int64
	Metrics() []models.Metrics
	Reset()
}

type metricsClient interface {
	SendMetric(metric models.Metrics) error
	SendMetrics(metrics []models.Metrics) error
}

type Agent struct {
	cfg           *config.Config
	scraper       scraper
	metricsClient metricsClient
}

func New(cfg *config.Config, scraper scraper, cl metricsClient) *Agent {
	return &Agent{
		cfg:           cfg,
		scraper:       scraper,
		metricsClient: cl,
	}
}

func (a *Agent) Start(ctx context.Context) {
	go func() {
		pollTicker := time.NewTicker(a.cfg.PollInterval)
		reportTicker := time.NewTicker(a.cfg.ReportInterval)

		for {
			select {
			case <-ctx.Done():
				pollTicker.Stop()
				reportTicker.Stop()

				return
			case <-pollTicker.C:
				a.pollMetrics()
			case <-reportTicker.C:
				a.reportMetrics()
			}
		}
	}()
}

func (a *Agent) pollMetrics() {
	a.scraper.Scrap()
}

func (a *Agent) reportMetrics() {
	metrics := a.scraper.Metrics()
	r := rand.Float64()
	c := a.scraper.Count()

	metrics = append(metrics, []models.Metrics{
		{
			ID:    "RandomValue",
			MType: models.Gauge,
			Value: &r,
		},
		{
			ID:    "PollCount",
			MType: models.Counter,
			Delta: &c,
		}}...,
	)

	err := a.metricsClient.SendMetrics(metrics)
	if err != nil {
		logger.Log.Error("error sending metrics", logger.Error(err))
		return
	}

	logger.Log.Info("sent metrics", logger.Int("count", len(metrics)))

	a.scraper.Reset()

}

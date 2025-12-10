package agent

import (
	"context"
	"math/rand/v2"
	"sync"
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

func (a *Agent) Start(ctx context.Context, wg *sync.WaitGroup, ch <-chan []models.Metrics) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		reportTicker := time.NewTicker(a.cfg.ReportInterval.Value())
		defer reportTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Log.Info("agent received shutdown signal, flushing remaining metrics")
				// Try to send any remaining metrics in the channel
				for {
					select {
					case metrics, ok := <-ch:
						if !ok {
							logger.Log.Info("metrics channel closed")
							return
						}
						a.reportMetrics(metrics)
					case <-time.After(100 * time.Millisecond):
						logger.Log.Info("no more metrics to send")
						return
					}
				}
			case <-reportTicker.C:
				select {
				case metrics, ok := <-ch:
					if !ok {
						logger.Log.Info("metrics channel closed, stopping agent")
						return
					}
					a.reportMetrics(metrics)
				default:
					logger.Log.Debug("no metrics available to send")
				}
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

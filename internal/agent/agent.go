package agent

import (
	"context"
	"fmt"
	"github.com/koyif/metrics/internal/agent/config"
	"log/slog"
	"math/rand/v2"
	"strconv"
	"time"
)

type scraper interface {
	Scrap()
	Count() int64
	Metrics() map[string]float64
	Reset()
}

type metricsClient interface {
	Send(metricType, metricName, value string) error
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
		pollTicker := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
		reportTicker := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)

	loop:
		for {
			select {
			case <-ctx.Done():
				pollTicker.Stop()
				reportTicker.Stop()

				break loop
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
	const op = "Agent.reportMetrics"
	gaugeMetrics := a.scraper.Metrics()
	gaugeMetrics["RandomValue"] = rand.Float64()

	counterMetrics := map[string]int64{
		"PollCount": a.scraper.Count(),
	}

	for k, v := range gaugeMetrics {
		slog.Info(fmt.Sprintf("%s: %s: %f", op, k, v))
		value := strconv.FormatFloat(v, 'f', -1, 64)

		if err := a.metricsClient.Send("gauge", k, value); err != nil {
			slog.Error(fmt.Sprintf("%s: %s: %v", op, k, err))
		}
	}

	for k, v := range counterMetrics {
		slog.Info(fmt.Sprintf("%s: %s: %d", op, k, v))
		value := strconv.FormatInt(v, 10)

		if err := a.metricsClient.Send("counter", k, value); err != nil {
			slog.Error(fmt.Sprintf("%s: %s: %v", op, k, err))
		}
	}

	a.scraper.Reset()
}

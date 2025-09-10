package scraper

import (
	"context"
	"github.com/koyif/metrics/internal/agent/config"
	"github.com/koyif/metrics/internal/models"
	"runtime"
	"time"
)

type Scraper struct {
	count int64
	cfg   *config.Config
}

func New(cfg *config.Config) *Scraper {
	return &Scraper{
		count: 0,
		cfg:   cfg,
	}
}

func (s *Scraper) Scrap() []models.Metrics {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)

	s.count++
	metrics := make([]models.Metrics, 0, 30)

	metrics = append(metrics, newGauge("Alloc", float64(memStats.Alloc)))
	metrics = append(metrics, newGauge("BuckHashSys", float64(memStats.BuckHashSys)))
	metrics = append(metrics, newGauge("Frees", float64(memStats.Frees)))
	metrics = append(metrics, newGauge("GCCPUFraction", memStats.GCCPUFraction))
	metrics = append(metrics, newGauge("GCSys", float64(memStats.GCSys)))
	metrics = append(metrics, newGauge("HeapAlloc", float64(memStats.HeapAlloc)))
	metrics = append(metrics, newGauge("HeapIdle", float64(memStats.HeapIdle)))
	metrics = append(metrics, newGauge("HeapInuse", float64(memStats.HeapInuse)))
	metrics = append(metrics, newGauge("HeapObjects", float64(memStats.HeapObjects)))
	metrics = append(metrics, newGauge("HeapReleased", float64(memStats.HeapReleased)))
	metrics = append(metrics, newGauge("HeapSys", float64(memStats.HeapSys)))
	metrics = append(metrics, newGauge("LastGC", float64(memStats.LastGC)))
	metrics = append(metrics, newGauge("Lookups", float64(memStats.Lookups)))
	metrics = append(metrics, newGauge("MCacheInuse", float64(memStats.MCacheInuse)))
	metrics = append(metrics, newGauge("MCacheSys", float64(memStats.MCacheSys)))
	metrics = append(metrics, newGauge("MSpanInuse", float64(memStats.MSpanInuse)))
	metrics = append(metrics, newGauge("MSpanSys", float64(memStats.MSpanSys)))
	metrics = append(metrics, newGauge("Mallocs", float64(memStats.Mallocs)))
	metrics = append(metrics, newGauge("NextGC", float64(memStats.NextGC)))
	metrics = append(metrics, newGauge("NumForcedGC", float64(memStats.NumForcedGC)))
	metrics = append(metrics, newGauge("NumGC", float64(memStats.NumGC)))
	metrics = append(metrics, newGauge("OtherSys", float64(memStats.OtherSys)))
	metrics = append(metrics, newGauge("PauseTotalNs", float64(memStats.PauseTotalNs)))
	metrics = append(metrics, newGauge("StackInuse", float64(memStats.StackInuse)))
	metrics = append(metrics, newGauge("StackSys", float64(memStats.StackSys)))
	metrics = append(metrics, newGauge("Sys", float64(memStats.Sys)))
	metrics = append(metrics, newGauge("TotalAlloc", float64(memStats.TotalAlloc)))
	metrics = append(metrics, models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &s.count,
	})

	return metrics
}

func newGauge(ID string, value float64) models.Metrics {
	return models.Metrics{
		ID:    ID,
		MType: models.Gauge,
		Value: &value,
	}
}

func (s *Scraper) Start(ctx context.Context) <-chan []models.Metrics {
	metricsCh := make(chan []models.Metrics, 256)

	go func() {
		ticker := time.NewTicker(s.cfg.PollInterval.Value())

		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				close(metricsCh)
				return
			case <-ticker.C:
				metricsCh <- s.Scrap()
			}
		}
	}()

	return metricsCh
}

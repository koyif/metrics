package storage

import (
	"maps"
	"slices"

	"github.com/koyif/metrics/internal/models"
)

type MetricStorage struct {
	metrics map[string]models.Metrics
}

func New() *MetricStorage {
	return &MetricStorage{
		metrics: make(map[string]models.Metrics),
	}
}

func (s *MetricStorage) Metrics() []models.Metrics {
	return slices.Collect(maps.Values(s.metrics))
}

func (s *MetricStorage) Store(metric string, value float64) {
	s.metrics[metric] = models.Metrics{
		ID:    metric,
		Value: &value,
		MType: models.Gauge,
	}
}

func (s *MetricStorage) Clean() {
	s.metrics = make(map[string]models.Metrics)
}

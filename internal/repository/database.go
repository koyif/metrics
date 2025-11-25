package repository

import (
	"context"

	models "github.com/koyif/metrics/internal/models"
)

type database interface {
	StoreMetric(metric models.Metrics) error
	StoreAll(metrics []models.Metrics) error
	Metric(metricName string) (models.Metrics, error)
	AllMetrics() []models.Metrics
	Ping(ctx context.Context) error
}

type DatabaseRepository struct {
	db database
}

func NewDatabaseRepository(db database) *DatabaseRepository {
	return &DatabaseRepository{
		db: db,
	}
}

func (r DatabaseRepository) StoreGauge(metricName string, value float64) error {
	return r.db.StoreMetric(
		models.Metrics{
			ID:    metricName,
			MType: models.Gauge,
			Value: &value,
		},
	)
}

func (r DatabaseRepository) StoreCounter(metricName string, value int64) error {
	return r.db.StoreMetric(
		models.Metrics{
			ID:    metricName,
			MType: models.Counter,
			Delta: &value,
		},
	)
}

func (r DatabaseRepository) StoreAll(metrics []models.Metrics) error {
	return r.db.StoreAll(metrics)
}

func (r DatabaseRepository) Counter(metricName string) (int64, error) {
	metric, err := r.db.Metric(metricName)
	if err != nil {
		return 0, err
	}

	return *metric.Delta, nil
}

func (r DatabaseRepository) AllCounters() map[string]int64 {
	metrics := r.db.AllMetrics()
	counters := make(map[string]int64, len(metrics))
	for _, metric := range metrics {
		if metric.MType == models.Counter {
			counters[metric.ID] = *metric.Delta
		}
	}

	return counters
}

func (r DatabaseRepository) Gauge(metricName string) (float64, error) {
	metric, err := r.db.Metric(metricName)
	if err != nil {
		return 0, err
	}

	return *metric.Value, nil
}

func (r DatabaseRepository) AllGauges() map[string]float64 {
	metrics := r.db.AllMetrics()
	gauges := make(map[string]float64, len(metrics))
	for _, metric := range metrics {
		if metric.MType == models.Gauge {
			gauges[metric.ID] = *metric.Value
		}
	}

	return gauges
}

func (r DatabaseRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

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

// DatabaseRepository provides persistent storage for metrics using a database backend.
// It implements the same interface as MetricsRepository but stores data in PostgreSQL
// instead of memory, providing durability and consistency.
//
// This repository is used when DATABASE_DSN is configured.
type DatabaseRepository struct {
	db database
}

// NewDatabaseRepository creates a new database-backed metrics repository.
// The db parameter should be a PostgreSQL database connection with migrations applied.
func NewDatabaseRepository(db database) *DatabaseRepository {
	return &DatabaseRepository{
		db: db,
	}
}

// StoreGauge stores or updates a gauge metric in the database.
// Existing gauge values are replaced with the new value.
func (r DatabaseRepository) StoreGauge(metricName string, value float64) error {
	return r.db.StoreMetric(
		models.Metrics{
			ID:    metricName,
			MType: models.Gauge,
			Value: &value,
		},
	)
}

// StoreCounter stores or updates a counter metric in the database.
// The delta value is added to the existing counter value (upsert with increment).
func (r DatabaseRepository) StoreCounter(metricName string, value int64) error {
	return r.db.StoreMetric(
		models.Metrics{
			ID:    metricName,
			MType: models.Counter,
			Delta: &value,
		},
	)
}

// StoreAll stores or updates multiple metrics in a single database transaction.
// This is more efficient than individual updates for batch operations.
func (r DatabaseRepository) StoreAll(metrics []models.Metrics) error {
	return r.db.StoreAll(metrics)
}

// Counter retrieves the current value of a counter metric from the database.
// Returns an error if the metric doesn't exist or cannot be retrieved.
func (r DatabaseRepository) Counter(metricName string) (int64, error) {
	metric, err := r.db.Metric(metricName)
	if err != nil {
		return 0, err
	}

	return *metric.Delta, nil
}

// AllCounters retrieves all counter metrics from the database.
// Returns a map of metric names to their current values.
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

// Gauge retrieves the current value of a gauge metric from the database.
// Returns an error if the metric doesn't exist or cannot be retrieved.
func (r DatabaseRepository) Gauge(metricName string) (float64, error) {
	metric, err := r.db.Metric(metricName)
	if err != nil {
		return 0, err
	}

	return *metric.Value, nil
}

// AllGauges retrieves all gauge metrics from the database.
// Returns a map of metric names to their current values.
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

// Ping checks the database connection health.
// Returns an error if the database is unreachable or connection has failed.
func (r DatabaseRepository) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

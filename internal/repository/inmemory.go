package repository

import (
	"context"
	"fmt"
	"maps"
	"sync"

	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/internal/repository/dberror"
)

// MetricsRepository provides thread-safe in-memory storage for metrics.
// It maintains separate maps for counter and gauge metrics, protected by a read-write mutex.
//
// This repository is used when database storage is not configured,
// and can be persisted to file using the file service.
type MetricsRepository struct {
	mu       sync.RWMutex
	counters map[string]int64
	gauges   map[string]float64
}

// NewMetricsRepository creates a new in-memory metrics repository.
// The repository is safe for concurrent access.
func NewMetricsRepository() *MetricsRepository {
	return &MetricsRepository{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

// AllCounters returns a copy of all counter metrics.
// The returned map can be safely modified without affecting the repository.
func (m *MetricsRepository) AllCounters() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]int64, len(m.counters))
	maps.Copy(result, m.counters)

	return result
}

// AllGauges returns a copy of all gauge metrics.
// The returned map can be safely modified without affecting the repository.
func (m *MetricsRepository) AllGauges() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]float64, len(m.gauges))
	maps.Copy(result, m.gauges)

	return result
}

// Counter retrieves the value of a counter metric by name.
// Returns dberror.ErrValueNotFound if the metric doesn't exist.
func (m *MetricsRepository) Counter(metricName string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if v, ok := m.counters[metricName]; ok {
		return v, nil
	}
	return 0, dberror.ErrValueNotFound
}

// Gauge retrieves the value of a gauge metric by name.
// Returns dberror.ErrValueNotFound if the metric doesn't exist.
func (m *MetricsRepository) Gauge(metricName string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if v, ok := m.gauges[metricName]; ok {
		return v, nil
	}
	return 0, dberror.ErrValueNotFound
}

// StoreCounter adds the given value to the counter metric.
// If the counter doesn't exist, it is created with the given value.
// Counter values are cumulative and always increase.
func (m *MetricsRepository) StoreCounter(metricName string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters[metricName] += value
	return nil
}

// StoreGauge sets the gauge metric to the given value.
// If the gauge doesn't exist, it is created. Existing values are replaced.
func (m *MetricsRepository) StoreGauge(metricName string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauges[metricName] = value
	return nil
}

// StoreAll stores multiple metrics in a single batch operation.
// All updates are performed atomically under a single lock.
// Returns an error if any metric has invalid data (nil value/delta or unknown type).
func (m *MetricsRepository) StoreAll(metrics []models.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				return fmt.Errorf("gauge value is nil")
			}
			m.gauges[metric.ID] = *metric.Value
		case models.Counter:
			if metric.Delta == nil {
				return fmt.Errorf("counter delta is nil")
			}
			m.counters[metric.ID] += *metric.Delta
		default:
			return fmt.Errorf("unknown metric type")
		}
	}

	return nil
}

// Ping always returns nil for in-memory storage.
// This method exists to satisfy the repository interface.
func (m *MetricsRepository) Ping(_ context.Context) error {
	return nil
}

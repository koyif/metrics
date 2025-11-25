package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/internal/repository/dberror"
)

type MetricsRepository struct {
	mu       sync.RWMutex
	counters map[string]int64
	gauges   map[string]float64
}

func NewMetricsRepository() *MetricsRepository {
	return &MetricsRepository{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (m *MetricsRepository) AllCounters() map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]int64, len(m.counters))
	for k, v := range m.counters {
		result[k] = v
	}
	return result
}

func (m *MetricsRepository) AllGauges() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]float64, len(m.gauges))
	for k, v := range m.gauges {
		result[k] = v
	}
	return result
}

func (m *MetricsRepository) Counter(metricName string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if v, ok := m.counters[metricName]; ok {
		return v, nil
	}
	return 0, dberror.ErrValueNotFound
}

func (m *MetricsRepository) Gauge(metricName string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if v, ok := m.gauges[metricName]; ok {
		return v, nil
	}
	return 0, dberror.ErrValueNotFound
}

func (m *MetricsRepository) StoreCounter(metricName string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counters[metricName] += value
	return nil
}

func (m *MetricsRepository) StoreGauge(metricName string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauges[metricName] = value
	return nil
}

func (m *MetricsRepository) StoreAll(metrics []models.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				return errors.New("gauge value is nil")
			}
			m.gauges[metric.ID] = *metric.Value
		case models.Counter:
			if metric.Delta == nil {
				return errors.New("counter delta is nil")
			}
			m.counters[metric.ID] += *metric.Delta
		default:
			return errors.New("unknown metric type")
		}
	}

	return nil
}

func (m *MetricsRepository) Ping(_ context.Context) error {
	return nil
}

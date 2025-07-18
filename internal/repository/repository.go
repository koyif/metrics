package repository

import (
	"errors"
)

var ErrValueNotFound = errors.New("value not found")

type MetricsRepository struct {
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
	return m.counters
}

func (m *MetricsRepository) AllGauges() map[string]float64 {
	return m.gauges
}

func (m *MetricsRepository) Counter(metricName string) (int64, error) {
	if v, ok := m.counters[metricName]; ok {
		return v, nil
	} else {
		return 0, ErrValueNotFound
	}
}

func (m *MetricsRepository) Gauge(metricName string) (float64, error) {
	if v, ok := m.gauges[metricName]; ok {
		return v, nil
	} else {
		return 0, ErrValueNotFound
	}
}

func (m *MetricsRepository) StoreCounter(metricName string, value int64) error {
	m.counters[metricName] += value

	return nil
}

func (m *MetricsRepository) StoreGauge(metricName string, value float64) error {
	m.gauges[metricName] = value

	return nil
}

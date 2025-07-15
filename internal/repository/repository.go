package repository

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

func (m *MetricsRepository) StoreCounter(metricName string, value int64) error {
	m.counters[metricName] += value

	return nil
}

func (m *MetricsRepository) StoreGauge(metricName string, value float64) error {
	m.gauges[metricName] = value

	return nil
}

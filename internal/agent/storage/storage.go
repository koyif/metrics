package storage

type MetricStorage struct {
	metrics map[string]float64
}

func New() *MetricStorage {
	return &MetricStorage{
		metrics: make(map[string]float64),
	}
}

func (s *MetricStorage) Metrics() map[string]float64 {
	return s.metrics
}

func (s *MetricStorage) Store(metric string, value float64) {
	s.metrics[metric] = value
}

func (s *MetricStorage) Clean() {
	s.metrics = make(map[string]float64)
}

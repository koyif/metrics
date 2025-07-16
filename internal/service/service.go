package service

type metricsRepository interface {
	StoreCounter(metricName string, value int64) error
	Counter(metricName string) (int64, error)
	AllCounters() map[string]int64
	StoreGauge(metricName string, value float64) error
	Gauge(metricName string) (float64, error)
	AllGauges() map[string]float64
}

type MetricsService struct {
	repository metricsRepository
}

func NewMetricsService(repository metricsRepository) *MetricsService {
	return &MetricsService{
		repository: repository,
	}
}

func (m MetricsService) StoreGauge(metricName string, value float64) error {
	return m.repository.StoreGauge(metricName, value)
}

func (m MetricsService) StoreCounter(metricName string, value int64) error {
	return m.repository.StoreCounter(metricName, value)
}

func (m MetricsService) Counter(metricName string) (int64, error) {
	return m.repository.Counter(metricName)
}

func (m MetricsService) AllCounters() map[string]int64 {
	return m.repository.AllCounters()
}

func (m MetricsService) Gauge(metricName string) (float64, error) {
	return m.repository.Gauge(metricName)
}

func (m MetricsService) AllGauges() map[string]float64 {
	return m.repository.AllGauges()
}

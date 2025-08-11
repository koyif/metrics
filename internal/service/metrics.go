package service

import "github.com/koyif/metrics/internal/app/logger"

type metricsRepository interface {
	StoreCounter(metricName string, value int64) error
	Counter(metricName string) (int64, error)
	AllCounters() map[string]int64
	StoreGauge(metricName string, value float64) error
	Gauge(metricName string) (float64, error)
	AllGauges() map[string]float64
}

type fileService interface {
	Persist() error
}

type MetricsService struct {
	repository  metricsRepository
	fileService fileService
}

func NewMetricsService(repository metricsRepository, fileService fileService) *MetricsService {
	return &MetricsService{
		repository:  repository,
		fileService: fileService,
	}
}

func (m MetricsService) Persist() error {
	logger.Log.Info("persisting metrics")
	return m.fileService.Persist()
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

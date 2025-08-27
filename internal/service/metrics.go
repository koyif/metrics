package service

import (
	"context"
	"github.com/koyif/metrics/internal/models"
)

type repository interface {
	StoreCounter(metricName string, value int64) error
	Counter(metricName string) (int64, error)
	AllCounters() map[string]int64
	StoreGauge(metricName string, value float64) error
	Gauge(metricName string) (float64, error)
	AllGauges() map[string]float64
	StoreAll(metrics []models.Metrics) error
	Ping(ctx context.Context) error
}

type fileService interface {
	Persist() error
}

type MetricsService struct {
	repository  repository
	fileService fileService
}

func NewMetricsService(repository repository, fileService fileService) *MetricsService {
	return &MetricsService{
		repository:  repository,
		fileService: fileService,
	}
}

func (m MetricsService) Persist() error {
	return m.fileService.Persist()
}

func (m MetricsService) StoreGauge(metricName string, value float64) error {
	return m.repository.StoreGauge(metricName, value)
}

func (m MetricsService) StoreCounter(metricName string, value int64) error {
	return m.repository.StoreCounter(metricName, value)
}

func (m MetricsService) StoreAll(metrics []models.Metrics) error {
	return m.repository.StoreAll(metrics)
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

func (m MetricsService) Ping(ctx context.Context) error {
	return m.repository.Ping(ctx)
}

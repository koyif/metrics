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

// MetricsService provides business logic for metrics storage and retrieval.
// It acts as an intermediary between HTTP handlers and the storage layer (repository).
//
// The service supports both in-memory and database-backed storage through
// the repository interface, and handles file persistence when configured.
type MetricsService struct {
	repository  repository
	fileService fileService
}

// NewMetricsService creates a new metrics service with the specified repository and file service.
// The fileService can be nil if file persistence is not required (e.g., when using database storage).
func NewMetricsService(repository repository, fileService fileService) *MetricsService {
	return &MetricsService{
		repository:  repository,
		fileService: fileService,
	}
}

// Persist triggers immediate file persistence of all metrics.
// This is typically called when StoreInterval is set to 0 (synchronous mode).
// Returns an error if file writing fails.
func (m MetricsService) Persist() error {
	return m.fileService.Persist()
}

// StoreGauge stores or updates a gauge metric with the given name and value.
// Gauge metrics represent current state and are replaced on each update.
func (m MetricsService) StoreGauge(metricName string, value float64) error {
	return m.repository.StoreGauge(metricName, value)
}

// StoreCounter stores or updates a counter metric with the given name and delta value.
// Counter metrics are cumulative - the delta is added to the existing value.
func (m MetricsService) StoreCounter(metricName string, value int64) error {
	return m.repository.StoreCounter(metricName, value)
}

// StoreAll stores or updates multiple metrics in a single batch operation.
// This is more efficient than individual updates when processing many metrics at once.
func (m MetricsService) StoreAll(metrics []models.Metrics) error {
	return m.repository.StoreAll(metrics)
}

// Counter retrieves the current value of a counter metric by name.
// Returns an error if the metric doesn't exist or cannot be retrieved.
func (m MetricsService) Counter(metricName string) (int64, error) {
	return m.repository.Counter(metricName)
}

// AllCounters returns a map of all counter metrics and their current values.
// The returned map is a copy and can be safely modified.
func (m MetricsService) AllCounters() map[string]int64 {
	return m.repository.AllCounters()
}

// Gauge retrieves the current value of a gauge metric by name.
// Returns an error if the metric doesn't exist or cannot be retrieved.
func (m MetricsService) Gauge(metricName string) (float64, error) {
	return m.repository.Gauge(metricName)
}

// AllGauges returns a map of all gauge metrics and their current values.
// The returned map is a copy and can be safely modified.
func (m MetricsService) AllGauges() map[string]float64 {
	return m.repository.AllGauges()
}

// Ping checks the health of the underlying storage layer.
// For database storage, this performs a database connectivity check.
// For in-memory storage, this always returns nil.
func (m MetricsService) Ping(ctx context.Context) error {
	return m.repository.Ping(ctx)
}

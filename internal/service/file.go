package service

import (
	"context"
	"github.com/koyif/metrics/pkg/logger"
	"sync"
	"time"

	models "github.com/koyif/metrics/internal/models"
)

type metricsRepository interface {
	StoreCounter(metricName string, value int64) error
	Counter(metricName string) (int64, error)
	AllCounters() map[string]int64
	StoreGauge(metricName string, value float64) error
	Gauge(metricName string) (float64, error)
	AllGauges() map[string]float64
}

type fileRepository interface {
	Save(metrics []models.Metrics) error
	Load() ([]models.Metrics, error)
}

type FileService struct {
	fileRepository    fileRepository
	metricsRepository metricsRepository
}

func NewFileService(fileRepository fileRepository, metricsRepository metricsRepository) *FileService {
	return &FileService{
		fileRepository:    fileRepository,
		metricsRepository: metricsRepository,
	}
}

func (s *FileService) Persist() error {
	logger.Log.Info("persisting metrics")

	metrics := make([]models.Metrics, 0)
	for metricName, value := range s.metricsRepository.AllGauges() {
		metrics = append(metrics, models.Metrics{
			ID:    metricName,
			MType: models.Gauge,
			Value: &value,
		})
	}
	for metricName, value := range s.metricsRepository.AllCounters() {
		metrics = append(metrics, models.Metrics{
			ID:    metricName,
			MType: models.Counter,
			Delta: &value,
		})
	}

	return s.fileRepository.Save(metrics)
}

func (s *FileService) Restore() error {
	metrics, err := s.fileRepository.Load()
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			s.metricsRepository.StoreGauge(metric.ID, *metric.Value)
		case models.Counter:
			s.metricsRepository.StoreCounter(metric.ID, *metric.Delta)
		default:
			logger.Log.Warn("unknown metric type", logger.String("metricType", metric.MType))
		}
	}
	return nil
}

func (s *FileService) SchedulePersist(ctx context.Context, wg *sync.WaitGroup, interval time.Duration) {
	if interval == 0 {
		return
	}

	logger.Log.Info("scheduling persist", logger.String("interval", interval.String()))

	go func() {
		ticker := time.NewTicker(interval)
		defer wg.Done()
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.persist()
				return
			case <-ticker.C:
				s.persist()
			}
		}
	}()
}

func (s *FileService) persist() {
	if err := s.Persist(); err != nil {
		logger.Log.Error("error persisting metrics", logger.Error(err))
	}
}

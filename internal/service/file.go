package service

import (
	"fmt"
	"time"

	"github.com/koyif/metrics/internal/app/logger"
	models "github.com/koyif/metrics/internal/model"
	"github.com/koyif/metrics/internal/repository"
)

type FileService struct {
	fileRepository    *repository.FileRepository
	metricsRepository *repository.MetricsRepository
}

func NewFileService(fileRepository *repository.FileRepository, metricsRepository *repository.MetricsRepository) *FileService {
	return &FileService{
		fileRepository:    fileRepository,
		metricsRepository: metricsRepository,
	}
}

func (s *FileService) Persist() error {
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
			return fmt.Errorf("unknown metric type: %s", metric.MType)
		}
	}
	return nil
}

func (s *FileService) SchedulePersist(interval time.Duration) {
	if interval == 0 {
		return
	}

	logger.Log.Info("scheduling persist", logger.String("interval", interval.String()))

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := s.Persist(); err != nil {
				logger.Log.Error("error persisting metrics", logger.Error(err))
			}
		}
	}()
}

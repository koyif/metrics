package app

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/koyif/metrics/internal/app/logger"
	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/internal/persistence/database"
	"github.com/koyif/metrics/internal/repository"
	"github.com/koyif/metrics/internal/service"
	"github.com/koyif/metrics/internal/service/ping"
)

type metricsRepository interface {
	StoreCounter(metricName string, value int64) error
	Counter(metricName string) (int64, error)
	AllCounters() map[string]int64
	StoreGauge(metricName string, value float64) error
	Gauge(metricName string) (float64, error)
	AllGauges() map[string]float64
	StoreAll(metrics []models.Metrics) error
}

type App struct {
	Config         *config.Config
	MetricsService *service.MetricsService
	PingService    *ping.Service
}

func New(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config) *App {
	var metricsRepository metricsRepository
	var pingService *ping.Service
	var fileService *service.FileService

	if cfg.Storage.DatabaseURL != "" {
		wg.Done()
		db := database.New(ctx, cfg.Storage.DatabaseURL)
		pingService = ping.NewService(db)
		metricsRepository = repository.NewDatabaseRepository(db)
	} else {
		fileRepository := repository.NewFileRepository(cfg.Storage.FileStoragePath)
		metricsRepository = repository.NewMetricsRepository()
		fileService = service.NewFileService(fileRepository, metricsRepository)
		if cfg.Storage.Restore {
			if err := fileService.Restore(); err != nil && !errors.Is(err, io.EOF) {
				logger.Log.Error("error restoring metrics", logger.Error(err))
			}
		}

		fileService.SchedulePersist(ctx, wg, cfg.Storage.StoreInterval)
	}

	metricsService := service.NewMetricsService(metricsRepository, fileService)

	return &App{
		Config:         cfg,
		MetricsService: metricsService,
		PingService:    pingService,
	}
}

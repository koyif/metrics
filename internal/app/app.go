package app

import (
	"github.com/koyif/metrics/internal/app/logger"
	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/internal/repository"
	"github.com/koyif/metrics/internal/service"
)

type App struct {
	Config         *config.Config
	MetricsService *service.MetricsService
}

func New(cfg *config.Config) *App {
	fileRepository := repository.NewFileRepository(cfg.Storage.FileStoragePath)
	metricsRepository := repository.NewMetricsRepository()
	fileService := service.NewFileService(fileRepository, metricsRepository)
	fileService.SchedulePersist(cfg.Storage.StoreInterval)

	if cfg.Storage.Restore {
		err := fileService.Restore()
		if err != nil {
			logger.Log.Error("error restoring metrics", logger.Error(err))
		}
	}

	metricsService := service.NewMetricsService(metricsRepository, fileService)

	return &App{
		Config:         cfg,
		MetricsService: metricsService,
	}
}

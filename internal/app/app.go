package app

import (
	"context"
	"errors"
	"github.com/koyif/metrics/internal/app/logger"
	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/internal/repository"
	"github.com/koyif/metrics/internal/service"
	"io"
	"sync"
)

type App struct {
	Config         *config.Config
	MetricsService *service.MetricsService
	Context        context.Context
}

func New(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config) *App {
	fileRepository := repository.NewFileRepository(cfg.Storage.FileStoragePath)
	metricsRepository := repository.NewMetricsRepository()
	fileService := service.NewFileService(fileRepository, metricsRepository)
	if cfg.Storage.Restore {
		if err := fileService.Restore(); err != nil && !errors.Is(err, io.EOF) {
			logger.Log.Error("error restoring metrics", logger.Error(err))
		}
	}

	fileService.SchedulePersist(ctx, wg, cfg.Storage.StoreInterval)

	metricsService := service.NewMetricsService(metricsRepository, fileService)

	return &App{
		Config:         cfg,
		MetricsService: metricsService,
		Context:        ctx,
	}
}

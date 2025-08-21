package app

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/koyif/metrics/internal/app/logger"
	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/internal/persistence/database"
	"github.com/koyif/metrics/internal/repository"
	"github.com/koyif/metrics/internal/service"
	"github.com/koyif/metrics/internal/service/ping"
)

type App struct {
	Config         *config.Config
	MetricsService *service.MetricsService
	PingService    *ping.Service
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

	var pingService *ping.Service
	if cfg.Storage.DatabaseURL != "" {
		db := database.New(ctx, cfg.Storage.DatabaseURL)
		pingService = ping.NewService(db)
	}

	return &App{
		Config:         cfg,
		MetricsService: metricsService,
		PingService:    pingService,
	}
}

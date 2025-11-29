package app

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/koyif/metrics/pkg/logger"

	"github.com/koyif/metrics/internal/audit"
	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/internal/models"
	"github.com/koyif/metrics/internal/persistence/database"
	"github.com/koyif/metrics/internal/repository"
	"github.com/koyif/metrics/internal/service"
)

type metricsRepository interface {
	StoreCounter(metricName string, value int64) error
	Counter(metricName string) (int64, error)
	AllCounters() map[string]int64
	StoreGauge(metricName string, value float64) error
	Gauge(metricName string) (float64, error)
	AllGauges() map[string]float64
	StoreAll(metrics []models.Metrics) error
	Ping(ctx context.Context) error
}

type App struct {
	Config         *config.Config
	MetricsService *service.MetricsService
	AuditManager   *audit.Manager
}

func New(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config) (*App, error) {
	var metricsRepository metricsRepository
	var fileService *service.FileService

	if cfg.Storage.DatabaseURL != "" {
		wg.Done()
		db, err := database.New(ctx, cfg.Storage.DatabaseURL)
		if err != nil {
			return nil, err
		}
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

		fileService.SchedulePersist(ctx, wg, cfg.Storage.StoreInterval.Value())
	}

	metricsService := service.NewMetricsService(metricsRepository, fileService)

	auditManager := initializeAudit(cfg)

	return &App{
		Config:         cfg,
		MetricsService: metricsService,
		AuditManager:   auditManager,
	}, nil
}

func initializeAudit(cfg *config.Config) *audit.Manager {
	manager := audit.NewManager()

	if cfg.Audit.FilePath != "" {
		fileAuditor, err := audit.NewFileAuditor(cfg.Audit.FilePath)
		if err != nil {
			logger.Log.Error("failed to create file auditor", logger.Error(err))
		} else {
			manager.AddObserver(fileAuditor)
			logger.Log.Info("file audit enabled", logger.String("path", cfg.Audit.FilePath))
		}
	}

	if cfg.Audit.URL != "" {
		httpAuditor, err := audit.NewHTTPAuditor(cfg.Audit.URL)
		if err != nil {
			logger.Log.Error("failed to create HTTP auditor", logger.Error(err))
		} else {
			manager.AddObserver(httpAuditor)
			logger.Log.Info("HTTP audit enabled", logger.String("url", cfg.Audit.URL))
		}
	}

	return manager
}

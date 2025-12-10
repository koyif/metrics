package app

import (
	"context"
	"crypto/rsa"
	"errors"
	"io"
	"sync"

	"github.com/koyif/metrics/pkg/crypto"
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
	PrivateKey     *rsa.PrivateKey
}

func New(ctx context.Context, wg *sync.WaitGroup, cfg *config.Config) (*App, error) {
	var metricsRepository metricsRepository
	var fileService *service.FileService

	if cfg.DatabaseURL != "" {
		wg.Done()
		db, err := database.New(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, err
		}
		metricsRepository = repository.NewDatabaseRepository(db)
	} else {
		fileRepository := repository.NewFileRepository(cfg.FileStoragePath)
		metricsRepository = repository.NewMetricsRepository()
		fileService = service.NewFileService(fileRepository, metricsRepository)
		if cfg.Restore {
			if err := fileService.Restore(); err != nil && !errors.Is(err, io.EOF) {
				logger.Log.Error("error restoring metrics", logger.Error(err))
			}
		}

		fileService.SchedulePersist(ctx, wg, cfg.StoreInterval.Value())
	}

	metricsService := service.NewMetricsService(metricsRepository, fileService)

	auditManager := initializeAudit(cfg)

	var privateKey *rsa.PrivateKey
	if cfg.CryptoKey != "" {
		key, err := crypto.LoadPrivateKey(cfg.CryptoKey)
		if err != nil {
			return nil, err
		}
		privateKey = key
		logger.Log.Info("private key loaded successfully for decryption")
	}

	return &App{
		Config:         cfg,
		MetricsService: metricsService,
		AuditManager:   auditManager,
		PrivateKey:     privateKey,
	}, nil
}

func initializeAudit(cfg *config.Config) *audit.Manager {
	manager := audit.NewManager()

	if cfg.FilePath != "" {
		fileAuditor, err := audit.NewFileAuditor(cfg.FilePath)
		if err != nil {
			logger.Log.Error("failed to create file auditor", logger.Error(err))
		} else {
			manager.AddObserver(fileAuditor)
			logger.Log.Info("file audit enabled", logger.String("path", cfg.FilePath))
		}
	}

	if cfg.URL != "" {
		httpAuditor, err := audit.NewHTTPAuditor(cfg.URL)
		if err != nil {
			logger.Log.Error("failed to create HTTP auditor", logger.Error(err))
		} else {
			manager.AddObserver(httpAuditor)
			logger.Log.Info("HTTP audit enabled", logger.String("url", cfg.URL))
		}
	}

	return manager
}

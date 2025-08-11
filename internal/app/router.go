package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/internal/handler"
	"github.com/koyif/metrics/internal/handler/metrics"
	"github.com/koyif/metrics/internal/handler/middleware"
	"github.com/koyif/metrics/internal/repository"
	"github.com/koyif/metrics/internal/service"
)

func Router(cfg *config.Config) *chi.Mux {
	mux := chi.NewMux()
	mux.Use(middleware.WithLogger)
	mux.Use(middleware.WithGzip)

	fileRepository := repository.NewFileRepository(cfg.Storage.FileStoragePath)
	metricsRepository := repository.NewMetricsRepository()
	fileService := service.NewFileService(fileRepository, metricsRepository)
	fileService.SchedulePersist(cfg.Storage.StoreInterval)

	if cfg.Storage.Restore {
		fileService.Restore()
	}

	metricsService := service.NewMetricsService(metricsRepository, fileService)

	summaryHandler := handler.NewSummaryHandler(metricsService)

	gaugesGetHandler := handler.NewGaugesGetHandler(metricsService)
	countersGetHandler := handler.NewCountersGetHandler(metricsService)

	countersPostHandler := handler.NewCountersPostHandler(metricsService)
	gaugesPostHandler := handler.NewGaugesPostHandler(metricsService)

	getHandler := metrics.NewGetHandler(metricsService)
	storeHandler := metrics.NewStoreHandler(metricsService, cfg)

	mux.HandleFunc("/", summaryHandler.Handle)
	mux.HandleFunc("/value/gauge/{metric}", gaugesGetHandler.Handle)
	mux.HandleFunc("/value/counter/{metric}", countersGetHandler.Handle)

	mux.HandleFunc("/update/counter/{metric}/{value}", countersPostHandler.Handle)
	mux.HandleFunc("/update/counter/", countersPostHandler.Handle)
	mux.HandleFunc("/update/gauge/{metric}/{value}", gaugesPostHandler.Handle)
	mux.HandleFunc("/update/gauge/", gaugesPostHandler.Handle)

	mux.HandleFunc("/update", storeHandler.Handle)
	mux.HandleFunc("/update/", storeHandler.Handle)
	mux.HandleFunc("/value", getHandler.Handle)
	mux.HandleFunc("/value/", getHandler.Handle)

	mux.HandleFunc("/update/{anything}/", handler.UnknownMetricTypeHandler)
	mux.HandleFunc("/update/{anything}/{metric}/", handler.UnknownMetricTypeHandler)
	mux.HandleFunc("/update/{anything}/{metric}/{value}", handler.UnknownMetricTypeHandler)

	return mux
}

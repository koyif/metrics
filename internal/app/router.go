package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/koyif/metrics/internal/handler"
	"github.com/koyif/metrics/internal/repository"
	"github.com/koyif/metrics/internal/service"
)

func Router() *chi.Mux {
	mux := chi.NewMux()

	metricsRepository := repository.NewMetricsRepository()

	metricsService := service.NewMetricsService(metricsRepository)

	summaryHandler := handler.NewSummaryHandler(metricsService)

	gaugesGetHandler := handler.NewGaugesGetHandler(metricsService)
	countersGetHandler := handler.NewCountersGetHandler(metricsService)

	countersPostHandler := handler.NewCountersPostHandler(metricsService)
	gaugesPostHandler := handler.NewGaugesPostHandler(metricsService)

	mux.HandleFunc("/", summaryHandler.Handle)
	mux.HandleFunc("/value/gauge/{metric}", gaugesGetHandler.Handle)
	mux.HandleFunc("/value/counter/{metric}", countersGetHandler.Handle)

	mux.HandleFunc("/update/counter/{metric}/{value}", countersPostHandler.Handle)
	mux.HandleFunc("/update/counter/", countersPostHandler.Handle)
	mux.HandleFunc("/update/gauge/{metric}/{value}", gaugesPostHandler.Handle)
	mux.HandleFunc("/update/gauge/", gaugesPostHandler.Handle)

	mux.HandleFunc("/update/{anything}/", handler.UnknownMetricTypeHandler)
	mux.HandleFunc("/update/{anything}/{metric}/", handler.UnknownMetricTypeHandler)
	mux.HandleFunc("/update/{anything}/{metric}/{value}", handler.UnknownMetricTypeHandler)

	return mux
}

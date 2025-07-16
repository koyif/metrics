package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/internal/handler"
	"github.com/koyif/metrics/internal/repository"
	"github.com/koyif/metrics/internal/service"
	"net/http"
)

func main() {
	cfg := config.Load()

	if err := run(cfg); err != nil {
		panic(err)
	}
}

func run(cfg *config.Config) error {
	return http.ListenAndServe(cfg.Server.Addr, router())
}

func router() *chi.Mux {
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

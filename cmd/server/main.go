package main

import (
	"fmt"
	"github.com/koyif/metrics/internal/config"
	"github.com/koyif/metrics/internal/handler"
	"github.com/koyif/metrics/internal/repository"
	"net/http"
)

func main() {
	cfg := config.Load()

	if err := run(cfg); err != nil {
		panic(err)
	}
}

func run(cfg *config.Config) error {
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	return http.ListenAndServe(addr, router())
}

func router() *http.ServeMux {
	mux := http.NewServeMux()

	metricsRepository := repository.NewMetricsRepository()

	counterHandler := handler.NewCountersHandler(metricsRepository)
	gaugeHandler := handler.NewGaugesHandler(metricsRepository)

	mux.HandleFunc("/update/counter/{metric}/{value}", counterHandler.Handle)
	mux.HandleFunc("/update/counter/", counterHandler.Handle)
	mux.HandleFunc("/update/gauge/{metric}/{value}", gaugeHandler.Handle)
	mux.HandleFunc("/update/gauge/", gaugeHandler.Handle)
	mux.HandleFunc("/update/{anything}/", handler.UnknownMetricTypeHandler)

	return mux
}

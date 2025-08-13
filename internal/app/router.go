package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/koyif/metrics/internal/handler"
	"github.com/koyif/metrics/internal/handler/metrics"
	"github.com/koyif/metrics/internal/handler/middleware"
)

func (app *App) Router() *chi.Mux {
	mux := chi.NewMux()
	mux.Use(middleware.WithLogger)
	mux.Use(middleware.WithGzip)

	mux.HandleFunc("/", handler.NewSummaryHandler(app.MetricsService).Handle)
	mux.HandleFunc("/value/gauge/{metric}", handler.NewGaugesGetHandler(app.MetricsService).Handle)
	mux.HandleFunc("/value/counter/{metric}", handler.NewCountersGetHandler(app.MetricsService).Handle)

	countersPostHandler := handler.NewCountersPostHandler(app.MetricsService)
	mux.HandleFunc("/update/counter/{metric}/{value}", countersPostHandler.Handle)
	mux.HandleFunc("/update/counter/", countersPostHandler.Handle)

	gaugesPostHandler := handler.NewGaugesPostHandler(app.MetricsService)
	mux.HandleFunc("/update/gauge/{metric}/{value}", gaugesPostHandler.Handle)
	mux.HandleFunc("/update/gauge/", gaugesPostHandler.Handle)

	storeHandler := metrics.NewStoreHandler(app.MetricsService, app.Config)
	mux.HandleFunc("/update", storeHandler.Handle)
	mux.HandleFunc("/update/", storeHandler.Handle)

	getHandler := metrics.NewGetHandler(app.MetricsService)
	mux.HandleFunc("/value", getHandler.Handle)
	mux.HandleFunc("/value/", getHandler.Handle)

	mux.HandleFunc("/update/{anything}/", handler.UnknownMetricTypeHandler)
	mux.HandleFunc("/update/{anything}/{metric}/", handler.UnknownMetricTypeHandler)
	mux.HandleFunc("/update/{anything}/{metric}/{value}", handler.UnknownMetricTypeHandler)

	return mux
}

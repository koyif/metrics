package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/koyif/metrics/internal/handler/health"

	"github.com/koyif/metrics/internal/handler"
	"github.com/koyif/metrics/internal/handler/deprecated"
	"github.com/koyif/metrics/internal/handler/metrics"
	"github.com/koyif/metrics/internal/handler/middleware"
)

func (app App) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.WithLogger)
	r.Use(middleware.WithGzip)

	summaryHandler := metrics.NewSummaryHandler(app.MetricsService)
	getHandler := metrics.NewGetHandler(app.MetricsService)
	storeHandler := metrics.NewStoreHandler(app.MetricsService, app.Config)
	storeAllHandler := metrics.NewStoreAllHandler(app.MetricsService, app.Config)

	counterGetHandler := deprecated.NewCountersGetHandler(app.MetricsService)
	gaugeGetHandler := deprecated.NewGaugesGetHandler(app.MetricsService)
	counterPostHandler := deprecated.NewCountersPostHandler(app.MetricsService)
	gaugePostHandler := deprecated.NewGaugesPostHandler(app.MetricsService)

	r.Get("/", summaryHandler.Handle)

	r.Post("/updates/", storeAllHandler.Handle)

	pingHandler := health.NewPingHandler(app.MetricsService)
	r.Get("/ping", pingHandler.Handle)

	r.Route("/value", func(r chi.Router) {
		r.Post("/", getHandler.Handle)

		r.Route("/counter", func(r chi.Router) {
			r.NotFound(handler.MetricNotFound)
			r.Get("/{metric}", counterGetHandler.Handle)
		})

		r.Route("/gauge", func(r chi.Router) {
			r.NotFound(handler.MetricNotFound)
			r.Get("/{metric}", gaugeGetHandler.Handle)
		})
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/", storeHandler.Handle)

		r.Route("/counter", func(r chi.Router) {
			r.NotFound(handler.MetricNotFound)
			r.Post("/{metric}/{value}", counterPostHandler.Handle)
		})

		r.Route("/gauge", func(r chi.Router) {
			r.NotFound(handler.MetricNotFound)
			r.Post("/{metric}/{value}", gaugePostHandler.Handle)
		})
	})

	r.NotFound(handler.UnknownMetricTypeHandler)

	return r
}

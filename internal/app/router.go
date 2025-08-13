package app

import (
	"github.com/go-chi/chi/v5"
	"github.com/koyif/metrics/internal/handler"
	"github.com/koyif/metrics/internal/handler/deprecated"
	"github.com/koyif/metrics/internal/handler/metrics"
	"github.com/koyif/metrics/internal/handler/middleware"
)

func (app *App) Router() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.WithLogger)
	r.Use(middleware.WithGzip)

	r.Get("/", metrics.NewSummaryHandler(app.MetricsService).Handle)

	r.Route("/value", func(r chi.Router) {
		r.Post("/", metrics.NewGetHandler(app.MetricsService).Handle)

		r.Route("/counter", func(r chi.Router) {
			r.NotFound(handler.MetricNotFound)
			r.Get("/{metric}", deprecated.NewCountersGetHandler(app.MetricsService).Handle)
		})

		r.Route("/gauge", func(r chi.Router) {
			r.NotFound(handler.MetricNotFound)
			r.Get("/{metric}", deprecated.NewGaugesGetHandler(app.MetricsService).Handle)
		})
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/", metrics.NewStoreHandler(app.MetricsService, app.Config).Handle)

		r.Route("/counter", func(r chi.Router) {
			r.NotFound(handler.MetricNotFound)
			r.Post("/{metric}/{value}", deprecated.NewCountersPostHandler(app.MetricsService).Handle)
		})

		r.Route("/gauge", func(r chi.Router) {
			r.NotFound(handler.MetricNotFound)
			r.Post("/{metric}/{value}", deprecated.NewGaugesPostHandler(app.MetricsService).Handle)
		})
	})

	r.NotFound(handler.UnknownMetricTypeHandler)

	return r
}

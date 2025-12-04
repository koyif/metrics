package app

import (
	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	swagger "github.com/swaggo/http-swagger/v2"

	"github.com/koyif/metrics/internal/handler/health"

	"github.com/koyif/metrics/internal/handler"
	"github.com/koyif/metrics/internal/handler/deprecated"
	"github.com/koyif/metrics/internal/handler/metrics"
	custommiddleware "github.com/koyif/metrics/internal/handler/middleware"
)

func (app App) Router() *chi.Mux {
	r := chi.NewRouter()

	r.Use(custommiddleware.WithLogger)
	r.Use(custommiddleware.WithGzip)

	if app.PrivateKey != nil {
		r.Use(custommiddleware.WithDecryption(app.PrivateKey))
	}

	if app.Config.HashKey != "" {
		r.Use(custommiddleware.WithHashCheck(app.Config.HashKey))
	}

	summaryHandler := metrics.NewSummaryHandler(app.MetricsService)
	getHandler := metrics.NewGetHandler(app.MetricsService)
	storeHandler := metrics.NewStoreHandler(app.MetricsService, app.Config, app.AuditManager)
	storeAllHandler := metrics.NewStoreAllHandler(app.MetricsService, app.Config, app.AuditManager)

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

	r.Mount("/debug", middleware.Profiler())

	// Swagger documentation
	r.Get("/swagger/*", swagger.Handler(
		swagger.URL("/swagger/doc.json"),
	))

	return r
}

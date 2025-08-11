package main

import (
	"net/http"

	"github.com/koyif/metrics/internal/app"
	"github.com/koyif/metrics/internal/app/logger"
	"github.com/koyif/metrics/internal/config"
)

func main() {
	cfg := config.Load()
	if err := logger.Initialize(); err != nil {
		logger.Log.Fatal("error starting logger", logger.Error(err))
	}

	if err := run(cfg); err != nil {
		logger.Log.Fatal("error starting server", logger.Error(err))
	}
}

func run(cfg *config.Config) error {
	return http.ListenAndServe(cfg.Server.Addr, app.Router(cfg))
}

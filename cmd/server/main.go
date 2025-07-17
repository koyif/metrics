package main

import (
	"fmt"
	"github.com/koyif/metrics/internal/app"
	"github.com/koyif/metrics/internal/config"
	"log/slog"
	"net/http"
)

func main() {
	cfg := config.Load()

	if err := run(cfg); err != nil {
		slog.Error(fmt.Sprintf("error starting server %v", err))
	}
}

func run(cfg *config.Config) error {
	return http.ListenAndServe(cfg.Server.Addr, app.Router())
}

package main

import (
	"fmt"
	"github.com/koyif/metrics/internal/agent/app"
	"github.com/koyif/metrics/internal/agent/config"
	"log/slog"
	"os"
)

func main() {
	cfg := config.Load()
	configureLogger()

	a := app.New(cfg)

	if err := a.Run(); err != nil {
		slog.Error(fmt.Sprintf("error starting agent %v", err))
		os.Exit(1)
	}
}

func configureLogger() {
	log := slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelDebug,
			AddSource: true,
		},
	))

	slog.SetDefault(log)
}
